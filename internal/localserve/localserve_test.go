package localserve

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writeTempFile(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}
	return path
}

func TestNewHandler_ServesFile(t *testing.T) {
	file := writeTempFile(t, "video.mp4", "0123456789")
	h := NewHandler(file)

	req := httptest.NewRequest(http.MethodGet, "/video.mp4", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if body := rec.Body.String(); body != "0123456789" {
		t.Errorf("body = %q, want %q", body, "0123456789")
	}
}

func TestNewHandler_NotFound(t *testing.T) {
	file := writeTempFile(t, "video.mp4", "data")
	h := NewHandler(file)

	tests := []struct {
		name string
		path string
	}{
		{"other file", "/other.mp4"},
		{"directory listing", "/"},
		{"traversal", "/../video.mp4"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code != http.StatusNotFound {
				t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
			}
		})
	}
}

func TestStart_ServesAndShutsDown(t *testing.T) {
	file := writeTempFile(t, "clip.mkv", "stream-bytes")
	srv, err := Start(file, "127.0.0.1:5555", 0)
	if err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	}()

	if !strings.HasPrefix(srv.URL, "http://127.0.0.1:") {
		t.Errorf("URL = %q, want 127.0.0.1 host", srv.URL)
	}
	if !strings.Contains(srv.URL, "/clip.mkv") {
		t.Errorf("URL = %q, want /clip.mkv path", srv.URL)
	}

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(srv.URL)
	if err != nil {
		t.Fatalf("GET %s error: %v", srv.URL, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ReadAll() error: %v", err)
	}
	if string(body) != "stream-bytes" {
		t.Errorf("body = %q, want %q", body, "stream-bytes")
	}
}

func TestAdvertiseIP(t *testing.T) {
	tests := []struct {
		name string
		tvIP string
	}{
		{"host and port", "127.0.0.1:5555"},
		{"bare host", "127.0.0.1"},
		{"wildcard", "0.0.0.0:5555"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AdvertiseIP(tt.tvIP)
			if got == "" {
				t.Error("AdvertiseIP() returned empty string")
			}
		})
	}
}
