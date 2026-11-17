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
	// KindUnknown is never produced from mDNS; it is the result of parsing an
	// unrecognized --kind filter.
	KindUnknown Kind = "unknown"
)

// ParseKind maps a --kind value to its typed kind. Anything unrecognized
// becomes KindUnknown so callers can reject it instead of filtering to an
// empty result.
func ParseKind(raw string) Kind {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "pairing":
		return KindPairing
	case "connect":
		return KindConnect
	case "legacy":
		return KindLegacy
	default:
		return KindUnknown
	}
}

// Valid reports whether k is a kind discovery can produce.
func (k Kind) Valid() bool {
	switch k {
	case KindPairing, KindConnect, KindLegacy:
		return true
	default:
		return false
	}
}

func (k Kind) String() string {
	if k == "" {
		return string(KindUnknown)
	}
	return string(k)
}

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
