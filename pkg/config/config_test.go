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
