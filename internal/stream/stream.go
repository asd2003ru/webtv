package stream

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/asd2003ru/webtv/internal/logx"
)

type Mode string

type AudioTrack struct {
	Index    int    `json:"index"`
	Codec    string `json:"codec"`
	Language string `json:"language"`
	Title    string `json:"title"`
	Default  bool   `json:"default"`
}

type transcodePlan struct {
	videoCodec string
	audioCodec string
	videoCopy  bool
	audioCopy  bool
}

type PathPlan struct {
	VideoPath  string
	AudioPath  string
	VideoCodec string
	AudioCodec string
}

const (
	ModeDirect    Mode = "direct"
	ModeTranscode Mode = "transcode"
)

type Manager struct {
	mu                    sync.Mutex
	policy                string
	ffmpegBin             string
	idleTimeout           time.Duration
	maxRetries            int
	enableManifestBackoff bool
	client                *http.Client
	probeMu               sync.Mutex
	probe                 map[string]probeEntry
	manifestMu            sync.Mutex
	manifests             map[string]manifestEntry
	inFlight              map[string]chan struct{}
	logMu                 sync.Mutex
	logStats              map[string]logEntry
}

type probeEntry struct {
	mode Mode
	at   time.Time
}

type manifestEntry struct {
	status     int
	header     http.Header
	body       []byte
	freshUntil time.Time
	staleUntil time.Time
	unchanged  int
	backoffTo  time.Time
}

type logEntry struct {
	count int
	since time.Time
}

const (
	manifestFreshTTL = 2 * time.Second
	manifestStaleTTL = 20 * time.Second
	logWindow        = 5 * time.Second
)

func NewManager(policy, ffmpegBin string, idleTimeout time.Duration, maxRetries int, enableManifestBackoff bool) *Manager {
	if idleTimeout <= 0 {
		idleTimeout = 12 * time.Second
	}
	if maxRetries < 0 {
		maxRetries = 0
	}
	return &Manager{
		policy:                policy,
		ffmpegBin:             ffmpegBin,
		idleTimeout:           idleTimeout,
		maxRetries:            maxRetries,
		enableManifestBackoff: enableManifestBackoff,
		client:                &http.Client{Timeout: 8 * time.Second},
		probe:                 make(map[string]probeEntry),
		manifests:             make(map[string]manifestEntry),
		inFlight:              make(map[string]chan struct{}),
		logStats:              make(map[string]logEntry),
	}
}

func (m *Manager) lock() error {
	// Browser players can reopen/retry stream URLs; do not hard-block concurrent
	// requests with 409 because it breaks playback.
	return nil
}

func (m *Manager) unlock() {
}

func BrowserMode(streamURL string) Mode {
	u := strings.ToLower(streamURL)
	if strings.Contains(u, "ac3") ||
		strings.Contains(u, "a52") ||
		strings.Contains(u, "mp2") ||
		strings.Contains(u, "mpga") {
		return ModeTranscode
	}
	return ModeDirect
}

func BuildArchiveURL(streamURL string, startAt, endAt time.Time) string {
	startUnix := startAt.UTC().Unix()
	endUnix := endAt.UTC().Unix()
	duration := int(endAt.Sub(startAt).Seconds())
	if duration < 0 {
		duration = 0
	}

	replacer := strings.NewReplacer(
		"{start}", strconv.FormatInt(startUnix, 10),
		"{end}", strconv.FormatInt(endUnix, 10),
		"{utc}", strconv.FormatInt(startUnix, 10),
		"{lutc}", strconv.FormatInt(endUnix, 10),
		"{timestamp}", strconv.FormatInt(startUnix, 10),
		"{duration}", strconv.Itoa(duration),
	)
	rewritten := replacer.Replace(streamURL)
	if rewritten != streamURL {
		return rewritten
	}

	u, err := url.Parse(streamURL)
	if err != nil {
		return streamURL
	}
	q := u.Query()
	q.Set("start", strconv.FormatInt(startUnix, 10))
	q.Set("end", strconv.FormatInt(endUnix, 10))
	q.Set("utc", strconv.FormatInt(startUnix, 10))
	q.Set("lutc", strconv.FormatInt(endUnix, 10))
	q.Set("duration", fmt.Sprintf("%d", duration))
	u.RawQuery = q.Encode()
	return u.String()
}

func (m *Manager) Handle(w http.ResponseWriter, r *http.Request, streamURL string) {
	if err := m.lock(); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	defer m.unlock()

	mode := m.modeForURL(r, streamURL, true)
	w.Header().Set("X-Stream-Mode", string(mode))
	m.logBurst("stream_handle mode="+string(mode), streamURL)

	if mode == ModeDirect {
		m.proxy(w, r, streamURL)
		return
	}
	m.transcode(w, r, streamURL)
}

func (m *Manager) DetectMode(r *http.Request, streamURL string) Mode {
	return m.modeForURL(r, streamURL, true)
}

func (m *Manager) modeForURL(r *http.Request, streamURL string, allowProbe bool) Mode {
	mode := BrowserMode(streamURL)
	if mode == ModeTranscode {
		return mode
	}
	if !allowProbe {
		return mode
	}
	if !strings.Contains(strings.ToLower(streamURL), ".m3u8") {
		return mode
	}
	if probed, ok := m.probeCodecMode(streamURL); ok {
		return probed
	}
	probed, ok := m.probeHLSMode(r, streamURL)
	if ok {
		return probed
	}
	return mode
}

func (m *Manager) HandleWithMode(w http.ResponseWriter, r *http.Request, streamURL string, mode Mode) {
	if err := m.lock(); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	defer m.unlock()

	w.Header().Set("X-Stream-Mode", string(mode))
	m.logBurst("stream_handle mode="+string(mode), streamURL)

	if mode == ModeDirect {
		m.proxy(w, r, streamURL)
		return
	}
	m.transcode(w, r, streamURL)
}

func (m *Manager) probeCodecMode(streamURL string) (Mode, bool) {
	const ttl = 10 * time.Minute
	now := time.Now()

	m.probeMu.Lock()
	if p, ok := m.probe[streamURL]; ok && now.Sub(p.at) < ttl {
		m.probeMu.Unlock()
		return p.mode, true
	}
	m.probeMu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, m.ffmpegBin[:len(m.ffmpegBin)-len("ffmpeg")]+"ffprobe",
		"-v", "error",
		"-show_entries", "stream=codec_name,codec_type",
		"-of", "json",
		streamURL,
	)
	// Fallback when ffmpeg path is not named as ".../ffmpeg".
	if strings.TrimSpace(m.ffmpegBin) == "ffmpeg" || strings.HasSuffix(strings.TrimSpace(m.ffmpegBin), "/ffmpeg") == false {
		cmd = exec.CommandContext(ctx, "ffprobe",
			"-v", "error",
			"-show_entries", "stream=codec_name,codec_type",
			"-of", "json",
			streamURL,
		)
	}

	out, err := cmd.Output()
	if err != nil {
		return ModeDirect, false
	}

	var payload struct {
		Streams []struct {
			CodecName string `json:"codec_name"`
			CodecType string `json:"codec_type"`
		} `json:"streams"`
	}
	if err := json.Unmarshal(out, &payload); err != nil {
		return ModeDirect, false
	}
	if len(payload.Streams) == 0 {
		return ModeDirect, false
	}

	mode := ModeDirect
	for _, s := range payload.Streams {
		cn := strings.ToLower(strings.TrimSpace(s.CodecName))
		ct := strings.ToLower(strings.TrimSpace(s.CodecType))
		if ct == "audio" {
			if cn == "ac3" || cn == "eac3" || cn == "a52" || cn == "mp2" {
				mode = ModeTranscode
				break
			}
		}
		if ct == "video" {
			if cn == "hevc" || cn == "mpeg2video" || cn == "mpeg4" {
				mode = ModeTranscode
				break
			}
		}
	}

	m.probeMu.Lock()
	m.probe[streamURL] = probeEntry{mode: mode, at: now}
	m.probeMu.Unlock()
	logx.Debugf("codec probe mode=%s url=%s", mode, streamURL)
	return mode, true
}

func (m *Manager) ProbeAudioTracks(streamURL string) ([]AudioTrack, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, m.ffmpegBin[:len(m.ffmpegBin)-len("ffmpeg")]+"ffprobe",
		"-v", "error",
		"-show_entries", "stream=index,codec_name,codec_type:stream_tags=language,title:stream_disposition=default",
		"-of", "json",
		streamURL,
	)
	if strings.TrimSpace(m.ffmpegBin) == "ffmpeg" || !strings.HasSuffix(strings.TrimSpace(m.ffmpegBin), "/ffmpeg") {
		cmd = exec.CommandContext(ctx, "ffprobe",
			"-v", "error",
			"-show_entries", "stream=index,codec_name,codec_type:stream_tags=language,title:stream_disposition=default",
			"-of", "json",
			streamURL,
		)
	}

	out, err := cmd.Output()
	if err != nil {
		return nil, false
	}

	var payload struct {
		Streams []struct {
			Index       int    `json:"index"`
			CodecName   string `json:"codec_name"`
			CodecType   string `json:"codec_type"`
			Disposition struct {
				Default int `json:"default"`
			} `json:"disposition"`
			Tags struct {
				Language string `json:"language"`
				Title    string `json:"title"`
			} `json:"tags"`
		} `json:"streams"`
	}
	if err := json.Unmarshal(out, &payload); err != nil {
		return nil, false
	}
	if len(payload.Streams) == 0 {
		return nil, false
	}

	audioTracks := make([]AudioTrack, 0, len(payload.Streams))
	ordinal := 0
	for _, s := range payload.Streams {
		if strings.ToLower(strings.TrimSpace(s.CodecType)) != "audio" {
			continue
		}
		audioTracks = append(audioTracks, AudioTrack{
			Index:    ordinal,
			Codec:    strings.ToLower(strings.TrimSpace(s.CodecName)),
			Language: strings.TrimSpace(s.Tags.Language),
			Title:    strings.TrimSpace(s.Tags.Title),
			Default:  s.Disposition.Default == 1,
		})
		ordinal++
	}
	if len(audioTracks) == 0 {
		return nil, false
	}
	return audioTracks, true
}

func (m *Manager) probeHLSMode(r *http.Request, streamURL string) (Mode, bool) {
	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, streamURL, nil)
	if err != nil {
		return ModeDirect, false
	}
	req.Header.Set("Range", "bytes=0-65535")
	resp, err := m.client.Do(req)
	if err != nil {
		return ModeDirect, false
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ModeDirect, false
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil {
		return ModeDirect, false
	}
	body := strings.ToLower(string(data))
	// Typical browser-incompatible codecs in IPTV HLS manifests.
	if strings.Contains(body, "ac-3") ||
		strings.Contains(body, "a52") ||
		strings.Contains(body, "ec-3") ||
		strings.Contains(body, "mp2") ||
		strings.Contains(body, "mpga") ||
		strings.Contains(body, "hvc1") ||
		strings.Contains(body, "hev1") {
		return ModeTranscode, true
	}
	return ModeDirect, true
}

func (m *Manager) proxy(w http.ResponseWriter, r *http.Request, streamURL string) {
	if strings.Contains(strings.ToLower(streamURL), ".m3u8") {
		m.proxyManifest(w, r, streamURL)
		return
	}

	resp, err := m.openUpstream(r.Context(), r, streamURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		m.logBurst(fmt.Sprintf("upstream_status=%d", resp.StatusCode), streamURL)
	}
	isM3U8 := isM3U8Response(streamURL, resp.Header.Get("Content-Type"))
	copyUpstreamHeaders(w, resp, isM3U8)
	w.WriteHeader(resp.StatusCode)
	if !isM3U8 {
		m.copyStreamWithRetry(w, r, streamURL, resp)
		return
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	rewritten := rewriteM3U8(string(body), streamURL, r.URL.Path)
	_, _ = io.WriteString(w, rewritten)
}

func (m *Manager) proxyManifest(w http.ResponseWriter, r *http.Request, streamURL string) {
	now := time.Now()
	if m.enableManifestBackoff {
		if cached, ok := m.getManifestBackoffCache(streamURL, now); ok {
			m.writeManifestFromCache(w, cached, "backoff")
			return
		}
	}
	if cached, ok := m.getManifestCache(streamURL, now, false); ok {
		m.writeManifestFromCache(w, cached, "fresh")
		return
	}

	ch, owner := m.beginManifestFetch(streamURL)
	if !owner {
		select {
		case <-ch:
		case <-r.Context().Done():
			return
		}
		if cached, ok := m.getManifestCache(streamURL, time.Now(), false); ok {
			m.writeManifestFromCache(w, cached, "fresh")
			return
		}
	}
	defer m.endManifestFetch(streamURL, ch, owner)

	resp, err := m.openUpstream(r.Context(), r, streamURL)
	if err != nil {
		if cached, ok := m.getManifestCache(streamURL, time.Now(), true); ok {
			logx.Debugf("upstream manifest fallback=stale err=%v url=%s", err, streamURL)
			m.writeManifestFromCache(w, cached, "stale")
			return
		}
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		if cached, ok := m.getManifestCache(streamURL, time.Now(), true); ok {
			logx.Debugf("upstream manifest read fallback=stale err=%v url=%s", readErr, streamURL)
			m.writeManifestFromCache(w, cached, "stale")
			return
		}
		http.Error(w, readErr.Error(), http.StatusBadGateway)
		return
	}

	if resp.StatusCode >= 400 {
		m.logBurst(fmt.Sprintf("upstream_status=%d", resp.StatusCode), streamURL)
	}

	isM3U8 := isM3U8Response(streamURL, resp.Header.Get("Content-Type"))
	if !isM3U8 {
		copyUpstreamHeaders(w, resp, false)
		w.WriteHeader(resp.StatusCode)
		_, _ = w.Write(body)
		return
	}

	rewritten := []byte(rewriteM3U8(string(body), streamURL, r.URL.Path))
	if resp.StatusCode >= http.StatusTooManyRequests || resp.StatusCode >= http.StatusInternalServerError {
		if cached, ok := m.getManifestCache(streamURL, time.Now(), true); ok {
			logx.Debugf("upstream manifest status=%d fallback=stale url=%s", resp.StatusCode, streamURL)
			m.writeManifestFromCache(w, cached, "stale")
			return
		}
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		prev, _ := m.getManifestEntry(streamURL)
		unchanged := 0
		if prev.status >= 200 && prev.status < 300 && bytes.Equal(prev.body, rewritten) {
			unchanged = prev.unchanged + 1
		}
		backoff := time.Time{}
		if m.enableManifestBackoff && unchanged >= 1 {
			target := parseTargetDurationSeconds(body)
			backoff = time.Now().Add(playlistBackoffDuration(target))
			logx.Debugf("upstream manifest unchanged=%d backoff=%s url=%s", unchanged+1, backoff.Sub(time.Now()).Round(100*time.Millisecond), streamURL)
		}
		m.setManifestCache(streamURL, manifestEntry{
			status:     resp.StatusCode,
			header:     cloneHeader(resp.Header),
			body:       rewritten,
			freshUntil: time.Now().Add(manifestFreshTTL),
			staleUntil: time.Now().Add(manifestFreshTTL + manifestStaleTTL),
			unchanged:  unchanged,
			backoffTo:  backoff,
		})
	}

	copyUpstreamHeaders(w, resp, true)
	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(rewritten)
}

func parseTargetDurationSeconds(body []byte) int {
	for _, line := range strings.Split(string(body), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#EXT-X-TARGETDURATION:") {
			v := strings.TrimSpace(strings.TrimPrefix(trimmed, "#EXT-X-TARGETDURATION:"))
			n, err := strconv.Atoi(v)
			if err == nil && n > 0 {
				return n
			}
			return 0
		}
	}
	return 0
}

func playlistBackoffDuration(targetDurationSec int) time.Duration {
	if targetDurationSec <= 0 {
		return 2 * time.Second
	}
	d := time.Duration(targetDurationSec) * time.Second / 2
	if d < time.Second {
		return time.Second
	}
	if d > 5*time.Second {
		return 5 * time.Second
	}
	return d
}

func cloneHeader(h http.Header) http.Header {
	out := make(http.Header, len(h))
	for k, vv := range h {
		cp := make([]string, len(vv))
		copy(cp, vv)
		out[k] = cp
	}
	return out
}

func (m *Manager) beginManifestFetch(streamURL string) (chan struct{}, bool) {
	m.manifestMu.Lock()
	defer m.manifestMu.Unlock()
	if ch, ok := m.inFlight[streamURL]; ok {
		return ch, false
	}
	ch := make(chan struct{})
	m.inFlight[streamURL] = ch
	return ch, true
}

func (m *Manager) endManifestFetch(streamURL string, ch chan struct{}, owner bool) {
	if !owner {
		return
	}
	m.manifestMu.Lock()
	if current, ok := m.inFlight[streamURL]; ok && current == ch {
		delete(m.inFlight, streamURL)
		close(ch)
	}
	m.manifestMu.Unlock()
}

func (m *Manager) setManifestCache(streamURL string, entry manifestEntry) {
	m.manifestMu.Lock()
	m.manifests[streamURL] = entry
	m.manifestMu.Unlock()
}

func (m *Manager) getManifestEntry(streamURL string) (manifestEntry, bool) {
	m.manifestMu.Lock()
	defer m.manifestMu.Unlock()
	entry, ok := m.manifests[streamURL]
	return entry, ok
}

func (m *Manager) getManifestCache(streamURL string, now time.Time, allowStale bool) (manifestEntry, bool) {
	m.manifestMu.Lock()
	defer m.manifestMu.Unlock()
	entry, ok := m.manifests[streamURL]
	if !ok {
		return manifestEntry{}, false
	}
	if now.Before(entry.freshUntil) {
		return entry, true
	}
	if allowStale && now.Before(entry.staleUntil) {
		return entry, true
	}
	return manifestEntry{}, false
}

func (m *Manager) getManifestBackoffCache(streamURL string, now time.Time) (manifestEntry, bool) {
	m.manifestMu.Lock()
	defer m.manifestMu.Unlock()
	entry, ok := m.manifests[streamURL]
	if !ok {
		return manifestEntry{}, false
	}
	if entry.backoffTo.IsZero() || !now.Before(entry.backoffTo) {
		return manifestEntry{}, false
	}
	if now.Before(entry.staleUntil) {
		return entry, true
	}
	return manifestEntry{}, false
}

func (m *Manager) writeManifestFromCache(w http.ResponseWriter, entry manifestEntry, cacheState string) {
	copyUpstreamHeaders(w, &http.Response{Header: entry.header}, true)
	w.Header().Set("X-Stream-Cache", cacheState)
	w.WriteHeader(entry.status)
	_, _ = w.Write(entry.body)
}

func (m *Manager) copyStreamWithRetry(w http.ResponseWriter, r *http.Request, streamURL string, initialResp *http.Response) {
	resp := initialResp
	retries := 0
	for {
		_, err := copyWithIdleTimeout(r.Context(), w, resp.Body, m.idleTimeout)
		_ = resp.Body.Close()
		if err == nil || errors.Is(err, io.EOF) {
			return
		}
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			if r.Context().Err() != nil {
				return
			}
		}
		if retries >= m.maxRetries {
			logx.Errorf("stream retry exhausted retries=%d url=%s err=%v", retries, streamURL, err)
			return
		}
		retries++
		logx.Debugf("stream stalled; reconnecting attempt=%d/%d url=%s err=%v", retries, m.maxRetries, streamURL, err)

		nextResp, nextErr := m.openUpstream(r.Context(), r, streamURL)
		if nextErr != nil {
			logx.Errorf("stream reconnect failed attempt=%d/%d url=%s err=%v", retries, m.maxRetries, streamURL, nextErr)
			time.Sleep(300 * time.Millisecond)
			continue
		}
		resp = nextResp
	}
}

func copyWithIdleTimeout(ctx context.Context, dst io.Writer, src io.ReadCloser, idle time.Duration) (int64, error) {
	ctxRead, cancel := context.WithCancel(ctx)
	defer cancel()
	timer := time.AfterFunc(idle, cancel)
	defer timer.Stop()

	buf := make([]byte, 32*1024)
	var total int64
	for {
		n, rerr := src.Read(buf)
		if n > 0 {
			timer.Reset(idle)
			wn, werr := dst.Write(buf[:n])
			total += int64(wn)
			if werr != nil {
				return total, werr
			}
			if wn != n {
				return total, io.ErrShortWrite
			}
		}
		if rerr != nil {
			if errors.Is(rerr, context.Canceled) && ctxRead.Err() != nil {
				return total, context.DeadlineExceeded
			}
			return total, rerr
		}
		if ctxRead.Err() != nil {
			return total, context.DeadlineExceeded
		}
	}
}

func (m *Manager) openUpstream(ctx context.Context, clientReq *http.Request, streamURL string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, streamURL, nil)
	if err != nil {
		return nil, err
	}
	applyUpstreamHeaders(req, clientReq, streamURL)
	return m.client.Do(req)
}

func copyUpstreamHeaders(w http.ResponseWriter, resp *http.Response, isM3U8 bool) {
	for k, vv := range resp.Header {
		if strings.EqualFold(k, "Content-Length") && isM3U8 {
			continue
		}
		if isM3U8 && (strings.EqualFold(k, "Cache-Control") || strings.EqualFold(k, "Expires") || strings.EqualFold(k, "Pragma")) {
			continue
		}
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	if isM3U8 {
		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
		// Guard against aggressive client reload loops on stale live playlists.
		w.Header().Set("Cache-Control", "public, max-age=2, stale-while-revalidate=2")
	}
}

func applyUpstreamHeaders(upReq, clientReq *http.Request, streamURL string) {
	copyHeader := func(name string) {
		if v := strings.TrimSpace(clientReq.Header.Get(name)); v != "" {
			upReq.Header.Set(name, v)
		}
	}
	copyHeader("User-Agent")
	copyHeader("Accept")
	copyHeader("Accept-Language")
	copyHeader("Range")

	u, err := url.Parse(streamURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return
	}
	base := u.Scheme + "://" + u.Host + "/"
	if upReq.Header.Get("Referer") == "" {
		upReq.Header.Set("Referer", base)
	}
	if upReq.Header.Get("Origin") == "" {
		upReq.Header.Set("Origin", u.Scheme+"://"+u.Host)
	}
}

func isM3U8Response(streamURL, contentType string) bool {
	u := strings.ToLower(streamURL)
	if strings.Contains(u, ".m3u8") {
		return true
	}
	ct := strings.ToLower(contentType)
	return strings.Contains(ct, "application/vnd.apple.mpegurl") ||
		strings.Contains(ct, "application/x-mpegurl")
}

func rewriteM3U8(content, baseURL, streamPath string) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "#") {
			lines[i] = rewriteTagURI(line, baseURL, streamPath)
			continue
		}
		lines[i] = buildProxySrcURI(baseURL, trimmed, streamPath)
	}
	return strings.Join(lines, "\n")
}

func rewriteTagURI(line, baseURL, streamPath string) string {
	const key = `URI="`
	idx := strings.Index(line, key)
	if idx == -1 {
		return line
	}
	start := idx + len(key)
	endRel := strings.Index(line[start:], `"`)
	if endRel == -1 {
		return line
	}
	end := start + endRel
	uri := line[start:end]
	rewritten := buildProxySrcURI(baseURL, uri, streamPath)
	return line[:start] + rewritten + line[end:]
}

func buildProxySrcURI(baseURL, ref, streamPath string) string {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return ref
	}
	abs := ref
	if u, err := url.Parse(ref); err == nil && (!u.IsAbs() || u.Host == "") {
		if b, berr := url.Parse(baseURL); berr == nil {
			abs = b.ResolveReference(u).String()
		}
	}
	return streamPath + "?src=" + url.QueryEscape(abs)
}

func (m *Manager) transcode(w http.ResponseWriter, r *http.Request, streamURL string) {
	// Emit fragmented MP4 for broad HTML5 video compatibility.
	w.Header().Set("Content-Type", "video/mp4")
	ctx := r.Context()
	audioTrack := parseAudioTrackIndex(r)
	forceAudioTranscode := parseAudioFallbackEnabled(r)
	plan, ok := m.buildTranscodePlan(streamURL, audioTrack)
	if !ok {
		// Safe fallback when probing fails: keep previous full transcode behavior.
		plan = transcodePlan{videoCopy: false, audioCopy: false}
	}
	if isDeinterlaceEnabled(r) {
		// Video filters require decode+encode, so bypass copy mode for video.
		plan.videoCopy = false
	}
	if forceAudioTranscode {
		plan.audioCopy = false
	}
	args := []string{
		"-hide_banner",
		"-loglevel", "error",
		// Be tolerant to occasional corrupt frames in unstable IPTV feeds.
		"-fflags", "+discardcorrupt+genpts",
		"-err_detect", "ignore_err",
		"-i", streamURL,
		"-map", "0:v:0",
	}
	if audioTrack >= 0 {
		args = append(args, "-map", fmt.Sprintf("0:a:%d?", audioTrack))
	} else {
		// Let ffmpeg pick the most suitable audio stream when explicit selection is absent.
		args = append(args, "-map", "0:a?")
	}
	args = append(args,
		"-sn",
		"-dn",
	)
	if isDeinterlaceEnabled(r) {
		// YADIF is fast enough for live playback and removes interlacing artifacts.
		args = append(args, "-vf", "yadif=0:-1:0")
	}
	videoCodecArg := "libx264"
	if plan.videoCopy {
		videoCodecArg = "copy"
	}
	audioCodecArg := "aac"
	if plan.audioCopy {
		audioCodecArg = "copy"
	}
	args = append(args,
		"-c:v", videoCodecArg,
		"-preset", "veryfast",
		"-tune", "zerolatency",
		"-pix_fmt", "yuv420p",
		"-c:a", audioCodecArg,
		"-ac", "2",
		"-ar", "48000",
		"-movflags", "+frag_keyframe+empty_moov+default_base_moof",
		"-f", "mp4",
		"pipe:1",
	)
	if plan.videoCopy {
		// Encoder-only options are invalid for stream copy.
		args = removeArgPair(args, "-preset")
		args = removeArgPair(args, "-tune")
		args = removeArgPair(args, "-pix_fmt")
	}
	if plan.audioCopy {
		// Resample/downmix options are invalid for stream copy.
		args = removeArgPair(args, "-ac")
		args = removeArgPair(args, "-ar")
	}
	logx.Debugf("transcode plan video=%s(%s) audio=%s(%s) url=%s",
		map[bool]string{true: "copy", false: "encode"}[plan.videoCopy], plan.videoCodec,
		map[bool]string{true: "copy", false: "encode"}[plan.audioCopy], plan.audioCodec,
		streamURL,
	)
	cmd := exec.CommandContext(ctx, m.ffmpegBin, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := cmd.Start(); err != nil {
		if stderr.Len() > 0 {
			logx.Errorf("ffmpeg start failed: %s", stderr.String())
		}
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	_, copyErr := io.Copy(w, stdout)
	if copyErr != nil && cmd.Process != nil {
		_ = cmd.Process.Kill()
	}
	_ = stdout.Close()
	waitDone := make(chan error, 1)
	go func() { waitDone <- cmd.Wait() }()
	select {
	case <-waitDone:
	case <-time.After(2 * time.Second):
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		select {
		case <-waitDone:
		case <-time.After(1500 * time.Millisecond):
		}
	}
	if stderr.Len() > 0 {
		msg := strings.TrimSpace(stderr.String())
		if copyErr != nil || !isBenignFFmpegWarning(msg) {
			if copyErr != nil {
				logx.Errorf("ffmpeg stream ended err=%v detail=%s", copyErr, msg)
			} else {
				logx.Debugf("ffmpeg stream ended err=%v detail=%s", copyErr, msg)
			}
		}
	}
}

func removeArgPair(args []string, key string) []string {
	out := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		if args[i] == key {
			i++
			continue
		}
		out = append(out, args[i])
	}
	return out
}

func (m *Manager) buildTranscodePlan(streamURL string, audioTrack int) (transcodePlan, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, m.ffmpegBin[:len(m.ffmpegBin)-len("ffmpeg")]+"ffprobe",
		"-v", "error",
		"-show_entries", "stream=codec_name,codec_type",
		"-of", "json",
		streamURL,
	)
	if strings.TrimSpace(m.ffmpegBin) == "ffmpeg" || !strings.HasSuffix(strings.TrimSpace(m.ffmpegBin), "/ffmpeg") {
		cmd = exec.CommandContext(ctx, "ffprobe",
			"-v", "error",
			"-show_entries", "stream=codec_name,codec_type",
			"-of", "json",
			streamURL,
		)
	}
	out, err := cmd.Output()
	if err != nil {
		return transcodePlan{}, false
	}
	var payload struct {
		Streams []struct {
			CodecName string `json:"codec_name"`
			CodecType string `json:"codec_type"`
		} `json:"streams"`
	}
	if err := json.Unmarshal(out, &payload); err != nil {
		return transcodePlan{}, false
	}
	if len(payload.Streams) == 0 {
		return transcodePlan{}, false
	}

	videoCodec := ""
	audioCodecs := make([]string, 0, 4)
	for _, s := range payload.Streams {
		ct := strings.ToLower(strings.TrimSpace(s.CodecType))
		cn := strings.ToLower(strings.TrimSpace(s.CodecName))
		switch ct {
		case "video":
			if videoCodec == "" {
				videoCodec = cn
			}
		case "audio":
			audioCodecs = append(audioCodecs, cn)
		}
	}
	if videoCodec == "" {
		videoCodec = "unknown"
	}
	audioCodec := "unknown"
	if len(audioCodecs) > 0 {
		if audioTrack >= 0 && audioTrack < len(audioCodecs) {
			audioCodec = audioCodecs[audioTrack]
		} else {
			audioCodec = audioCodecs[0]
		}
	}

	videoNeedsTranscode := isVideoCodecBrowserUnsafe(videoCodec)
	audioNeedsTranscode := isAudioCodecBrowserUnsafe(audioCodec)

	return transcodePlan{
		videoCodec: videoCodec,
		audioCodec: audioCodec,
		videoCopy:  !videoNeedsTranscode,
		audioCopy:  !audioNeedsTranscode,
	}, true
}

func isVideoCodecBrowserUnsafe(codec string) bool {
	switch strings.ToLower(strings.TrimSpace(codec)) {
	case "hevc", "h265", "mpeg2video", "mpeg4":
		return true
	default:
		return false
	}
}

func isAudioCodecBrowserUnsafe(codec string) bool {
	switch strings.ToLower(strings.TrimSpace(codec)) {
	case "ac3", "eac3", "a52", "mp2", "mpga":
		return true
	default:
		return false
	}
}

func (m *Manager) DescribePathPlan(streamURL string, audioTrack int, deinterlace bool, forceAudioTranscode bool, mode Mode) PathPlan {
	if mode == ModeDirect {
		return PathPlan{
			VideoPath:  "bypass",
			AudioPath:  "bypass",
			VideoCodec: "-",
			AudioCodec: "-",
		}
	}
	plan, ok := m.buildTranscodePlan(streamURL, audioTrack)
	if !ok {
		return PathPlan{
			VideoPath:  "ffmpeg",
			AudioPath:  "ffmpeg",
			VideoCodec: "unknown",
			AudioCodec: "unknown",
		}
	}
	videoCopy := plan.videoCopy && !deinterlace
	audioCopy := plan.audioCopy && !forceAudioTranscode
	videoPath := "ffmpeg"
	if videoCopy {
		videoPath = "bypass"
	}
	audioPath := "ffmpeg"
	if audioCopy {
		audioPath = "bypass"
	}
	return PathPlan{
		VideoPath:  videoPath,
		AudioPath:  audioPath,
		VideoCodec: plan.videoCodec,
		AudioCodec: plan.audioCodec,
	}
}

func parseAudioFallbackEnabled(r *http.Request) bool {
	return strings.TrimSpace(r.URL.Query().Get("audio_fallback")) == "1"
}

func parseAudioTrackIndex(r *http.Request) int {
	raw := strings.TrimSpace(r.URL.Query().Get("audio"))
	if raw == "" {
		return -1
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return -1
	}
	return n
}

func isDeinterlaceEnabled(r *http.Request) bool {
	raw := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("deinterlace")))
	return raw == "1" || raw == "true" || raw == "on" || raw == "yes"
}

func isBenignFFmpegWarning(msg string) bool {
	if msg == "" {
		return true
	}
	lower := strings.ToLower(msg)
	// Known noisy decoder warnings on damaged live inputs; transcoding can continue.
	return strings.Contains(lower, "mmco: unref short failure") ||
		strings.Contains(lower, "number of reference frames")
}

func (m *Manager) logBurst(event, rawURL string) {
	keyURL := normalizeLogURL(rawURL)
	key := event + "|" + keyURL
	now := time.Now()

	m.logMu.Lock()
	entry := m.logStats[key]
	if entry.since.IsZero() {
		entry.since = now
	}
	entry.count++
	elapsed := now.Sub(entry.since)
	if elapsed < logWindow {
		m.logStats[key] = entry
		m.logMu.Unlock()
		return
	}
	count := entry.count
	entry = logEntry{since: now}
	m.logStats[key] = entry
	m.logMu.Unlock()

	logx.Debugf("stream burst event=%s url=%s hits=%d window=%s", event, keyURL, count, elapsed.Round(100*time.Millisecond))
}

func normalizeLogURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	p := u.Path
	if strings.HasSuffix(strings.ToLower(p), ".ts") {
		dir := path.Dir(p)
		p = path.Join(dir, "*.ts")
	}
	u.Path = p
	u.RawQuery = ""
	u.Fragment = ""
	return u.String()
}
