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
