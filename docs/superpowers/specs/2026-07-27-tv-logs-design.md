# tv logs — Design Spec

**GitHub account:** mutedfalcontv
**Date:** 2026-07-27

## Goal

Add `tv logs` subcommand wrapping `adb logcat` for viewing, filtering, and debugging Android TV logs — app-specific, level-filtered, follow mode, and buffer management.

## Architecture

```cmd/tv/main.go      →  logsCmd handler
pkg/logcat/
  logcat.go         →  ArgsBuilder, PID resolution, Run
  logcat_test.go    →  tests with adb.MockRunner
```

Reuses `pkg/adb` (Runner interface) and `pkg/config`. The logcat package builds correct `adb shell logcat` arguments, resolves package names to PIDs for app filtering, and streams output to stdout.

## Command Set

```
tv logs                     # live follow (default)
tv logs -n 100              # last 100 lines, exit
tv logs -p <package>        # filter by app (com.brouken.player or "Just Player")
tv logs -l error            # filter by min level (v/d/i/w/e/f)
tv logs -t <tag>            # filter by log tag
tv logs -v threadtime       # verbose format with timestamps
tv logs -c                  # clear log buffer
tv logs dump                # one-shot full dump, no follow
```

Flags combine: `tv logs -p VLC -l error -n 50`

## Implementation

### pkg/logcat/logcat.go

```go
package logcat

type Options struct {
    Package string   // resolve to PID
    Level   string   // v/d/i/w/e/f
    Tag     string   // filter by tag
    Lines   int      // last N lines (-n)
    Format  string   // brief, threadtime, etc
    Clear   bool     // clear buffer
    Dump    bool     // one-shot, no follow
}

func Run(cfg *adb.Config, r adb.Runner, opts Options) error
func ResolvePID(cfg *adb.Config, r adb.Runner, pkg string) (int, error)
func BuildArgs(opts Options) []string
```

- `Run` builds logcat args, calls `r.ShellWithStderr()` or executes `adb logcat` directly for streaming
- `ResolvePID` runs `ps -A | grep <pkg>` and parses PID from output
- For live follow mode, uses `os/exec` directly (not through Shell wrapper) since it needs stdout streaming

### PID Resolution

```
adb shell ps -A | grep com.brouken.player
→ "u0_a123   4567   ... com.brouken.player"
→ PID = 4567
→ logcat --pid=4567
```

Fallback: if `ps -A` not available, try `ps | grep <pkg>`.

### Argument Building

| Option     | logcat args                            |
|-----------|----------------------------------------|
| Package   | `--pid=<PID>`                          |
| Level     | `*:<level>` (e.g. `*:E`)              |
| Tag       | `<tag>:*`                              |
| Lines     | `-t <lines>`                           |
| Format    | `-v <format>`                          |
| Clear     | `-b all -c`                            |
| Dump      | `-d` (dump and exit)                   |
| Default   | `-v brief` (follow mode)               |

### Streaming Mode

For live follow (default), `Run` uses `os/exec.Command(adb, "-s", ip, "shell", logcatArgs...)` directly with stdout connected to `os.Stdout` — real-time streaming. Ctrl-C stops.

For dump/clear/numbered modes (one-shot), uses `ShellWithStderr` and prints output.

**Testable split:** `BuildArgs` and `ResolvePID` are pure functions tested with MockRunner. `Run`'s streaming path is verified in integration tests only.

## Testing

8 tests in `pkg/logcat/logcat_test.go`:

1. TestBuildArgs_Default — no flags → `-v brief`
2. TestBuildArgs_WithLines — `-n 100` → `-t 100 -v brief`
3. TestBuildArgs_WithLevel — `-l error` → `*:E -v brief`
4. TestBuildArgs_WithPackage — resolves PID, adds `--pid=<PID>`
5. TestBuildArgs_Clear — `-c` → `-b all -c`
6. TestBuildArgs_Dump — `dump` → `-d -v brief`
7. TestBuildArgs_Combined — `-p VLC -l error -n 50` → `-t 50 *:E --pid=<PID> -v brief`
8. TestResolvePID — parses `ps -A` output, returns correct PID

## Files

| File | Action |
|------|--------|
| `pkg/logcat/logcat.go` | Create |
| `pkg/logcat/logcat_test.go` | Create |
| `cmd/tv/main.go` | Modify — add logsCmd |
