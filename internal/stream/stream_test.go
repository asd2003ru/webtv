package stream

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBrowserMode(t *testing.T) {
	if BrowserMode("http://x/live_ac3.m3u8") != ModeTranscode {
		t.Fatal("expected transcode mode")
	}
	if BrowserMode("http://x/live_aac.m3u8") != ModeDirect {
		t.Fatal("expected direct mode")
	}
}

func TestProbeHLSModeChecksVariantPlaylistCodecs(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/master.m3u8", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
		_, _ = w.Write([]byte(`#EXTM3U
#EXT-X-STREAM-INF:BANDWIDTH=4000000,RESOLUTION=1920x1080
variant/main.m3u8
`))
	})
	mux.HandleFunc("/variant/main.m3u8", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
		_, _ = w.Write([]byte(`#EXTM3U
#EXT-X-STREAM-INF:BANDWIDTH=4000000,CODECS="hvc1.1.6.L123.B0,mp4a.40.2"
chunklist.m3u8
`))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	manager := NewManager("", "ffmpeg", 0, 0, false)
	req := httptest.NewRequest(http.MethodHead, "/stream", nil)

	mode, ok := manager.probeHLSMode(req, server.URL+"/master.m3u8")
	if !ok {
		t.Fatal("expected HLS probe to succeed")
	}
	if mode != ModeTranscode {
		t.Fatalf("expected transcode mode, got %s", mode)
	}
}
