package player

import (
	"testing"

	"github.com/mutedfalcontv/tv/pkg/adb"
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
