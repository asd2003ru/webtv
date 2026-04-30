package scheduler

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/asd2003ru/webtv/internal/epg"
	"github.com/asd2003ru/webtv/internal/logx"
	"github.com/asd2003ru/webtv/internal/model"
	"github.com/asd2003ru/webtv/internal/playlist"
	"github.com/asd2003ru/webtv/internal/storage"
)

type SyncService struct {
	store  *storage.Store
	client *http.Client
}

func NewSyncService(store *storage.Store) *SyncService {
	return &SyncService{store: store, client: &http.Client{Timeout: 45 * time.Second}}
}

func (s *SyncService) SyncPlaylist(ctx context.Context, p model.Playlist) error {
	logx.Infof("playlist sync started id=%d name=%q", p.ID, p.Name)
	channels, err := s.fetchChannels(ctx, p)
	if err != nil {
		_ = s.store.SetPlaylistSyncStatus(ctx, p.ID, err.Error())
		logx.Errorf("playlist sync fetch channels failed id=%d: %v", p.ID, err)
		return err
	}
	if err := s.store.ReplaceChannels(ctx, p.ID, channels); err != nil {
		_ = s.store.SetPlaylistSyncStatus(ctx, p.ID, err.Error())
		logx.Errorf("playlist sync replace channels failed id=%d: %v", p.ID, err)
		return err
	}
	storedChannels, err := s.store.ListChannelsByPlaylist(ctx, p.ID)
	if err != nil {
		_ = s.store.SetPlaylistSyncStatus(ctx, p.ID, err.Error())
		logx.Errorf("playlist sync list stored channels failed id=%d: %v", p.ID, err)
		return err
	}
	if err := s.syncEPG(ctx, p, storedChannels); err != nil {
		_ = s.store.SetPlaylistSyncStatus(ctx, p.ID, err.Error())
		logx.Errorf("playlist sync epg failed id=%d: %v", p.ID, err)
		return err
	}
	if err := s.store.SetPlaylistSyncStatus(ctx, p.ID, ""); err != nil {
		logx.Errorf("playlist sync set status failed id=%d: %v", p.ID, err)
		return err
	}
	logx.Infof("playlist sync completed id=%d channels=%d", p.ID, len(storedChannels))
	return nil
}

func (s *SyncService) fetchChannels(ctx context.Context, p model.Playlist) ([]model.Channel, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, p.M3UURL, nil)
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	parsed, err := playlist.ParseM3U(resp.Body)
	if err != nil {
		return nil, err
	}
	if len(parsed) == 0 {
		return nil, fmt.Errorf("empty playlist: fetched 0 channels from %s", p.M3UURL)
	}
	return playlist.ToModel(p.ID, parsed), nil
}

func (s *SyncService) syncEPG(ctx context.Context, p model.Playlist, channels []model.Channel) error {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, p.EPGURL, nil)
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	programs, err := epg.ParseXMLTVGZ(resp.Body)
	if err != nil {
		return err
	}

	epgProgramsByID := make(map[string][]epg.Program)
	epgIDByNormID := make(map[string]string)
	epgIDsByNameKey := make(map[string][]string)
	epgIconByID := make(map[string]string)
	epgNameByID := make(map[string]string)
	for _, prog := range programs {
		epgProgramsByID[prog.ChannelID] = append(epgProgramsByID[prog.ChannelID], prog)
		if k := normalizeMatchKey(prog.ChannelID); k != "" {
			if _, ok := epgIDByNormID[k]; !ok {
				epgIDByNormID[k] = prog.ChannelID
			}
		}
		if _, ok := epgNameByID[prog.ChannelID]; !ok {
			epgNameByID[prog.ChannelID] = strings.TrimSpace(prog.ChannelName)
		}
		if strings.TrimSpace(prog.ChannelIcon) != "" {
			if _, ok := epgIconByID[prog.ChannelID]; !ok {
				epgIconByID[prog.ChannelID] = strings.TrimSpace(prog.ChannelIcon)
			}
		}
		for _, nm := range append([]string{prog.ChannelName}, prog.ChannelAltNames...) {
			for _, key := range []string{
				normalizeMatchKey(nm),
				normalizeMatchKey(strings.ReplaceAll(nm, "_", " ")),
				baseNameKey(nm),
				donorNameKey(nm),
			} {
				if key == "" {
					continue
				}
				epgIDsByNameKey[key] = appendUniqueString(epgIDsByNameKey[key], prog.ChannelID)
			}
		}
	}

	chPrograms := map[int64][]model.Program{}
	channelLogos := map[int64]string{}
	var matchedByID, matchedByNormID, matchedByName, matchedByRelaxedName, matchedByBaseName int
	for _, ch := range channels {
		epgID, kind := resolveEPGChannelID(ch, epgProgramsByID, epgIDByNormID, epgIDsByNameKey, epgNameByID)
		if epgID == "" {
			continue
		}
		switch kind {
		case "id":
			matchedByID++
		case "norm_id":
			matchedByNormID++
		case "name":
			matchedByName++
		case "relaxed":
			matchedByRelaxedName++
		case "base":
			matchedByBaseName++
		}
		items := epgProgramsByID[epgID]
		out := make([]model.Program, 0, len(items))
		for _, it := range items {
			out = append(out, model.Program{
				PlaylistID:  p.ID,
				ChannelID:   ch.ID,
				Title:       it.Title,
				StartAt:     it.StartAt,
				EndAt:       it.EndAt,
				Description: it.Description,
			})
		}
		chPrograms[ch.ID] = out
		if strings.TrimSpace(ch.Logo) == "" {
			if logo := strings.TrimSpace(epgIconByID[epgID]); logo != "" {
				channelLogos[ch.ID] = logo
			}
		}
	}

	for chID, items := range chPrograms {
		if err := s.store.ReplacePrograms(ctx, chID, items); err != nil {
			return err
		}
	}
	for chID, logo := range channelLogos {
		if err := s.store.SetChannelLogo(ctx, chID, logo); err != nil {
			return err
		}
	}
	logx.Infof("sync playlist %d EPG matched: by_id=%d by_norm_id=%d by_name=%d by_relaxed_name=%d by_base_name=%d total_programmes=%d",
		p.ID, matchedByID, matchedByNormID, matchedByName, matchedByRelaxedName, matchedByBaseName, len(programs))
	return nil
}

func resolveEPGChannelID(
	ch model.Channel,
	epgProgramsByID map[string][]epg.Program,
	epgIDByNormID map[string]string,
	epgIDsByNameKey map[string][]string,
	epgNameByID map[string]string,
) (string, string) {
	ext := strings.TrimSpace(ch.ExternalID)
	if ext != "" {
		if _, ok := epgProgramsByID[ext]; ok {
			return ext, "id"
		}
		if id, ok := epgIDByNormID[normalizeMatchKey(ext)]; ok {
			return id, "norm_id"
		}
		for _, key := range []string{
			normalizeMatchKey(ext),
			normalizeMatchKey(strings.ReplaceAll(ext, "_", " ")),
		} {
			if key == "" {
				continue
			}
			if ids := epgIDsByNameKey[key]; len(ids) > 0 {
				return pickBestEPGID(ids, ch.Name, epgNameByID), "name"
			}
		}
	}
	if key := normalizeMatchKey(ch.Name); key != "" {
		if ids := epgIDsByNameKey[key]; len(ids) > 0 {
			return pickBestEPGID(ids, ch.Name, epgNameByID), "name"
		}
	}
	if key := normalizeRelaxedChannelName(ch.Name); key != "" {
		if ids := epgIDsByNameKey[key]; len(ids) > 0 {
			return pickBestEPGID(ids, ch.Name, epgNameByID), "relaxed"
		}
	}
	if key := donorNameKey(ch.Name); key != "" {
		if ids := epgIDsByNameKey[key]; len(ids) > 0 {
			return pickBestEPGID(ids, ch.Name, epgNameByID), "base"
		}
	}
	return "", ""
}

func pickBestEPGID(ids []string, channelName string, epgNameByID map[string]string) string {
	if len(ids) == 1 {
		return ids[0]
	}
	target := normalizeMatchKey(channelName)
	base := donorNameKey(channelName)
	best := ids[0]
	bestScore := -1
	for _, id := range ids {
		score := 0
		nm := normalizeMatchKey(epgNameByID[id])
		if nm == target {
			score += 3
		}
		if base != "" && donorNameKey(epgNameByID[id]) == base {
			score += 2
		}
		if score > bestScore {
			bestScore = score
			best = id
		}
	}
	return best
}

func appendUniqueString(list []string, v string) []string {
	for _, it := range list {
		if it == v {
			return list
		}
	}
	return append(list, v)
}

func normalizeMatchKey(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	if v == "" {
		return ""
	}
	var b strings.Builder
	b.Grow(len(v))
	lastSpace := false
	for _, r := range v {
		isLetter := (r >= 'a' && r <= 'z') || (r >= 'а' && r <= 'я') || r == 'ё'
		isDigit := r >= '0' && r <= '9'
		if isLetter || isDigit {
			b.WriteRune(r)
			lastSpace = false
			continue
		}
		if !lastSpace {
			b.WriteByte(' ')
			lastSpace = true
		}
	}
	out := strings.TrimSpace(b.String())
	out = strings.Join(strings.Fields(out), " ")
	return out
}

func normalizeRelaxedChannelName(v string) string {
	k := normalizeMatchKey(v)
	if k == "" {
		return ""
	}
	noise := map[string]struct{}{
		"tv": {}, "канал": {}, "телеканал": {},
		"hd": {}, "fhd": {}, "uhd": {}, "sd": {}, "4k": {},
		"hevc": {}, "h264": {}, "h265": {},
	}
	parts := strings.Fields(k)
	cleaned := make([]string, 0, len(parts))
	for _, p := range parts {
		if _, isNoise := noise[p]; isNoise {
			continue
		}
		if strings.HasPrefix(p, "+") {
			if len(p) > 1 && allDigits(p[1:]) {
				continue
			}
		}
		if allDigits(p) {
			cleaned = append(cleaned, p)
			continue
		}
		cleaned = append(cleaned, p)
	}
	return strings.Join(cleaned, " ")
}

func allDigits(v string) bool {
	if v == "" {
		return false
	}
	for _, r := range v {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func pickBestChannelCandidate(candidates []int64, channelByID map[int64]model.Channel, channelName string, altNames []string) int64 {
	if len(candidates) == 1 {
		return candidates[0]
	}
	epgTokens := qualityTokens(channelName)
	for _, alt := range altNames {
		epgTokens = mergeTokens(epgTokens, qualityTokens(alt))
	}
	bestID := candidates[0]
	bestScore := -1
	for _, id := range candidates {
		ch, ok := channelByID[id]
		if !ok {
			continue
		}
		score := 0
		nameNorm := normalizeMatchKey(ch.Name)
		if nameNorm == normalizeMatchKey(channelName) {
			score += 4
		}
		for _, alt := range altNames {
			if nameNorm == normalizeMatchKey(alt) {
				score += 3
				break
			}
		}
		chTokens := qualityTokens(ch.Name)
		if tokensEqual(epgTokens, chTokens) {
			score += 2
		} else if len(epgTokens) == 0 && len(chTokens) == 0 {
			score += 1
		}
		if score > bestScore {
			bestScore = score
			bestID = id
		}
	}
	return bestID
}

func qualityTokens(v string) map[string]struct{} {
	k := normalizeMatchKey(v)
	out := map[string]struct{}{}
	for _, p := range strings.Fields(k) {
		switch p {
		case "hd", "fhd", "uhd", "4k":
			out[p] = struct{}{}
		}
	}
	return out
}

func mergeTokens(base, add map[string]struct{}) map[string]struct{} {
	if base == nil {
		base = map[string]struct{}{}
	}
	for k := range add {
		base[k] = struct{}{}
	}
	return base
}

func tokensEqual(a, b map[string]struct{}) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if _, ok := b[k]; !ok {
			return false
		}
	}
	return true
}

var timeshiftSuffixRE = regexp.MustCompile(`(?i)\+\d+$`)

func baseNameKey(v string) string {
	raw := strings.TrimSpace(v)
	if raw == "" {
		return ""
	}
	// Keep time-shift channels (+2/+3/...) as-is.
	if timeshiftSuffixRE.MatchString(strings.ReplaceAll(raw, " ", "")) {
		return normalizeMatchKey(raw)
	}

	for _, sep := range []string{" | ", " - ", " — ", " / "} {
		if idx := strings.Index(raw, sep); idx > 0 {
			left := strings.TrimSpace(raw[:idx])
			if left != "" {
				raw = left
				break
			}
		}
	}

	k := normalizeMatchKey(raw)
	if k == "" {
		return ""
	}
	parts := strings.Fields(k)
	if len(parts) < 2 {
		return k
	}
	last := parts[len(parts)-1]
	suffixes := map[string]struct{}{
		"hd": {}, "fhd": {}, "uhd": {}, "sd": {}, "4k": {}, "hdr": {},
		"hevc": {}, "h264": {}, "h265": {}, "fullhd": {}, "ultrahd": {},
	}
	if _, ok := suffixes[last]; ok {
		return strings.Join(parts[:len(parts)-1], " ")
	}
	return k
}

func findEPGDonorChannelID(name string, byNormalizedName map[string][]int64, channelByID map[int64]model.Channel, chPrograms map[int64][]model.Program) int64 {
	norm := normalizeMatchKey(name)
	if norm == "" {
		return 0
	}
	parts := strings.Fields(norm)
	if len(parts) < 2 {
		return 0
	}
	joinedNoSpace := strings.ReplaceAll(norm, " ", "")
	if timeshiftSuffixRE.MatchString(joinedNoSpace) {
		return 0
	}
	for i := len(parts) - 1; i >= 1; i-- {
		base := strings.Join(parts[:i], " ")
		candidates := byNormalizedName[base]
		for _, id := range candidates {
			if len(chPrograms[id]) == 0 {
				continue
			}
			if _, ok := channelByID[id]; ok {
				return id
			}
		}
	}
	return 0
}

func donorNameKey(name string) string {
	norm := normalizeMatchKey(name)
	if norm == "" {
		return ""
	}
	joined := strings.ReplaceAll(norm, " ", "")
	if timeshiftSuffixRE.MatchString(joined) {
		return ""
	}
	parts := strings.Fields(norm)
	if len(parts) == 0 {
		return ""
	}
	skip := map[string]struct{}{
		"hd": {}, "fhd": {}, "uhd": {}, "sd": {}, "4k": {}, "hdr": {},
		"hevc": {}, "h264": {}, "h265": {}, "fullhd": {}, "ultrahd": {},
		"orig": {}, "original": {},
	}
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if _, ok := skip[p]; ok {
			continue
		}
		out = append(out, p)
	}
	return strings.TrimSpace(strings.Join(out, " "))
}

type Scheduler struct {
	syncer *SyncService
	store  *storage.Store
}

func New(syncer *SyncService, store *storage.Store) *Scheduler {
	return &Scheduler{syncer: syncer, store: store}
}

func (s *Scheduler) Start(ctx context.Context) {
	t := time.NewTicker(1 * time.Minute)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			pls, err := s.store.ListPlaylists(ctx)
			if err != nil {
				logx.Errorf("scheduler list playlists: %v", err)
				continue
			}
			now := time.Now().UTC()
			for _, p := range pls {
				if !p.Enabled {
					continue
				}
				if p.LastSyncAt != nil && now.Sub(*p.LastSyncAt) < time.Duration(p.UpdateIntervalMinutes)*time.Minute {
					continue
				}
				if err := s.syncer.SyncPlaylist(ctx, p); err != nil {
					logx.Errorf("sync playlist %d: %v", p.ID, err)
				}
			}
		}
	}
}
