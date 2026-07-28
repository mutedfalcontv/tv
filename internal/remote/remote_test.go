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
