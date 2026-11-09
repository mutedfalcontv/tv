# AGENTS.md

Guidance for AI coding agents working in this repository. Read this before changing code.
Frontend-specific rules live in [`web/AGENTS.md`](./web/AGENTS.md).

## What this project is

`tv` is a Go command-line tool plus an embedded Angular progressive web app that controls an
Android TV over ADB. It can connect to the TV, list/launch/kill apps, send remote key events,
type text, play a media URL in a chosen video player, and read/filter logcat. `tv serve` exposes
the same capabilities over a small JSON HTTP API and serves the web UI from the same binary.

Module: `github.com/mutedfalcontv/tv`

## Toolchain and requirements

- Go (see `go.mod`; currently Go 1.26.x)
- Node.js + npm (see `web/package.json`; Angular 22, npm 10.9.8)
- `adb` (Android platform tools) must be on `PATH` for any TV operation
- A reachable Android TV with ADB debugging enabled

## Build, test, run

```sh
# Build web assets, then the Go binary (output: tv.exe on Windows)
make build

# Build only the web app (writes internal/serve/static via Angular output)
make web

# Go tests (verbose) / plain
make test
go test ./...

# Compile all packages
go build ./...

# Build the CLI explicitly (Windows)
go build -o tv.exe ./cmd/tv/
```

Notes:
- `make build` depends on `web`, so the Angular bundle is rebuilt before the Go build. The
  Angular output is embedded into `internal/serve/static`.
- Web unit tests use Vitest via `cd web && npm test`.
- Cross-compile targets are in the `build-all` Make target.

## Architecture

| Path | Responsibility |
| --- | --- |
| `cmd/tv/` | CLI entry point. `main.go` defines the command table, dispatches subcommands, and handles `help`. `serve.go` implements `tv serve` (flags, signals, browser open). |
| `internal/adb/` | ADB abstraction. `Runner` interface + `RealRunner` (executes `adb`) + `MockRunner` (test double). Also `Config{TVIP, DefaultPlayer}` and `EnsureConnected`. |
| `internal/config/` | Persisted user config at `~/.config/tv/config.json` (JSON: `tv_ip`, `device`, `ip`, `default_player`, `token`, `cors_origins`). `Load`, `Save`, default TV IP. |
| `internal/device/` | Screenshot capture (`CaptureScreenshot`, `NormalizePNG`, `WriteScreenshot`) and live state (`GetState`). |
| `internal/discovery/` | mDNS wireless-debugging discovery behind `tv discover`. |
| `internal/mcp/` | stdio MCP server exposing TV tools to AI agents (`tv mcp`). |
| `internal/output/` | JSON/human output helpers used by the CLI (`--json`). |
| `internal/localserve/` | Ephemeral HTTP server used by `tv play <local-file>`. |
| `internal/player/` | Video player detection (`Detect`, `Resolve`) and media playback via an `am start` intent (`PlayOnTV`, MIME guessing). |
| `internal/remote/` | Remote key map (`Press`), text input (`Type`), and `Keys()`. |
| `internal/logcat/` | logcat argument construction (`BuildArgs`), PID resolution, one-shot and streaming capture. |
| `internal/serve/` | HTTP server: route registration (`server.go`), handlers (`handlers.go`), WebSocket log streaming (`websocket.go`), embedded static Angular bundle, and the OpenAPI document (`openapi.go` + `openapi.json`). |
| `internal/help/` | Machine-readable CLI command tree used by `tv help --json` / `tv --help --json`. |
| `web/` | Angular PWA source. Build output is copied/generated into `internal/serve/static` and embedded. |
| `docs/` | Workflow notes and per-feature plans. |

Data flow: CLI or HTTP handler → `adb.Runner` → `adb` process on the TV. Player and remote
logic both go through `adb.Runner`, which is why tests can substitute `adb.MockRunner`.

## Conventions

- Follow the existing standard-library style; keep packages small and single-purpose.
- Do not add comments unless the logic is not self-evident. Existing code is intentionally
  comment-light.
- Depend on the `adb.Runner` interface, not `RealRunner`, in business logic so tests stay fast.
- Tests use table-driven / focused subtests where it helps, and `adb.MockRunner` for TV I/O.
  Config tests redirect `HOME`/`USERPROFILE` to `t.TempDir()`.
- CLI errors print to `stderr` and exit non-zero; success output goes to `stdout`.
- HTTP handlers use `jsonOK` / `jsonError`; failures return `{"error": "..."}` with a proper
  status code (400 invalid input, 500 ADB/operation failure, 503 TV unreachable).
- New HTTP routes must be registered in `internal/serve/server.go` and reflected in
  `internal/serve/openapi.json`.
- New CLI commands must be added to the `commands` table in `cmd/tv/main.go` and to the tree in
  `internal/help/help.go` so JSON help stays accurate.
- Git: `master` is the ground-truth branch; feature work happens in worktrees (see
  `docs/WORKFLOW.md`). Do not commit directly to `master` for non-trivial changes.

## Using the CLI

```sh
tv adb connect                 # connect using configured/TV_IP address
tv app list                    # list installed packages
tv app launch com.brouken.player
tv app kill com.brouken.player # destructive: force-stop
tv remote home                 # send a named key
tv remote type "hello world"
tv play https://example.com/video.m3u8
tv player list                 # detect players, * marks default
tv player default "Just Player" # or the package name
tv logs -p vlc -l e -n 50      # filtered snapshot; omit filters to stream
tv config                      # show current config
tv serve --port 8080 --no-open # start UI + API
```

`TV_IP` environment variable overrides the configured TV address for `tv adb`.
`tv help --json` (or `tv --help --json`) emits the full command tree as JSON for automation.

## Using the HTTP API

Start it with `tv serve`. All responses are JSON. It binds `127.0.0.1` by default with no auth;
when `--token` is set, send `Authorization: Bearer <token>` (the bundled UI receives a
`tv_token` cookie, and WebSocket clients may use `?token=`). CORS reflects only loopback and
configured origins, and cross-origin mutations are rejected. Current routes:

| Method | Path | Purpose |
| --- | --- | --- |
| GET | `/api/status` | `{connected, tv_ip}` |
| GET | `/api/status/keys` | sorted remote key names |
| POST | `/api/remote/press` | `{key}` |
| POST | `/api/remote/type` | `{text}` |
| GET | `/api/apps` | installed package names |
| POST | `/api/apps/launch` | `{package}` |
| POST | `/api/apps/kill` | `{package}` (destructive) |
| GET | `/api/players` | `[{PackageName, Name}]` |
| PUT | `/api/player/default` | `{package}` (name or package) |
| POST | `/api/play` | `{url, player?}` |
| GET | `/api/logs` | query: `level`, `tag`, `format`, `lines`, `pkg`, `dump` |
| GET | `/api/screenshot` | PNG screenshot of the TV (`image/png`) |
| GET | `/api/state` | `{connected, tv_ip, foreground_app?, foreground_activity?, volume, muted, screen_on, wakefulness?}` |
| GET | `/api/config` | `{tv_ip, device?, ip?, default_player}` |
| PUT | `/api/config` | `{tv_ip?, device?, ip?, default_player?}` |
| GET | `/api/openapi.json` | OpenAPI 3.0 description of this API |
| WS | `/ws/logs` | stream `{line, timestamp}` frames; send `{type:"filter", level?, tag?, pid?}` |

Example:

```sh
curl -s localhost:8080/api/status
curl -s -X POST localhost:8080/api/remote/press -H 'Content-Type: application/json' -d '{"key":"home"}'
curl -s -X POST localhost:8080/api/play -H 'Content-Type: application/json' -d '{"url":"https://example.com/v.mp4"}'
```

The canonical machine-readable contract is `GET /api/openapi.json` (hand-written, embedded from
`internal/serve/openapi.json`). Keep it in sync with `server.go`.

## MCP server

`tv mcp` starts a stdio MCP server for AI agents. It exposes `tv_status`, `tv_connect`,
`tv_disconnect`, `tv_list_apps`, `tv_launch_app`, `tv_kill_app`, `tv_press_key`,
`tv_type_text`, `tv_play`, `tv_players`, `tv_logs` (bounded snapshot; never streams), and
`tv_screenshot`. Tools reuse `internal/adb.Runner`, honour `--device`/`--ip`, and `-timeout`
bounds each ADB call.

## Safety

- `tv app kill`, `/api/apps/kill`, and player/config changes mutate the TV or persisted state.
  Confirm intent before running them against a real device.
- `tv serve` binds `127.0.0.1:8080` by default. It has no authentication unless `--token` is
  set; CORS reflects only loopback/configured origins, and cross-origin mutations are rejected.
  Use `--host 0.0.0.0 --token <secret>` when exposing it on a network.
- Do not commit secrets or machine-specific config. `~/.config/tv/config.json` is outside the repo.
