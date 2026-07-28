# TV Remote UI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `tv serve` command with embedded Angular PWA for TV remote control.

**Architecture:** Go HTTP server (`internal/serve/`) serves REST API + WebSocket + embedded Angular static files. New `cmd/tv/serve.go` registers command. Angular PWA at `web/` built separately.

**Tech Stack:** Go 1.26 (net/http, gorilla/websocket), Angular latest, Tailwind CSS, PWA

---

### Task 1: Go internal/serve/server.go — HTTP server with embedded static files

**Files:**
- Create: `internal/serve/server.go`

Note: `//go:embed` paths are relative to the source file's dir and cannot contain `..`. So Angular build output gets copied into `internal/serve/static/` by Makefile before Go build (see Task 8).

- [ ] **Step 1: Create server.go**

```go
package serve

import (
	"context"
	"embed"
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
	w.Write([]byte(`{"error":"` + msg + `"}`))
}

func jsonOK(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	b, _ := json.Marshal(data)
	w.Write(b)
}
```

- [ ] **Step 2: Verify it builds**

Run: `go vet ./internal/serve/`
Expected: PASS (will fail on undefined handlers — that's fine, they come in Task 2)

- [ ] **Step 3: Commit**

```
git add internal/serve/server.go
git commit -m "feat: add HTTP server scaffold with embed and routing"
```

---

### Task 2: REST API handlers

**Files:**
- Create: `internal/serve/handlers.go`

- [ ] **Step 1: Create handlers.go**

```go
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
```

- [ ] **Step 2: Verify build**

Run: `go vet ./internal/serve/`
Expected: PASS (handlers now reference all imported packages)

- [ ] **Step 3: Commit**

```
git add internal/serve/handlers.go
git commit -m "feat: add REST API handlers for all TV operations"
```

---

### Task 3: WebSocket handler for logcat

**Files:**
- Create: `internal/serve/websocket.go`

- [ ] **Step 1: Create websocket.go**

```go
package serve

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/mutedfalcontv/tv/internal/logcat"

	gorilla "github.com/gorilla/websocket"
)

var upgrader = gorilla.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type logFrame struct {
	Line      string `json:"line"`
	Timestamp string `json:"timestamp"`
}

type filterMsg struct {
	Type  string `json:"type"`
	Level string `json:"level,omitempty"`
	Tag   string `json:"tag,omitempty"`
	PID   int    `json:"pid,omitempty"`
}

func (s *Server) handleWSLogs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WS upgrade error: %v", err)
		return
	}
	defer conn.Close()

	adbPath, err := exec.LookPath("adb")
	if err != nil {
		conn.WriteJSON(logFrame{Line: "ADB not found on PATH", Timestamp: time.Now().Format(time.RFC3339)})
		return
	}

	var opts logcat.Options
	opts.Dump = false

	go func() {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			var f filterMsg
			if json.Unmarshal(msg, &f) == nil && f.Type == "filter" {
				if f.Level != "" {
					opts.Level = f.Level
				}
				if f.Tag != "" {
					opts.Tag = f.Tag
				}
				if f.PID > 0 {
					opts.Package = fmt.Sprintf("%d", f.PID)
				}
			}
		}
	}()

	logArgs := logcat.BuildArgs(opts)
	cmd := exec.Command(adbPath, append([]string{"-s", s.config.TVIP, "logcat"}, logArgs...)...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		conn.WriteJSON(logFrame{Line: fmt.Sprintf("Error: %v", err), Timestamp: time.Now().Format(time.RFC3339)})
		return
	}
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		conn.WriteJSON(logFrame{Line: fmt.Sprintf("Error: %v", err), Timestamp: time.Now().Format(time.RFC3339)})
		return
	}
	defer cmd.Wait()

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		frame := logFrame{Line: line, Timestamp: time.Now().Format(time.RFC3339)}
		if err := conn.WriteJSON(frame); err != nil {
			return
		}
	}
}
```

- [ ] **Step 2: Add gorilla/websocket dependency**

Run:
```
cd C:\Users\ajink\Projects\tv
go get github.com/gorilla/websocket
```

- [ ] **Step 3: Verify build**

Run: `go build ./...`
Expected: PASS

- [ ] **Step 4: Commit**

```
git add internal/serve/websocket.go go.mod go.sum
git commit -m "feat: add WebSocket handler for logcat streaming"
```

---

### Task 4: tv serve CLI command

**Files:**
- Create: `cmd/tv/serve.go`

- [ ] **Step 1: Create serve.go**

```go
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/mutedfalcontv/tv/internal/adb"
	"github.com/mutedfalcontv/tv/internal/config"
	"github.com/mutedfalcontv/tv/internal/serve"
)

func serveCmd(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	port := fs.Int("port", 8080, "HTTP server port")
	host := fs.String("host", "0.0.0.0", "HTTP server host")
	noOpen := fs.Bool("no-open", false, "Don't open browser")
	fs.Parse(args)

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %v (using defaults)\n", err)
		cfg = &config.Config{TVIP: "192.168.2.3:5555"}
	}

	adbCfg := &adb.Config{TVIP: cfg.TVIP, DefaultPlayer: cfg.DefaultPlayer}
	runner := &adb.RealRunner{}

	addr := fmt.Sprintf("%s:%d", *host, *port)
	server := serve.New(adbCfg, runner)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if !*noOpen {
		url := fmt.Sprintf("http://localhost:%d", *port)
		openBrowser(url)
	}

	go func() {
		if err := server.Start(addr); err != nil {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			stop()
		}
	}()

	<-ctx.Done()
	fmt.Println("\nShutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), *shutdownTimeout)
	defer cancel()
	server.Shutdown(shutdownCtx)
}

func openBrowser(url string) {
	switch {
	case isWindows():
		execCommand("rundll32", "url.dll,FileProtocolHandler", url)
	case isMac():
		execCommand("open", url)
	default:
		execCommand("xdg-open", url)
	}
}
```

Note: Add `isWindows()`, `isMac()`, `execCommand()` helpers and register `serveCmd` in `commands` slice. Also add `shutdownTimeout` var.

- [ ] **Step 2: Modify cmd/tv/main.go — register serve command**

Add to `commands` slice in main.go:
```go
{"serve", "Start TV Remote UI web server", serveCmd},
```

- [ ] **Step 3: Modify cmd/tv/main.go — add helper functions**

Add near the bottom of main.go:
```go
var shutdownTimeout = flag.Duration("shutdown-timeout", 5*time.Second, "timeout for graceful shutdown")

func isWindows() bool {
	return len(os.Getenv("SYSTEMROOT")) > 0
}

func isMac() bool {
	return len(os.Getenv("HOME")) > 0 && func() bool {
		_, err := exec.LookPath("open")
		return err == nil
	}()
}

func execCommand(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Start()
}
```

- [ ] **Step 4: Verify build**

Run: `go build ./cmd/tv`
Expected: PASS, produces tv.exe

- [ ] **Step 5: Commit**

```
git add cmd/tv/serve.go cmd/tv/main.go
git commit -m "feat: add tv serve command with browser auto-open"
```

---

### Task 5: Scaffold Angular PWA project (following angular-new-app skill)

**Files:**
- Create: `web/` directory with Angular project

- [ ] **Step 1: Check Angular CLI availability**

Run: `gcm ng` (PowerShell)
If not found, install: `npm install -g @angular/cli`

- [ ] **Step 2: Create Angular app with PWA + Tailwind**

Run:
```
cd C:\Users\ajink\Projects\tv
npx @angular/cli@latest new web --routing --style=scss --interactive=false --ai-config=agents
cd web
npx ng add @angular/pwa --skip-confirmation
npx ng add tailwindcss --skip-confirmation
```

- [ ] **Step 3: Update angular.json — output directly to internal/serve/static**

Edit `web/angular.json` — set `outputPath` under `build` → `options`:
```json
"outputPath": "../internal/serve/static"
```

This makes Angular build directly to `internal/serve/static/` where `//go:embed all:static` picks it up.

- [ ] **Step 4: Generate components using Angular CLI**

```
cd web
npx ng generate class models/models --type=ts
npx ng generate service services/api
npx ng generate service services/websocket
npx ng generate component pages/layout --standalone
npx ng generate component pages/remote-page --standalone
npx ng generate component pages/apps-page --standalone
npx ng generate component pages/logs-page --standalone
npx ng generate component pages/player-page --standalone
npx ng generate component pages/settings-page --standalone
npx ng generate component components/dpad --standalone
npx ng generate component components/action-buttons --standalone
npx ng generate component components/volume-bar --standalone
npx ng generate component components/playback-controls --standalone
```

- [ ] **Step 5: Create TypeScript models**

Write `web/src/app/models/models.ts`:
```typescript
export interface AppStatus {
  connected: boolean;
  tv_ip: string;
}

export interface PlayerInfo {
  PackageName: string;
  Name: string;
}

export interface ServerConfig {
  tv_ip: string;
  default_player: string;
}

export interface LogEntry {
  line: string;
  timestamp: string;
}
```

- [ ] **Step 6: Set up routing**

Edit `web/src/app/app.routes.ts`:
```typescript
import { Routes } from '@angular/router';
import { RemotePageComponent } from './pages/remote-page/remote-page.component';
import { AppsPageComponent } from './pages/apps-page/apps-page.component';
import { LogsPageComponent } from './pages/logs-page/logs-page.component';
import { PlayerPageComponent } from './pages/player-page/player-page.component';
import { SettingsPageComponent } from './pages/settings-page/settings-page.component';

export const routes: Routes = [
  { path: '', component: RemotePageComponent },
  { path: 'apps', component: AppsPageComponent },
  { path: 'logs', component: LogsPageComponent },
  { path: 'player', component: PlayerPageComponent },
  { path: 'settings', component: SettingsPageComponent },
];
```

- [ ] **Step 7: Verify Angular build**

Run: `npx ng build`
Expected: PASS, produces `internal/serve/static/` with index.html

- [ ] **Step 8: Commit**

```
git add web/
git commit -m "feat: scaffold Angular PWA project with routing and Tailwind"
```

---

### Task 6: Angular services (scaffolded in Task 5, now implement)

**Files:**
- Modify: `web/src/app/services/api.service.ts`
- Modify: `web/src/app/services/websocket.service.ts`

- [ ] **Step 1: Register provideHttpClient in app config**

Edit `web/src/app/app.config.ts` — add `provideHttpClient(withFetch())`:
```typescript
import { ApplicationConfig, provideZoneChangeDetection } from '@angular/core';
import { provideRouter } from '@angular/router';
import { provideHttpClient, withFetch } from '@angular/common/http';
import { routes } from './app.routes';

export const appConfig: ApplicationConfig = {
  providers: [
    provideZoneChangeDetection({ eventCoalescing: true }),
    provideRouter(routes),
    provideHttpClient(withFetch()),
  ],
};
```

- [ ] **Step 2: Create ApiService with httpResource + HttpClient**

```typescript
import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { httpResource } from '@angular/common/http';
import { AppStatus, PlayerInfo, ServerConfig } from '../models/models';

@Injectable({ providedIn: 'root' })
export class ApiService {
  private http = inject(HttpClient);

  status = httpResource<AppStatus>('/api/status');
  keys = httpResource<string[]>('/api/status/keys');
  apps = httpResource<string[]>('/api/apps');
  players = httpResource<PlayerInfo[]>('/api/players');
  config = httpResource<ServerConfig>('/api/config');

  pressKey(key: string) {
    return this.http.post('/api/remote/press', { key });
  }

  typeText(text: string) {
    return this.http.post('/api/remote/type', { text });
  }

  launchApp(pkg: string) {
    return this.http.post('/api/apps/launch', { package: pkg });
  }

  killApp(pkg: string) {
    return this.http.post('/api/apps/kill', { package: pkg });
  }

  setDefaultPlayer(pkg: string) {
    return this.http.put('/api/player/default', { package: pkg });
  }

  play(url: string, player?: string) {
    return this.http.post('/api/play', { url, player });
  }

  updateConfig(config: Partial<ServerConfig>) {
    return this.http.put('/api/config', config);
  }

  refresh() {
    this.status.refresh?.();
    this.keys.refresh?.();
    this.apps.refresh?.();
    this.players.refresh?.();
    this.config.refresh?.();
  }
}
```

- [ ] **Step 2: Create WebSocketService**

```typescript
import { Injectable, signal } from '@angular/core';
import { LogEntry } from '../models/models';

@Injectable({ providedIn: 'root' })
export class WebSocketService {
  private ws: WebSocket | null = null;
  logs = signal<LogEntry[]>([]);

  connect(url: string) {
    if (this.ws) this.ws.close();
    this.ws = new WebSocket(url);
    this.ws.onmessage = (event) => {
      const entry: LogEntry = JSON.parse(event.data);
      this.logs.update((prev) => [...prev.slice(-999), entry]);
    };
    this.ws.onclose = () => {
      setTimeout(() => this.connect(url), 3000);
    };
  }

  disconnect() {
    this.ws?.close();
    this.ws = null;
  }

  filter(opts: { level?: string; tag?: string; pid?: number }) {
    this.ws?.send(JSON.stringify({ type: 'filter', ...opts }));
  }
}
```

- [ ] **Step 3: Verify build**

Run: `npx ng build`
Expected: PASS

- [ ] **Step 4: Commit**

```
git add web/src/app/services/
git commit -m "feat: add ApiService and WebSocketService"
```

---

### Task 7: Angular UI pages (scaffolded in Task 5, now implement)

**Files:**
- Modify: all 5 page components + 4 sub-components in `web/src/app/`

- [ ] **Step 1: Remote page — DPAD + controls**

Write `web/src/app/pages/remote-page/remote-page.component.ts` with DPAD grid, action buttons, volume, playback controls calling ApiService.

- [ ] **Step 2: Apps page — package list**

Write `web/src/app/pages/apps-page/apps-page.component.ts` with app list, launch/kill buttons.

- [ ] **Step 3: Logs page — real-time viewer**

Write `web/src/app/pages/logs-page/logs-page.component.ts` with virtual-scroll log display connected to WebSocketService.

- [ ] **Step 4: Player page — player selection**

Write `web/src/app/pages/player-page/player-page.component.ts` with radio list, set default.

- [ ] **Step 5: Settings page**

Write `web/src/app/pages/settings-page/settings-page.component.ts` with TV IP form.

- [ ] **Step 6: Verify build**

Run: `npx ng build`
Expected: PASS

- [ ] **Step 7: Commit**

```
git add web/src/app/pages/
git commit -m "feat: add all UI pages (remote, apps, logs, player, settings)"
```

---

### Task 8: Build pipeline and integration

**Files:**
- Modify: `Makefile`
- Modify: `go.mod` (already done from Task 3)
- Modify: `.gitignore`

- [ ] **Step 1: Update Makefile**

Update full Makefile (keep existing targets alongside these):
```makefile
.PHONY: web serve all clean

web:
	cd web && npx ng build --configuration production

serve:
	go build -o tv.exe ./cmd/tv

all: serve

clean:
	del /f /q tv.exe 2>NUL || true
	if exist internal\serve\static rmdir /s /q internal\serve\static
```

Note: `web` builds Angular directly to `internal/serve/static/`. `serve` (without `web` dep) is for quick Go-only rebuilds after first `make web`. Run `make web serve` or `make all` for full build.

- [ ] **Step 2: Update .gitignore**

Add:
```
# Go build artifacts
tv.exe
tv-*-*

# Angular
web/node_modules/
web/.angular/

# Embedded Angular build (output from ng build)
internal/serve/static/
```

- [ ] **Step 3: Full build test**

Run: `make all`
Expected: Angular builds, Go embeds and builds, tv.exe produced

- [ ] **Step 4: Verify tv serve can start**

Run: `.\tv.exe serve --port 8080 --no-open &`
Expected: Server starts, press Ctrl-C to stop

- [ ] **Step 5: Commit**

```
git add Makefile .gitignore
git commit -m "feat: update build pipeline for Angular + Go embedding"
```

---

### Task 9: Final verification

- [ ] **Step 1: Run Go tests**

Run: `go test ./...`
Expected: All tests pass

- [ ] **Step 2: Run Angular build**

Run: `npx ng build --configuration production`
Expected: PASS

- [ ] **Step 3: Run integration**

Run: `.\tv.exe serve --port 8080 --no-open`
Expected: Starts, visit http://localhost:8080, API responds at /api/status

- [ ] **Step 4: Push to GitHub (if user confirms)**

```
git push origin master
```
