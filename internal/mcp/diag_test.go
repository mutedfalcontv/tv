package mcp

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/mutedfalcontv/tv/internal/adb"
)

func TestHandleLogs_ScopedToPackage(t *testing.T) {
	m := connectedMock()
	m.ShellResponses = map[string]string{"pidof com.nuvio.tv": "4028\n"}
	m.ShellWithStderrOut = "E/Nuvio: rebuffer"
	s := newTestServer(m)

	_, out, err := s.handleLogs(context.Background(), nil, logsIn{Package: "com.nuvio.tv", Lines: 10})
	if err != nil {
		t.Fatalf("handleLogs() error: %v", err)
	}
	if out.Count != 1 || out.Lines[0] != "E/Nuvio: rebuffer" {
		t.Errorf("lines = %v, want the scoped log line", out.Lines)
	}
}

func TestHandleLogs_PackageNotRunning(t *testing.T) {
	m := connectedMock()
	s := newTestServer(m)

	_, _, err := s.handleLogs(context.Background(), nil, logsIn{Package: "com.missing.app"})
	if err == nil {
		t.Fatal("handleLogs() = nil error, want not-running error")
	}
	if !strings.Contains(err.Error(), "not running") {
		t.Errorf("error = %v, want it to say the package is not running", err)
	}
}

func TestHandleLogs_RejectsInvalidPackage(t *testing.T) {
	m := connectedMock()
	s := newTestServer(m)

	_, _, err := s.handleLogs(context.Background(), nil, logsIn{Package: "com.x; reboot"})
	if err == nil {
		t.Fatal("handleLogs() = nil error, want validation error")
	}
}

func TestHandleState_ReportsForeground(t *testing.T) {
	m := connectedMock()
	m.ShellResponses = map[string]string{
		"dumpsys activity activities": "mResumedActivity: ActivityRecord{1 u0 com.nuvio.tv/.MainActivity t1}",
		"dumpsys audio":               "STREAM_MUSIC: 2",
		"dumpsys power":               "mWakefulness=Awake\nDisplay Power: state=ON",
	}
	s := newTestServer(m)

	_, out, err := s.handleState(context.Background(), nil, stateIn{})
	if err != nil {
		t.Fatalf("handleState() error: %v", err)
	}
	if !out.Connected {
		t.Error("Connected = false, want true")
	}
	if out.ForegroundApp != "com.nuvio.tv" {
		t.Errorf("ForegroundApp = %q, want com.nuvio.tv", out.ForegroundApp)
	}
}

func TestHandleDiagnose_UsesForegroundPackage(t *testing.T) {
	m := connectedMock()
	m.ShellResponses = map[string]string{
		"dumpsys activity activities": "mResumedActivity: ActivityRecord{1 u0 com.nuvio.tv/.MainActivity t1}",
		"dumpsys audio":               "STREAM_MUSIC: 2",
		"dumpsys power":               "mWakefulness=Awake\nDisplay Power: state=ON",
		"pidof com.nuvio.tv":          "4028\n",
	}
	m.ShellWithStderrOut = "E/Nuvio: boom"
	s := newTestServer(m)

	_, out, err := s.handleDiagnose(context.Background(), nil, diagnoseIn{Lines: 5})
	if err != nil {
		t.Fatalf("handleDiagnose() error: %v", err)
	}
	if out.ForegroundApp != "com.nuvio.tv" {
		t.Errorf("ForegroundApp = %q, want com.nuvio.tv", out.ForegroundApp)
	}
	if out.ForegroundPID != 4028 {
		t.Errorf("ForegroundPID = %d, want 4028", out.ForegroundPID)
	}
	if len(out.ErrorLines) != 1 || out.ErrorLines[0] != "E/Nuvio: boom" {
		t.Errorf("ErrorLines = %v, want the scoped error line", out.ErrorLines)
	}
}

func TestHandleDiagnose_HintsWhenPackageNotRunning(t *testing.T) {
	m := connectedMock()
	m.ShellResponses = map[string]string{
		"dumpsys activity activities": "mResumedActivity: ActivityRecord{1 u0 com.other.app/.Main t1}",
		"dumpsys audio":               "STREAM_MUSIC: 2",
		"dumpsys power":               "mWakefulness=Awake\nDisplay Power: state=ON",
	}
	s := newTestServer(m)

	_, out, err := s.handleDiagnose(context.Background(), nil, diagnoseIn{Package: "com.nuvio.tv", Lines: 5})
	if err != nil {
		t.Fatalf("handleDiagnose() error: %v", err)
	}
	if out.ForegroundPID != 0 {
		t.Errorf("ForegroundPID = %d, want 0 when the package is not running", out.ForegroundPID)
	}
	joined := strings.Join(out.Hints, " | ")
	if !strings.Contains(joined, "not running") {
		t.Errorf("hints = %v, want a not-running hint", out.Hints)
	}
	if !strings.Contains(joined, "foreground app") {
		t.Errorf("hints = %v, want a foreground mismatch hint", out.Hints)
	}
}

func TestHandleDiagnose_Disconnected(t *testing.T) {
	m := &adb.MockRunner{
		DevicesOut: "List of devices attached\n",
		ConnectErr: errors.New("unreachable"),
	}
	s := newTestServer(m)

	_, out, err := s.handleDiagnose(context.Background(), nil, diagnoseIn{})
	if err != nil {
		t.Fatalf("handleDiagnose() error: %v", err)
	}
	if out.Connected {
		t.Error("Connected = true, want false")
	}
	if len(out.Hints) == 0 || !strings.Contains(strings.Join(out.Hints, " "), "not connected") {
		t.Errorf("hints = %v, want a not-connected hint", out.Hints)
	}
}

func TestHandleDiagnose_RejectsInvalidPackage(t *testing.T) {
	s := newTestServer(connectedMock())
	_, _, err := s.handleDiagnose(context.Background(), nil, diagnoseIn{Package: "bad/pkg"})
	if err == nil {
		t.Fatal("handleDiagnose() = nil error, want validation error")
	}
}

func TestNormalizeDiagnoseLines(t *testing.T) {
	tests := []struct {
		in            int
		wantLines     int
		wantTruncated bool
	}{
		{0, defaultDiagnoseLines, false},
		{-5, defaultDiagnoseLines, false},
		{10, 10, false},
		{maxDiagnoseLines + 50, maxDiagnoseLines, true},
	}
	for _, tt := range tests {
		lines, truncated := normalizeDiagnoseLines(tt.in)
		if lines != tt.wantLines || truncated != tt.wantTruncated {
			t.Errorf("normalizeDiagnoseLines(%d) = (%d, %v), want (%d, %v)",
				tt.in, lines, truncated, tt.wantLines, tt.wantTruncated)
		}
	}
}

func TestResolvePID_RejectsInvalidPackage(t *testing.T) {
	m := &adb.MockRunner{}
	if _, err := resolvePID(m, &adb.Config{TVIP: testIP}, "com.x; reboot"); err == nil {
		t.Fatal("resolvePID() = nil error, want validation error")
	}
	if m.LastShellCmd != "" {
		t.Errorf("shell ran despite invalid package: %q", m.LastShellCmd)
	}
}
