package adb

import (
	"errors"
	"strings"
	"testing"
)

func TestPairDevice_Success(t *testing.T) {
	m := &MockRunner{PairOut: "Successfully paired to 192.168.2.3:37039 [guid=adb-ABC-XYZ]"}
	out, err := PairDevice(m, "192.168.2.3:37039", "123456")
	if err != nil {
		t.Fatalf("PairDevice() unexpected error: %v", err)
	}
	if !strings.Contains(out, "Successfully paired") {
		t.Errorf("PairDevice() output = %q, want success message", out)
	}
	if m.LastPairAddr != "192.168.2.3:37039" || m.LastPairCode != "123456" {
		t.Errorf("pair args = (%q, %q), want (%q, %q)", m.LastPairAddr, m.LastPairCode, "192.168.2.3:37039", "123456")
	}
}

func TestPairDevice_WrongCode(t *testing.T) {
	m := &MockRunner{PairOut: "Failed: Wrong password or connection was dropped."}
	_, err := PairDevice(m, "192.168.2.3:37039", "000000")
	if err == nil {
		t.Fatal("PairDevice() expected error for wrong code, got nil")
	}
	if !strings.Contains(err.Error(), "pairing failed") || !strings.Contains(err.Error(), "code") {
		t.Errorf("PairDevice() error = %q, want clear wrong-code message", err)
	}
}

func TestPairDevice_ErrorWithFailureOutput(t *testing.T) {
	m := &MockRunner{
		PairOut: "Failed: Wrong password or connection was dropped.",
		PairErr: errors.New("exit status 1"),
	}
	_, err := PairDevice(m, "192.168.2.3:37039", "000000")
	if err == nil {
		t.Fatal("PairDevice() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "pairing failed") {
		t.Errorf("PairDevice() error = %q, want pairing failed message", err)
	}
}

func TestPairDevice_ErrorWithoutOutput(t *testing.T) {
	m := &MockRunner{PairErr: assertAnError{}}
	_, err := PairDevice(m, "192.168.2.3:37039", "123456")
	if err == nil {
		t.Fatal("PairDevice() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "pairing failed") {
		t.Errorf("PairDevice() error = %q, want pairing failed message", err)
	}
}
