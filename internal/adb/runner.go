package adb

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Config struct {
	TVIP          string
	DefaultPlayer string
}

type Runner interface {
	Devices() (string, error)
	Connect(ip string) (string, error)
	Disconnect(ip string) (string, error)
	Shell(cfg *Config, cmd string) (string, error)
	ShellWithStderr(cfg *Config, cmd string) (string, error)
}

type RealRunner struct{}

func (r *RealRunner) binary() (string, error) {
	path, err := exec.LookPath("adb")
	if err != nil {
		return "", fmt.Errorf("ADB not found on PATH")
	}
	return path, nil
}

func (r *RealRunner) Devices() (string, error) {
	adb, err := r.binary()
	if err != nil {
		return "", err
	}
	out, err := exec.Command(adb, "devices").Output()
	if err != nil {
		return "", fmt.Errorf("adb devices: %w", err)
	}
	return string(out), nil
}

func (r *RealRunner) Connect(ip string) (string, error) {
	adb, err := r.binary()
	if err != nil {
		return "", err
	}
	out, err := exec.Command(adb, "connect", ip).CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func (r *RealRunner) Disconnect(ip string) (string, error) {
	adb, err := r.binary()
	if err != nil {
		return "", err
	}
	out, err := exec.Command(adb, "disconnect", ip).CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func (r *RealRunner) Shell(cfg *Config, cmd string) (string, error) {
	adb, err := r.binary()
	if err != nil {
		return "", err
	}
	out, err := exec.Command(adb, "-s", cfg.TVIP, "shell", cmd).Output()
	if err != nil {
		return "", fmt.Errorf("ADB error: %w\n%s", err, string(out))
	}
	return strings.TrimSpace(string(out)), nil
}

func (r *RealRunner) ShellWithStderr(cfg *Config, cmd string) (string, error) {
	adb, err := r.binary()
	if err != nil {
		return "", err
	}
	out, err := exec.Command(adb, "-s", cfg.TVIP, "shell", cmd).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("ADB error: %w\n%s", err, string(out))
	}
	return strings.TrimSpace(string(out)), nil
}

func GetTVIP() string {
	if v := os.Getenv("TV_IP"); v != "" {
		return v
	}
	cfg, err := loadConfig()
	if err != nil {
		return "192.168.2.3:5555"
	}
	return cfg.TVIP
}

func loadConfig() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(filepath.Join(home, ".config", "tv", "config.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{TVIP: "192.168.2.3:5555"}, nil
		}
		return nil, err
	}
	var cfg struct {
		TVIP          string `json:"tv_ip"`
		DefaultPlayer string `json:"default_player"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.TVIP == "" {
		return &Config{TVIP: "192.168.2.3:5555"}, nil
	}
	return &Config{TVIP: cfg.TVIP, DefaultPlayer: cfg.DefaultPlayer}, nil
}

type assertAnError struct{}

func (assertAnError) Error() string { return "expected error" }

func EnsureConnected(cfg *Config, r Runner) error {
	out, err := r.Devices()
	if err != nil {
		return err
	}
	if strings.Contains(out, cfg.TVIP) && !strings.Contains(out, "offline") {
		return nil
	}
	_, err = r.Connect(cfg.TVIP)
	return err
}

type MockRunner struct {
	DevicesOut    string
	DevicesErr    error
	ConnectOut    string
	ConnectErr    error
	DisconnectOut string
	DisconnectErr error
	ShellOut      string
	ShellErr      error
	ShellWithStderrOut string
	ShellWithStderrErr error
	ConnectCalled    bool
	DisconnectCalled bool
	LastIntent       string
	LastShellCmd     string
}

func (m *MockRunner) Devices() (string, error)                    { return m.DevicesOut, m.DevicesErr }
func (m *MockRunner) Connect(ip string) (string, error)           { m.ConnectCalled = true; return m.ConnectOut, m.ConnectErr }
func (m *MockRunner) Disconnect(ip string) (string, error)        { m.DisconnectCalled = true; return m.DisconnectOut, m.DisconnectErr }
func (m *MockRunner) Shell(cfg *Config, cmd string) (string, error)       { m.LastShellCmd = cmd; return m.ShellOut, m.ShellErr }
func (m *MockRunner) ShellWithStderr(cfg *Config, cmd string) (string, error) { return m.ShellWithStderrOut, m.ShellWithStderrErr }
