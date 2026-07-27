package logcat

import (
	"testing"
	"github.com/mutedfalcontv/tv/pkg/adb"
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

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
