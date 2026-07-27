package player

import (
	"fmt"
	"sort"
	"strings"

	"github.com/mutedfalcontv/tv/pkg/adb"
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
