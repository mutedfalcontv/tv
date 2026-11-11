package adb

import (
	"strings"
	"testing"
)

func TestEnsureConnected_AlreadyConnected(t *testing.T) {
	m := &MockRunner{
		DevicesOut: "List of devices attached\n192.168.2.3:5555   device\n",
	}
	cfg := &Config{TVIP: "192.168.2.3:5555"}
	err := EnsureConnected(cfg, m)
	if err != nil {
		t.Fatalf("EnsureConnected() unexpected error: %v", err)
	}
	if m.ConnectCalled {
		t.Error("Connect should not be called when device is already connected")
	}
}

func TestEnsureConnected_NotConnected(t *testing.T) {
	m := &MockRunner{
		DevicesResponses: []string{
			"List of devices attached\n",
			"List of devices attached\n192.168.2.3:5555   device\n",
		},
		ConnectOut: "connected to 192.168.2.3:5555",
	}
	cfg := &Config{TVIP: "192.168.2.3:5555"}
	err := EnsureConnected(cfg, m)
	if err != nil {
		t.Fatalf("EnsureConnected() unexpected error: %v", err)
	}
	if !m.ConnectCalled {
		t.Error("Connect should be called when device is not connected")
	}
}

func TestEnsureConnected_Offline(t *testing.T) {
	m := &MockRunner{
		DevicesResponses: []string{
			"List of devices attached\n192.168.2.3:5555   offline\n",
			"List of devices attached\n192.168.2.3:5555   device\n",
		},
		ConnectOut: "connected to 192.168.2.3:5555",
	}
	cfg := &Config{TVIP: "192.168.2.3:5555"}
	err := EnsureConnected(cfg, m)
	if err != nil {
		t.Fatalf("EnsureConnected() unexpected error: %v", err)
	}
	if !m.ConnectCalled {
		t.Error("Connect should be called when device is offline")
	}
}

func TestEnsureConnected_PostConnectStillOffline(t *testing.T) {
	m := &MockRunner{
		DevicesResponses: []string{
			"List of devices attached\n",
			"List of devices attached\n192.168.2.3:5555   offline\n",
		},
		ConnectOut: "connected to 192.168.2.3:5555",
	}
	err := EnsureConnected(&Config{TVIP: "192.168.2.3:5555"}, m)
	if err == nil {
		t.Fatal("EnsureConnected() expected error for offline device, got nil")
	}
	if !strings.Contains(err.Error(), "offline") {
		t.Errorf("error = %v, want offline explanation", err)
	}
}

func TestEnsureConnected_PostConnectUnauthorized(t *testing.T) {
	m := &MockRunner{
		DevicesResponses: []string{
			"List of devices attached\n",
			"List of devices attached\n192.168.2.3:5555   unauthorized\n",
		},
		ConnectOut: "connected to 192.168.2.3:5555",
	}
	err := EnsureConnected(&Config{TVIP: "192.168.2.3:5555"}, m)
	if err == nil {
		t.Fatal("EnsureConnected() expected error for unauthorized device, got nil")
	}
	if !strings.Contains(err.Error(), "unauthorized") {
		t.Errorf("error = %v, want unauthorized explanation", err)
	}
}

func TestEnsureConnected_PostConnectNotListed(t *testing.T) {
	m := &MockRunner{
		DevicesResponses: []string{
			"List of devices attached\n",
			"List of devices attached\n",
		},
		ConnectOut: "connected to 192.168.2.3:5555",
	}
	err := EnsureConnected(&Config{TVIP: "192.168.2.3:5555"}, m)
	if err == nil {
		t.Fatal("EnsureConnected() expected error when device never appears, got nil")
	}
}

func TestConnectFailed(t *testing.T) {
	tests := []struct {
		out  string
		want bool
	}{
		{"connected to 192.168.2.3:5555", false},
		{"already connected to 192.168.2.3:5555", false},
		{"failed to connect to 192.168.2.3:5555", true},
		{"cannot connect to 192.168.2.3:5555", true},
		{"error: protocol fault", true},
	}
	for _, tt := range tests {
		if got := ConnectFailed(tt.out); got != tt.want {
			t.Errorf("ConnectFailed(%q) = %v, want %v", tt.out, got, tt.want)
		}
	}
}

func TestMockRunnerLastShellCmd(t *testing.T) {
	m := &MockRunner{ShellOut: "output"}
	out, err := m.Shell(&Config{TVIP: "192.168.2.3:5555"}, "input keyevent KEYCODE_HOME")
	if err != nil {
		t.Fatalf("Shell() unexpected error: %v", err)
	}
	if out != "output" {
		t.Errorf("Shell() = %q, want %q", out, "output")
	}
	if m.LastShellCmd != "input keyevent KEYCODE_HOME" {
		t.Errorf("LastShellCmd = %q, want %q", m.LastShellCmd, "input keyevent KEYCODE_HOME")
	}
}

func TestEnsureConnected_ConnectFails(t *testing.T) {
	m := &MockRunner{
		DevicesOut: "List of devices attached\n",
		ConnectErr: assertAnError{},
	}
	cfg := &Config{TVIP: "192.168.2.3:5555"}
	err := EnsureConnected(cfg, m)
	if err == nil {
		t.Fatal("EnsureConnected() expected error, got nil")
	}
}

func TestMockRunnerShellWithStderrCapture(t *testing.T) {
	m := &MockRunner{ShellWithStderrOut: "output"}
	cfg := &Config{TVIP: "192.168.2.3:5555"}
	out, err := m.ShellWithStderr(cfg, "some command")
	if err != nil {
		t.Fatalf("ShellWithStderr() unexpected error: %v", err)
	}
	if out != "output" {
		t.Errorf("ShellWithStderr() = %q, want %q", out, "output")
	}
	if m.LastShellCmd != "some command" {
		t.Errorf("LastShellCmd = %q, want %q", m.LastShellCmd, "some command")
	}
}

func TestMockRunnerInstallCapture(t *testing.T) {
	m := &MockRunner{InstallOut: "Success"}
	out, err := m.Install(&Config{TVIP: "192.168.2.3:5555"}, "app.apk")
	if err != nil {
		t.Fatalf("Install() unexpected error: %v", err)
	}
	if out != "Success" {
		t.Errorf("Install() = %q, want %q", out, "Success")
	}
	if !m.InstallCalled {
		t.Error("InstallCalled = false, want true")
	}
	if m.LastInstallPath != "app.apk" {
		t.Errorf("LastInstallPath = %q, want %q", m.LastInstallPath, "app.apk")
	}
}
