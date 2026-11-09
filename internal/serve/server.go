package serve

import (
	"context"
	"crypto/subtle"
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/mutedfalcontv/tv/internal/adb"
)

//go:embed all:static
var angularFS embed.FS

// Options configures the HTTP server's security posture.
type Options struct {
	// Token, when non-empty, requires "Authorization: Bearer <Token>" on
	// /api/* and /ws/* requests.
	Token string
	// CORSOrigins lists extra origins allowed to make cross-origin requests.
	// Loopback origins are always allowed; "*" is never emitted.
	CORSOrigins []string
}

type Server struct {
	config      *adb.Config
	runner      adb.Runner
	server      *http.Server
	token       string
	corsOrigins []string
}

type Status struct {
	Connected bool   `json:"connected"`
	TVIP      string `json:"tv_ip"`
}

func New(cfg *adb.Config, runner adb.Runner, opts Options) *Server {
	return &Server{
		config:      cfg,
		runner:      runner,
		token:       opts.Token,
		corsOrigins: opts.CORSOrigins,
	}
}

func (s *Server) Start(addr string) error {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/status", s.handleStatus)
	mux.HandleFunc("GET /api/status/keys", s.handleKeys)
	mux.HandleFunc("POST /api/remote/press", s.handlePress)
	mux.HandleFunc("POST /api/remote/type", s.handleType)
	mux.HandleFunc("GET /api/apps", s.handleApps)
	mux.HandleFunc("POST /api/apps/launch", s.handleLaunch)
	mux.HandleFunc("POST /api/apps/kill", s.handleKill)
	mux.HandleFunc("GET /api/players", s.handlePlayers)
	mux.HandleFunc("PUT /api/player/default", s.handlePlayerDefault)
	mux.HandleFunc("POST /api/play", s.handlePlay)
	mux.HandleFunc("GET /api/logs", s.handleLogs)
	mux.HandleFunc("GET /api/config", s.handleGetConfig)
	mux.HandleFunc("PUT /api/config", s.handlePutConfig)
	mux.HandleFunc("/ws/logs", s.handleWSLogs)

	subFS, err := fs.Sub(angularFS, "static/browser")
	if err != nil {
		return err
	}
	mux.Handle("/", spaFallback(http.FileServer(http.FS(subFS)), subFS))

	s.server = &http.Server{
		Addr:         addr,
		Handler:      s.corsMiddleware(s.authMiddleware(mux)),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	log.Printf("TV Remote UI at http://%s", addr)
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// authMiddleware requires a bearer token on API and WebSocket routes when the
// server was configured with one. Static assets remain public so the SPA can
// load and prompt for/normalize auth.
func (s *Server) authMiddleware(next http.Handler) http.Handler {
	if s.token == "" {
		return next
	}
	want := "Bearer " + s.token
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !protectedPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}
		got := r.Header.Get("Authorization")
		if subtle.ConstantTimeCompare([]byte(got), []byte(want)) != 1 {
			jsonError(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func protectedPath(path string) bool {
	return strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/ws/")
}

// corsMiddleware reflects only origins the server trusts (configured origins
// plus loopback) and never sends a wildcard. Requests with no Origin (CLI,
// agents) are unaffected.
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && s.originAllowed(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Add("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// checkOrigin is the WebSocket upgrader's origin policy. Explicit configured
// origins and loopback are allowed, as is same-origin (browser pages served by
// this server, including on a non-loopback bind). Cross-origin requests from
// unknown sites are rejected. Requests without an Origin header (non-browser
// clients) are allowed.
func (s *Server) checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	if s.originAllowed(origin) {
		return true
	}
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	return strings.EqualFold(u.Host, r.Host)
}

func (s *Server) originAllowed(origin string) bool {
	trimmed := strings.TrimRight(origin, "/")
	for _, o := range s.corsOrigins {
		if strings.EqualFold(strings.TrimRight(o, "/"), trimmed) {
			return true
		}
	}
	return isLoopbackOrigin(origin)
}

func isLoopbackOrigin(origin string) bool {
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	host := u.Hostname()
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func spaFallback(h http.Handler, fsys fs.FS) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path != "/" {
			_, err := fsys.Open(path[1:])
			if err == nil {
				h.ServeHTTP(w, r)
				return
			}
		}
		r.URL.Path = "/"
		h.ServeHTTP(w, r)
	})
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func jsonOK(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}
