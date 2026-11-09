package serve

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
}

func TestAuthMiddleware_NoTokenAllowsEverything(t *testing.T) {
	s := New(nil, nil, Options{})
	h := s.authMiddleware(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestAuthMiddleware_TokenRequired(t *testing.T) {
	s := New(nil, nil, Options{Token: "s3cret"})
	h := s.authMiddleware(okHandler())

	tests := []struct {
		name   string
		path   string
		header string
		want   int
	}{
		{"api missing", "/api/status", "", http.StatusUnauthorized},
		{"api wrong", "/api/status", "Bearer nope", http.StatusUnauthorized},
		{"api wrong scheme", "/api/status", "s3cret", http.StatusUnauthorized},
		{"api correct", "/api/status", "Bearer s3cret", http.StatusOK},
		{"ws missing", "/ws/logs", "", http.StatusUnauthorized},
		{"ws correct", "/ws/logs", "Bearer s3cret", http.StatusOK},
		{"static public", "/", "", http.StatusOK},
		{"static asset public", "/main.js", "", http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code != tt.want {
				t.Fatalf("status = %d, want %d", rec.Code, tt.want)
			}
		})
	}
}

func TestCORS_RejectsWildcardAndUnknownOrigins(t *testing.T) {
	s := New(nil, nil, Options{})
	h := s.corsMiddleware(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	req.Header.Set("Origin", "http://evil.example")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want empty", got)
	}
}

func TestCORS_ReflectsLoopbackOrigin(t *testing.T) {
	s := New(nil, nil, Options{})
	h := s.corsMiddleware(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	req.Header.Set("Origin", "http://localhost:4200")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:4200" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want reflected loopback origin", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got == "*" {
		t.Fatal("must not emit wildcard origin")
	}
}

func TestCORS_ConfiguredOriginAndPreflight(t *testing.T) {
	s := New(nil, nil, Options{CORSOrigins: []string{"https://tv.example.com"}})
	h := s.corsMiddleware(okHandler())

	req := httptest.NewRequest(http.MethodOptions, "/api/status", nil)
	req.Header.Set("Origin", "https://tv.example.com")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("preflight status = %d, want 204", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://tv.example.com" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want configured origin", got)
	}
}

func TestCORS_AllowsSameOriginIPMutation(t *testing.T) {
	s := New(nil, nil, Options{})
	h := s.corsMiddleware(okHandler())

	req := httptest.NewRequest(http.MethodPost, "/api/remote/press", strings.NewReader(`{"key":"home"}`))
	req.Host = "192.168.1.10:8080"
	req.Header.Set("Origin", "http://192.168.1.10:8080")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("same-origin IP mutation status = %d, want 200", rec.Code)
	}
}

func TestCORS_RejectsCrossOriginMutation(t *testing.T) {
	s := New(nil, nil, Options{})
	h := s.corsMiddleware(okHandler())

	req := httptest.NewRequest(http.MethodPost, "/api/remote/press", strings.NewReader(`{}`))
	req.Header.Set("Origin", "http://evil.example")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("cross-origin mutation status = %d, want 403", rec.Code)
	}
}

func TestIsLoopbackOrigin(t *testing.T) {
	tests := map[string]bool{
		"http://localhost:8080":  true,
		"http://127.0.0.1:8080":  true,
		"http://[::1]:8080":      true,
		"https://evil.example":   false,
		"http://192.168.1.10:80": false,
		"not a url":              false,
	}
	for origin, want := range tests {
		if got := isLoopbackOrigin(origin); got != want {
			t.Errorf("isLoopbackOrigin(%q) = %v, want %v", origin, got, want)
		}
	}
}

func TestCheckOrigin(t *testing.T) {
	s := New(nil, nil, Options{Token: "t"})
	tests := []struct {
		name   string
		origin string
		host   string
		want   bool
	}{
		{"no origin allowed", "", "127.0.0.1:8080", true},
		{"loopback allowed", "http://localhost:8080", "127.0.0.1:8080", true},
		{"same origin on LAN allowed", "http://192.168.1.10:8080", "192.168.1.10:8080", true},
		{"same host DNS name rejected (rebinding)", "http://tv.local:8080", "tv.local:8080", false},
		{"cross origin rejected", "http://evil.example", "127.0.0.1:8080", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/ws/logs", nil)
			req.Host = tt.host
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			if got := s.checkOrigin(req); got != tt.want {
				t.Fatalf("checkOrigin = %v, want %v", got, tt.want)
			}
		})
	}
}
