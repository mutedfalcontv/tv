package localserve

import (
	"testing"
	"time"
)

func TestModeFor(t *testing.T) {
	tests := []struct {
		name    string
		forever bool
		timeout time.Duration
		want    Mode
	}{
		{"default is background", false, 0, ModeBackground},
		{"timeout is foreground", false, 30 * time.Minute, ModeForeground},
		{"forever wins", true, 0, ModeForever},
		{"forever wins over timeout", true, 30 * time.Minute, ModeForever},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ModeFor(tt.forever, tt.timeout); got != tt.want {
				t.Errorf("ModeFor(%v, %v) = %v, want %v", tt.forever, tt.timeout, got, tt.want)
			}
		})
	}
}

func TestModeDetached(t *testing.T) {
	if !ModeBackground.Detached() {
		t.Error("ModeBackground.Detached() = false, want true")
	}
	for _, m := range []Mode{ModeForeground, ModeForever} {
		if m.Detached() {
			t.Errorf("%v.Detached() = true, want false", m)
		}
	}
}

func TestModeTTL(t *testing.T) {
	tests := []struct {
		mode    Mode
		timeout time.Duration
		want    time.Duration
	}{
		{ModeBackground, 0, BackgroundTTL},
		{ModeBackground, 5 * time.Minute, BackgroundTTL},
		{ModeForeground, 30 * time.Minute, 30 * time.Minute},
		{ModeForever, 0, 0},
		{ModeForever, 30 * time.Minute, 0},
	}
	for _, tt := range tests {
		if got := tt.mode.TTL(tt.timeout); got != tt.want {
			t.Errorf("%v.TTL(%v) = %v, want %v", tt.mode, tt.timeout, got, tt.want)
		}
	}
}

func TestModeStringRoundTrip(t *testing.T) {
	for _, m := range []Mode{ModeBackground, ModeForeground, ModeForever} {
		if got := ParseMode(m.String()); got != m {
			t.Errorf("ParseMode(%q) = %v, want %v", m.String(), got, m)
		}
	}
}

func TestParseModeUnknown(t *testing.T) {
	for _, raw := range []string{"", "bogus", "Background"} {
		if got := ParseMode(raw); got.Valid() {
			t.Errorf("ParseMode(%q) = %v, want an invalid mode", raw, got)
		}
	}
}

func TestModeString(t *testing.T) {
	tests := map[Mode]string{
		ModeBackground: "background",
		ModeForeground: "foreground",
		ModeForever:    "forever",
		Mode(99):       "unknown",
	}
	for mode, want := range tests {
		if got := mode.String(); got != want {
			t.Errorf("Mode(%d).String() = %q, want %q", mode, got, want)
		}
	}
}
