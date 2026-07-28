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
