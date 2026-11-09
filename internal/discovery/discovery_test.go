package discovery

import (
	"errors"
	"strings"
	"testing"

	"github.com/mutedfalcontv/tv/internal/adb"
)

func TestADBDiscovery_Services(t *testing.T) {
	m := &adb.MockRunner{
		MDNSServicesOut: "List of discovered mdns services\n" +
			"adb-X\t_adb-tls-pairing._tcp\t192.168.2.3:37039\n" +
			"adb-X\t_adb-tls-connect._tcp\t192.168.2.3:37037\n" +
			"adb-X\t_adb._tcp\t192.168.2.3:5555\n",
	}
	d := &ADBDiscovery{Runner: m}

	services, err := d.Services()
	if err != nil {
		t.Fatalf("Services() unexpected error: %v", err)
	}
	if len(services) != 3 {
		t.Fatalf("got %d services, want 3", len(services))
	}
	wantKinds := []Kind{KindPairing, KindConnect, KindLegacy}
	for i, want := range wantKinds {
		if services[i].Kind != want {
			t.Errorf("services[%d].Kind = %q, want %q", i, services[i].Kind, want)
		}
	}
	if services[0].Address != "192.168.2.3:37039" {
		t.Errorf("services[0].Address = %q, want %q", services[0].Address, "192.168.2.3:37039")
	}
}

func TestADBDiscovery_Empty(t *testing.T) {
	m := &adb.MockRunner{MDNSServicesOut: "List of discovered mdns services\n"}
	services, err := (&ADBDiscovery{Runner: m}).Services()
	if err != nil {
		t.Fatalf("Services() unexpected error: %v", err)
	}
	if len(services) != 0 {
		t.Errorf("got %d services, want 0", len(services))
	}
}

func TestADBDiscovery_ErrorDocumentsFallback(t *testing.T) {
	m := &adb.MockRunner{MDNSServicesErr: errors.New("unknown command mdns")}
	_, err := (&ADBDiscovery{Runner: m}).Services()
	if err == nil {
		t.Fatal("Services() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "platform-tools") || !strings.Contains(err.Error(), "_adb-tls-pairing") {
		t.Errorf("Services() error = %q, want fallback documentation", err)
	}
}

func TestKindForType(t *testing.T) {
	tests := []struct {
		serviceType string
		want        Kind
	}{
		{"_adb-tls-pairing._tcp", KindPairing},
		{"_adb-tls-connect._tcp", KindConnect},
		{"_adb._tcp", KindLegacy},
		{"_something-else._tcp", KindLegacy},
	}
	for _, tt := range tests {
		if got := kindForType(tt.serviceType); got != tt.want {
			t.Errorf("kindForType(%q) = %q, want %q", tt.serviceType, got, tt.want)
		}
	}
}
