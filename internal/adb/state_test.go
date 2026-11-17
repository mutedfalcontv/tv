package adb

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestDeviceStateString(t *testing.T) {
	tests := []struct {
		state DeviceState
		want  string
	}{
		{DeviceStateUnknown, "unknown"},
		{DeviceStateAbsent, "absent"},
		{DeviceStateOffline, "offline"},
		{DeviceStateUnauthorized, "unauthorized"},
		{DeviceStateNoPermissions, "no permissions"},
		{DeviceStateDevice, "device"},
		{DeviceState(99), "unknown"},
	}
	for _, tt := range tests {
		if got := tt.state.String(); got != tt.want {
			t.Errorf("DeviceState(%d).String() = %q, want %q", tt.state, got, tt.want)
		}
	}
}

func TestParseDeviceState(t *testing.T) {
	tests := []struct {
		raw  string
		want DeviceState
	}{
		{"device", DeviceStateDevice},
		{" Device ", DeviceStateDevice},
		{"DEVICE", DeviceStateDevice},
		{"offline", DeviceStateOffline},
		{"unauthorized", DeviceStateUnauthorized},
		{"no permissions", DeviceStateNoPermissions},
		{"No Permissions", DeviceStateNoPermissions},
		{"", DeviceStateUnknown},
		{"recovery", DeviceStateUnknown},
		{"bootloader", DeviceStateUnknown},
	}
	for _, tt := range tests {
		if got := ParseDeviceState(tt.raw); got != tt.want {
			t.Errorf("ParseDeviceState(%q) = %v (%s), want %v", tt.raw, got, got, tt.want)
		}
	}
}

func TestDeviceStateReady(t *testing.T) {
	states := []DeviceState{
		DeviceStateUnknown,
		DeviceStateAbsent,
		DeviceStateOffline,
		DeviceStateUnauthorized,
		DeviceStateNoPermissions,
		DeviceStateDevice,
		DeviceState(99),
	}
	for _, state := range states {
		want := state == DeviceStateDevice
		if got := state.Ready(); got != want {
			t.Errorf("%v.Ready() = %v, want %v", state, got, want)
		}
	}
}

func TestDeviceStateFor(t *testing.T) {
	out := "List of devices attached\n" +
		"A\tdevice\n" +
		"B\toffline\n" +
		"C\tunauthorized\n" +
		"D\tno permissions (user in plugdev group); see [http://example]\n" +
		"E\trecovery\n"

	tests := []struct {
		name       string
		serial     string
		wantState  DeviceState
		wantRaw    string
		wantListed bool
	}{
		{"ready", "A", DeviceStateDevice, "device", true},
		{"offline", "B", DeviceStateOffline, "offline", true},
		{"unauthorized", "C", DeviceStateUnauthorized, "unauthorized", true},
		{"no permissions", "D", DeviceStateNoPermissions, "no permissions", true},
		{"unknown raw state", "E", DeviceStateUnknown, "recovery", true},
		{"absent", "ZZZ", DeviceStateAbsent, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state, raw, listed := DeviceStateFor(out, tt.serial)
			if state != tt.wantState || raw != tt.wantRaw || listed != tt.wantListed {
				t.Errorf("DeviceStateFor(%q) = (%v, %q, %v), want (%v, %q, %v)",
					tt.serial, state, raw, listed, tt.wantState, tt.wantRaw, tt.wantListed)
			}
		})
	}
}

// TestDeviceStateJSONUnchanged pins the `tv devices --json` shape: states stay
// the raw adb words even though decision points now use DeviceState.
func TestDeviceStateJSONUnchanged(t *testing.T) {
	out := "List of devices attached\n" +
		"A\tdevice\n" +
		"B\toffline\n" +
		"C\tunauthorized\n" +
		"D\tno permissions (user in plugdev group); see [http://example]\n"

	data, err := json.Marshal(ParseDevices(out))
	if err != nil {
		t.Fatalf("Marshal() error: %v", err)
	}
	var rows []struct {
		State string `json:"state"`
	}
	if err := json.Unmarshal(data, &rows); err != nil {
		t.Fatalf("Unmarshal() error: %v", err)
	}
	want := []string{"device", "offline", "unauthorized", "no permissions"}
	if len(rows) != len(want) {
		t.Fatalf("encoded %d rows, want %d", len(rows), len(want))
	}
	for i, w := range want {
		if rows[i].State != w {
			t.Errorf("rows[%d].state = %q, want %q", i, rows[i].State, w)
		}
	}
}

func TestEnsureConnected_TransitionsFromEveryState(t *testing.T) {
	tests := []struct {
		name  string
		first string
	}{
		{"absent", "List of devices attached\n"},
		{"offline", "List of devices attached\n192.168.2.3:5555\toffline\n"},
		{"unauthorized", "List of devices attached\n192.168.2.3:5555\tunauthorized\n"},
		{"no permissions", "List of devices attached\n192.168.2.3:5555\tno permissions (user in plugdev group)\n"},
		{"unknown raw state", "List of devices attached\n192.168.2.3:5555\trecovery\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MockRunner{
				DevicesResponses: []string{
					tt.first,
					"List of devices attached\n192.168.2.3:5555\tdevice\n",
				},
				ConnectOut: "connected to 192.168.2.3:5555",
			}
			if err := EnsureConnected(&Config{TVIP: "192.168.2.3:5555"}, m); err != nil {
				t.Fatalf("EnsureConnected() unexpected error: %v", err)
			}
			if !m.ConnectCalled {
				t.Error("Connect() not called from a non-ready state")
			}
		})
	}
}

func TestEnsureConnected_PostConnectStates(t *testing.T) {
	tests := []struct {
		name    string
		second  string
		wantErr string
	}{
		{"absent", "List of devices attached\n", "not listed"},
		{"offline", "List of devices attached\n192.168.2.3:5555\toffline\n", "offline"},
		{"unauthorized", "List of devices attached\n192.168.2.3:5555\tunauthorized\n", "unauthorized"},
		{"no permissions", "List of devices attached\n192.168.2.3:5555\tno permissions (user in plugdev group)\n", "no permissions"},
		{"unknown raw state", "List of devices attached\n192.168.2.3:5555\trecovery\n", "recovery"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MockRunner{
				DevicesResponses: []string{"List of devices attached\n", tt.second},
				ConnectOut:       "connected to 192.168.2.3:5555",
			}
			err := EnsureConnected(&Config{TVIP: "192.168.2.3:5555"}, m)
			if err == nil {
				t.Fatalf("EnsureConnected() = nil, want error for %s", tt.name)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want it to mention %q", err, tt.wantErr)
			}
		})
	}
}

// A connect target that differs from the serial must not be accepted: later
// commands address the serial, so a ready connect address would otherwise
// report success for an unreachable device.
func TestEnsureConnected_DoesNotAcceptDifferentConnectTarget(t *testing.T) {
	m := &MockRunner{
		DevicesResponses: []string{
			"List of devices attached\n",
			"List of devices attached\n10.0.0.2:5555\tdevice\n",
		},
		ConnectOut: "connected to 10.0.0.2:5555",
	}
	cfg := &Config{TVIP: "192.168.2.3:5555", Device: "ABC123", IP: "10.0.0.2:5555"}

	err := EnsureConnected(cfg, m)
	if err == nil {
		t.Fatal("EnsureConnected() = nil, want error when only the connect target is listed")
	}
	if !strings.Contains(err.Error(), "ABC123") {
		t.Errorf("error = %q, want it to name the serial ABC123", err)
	}
	if m.LastConnectAddr != "10.0.0.2:5555" {
		t.Errorf("Connect() address = %q, want the connect target", m.LastConnectAddr)
	}
}
