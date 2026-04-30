package playlist

import (
	"strings"
	"testing"
)

func TestParseM3U(t *testing.T) {
	input := `#EXTM3U
#EXTINF:-1 tvg-id="ch1" tvg-logo="logo.png" tvg-chno="11" group-title="News" catchup="default",Channel One
http://example.com/stream1.m3u8
`
	out, err := ParseM3U(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 channel, got %d", len(out))
	}
	if out[0].ExternalID != "ch1" || !out[0].ArchiveSupported {
		t.Fatalf("unexpected parse result: %+v", out[0])
	}
	if out[0].SortIndex != 11 {
		t.Fatalf("expected sort index 11, got %d", out[0].SortIndex)
	}
}

func TestParseM3U_EXTGRP(t *testing.T) {
	input := `#EXTM3U
#EXTINF:-1 tvg-id='ch2' tvg-logo='logo2.png',Channel Two
#EXTGRP:Movies
http://example.com/stream2.m3u8
`
	out, err := ParseM3U(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 channel, got %d", len(out))
	}
	if out[0].Group != "Movies" {
		t.Fatalf("expected group Movies, got %q", out[0].Group)
	}
	if out[0].ExternalID != "ch2" {
		t.Fatalf("expected external id ch2, got %q", out[0].ExternalID)
	}
}

func TestParseM3U_ExternalIDFallbackToTvgName(t *testing.T) {
	input := `#EXTM3U
#EXTINF:-1 tvg-name="rossiya_1" group-title="General",Россия 1
http://example.com/stream3.m3u8
`
	out, err := ParseM3U(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 channel, got %d", len(out))
	}
	if out[0].ExternalID != "rossiya_1" {
		t.Fatalf("expected external id rossiya_1, got %q", out[0].ExternalID)
	}
}

func TestParseM3U_AttrsCaseInsensitive(t *testing.T) {
	input := `#EXTM3U
#EXTINF:-1 TVG-ID="ch4" TVG-LOGO="logo4.png" GROUP-TITLE="Kids",Channel Four
http://example.com/stream4.m3u8
`
	out, err := ParseM3U(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 channel, got %d", len(out))
	}
	if out[0].ExternalID != "ch4" || out[0].Logo != "logo4.png" || out[0].Group != "Kids" {
		t.Fatalf("unexpected parse result: %+v", out[0])
	}
}

func TestParseM3U_UsesPlaylistOrderAsSortIndex(t *testing.T) {
	input := `#EXTM3U
#EXTINF:-1 tvg-id="ch10",Ten
http://example.com/10.m3u8
#EXTINF:-1 tvg-id="ch20",Twenty
http://example.com/20.m3u8
`
	out, err := ParseM3U(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 2 {
		t.Fatalf("expected 2 channels, got %d", len(out))
	}
	if out[0].SortIndex != 1 || out[1].SortIndex != 2 {
		t.Fatalf("unexpected sort indexes: %d, %d", out[0].SortIndex, out[1].SortIndex)
	}
}

func TestParseM3U_UnquotedAttributes(t *testing.T) {
	input := `#EXTM3U
#EXTINF:-1 tvg-id=ch5 tvg-logo=http://img.local/logo5.png group-title=Sport,Channel Five
http://example.com/stream5.m3u8
`
	out, err := ParseM3U(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 channel, got %d", len(out))
	}
	if out[0].ExternalID != "ch5" || out[0].Logo != "http://img.local/logo5.png" || out[0].Group != "Sport" {
		t.Fatalf("unexpected parse result: %+v", out[0])
	}
}

func TestParseM3U_GlobalCatchupFromHeader(t *testing.T) {
	input := `#EXTM3U catchup="append" catchup-days="3"
#EXTINF:-1 tvg-id="ch6",Channel Six
http://example.com/stream6.m3u8
`
	out, err := ParseM3U(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 channel, got %d", len(out))
	}
	if !out[0].ArchiveSupported {
		t.Fatalf("expected archive supported, got false")
	}
}

func TestParseM3U_EdemArchiveByTvgRec(t *testing.T) {
	input := `#EXTM3U
#EXTINF:-1 tvg-id="edem_ch" tvg-rec="4" group-title="General",Edem Channel
http://example.com/edem.m3u8
`
	out, err := ParseM3U(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 channel, got %d", len(out))
	}
	if !out[0].ArchiveSupported {
		t.Fatalf("expected archive supported for tvg-rec channel, got false")
	}
}
