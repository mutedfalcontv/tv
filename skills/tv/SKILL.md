---
name: tv
description: Control an Android TV over ADB with the `tv` CLI or its HTTP API. Use when connecting to a TV, listing/launching/killing apps, sending remote keys, typing text, playing a media URL, reading logcat, or setting a default video player.
---

# Control an Android TV with `tv`

`tv` is a Go CLI (and optional HTTP API via `tv serve`) that drives an Android TV through
`adb`. It also ships a machine-readable command tree: **`tv help --json`** (or
`tv --help --json`). Prefer that output over guessing flags.

## Prerequisites

- `adb` is on `PATH`.
- The TV is reachable and ADB debugging is enabled (usually port `5555`).
- Configuration lives at `~/.config/tv/config.json`:

```json
{ "tv_ip": "192.168.2.3:5555", "default_player": "com.brouken.player" }
```

`TV_IP` overrides the configured address for `tv adb`.

## Discover the interface

```sh
tv help --json          # full command tree: subcommands, flags, arguments as JSON
tv --help --json        # alias
```

The JSON document has `{name, version, description, commands[]}`; each command has
`{name, description, usage, arguments[], flags[], subcommands[]}`. Remote keys are enumerated
under the `remote` command's `key` argument `choices`.

## Connect

```sh
tv adb connect          # connect to the configured TV
tv adb disconnect
```

## Apps

```sh
tv app list                        # installed packages (one per line, "package:..." on TV)
tv app launch com.brouken.player   # start an app
tv app kill com.brouken.player     # DESTRUCTIVE: force-stop
```

## Remote control

```sh
tv remote home
tv remote ok
tv remote volup
tv remote 5
tv remote type "hello world"
```

Available keys: `up down left right ok home back menu power input info subtitle volup voldown
mute play pause stop ff rew next prev chup chdown 0-9`. `tv remote type <text>` types into the
focused field.

## Players and playback

```sh
tv player list                      # detected players; "*" marks the default
tv player default "Just Player"     # accepts display name or package
tv play https://example.com/video.m3u8
tv play -player vlc https://example.com/v.mp4
tv play -ip 192.168.2.50:5555 https://example.com/v.mp4
```

MIME type is inferred from the URL extension (m3u8, mp4, mkv, avi, mov, webm; default mp4) and
opened with an `am start` VIEW intent in the chosen player.

## Logs

```sh
tv logs -p vlc -l e -n 50     # snapshot: package, level=error, last 50 lines
tv logs -t MyTag -n 100       # filter by tag
tv logs -d                    # dump buffer and exit
tv logs                       # stream until Ctrl-C
```

Flags: `-p` package, `-l` level (v/d/i/w/e/f), `-t` tag, `-n` line count, `-v` format,
`-c` clear, `-d` dump. Passing `-n`, `-l`, `-t`, or `-d` makes it a one-shot call; otherwise
it streams.

## HTTP API (`tv serve`)

Start the UI + API: `tv serve --port 8080 --no-open`. The OpenAPI 3.0 contract is at
`GET /api/openapi.json`. Key endpoints and shapes:

| Method | Path | Body / Query | Response |
| --- | --- | --- | --- |
| GET | `/api/status` | — | `{ "connected": bool, "tv_ip": string }` |
| GET | `/api/status/keys` | — | `["back","down", ...]` |
| POST | `/api/remote/press` | `{ "key": "home" }` | `{ "ok": true }` |
| POST | `/api/remote/type` | `{ "text": "hello" }` | `{ "ok": true }` |
| GET | `/api/apps` | — | `["com.brouken.player", ...]` |
| POST | `/api/apps/launch` | `{ "package": "com.brouken.player" }` | `{ "ok": true }` |
| POST | `/api/apps/kill` | `{ "package": "com.brouken.player" }` | `{ "ok": true }` |
| GET | `/api/players` | — | `[{ "PackageName": "com.brouken.player", "Name": "Just Player" }]` |
| PUT | `/api/player/default` | `{ "package": "Just Player" }` | `{ "ok": true }` |
| POST | `/api/play` | `{ "url": "https://...", "player": "com.brouken.player" }` | `{ "ok": true }` |
| GET | `/api/logs` | `?level=&tag=&format=&lines=&pkg=&dump=` | `{ "logs": "..." }` |
| GET | `/api/screenshot` | — | PNG image (`image/png`) |
| GET | `/api/state` | — | `{ "connected": bool, "tv_ip": "...", "foreground_app"?, "volume": int, "muted": bool, "screen_on": bool, "wakefulness"? }` |
| GET | `/api/config` | — | `{ "tv_ip": "...", "device"?, "ip"?, "default_player": "..." }` |
| PUT | `/api/config` | `{ "tv_ip"?, "device"?, "ip"?, "default_player"? }` | `{ "ok": true }` |
| GET | `/api/openapi.json` | — | OpenAPI document |
| WS | `/ws/logs` | — | frames `{ "line": "...", "timestamp": "..." }` |

WebSocket clients may send `{ "type": "filter", "level"?, "tag"?, "pid"? }` to update filters
while streaming.

All errors are `{ "error": "..." }` with 400 (invalid input), 500 (operation failed), or 503
(TV unreachable).

Example calls:

```sh
curl -s localhost:8080/api/status
curl -s -X POST localhost:8080/api/remote/press -H 'Content-Type: application/json' -d '{"key":"home"}'
curl -s -X POST localhost:8080/api/play -H 'Content-Type: application/json' -d '{"url":"https://example.com/v.mp4"}'
```

## MCP server (for AI agents)

`tv mcp` starts a stdio MCP server. Tools: `tv_status`, `tv_connect`, `tv_disconnect`,
`tv_list_apps`, `tv_launch_app`, `tv_kill_app`, `tv_press_key`, `tv_type_text`, `tv_play`,
`tv_players`, `tv_logs` (bounded snapshot), `tv_screenshot`. Register it in an MCP client as a
stdio command, e.g. `{ "command": "tv", "args": ["mcp"] }`. It honours `--device`/`--ip` and
`-timeout`.

## Safety

- **Destructive:** `tv app kill` and `POST /api/apps/kill` force-stop a package. Confirm intent
  before running against a real device.
- **State-changing:** `tv player default` and `PUT /api/config` persist changes to
  `~/.config/tv/config.json`.
- **Network exposure:** `tv serve` binds `127.0.0.1:8080` by default. It has no authentication
  unless `--token` is set; CORS reflects only loopback/configured origins. For LAN exposure use
  `--host 0.0.0.0 --token <secret>`.
