# tv remote — Design Spec

**GitHub account:** mutedfalcontv
**Date:** 2026-07-27

## Goal

Add `tv remote` subcommand to the `tv` CLI for full Android TV remote control via ADB keyevents. Reuses existing `pkg/adb` (Runner interface + MockRunner) and `pkg/config`.

## Architecture

```
cmd/tv/main.go      →  remoteCmd handler (flat dispatch)
pkg/remote/
  remote.go         →  Press(), Type(), key map
  remote_test.go    →  tests with adb.MockRunner
```

`tv remote` piggybacks on the `Runner` interface — no new ADB logic.

## Command Set (Flat)

### Navigation
```
tv remote up
tv remote down
tv remote left
tv remote right
tv remote ok
```

### System
```
tv remote home
tv remote back
tv remote menu
tv remote power
tv remote input
tv remote info
tv remote subtitle
```

### Volume
```
tv remote volup
tv remote voldown
tv remote mute
```

### Media
```
tv remote play
tv remote pause
tv remote stop
tv remote ff
tv remote rew
tv remote next
tv remote prev
```

### Channel
```
tv remote chup
tv remote chdown
```

### Numbers
```
tv remote 0
tv remote 1
...
tv remote 9
```

### Text input
```
tv remote type "search query"
```

## Key Mapping

Each command maps to an Android `KEYCODE_*` constant sent via `input keyevent`:

| Command    | ADB Keyevent            |
|-----------|------------------------|
| up        | KEYCODE_DPAD_UP        |
| down      | KEYCODE_DPAD_DOWN      |
| left      | KEYCODE_DPAD_LEFT      |
| right     | KEYCODE_DPAD_RIGHT     |
| ok        | KEYCODE_DPAD_CENTER    |
| home      | KEYCODE_HOME           |
| back      | KEYCODE_BACK           |
| menu      | KEYCODE_MENU           |
| power     | KEYCODE_POWER          |
| input     | KEYCODE_TV_INPUT       |
| info      | KEYCODE_INFO           |
| subtitle  | KEYCODE_CAPTIONS       |
| volup     | KEYCODE_VOLUME_UP      |
| voldown   | KEYCODE_VOLUME_DOWN    |
| mute      | KEYCODE_VOLUME_MUTE    |
| play      | KEYCODE_MEDIA_PLAY     |
| pause     | KEYCODE_MEDIA_PAUSE    |
| stop      | KEYCODE_MEDIA_STOP     |
| ff        | KEYCODE_MEDIA_FAST_FORWARD |
| rew       | KEYCODE_MEDIA_REWIND   |
| next      | KEYCODE_MEDIA_NEXT     |
| prev      | KEYCODE_MEDIA_PREVIOUS |
| chup      | KEYCODE_CHANNEL_UP     |
| chdown    | KEYCODE_CHANNEL_DOWN   |
| 0-9       | KEYCODE_0 — KEYCODE_9  |
| type      | `input text <escaped>` |

## Implementation

### pkg/remote/remote.go

```go
package remote

type KeyMap map[string]string

var keys = KeyMap{...}  // all key name → KEYCODE mappings

func Press(cfg *adb.Config, r adb.Runner, key string) error
func Type(cfg *adb.Config, r adb.Runner, text string) error
func Keys() []string  // returns sorted key names for error message
```

- `Press` looks up key in map, calls `r.Shell(cfg, "input keyevent "+keycode)`
- `Type` calls `r.Shell(cfg, "input text "+escapeForADBShell(text))` — wraps text in single quotes so Android shell doesn't split on spaces. Single quotes in text escaped via `'\\''` (same as `EscapeURLForShell` in player package)
- Unknown key returns error listing available keys

### MockRunner changes (pkg/adb/runner.go)

Add `LastShellCmd string` field to `MockRunner`. `MockRunner.Shell()` captures the `cmd` argument so `Press()` and `Type()` can verify correct ADB command.

### CLI handler (cmd/tv/main.go)

Add `remoteCmd` to commands slice:
```go
{"remote", "TV remote control", remoteCmd},
```

Flat dispatch:
```go
func remoteCmd(args []string) {
    if len(args) < 1 { usage + exit }
    cmd := args[0]
    if cmd == "type" {
        if len(args) < 2 { usage + exit }
        remote.Type(cfg, r, args[1])
    } else if isNumber(cmd) {
        remote.Press(cfg, r, cmd)
    } else {
        remote.Press(cfg, r, cmd)
    }
}
```

## Testing

8 tests in `pkg/remote/remote_test.go` using `adb.MockRunner`:

1. TestPress_NavKey — up sends KEYCODE_DPAD_UP
2. TestPress_VolumeKey — volup sends KEYCODE_VOLUME_UP
3. TestPress_MediaKey — play sends KEYCODE_MEDIA_PLAY
4. TestPress_NumberKey — "5" sends KEYCODE_5
5. TestPress_UnknownKey — returns error
6. TestType_SendsInputText — type "hello" sends `input text hello`
7. TestType_WithSpaces — type "search term" sends `input text 'search term'`
8. TestKeys_ReturnsSorted — Keys() returns sorted slice

## Files Changed/Created

| File | Action |
|------|--------|
| `pkg/remote/remote.go` | Create |
| `pkg/remote/remote_test.go` | Create |
| `pkg/adb/runner.go` | Modify — add `LastShellCmd` to MockRunner |
| `cmd/tv/main.go` | Modify — add remoteCmd |
