package logcat

import (
	"github.com/mutedfalcontv/tv/internal/adb"
	"testing"
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

// TestOnceArgs_AlwaysDumps is the regression test for the hanging GET /api/logs
// bug: RunOnce built args without -d when Dump was false, so `adb shell logcat`
// streamed forever. onceArgs (the RunOnce path) must always include -d, so
// BuildArgs must emit -d whenever the options reach it from RunOnce.
func TestOnceArgs_AlwaysDumps(t *testing.T) {
	tests := []struct {
		name string
		opts Options
	}{
		{"empty", Options{}},
		{"level", Options{Level: "error"}},
		{"tag", Options{Tag: "ActivityManager"}},
		{"lines", Options{Lines: 100}},
		{"package", Options{Package: "1234"}},
		{"format", Options{Format: "threadtime"}},
		{"dump-explicit", Options{Dump: true}},
		{"combined", Options{Level: "info", Tag: "VLC", Lines: 50, Package: "42"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := onceArgs(tt.opts)
			if !hasPair(args, "-d") {
				t.Fatalf("onceArgs(%+v) = %v, missing -d (would stream forever)", tt.opts, args)
			}
			// -d must survive alongside a bounded -t line count too.
			if tt.opts.Lines > 0 && !hasPair(args, "-t") {
				t.Fatalf("onceArgs(%+v) = %v, missing -t", tt.opts, args)
			}
		})
	}
}

func hasPair(args []string, flag string) bool {
	for _, a := range args {
		if a == flag {
			return true
		}
	}
	return false
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestValidate_RejectsUnsafeFields(t *testing.T) {
	tests := []struct {
		name string
		opts Options
	}{
		{"tag with semicolon", Options{Tag: "ActivityManager; reboot"}},
		{"tag with pipe", Options{Tag: "VLC | reboot"}},
		{"format with space", Options{Format: "brief extra"}},
		{"non-numeric pid", Options{Package: "1234; reboot"}},
	}
	for _, tt := range tests {
		if err := Validate(tt.opts); err == nil {
			t.Errorf("Validate(%+v) = nil, want error", tt.opts)
		}
	}
}

func TestValidate_AcceptsSafeFields(t *testing.T) {
	opts := Options{Tag: "ActivityManager", Format: "threadtime", Package: "1234"}
	if err := Validate(opts); err != nil {
		t.Errorf("Validate(%+v) = %v, want nil", opts, err)
	}
}

func TestRunOnce_RejectsInvalidTagBeforeShell(t *testing.T) {
	m := &adb.MockRunner{}
	_, err := RunOnce(&adb.Config{TVIP: "192.168.2.3:5555"}, m, Options{Tag: "x; reboot"})
	if err == nil {
		t.Fatal("RunOnce() expected error for unsafe tag, got nil")
	}
	if m.LastShellCmd != "" {
		t.Errorf("shell ran despite invalid tag: %q", m.LastShellCmd)
	}
}

func TestResolvePID_RejectsInvalidPackage(t *testing.T) {
	m := &adb.MockRunner{}
	_, err := ResolvePID(&adb.Config{TVIP: "192.168.2.3:5555"}, m, "com.x; reboot")
	if err == nil {
		t.Fatal("ResolvePID() expected error for invalid package, got nil")
	}
	if m.LastShellCmd != "" {
		t.Errorf("shell ran despite invalid package: %q", m.LastShellCmd)
	}
}
