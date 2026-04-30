package stream

import "testing"

func TestBrowserMode(t *testing.T) {
	if BrowserMode("http://x/live_ac3.m3u8") != ModeTranscode {
		t.Fatal("expected transcode mode")
	}
	if BrowserMode("http://x/live_aac.m3u8") != ModeDirect {
		t.Fatal("expected direct mode")
	}
}
