package mcp

import (
	"context"
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mutedfalcontv/tv/internal/adb"
)

const testIP = "192.168.2.3:5555"

func newTestServer(m *adb.MockRunner) *Server {
	return New(&adb.Config{TVIP: testIP, DefaultPlayer: "org.videolan.vlc"}, m)
}

func connectedMock() *adb.MockRunner {
	return &adb.MockRunner{
		DevicesOut: "List of devices attached\n" + testIP + "   device\n",
	}
}

func TestHandleStatus_Connected(t *testing.T) {
	s := newTestServer(connectedMock())
	_, out, err := s.handleStatus(context.Background(), nil, statusIn{})
	if err != nil {
		t.Fatalf("handleStatus() error: %v", err)
	}
	if !out.Connected {
		t.Error("Connected = false, want true")
	}
	if out.TVIP != testIP {
		t.Errorf("TVIP = %q, want %q", out.TVIP, testIP)
	}
	if out.DefaultPlayer != "org.videolan.vlc" {
		t.Errorf("DefaultPlayer = %q", out.DefaultPlayer)
	}
}

func TestHandleStatus_Disconnected(t *testing.T) {
	m := &adb.MockRunner{
		DevicesOut: "List of devices attached\n",
		ConnectErr: errTest,
	}
	s := newTestServer(m)
	_, out, err := s.handleStatus(context.Background(), nil, statusIn{})
	if err != nil {
		t.Fatalf("handleStatus() error: %v", err)
	}
	if out.Connected {
		t.Error("Connected = true, want false")
	}
}

func TestHandleListApps_SortedAndParsed(t *testing.T) {
	m := connectedMock()
	m.ShellOut = "package:org.videolan.vlc\npackage:com.netflix.ninja\n\n"
	s := newTestServer(m)
	_, out, err := s.handleListApps(context.Background(), nil, listAppsIn{})
	if err != nil {
		t.Fatalf("handleListApps() error: %v", err)
	}
	if out.Count != 2 {
		t.Fatalf("Count = %d, want 2", out.Count)
	}
	want := []string{"com.netflix.ninja", "org.videolan.vlc"}
	for i := range want {
		if out.Packages[i] != want[i] {
			t.Errorf("Packages[%d] = %q, want %q", i, out.Packages[i], want[i])
		}
	}
}

func TestHandleLaunchApp(t *testing.T) {
	m := connectedMock()
	s := newTestServer(m)
	_, out, err := s.handleLaunchApp(context.Background(), nil, packageIn{Package: "com.netflix.ninja"})
	if err != nil {
		t.Fatalf("handleLaunchApp() error: %v", err)
	}
	if !out.OK || out.Action != "launch" {
		t.Errorf("unexpected output: %+v", out)
	}
	if !strings.Contains(m.LastShellCmd, "monkey -p com.netflix.ninja") {
		t.Errorf("LastShellCmd = %q", m.LastShellCmd)
	}
}

func TestHandleKillApp(t *testing.T) {
	m := connectedMock()
	s := newTestServer(m)
	_, out, err := s.handleKillApp(context.Background(), nil, packageIn{Package: "com.netflix.ninja"})
	if err != nil {
		t.Fatalf("handleKillApp() error: %v", err)
	}
	if !out.OK || out.Action != "kill" {
		t.Errorf("unexpected output: %+v", out)
	}
	if m.LastShellCmd != "am force-stop com.netflix.ninja" {
		t.Errorf("LastShellCmd = %q", m.LastShellCmd)
	}
}

func TestHandleLaunchApp_RejectsInjection(t *testing.T) {
	s := newTestServer(connectedMock())
	_, _, err := s.handleLaunchApp(context.Background(), nil, packageIn{Package: "com.x; rm -rf /"})
	if err == nil {
		t.Fatal("expected error for invalid package name")
	}
}

func TestHandlePressKey(t *testing.T) {
	m := connectedMock()
	s := newTestServer(m)
	_, out, err := s.handlePressKey(context.Background(), nil, pressKeyIn{Key: "up"})
	if err != nil {
		t.Fatalf("handlePressKey() error: %v", err)
	}
	if !out.OK || out.Key != "up" {
		t.Errorf("unexpected output: %+v", out)
	}
	if m.LastShellCmd != "input keyevent KEYCODE_DPAD_UP" {
		t.Errorf("LastShellCmd = %q", m.LastShellCmd)
	}
}

func TestHandlePressKey_Unknown(t *testing.T) {
	s := newTestServer(connectedMock())
	_, _, err := s.handlePressKey(context.Background(), nil, pressKeyIn{Key: "nope"})
	if err == nil {
		t.Fatal("expected error for unknown key")
	}
}

func TestHandleTypeText(t *testing.T) {
	m := connectedMock()
	s := newTestServer(m)
	_, out, err := s.handleTypeText(context.Background(), nil, typeTextIn{Text: "hello world"})
	if err != nil {
		t.Fatalf("handleTypeText() error: %v", err)
	}
	if !out.OK {
		t.Error("OK = false")
	}
	if m.LastShellCmd != "input text 'hello world'" {
		t.Errorf("LastShellCmd = %q", m.LastShellCmd)
	}
}

func TestHandlePlay_ResolvesNamedPlayer(t *testing.T) {
	m := connectedMock()
	m.ShellOut = "package:org.videolan.vlc"
	s := newTestServer(m)
	if _, _, err := s.handlePlay(context.Background(), nil, playIn{URL: "http://x/video.mp4", Player: "VLC"}); err != nil {
		t.Fatalf("handlePlay() error: %v", err)
	}
	if !strings.Contains(m.LastShellCmd, "org.videolan.vlc") {
		t.Errorf("LastShellCmd = %q, want VLC intent", m.LastShellCmd)
	}
}

func TestHandlePlay_NoPlayer(t *testing.T) {
	m := connectedMock()
	s := New(&adb.Config{TVIP: testIP}, m)
	if _, _, err := s.handlePlay(context.Background(), nil, playIn{URL: "http://x/v.mp4"}); err == nil {
		t.Fatal("expected error with no player configured")
	}
}

func TestHandlePlayers(t *testing.T) {
	m := connectedMock()
	m.ShellOut = "package:org.videolan.vlc\npackage:com.brouken.player"
	s := newTestServer(m)
	_, out, err := s.handlePlayers(context.Background(), nil, playersIn{})
	if err != nil {
		t.Fatalf("handlePlayers() error: %v", err)
	}
	if out.Count != 2 {
		t.Fatalf("Count = %d, want 2", out.Count)
	}
}

func TestHandleLogs_Bounded(t *testing.T) {
	m := connectedMock()
	m.ShellWithStderrOut = "line1\nline2\nline3\nline4\nline5"
	s := newTestServer(m)
	_, out, err := s.handleLogs(context.Background(), nil, logsIn{Lines: 3})
	if err != nil {
		t.Fatalf("handleLogs() error: %v", err)
	}
	if out.Count != 3 {
		t.Fatalf("Count = %d, want 3", out.Count)
	}
	if !out.Truncated {
		t.Error("Truncated = false, want true")
	}
	if out.Lines[2] != "line5" {
		t.Errorf("last line = %q, want line5", out.Lines[2])
	}
}

func TestNormalizeLines(t *testing.T) {
	if got, _ := normalizeLines(0); got != defaultLogLines {
		t.Errorf("normalizeLines(0) = %d, want %d", got, defaultLogLines)
	}
	if got, trunc := normalizeLines(maxLogLines + 500); got != maxLogLines || !trunc {
		t.Errorf("normalizeLines(over) = %d, %v", got, trunc)
	}
	if got, trunc := normalizeLines(42); got != 42 || trunc {
		t.Errorf("normalizeLines(42) = %d, %v", got, trunc)
	}
}

func TestHandleScreenshot_SavesFile(t *testing.T) {
	m := connectedMock()
	png := []byte{0x89, 0x50, 0x4e, 0x47}
	m.ExecOutOut = png
	s := newTestServer(m)
	path := filepath.Join(t.TempDir(), "shot.png")

	_, out, err := s.handleScreenshot(context.Background(), nil, screenshotIn{Path: path})
	if err != nil {
		t.Fatalf("handleScreenshot() error: %v", err)
	}
	if out.Path != path || out.Bytes != len(png) {
		t.Errorf("unexpected output: %+v", out)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading screenshot: %v", err)
	}
	if string(data) != string(png) {
		t.Error("saved bytes do not match")
	}
	if strings.Join(m.LastExecOutArgs, " ") != "screencap -p" {
		t.Errorf("ExecOut args = %v", m.LastExecOutArgs)
	}
}

func TestHandleScreenshot_Base64(t *testing.T) {
	m := connectedMock()
	png := []byte{1, 2, 3}
	m.ExecOutOut = png
	s := newTestServer(m)

	_, out, err := s.handleScreenshot(context.Background(), nil, screenshotIn{})
	if err != nil {
		t.Fatalf("handleScreenshot() error: %v", err)
	}
	if out.Base64 != base64.StdEncoding.EncodeToString(png) {
		t.Errorf("Base64 = %q", out.Base64)
	}
	if out.Path != "" {
		t.Errorf("Path = %q, want empty", out.Path)
	}
	if out.MimeType != "image/png" {
		t.Errorf("MimeType = %q", out.MimeType)
	}
}

type testError struct{}

func (testError) Error() string { return "test error" }

var errTest = testError{}
