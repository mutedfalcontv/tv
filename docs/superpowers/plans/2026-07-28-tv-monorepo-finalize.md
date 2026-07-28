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
