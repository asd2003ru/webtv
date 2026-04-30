package stream

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestProxyManifestSingleflight(t *testing.T) {
	var calls atomic.Int32
	m := NewManager("", "ffmpeg", 12*time.Second, 0, false)
	m.client = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			calls.Add(1)
			time.Sleep(80 * time.Millisecond)
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/vnd.apple.mpegurl"}},
				Body:       io.NopCloser(bytes.NewBufferString("#EXTM3U\nseg.ts\n")),
			}, nil
		}),
	}
	streamURL := "http://upstream/live.m3u8"

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := newRequest(http.MethodGet, "/stream/1")
			rr := newRecorder()
			m.HandleWithMode(rr, req, streamURL, ModeDirect)
			if rr.Code() != http.StatusOK {
				t.Errorf("unexpected status: %d", rr.Code())
			}
			if !strings.Contains(rr.BodyString(), "/stream/1?src=") {
				t.Errorf("expected rewritten manifest, got: %q", rr.BodyString())
			}
		}()
	}
	wg.Wait()

	if got := calls.Load(); got != 1 {
		t.Fatalf("expected 1 upstream call, got %d", got)
	}
}

func TestProxyManifestFallbackStaleOn429(t *testing.T) {
	var calls atomic.Int32
	m := NewManager("", "ffmpeg", 12*time.Second, 0, false)
	m.client = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			n := calls.Add(1)
			if n == 1 {
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     http.Header{"Content-Type": []string{"application/vnd.apple.mpegurl"}},
					Body:       io.NopCloser(bytes.NewBufferString("#EXTM3U\nfirst.ts\n")),
				}, nil
			}
			return &http.Response{
				StatusCode: http.StatusTooManyRequests,
				Header:     http.Header{"Content-Type": []string{"application/vnd.apple.mpegurl"}},
				Body:       io.NopCloser(bytes.NewBufferString("rate limited")),
			}, nil
		}),
	}
	streamURL := "http://upstream/live.m3u8"

	req1 := newRequest(http.MethodGet, "/stream/2")
	rr1 := newRecorder()
	m.HandleWithMode(rr1, req1, streamURL, ModeDirect)
	if rr1.Code() != http.StatusOK {
		t.Fatalf("first request status: %d", rr1.Code())
	}

	m.manifestMu.Lock()
	e := m.manifests[streamURL]
	e.freshUntil = time.Now().Add(-time.Second)
	m.manifests[streamURL] = e
	m.manifestMu.Unlock()

	req2 := newRequest(http.MethodGet, "/stream/2")
	rr2 := newRecorder()
	m.HandleWithMode(rr2, req2, streamURL, ModeDirect)

	if rr2.Code() != http.StatusOK {
		t.Fatalf("fallback request status: %d", rr2.Code())
	}
	if rr2.Header().Get("X-Stream-Cache") != "stale" {
		t.Fatalf("expected stale cache header")
	}
	if !strings.Contains(rr2.BodyString(), "first.ts") {
		t.Fatalf("expected cached manifest body, got %q", rr2.BodyString())
	}
}

func TestProxyManifestBackoffOnUnchangedPlaylist(t *testing.T) {
	var calls atomic.Int32
	playlist := "#EXTM3U\n#EXT-X-TARGETDURATION:5\n#EXTINF:5.000,\nseg.ts\n"
	m := NewManager("", "ffmpeg", 12*time.Second, 0, true)
	m.client = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			calls.Add(1)
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/vnd.apple.mpegurl"}},
				Body:       io.NopCloser(bytes.NewBufferString(playlist)),
			}, nil
		}),
	}
	streamURL := "http://upstream/live.m3u8"

	req1 := newRequest(http.MethodGet, "/stream/3")
	rr1 := newRecorder()
	m.HandleWithMode(rr1, req1, streamURL, ModeDirect)
	if rr1.Code() != http.StatusOK {
		t.Fatalf("first request status: %d", rr1.Code())
	}

	m.manifestMu.Lock()
	e := m.manifests[streamURL]
	e.freshUntil = time.Now().Add(-time.Second)
	m.manifests[streamURL] = e
	m.manifestMu.Unlock()

	req2 := newRequest(http.MethodGet, "/stream/3")
	rr2 := newRecorder()
	m.HandleWithMode(rr2, req2, streamURL, ModeDirect)
	if rr2.Code() != http.StatusOK {
		t.Fatalf("second request status: %d", rr2.Code())
	}

	m.manifestMu.Lock()
	e = m.manifests[streamURL]
	e.freshUntil = time.Now().Add(-time.Second)
	m.manifests[streamURL] = e
	m.manifestMu.Unlock()

	req3 := newRequest(http.MethodGet, "/stream/3")
	rr3 := newRecorder()
	m.HandleWithMode(rr3, req3, streamURL, ModeDirect)
	if rr3.Code() != http.StatusOK {
		t.Fatalf("third request status: %d", rr3.Code())
	}
	if rr3.Header().Get("X-Stream-Cache") != "backoff" {
		t.Fatalf("expected backoff cache, got %q", rr3.Header().Get("X-Stream-Cache"))
	}
	if got := calls.Load(); got != 2 {
		t.Fatalf("expected 2 upstream calls due to backoff, got %d", got)
	}
}

func TestProxyManifestBackoffReturnsCached200(t *testing.T) {
	m := NewManager("", "ffmpeg", 12*time.Second, 0, true)
	streamURL := "http://upstream/live.m3u8"
	m.setManifestCache(streamURL, manifestEntry{
		status:     http.StatusOK,
		header:     http.Header{"Content-Type": []string{"application/vnd.apple.mpegurl"}},
		body:       []byte("#EXTM3U\n"),
		freshUntil: time.Now().Add(-time.Second),
		staleUntil: time.Now().Add(10 * time.Second),
		unchanged:  5,
		backoffTo:  time.Now().Add(5 * time.Second),
	})

	// Request should be served from backoff cache as 200 to avoid client retry storms.
	req := newRequest(http.MethodGet, "/stream/4")
	rr := newRecorder()
	m.HandleWithMode(rr, req, streamURL, ModeDirect)
	if rr.Code() != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code())
	}
	if rr.Header().Get("X-Stream-Cache") != "backoff" {
		t.Fatalf("expected backoff cache, got %q", rr.Header().Get("X-Stream-Cache"))
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func newRequest(method, path string) *http.Request {
	req, err := http.NewRequest(method, "http://local"+path, nil)
	if err != nil {
		panic(err)
	}
	return req
}

func newRecorder() *responseRecorder {
	return &responseRecorder{header: make(http.Header)}
}

type responseRecorder struct {
	header http.Header
	body   bytes.Buffer
	code   int
}

func (r *responseRecorder) Header() http.Header {
	return r.header
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.code = statusCode
}

func (r *responseRecorder) Write(p []byte) (int, error) {
	if r.code == 0 {
		r.code = http.StatusOK
	}
	return r.body.Write(p)
}

func (r *responseRecorder) Code() int {
	if r.code == 0 {
		return http.StatusOK
	}
	return r.code
}

func (r *responseRecorder) BodyString() string {
	return r.body.String()
}
