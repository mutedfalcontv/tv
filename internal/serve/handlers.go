package serve

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/mutedfalcontv/tv/internal/adb"
	"github.com/mutedfalcontv/tv/internal/config"
	"github.com/mutedfalcontv/tv/internal/logcat"
	"github.com/mutedfalcontv/tv/internal/player"
	"github.com/mutedfalcontv/tv/internal/remote"
)

type pressRequest struct {
	Key string `json:"key"`
}

type typeRequest struct {
	Text string `json:"text"`
}

type packageRequest struct {
	Package string `json:"package"`
}

type playRequest struct {
	URL    string `json:"url"`
	Player string `json:"player,omitempty"`
}

type configRequest struct {
	TVIP          string `json:"tv_ip,omitempty"`
	DefaultPlayer string `json:"default_player,omitempty"`
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	err := adb.EnsureConnected(s.config, s.runner)
	status := Status{Connected: err == nil, TVIP: s.config.TVIP}
	jsonOK(w, status)
}

func (s *Server) handleKeys(w http.ResponseWriter, r *http.Request) {
	jsonOK(w, remote.Keys())
}

func (s *Server) handlePress(w http.ResponseWriter, r *http.Request) {
	var req pressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := adb.EnsureConnected(s.config, s.runner); err != nil {
		jsonError(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	if err := remote.Press(s.config, s.runner, req.Key); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, map[string]bool{"ok": true})
}

func (s *Server) handleType(w http.ResponseWriter, r *http.Request) {
	var req typeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := adb.EnsureConnected(s.config, s.runner); err != nil {
		jsonError(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	if err := remote.Type(s.config, s.runner, req.Text); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, map[string]bool{"ok": true})
}

func (s *Server) handleApps(w http.ResponseWriter, r *http.Request) {
	if err := adb.EnsureConnected(s.config, s.runner); err != nil {
		jsonError(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	out, err := s.runner.Shell(s.config, "pm list packages")
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	lines := strings.Split(out, "\n")
	var pkgs []string
	for _, line := range lines {
		pkg := strings.TrimPrefix(line, "package:")
		pkg = strings.TrimSpace(pkg)
		if pkg != "" {
			pkgs = append(pkgs, pkg)
		}
	}
	jsonOK(w, pkgs)
}

func (s *Server) handleLaunch(w http.ResponseWriter, r *http.Request) {
	var req packageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := adb.EnsureConnected(s.config, s.runner); err != nil {
		jsonError(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	_, err := s.runner.ShellWithStderr(s.config, "monkey -p "+req.Package+" -c android.intent.category.LAUNCHER 1")
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, map[string]bool{"ok": true})
}

func (s *Server) handleKill(w http.ResponseWriter, r *http.Request) {
	var req packageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := adb.EnsureConnected(s.config, s.runner); err != nil {
		jsonError(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	_, err := s.runner.ShellWithStderr(s.config, "am force-stop "+req.Package)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, map[string]bool{"ok": true})
}

func (s *Server) handlePlayers(w http.ResponseWriter, r *http.Request) {
	players, err := player.Detect(s.config, s.runner)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, players)
}

func (s *Server) handlePlayerDefault(w http.ResponseWriter, r *http.Request) {
	var req packageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	pkg, err := player.Resolve(s.config, s.runner, req.Package)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	cfg, err := config.Load()
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	cfg.DefaultPlayer = pkg
	if err := cfg.Save(); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, map[string]bool{"ok": true})
}

func (s *Server) handlePlay(w http.ResponseWriter, r *http.Request) {
	var req playRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := adb.EnsureConnected(s.config, s.runner); err != nil {
		jsonError(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	playerPkg := req.Player
	if playerPkg == "" {
		playerPkg = s.config.DefaultPlayer
	}
	if playerPkg == "" {
		jsonError(w, "no player specified and no default player set", http.StatusBadRequest)
		return
	}
	if err := player.PlayOnTV(s.config, s.runner, playerPkg, req.URL); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, map[string]bool{"ok": true})
}

func (s *Server) handleLogs(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	var opts logcat.Options
	opts.Level = q.Get("level")
	opts.Tag = q.Get("tag")
	opts.Format = q.Get("format")
	if l := q.Get("lines"); l != "" {
		n, err := strconv.Atoi(l)
		if err == nil {
			opts.Lines = n
		}
	}
	if pkg := q.Get("pkg"); pkg != "" {
		pid, err := logcat.ResolvePID(s.config, s.runner, pkg)
		if err != nil {
			jsonError(w, err.Error(), http.StatusBadRequest)
			return
		}
		opts.Package = strconv.Itoa(pid)
	}
	if q.Has("dump") {
		opts.Dump = true
	}
	out, err := logcat.RunOnce(s.config, s.runner, opts)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, map[string]string{"logs": out})
}

func (s *Server) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	jsonOK(w, s.config)
}

func (s *Server) handlePutConfig(w http.ResponseWriter, r *http.Request) {
	var req configRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	cfg, err := config.Load()
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if req.TVIP != "" {
		cfg.TVIP = req.TVIP
		s.config.TVIP = req.TVIP
	}
	if req.DefaultPlayer != "" {
		cfg.DefaultPlayer = req.DefaultPlayer
		s.config.DefaultPlayer = req.DefaultPlayer
	}
	if err := cfg.Save(); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, map[string]bool{"ok": true})
}
