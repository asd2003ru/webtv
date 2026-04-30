package epg

import (
	"bytes"
	"compress/gzip"
	"testing"
)

func TestParseXMLTVGZ(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<tv>
  <channel id="ch1">
    <display-name>Channel One</display-name>
    <display-name>Channel One HD</display-name>
    <icon src="http://example.com/ch1.png"></icon>
  </channel>
  <programme channel="ch1" start="20260101090000 +0300" stop="20260101100000 +0300">
    <title>Morning News</title>
    <desc>Daily bulletin</desc>
  </programme>
</tv>`
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	_, _ = gz.Write([]byte(xml))
	_ = gz.Close()

	out, err := ParseXMLTVGZ(&b)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 item, got %d", len(out))
	}
	if out[0].Title != "Morning News" || out[0].ChannelID != "ch1" {
		t.Fatalf("unexpected output: %+v", out[0])
	}
	if out[0].ChannelName != "Channel One" {
		t.Fatalf("unexpected channel name: %+v", out[0])
	}
	if len(out[0].ChannelAltNames) != 2 || out[0].ChannelAltNames[1] != "Channel One HD" {
		t.Fatalf("unexpected channel alt names: %+v", out[0].ChannelAltNames)
	}
	if out[0].ChannelIcon != "http://example.com/ch1.png" {
		t.Fatalf("unexpected channel icon: %+v", out[0])
	}
}
