package adb

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/mutedfalcontv/tv/internal/config"
)

type Config struct {
	TVIP          string
	Device        string
	IP            string
	DefaultPlayer string
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
	Pair(addr, code string) (string, error)
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
	out, err := exec.Command(adb, "devices", "-l").Output()
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
	out, err := exec.Command(adb, "mdns", "services").CombinedOutput()
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

func (r *RealRunner) Pair(addr, code string) (string, error) {
	adb, err := r.binary()
	if err != nil {
		return "", err
	}
	out, err := exec.Command(adb, "pair", addr, code).CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func (r *RealRunner) Shell(cfg *Config, cmd string) (string, error) {
	adb, err := r.binary()
	if err != nil {
		return "", err
	}
	out, err := exec.Command(adb, "-s", cfg.Serial(), "shell", cmd).Output()
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
	out, err := exec.Command(adb, "-s", cfg.Serial(), "shell", cmd).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("ADB error: %w\n%s", err, string(out))
	}
	return strings.TrimSpace(string(out)), nil
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
	_, err = r.Connect(cfg.ConnectTarget())
	return err
}

type MockRunner struct {
	DevicesOut         string
	DevicesErr         error
	MDNSServicesOut    string
	MDNSServicesErr    error
	ConnectOut         string
	ConnectErr         error
	DisconnectOut      string
	DisconnectErr      error
	PairOut            string
	PairErr            error
	ShellOut           string
	ShellErr           error
	ShellWithStderrOut string
	ShellWithStderrErr error
	ConnectCalled      bool
	DisconnectCalled   bool
	PairCalled         bool
	LastConnectAddr    string
	LastPairAddr       string
	LastPairCode       string
	LastIntent         string
	LastShellCmd       string
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

func (m *MockRunner) Pair(addr, code string) (string, error) {
	m.PairCalled = true
	m.LastPairAddr = addr
	m.LastPairCode = code
	return m.PairOut, m.PairErr
}

func (m *MockRunner) Shell(cfg *Config, cmd string) (string, error) {
	m.LastShellCmd = cmd
	return m.ShellOut, m.ShellErr
}

func (m *MockRunner) ShellWithStderr(cfg *Config, cmd string) (string, error) {
	m.LastShellCmd = cmd
	return m.ShellWithStderrOut, m.ShellWithStderrErr
}
