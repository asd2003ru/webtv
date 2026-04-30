package epg

import (
	"compress/gzip"
	"encoding/xml"
	"io"
	"strings"
	"time"
)

type TV struct {
	Channels []ChannelXML `xml:"channel"`
	Programs []ProgramXML `xml:"programme"`
}

type ChannelXML struct {
	ID           string   `xml:"id,attr"`
	DisplayNames []string `xml:"display-name"`
	Icons        []IconXML `xml:"icon"`
}

type IconXML struct {
	Src string `xml:"src,attr"`
}

type ProgramXML struct {
	Channel string `xml:"channel,attr"`
	Start   string `xml:"start,attr"`
	Stop    string `xml:"stop,attr"`
	Title   string `xml:"title"`
	Desc    string `xml:"desc"`
}

type Program struct {
	ChannelID      string
	ChannelName    string
	ChannelAltNames []string
	ChannelIcon    string
	StartAt        time.Time
	EndAt          time.Time
	Title          string
	Description    string
}

func ParseXMLTVGZ(r io.Reader) ([]Program, error) {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	defer gz.Close()
	return ParseXMLTV(gz)
}

func ParseXMLTV(r io.Reader) ([]Program, error) {
	dec := xml.NewDecoder(r)
	var tv TV
	if err := dec.Decode(&tv); err != nil {
		return nil, err
	}
	channelNames := make(map[string]string, len(tv.Channels))
	channelAllNames := make(map[string][]string, len(tv.Channels))
	channelIcons := make(map[string]string, len(tv.Channels))
	for _, ch := range tv.Channels {
		for _, name := range ch.DisplayNames {
			name = strings.TrimSpace(name)
			if name == "" {
				continue
			}
			channelAllNames[ch.ID] = append(channelAllNames[ch.ID], name)
			if _, exists := channelNames[ch.ID]; !exists {
				channelNames[ch.ID] = name
			}
		}
		for _, icon := range ch.Icons {
			src := strings.TrimSpace(icon.Src)
			if src == "" {
				continue
			}
			if _, exists := channelIcons[ch.ID]; !exists {
				channelIcons[ch.ID] = src
			}
		}
	}
	out := make([]Program, 0, len(tv.Programs))
	for _, p := range tv.Programs {
		start, err := parseXMLTVTime(p.Start)
		if err != nil {
			continue
		}
		end, err := parseXMLTVTime(p.Stop)
		if err != nil {
			continue
		}
		out = append(out, Program{
			ChannelID:       p.Channel,
			ChannelName:     channelNames[p.Channel],
			ChannelAltNames: channelAllNames[p.Channel],
			ChannelIcon:     channelIcons[p.Channel],
			StartAt:         start,
			EndAt:           end,
			Title:           strings.TrimSpace(p.Title),
			Description:     strings.TrimSpace(p.Desc),
		})
	}
	return out, nil
}

func parseXMLTVTime(v string) (time.Time, error) {
	v = strings.TrimSpace(v)
	if len(v) >= len("20060102150405 -0700") {
		v = v[:len("20060102150405 -0700")]
		return time.Parse("20060102150405 -0700", v)
	}
	return time.Parse("20060102150405", v)
}
