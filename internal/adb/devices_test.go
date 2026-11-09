package adb

import (
	"reflect"
	"testing"
)

func TestParseDevices(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []Device
	}{
		{
			name:  "empty output",
			input: "List of devices attached\n",
			want:  nil,
		},
		{
			name:  "single usb device with metadata",
			input: "List of devices attached\nemulator-5554\tdevice product:sdk_gphone64_x86_64 model:sdk_gphone64_x86_64 device:emu64xa transport_id:1\n",
			want: []Device{{
				Serial:    "emulator-5554",
				State:     "device",
				Model:     "sdk_gphone64_x86_64",
				Product:   "sdk_gphone64_x86_64",
				Device:    "emu64xa",
				Transport: "1",
			}},
		},
		{
			name:  "network device",
			input: "List of devices attached\n192.168.2.3:5555\tdevice product:sabrina model:Chromecast_Google_TV device:sabrina transport_id:2\n",
			want: []Device{{
				Serial:    "192.168.2.3:5555",
				State:     "device",
				Model:     "Chromecast_Google_TV",
				Product:   "sabrina",
				Device:    "sabrina",
				Transport: "2",
			}},
		},
		{
			name:  "multiple devices with mixed states",
			input: "List of devices attached\nABC123\tdevice\n192.168.2.3:5555\toffline\n",
			want: []Device{
				{Serial: "ABC123", State: "device"},
				{Serial: "192.168.2.3:5555", State: "offline"},
			},
		},
		{
			name:  "unauthorized device",
			input: "List of devices attached\nABC123\tunauthorized\n",
			want:  []Device{{Serial: "ABC123", State: "unauthorized"}},
		},
		{
			name:  "no permissions device",
			input: "List of devices attached\n????????????\tno permissions (user in plugdev group); see [http://developer.android.com/tools/device.html]\n",
			want:  []Device{{Serial: "????????????", State: "no permissions"}},
		},
		{
			name:  "blank lines and daemon chatter",
			input: "* daemon not running; starting now at tcp:5037\n* daemon started successfully\nList of devices attached\n\nABC123 device\n\n",
			want:  []Device{{Serial: "ABC123", State: "device"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseDevices(tt.input); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseDevices() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestParseDevicesFromMockRunner(t *testing.T) {
	m := &MockRunner{
		DevicesOut: "List of devices attached\n192.168.2.3:5555\tdevice product:sabrina model:Chromecast_Google_TV device:sabrina transport_id:3\n",
	}
	out, err := m.Devices()
	if err != nil {
		t.Fatalf("Devices() unexpected error: %v", err)
	}
	devices := ParseDevices(out)
	if len(devices) != 1 {
		t.Fatalf("got %d devices, want 1", len(devices))
	}
	if devices[0].Serial != "192.168.2.3:5555" || devices[0].Model != "Chromecast_Google_TV" || devices[0].Transport != "3" {
		t.Errorf("unexpected device: %#v", devices[0])
	}
}

func TestParseMDNSServices(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []MDNSService
	}{
		{
			name:  "empty output",
			input: "List of discovered mdns services\n",
			want:  nil,
		},
		{
			name:  "pairing and connect services",
			input: "List of discovered mdns services\nadb-ABC123-XYZ\t_adb-tls-pairing._tcp\t192.168.2.3:37039\nadb-ABC123-XYZ\t_adb-tls-connect._tcp\t192.168.2.3:37037\n",
			want: []MDNSService{
				{Name: "adb-ABC123-XYZ", Type: "_adb-tls-pairing._tcp", Address: "192.168.2.3:37039"},
				{Name: "adb-ABC123-XYZ", Type: "_adb-tls-connect._tcp", Address: "192.168.2.3:37037"},
			},
		},
		{
			name:  "legacy adb tcp service and trailing dot",
			input: "List of discovered mdns services\nadb-LEGACY\t_adb._tcp.\t192.168.2.3:5555\n",
			want:  []MDNSService{{Name: "adb-LEGACY", Type: "_adb._tcp", Address: "192.168.2.3:5555"}},
		},
		{
			name:  "space separated and blank lines",
			input: "\nList of discovered mdns services\n\nadb-X   _adb-tls-pairing._tcp   10.0.0.5:40000\n\n",
			want:  []MDNSService{{Name: "adb-X", Type: "_adb-tls-pairing._tcp", Address: "10.0.0.5:40000"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseMDNSServices(tt.input); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseMDNSServices() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestParseMDNSServicesFromMockRunner(t *testing.T) {
	m := &MockRunner{
		MDNSServicesOut: "List of discovered mdns services\nadb-X\t_adb-tls-connect._tcp\t192.168.2.3:37037\n",
	}
	out, err := m.MDNSServices()
	if err != nil {
		t.Fatalf("MDNSServices() unexpected error: %v", err)
	}
	services := ParseMDNSServices(out)
	if len(services) != 1 {
		t.Fatalf("got %d services, want 1", len(services))
	}
	if services[0].Address != "192.168.2.3:37037" {
		t.Errorf("unexpected service: %#v", services[0])
	}
}

func TestConfigSerial(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want string
	}{
		{"device wins", Config{TVIP: "192.168.2.3:5555", Device: "ABC123", IP: "10.0.0.2:5555"}, "ABC123"},
		{"ip when no device", Config{TVIP: "192.168.2.3:5555", IP: "10.0.0.2:5555"}, "10.0.0.2:5555"},
		{"tvip fallback", Config{TVIP: "192.168.2.3:5555"}, "192.168.2.3:5555"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.Serial(); got != tt.want {
				t.Errorf("Serial() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestConfigConnectTarget(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want string
	}{
		{"network device serial", Config{TVIP: "192.168.2.3:5555", Device: "10.0.0.2:5555"}, "10.0.0.2:5555"},
		{"ip wins over tvip", Config{TVIP: "192.168.2.3:5555", IP: "10.0.0.2:5555"}, "10.0.0.2:5555"},
		{"usb serial falls back to tvip", Config{TVIP: "192.168.2.3:5555", Device: "ABC123"}, "192.168.2.3:5555"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.ConnectTarget(); got != tt.want {
				t.Errorf("ConnectTarget() = %q, want %q", got, tt.want)
			}
		})
	}
}
