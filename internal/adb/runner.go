package adb

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/mutedfalcontv/tv/internal/config"
)

type Config struct {
	TVIP          string `json:"tv_ip"`
	Device        string `json:"device,omitempty"`
	IP            string `json:"ip,omitempty"`
	DefaultPlayer string `json:"default_player"`
}

// Serial returns the adb serial used for `adb -s`. A configured device wins,
// then an explicit IP, and finally the legacy TVIP default.
func (c *Config) Serial() string {
	if c.Device != "" {
		return c.Device
	}
	if c.IP != "" {
		return c.IP
	}
	return c.TVIP
}

// ConnectTarget returns the address used when (re)connecting to a device.
func (c *Config) ConnectTarget() string {
	if c.Device != "" && strings.Contains(c.Device, ":") {
		return c.Device
	}
	if c.IP != "" {
		return c.IP
	}
	return c.TVIP
}

// Device is one row of `adb devices -l`.
type Device struct {
	Serial    string `json:"serial"`
	State     string `json:"state"`
	Model     string `json:"model"`
	Product   string `json:"product"`
	Transport string `json:"transport"`
	Device    string `json:"device"`
}

// MDNSService is one row of `adb mdns services`.
type MDNSService struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Address string `json:"address"`
}

type Runner interface {
	Devices() (string, error)
	MDNSServices() (string, error)
	Connect(ip string) (string, error)
	Disconnect(ip string) (string, error)
	Install(cfg *Config, apkPath string) (string, error)
	Pair(addr, code string) (string, error)
	Shell(cfg *Config, cmd string) (string, error)
	ShellWithStderr(cfg *Config, cmd string) (string, error)
	ExecOut(cfg *Config, args ...string) ([]byte, error)
}

type RealRunner struct {
	// Timeout, when non-zero, bounds each adb invocation.
	Timeout time.Duration
}

func (r *RealRunner) binary() (string, error) {
	path, err := exec.LookPath("adb")
	if err != nil {
		return "", fmt.Errorf("ADB not found on PATH")
	}
	return path, nil
}

// command returns an exec.Cmd bound to r.Timeout, plus a cancel function the
// caller must defer. A zero Timeout means no deadline.
func (r *RealRunner) command(name string, args ...string) (*exec.Cmd, context.CancelFunc) {
	if r.Timeout > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), r.Timeout)
		return exec.CommandContext(ctx, name, args...), cancel
	}
	return exec.Command(name, args...), func() {}
}

func (r *RealRunner) Devices() (string, error) {
	adb, err := r.binary()
	if err != nil {
		return "", err
	}
	c, cancel := r.command(adb, "devices", "-l")
	defer cancel()
	out, err := c.Output()
	if err != nil {
		return "", fmt.Errorf("adb devices: %w", err)
	}
	return string(out), nil
}

func (r *RealRunner) MDNSServices() (string, error) {
	adb, err := r.binary()
	if err != nil {
		return "", err
	}
	c, cancel := r.command(adb, "mdns", "services")
	defer cancel()
	out, err := c.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("adb mdns services: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return string(out), nil
}

func (r *RealRunner) Connect(ip string) (string, error) {
	adb, err := r.binary()
	if err != nil {
		return "", err
	}
	c, cancel := r.command(adb, "connect", ip)
	defer cancel()
	out, err := c.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func (r *RealRunner) Disconnect(ip string) (string, error) {
	adb, err := r.binary()
	if err != nil {
		return "", err
	}
	c, cancel := r.command(adb, "disconnect", ip)
	defer cancel()
	out, err := c.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func (r *RealRunner) Install(cfg *Config, apkPath string) (string, error) {
	adb, err := r.binary()
	if err != nil {
		return "", err
	}
	c, cancel := r.command(adb, "-s", cfg.Serial(), "install", "-r", apkPath)
	defer cancel()
	out, err := c.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func (r *RealRunner) Pair(addr, code string) (string, error) {
	adb, err := r.binary()
	if err != nil {
		return "", err
	}
	c, cancel := r.command(adb, "pair", addr, code)
	defer cancel()
	out, err := c.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func (r *RealRunner) Shell(cfg *Config, cmd string) (string, error) {
	adb, err := r.binary()
	if err != nil {
		return "", err
	}
	c, cancel := r.command(adb, "-s", cfg.Serial(), "shell", cmd)
	defer cancel()
	out, err := c.Output()
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
	c, cancel := r.command(adb, "-s", cfg.Serial(), "shell", cmd)
	defer cancel()
	out, err := c.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("ADB error: %w\n%s", err, string(out))
	}
	return strings.TrimSpace(string(out)), nil
}

// ExecOut runs an adb command with exec-out, which bypasses the device PTY and
// returns raw stdout bytes. This is required for binary payloads such as PNG
// screencaps: piping through "adb shell" corrupts them with CRLF translation on
// Windows. The command is executed directly (no local shell) and no argument
// parsing is done, so args are passed to adb verbatim.
func (r *RealRunner) ExecOut(cfg *Config, args ...string) ([]byte, error) {
	adb, err := r.binary()
	if err != nil {
		return nil, err
	}
	full := append([]string{"-s", cfg.Serial(), "exec-out"}, args...)
	c, cancel := r.command(adb, full...)
	defer cancel()
	out, err := c.Output()
	if err != nil {
		return nil, fmt.Errorf("ADB exec-out error: %w", err)
	}
	return out, nil
}

func GetTVIP() string {
	if v := os.Getenv("TV_IP"); v != "" {
		return v
	}
	cfg, err := config.Load()
	if err != nil {
		return "192.168.2.3:5555"
	}
	return cfg.TVIP
}

// ParseDevices parses `adb devices -l` output into stable rows.
func ParseDevices(out string) []Device {
	var devices []Device
	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "*") || strings.HasPrefix(line, "List of devices") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		d := Device{Serial: fields[0], State: fields[1]}
		if d.State == "no" && len(fields) >= 3 && fields[2] == "permissions" {
			d.State = "no permissions"
		}
		for _, field := range fields[2:] {
			key, val, ok := strings.Cut(field, ":")
			if !ok {
				continue
			}
			switch key {
			case "model":
				d.Model = val
			case "product":
				d.Product = val
			case "device":
				d.Device = val
			case "transport_id":
				d.Transport = val
			}
		}
		devices = append(devices, d)
	}
	return devices
}

// ParseMDNSServices parses `adb mdns services` output into stable rows.
func ParseMDNSServices(out string) []MDNSService {
	var services []MDNSService
	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "List of discovered") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		services = append(services, MDNSService{
			Name:    fields[0],
			Type:    strings.TrimSuffix(fields[1], "."),
			Address: fields[2],
		})
	}
	return services
}

// PairDevice pairs with a device and turns wrong-code output into a clear error.
func PairDevice(r Runner, addr, code string) (string, error) {
	out, err := r.Pair(addr, code)
	if err != nil || pairingFailed(out) {
		if err == nil {
			err = fmt.Errorf("adb pair reported a failure")
		}
		return out, fmt.Errorf("pairing failed: %w\n%s\nCheck that the 6-digit code matches the one shown on the TV and that the pairing address/port are correct", err, strings.TrimSpace(out))
	}
	return out, nil
}

func pairingFailed(out string) bool {
	low := strings.ToLower(out)
	for _, marker := range []string{"failed", "wrong", "incorrect", "unable to pair", "cannot pair", "authentication"} {
		if strings.Contains(low, marker) {
			return true
		}
	}
	return false
}

type assertAnError struct{}

func (assertAnError) Error() string { return "expected error" }

func EnsureConnected(cfg *Config, r Runner) error {
	out, err := r.Devices()
	if err != nil {
		return err
	}
	serial := cfg.Serial()
	for _, d := range ParseDevices(out) {
		if d.Serial == serial && d.State == "device" {
			return nil
		}
	}
	out, err = r.Connect(cfg.ConnectTarget())
	if err != nil {
		return err
	}
	if connectFailed(out) {
		return fmt.Errorf("adb connect failed: %s", strings.TrimSpace(out))
	}
	return nil
}

// connectFailed reports whether adb connect output indicates a failure. adb
// often exits 0 even when it cannot connect, so callers must check the text.
func connectFailed(out string) bool {
	low := strings.ToLower(out)
	for _, marker := range []string{"failed", "cannot", "unable", "refused", "error"} {
		if strings.Contains(low, marker) {
			return true
		}
	}
	return false
}

// ValidatePackage reports whether pkg is a safe Android package name.
func ValidatePackage(pkg string) error {
	if pkg == "" {
		return fmt.Errorf("package must not be empty")
	}
	for _, r := range pkg {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '.', r == '_':
		default:
			return fmt.Errorf("invalid package name %q", pkg)
		}
	}
	return nil
}

type MockRunner struct {
	DevicesOut      string
	DevicesErr      error
	MDNSServicesOut string
	MDNSServicesErr error
	ConnectOut      string
	ConnectErr      error
	DisconnectOut   string
	DisconnectErr   error
	InstallOut      string
	InstallErr      error
	PairOut         string
	PairErr         error
	ShellOut        string
	ShellErr        error
	// ShellResponses lets a test return a different result per shell command.
	// Commands not present fall back to ShellOut/ShellErr.
	ShellResponses     map[string]string
	ShellWithStderrOut string
	ShellWithStderrErr error
	ExecOutOut         []byte
	ExecOutErr         error
	ConnectCalled      bool
	DisconnectCalled   bool
	InstallCalled      bool
	PairCalled         bool
	LastInstallPath    string
	LastConnectAddr    string
	LastPairAddr       string
	LastPairCode       string
	LastIntent         string
	LastShellCmd       string
	LastExecOutArgs    []string
}

func (m *MockRunner) Devices() (string, error) { return m.DevicesOut, m.DevicesErr }

func (m *MockRunner) MDNSServices() (string, error) {
	return m.MDNSServicesOut, m.MDNSServicesErr
}

func (m *MockRunner) Connect(ip string) (string, error) {
	m.ConnectCalled = true
	m.LastConnectAddr = ip
	return m.ConnectOut, m.ConnectErr
}

func (m *MockRunner) Disconnect(ip string) (string, error) {
	m.DisconnectCalled = true
	return m.DisconnectOut, m.DisconnectErr
}

func (m *MockRunner) Install(cfg *Config, apkPath string) (string, error) {
	m.InstallCalled = true
	m.LastInstallPath = apkPath
	return m.InstallOut, m.InstallErr
}

func (m *MockRunner) Pair(addr, code string) (string, error) {
	m.PairCalled = true
	m.LastPairAddr = addr
	m.LastPairCode = code
	return m.PairOut, m.PairErr
}

func (m *MockRunner) Shell(cfg *Config, cmd string) (string, error) {
	m.LastShellCmd = cmd
	if m.ShellResponses != nil {
		if out, ok := m.ShellResponses[cmd]; ok {
			return out, nil
		}
	}
	return m.ShellOut, m.ShellErr
}
func (m *MockRunner) ShellWithStderr(cfg *Config, cmd string) (string, error) {
	m.LastShellCmd = cmd
	return m.ShellWithStderrOut, m.ShellWithStderrErr
}
func (m *MockRunner) ExecOut(cfg *Config, args ...string) ([]byte, error) {
	m.LastExecOutArgs = args
	return m.ExecOutOut, m.ExecOutErr
}
