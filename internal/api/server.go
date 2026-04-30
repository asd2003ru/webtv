package api

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"webtv/internal/logx"
	"webtv/internal/model"
	"webtv/internal/scheduler"
	"webtv/internal/storage"
	"webtv/internal/stream"
)

var lastStreamChannelByClient sync.Map
var streamInflight sync.Map

type Server struct {
	mux                *http.ServeMux
	store              *storage.Store
	syncer             *scheduler.SyncService
	stream             *stream.Manager
	streamModeCacheTTL time.Duration
	appTitle           string
}

type hiddenPrefsPayload struct {
	Groups   []string `json:"groups"`
	Channels []string `json:"channels"`
}

func New(store *storage.Store, syncer *scheduler.SyncService, streamMgr *stream.Manager, staticFS http.Handler, streamModeCacheTTL time.Duration) *Server {
	return NewWithAppTitle(store, syncer, streamMgr, staticFS, streamModeCacheTTL, "WebTV")
}

func NewWithAppTitle(store *storage.Store, syncer *scheduler.SyncService, streamMgr *stream.Manager, staticFS http.Handler, streamModeCacheTTL time.Duration, appTitle string) *Server {
	if streamModeCacheTTL <= 0 {
		streamModeCacheTTL = 6 * time.Hour
	}
	if strings.TrimSpace(appTitle) == "" {
		appTitle = "WebTV"
	}
	s := &Server{
		mux:                http.NewServeMux(),
		store:              store,
		syncer:             syncer,
		stream:             streamMgr,
		streamModeCacheTTL: streamModeCacheTTL,
		appTitle:           appTitle,
	}
	s.routes(staticFS)
	return s
}

func (s *Server) Handler() http.Handler { return logging(s.mux) }

func (s *Server) routes(staticFS http.Handler) {
	s.mux.HandleFunc("/healthz", s.health)
	s.mux.HandleFunc("/readyz", s.ready)
	s.mux.HandleFunc("/api/config", s.uiConfig)
	s.mux.HandleFunc("/api/playlists", s.playlists)
	s.mux.HandleFunc("/api/playlists/", s.playlistsSub)
	s.mux.HandleFunc("/api/channels/", s.channelsSub)
	s.mux.Handle("/", staticFS)
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) ready(w http.ResponseWriter, r *http.Request) {
	if err := s.store.Ping(r.Context()); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "db_down"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func (s *Server) uiConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"app_title": s.appTitle,
	})
}

func (s *Server) playlists(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		items, err := s.store.ListPlaylists(r.Context())
		if err != nil {
			writeErr(w, err, 500)
			return
		}
		writeJSON(w, 200, items)
	case http.MethodPost:
		var p model.Playlist
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			writeErr(w, err, 400)
			return
		}
		if p.UpdateIntervalMinutes <= 0 {
			p.UpdateIntervalMinutes = 1440
		}
		if !p.Enabled {
			p.Enabled = true
		}
		created, err := s.store.CreatePlaylist(r.Context(), p)
		if err != nil {
			writeErr(w, err, 500)
			return
		}
		logx.Infof("playlist created id=%d name=%q enabled=%t interval_min=%d", created.ID, created.Name, created.Enabled, created.UpdateIntervalMinutes)
		writeJSON(w, 201, created)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) playlistsSub(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/playlists/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		w.WriteHeader(404)
		return
	}
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		writeErr(w, err, 400)
		return
	}
	if len(parts) == 1 {
		switch r.Method {
		case http.MethodPut:
			var p model.Playlist
			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				writeErr(w, err, 400)
				return
			}
			if err := s.store.UpdatePlaylist(r.Context(), id, p); err != nil {
				writeErr(w, err, 500)
				return
			}
			logx.Infof("playlist updated id=%d", id)
			w.WriteHeader(204)
		case http.MethodDelete:
			if err := s.store.DeletePlaylist(r.Context(), id); err != nil {
				writeErr(w, err, 500)
				return
			}
			logx.Infof("playlist deleted id=%d", id)
			w.WriteHeader(204)
		default:
			w.WriteHeader(405)
		}
		return
	}
	switch parts[1] {
	case "refresh":
		if r.Method != http.MethodPost {
			w.WriteHeader(405)
			return
		}
		p, err := s.store.GetPlaylist(r.Context(), id)
		if err != nil {
			writeErr(w, err, 404)
			return
		}
		if err := s.syncer.SyncPlaylist(r.Context(), p); err != nil {
			writeErr(w, err, 502)
			return
		}
		logx.Infof("playlist refresh requested id=%d", id)
		w.WriteHeader(202)
	case "channels":
		if r.Method != http.MethodGet {
			w.WriteHeader(405)
			return
		}
		limit := 0
		offset := 0
		if v := strings.TrimSpace(r.URL.Query().Get("limit")); v != "" {
			n, parseErr := strconv.Atoi(v)
			if parseErr != nil || n < 0 {
				writeErr(w, errors.New("invalid limit"), 400)
				return
			}
			limit = n
		}
		if v := strings.TrimSpace(r.URL.Query().Get("offset")); v != "" {
			n, parseErr := strconv.Atoi(v)
			if parseErr != nil || n < 0 {
				writeErr(w, errors.New("invalid offset"), 400)
				return
			}
			offset = n
		}
		chs, err := s.store.ListChannelsByPlaylistPage(r.Context(), id, limit, offset)
		if err != nil {
			writeErr(w, err, 500)
			return
		}
		writeJSON(w, 200, chs)
	case "now-programs":
		if r.Method != http.MethodGet {
			w.WriteHeader(405)
			return
		}
		items, err := s.store.GetNowProgramsByPlaylist(r.Context(), id, time.Now().UTC())
		if err != nil {
			writeErr(w, err, 500)
			return
		}
		writeJSON(w, 200, items)
	case "hidden":
		switch r.Method {
		case http.MethodGet:
			groups, channels, err := s.store.GetHiddenPrefsByPlaylist(r.Context(), id)
			if err != nil {
				writeErr(w, err, 500)
				return
			}
			writeJSON(w, 200, hiddenPrefsPayload{
				Groups:   groups,
				Channels: channels,
			})
		case http.MethodPut:
			var payload hiddenPrefsPayload
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				writeErr(w, err, 400)
				return
			}
			if err := s.store.ReplaceHiddenPrefsByPlaylist(r.Context(), id, payload.Groups, payload.Channels); err != nil {
				writeErr(w, err, 500)
				return
			}
			logx.Infof("playlist hidden prefs updated id=%d groups=%d channels=%d", id, len(payload.Groups), len(payload.Channels))
			w.WriteHeader(204)
		default:
			w.WriteHeader(405)
		}
	default:
		w.WriteHeader(404)
	}
}

func (s *Server) channelsSub(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/channels/"), "/")
	if len(parts) < 2 {
		w.WriteHeader(404)
		return
	}
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		writeErr(w, err, 400)
		return
	}
	switch parts[1] {
	case "audio-tracks":
		if r.Method != http.MethodGet {
			w.WriteHeader(405)
			return
		}
		ch, err := s.store.GetChannel(r.Context(), id)
		if err != nil {
			writeErr(w, err, 404)
			return
		}
		streamURL := ch.StreamURL
		start, end, ok := parseArchiveWindow(r)
		if ok && ch.ArchiveSupported {
			streamURL = stream.BuildArchiveURL(ch.StreamURL, start, end)
		}
		tracks, ok := s.stream.ProbeAudioTracks(streamURL)
		if !ok {
			writeJSON(w, 200, []stream.AudioTrack{})
			return
		}
		writeJSON(w, 200, tracks)
	case "epg":
		if r.Method != http.MethodGet {
			w.WriteHeader(405)
			return
		}
		from, to := parseWindow(r)
		items, err := s.store.GetProgramsByWindow(r.Context(), id, from, to)
		if err != nil {
			writeErr(w, err, 500)
			return
		}
		writeJSON(w, 200, items)
	case "stream":
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.WriteHeader(405)
			return
		}
		if src := strings.TrimSpace(r.URL.Query().Get("src")); src != "" {
			u, err := url.Parse(src)
			if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
				writeErr(w, errors.New("invalid stream src"), 400)
				return
			}
			if r.Method == http.MethodHead {
				mode := stream.BrowserMode(src)
				w.Header().Set("X-Stream-Mode", string(mode))
				setStreamPathPlanHeaders(w, s.stream.DescribePathPlan(src, -1, isDeinterlaceEnabled(r), false, mode))
				w.WriteHeader(http.StatusOK)
				return
			}
			// Nested HLS requests should not trigger codec probing repeatedly.
			s.stream.HandleWithMode(w, r, src, stream.BrowserMode(src))
			return
		}
		ch, err := s.store.GetChannel(r.Context(), id)
		if err != nil {
			writeErr(w, err, 404)
			return
		}
		streamURL := ch.StreamURL
		start, end, ok := parseArchiveWindow(r)
		if ok && ch.ArchiveSupported {
			streamURL = stream.BuildArchiveURL(ch.StreamURL, start, end)
		}
		mode := s.resolveChannelStreamMode(r.Context(), r, ch, streamURL)
		if strings.TrimSpace(r.URL.Query().Get("audio_fallback")) == "1" {
			mode = stream.ModeTranscode
		}
		if r.Method == http.MethodHead {
			w.Header().Set("X-Stream-Mode", string(mode))
			audioTrack := parseAudioTrackIndex(r)
			setStreamPathPlanHeaders(w, s.stream.DescribePathPlan(streamURL, audioTrack, isDeinterlaceEnabled(r), isAudioFallbackEnabled(r), mode))
			w.WriteHeader(http.StatusOK)
			return
		}
		if isStreamChannelSwitch(r, strconv.FormatInt(id, 10)) {
			playlistName := ""
			if p, err := s.store.GetPlaylist(r.Context(), ch.PlaylistID); err == nil {
				playlistName = p.Name
			}
			if playlistName == "" {
				playlistName = "unknown"
			}
			logx.Infof("channel switched playlist=%q(%d) channel=%q(%d)", playlistName, ch.PlaylistID, ch.Name, ch.ID)
		}
		s.stream.HandleWithMode(w, r, streamURL, mode)
	default:
		w.WriteHeader(404)
	}
}

func setStreamPathPlanHeaders(w http.ResponseWriter, plan stream.PathPlan) {
	w.Header().Set("X-Video-Path", plan.VideoPath)
	w.Header().Set("X-Audio-Path", plan.AudioPath)
	w.Header().Set("X-Video-Codec", plan.VideoCodec)
	w.Header().Set("X-Audio-Codec", plan.AudioCodec)
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

func isAudioFallbackEnabled(r *http.Request) bool {
	return strings.TrimSpace(r.URL.Query().Get("audio_fallback")) == "1"
}

func (s *Server) resolveChannelStreamMode(ctx context.Context, r *http.Request, ch model.Channel, streamURL string) stream.Mode {
	fallback := stream.BrowserMode(streamURL)
	if fallback == stream.ModeTranscode {
		return fallback
	}
	if ch.StreamModeCache != "" && ch.StreamModeAt != nil && time.Since(*ch.StreamModeAt) <= s.streamModeCacheTTL {
		if ch.StreamModeCache == string(stream.ModeTranscode) {
			return stream.ModeTranscode
		}
		return stream.ModeDirect
	}
	mode := s.stream.DetectMode(r, streamURL)
	if err := s.store.SetChannelStreamModeCache(ctx, ch.ID, string(mode), time.Now().UTC()); err != nil {
		logx.Errorf("set stream mode cache channel=%d: %v", ch.ID, err)
	}
	return mode
}

func parseWindow(r *http.Request) (time.Time, time.Time) {
	from := time.Now().UTC().Add(-1 * time.Hour)
	to := time.Now().UTC().Add(6 * time.Hour)
	if v := r.URL.Query().Get("from"); v != "" {
		if p, err := time.Parse(time.RFC3339, v); err == nil {
			from = p
		}
	}
	if v := r.URL.Query().Get("to"); v != "" {
		if p, err := time.Parse(time.RFC3339, v); err == nil {
			to = p
		}
	}
	return from, to
}

func parseArchiveWindow(r *http.Request) (time.Time, time.Time, bool) {
	startRaw := r.URL.Query().Get("start")
	endRaw := r.URL.Query().Get("end")
	if startRaw == "" || endRaw == "" {
		return time.Time{}, time.Time{}, false
	}
	start, err := time.Parse(time.RFC3339, startRaw)
	if err != nil {
		return time.Time{}, time.Time{}, false
	}
	end, err := time.Parse(time.RFC3339, endRaw)
	if err != nil {
		return time.Time{}, time.Time{}, false
	}
	if !end.After(start) {
		return time.Time{}, time.Time{}, false
	}
	return start, end, true
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, err error, code int) {
	if errors.Is(err, context.Canceled) {
		code = 499
	}
	writeJSON(w, code, map[string]string{"error": err.Error()})
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lw := &logResponseWriter{ResponseWriter: w, status: http.StatusOK}
		key, track := streamRequestKey(r)
		var inflight int64
		if track {
			inflight = addInflight(key, 1)
			defer addInflight(key, -1)
		}

		next.ServeHTTP(lw, r)
		if lw.status >= http.StatusBadRequest {
			logRequest("error", r, lw, start, inflight, track)
			return
		}
		if logx.Enabled(logx.LevelDebug) {
			logRequest("debug", r, lw, start, inflight, track)
			return
		}
		if !shouldLogRequestInfo(r) {
			return
		}
		logRequest("info", r, lw, start, inflight, track)
	})
}

func logRequest(level string, r *http.Request, lw *logResponseWriter, start time.Time, inflight int64, track bool) {
	if track {
		msg := "%s %s status=%d bytes=%d inflight=%d canceled=%t in %s"
		args := []any{r.Method, r.URL.RequestURI(), lw.status, lw.bytes, inflight, errors.Is(r.Context().Err(), context.Canceled), time.Since(start)}
		switch level {
		case "error":
			logx.Errorf(msg, args...)
		case "debug":
			logx.Debugf(msg, args...)
		default:
			logx.Infof(msg, args...)
		}
		return
	}
	msg := "%s %s status=%d bytes=%d in %s"
	args := []any{r.Method, r.URL.Path, lw.status, lw.bytes, time.Since(start)}
	switch level {
	case "error":
		logx.Errorf(msg, args...)
	case "debug":
		logx.Debugf(msg, args...)
	default:
		logx.Infof(msg, args...)
	}
}

type logResponseWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (w *logResponseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *logResponseWriter) Write(p []byte) (int, error) {
	n, err := w.ResponseWriter.Write(p)
	w.bytes += n
	return n, err
}

func streamRequestKey(r *http.Request) (string, bool) {
	if !strings.HasPrefix(r.URL.Path, "/api/channels/") || !strings.HasSuffix(r.URL.Path, "/stream") {
		return "", false
	}
	client := clientAddrKey(r.RemoteAddr)
	srcType := "base"
	if strings.TrimSpace(r.URL.Query().Get("src")) != "" {
		srcType = "src"
	}
	return client + "|" + r.URL.Path + "|" + srcType + "|" + r.Method, true
}

func addInflight(key string, delta int64) int64 {
	v, _ := streamInflight.LoadOrStore(key, new(int64))
	ptr := v.(*int64)
	return atomic.AddInt64(ptr, delta)
}

func shouldLogRequestInfo(r *http.Request) bool {
	if r.Method != http.MethodGet {
		return false
	}
	channelID, ok := streamChannelIDFromPath(r.URL.Path)
	if !ok {
		return false
	}
	return isStreamChannelSwitch(r, channelID)
}

func isStreamChannelSwitch(r *http.Request, channelID string) bool {
	client := clientAddrKey(r.RemoteAddr)
	last, found := lastStreamChannelByClient.Load(client)
	if found {
		if prev, ok := last.(string); ok && prev == channelID {
			return false
		}
	}
	lastStreamChannelByClient.Store(client, channelID)
	return true
}

func streamChannelIDFromPath(path string) (string, bool) {
	parts := strings.Split(strings.TrimPrefix(path, "/api/channels/"), "/")
	if len(parts) != 2 || parts[1] != "stream" || parts[0] == "" {
		return "", false
	}
	return parts[0], true
}

func clientAddrKey(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err == nil && host != "" {
		return host
	}
	return remoteAddr
}
