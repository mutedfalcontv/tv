# TV Remote UI — Design Spec

## Overview

A remote control web UI for Android TV, built as a PWA frontend (Angular) served by the existing Go CLI (`tv serve`). The Go binary embeds the Angular static build and serves both the API and the frontend on a single port. Phone/tablet browsers on the same network can access the UI via the laptop's IP.

## Architecture

```
tv serve --port 8080 [--host 0.0.0.0]
├── Go HTTP server (net/http)
│   ├── /api/*        → REST handlers → ADB → TV
│   ├── /ws/*         → WebSocket (logcat streaming)
│   └── /             → Static files (embedded Angular PWA)
├── Auto-opens http://localhost:8080 in browser
└── Phone: http://<laptop-ip>:8080
```

- Single Go binary. No Node/SSR server required in production.
- Same origin for API and frontend → no CORS config needed.
- Angular PWA with service worker for offline-capable install.

## Go Backend

### New package: `internal/serve/`

Files:
- `server.go` — HTTP server setup, route registration, static file serving
- `handlers.go` — Per-endpoint handler functions
- `websocket.go` — WebSocket handler for logcat streaming

### New command: `cmd/tv/serve.go`

```
tv serve [--port 8080] [--host 0.0.0.0] [--no-open]
```

Flags:
- `--port` (default 8080)
- `--host` (default 0.0.0.0)
- `--no-open` — skip opening browser (for headless/phone-only)

Behavior:
1. Load config via `config.Load()`
2. Ensure ADB connected via `adb.EnsureConnected()`
3. Start HTTP server on `host:port`
4. Open `http://localhost:port` in default browser (skip with `--no-open`)
5. Block until SIGINT/SIGTERM

### REST Endpoints

All responses JSON with `Content-Type: application/json`. Errors return `{"error": "..."}` with appropriate status code.

| Method | Path | Request | Response | Maps to |
|---|---|---|---|---|
| `GET` | `/api/status` | — | `{"connected": bool, "tv_ip": "..."}` | `adb.EnsureConnected` |
| `GET` | `/api/status/keys` | — | `["up", "down", ...]` | `remote.Keys()` |
| `POST` | `/api/remote/press` | `{"key": "home"}` | `{"ok": true}` | `remote.Press()` |
| `POST` | `/api/remote/type` | `{"text": "hello"}` | `{"ok": true}` | `remote.Type()` |
| `GET` | `/api/apps` | `?prefix=` (optional) | `["pkg1", "pkg2", ...]` | shell `pm list packages` |
| `POST` | `/api/apps/launch` | `{"package": "..."}` | `{"ok": true}` | shell `monkey -p ... 1` |
| `POST` | `/api/apps/kill` | `{"package": "..."}` | `{"ok": true}` | shell `am force-stop ...` |
| `GET` | `/api/players` | — | `[{"name": "VLC", "package": "..."}]` | `player.Detect()` |
| `PUT` | `/api/player/default` | `{"package": "..."}` | `{"ok": true}` | `player.Resolve()` + `config.Save()` |
| `POST` | `/api/play` | `{"url": "...", "player?": "..."}` | `{"ok": true}` | `player.PlayOnTV()` |
| `GET` | `/api/logs` | `?pkg=&level=&lines=&tag=` | log text | `logcat.RunOnce()` |
| `GET` | `/api/config` | — | `{"tv_ip": "...", "default_player": "..."}` | `config.Load()` |
| `PUT` | `/api/config` | `{"tv_ip": "...", "default_player": "..."}` | `{"ok": true}` | `config.Save()` |

### WebSocket

| Path | Protocol | Description |
|---|---|---|
| `/ws/logs` | Text frames (JSON) | Each frame: `{"line": "...", "timestamp": "..."}` |

Client sends optional filter message: `{"type": "filter", "level": "i", "tag": "...", "pid": 1234}`

### Embedded Static Files

Angular build output at `web/dist/tv/browser/` embedded via `//go:embed web/dist/tv/browser/*`. Served under `/` with proper MIME types and `Cache-Control: max-age=3600` for hashed assets.

## Angular Frontend

### Tech stack

- Angular latest version
- Standalone components (no NgModules)
- Signals for state management
- Tailwind CSS for styling
- PWA via `@angular/pwa`
- No SSR — pure static SPA

### Routes

| Path | Component | Description |
|---|---|---|
| `/` | `RemotePageComponent` | DPAD + action buttons + volume + playback |
| `/apps` | `AppsPageComponent` | App list with launch/kill |
| `/logs` | `LogsPageComponent` | Real-time logcat viewer |
| `/player` | `PlayerPageComponent` | Select default player |
| `/settings` | `SettingsPageComponent` | TV IP, server config |

### Component tree

```
AppComponent
├── LayoutComponent
│   ├── SidebarComponent (nav links)
│   ├── Router outlet
│   │   ├── RemotePageComponent
│   │   │   ├── DpadComponent (up/down/left/right/ok)
│   │   │   ├── ActionButtonsComponent (home/back/menu/power)
│   │   │   ├── VolumeComponent (vol up/down/mute)
│   │   │   └── PlaybackComponent (play/pause/stop/ff/rew)
│   │   ├── AppsPageComponent
│   │   │   └── AppCardComponent
│   │   ├── LogsPageComponent (virtual scroll + WebSocket)
│   │   ├── PlayerPageComponent (radio list)
│   │   └── SettingsPageComponent (form)
│   └── StatusBarComponent (connection status)
```

### Services

| Service | Responsibility |
|---|---|
| `ApiService` | HTTP calls to `/api/*`, error handling, loading state |
| `WebSocketService` | Connect/reconnect to `/ws/logs`, message stream via signal |
| `ConfigService` | Holds config + connection state as signals |

### PWA

- Service worker via `@angular/pwa`
- `manifest.webmanifest` with app name "TV Remote", theme color
- Install prompt on supported browsers
- Offline fallback shows cached remote page

## Project Structure

```
tv/
├── cmd/tv/
│   ├── main.go
│   └── serve.go               ← NEW: tv serve command
├── internal/
│   ├── serve/
│   │   ├── server.go           ← NEW: HTTP server, routing, static embedding
│   │   ├── handlers.go         ← NEW: API handlers
│   │   └── websocket.go        ← NEW: logcat WebSocket
│   ├── adb/
│   ├── remote/
│   ├── logcat/
│   ├── player/
│   └── config/
├── web/                         ← NEW: Angular project
│   ├── src/
│   │   ├── app/
│   │   │   ├── components/
│   │   │   ├── pages/
│   │   │   ├── services/
│   │   │   └── models/
│   │   ├── index.html
│   │   └── manifest.webmanifest
│   ├── angular.json
│   ├── package.json
│   ├── tsconfig.json
│   └── tailwind.config.js
├── go.mod
├── Makefile
└── .gitignore
```

## Build & Run

```makefile
.PHONY: web serve all

web:
	cd web && npx ng build --configuration production

serve:
	go build -o tv.exe ./cmd/tv

all: web serve

# Run:
#   tv serve --port 8080
# Open:
#   http://localhost:8080 (desktop)
#   http://<laptop-ip>:8080 (phone/tablet)
```

## Error Handling

- Server errors: JSON `{"error": "..."}` with HTTP 4xx/5xx
- WebSocket: Close frame on internal error with status code 1011
- Frontend: Toast notifications for API errors, reconnection banner for WS disconnect
- ADB connection: `/api/status` returns `connected: false`; frontend shows "Disconnected" banner with Connect button

## Security

- No auth for LAN-only usage (same network as ADB)
- `--host 0.0.0.0` to allow phone access; `--host 127.0.0.1` for laptop-only
- CORS: restricted to same origin (or configurable for development)

## Out of Scope (v1)

- User authentication
- Multiple TV support
- ADB over USB
- Drag-and-drop file transfer
- Voice search
