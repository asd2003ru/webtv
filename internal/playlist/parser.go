package playlist

import (
	"bufio"
	"io"
	"strconv"
	"strings"

	"github.com/asd2003ru/webtv/internal/model"
)

type ParsedChannel struct {
	Name             string
	Group            string
	SortIndex        int
	Logo             string
	StreamURL        string
	ArchiveSupported bool
	ExternalID       string
}

func ParseM3U(r io.Reader) ([]ParsedChannel, error) {
	s := bufio.NewScanner(r)
	var out []ParsedChannel
	var current ParsedChannel
	var globalArchiveSupported bool
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" {
			continue
		}
		if strings.HasPrefix(strings.ToUpper(line), "#EXTM3U") {
			if hasArchiveHints(line) {
				globalArchiveSupported = true
			}
			continue
		}
		if strings.HasPrefix(line, "#EXTINF:") {
			current = parseExtInf(line)
			if globalArchiveSupported {
				current.ArchiveSupported = true
			}
			continue
		}
		if strings.HasPrefix(strings.ToUpper(line), "#EXTGRP:") {
			group := strings.TrimSpace(line[len("#EXTGRP:"):])
			if group != "" {
				current.Group = group
			}
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}
		current.StreamURL = line
		if current.SortIndex <= 0 {
			current.SortIndex = len(out) + 1
		}
		if current.ExternalID == "" {
			current.ExternalID = strings.ToLower(strings.ReplaceAll(current.Name, " ", "_"))
		}
		out = append(out, current)
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func parseExtInf(line string) ParsedChannel {
	p := ParsedChannel{}
	meta, name, _ := strings.Cut(line, ",")
	p.Name = strings.TrimSpace(name)
	p.Group = attrCI(meta, "group-title")
	p.SortIndex = parseSortIndex(meta)
	p.Logo = attrCI(meta, "tvg-logo")
	p.ExternalID = firstNonEmpty(
		attrCI(meta, "tvg-id"),
		attrCI(meta, "channel-id"),
		attrCI(meta, "tvg-name"),
		attrCI(meta, "tvg-channel-id"),
	)
	a := strings.ToLower(meta)
	p.ArchiveSupported = hasArchiveHints(a)
	return p
}

func hasArchiveHints(s string) bool {
	s = strings.ToLower(s)
	return strings.Contains(s, "timeshift") ||
		strings.Contains(s, "catchup") ||
		strings.Contains(s, "archive") ||
		strings.Contains(s, "tvg-rec")
}

func attrCI(meta, key string) string {
	lowerMeta := strings.ToLower(meta)
	lowerKey := strings.ToLower(key)
	for _, quote := range []string{`"`, `'`} {
		needle := lowerKey + "=" + quote
		i := strings.Index(lowerMeta, needle)
		if i == -1 {
			continue
		}
		v := meta[i+len(needle):]
		j := strings.Index(v, quote)
		if j == -1 {
			continue
		}
		return v[:j]
	}
	needle := lowerKey + "="
	i := strings.Index(lowerMeta, needle)
	if i == -1 {
		return ""
	}
	v := strings.TrimSpace(meta[i+len(needle):])
	if v == "" {
		return ""
	}
	end := len(v)
	for i, r := range v {
		if r == ' ' || r == '\t' || r == ',' {
			end = i
			break
		}
	}
	return v[:end]
	return ""
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func parseSortIndex(meta string) int {
	raw := firstNonEmpty(
		attrCI(meta, "tvg-chno"),
		attrCI(meta, "chno"),
		attrCI(meta, "channel-number"),
		attrCI(meta, "tvg-num"),
	)
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return 0
	}
	return n
}

func ToModel(playlistID int64, parsed []ParsedChannel) []model.Channel {
	out := make([]model.Channel, 0, len(parsed))
	for _, c := range parsed {
		out = append(out, model.Channel{
			PlaylistID:       playlistID,
			Name:             c.Name,
			Group:            c.Group,
			SortIndex:        c.SortIndex,
			Logo:             c.Logo,
			StreamURL:        c.StreamURL,
			ArchiveSupported: c.ArchiveSupported,
			ExternalID:       c.ExternalID,
		})
	}
	return out
}
