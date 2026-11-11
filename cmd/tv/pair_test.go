package main

import (
	"testing"

	"github.com/mutedfalcontv/tv/internal/discovery"
)

func TestConnectTargetFor_MatchesConnectService(t *testing.T) {
	services := []discovery.Service{
		{Kind: discovery.KindPairing, Address: "192.168.2.3:37123"},
		{Kind: discovery.KindConnect, Address: "192.168.2.3:5555"},
	}
	got, err := connectTargetFor("192.168.2.3:37123", services)
	if err != nil {
		t.Fatalf("connectTargetFor() unexpected error: %v", err)
	}
	if got != "192.168.2.3:5555" {
		t.Errorf("connectTargetFor() = %q, want connect service 192.168.2.3:5555", got)
	}
}

func TestConnectTargetFor_DoesNotReusePairingAddress(t *testing.T) {
	services := []discovery.Service{
		{Kind: discovery.KindPairing, Address: "192.168.2.3:37123"},
	}
	if got, err := connectTargetFor("192.168.2.3:37123", services); err == nil {
		t.Errorf("connectTargetFor() = %q, want error when only a pairing endpoint exists", got)
	}
}

func TestConnectTargetFor_IgnoresOtherHosts(t *testing.T) {
	services := []discovery.Service{
		{Kind: discovery.KindConnect, Address: "192.168.2.99:5555"},
	}
	if got, err := connectTargetFor("192.168.2.3:37123", services); err == nil {
		t.Errorf("connectTargetFor() = %q, want error for a different host", got)
	}
}

func TestHostOnly(t *testing.T) {
	tests := map[string]string{
		"192.168.2.3:37123": "192.168.2.3",
		"192.168.2.3":       "192.168.2.3",
		"[::1]:5555":        "::1",
	}
	for in, want := range tests {
		if got := hostOnly(in); got != want {
			t.Errorf("hostOnly(%q) = %q, want %q", in, got, want)
		}
	}
}
