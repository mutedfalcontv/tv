package serve

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthSourceString(t *testing.T) {
	tests := []struct {
		source AuthSource
		want   string
	}{
		{AuthNone, "none"},
		{AuthHeader, "header"},
		{AuthQuery, "query"},
		{AuthCookie, "cookie"},
		{AuthSource(99), "none"},
	}
	for _, tt := range tests {
		if got := tt.source.String(); got != tt.want {
			t.Errorf("AuthSource(%d).String() = %q, want %q", tt.source, got, tt.want)
		}
	}
}

func TestBearerCredential(t *testing.T) {
	tests := []struct {
		name       string
		header     string
		query      string
		cookie     string
		wantValue  string
		wantSource AuthSource
	}{
		{name: "none", wantValue: "", wantSource: AuthNone},
		{name: "header", header: "Bearer h", wantValue: "Bearer h", wantSource: AuthHeader},
		{name: "query", query: "q", wantValue: "Bearer q", wantSource: AuthQuery},
		{name: "cookie", cookie: "c", wantValue: "Bearer c", wantSource: AuthCookie},
		{
			name:       "header wins over query and cookie",
			header:     "Bearer h",
			query:      "q",
			cookie:     "c",
			wantValue:  "Bearer h",
			wantSource: AuthHeader,
		},
		{
			name:       "query wins over cookie",
			query:      "q",
			cookie:     "c",
			wantValue:  "Bearer q",
			wantSource: AuthQuery,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := "/api/status"
			if tt.query != "" {
				target += "?token=" + tt.query
			}
			req := httptest.NewRequest(http.MethodGet, target, nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			if tt.cookie != "" {
				req.AddCookie(&http.Cookie{Name: "tv_token", Value: tt.cookie})
			}

			got, source := bearerCredential(req)
			if got != tt.wantValue || source != tt.wantSource {
				t.Errorf("bearerCredential() = (%q, %v), want (%q, %v)",
					got, source, tt.wantValue, tt.wantSource)
			}
		})
	}
}

func TestAuthMiddleware_TokenSources(t *testing.T) {
	s := New(nil, nil, Options{Token: "s3cret"})
	h := s.authMiddleware(okHandler())

	tests := []struct {
		name   string
		target string
		header string
		cookie string
		want   int
	}{
		{"query", "/api/status?token=s3cret", "", "", http.StatusOK},
		{"query wrong", "/api/status?token=nope", "", "", http.StatusUnauthorized},
		{"cookie", "/api/status", "", "s3cret", http.StatusOK},
		{"cookie wrong", "/api/status", "", "nope", http.StatusUnauthorized},
		{"header wins over bad query", "/api/status?token=nope", "Bearer s3cret", "", http.StatusOK},
		{"query wins over bad cookie", "/api/status?token=s3cret", "", "nope", http.StatusOK},
		{"no credential", "/api/status", "", "", http.StatusUnauthorized},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.target, nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			if tt.cookie != "" {
				req.AddCookie(&http.Cookie{Name: "tv_token", Value: tt.cookie})
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code != tt.want {
				t.Fatalf("status = %d, want %d", rec.Code, tt.want)
			}
		})
	}
}

func TestOriginClassString(t *testing.T) {
	tests := []struct {
		class OriginClass
		want  string
	}{
		{OriginNone, "none"},
		{OriginLoopback, "loopback"},
		{OriginConfigured, "configured"},
		{OriginSameOrigin, "same-origin"},
		{OriginRejectedDNS, "rejected-dns"},
		{OriginRejected, "rejected"},
		{OriginClass(99), "none"},
	}
	for _, tt := range tests {
		if got := tt.class.String(); got != tt.want {
			t.Errorf("OriginClass(%d).String() = %q, want %q", tt.class, got, tt.want)
		}
	}
}

func TestOriginClassTrustAndReflect(t *testing.T) {
	tests := []struct {
		class       OriginClass
		wantTrusted bool
		wantReflect bool
	}{
		{OriginNone, true, false},
		{OriginLoopback, true, true},
		{OriginConfigured, true, true},
		{OriginSameOrigin, true, false},
		{OriginRejectedDNS, false, false},
		{OriginRejected, false, false},
		{OriginClass(99), false, false},
	}
	for _, tt := range tests {
		if got := tt.class.Trusted(); got != tt.wantTrusted {
			t.Errorf("OriginClass(%d).Trusted() = %v, want %v", tt.class, got, tt.wantTrusted)
		}
		if got := tt.class.Reflect(); got != tt.wantReflect {
			t.Errorf("OriginClass(%d).Reflect() = %v, want %v", tt.class, got, tt.wantReflect)
		}
	}
}

func TestClassifyOrigin(t *testing.T) {
	s := New(nil, nil, Options{CORSOrigins: []string{"https://tv.example.com"}})
	tests := []struct {
		name   string
		origin string
		host   string
		want   OriginClass
	}{
		{"no origin", "", "127.0.0.1:8080", OriginNone},
		{"loopback localhost", "http://localhost:4200", "127.0.0.1:8080", OriginLoopback},
		{"loopback ip", "http://127.0.0.1:8080", "127.0.0.1:8080", OriginLoopback},
		{"loopback ipv6", "http://[::1]:8080", "127.0.0.1:8080", OriginLoopback},
		{"configured", "https://tv.example.com", "127.0.0.1:8080", OriginConfigured},
		{"configured loopback entry", "http://localhost:5555", "127.0.0.1:8080", OriginConfigured},
		{"same origin ip", "http://192.168.1.10:8080", "192.168.1.10:8080", OriginSameOrigin},
		{"same origin dns name", "http://tv.local:8080", "tv.local:8080", OriginRejectedDNS},
		{"different host", "http://evil.example", "127.0.0.1:8080", OriginRejected},
		{"unparseable", "not a url", "127.0.0.1:8080", OriginRejected},
	}
	// An explicit allow-list entry must classify as configured even when it is
	// otherwise a loopback origin.
	s.corsOrigins = append(s.corsOrigins, "http://localhost:5555")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
			req.Host = tt.host
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			if got := s.classifyOrigin(req); got != tt.want {
				t.Errorf("classifyOrigin() = %v, want %v", got, tt.want)
			}
		})
	}
}
