package adb

import (
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
		DevicesOut: "List of devices attached\n",
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
		DevicesOut: "List of devices attached\n192.168.2.3:5555   offline\n",
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
