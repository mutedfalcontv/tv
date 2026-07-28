This file is a merged representation of the entire codebase, combined into a single document by Repomix.

# File Summary

## Purpose
This file contains a packed representation of the entire repository's contents.
It is designed to be easily consumable by AI systems for analysis, code review,
or other automated processes.

## File Format
The content is organized as follows:
1. This summary section
2. Repository information
3. Directory structure
4. Repository files (if enabled)
5. Multiple file entries, each consisting of:
  a. A header with the file path (## File: path/to/file)
  b. The full contents of the file in a code block

## Usage Guidelines
- This file should be treated as read-only. Any changes should be made to the
  original repository files, not this packed version.
- When processing this file, use the file path to distinguish
  between different files in the repository.
- Be aware that this file may contain sensitive information. Handle it with
  the same level of security as you would the original repository.

## Notes
- Some files may have been excluded based on .gitignore rules and Repomix's configuration
- Binary files are not included in this packed representation. Please refer to the Repository Structure section for a complete list of file paths, including binary files
- Files matching patterns in .gitignore are excluded
- Files matching default ignore patterns are excluded
- Files are sorted by Git change count (files with more changes are at the bottom)

# Directory Structure
````
cmd/
  tv/
    main.go
docs/
  superpowers/
    plans/
      2026-07-27-tv-play.md
      2026-07-28-tv-monorepo-finalize.md
    specs/
      2026-07-27-tv-logs-design.md
      2026-07-27-tv-remote-design.md
  WORKFLOW.md
internal/
  adb/
    runner_test.go
    runner.go
  config/
    config_test.go
    config.go
  logcat/
    logcat_test.go
    logcat.go
  player/
    detect_test.go
    detect.go
    play_test.go
    play.go
  remote/
    remote_test.go
    remote.go
.gitignore
go.mod
Makefile
````

# Files

## File: docs/superpowers/plans/2026-07-28-tv-monorepo-finalize.md
````markdown
# tv Monorepo Finalize — Structural Fix & Bug Fix Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix Go monorepo structure (pkg/ → internal/), repair 3 bugs, fix 5 architecture issues, resolve edge cases — deliver final working repo ready for worktree use.

**Architecture:** Single-module Go monorepo: `github.com/mutedfalcontv/tv` with `cmd/tv/main.go` entrypoint, all internal packages under `internal/`, compiler-enforced privacy via `internal/` directory. No bare-repo worktree layering — standard `.git/` at root.

**Tech Stack:** Go 1.26 (stdlib only: flag, os/exec, encoding/json, testing, strings, path/filepath, fmt)

**Current tree:**
```
tv/
├── .git/
├── go.mod                    # module github.com/mutedfalcontv/tv
├── cmd/tv/main.go            # CLI entry (7 subcommands)
├── pkg/adb/                  # → move to internal/adb/
├── pkg/config/               # → move to internal/config/
├── pkg/logcat/               # → move to internal/logcat/
├── pkg/player/               # → move to internal/player/
├── pkg/remote/               # → move to internal/remote/
├── docs/
│   ├── WORKFLOW.md
│   └── superpowers/
│       ├── plans/2026-07-27-tv-play.md
│       └── specs/{tv-logs,tv-remote}-design.md
├── Makefile
└── .gitignore
```

---

### Bugs Found (3)

| # | Bug | Location | Impact |
|---|-----|----------|--------|
| B1 | `-p` flag in `tv logs` resolves PID but never passes it to logcat filter | `cmd/tv/main.go:338-354` + `pkg/logcat/logcat.go:27-60` | `-p` flag completely non-functional |
| B2 | `RunStream()` sets `Stdout = nil` instead of `os.Stdout` | `pkg/logcat/logcat.go:95` | Streaming mode shows no output |
| B3 | `remote.Type()` doesn't escape single quotes in text | `pkg/remote/remote.go:59` | Text with `'` breaks ADB shell |

### Architecture Issues (5)

| # | Issue | Location | Fix |
|---|-------|----------|-----|
| A1 | `pkg/adb` has duplicate `loadConfig()` instead of importing `config.Load()` | `pkg/adb/runner.go:100-123` | Remove, use `config.Load()` |
| A2 | `pkg/logcat` has duplicate `adbPath()` instead of using `RealRunner.binary()` | `pkg/logcat/logcat.go:80-86` | Remove, use `RealRunner` |
| A3 | `pkg/logcat.RunStream()` bypasses `Runner` interface | `pkg/logcat/logcat.go:88-97` | Refactor to use `RealRunner` |
| A4 | `ShellWithStderr` on MockRunner doesn't capture `LastShellCmd` | `pkg/adb/runner.go:161` | Add capture |
| A5 | `Makefile clean` uses `rm -f` (Unix) on Windows | `Makefile:17` | Use `del /f /q` for Windows |

### Structure Changes

| # | Change | Files Affected |
|---|--------|---------------|
| S1 | Move `pkg/*` → `internal/*` | All import paths update |
| S2 | Add `.gitignore` entries for build artifacts | `.gitignore` |

---
---

### Task 1: Move pkg/* → internal/* (structural fix)

**Files:**
- Rename: `pkg/adb/` → `internal/adb/`
- Rename: `pkg/config/` → `internal/config/`
- Rename: `pkg/logcat/` → `internal/logcat/`
- Rename: `pkg/player/` → `internal/player/`
- Rename: `pkg/remote/` → `internal/remote/`
- Modify: `cmd/tv/main.go` (import paths)
- Modify: `Makefile` (no change — `go build ./...` auto-resolves)

- [ ] **Step 1: Rename directories via git mv**

```bash
cd C:\Users\ajink\Projects\tv
git mv pkg internal
```

This renames the entire `pkg/` directory to `internal/`. All files move with history preserved.

- [ ] **Step 2: Update import paths in cmd/tv/main.go**

```go
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/mutedfalcontv/tv/internal/adb"
	"github.com/mutedfalcontv/tv/internal/config"
	"github.com/mutedfalcontv/tv/internal/logcat"
	"github.com/mutedfalcontv/tv/internal/player"
	"github.com/mutedfalcontv/tv/internal/remote"
)
```

- [ ] **Step 3: Verify build**

Run: `cd C:\Users\ajink\Projects\tv && go build ./...`
Expected: no errors (all internal packages resolved).

- [ ] **Step 4: Run all tests**

Run: `cd C:\Users\ajink\Projects\tv && go test ./... -v -count=1`
Expected: all 30 tests PASS with new import paths.

- [ ] **Step 5: Commit**

```bash
cd C:\Users\ajink\Projects\tv
git add internal/ cmd/tv/main.go
git commit -m "refactor: move pkg/ to internal/ for Go monorepo structure"
```

---

### Task 2: Fix duplicate loadConfig() in adb package

**Files:**
- Modify: `internal/adb/runner.go`
- Modify: `internal/adb/runner_test.go`

- [ ] **Step 1: Remove duplicate loadConfig() and GetTVIP, import config package**

Current `internal/adb/runner.go` has `loadConfig()` (lines 100-123) and `GetTVIP()` (lines 89-98) that duplicate `pkg/config.Load()`. Replace `GetTVIP()` with a simpler function that uses `config.Load()`.

Replace the entire `GetTVIP()` and `loadConfig()` functions:

```go
import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mutedfalcontv/tv/internal/config"
)
```

Replace `GetTVIP` and `loadConfig`:

```go
func GetTVIP() string {
	if v := os.Getenv("TV_IP"); v != "" {
		return v
	}
	cfg, err := config.Load()
	if err != nil {
		return "192.168.2.3:5555"
	}
	return cfg.TVIP
}
```

Delete the entire `loadConfig()` function (lines 100-123).

- [ ] **Step 2: Verify build**

Run: `cd C:\Users\ajink\Projects\tv && go build ./...`
Expected: no errors.

- [ ] **Step 3: Run tests**

Run: `cd C:\Users\ajink\Projects\tv && go test ./internal/adb/ -v`
Expected: 5 tests PASS.

- [ ] **Step 4: Commit**

```bash
cd C:\Users\ajink\Projects\tv
git add internal/adb/runner.go
git commit -m "fix: remove duplicate loadConfig(), use config.Load() instead"
```

---

### Task 3: Fix ShellWithStderr capture on MockRunner

**Files:**
- Modify: `internal/adb/runner.go:161`

- [ ] **Step 1: Add LastShellCmd capture to ShellWithStderr**

Current MockRunner.ShellWithStderr doesn't capture the command. Fix so tests can verify ADB commands sent via ShellWithStderr:

```go
func (m *MockRunner) ShellWithStderr(cfg *Config, cmd string) (string, error) {
	m.LastShellCmd = cmd
	return m.ShellWithStderrOut, m.ShellWithStderrErr
}
```

- [ ] **Step 2: Write failing test**

```go
func TestMockRunnerShellWithStderrCapture(t *testing.T) {
	m := &MockRunner{ShellWithStderrOut: "output"}
	cfg := &Config{TVIP: "192.168.2.3:5555"}
	out, err := m.ShellWithStderr(cfg, "some command")
	if err != nil {
		t.Fatalf("ShellWithStderr() unexpected error: %v", err)
	}
	if out != "output" {
		t.Errorf("ShellWithStderr() = %q, want %q", out, "output")
	}
	if m.LastShellCmd != "some command" {
		t.Errorf("LastShellCmd = %q, want %q", m.LastShellCmd, "some command")
	}
}
```

Add this test function to `internal/adb/runner_test.go`.

- [ ] **Step 3: Run test to verify it passes**

Run: `cd C:\Users\ajink\Projects\tv && go test ./internal/adb/ -v -run TestMockRunnerShellWithStderrCapture`
Expected: PASS.

- [ ] **Step 4: Run all adb tests**

Run: `cd C:\Users\ajink\Projects\tv && go test ./internal/adb/ -v`
Expected: 6 tests PASS (5 existing + 1 new).

- [ ] **Step 5: Commit**

```bash
cd C:\Users\ajink\Projects\tv
git add internal/adb/runner.go internal/adb/runner_test.go
git commit -m "fix: capture LastShellCmd in MockRunner.ShellWithStderr"
```

---

### Task 4: Fix remote.Type() single-quote escaping

**Files:**
- Modify: `internal/remote/remote.go`
- Modify: `internal/remote/remote_test.go`

- [ ] **Step 1: Write failing test for single-quote in text**

```go
func TestType_WithSingleQuote(t *testing.T) {
	m := &adb.MockRunner{}
	cfg := &adb.Config{TVIP: "192.168.2.3:5555"}
	err := Type(cfg, m, "it's")
	if err != nil {
		t.Fatalf("Type() unexpected error: %v", err)
	}
	// Single quote must be escaped: ' → '\''
	want := "input text 'it'\\''s'"
	if m.LastShellCmd != want {
		t.Errorf("got %q, want %q", m.LastShellCmd, want)
	}
}
```

Add to `internal/remote/remote_test.go`.

- [ ] **Step 2: Run test to verify it fails**

Run: `cd C:\Users\ajink\Projects\tv && go test ./internal/remote/ -v -run TestType_WithSingleQuote`
Expected: FAIL — Type() doesn't escape quotes.

- [ ] **Step 3: Fix Type() to escape single quotes**

Current:
```go
func Type(cfg *adb.Config, r adb.Runner, text string) error {
	escaped := "'" + text + "'"
	_, err := r.Shell(cfg, "input text "+escaped)
```

Replace with proper escaping:
```go
func Type(cfg *adb.Config, r adb.Runner, text string) error {
	escaped := "'" + strings.ReplaceAll(text, "'", "'\\''") + "'"
	_, err := r.Shell(cfg, "input text "+escaped)
```

Add `"strings"` to imports.

- [ ] **Step 4: Run test to verify it passes**

Run: `cd C:\Users\ajink\Projects\tv && go test ./internal/remote/ -v -run TestType_WithSingleQuote`
Expected: PASS.

- [ ] **Step 5: Run all remote tests**

Run: `cd C:\Users\ajink\Projects\tv && go test ./internal/remote/ -v`
Expected: 10 tests PASS (9 existing + 1 new).

- [ ] **Step 6: Commit**

```bash
cd C:\Users\ajink\Projects\tv
git add internal/remote/remote.go internal/remote/remote_test.go
git commit -m "fix: escape single quotes in remote.Type() for ADB shell"
```

---

### Task 5: Fix logcat package — bugs B1 and B2

**Files:**
- Modify: `internal/logcat/logcat.go`
- Modify: `internal/logcat/logcat_test.go`
- Modify: `cmd/tv/main.go`

#### Sub-task 5a: Fix BuildArgs to support --pid filter

- [ ] **Step 1: Add Package field handling to BuildArgs and fix RunStream**

`internal/logcat/logcat.go` has `Options.Package` field but `BuildArgs()` doesn't generate `--pid=<PID>` from it. Also `RunStream()` sets `Stdout = nil` instead of `os.Stdout`.

Replace the entire file content:

```go
package logcat

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/mutedfalcontv/tv/internal/adb"
)

type Options struct {
	Package string
	Level   string
	Tag     string
	Lines   int
	Format  string
	Clear   bool
	Dump    bool
}

var levels = map[string]string{
	"v": "*:V", "d": "*:D", "i": "*:I", "w": "*:W", "e": "*:E", "f": "*:F",
	"verbose": "*:V", "debug": "*:D", "info": "*:I",
	"warning": "*:W", "error": "*:E", "fatal": "*:F",
}

func BuildArgs(opts Options) []string {
	var args []string

	if opts.Clear {
		args = append(args, "-b", "all", "-c")
		return args
	}

	if opts.Dump {
		args = append(args, "-d")
	}

	if opts.Lines > 0 {
		args = append(args, "-t", strconv.Itoa(opts.Lines))
	}

	if opts.Package != "" {
		args = append(args, "--pid="+opts.Package)
	}

	if opts.Tag != "" {
		args = append(args, "-s", opts.Tag+":*")
	}

	if opts.Level != "" {
		if l, ok := levels[strings.ToLower(opts.Level)]; ok {
			args = append(args, l)
		}
	}

	format := opts.Format
	if format == "" {
		format = "brief"
	}
	args = append(args, "-v", format)

	return args
}

func ResolvePID(cfg *adb.Config, r adb.Runner, pkg string) (int, error) {
	out, err := r.Shell(cfg, "ps -A | grep "+pkg)
	if err != nil {
		return 0, fmt.Errorf("failed to list processes: %w", err)
	}
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			pid, err := strconv.Atoi(fields[1])
			if err == nil {
				return pid, nil
			}
		}
	}
	return 0, fmt.Errorf("process not found for package: %s", pkg)
}

func RunStream(cfg *adb.Config, opts Options) error {
	adbPath, err := exec.LookPath("adb")
	if err != nil {
		return fmt.Errorf("ADB not found on PATH")
	}
	logArgs := BuildArgs(opts)
	cmd := exec.Command(adbPath, append([]string{"-s", cfg.TVIP, "logcat"}, logArgs...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func RunOnce(cfg *adb.Config, r adb.Runner, opts Options) (string, error) {
	logArgs := BuildArgs(opts)
	cmd := "logcat " + strings.Join(logArgs, " ")
	return r.ShellWithStderr(cfg, cmd)
}
```

Key changes:
- `BuildArgs`: added `if opts.Package != "" { args = append(args, "--pid="+opts.Package) }`
- `RunStream`: removed `adbPath()` function, use `exec.LookPath` inline. Set `cmd.Stdout = os.Stdout`, `cmd.Stderr = os.Stderr`
- Removed duplicate `adbPath()` function

- [ ] **Step 2: Write test for Package filter in BuildArgs**

```go
func TestBuildArgs_WithPackage(t *testing.T) {
	opts := Options{Package: "1234"}
	args := BuildArgs(opts)
	hasPID := false
	for _, a := range args {
		if a == "--pid=1234" {
			hasPID = true
		}
	}
	if !hasPID {
		t.Errorf("BuildArgs() = %v, missing --pid=1234", args)
	}
}
```

Add to `internal/logcat/logcat_test.go`.

- [ ] **Step 3: Run tests**

Run: `cd C:\Users\ajink\Projects\tv && go test ./internal/logcat/ -v`
Expected: all 9 tests PASS (8 existing + 1 new).

#### Sub-task 5b: Fix cmd/tv/main.go to pass resolved PID to logcat

- [ ] **Step 4: Fix logsCmd to pass resolved PID into opts.Package**

Current code resolves PID but stores it in a local `pid` variable (line 348), never putting it into `opts.Package`.

Replace the `-p` handling section in `logsCmd` (lines 338-354):

```go
	if *packageFlag != "" {
		pkg := *packageFlag
		if !strings.Contains(pkg, ".") {
			var err error
			pkg, err = player.Resolve(cfg, r, pkg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		}
		pid, err := logcat.ResolvePID(cfg, r, pkg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		opts.Package = strconv.Itoa(pid)
	}
```

Add `"strconv"` to imports in `cmd/tv/main.go`.

- [ ] **Step 5: Verify build**

Run: `cd C:\Users\ajink\Projects\tv && go build ./...`
Expected: no errors.

- [ ] **Step 6: Run all tests**

Run: `cd C:\Users\ajink\Projects\tv && go test ./... -v -count=1`
Expected: all 31+ tests PASS.

- [ ] **Step 7: Commit**

```bash
cd C:\Users\ajink\Projects\tv
git add internal/logcat/ cmd/tv/main.go
git commit -m "fix: logcat -p PID filter, RunStream stdout, remove duplicate adbPath()"
```

---

### Task 6: Fix Makefile for Windows

**Files:**
- Modify: `Makefile:17`

- [ ] **Step 1: Replace rm -f with Windows-compatible clean**

```makefile
clean:
	del /f /q $(BINARY).exe $(BINARY)-windows-*.exe $(BINARY)-linux-* $(BINARY)-darwin-* 2>nul || true
```

- [ ] **Step 2: Verify build works**

Run: `cd C:\Users\ajink\Projects\tv && go build -o tv.exe ./cmd/tv/`
Expected: binary created, no errors.

- [ ] **Step 3: Commit**

```bash
cd C:\Users\ajink\Projects\tv
git add Makefile
git commit -m "fix: Makefile clean target for Windows (del instead of rm)"
```

---

### Task 7: Add tv logs dump positional subcommand support

**Files:**
- Modify: `cmd/tv/main.go`

- [ ] **Step 1: Add positional "dump" support to logsCmd**

The spec says `tv logs dump` should work. Currently only `-d` flag works. Add check at start of logsCmd:

After `fs.Parse(args)` and after setting opts, add:

```go
	// Support positional "dump" subcommand
	if len(fs.Args()) > 0 && fs.Args()[0] == "dump" {
		opts.Dump = true
	}
```

Insert this after `opts.Dump = *dumpFlag` and before the first `EnsureConnected` call.

- [ ] **Step 2: Verify build**

Run: `cd C:\Users\ajink\Projects\tv && go build ./...`
Expected: no errors.

- [ ] **Step 3: Commit**

```bash
cd C:\Users\ajink\Projects\tv
git add cmd/tv/main.go
git commit -m "feat: support 'tv logs dump' positional subcommand"
```

---

### Task 8: Update .gitignore for worktree/build artifacts

**Files:**
- Modify: `.gitignore`

- [ ] **Step 1: Add build output and coverage patterns**

```gitignore
*.exe
*.test
*.out
*.log
tv.exe
tv-windows-*.exe
tv-linux-*
tv-darwin-*
.worktrees/
```

- [ ] **Step 2: Commit**

```bash
cd C:\Users\ajink\Projects\tv
git add .gitignore
git commit -m "chore: update gitignore with build artifacts"
```

---

### Task 9: Final verification

- [ ] **Step 1: Binary smoke test**

```bash
cd C:\Users\ajink\Projects\tv
go build -o tv.exe ./cmd/tv/
.\tv.exe
```

Expected: prints usage. No crash.

- [ ] **Step 2: Run full test suite**

```bash
cd C:\Users\ajink\Projects\tv
go test -v -count=1 ./...
```

Expected: all tests PASS.

- [ ] **Step 3: Verify git log is clean**

```bash
cd C:\Users\ajink\Projects\tv
git status
git log --oneline -10
```

Expected: working tree clean, history shows all 11+ commits.

---

## Edge Cases Documented (not code changes)

These were identified during audit but don't need code fixes right now:

1. **app launch uses deprecated `monkey`** — Android has deprecated `monkey`. Should use `am start -n <package>/.<activity>` instead. Mitigation: `monkey` still works on Android 14 and below.

2. **PID resolution via `ps -A | grep` is fragile** — grep pattern could match partial package names. Mitigation: documented in spec, acceptable for internal tool.

3. **logcat streaming bypasses Runner interface** — `RunStream` uses `os/exec` directly. Mitigation: `RunOnce` uses `Runner` for testable dump/clear paths; streaming is inherently hard to mock.

4. **knownPrefixes in Detect() may false-positive** — `strings.Contains(lower, "player")` matches anything with "player" in package name. Mitigation: knownPlayers map takes priority; prefix matches only catch unrecognized players.

5. **`remote.Type()` escaping duplicated from `player.EscapeURLForShell()`** — Both do `'\''` escaping. Extracting to shared utility would be cleaner but adds coupling; acceptable for a CLI tool of this size.
````

## File: docs/superpowers/plans/2026-07-27-tv-play.md
````markdown
# tv Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. REQUIRED SUB-SKILL: Use superpowers:using-git-worktrees for isolated workspace.

**GitHub account:** mutedfalcontv (gh auth switch --user mutedfalcontv before push)
**Default branch:** master (not main)

**Goal:** Monorepo for Android TV CLI tools — parent `tv` binary with subcommands (ADB control, app management, player detection, video playback). Shared packages in `pkg/` so future tools reuse ADB, config, and player logic.

**Architecture:**
```
tv/
├── cmd/tv/main.go         # Parent CLI — dispatches subcommands
├── pkg/
│   ├── adb/               # adbRunner interface, realADBRunner, mockADBRunner
│   ├── config/            # Config struct, LoadConfig, Save
│   └── player/            # Player detection, MIME, URL escaping, intent building
├── go.mod                 # Module: github.com/mutedfalcontv/tv
├── Makefile
└── .gitignore
```

**Subcommands under `tv`:**
```
tv adb connect                Connect ADB to TV
tv adb disconnect             Disconnect ADB from TV
tv app list                   List all installed apps
tv app kill <package>         Force-stop an app
tv app launch <package>       Launch an app by package name
tv config                     Show configuration
tv player list                List video players (* = default)
tv player default <name>      Set default player
tv play [options] <url>       Play URL on TV
```

**Shared packages:**
- `pkg/adb` — `Runner` interface + `RealRunner` + `MockRunner`, `EnsureConnected`, `GetTVIP`
- `pkg/config` — `Config` struct, `Load()`, `Save()`, default `configDir()`
- `pkg/player` — `Detect()`, `Resolve()`, `MimeForURL()`, `EscapeURLForShell()`, `PlayOnTV()`

**Tech Stack:** Go 1.26 (stdlib only: flag, os/exec, encoding/json, testing, path/filepath, strings)

**Key design decisions:**
- `Runner` interface in `pkg/adb` — all higher code takes `Runner` param, mockable in tests
- `flag.FlagSet` per subcommand in parent CLI
- `TV_IP` env var overrides config
- URL escaped for ADB shell with single-quote wrapping
- `pkg/player.PlayOnTV` captures intent on mock via `LastIntent` field

---
### Task 1: Module init + directory structure

**Files:**
- Create: `tv/go.mod`

- [ ] **Step 1: Create directories**

```bash
mkdir -p C:\Users\ajink\Projects\tv\cmd\tv
mkdir -p C:\Users\ajink\Projects\tv\pkg\adb
mkdir -p C:\Users\ajink\Projects\tv\pkg\config
mkdir -p C:\Users\ajink\Projects\tv\pkg\player
mkdir -p C:\Users\ajink\Projects\tv\docs\superpowers\plans
```

- [ ] **Step 2: Remove old files (if redoing)**

```bash
Remove-Item -Recurse -Force C:\Users\ajink\Projects\tv-play  # only if exists
```

- [ ] **Step 3: Init Go module**

```bash
cd C:\Users\ajink\Projects\tv
go mod init github.com/mutedfalcontv/tv
```

- [ ] **Step 4: Verify empty module**

Run: `go build ./...`
Expected: no errors (nothing to build yet).

- [ ] **Step 5: Commit**

```bash
git init
git add go.mod
git commit -m "chore: init monorepo module github.com/mutedfalcontv/tv"
```

---
### Task 2: pkg/config — shared config package

**Files:**
- Create: `tv/pkg/config/config.go`
- Create: `tv/pkg/config/config_test.go`

- [ ] **Step 1: Write failing test (config_test.go)**

```go
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_ReturnsDefaultsWhenMissing(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if cfg.TVIP != "192.168.2.3:5555" {
		t.Errorf("default TVIP = %q, want 192.168.2.3:5555", cfg.TVIP)
	}
	if cfg.DefaultPlayer != "" {
		t.Errorf("default DefaultPlayer = %q, want empty", cfg.DefaultPlayer)
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	orig := &Config{TVIP: "10.0.0.1:5555", DefaultPlayer: "com.brouken.player"}
	if err := orig.Save(); err != nil {
		t.Fatalf("Save() unexpected error: %v", err)
	}
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if loaded.TVIP != orig.TVIP {
		t.Errorf("TVIP = %q, want %q", loaded.TVIP, orig.TVIP)
	}
	if loaded.DefaultPlayer != orig.DefaultPlayer {
		t.Errorf("DefaultPlayer = %q, want %q", loaded.DefaultPlayer, orig.DefaultPlayer)
	}
}

func TestDir(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	dir, err := Dir()
	if err != nil {
		t.Fatalf("Dir() unexpected error: %v", err)
	}
	if filepath.Base(dir) != "tv" {
		t.Errorf("Dir() base = %q, want tv", filepath.Base(dir))
	}
}
```

- [ ] **Step 2: Run test, expect failure**

```bash
cd C:\Users\ajink\Projects\tv
go test ./pkg/config/ -v -run TestLoad
```

Expected: FAIL (package not found).

- [ ] **Step 3: Write config.go**

```go
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	TVIP          string `json:"tv_ip"`
	DefaultPlayer string `json:"default_player"`
}

func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "tv"), nil
}

func path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func Load() (*Config, error) {
	p, err := path()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{TVIP: "192.168.2.3:5555"}, nil
		}
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.TVIP == "" {
		cfg.TVIP = "192.168.2.3:5555"
	}
	return &cfg, nil
}

func (c *Config) Save() error {
	dir, err := Dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	p, err := path()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0644)
}
```

- [ ] **Step 4: Run tests, expect pass**

```bash
cd C:\Users\ajink\Projects\tv
go test ./pkg/config/ -v
```

Expected: all 3 tests PASS.

- [ ] **Step 5: Commit**

```bash
git add pkg/config/
git commit -m "feat: pkg/config with Load/Save and tests"
```

---
### Task 3: pkg/adb — shared ADB interface + mock

**Files:**
- Create: `tv/pkg/adb/runner.go`
- Create: `tv/pkg/adb/runner_test.go`

- [ ] **Step 1: Write failing test (runner_test.go)**

```go
package adb

import (
	"testing"
)

func TestEnsureConnected_AlreadyConnected(t *testing.T) {
	m := &MockRunner{
		DevicesOut: "List of devices attached\n192.168.2.3:5555   device\n",
	}
	cfg := &Config{TVIP: "192.168.2.3:5555"}
	err := EnsureConnected(cfg, m)
	if err != nil {
		t.Fatalf("EnsureConnected() unexpected error: %v", err)
	}
	if m.ConnectCalled {
		t.Error("Connect should not be called when device is already connected")
	}
}

func TestEnsureConnected_NotConnected(t *testing.T) {
	m := &MockRunner{
		DevicesOut: "List of devices attached\n",
		ConnectOut: "connected to 192.168.2.3:5555",
	}
	cfg := &Config{TVIP: "192.168.2.3:5555"}
	err := EnsureConnected(cfg, m)
	if err != nil {
		t.Fatalf("EnsureConnected() unexpected error: %v", err)
	}
	if !m.ConnectCalled {
		t.Error("Connect should be called when device is not connected")
	}
}

func TestEnsureConnected_Offline(t *testing.T) {
	m := &MockRunner{
		DevicesOut: "List of devices attached\n192.168.2.3:5555   offline\n",
		ConnectOut: "connected to 192.168.2.3:5555",
	}
	cfg := &Config{TVIP: "192.168.2.3:5555"}
	err := EnsureConnected(cfg, m)
	if err != nil {
		t.Fatalf("EnsureConnected() unexpected error: %v", err)
	}
	if !m.ConnectCalled {
		t.Error("Connect should be called when device is offline")
	}
}

func TestEnsureConnected_ConnectFails(t *testing.T) {
	m := &MockRunner{
		DevicesOut: "List of devices attached\n",
		ConnectOut: "failed",
		ConnectErr: assertAnError{},
	}
	cfg := &Config{TVIP: "192.168.2.3:5555"}
	err := EnsureConnected(cfg, m)
	if err == nil {
		t.Fatal("EnsureConnected() expected error, got nil")
	}
}
```

- [ ] **Step 2: Run test, expect failure**

```bash
cd C:\Users\ajink\Projects\tv
go test ./pkg/adb/ -v
```

Expected: FAIL.

- [ ] **Step 3: Write runner.go**

```go
package adb

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Config struct {
	TVIP          string
	DefaultPlayer string
}

type Runner interface {
	Devices() (string, error)
	Connect(ip string) (string, error)
	Disconnect(ip string) (string, error)
	Shell(cfg *Config, cmd string) (string, error)
	ShellWithStderr(cfg *Config, cmd string) (string, error)
}

type RealRunner struct{}

func (r *RealRunner) binary() (string, error) {
	path, err := exec.LookPath("adb")
	if err != nil {
		return "", fmt.Errorf("ADB not found on PATH — install Android platform-tools")
	}
	return path, nil
}

func (r *RealRunner) Devices() (string, error) {
	adb, err := r.binary()
	if err != nil {
		return "", err
	}
	out, err := exec.Command(adb, "devices").Output()
	if err != nil {
		return "", fmt.Errorf("adb devices: %w", err)
	}
	return string(out), nil
}

func (r *RealRunner) Connect(ip string) (string, error) {
	adb, err := r.binary()
	if err != nil {
		return "", err
	}
	out, err := exec.Command(adb, "connect", ip).CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func (r *RealRunner) Disconnect(ip string) (string, error) {
	adb, err := r.binary()
	if err != nil {
		return "", err
	}
	out, err := exec.Command(adb, "disconnect", ip).CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func (r *RealRunner) Shell(cfg *Config, cmd string) (string, error) {
	adb, err := r.binary()
	if err != nil {
		return "", err
	}
	out, err := exec.Command(adb, "-s", cfg.TVIP, "shell", cmd).Output()
	if err != nil {
		return "", fmt.Errorf("ADB error: %w\n%s", err, string(out))
	}
	return strings.TrimSpace(string(out)), nil
}

func (r *RealRunner) ShellWithStderr(cfg *Config, cmd string) (string, error) {
	adb, err := r.binary()
	if err != nil {
		return "", err
	}
	out, err := exec.Command(adb, "-s", cfg.TVIP, "shell", cmd).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("ADB error: %w\n%s", err, string(out))
	}
	return strings.TrimSpace(string(out)), nil
}

func GetTVIP() string {
	if v := os.Getenv("TV_IP"); v != "" {
		return v
	}
	cfg, err := loadConfig()
	if err != nil {
		return "192.168.2.3:5555"
	}
	return cfg.TVIP
}

func loadConfig() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(filepath.Join(home, ".config", "tv", "config.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{TVIP: "192.168.2.3:5555"}, nil
		}
		return nil, err
	}
	var cfg struct {
		TVIP          string `json:"tv_ip"`
		DefaultPlayer string `json:"default_player"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.TVIP == "" {
		return &Config{TVIP: "192.168.2.3:5555"}, nil
	}
	return &Config{TVIP: cfg.TVIP, DefaultPlayer: cfg.DefaultPlayer}, nil
}

func EnsureConnected(cfg *Config, r Runner) error {
	out, err := r.Devices()
	if err != nil {
		return err
	}
	if strings.Contains(out, cfg.TVIP) && !strings.Contains(out, "offline") {
		return nil
	}
	_, err = r.Connect(cfg.TVIP)
	return err
}

type assertAnError struct{}
func (assertAnError) Error() string { return "expected error" }

type MockRunner struct {
	DevicesOut    string
	DevicesErr    error
	ConnectOut    string
	ConnectErr    error
	DisconnectOut string
	DisconnectErr error
	ShellOut      string
	ShellErr      error
	ShellWithStderrOut string
	ShellWithStderrErr error
	ConnectCalled    bool
	DisconnectCalled bool
	LastIntent       string
}

func (m *MockRunner) Devices() (string, error)       { return m.DevicesOut, m.DevicesErr }
func (m *MockRunner) Connect(ip string) (string, error)    { m.ConnectCalled = true; return m.ConnectOut, m.ConnectErr }
func (m *MockRunner) Disconnect(ip string) (string, error) { m.DisconnectCalled = true; return m.DisconnectOut, m.DisconnectErr }
func (m *MockRunner) Shell(cfg *Config, cmd string) (string, error)       { return m.ShellOut, m.ShellErr }
func (m *MockRunner) ShellWithStderr(cfg *Config, cmd string) (string, error) { return m.ShellWithStderrOut, m.ShellWithStderrErr }
```

- [ ] **Step 4: Run tests, expect pass**

```bash
cd C:\Users\ajink\Projects\tv
go test ./pkg/adb/ -v
```

Expected: all 4 tests PASS.

- [ ] **Step 5: Commit**

```bash
git add pkg/adb/
git commit -m "feat: pkg/adb with Runner interface, RealRunner, MockRunner, EnsureConnected, tests"
```

---
### Task 4: pkg/player — detection + playback logic

**Files:**
- Create: `tv/pkg/player/detect.go`
- Create: `tv/pkg/player/detect_test.go`
- Create: `tv/pkg/player/play.go`
- Create: `tv/pkg/player/play_test.go`

- [ ] **Step 1: Write failing detect test (detect_test.go)**

```go
package player

import (
	"testing"
	"tv/pkg/adb"
)

func TestDetect_Known(t *testing.T) {
	mock := &adb.MockRunner{
		ShellOut: `package:com.brouken.player
package:org.videolan.vlc
package:com.android.chrome
package:com.google.android.youtube`,
	}
	cfg := &adb.Config{TVIP: "192.168.2.3:5555"}
	players, err := Detect(cfg, mock)
	if err != nil {
		t.Fatalf("Detect() unexpected error: %v", err)
	}
	if len(players) != 3 {
		t.Fatalf("got %d players, want 3", len(players))
	}
	got := map[string]bool{}
	for _, p := range players {
		got[p.Name] = true
	}
	for _, name := range []string{"Just Player", "VLC", "YouTube"} {
		if !got[name] {
			t.Errorf("missing player: %s", name)
		}
	}
}

func TestDetect_PrefixMatch(t *testing.T) {
	mock := &adb.MockRunner{
		ShellOut: `package:com.some.videoplayer
package:org.example.kodi
package:com.android.chrome`,
	}
	cfg := &adb.Config{TVIP: "192.168.2.3:5555"}
	players, err := Detect(cfg, mock)
	if err != nil {
		t.Fatalf("Detect() unexpected error: %v", err)
	}
	if len(players) != 2 {
		t.Fatalf("got %d players, want 2", len(players))
	}
}

func TestDetect_Sorted(t *testing.T) {
	mock := &adb.MockRunner{
		ShellOut: `package:org.videolan.vlc
package:com.brouken.player`,
	}
	cfg := &adb.Config{TVIP: "192.168.2.3:5555"}
	players, err := Detect(cfg, mock)
	if err != nil {
		t.Fatalf("Detect() unexpected error: %v", err)
	}
	if len(players) < 2 {
		t.Fatal("need at least 2 players for sort test")
	}
	if players[0].Name > players[1].Name {
		t.Errorf("players not sorted: %s > %s", players[0].Name, players[1].Name)
	}
}

func TestResolve_ByName(t *testing.T) {
	mock := &adb.MockRunner{ShellOut: `package:com.brouken.player`}
	cfg := &adb.Config{TVIP: "192.168.2.3:5555"}
	pkg, err := Resolve(cfg, mock, "Just Player")
	if err != nil {
		t.Fatalf("Resolve() unexpected error: %v", err)
	}
	if pkg != "com.brouken.player" {
		t.Errorf("got %q, want com.brouken.player", pkg)
	}
}

func TestResolve_ByPackage(t *testing.T) {
	mock := &adb.MockRunner{ShellOut: `package:org.videolan.vlc`}
	cfg := &adb.Config{TVIP: "192.168.2.3:5555"}
	pkg, err := Resolve(cfg, mock, "org.videolan.vlc")
	if err != nil {
		t.Fatalf("Resolve() unexpected error: %v", err)
	}
	if pkg != "org.videolan.vlc" {
		t.Errorf("got %q, want org.videolan.vlc", pkg)
	}
}

func TestResolve_NotFound(t *testing.T) {
	mock := &adb.MockRunner{ShellOut: `package:com.brouken.player`}
	cfg := &adb.Config{TVIP: "192.168.2.3:5555"}
	_, err := Resolve(cfg, mock, "NonExistent")
	if err == nil {
		t.Fatal("Resolve() expected error, got nil")
	}
}
```

- [ ] **Step 2: Write failing play test (play_test.go)**

```go
package player

import (
	"strings"
	"testing"
	"tv/pkg/adb"
)

func TestMimeForURL_M3U8(t *testing.T) {
	got := MimeForURL("https://example.com/stream.m3u8")
	want := "application/x-mpegURL"
	if got != want {
		t.Errorf("MimeForURL() = %q, want %q", got, want)
	}
}

func TestMimeForURL_MP4(t *testing.T) {
	got := MimeForURL("https://example.com/video.mp4?token=abc")
	want := "video/mp4"
	if got != want {
		t.Errorf("MimeForURL() = %q, want %q", got, want)
	}
}

func TestMimeForURL_CaseInsensitive(t *testing.T) {
	got := MimeForURL("https://example.com/video.MKV")
	want := "video/x-matroska"
	if got != want {
		t.Errorf("MimeForURL() = %q, want %q", got, want)
	}
}

func TestMimeForURL_Fallback(t *testing.T) {
	got := MimeForURL("https://example.com/file.unknown")
	want := "video/mp4"
	if got != want {
		t.Errorf("MimeForURL() = %q, want %q", got, want)
	}
}

func TestEscapeURLForShell_Simple(t *testing.T) {
	got := EscapeURLForShell("https://example.com/video.mp4")
	want := "'https://example.com/video.mp4'"
	if got != want {
		t.Errorf("EscapeURLForShell() = %q, want %q", got, want)
	}
}

func TestEscapeURLForShell_WithAmpersand(t *testing.T) {
	url := "https://example.com/v?token=a&b=c"
	got := EscapeURLForShell(url)
	if !strings.HasPrefix(got, "'") || !strings.HasSuffix(got, "'") {
		t.Errorf("EscapeURLForShell() = %q, should be single-quote wrapped", got)
	}
}

func TestEscapeURLForShell_WithSingleQuote(t *testing.T) {
	got := EscapeURLForShell("https://example.com/it's.mp4")
	want := "'https://example.com/it'\\''s.mp4'"
	if got != want {
		t.Errorf("EscapeURLForShell() = %q, want %q", got, want)
	}
}

func TestActivityForPlayer_Known(t *testing.T) {
	got := ActivityForPlayer("com.brouken.player")
	want := ".PlayerActivity"
	if got != want {
		t.Errorf("ActivityForPlayer() = %q, want %q", got, want)
	}
}

func TestActivityForPlayer_Unknown(t *testing.T) {
	got := ActivityForPlayer("com.example.player")
	want := ".PlayerActivity"
	if got != want {
		t.Errorf("ActivityForPlayer() = %q, want %q", got, want)
	}
}

func TestPlayOnTV_BuildsIntent(t *testing.T) {
	mock := &adb.MockRunner{}
	cfg := &adb.Config{TVIP: "192.168.2.3:5555"}
	err := PlayOnTV(cfg, mock, "org.videolan.vlc", "https://example.com/v.m3u8")
	if err != nil {
		t.Fatalf("PlayOnTV() unexpected error: %v", err)
	}
	expected := "am start -a android.intent.action.VIEW -d 'https://example.com/v.m3u8' -t application/x-mpegURL -n org.videolan.vlc/.gui.video.VideoPlayerActivity"
	if mock.LastIntent != expected {
		t.Errorf("intent = %q\nwant = %q", mock.LastIntent, expected)
	}
}

func TestPlayOnTV_UnknownPlayerDefaults(t *testing.T) {
	mock := &adb.MockRunner{}
	cfg := &adb.Config{TVIP: "192.168.2.3:5555"}
	err := PlayOnTV(cfg, mock, "com.unknown.player", "https://example.com/v.mp4")
	if err != nil {
		t.Fatalf("PlayOnTV() unexpected error: %v", err)
	}
	if !strings.Contains(mock.LastIntent, "com.unknown.player/.PlayerActivity") {
		t.Errorf("intent should use default .PlayerActivity for unknown players")
	}
}
```

- [ ] **Step 3: Run tests, expect failure**

```bash
cd C:\Users\ajink\Projects\tv
go test ./pkg/player/ -v
```

Expected: FAIL — functions not defined.

- [ ] **Step 4: Write detect.go**

```go
package player

import (
	"fmt"
	"sort"
	"strings"
	"tv/pkg/adb"
)

type Info struct {
	PackageName string
	Name        string
}

var knownPlayers = map[string]string{
	"com.brouken.player":        "Just Player",
	"org.videolan.vlc":          "VLC",
	"app.mpvnova.player":        "MPV Nova",
	"com.mxtech.videoplayer.ad": "MX Player",
	"com.google.android.youtube": "YouTube",
}

var knownPrefixes = []string{
	"player", "vlc", "mpv", "mxplayer", "videoplayer", "kodi", "nova",
}

func Detect(cfg *adb.Config, r adb.Runner) ([]Info, error) {
	if err := adb.EnsureConnected(cfg, r); err != nil {
		return nil, err
	}
	out, err := r.Shell(cfg, "pm list packages")
	if err != nil {
		return nil, fmt.Errorf("failed to list packages: %w", err)
	}
	lines := strings.Split(out, "\n")
	installed := make(map[string]bool)
	for _, line := range lines {
		pkg := strings.TrimPrefix(line, "package:")
		pkg = strings.TrimSpace(pkg)
		if pkg != "" {
			installed[pkg] = true
		}
	}
	var players []Info
	for pkg, name := range knownPlayers {
		if installed[pkg] {
			players = append(players, Info{PackageName: pkg, Name: name})
		}
	}
	for pkg := range installed {
		if _, known := knownPlayers[pkg]; known {
			continue
		}
		lower := strings.ToLower(pkg)
		for _, prefix := range knownPrefixes {
			if strings.Contains(lower, prefix) {
				players = append(players, Info{PackageName: pkg, Name: pkg})
				break
			}
		}
	}
	sort.Slice(players, func(i, j int) bool {
		return players[i].Name < players[j].Name
	})
	return players, nil
}

func Resolve(cfg *adb.Config, r adb.Runner, name string) (string, error) {
	players, err := Detect(cfg, r)
	if err != nil {
		return "", err
	}
	for _, p := range players {
		if p.Name == name || p.PackageName == name {
			return p.PackageName, nil
		}
	}
	return "", fmt.Errorf("player '%s' not found. Run 'tv player list'", name)
}
```

- [ ] **Step 5: Write play.go**

```go
package player

import (
	"fmt"
	"strings"
	"tv/pkg/adb"
)

var mimeByExt = map[string]string{
	"m3u8": "application/x-mpegURL",
	"mp4":  "video/mp4",
	"mkv":  "video/x-matroska",
	"avi":  "video/avi",
	"mov":  "video/quicktime",
	"webm": "video/webm",
}

var playerActivity = map[string]string{
	"com.brouken.player":        ".PlayerActivity",
	"org.videolan.vlc":          ".gui.video.VideoPlayerActivity",
	"app.mpvnova.player":        ".PlayerActivity",
	"com.mxtech.videoplayer.ad": ".ActivityMedia",
}

func MimeForURL(url string) string {
	lower := strings.ToLower(url)
	for ext, mime := range mimeByExt {
		if strings.HasSuffix(lower, "."+ext) {
			return mime
		}
		if strings.Contains(lower, "."+ext+"?") || strings.Contains(lower, "."+ext+"&") {
			return mime
		}
	}
	return "video/mp4"
}

func EscapeURLForShell(url string) string {
	return "'" + strings.ReplaceAll(url, "'", "'\\''") + "'"
}

func ActivityForPlayer(pkg string) string {
	if act, ok := playerActivity[pkg]; ok {
		return act
	}
	return ".PlayerActivity"
}

func PlayOnTV(cfg *adb.Config, r adb.Runner, playerPkg, url string) error {
	mime := MimeForURL(url)
	activity := ActivityForPlayer(playerPkg)
	quotedURL := EscapeURLForShell(url)
	intent := fmt.Sprintf("am start -a android.intent.action.VIEW -d %s -t %s -n %s/%s",
		quotedURL, mime, playerPkg, activity)
	if m, ok := r.(*adb.MockRunner); ok {
		m.LastIntent = intent
		return nil
	}
	_, err := r.ShellWithStderr(cfg, intent)
	if err != nil {
		return fmt.Errorf("failed to launch player: %w", err)
	}
	return nil
}
```

- [ ] **Step 6: Run tests, expect pass**

```bash
cd C:\Users\ajink\Projects\tv
go test ./pkg/player/ -v
```

Expected: all 12 tests PASS.

- [ ] **Step 7: Run all tests**

```bash
cd C:\Users\ajink\Projects\tv
go test ./... -v
```

Expected: all 19 tests PASS (3 config + 4 adb + 12 player).

- [ ] **Step 8: Commit**

```bash
git add pkg/player/
git commit -m "feat: pkg/player with detection, playback, tests"
```

---
### Task 5: cmd/tv — parent CLI entry point

**Files:**
- Create: `tv/cmd/tv/main.go`

- [ ] **Step 1: Write main.go**

```go
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"tv/pkg/adb"
	"tv/pkg/config"
	"tv/pkg/player"
)

type command struct {
	name string
	desc string
	run  func(args []string)
}

var commands = []command{
	{"adb", "ADB connection management", adbCmd},
	{"app", "App management (list, launch, kill)", appCmd},
	{"config", "Show configuration", configCmd},
	{"player", "List/set default video player", playerCmd},
	{"play", "Play a URL on TV", playCmd},
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
	sub := os.Args[1]
	for _, c := range commands {
		if c.name == sub {
			c.run(os.Args[2:])
			return
		}
	}
	fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", sub)
	printUsage()
	os.Exit(1)
}

func printUsage() {
	fmt.Println(`tv — Android TV CLI tools

Usage:
  tv <command> [options] [args]

Commands:`)
	for _, c := range commands {
		fmt.Printf("  %-15s %s\n", c.name, c.desc)
	}
}

func adbCmd(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: tv adb <connect|disconnect>")
		os.Exit(1)
	}
	r := &adb.RealRunner{}
	ip := adb.GetTVIP()
	switch args[0] {
	case "connect":
		fmt.Printf("Connecting to %s...\n", ip)
		result, err := r.Connect(ip)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(result)
	case "disconnect":
		fmt.Printf("Disconnecting from %s...\n", ip)
		result, err := r.Disconnect(ip)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(result)
	default:
		fmt.Fprintf(os.Stderr, "unknown adb subcommand: %s\n", args[0])
		os.Exit(1)
	}
}

func appCmd(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: tv app <list|launch|kill> [package]")
		os.Exit(1)
	}

	cfg := loadAdbConfig()
	r := &adb.RealRunner{}

	switch args[0] {
	case "list":
		if err := adb.EnsureConnected(cfg, r); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		out, err := r.Shell(cfg, "pm list packages")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(out)
	case "launch":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: tv app launch <package>")
			os.Exit(1)
		}
		if err := adb.EnsureConnected(cfg, r); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		_, err := r.ShellWithStderr(cfg, "monkey -p "+args[1]+" -c android.intent.category.LAUNCHER 1")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error launching %s: %v\n", args[1], err)
			os.Exit(1)
		}
		fmt.Printf("Launched %s\n", args[1])
	case "kill":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: tv app kill <package>")
			os.Exit(1)
		}
		if err := adb.EnsureConnected(cfg, r); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		_, err := r.ShellWithStderr(cfg, "am force-stop "+args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error killing %s: %v\n", args[1], err)
			os.Exit(1)
		}
		fmt.Printf("Killed %s\n", args[1])
	default:
		fmt.Fprintf(os.Stderr, "unknown app subcommand: %s\n", args[0])
		os.Exit(1)
	}
}

func configCmd(args []string) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Config file: ~/.config/tv/config.json\n")
	fmt.Printf("TV IP:        %s\n", cfg.TVIP)
	if cfg.DefaultPlayer != "" {
		fmt.Printf("Default player: %s\n", cfg.DefaultPlayer)
	} else {
		fmt.Println("Default player: (none)")
	}
}

func playerCmd(args []string) {
	cfg := loadAdbConfig()
	r := &adb.RealRunner{}

	fs := flag.NewFlagSet("player", flag.ExitOnError)
	ip := fs.String("ip", cfg.TVIP, "TV IP address")
	fs.Parse(args)
	cfg.TVIP = *ip

	remaining := fs.Args()
	if len(remaining) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: tv player <list|default> [options] [name]")
		os.Exit(1)
	}

	switch remaining[0] {
	case "list":
		players, err := player.Detect(cfg, r)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if len(players) == 0 {
			fmt.Println("No video players found on TV.")
			return
		}
		fmt.Println("Installed video players:")
		for _, p := range players {
			mark := " "
			if cfg.DefaultPlayer == p.PackageName {
				mark = "*"
			}
			fmt.Printf("  %s %-20s %s\n", mark, p.Name, p.PackageName)
		}
	case "default":
		if len(remaining) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: tv player default <name>")
			os.Exit(1)
		}
		name := remaining[1]
		pkg, err := player.Resolve(cfg, r, name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		conf, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		conf.DefaultPlayer = pkg
		conf.TVIP = cfg.TVIP
		if err := conf.Save(); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Default player set to: %s\n", pkg)
	default:
		fmt.Fprintf(os.Stderr, "unknown player subcommand: %s\n", remaining[0])
		os.Exit(1)
	}
}

func playCmd(args []string) {
	cfg := loadAdbConfig()
	r := &adb.RealRunner{}

	fs := flag.NewFlagSet("play", flag.ExitOnError)
	playerFlag := fs.String("player", cfg.DefaultPlayer, "Player name or package name")
	ip := fs.String("ip", cfg.TVIP, "TV IP address")
	fs.Parse(args)
	cfg.TVIP = *ip

	remaining := fs.Args()
	if len(remaining) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: tv play [options] <url>")
		fmt.Fprintln(os.Stderr, "  -player <name>   Player to use")
		fmt.Fprintln(os.Stderr, "  -ip <address>    TV IP address")
		os.Exit(1)
	}

	url := remaining[0]
	playerPkg := *playerFlag
	if playerPkg == "" {
		fmt.Fprintln(os.Stderr, "No default player set. Use 'tv player default <name>' or 'tv play -player <name> <url>'")
		os.Exit(1)
	}
	if !strings.Contains(playerPkg, ".") {
		pkg, err := player.Resolve(cfg, r, playerPkg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		playerPkg = pkg
	}
	if err := adb.EnsureConnected(cfg, r); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Playing on %s via %s...\n", cfg.TVIP, playerPkg)
	if err := player.PlayOnTV(cfg, r, playerPkg, url); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Done.")
}

func loadAdbConfig() *adb.Config {
	c, err := config.Load()
	if err != nil {
		return &adb.Config{TVIP: "192.168.2.3:5555"}
	}
	return &adb.Config{TVIP: c.TVIP, DefaultPlayer: c.DefaultPlayer}
}
```

- [ ] **Step 2: Build**

```bash
cd C:\Users\ajink\Projects\tv
go build -o tv.exe ./cmd/tv/
```

Expected: binary created. `.\tv.exe` prints usage.

- [ ] **Step 3: Update go.sum**

```bash
cd C:\Users\ajink\Projects\tv
go mod tidy
```

- [ ] **Step 4: Run full test suite**

```bash
cd C:\Users\ajink\Projects\tv
go test ./... -v
```

Expected: all 19 tests PASS.

- [ ] **Step 5: Commit**

```bash
git add cmd/tv/ go.sum
git commit -m "feat: cmd/tv parent CLI with adb, app, config, player, play commands"
```

---
### Task 6: Build artifacts

**Files:**
- Create: `tv/Makefile`
- Create: `tv/.gitignore`

- [ ] **Step 1: Write Makefile**

```makefile
BINARY := tv
GO := go

.PHONY: build test clean build-all

build:
	$(GO) build -o $(BINARY).exe ./cmd/$(BINARY)/

test:
	$(GO) test -v ./...

build-all:
	GOOS=windows GOARCH=amd64 $(GO) build -o $(BINARY)-windows-amd64.exe ./cmd/$(BINARY)/
	GOOS=linux GOARCH=amd64 $(GO) build -o $(BINARY)-linux-amd64 ./cmd/$(BINARY)/
	GOOS=darwin GOARCH=amd64 $(GO) build -o $(BINARY)-darwin-amd64 ./cmd/$(BINARY)/

clean:
	rm -f $(BINARY).exe $(BINARY)-windows-*.exe $(BINARY)-linux-* $(BINARY)-darwin-*
```

- [ ] **Step 2: Write .gitignore**

```
*.exe
*-windows-*.exe
*-linux-*
*-darwin-*
```

- [ ] **Step 3: Build + test**

```bash
cd C:\Users\ajink\Projects\tv
go build -o tv.exe ./cmd/tv/
go test ./... -v
```

Expected: build + all 19 tests pass.

- [ ] **Step 4: Commit**

```bash
git add Makefile .gitignore
git commit -m "chore: build artifacts and gitignore"
```

---
### Task 7: Integration + GitHub push

- [ ] **Step 1: Final test suite**

```bash
cd C:\Users\ajink\Projects\tv
go test -v -count=1 ./...
```

Expected: all 19 tests PASS.

- [ ] **Step 2: Binary smoke test**

```bash
.\tv.exe
.\tv.exe config
.\tv.exe player list
.\tv.exe player default "Just Player"
.\tv.exe play "https://example.com/video.mp4"
.\tv.exe play -player VLC "https://example.com/video.m3u8"
.\tv.exe adb connect
.\tv.exe adb disconnect
.\tv.exe app list
.\tv.exe app kill com.brouken.player
.\tv.exe app launch com.brouken.player
```

Expected: no crashes.

- [ ] **Step 3: Push to GitHub**

```bash
cd C:\Users\ajink\Projects\tv
gh auth switch --user mutedfalcontv
gh repo create mutedfalcontv/tv --public --push --source=. --remote=origin
git push -u origin master
```

- [ ] **Step 4: Mark complete**
````

## File: docs/superpowers/specs/2026-07-27-tv-logs-design.md
````markdown
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
````

## File: docs/superpowers/specs/2026-07-27-tv-remote-design.md
````markdown
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
````

## File: docs/WORKFLOW.md
````markdown
# Workflow Conventions

## Branch
- Default branch: `master`

## Git Worktrees
All feature work uses git worktrees for isolation. See `docs/superpowers/plans/` per feature.

Process:
1. `master` branch is ground truth — always up-to-date with origin
2. Each feature gets a worktree: `../tv-<feature>/`
3. Worktrees are temporary — deleted after merge to master
4. Never commit directly to master (except trivial fixes)

## GitHub
- Account: mutedfalcontv
- Repo: mutedfalcontv/tv
````

## File: internal/config/config_test.go
````go
package config

import (
	"path/filepath"
	"testing"
)

func TestLoad_ReturnsDefaultsWhenMissing(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if cfg.TVIP != "192.168.2.3:5555" {
		t.Errorf("default TVIP = %q, want 192.168.2.3:5555", cfg.TVIP)
	}
	if cfg.DefaultPlayer != "" {
		t.Errorf("default DefaultPlayer = %q, want empty", cfg.DefaultPlayer)
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	orig := &Config{TVIP: "10.0.0.1:5555", DefaultPlayer: "com.brouken.player"}
	if err := orig.Save(); err != nil {
		t.Fatalf("Save() unexpected error: %v", err)
	}
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if loaded.TVIP != orig.TVIP {
		t.Errorf("TVIP = %q, want %q", loaded.TVIP, orig.TVIP)
	}
	if loaded.DefaultPlayer != orig.DefaultPlayer {
		t.Errorf("DefaultPlayer = %q, want %q", loaded.DefaultPlayer, orig.DefaultPlayer)
	}
}

func TestDir(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	dir, err := Dir()
	if err != nil {
		t.Fatalf("Dir() unexpected error: %v", err)
	}
	if filepath.Base(dir) != "tv" {
		t.Errorf("Dir() base = %q, want tv", filepath.Base(dir))
	}
}
````

## File: internal/config/config.go
````go
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const defaultTVIP = "192.168.2.3:5555"

type Config struct {
	TVIP          string `json:"tv_ip"`
	DefaultPlayer string `json:"default_player"`
}

func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "tv"), nil
}

func path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func Load() (*Config, error) {
	p, err := path()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{TVIP: defaultTVIP}, nil
		}
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.TVIP == "" {
		cfg.TVIP = defaultTVIP
	}
	return &cfg, nil
}

func (c *Config) Save() error {
	dir, err := Dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	p, err := path()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0644)
}
````

## File: internal/player/detect_test.go
````go
package player

import (
	"testing"

	"github.com/mutedfalcontv/tv/internal/adb"
)

func TestDetect_Known(t *testing.T) {
	mock := &adb.MockRunner{
		ShellOut: `package:com.brouken.player
package:org.videolan.vlc
package:com.android.chrome
package:com.google.android.youtube`,
	}
	cfg := &adb.Config{TVIP: "192.168.2.3:5555"}
	players, err := Detect(cfg, mock)
	if err != nil {
		t.Fatalf("Detect() unexpected error: %v", err)
	}
	if len(players) != 3 {
		t.Fatalf("got %d players, want 3", len(players))
	}
	got := map[string]bool{}
	for _, p := range players {
		got[p.Name] = true
	}
	for _, name := range []string{"Just Player", "VLC", "YouTube"} {
		if !got[name] {
			t.Errorf("missing player: %s", name)
		}
	}
}

func TestDetect_PrefixMatch(t *testing.T) {
	mock := &adb.MockRunner{
		ShellOut: `package:com.some.videoplayer
package:org.example.kodi
package:com.android.chrome`,
	}
	cfg := &adb.Config{TVIP: "192.168.2.3:5555"}
	players, err := Detect(cfg, mock)
	if err != nil {
		t.Fatalf("Detect() unexpected error: %v", err)
	}
	if len(players) != 2 {
		t.Fatalf("got %d players, want 2", len(players))
	}
}

func TestDetect_Sorted(t *testing.T) {
	mock := &adb.MockRunner{
		ShellOut: `package:org.videolan.vlc
package:com.brouken.player`,
	}
	cfg := &adb.Config{TVIP: "192.168.2.3:5555"}
	players, err := Detect(cfg, mock)
	if err != nil {
		t.Fatalf("Detect() unexpected error: %v", err)
	}
	if len(players) < 2 {
		t.Fatal("need at least 2 players for sort test")
	}
	if players[0].Name > players[1].Name {
		t.Errorf("players not sorted: %s > %s", players[0].Name, players[1].Name)
	}
}

func TestResolve_ByName(t *testing.T) {
	mock := &adb.MockRunner{ShellOut: `package:com.brouken.player`}
	cfg := &adb.Config{TVIP: "192.168.2.3:5555"}
	pkg, err := Resolve(cfg, mock, "Just Player")
	if err != nil {
		t.Fatalf("Resolve() unexpected error: %v", err)
	}
	if pkg != "com.brouken.player" {
		t.Errorf("got %q, want com.brouken.player", pkg)
	}
}

func TestResolve_ByPackage(t *testing.T) {
	mock := &adb.MockRunner{ShellOut: `package:org.videolan.vlc`}
	cfg := &adb.Config{TVIP: "192.168.2.3:5555"}
	pkg, err := Resolve(cfg, mock, "org.videolan.vlc")
	if err != nil {
		t.Fatalf("Resolve() unexpected error: %v", err)
	}
	if pkg != "org.videolan.vlc" {
		t.Errorf("got %q, want org.videolan.vlc", pkg)
	}
}

func TestResolve_NotFound(t *testing.T) {
	mock := &adb.MockRunner{ShellOut: `package:com.brouken.player`}
	cfg := &adb.Config{TVIP: "192.168.2.3:5555"}
	_, err := Resolve(cfg, mock, "NonExistent")
	if err == nil {
		t.Fatal("Resolve() expected error, got nil")
	}
}
````

## File: internal/player/detect.go
````go
package player

import (
	"fmt"
	"sort"
	"strings"

	"github.com/mutedfalcontv/tv/internal/adb"
)

type Info struct {
	PackageName string
	Name        string
}

var knownPlayers = map[string]string{
	"com.brouken.player":         "Just Player",
	"org.videolan.vlc":           "VLC",
	"app.mpvnova.player":         "MPV Nova",
	"com.mxtech.videoplayer.ad":  "MX Player",
	"com.google.android.youtube": "YouTube",
}

var knownPrefixes = []string{
	"player", "vlc", "mpv", "mxplayer", "videoplayer", "kodi", "nova",
}

func Detect(cfg *adb.Config, r adb.Runner) ([]Info, error) {
	if err := adb.EnsureConnected(cfg, r); err != nil {
		return nil, err
	}
	out, err := r.Shell(cfg, "pm list packages")
	if err != nil {
		return nil, fmt.Errorf("failed to list packages: %w", err)
	}
	lines := strings.Split(out, "\n")
	installed := make(map[string]bool)
	for _, line := range lines {
		pkg := strings.TrimPrefix(line, "package:")
		pkg = strings.TrimSpace(pkg)
		if pkg != "" {
			installed[pkg] = true
		}
	}
	var players []Info
	for pkg, name := range knownPlayers {
		if installed[pkg] {
			players = append(players, Info{PackageName: pkg, Name: name})
		}
	}
	for pkg := range installed {
		if _, known := knownPlayers[pkg]; known {
			continue
		}
		lower := strings.ToLower(pkg)
		for _, prefix := range knownPrefixes {
			if strings.Contains(lower, prefix) {
				players = append(players, Info{PackageName: pkg, Name: pkg})
				break
			}
		}
	}
	sort.Slice(players, func(i, j int) bool {
		return players[i].Name < players[j].Name
	})
	return players, nil
}

func Resolve(cfg *adb.Config, r adb.Runner, name string) (string, error) {
	players, err := Detect(cfg, r)
	if err != nil {
		return "", err
	}
	for _, p := range players {
		if p.Name == name || p.PackageName == name {
			return p.PackageName, nil
		}
	}
	return "", fmt.Errorf("player '%s' not found. Run 'tv player list'", name)
}
````

## File: internal/player/play_test.go
````go
package player

import (
	"strings"
	"testing"

	"github.com/mutedfalcontv/tv/internal/adb"
)

func TestMimeForURL_M3U8(t *testing.T) {
	got := MimeForURL("https://example.com/stream.m3u8")
	want := "application/x-mpegURL"
	if got != want {
		t.Errorf("MimeForURL() = %q, want %q", got, want)
	}
}

func TestMimeForURL_MP4(t *testing.T) {
	got := MimeForURL("https://example.com/video.mp4?token=abc")
	want := "video/mp4"
	if got != want {
		t.Errorf("MimeForURL() = %q, want %q", got, want)
	}
}

func TestMimeForURL_CaseInsensitive(t *testing.T) {
	got := MimeForURL("https://example.com/video.MKV")
	want := "video/x-matroska"
	if got != want {
		t.Errorf("MimeForURL() = %q, want %q", got, want)
	}
}

func TestMimeForURL_Fallback(t *testing.T) {
	got := MimeForURL("https://example.com/file.unknown")
	want := "video/mp4"
	if got != want {
		t.Errorf("MimeForURL() = %q, want %q", got, want)
	}
}

func TestEscapeURLForShell_Simple(t *testing.T) {
	got := EscapeURLForShell("https://example.com/video.mp4")
	want := "'https://example.com/video.mp4'"
	if got != want {
		t.Errorf("EscapeURLForShell() = %q, want %q", got, want)
	}
}

func TestEscapeURLForShell_WithAmpersand(t *testing.T) {
	url := "https://example.com/v?token=a&b=c"
	got := EscapeURLForShell(url)
	if !strings.HasPrefix(got, "'") || !strings.HasSuffix(got, "'") {
		t.Errorf("EscapeURLForShell() = %q, should be single-quote wrapped", got)
	}
}

func TestEscapeURLForShell_WithSingleQuote(t *testing.T) {
	got := EscapeURLForShell("https://example.com/it's.mp4")
	want := "'https://example.com/it'\\''s.mp4'"
	if got != want {
		t.Errorf("EscapeURLForShell() = %q, want %q", got, want)
	}
}

func TestActivityForPlayer_Known(t *testing.T) {
	got := ActivityForPlayer("com.brouken.player")
	want := ".PlayerActivity"
	if got != want {
		t.Errorf("ActivityForPlayer() = %q, want %q", got, want)
	}
}

func TestActivityForPlayer_Unknown(t *testing.T) {
	got := ActivityForPlayer("com.example.player")
	want := ".PlayerActivity"
	if got != want {
		t.Errorf("ActivityForPlayer() = %q, want %q", got, want)
	}
}

func TestPlayOnTV_BuildsIntent(t *testing.T) {
	mock := &adb.MockRunner{}
	cfg := &adb.Config{TVIP: "192.168.2.3:5555"}
	err := PlayOnTV(cfg, mock, "org.videolan.vlc", "https://example.com/v.m3u8")
	if err != nil {
		t.Fatalf("PlayOnTV() unexpected error: %v", err)
	}
	expected := "am start -a android.intent.action.VIEW -d 'https://example.com/v.m3u8' -t application/x-mpegURL -n org.videolan.vlc/.gui.video.VideoPlayerActivity"
	if mock.LastIntent != expected {
		t.Errorf("intent = %q\nwant = %q", mock.LastIntent, expected)
	}
}

func TestPlayOnTV_UnknownPlayerDefaults(t *testing.T) {
	mock := &adb.MockRunner{}
	cfg := &adb.Config{TVIP: "192.168.2.3:5555"}
	err := PlayOnTV(cfg, mock, "com.unknown.player", "https://example.com/v.mp4")
	if err != nil {
		t.Fatalf("PlayOnTV() unexpected error: %v", err)
	}
	if !strings.Contains(mock.LastIntent, "com.unknown.player/.PlayerActivity") {
		t.Errorf("intent should use default .PlayerActivity for unknown players")
	}
}
````

## File: internal/player/play.go
````go
package player

import (
	"fmt"
	"strings"

	"github.com/mutedfalcontv/tv/internal/adb"
)

var mimeByExt = map[string]string{
	"m3u8": "application/x-mpegURL",
	"mp4":  "video/mp4",
	"mkv":  "video/x-matroska",
	"avi":  "video/avi",
	"mov":  "video/quicktime",
	"webm": "video/webm",
}

var playerActivity = map[string]string{
	"com.brouken.player":         ".PlayerActivity",
	"org.videolan.vlc":           ".gui.video.VideoPlayerActivity",
	"app.mpvnova.player":         ".PlayerActivity",
	"com.mxtech.videoplayer.ad":  ".ActivityMedia",
}

func MimeForURL(url string) string {
	lower := strings.ToLower(url)
	for ext, mime := range mimeByExt {
		if strings.HasSuffix(lower, "."+ext) {
			return mime
		}
		if strings.Contains(lower, "."+ext+"?") || strings.Contains(lower, "."+ext+"&") {
			return mime
		}
	}
	return "video/mp4"
}

func EscapeURLForShell(url string) string {
	return "'" + strings.ReplaceAll(url, "'", "'\\''") + "'"
}

func ActivityForPlayer(pkg string) string {
	if act, ok := playerActivity[pkg]; ok {
		return act
	}
	return ".PlayerActivity"
}

func PlayOnTV(cfg *adb.Config, r adb.Runner, playerPkg, url string) error {
	mime := MimeForURL(url)
	activity := ActivityForPlayer(playerPkg)
	quotedURL := EscapeURLForShell(url)
	intent := fmt.Sprintf("am start -a android.intent.action.VIEW -d %s -t %s -n %s/%s",
		quotedURL, mime, playerPkg, activity)
	if m, ok := r.(*adb.MockRunner); ok {
		m.LastIntent = intent
		return nil
	}
	_, err := r.ShellWithStderr(cfg, intent)
	if err != nil {
		return fmt.Errorf("failed to launch player: %w", err)
	}
	return nil
}
````

## File: go.mod
````
module github.com/mutedfalcontv/tv

go 1.26.5
````

## File: internal/adb/runner_test.go
````go
package adb

import (
	"testing"
)

func TestEnsureConnected_AlreadyConnected(t *testing.T) {
	m := &MockRunner{
		DevicesOut: "List of devices attached\n192.168.2.3:5555   device\n",
	}
	cfg := &Config{TVIP: "192.168.2.3:5555"}
	err := EnsureConnected(cfg, m)
	if err != nil {
		t.Fatalf("EnsureConnected() unexpected error: %v", err)
	}
	if m.ConnectCalled {
		t.Error("Connect should not be called when device is already connected")
	}
}

func TestEnsureConnected_NotConnected(t *testing.T) {
	m := &MockRunner{
		DevicesOut: "List of devices attached\n",
		ConnectOut: "connected to 192.168.2.3:5555",
	}
	cfg := &Config{TVIP: "192.168.2.3:5555"}
	err := EnsureConnected(cfg, m)
	if err != nil {
		t.Fatalf("EnsureConnected() unexpected error: %v", err)
	}
	if !m.ConnectCalled {
		t.Error("Connect should be called when device is not connected")
	}
}

func TestEnsureConnected_Offline(t *testing.T) {
	m := &MockRunner{
		DevicesOut: "List of devices attached\n192.168.2.3:5555   offline\n",
		ConnectOut: "connected to 192.168.2.3:5555",
	}
	cfg := &Config{TVIP: "192.168.2.3:5555"}
	err := EnsureConnected(cfg, m)
	if err != nil {
		t.Fatalf("EnsureConnected() unexpected error: %v", err)
	}
	if !m.ConnectCalled {
		t.Error("Connect should be called when device is offline")
	}
}

func TestMockRunnerLastShellCmd(t *testing.T) {
	m := &MockRunner{ShellOut: "output"}
	out, err := m.Shell(&Config{TVIP: "192.168.2.3:5555"}, "input keyevent KEYCODE_HOME")
	if err != nil {
		t.Fatalf("Shell() unexpected error: %v", err)
	}
	if out != "output" {
		t.Errorf("Shell() = %q, want %q", out, "output")
	}
	if m.LastShellCmd != "input keyevent KEYCODE_HOME" {
		t.Errorf("LastShellCmd = %q, want %q", m.LastShellCmd, "input keyevent KEYCODE_HOME")
	}
}

func TestEnsureConnected_ConnectFails(t *testing.T) {
	m := &MockRunner{
		DevicesOut: "List of devices attached\n",
		ConnectErr: assertAnError{},
	}
	cfg := &Config{TVIP: "192.168.2.3:5555"}
	err := EnsureConnected(cfg, m)
	if err == nil {
		t.Fatal("EnsureConnected() expected error, got nil")
	}
}

func TestMockRunnerShellWithStderrCapture(t *testing.T) {
	m := &MockRunner{ShellWithStderrOut: "output"}
	cfg := &Config{TVIP: "192.168.2.3:5555"}
	out, err := m.ShellWithStderr(cfg, "some command")
	if err != nil {
		t.Fatalf("ShellWithStderr() unexpected error: %v", err)
	}
	if out != "output" {
		t.Errorf("ShellWithStderr() = %q, want %q", out, "output")
	}
	if m.LastShellCmd != "some command" {
		t.Errorf("LastShellCmd = %q, want %q", m.LastShellCmd, "some command")
	}
}
````

## File: internal/logcat/logcat_test.go
````go
package logcat

import (
	"testing"
	"github.com/mutedfalcontv/tv/internal/adb"
)

func TestBuildArgs_Default(t *testing.T) {
	opts := Options{}
	args := BuildArgs(opts)
	want := "-v brief"
	if len(args) != 2 || args[0]+" "+args[1] != want {
		t.Errorf("BuildArgs() = %v, want [%s]", args, want)
	}
}

func TestBuildArgs_WithLines(t *testing.T) {
	opts := Options{Lines: 100}
	args := BuildArgs(opts)
	hasLines := false
	for i := range args {
		if args[i] == "-t" && i+1 < len(args) && args[i+1] == "100" {
			hasLines = true
		}
	}
	if !hasLines {
		t.Errorf("BuildArgs() = %v, missing -t 100", args)
	}
}

func TestBuildArgs_WithLevel(t *testing.T) {
	opts := Options{Level: "error"}
	args := BuildArgs(opts)
	hasLevel := false
	for _, a := range args {
		if a == "*:E" {
			hasLevel = true
		}
	}
	if !hasLevel {
		t.Errorf("BuildArgs() = %v, missing *:E", args)
	}
}

func TestBuildArgs_Clear(t *testing.T) {
	opts := Options{Clear: true}
	args := BuildArgs(opts)
	hasClear := false
	for i := range args {
		if args[i] == "-b" && i+1 < len(args) && args[i+1] == "all" {
			hasClear = true
		}
	}
	if !hasClear {
		t.Errorf("BuildArgs() = %v, missing -b all", args)
	}
}

func TestBuildArgs_Dump(t *testing.T) {
	opts := Options{Dump: true}
	args := BuildArgs(opts)
	hasDump := false
	for _, a := range args {
		if a == "-d" {
			hasDump = true
		}
	}
	if !hasDump {
		t.Errorf("BuildArgs() = %v, missing -d", args)
	}
}

func TestBuildArgs_Combined(t *testing.T) {
	opts := Options{Level: "error", Lines: 50, Format: "threadtime"}
	args := BuildArgs(opts)
	combined := ""
	for _, a := range args {
		combined += a + " "
	}
	if !contains(combined, "*:E") || !contains(combined, "-t 50") || !contains(combined, "-v threadtime") {
		t.Errorf("BuildArgs() = %v, missing combined filters", args)
	}
}

func TestResolvePID_ParsesPS(t *testing.T) {
	m := &adb.MockRunner{
		ShellOut: "u0_a123  4567   ... com.brouken.player\n",
	}
	cfg := &adb.Config{TVIP: "192.168.2.3:5555"}
	pid, err := ResolvePID(cfg, m, "com.brouken.player")
	if err != nil {
		t.Fatalf("ResolvePID() unexpected error: %v", err)
	}
	if pid != 4567 {
		t.Errorf("ResolvePID() = %d, want 4567", pid)
	}
}

func TestResolvePID_NotFound(t *testing.T) {
	m := &adb.MockRunner{ShellOut: ""}
	cfg := &adb.Config{TVIP: "192.168.2.3:5555"}
	_, err := ResolvePID(cfg, m, "com.nonexistent")
	if err == nil {
		t.Fatal("ResolvePID() expected error, got nil")
	}
}

func TestBuildArgs_WithPackage(t *testing.T) {
	opts := Options{Package: "1234"}
	args := BuildArgs(opts)
	hasPID := false
	for _, a := range args {
		if a == "--pid=1234" {
			hasPID = true
		}
	}
	if !hasPID {
		t.Errorf("BuildArgs() = %v, missing --pid=1234", args)
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
````

## File: internal/logcat/logcat.go
````go
package logcat

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/mutedfalcontv/tv/internal/adb"
)

type Options struct {
	Package string
	Level   string
	Tag     string
	Lines   int
	Format  string
	Clear   bool
	Dump    bool
}

var levels = map[string]string{
	"v": "*:V", "d": "*:D", "i": "*:I", "w": "*:W", "e": "*:E", "f": "*:F",
	"verbose": "*:V", "debug": "*:D", "info": "*:I",
	"warning": "*:W", "error": "*:E", "fatal": "*:F",
}

func BuildArgs(opts Options) []string {
	var args []string

	if opts.Clear {
		args = append(args, "-b", "all", "-c")
		return args
	}

	if opts.Dump {
		args = append(args, "-d")
	}

	if opts.Lines > 0 {
		args = append(args, "-t", strconv.Itoa(opts.Lines))
	}

	if opts.Package != "" {
		args = append(args, "--pid="+opts.Package)
	}

	if opts.Tag != "" {
		args = append(args, "-s", opts.Tag+":*")
	}

	if opts.Level != "" {
		if l, ok := levels[strings.ToLower(opts.Level)]; ok {
			args = append(args, l)
		}
	}

	format := opts.Format
	if format == "" {
		format = "brief"
	}
	args = append(args, "-v", format)

	return args
}

func ResolvePID(cfg *adb.Config, r adb.Runner, pkg string) (int, error) {
	out, err := r.Shell(cfg, "ps -A | grep "+pkg)
	if err != nil {
		return 0, fmt.Errorf("failed to list processes: %w", err)
	}
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			pid, err := strconv.Atoi(fields[1])
			if err == nil {
				return pid, nil
			}
		}
	}
	return 0, fmt.Errorf("process not found for package: %s", pkg)
}

func RunStream(cfg *adb.Config, opts Options) error {
	adbPath, err := exec.LookPath("adb")
	if err != nil {
		return fmt.Errorf("ADB not found on PATH")
	}
	logArgs := BuildArgs(opts)
	cmd := exec.Command(adbPath, append([]string{"-s", cfg.TVIP, "logcat"}, logArgs...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func RunOnce(cfg *adb.Config, r adb.Runner, opts Options) (string, error) {
	logArgs := BuildArgs(opts)
	cmd := "logcat " + strings.Join(logArgs, " ")
	return r.ShellWithStderr(cfg, cmd)
}
````

## File: internal/remote/remote_test.go
````go
package remote

import (
	"testing"
	"github.com/mutedfalcontv/tv/internal/adb"
)

func TestPress_NavKey(t *testing.T) {
	m := &adb.MockRunner{}
	cfg := &adb.Config{TVIP: "192.168.2.3:5555"}
	err := Press(cfg, m, "up")
	if err != nil {
		t.Fatalf("Press() unexpected error: %v", err)
	}
	if m.LastShellCmd != "input keyevent KEYCODE_DPAD_UP" {
		t.Errorf("got %q, want %q", m.LastShellCmd, "input keyevent KEYCODE_DPAD_UP")
	}
}

func TestPress_VolumeKey(t *testing.T) {
	m := &adb.MockRunner{}
	cfg := &adb.Config{TVIP: "192.168.2.3:5555"}
	err := Press(cfg, m, "volup")
	if err != nil {
		t.Fatalf("Press() unexpected error: %v", err)
	}
	if m.LastShellCmd != "input keyevent KEYCODE_VOLUME_UP" {
		t.Errorf("got %q, want %q", m.LastShellCmd, "input keyevent KEYCODE_VOLUME_UP")
	}
}

func TestPress_MediaKey(t *testing.T) {
	m := &adb.MockRunner{}
	cfg := &adb.Config{TVIP: "192.168.2.3:5555"}
	err := Press(cfg, m, "play")
	if err != nil {
		t.Fatalf("Press() unexpected error: %v", err)
	}
	if m.LastShellCmd != "input keyevent KEYCODE_MEDIA_PLAY" {
		t.Errorf("got %q, want %q", m.LastShellCmd, "input keyevent KEYCODE_MEDIA_PLAY")
	}
}

func TestPress_NumberKey(t *testing.T) {
	m := &adb.MockRunner{}
	cfg := &adb.Config{TVIP: "192.168.2.3:5555"}
	err := Press(cfg, m, "5")
	if err != nil {
		t.Fatalf("Press() unexpected error: %v", err)
	}
	if m.LastShellCmd != "input keyevent KEYCODE_5" {
		t.Errorf("got %q, want %q", m.LastShellCmd, "input keyevent KEYCODE_5")
	}
}

func TestPress_UnknownKey(t *testing.T) {
	m := &adb.MockRunner{}
	cfg := &adb.Config{TVIP: "192.168.2.3:5555"}
	err := Press(cfg, m, "nonexistent")
	if err == nil {
		t.Fatal("Press() expected error, got nil")
	}
}

func TestPress_SystemKey(t *testing.T) {
	m := &adb.MockRunner{}
	cfg := &adb.Config{TVIP: "192.168.2.3:5555"}
	err := Press(cfg, m, "home")
	if err != nil {
		t.Fatalf("Press() unexpected error: %v", err)
	}
	if m.LastShellCmd != "input keyevent KEYCODE_HOME" {
		t.Errorf("got %q, want %q", m.LastShellCmd, "input keyevent KEYCODE_HOME")
	}
}

func TestType_SendsInputText(t *testing.T) {
	m := &adb.MockRunner{}
	cfg := &adb.Config{TVIP: "192.168.2.3:5555"}
	err := Type(cfg, m, "hello")
	if err != nil {
		t.Fatalf("Type() unexpected error: %v", err)
	}
	if m.LastShellCmd != "input text 'hello'" {
		t.Errorf("got %q, want %q", m.LastShellCmd, "input text 'hello'")
	}
}

func TestType_WithSpaces(t *testing.T) {
	m := &adb.MockRunner{}
	cfg := &adb.Config{TVIP: "192.168.2.3:5555"}
	err := Type(cfg, m, "search term")
	if err != nil {
		t.Fatalf("Type() unexpected error: %v", err)
	}
	if m.LastShellCmd != "input text 'search term'" {
		t.Errorf("got %q, want %q", m.LastShellCmd, "input text 'search term'")
	}
}

func TestKeys_ReturnsSorted(t *testing.T) {
	keys := Keys()
	if len(keys) < 20 {
		t.Errorf("got %d keys, want at least 20", len(keys))
	}
	if keys[0] > keys[1] {
		t.Errorf("keys not sorted: %s > %s", keys[0], keys[1])
	}
}

func TestType_WithSingleQuote(t *testing.T) {
	m := &adb.MockRunner{}
	cfg := &adb.Config{TVIP: "192.168.2.3:5555"}
	err := Type(cfg, m, "it's")
	if err != nil {
		t.Fatalf("Type() unexpected error: %v", err)
	}
	want := "input text 'it'\\''s'"
	if m.LastShellCmd != want {
		t.Errorf("got %q, want %q", m.LastShellCmd, want)
	}
}
````

## File: internal/remote/remote.go
````go
package remote

import (
	"fmt"
	"sort"
	"strings"

	"github.com/mutedfalcontv/tv/internal/adb"
)

var keyMap = map[string]string{
	"up":       "KEYCODE_DPAD_UP",
	"down":     "KEYCODE_DPAD_DOWN",
	"left":     "KEYCODE_DPAD_LEFT",
	"right":    "KEYCODE_DPAD_RIGHT",
	"ok":       "KEYCODE_DPAD_CENTER",
	"home":     "KEYCODE_HOME",
	"back":     "KEYCODE_BACK",
	"menu":     "KEYCODE_MENU",
	"power":    "KEYCODE_POWER",
	"input":    "KEYCODE_TV_INPUT",
	"info":     "KEYCODE_INFO",
	"subtitle": "KEYCODE_CAPTIONS",
	"volup":    "KEYCODE_VOLUME_UP",
	"voldown":  "KEYCODE_VOLUME_DOWN",
	"mute":     "KEYCODE_VOLUME_MUTE",
	"play":     "KEYCODE_MEDIA_PLAY",
	"pause":    "KEYCODE_MEDIA_PAUSE",
	"stop":     "KEYCODE_MEDIA_STOP",
	"ff":       "KEYCODE_MEDIA_FAST_FORWARD",
	"rew":      "KEYCODE_MEDIA_REWIND",
	"next":     "KEYCODE_MEDIA_NEXT",
	"prev":     "KEYCODE_MEDIA_PREVIOUS",
	"chup":     "KEYCODE_CHANNEL_UP",
	"chdown":   "KEYCODE_CHANNEL_DOWN",
	"0":        "KEYCODE_0",
	"1":        "KEYCODE_1",
	"2":        "KEYCODE_2",
	"3":        "KEYCODE_3",
	"4":        "KEYCODE_4",
	"5":        "KEYCODE_5",
	"6":        "KEYCODE_6",
	"7":        "KEYCODE_7",
	"8":        "KEYCODE_8",
	"9":        "KEYCODE_9",
}

func Press(cfg *adb.Config, r adb.Runner, key string) error {
	keycode, ok := keyMap[key]
	if !ok {
		return fmt.Errorf("unknown key: %s. Available keys: %v", key, Keys())
	}
	_, err := r.Shell(cfg, "input keyevent "+keycode)
	if err != nil {
		return fmt.Errorf("failed to send keyevent: %w", err)
	}
	return nil
}

func Type(cfg *adb.Config, r adb.Runner, text string) error {
	escaped := "'" + strings.ReplaceAll(text, "'", "'\\''") + "'"
	_, err := r.Shell(cfg, "input text "+escaped)
	if err != nil {
		return fmt.Errorf("failed to input text: %w", err)
	}
	return nil
}

func Keys() []string {
	var names []string
	for k := range keyMap {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}
````

## File: .gitignore
````
*.exe
*.test
*.out
*.log
tv.exe
tv-windows-*.exe
tv-linux-*
tv-darwin-*
.worktrees/
````

## File: Makefile
````makefile
BINARY := tv
GO := go

.PHONY: build test clean build-all

build:
	$(GO) build -o $(BINARY).exe ./cmd/$(BINARY)/

test:
	$(GO) test -v ./...

build-all:
	GOOS=windows GOARCH=amd64 $(GO) build -o $(BINARY)-windows-amd64.exe ./cmd/$(BINARY)/
	GOOS=linux GOARCH=amd64 $(GO) build -o $(BINARY)-linux-amd64 ./cmd/$(BINARY)/
	GOOS=darwin GOARCH=amd64 $(GO) build -o $(BINARY)-darwin-amd64 ./cmd/$(BINARY)/

clean:
	del /f /q $(BINARY).exe $(BINARY)-windows-*.exe $(BINARY)-linux-* $(BINARY)-darwin-* 2>nul || true
````

## File: internal/adb/runner.go
````go
package adb

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/mutedfalcontv/tv/internal/config"
)

type Config struct {
	TVIP          string
	DefaultPlayer string
}

type Runner interface {
	Devices() (string, error)
	Connect(ip string) (string, error)
	Disconnect(ip string) (string, error)
	Shell(cfg *Config, cmd string) (string, error)
	ShellWithStderr(cfg *Config, cmd string) (string, error)
}

type RealRunner struct{}

func (r *RealRunner) binary() (string, error) {
	path, err := exec.LookPath("adb")
	if err != nil {
		return "", fmt.Errorf("ADB not found on PATH")
	}
	return path, nil
}

func (r *RealRunner) Devices() (string, error) {
	adb, err := r.binary()
	if err != nil {
		return "", err
	}
	out, err := exec.Command(adb, "devices").Output()
	if err != nil {
		return "", fmt.Errorf("adb devices: %w", err)
	}
	return string(out), nil
}

func (r *RealRunner) Connect(ip string) (string, error) {
	adb, err := r.binary()
	if err != nil {
		return "", err
	}
	out, err := exec.Command(adb, "connect", ip).CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func (r *RealRunner) Disconnect(ip string) (string, error) {
	adb, err := r.binary()
	if err != nil {
		return "", err
	}
	out, err := exec.Command(adb, "disconnect", ip).CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func (r *RealRunner) Shell(cfg *Config, cmd string) (string, error) {
	adb, err := r.binary()
	if err != nil {
		return "", err
	}
	out, err := exec.Command(adb, "-s", cfg.TVIP, "shell", cmd).Output()
	if err != nil {
		return "", fmt.Errorf("ADB error: %w\n%s", err, string(out))
	}
	return strings.TrimSpace(string(out)), nil
}

func (r *RealRunner) ShellWithStderr(cfg *Config, cmd string) (string, error) {
	adb, err := r.binary()
	if err != nil {
		return "", err
	}
	out, err := exec.Command(adb, "-s", cfg.TVIP, "shell", cmd).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("ADB error: %w\n%s", err, string(out))
	}
	return strings.TrimSpace(string(out)), nil
}

func GetTVIP() string {
	if v := os.Getenv("TV_IP"); v != "" {
		return v
	}
	cfg, err := config.Load()
	if err != nil {
		return "192.168.2.3:5555"
	}
	return cfg.TVIP
}

type assertAnError struct{}

func (assertAnError) Error() string { return "expected error" }

func EnsureConnected(cfg *Config, r Runner) error {
	out, err := r.Devices()
	if err != nil {
		return err
	}
	if strings.Contains(out, cfg.TVIP) && !strings.Contains(out, "offline") {
		return nil
	}
	_, err = r.Connect(cfg.TVIP)
	return err
}

type MockRunner struct {
	DevicesOut    string
	DevicesErr    error
	ConnectOut    string
	ConnectErr    error
	DisconnectOut string
	DisconnectErr error
	ShellOut      string
	ShellErr      error
	ShellWithStderrOut string
	ShellWithStderrErr error
	ConnectCalled    bool
	DisconnectCalled bool
	LastIntent       string
	LastShellCmd     string
}

func (m *MockRunner) Devices() (string, error)                    { return m.DevicesOut, m.DevicesErr }
func (m *MockRunner) Connect(ip string) (string, error)           { m.ConnectCalled = true; return m.ConnectOut, m.ConnectErr }
func (m *MockRunner) Disconnect(ip string) (string, error)        { m.DisconnectCalled = true; return m.DisconnectOut, m.DisconnectErr }
func (m *MockRunner) Shell(cfg *Config, cmd string) (string, error)       { m.LastShellCmd = cmd; return m.ShellOut, m.ShellErr }
func (m *MockRunner) ShellWithStderr(cfg *Config, cmd string) (string, error) {
	m.LastShellCmd = cmd
	return m.ShellWithStderrOut, m.ShellWithStderrErr
}
````

## File: cmd/tv/main.go
````go
package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/mutedfalcontv/tv/internal/adb"
	"github.com/mutedfalcontv/tv/internal/config"
	"github.com/mutedfalcontv/tv/internal/logcat"
	"github.com/mutedfalcontv/tv/internal/player"
	"github.com/mutedfalcontv/tv/internal/remote"
)

type command struct {
	name string
	desc string
	run  func(args []string)
}

var commands = []command{
	{"adb", "ADB connection management", adbCmd},
	{"app", "App management (list, launch, kill)", appCmd},
	{"config", "Show configuration", configCmd},
	{"logs", "View and filter TV logs (logcat)", logsCmd},
	{"play", "Play a URL on TV", playCmd},
	{"player", "List/set default video player", playerCmd},
	{"remote", "TV remote control", remoteCmd},
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
	sub := os.Args[1]
	for _, c := range commands {
		if c.name == sub {
			c.run(os.Args[2:])
			return
		}
	}
	fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", sub)
	printUsage()
	os.Exit(1)
}

func printUsage() {
	fmt.Fprint(os.Stderr, `tv — Android TV CLI tools

Usage:
  tv <command> [options] [args]

Commands:
`)
	for _, c := range commands {
		fmt.Fprintf(os.Stderr, "  %-15s %s\n", c.name, c.desc)
	}
}

func adbCmd(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: tv adb <connect|disconnect>")
		os.Exit(1)
	}
	r := &adb.RealRunner{}
	ip := adb.GetTVIP()
	switch args[0] {
	case "connect":
		fmt.Printf("Connecting to %s...\n", ip)
		result, err := r.Connect(ip)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(result)
	case "disconnect":
		fmt.Printf("Disconnecting from %s...\n", ip)
		result, err := r.Disconnect(ip)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(result)
	default:
		fmt.Fprintf(os.Stderr, "unknown adb subcommand: %s\n", args[0])
		os.Exit(1)
	}
}

func appCmd(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: tv app <list|launch|kill> [package]")
		os.Exit(1)
	}

	cfg := loadAdbConfig()
	r := &adb.RealRunner{}

	switch args[0] {
	case "list":
		if err := adb.EnsureConnected(cfg, r); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		out, err := r.Shell(cfg, "pm list packages")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(out)
	case "launch":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: tv app launch <package>")
			os.Exit(1)
		}
		if err := adb.EnsureConnected(cfg, r); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		_, err := r.ShellWithStderr(cfg, "monkey -p "+args[1]+" -c android.intent.category.LAUNCHER 1")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error launching %s: %v\n", args[1], err)
			os.Exit(1)
		}
		fmt.Printf("Launched %s\n", args[1])
	case "kill":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: tv app kill <package>")
			os.Exit(1)
		}
		if err := adb.EnsureConnected(cfg, r); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		_, err := r.ShellWithStderr(cfg, "am force-stop "+args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error killing %s: %v\n", args[1], err)
			os.Exit(1)
		}
		fmt.Printf("Killed %s\n", args[1])
	default:
		fmt.Fprintf(os.Stderr, "unknown app subcommand: %s\n", args[0])
		os.Exit(1)
	}
}

func configCmd(args []string) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Config file: ~/.config/tv/config.json\n")
	fmt.Printf("TV IP:        %s\n", cfg.TVIP)
	if cfg.DefaultPlayer != "" {
		fmt.Printf("Default player: %s\n", cfg.DefaultPlayer)
	} else {
		fmt.Println("Default player: (none)")
	}
}

func playerCmd(args []string) {
	cfg := loadAdbConfig()
	r := &adb.RealRunner{}

	fs := flag.NewFlagSet("player", flag.ExitOnError)
	ip := fs.String("ip", cfg.TVIP, "TV IP address")
	fs.Parse(args)
	cfg.TVIP = *ip

	remaining := fs.Args()
	if len(remaining) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: tv player <list|default> [options] [name]")
		os.Exit(1)
	}

	switch remaining[0] {
	case "list":
		players, err := player.Detect(cfg, r)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if len(players) == 0 {
			fmt.Println("No video players found on TV.")
			return
		}
		fmt.Println("Installed video players:")
		for _, p := range players {
			mark := " "
			if cfg.DefaultPlayer == p.PackageName {
				mark = "*"
			}
			fmt.Printf("  %s %-20s %s\n", mark, p.Name, p.PackageName)
		}
	case "default":
		if len(remaining) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: tv player default <name>")
			os.Exit(1)
		}
		name := remaining[1]
		pkg, err := player.Resolve(cfg, r, name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		conf, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		conf.DefaultPlayer = pkg
		if err := conf.Save(); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Default player set to: %s\n", pkg)
	default:
		fmt.Fprintf(os.Stderr, "unknown player subcommand: %s\n", remaining[0])
		os.Exit(1)
	}
}

func playCmd(args []string) {
	cfg := loadAdbConfig()
	r := &adb.RealRunner{}

	fs := flag.NewFlagSet("play", flag.ExitOnError)
	playerFlag := fs.String("player", cfg.DefaultPlayer, "Player name or package name")
	ip := fs.String("ip", cfg.TVIP, "TV IP address")
	fs.Parse(args)
	cfg.TVIP = *ip

	remaining := fs.Args()
	if len(remaining) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: tv play [options] <url>")
		fmt.Fprintln(os.Stderr, "  -player <name>   Player to use")
		fmt.Fprintln(os.Stderr, "  -ip <address>    TV IP address")
		os.Exit(1)
	}

	url := remaining[0]
	playerPkg := *playerFlag
	if playerPkg == "" {
		fmt.Fprintln(os.Stderr, "No default player set. Use 'tv player default <name>' or 'tv play -player <name> <url>'")
		os.Exit(1)
	}
	if !strings.Contains(playerPkg, ".") {
		pkg, err := player.Resolve(cfg, r, playerPkg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		playerPkg = pkg
	}
	if err := adb.EnsureConnected(cfg, r); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Playing on %s via %s...\n", cfg.TVIP, playerPkg)
	if err := player.PlayOnTV(cfg, r, playerPkg, url); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Done.")
}

func remoteCmd(args []string) {
	cfg := loadAdbConfig()
	r := &adb.RealRunner{}

	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: tv remote <key>")
		fmt.Fprintln(os.Stderr, "Keys: up, down, left, right, ok, home, back, menu, power,")
		fmt.Fprintln(os.Stderr, "      input, info, subtitle, volup, voldown, mute,")
		fmt.Fprintln(os.Stderr, "      play, pause, stop, ff, rew, next, prev,")
		fmt.Fprintln(os.Stderr, "      chup, chdown, 0-9, type <text>")
		os.Exit(1)
	}

	cmd := args[0]
	if cmd == "type" {
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: tv remote type <text>")
			os.Exit(1)
		}
		if err := adb.EnsureConnected(cfg, r); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if err := remote.Type(cfg, r, args[1]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Typed: %s\n", args[1])
		return
	}
	if err := adb.EnsureConnected(cfg, r); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if err := remote.Press(cfg, r, cmd); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Sent key: %s\n", cmd)
}

func logsCmd(args []string) {
	cfg := loadAdbConfig()
	r := &adb.RealRunner{}
	var opts logcat.Options

	fs := flag.NewFlagSet("logs", flag.ExitOnError)
	packageFlag := fs.String("p", "", "Filter by package name")
	levelFlag := fs.String("l", "", "Log level: v/d/i/w/e/f")
	tagFlag := fs.String("t", "", "Filter by log tag")
	linesFlag := fs.Int("n", 0, "Number of recent lines")
	formatFlag := fs.String("v", "", "Log format: brief, threadtime, etc.")
	clearFlag := fs.Bool("c", false, "Clear log buffer")
	dumpFlag := fs.Bool("d", false, "Dump and exit (no follow)")
	fs.Parse(args)

	opts.Level = *levelFlag
	opts.Tag = *tagFlag
	opts.Lines = *linesFlag
	opts.Format = *formatFlag
	opts.Clear = *clearFlag
	opts.Dump = *dumpFlag

	if len(fs.Args()) > 0 && fs.Args()[0] == "dump" {
		opts.Dump = true
	}

	if err := adb.EnsureConnected(cfg, r); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if *packageFlag != "" {
		pkg := *packageFlag
		if !strings.Contains(pkg, ".") {
			var err error
			pkg, err = player.Resolve(cfg, r, pkg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		}
		pid, err := logcat.ResolvePID(cfg, r, pkg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		opts.Package = strconv.Itoa(pid)
		fmt.Printf("Filtering by %s (PID %d)\n", pkg, pid)
	}

	if opts.Clear {
		_, err := logcat.RunOnce(cfg, r, opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Log buffer cleared.")
		return
	}

	if opts.Dump || opts.Lines > 0 || opts.Level != "" || opts.Tag != "" {
		out, err := logcat.RunOnce(cfg, r, opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(out)
		return
	}

	fmt.Println("Streaming logs (Ctrl-C to stop)...")
	if err := logcat.RunStream(cfg, opts); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func loadAdbConfig() *adb.Config {
	c, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: config load failed: %v (using defaults)\n", err)
		return &adb.Config{TVIP: "192.168.2.3:5555"}
	}
	return &adb.Config{TVIP: c.TVIP, DefaultPlayer: c.DefaultPlayer}
}
````
