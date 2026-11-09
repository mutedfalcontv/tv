package discovery

import (
	"fmt"
	"strings"

	"github.com/mutedfalcontv/tv/internal/adb"
)

type Kind string

const (
	KindPairing Kind = "pairing"
	KindConnect Kind = "connect"
	KindLegacy  Kind = "legacy"
)

type Service struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Kind    Kind   `json:"kind"`
	Type    string `json:"type"`
}

// Discovery is the seam used by callers so tests never touch the network.
type Discovery interface {
	Services() ([]Service, error)
}

// ADBDiscovery discovers wireless-debugging endpoints via `adb mdns services`.
type ADBDiscovery struct {
	Runner adb.Runner
}

func (d *ADBDiscovery) Services() ([]Service, error) {
	out, err := d.Runner.MDNSServices()
	if err != nil {
		return nil, fmt.Errorf("%w\n\nadb mdns services requires platform-tools 30.0.0+.\nEnable Wireless debugging on the TV (Settings > System > Developer options) and read the IP:port from that screen.\nAndroid 11+ advertises pairing via _adb-tls-pairing._tcp and connections via _adb-tls-connect._tcp.", err)
	}
	raw := adb.ParseMDNSServices(out)
	services := make([]Service, 0, len(raw))
	for _, s := range raw {
		services = append(services, Service{
			Name:    s.Name,
			Address: s.Address,
			Type:    s.Type,
			Kind:    kindForType(s.Type),
		})
	}
	return services, nil
}

func kindForType(serviceType string) Kind {
	switch {
	case strings.Contains(serviceType, "_adb-tls-pairing"):
		return KindPairing
	case strings.Contains(serviceType, "_adb-tls-connect"):
		return KindConnect
	default:
		return KindLegacy
	}
}
