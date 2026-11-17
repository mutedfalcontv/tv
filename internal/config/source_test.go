package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadWithSource_DefaultsWhenMissing(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)

	l := LoadWithSource()
	if l.Source != SourceDefaults {
		t.Errorf("Source = %v, want %v", l.Source, SourceDefaults)
	}
	if !l.Defaults() {
		t.Error("Defaults() = false, want true")
	}
	if !l.SafeToSave() {
		t.Error("SafeToSave() = false, want true for a missing file")
	}
	if l.Err() != nil {
		t.Errorf("Err() = %v, want nil", l.Err())
	}
	if l.Config.TVIP == "" {
		t.Error("Config.TVIP is empty, want the default")
	}
}

func TestLoadWithSource_File(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	if err := (&Config{TVIP: "10.0.0.9:5555"}).Save(); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	l := LoadWithSource()
	if l.Source != SourceFile {
		t.Errorf("Source = %v, want %v", l.Source, SourceFile)
	}
	if l.Defaults() {
		t.Error("Defaults() = true, want false")
	}
	if l.Config.TVIP != "10.0.0.9:5555" {
		t.Errorf("TVIP = %q, want the saved value", l.Config.TVIP)
	}
	if filepath.Base(l.Path) != "config.json" {
		t.Errorf("Path = %q, want the config file", l.Path)
	}
}

func TestLoadWithSource_InvalidIsUnsafeToSave(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	dir, err := Dir()
	if err != nil {
		t.Fatalf("Dir() error: %v", err)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte("{not json"), 0o644); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	l := LoadWithSource()
	if l.Source != SourceInvalid {
		t.Errorf("Source = %v, want %v", l.Source, SourceInvalid)
	}
	if l.Err() == nil {
		t.Error("Err() = nil, want the parse error")
	}
	if l.Defaults() {
		t.Error("Defaults() = true, want false: an unreadable file is not the same as no file")
	}
	if l.SafeToSave() {
		t.Error("SafeToSave() = true, want false: saving would destroy the unreadable file")
	}
}

// Load must keep its previous behaviour: an unreadable file is an error, so
// callers that already handle errors are unaffected.
func TestLoad_StillReturnsErrorOnInvalid(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	dir, err := Dir()
	if err != nil {
		t.Fatalf("Dir() error: %v", err)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte("{not json"), 0o644); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}
	if _, err := Load(); err == nil {
		t.Fatal("Load() = nil error, want a parse error")
	}
}
