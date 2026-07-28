package serve

import (
	"context"
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"time"

	"github.com/mutedfalcontv/tv/internal/adb"
)

//go:embed all:static
var angularFS embed.FS

type Server struct {
	config *adb.Config
	runner adb.Runner
	server *http.Server
}

type Status struct {
	Connected bool   `json:"connected"`
	TVIP      string `json:"tv_ip"`
}

func New(cfg *adb.Config, runner adb.Runner) *Server {
	return &Server{config: cfg, runner: runner}
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

	subFS, err := fs.Sub(angularFS, "static")
	if err != nil {
		return err
	}
	fileServer := http.FileServer(http.FS(subFS))
	mux.Handle("/", fileServer)

	s.server = &http.Server{
		Addr:         addr,
		Handler:      corsMiddleware(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	log.Printf("TV Remote UI at http://%s", addr)
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
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
