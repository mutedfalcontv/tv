package logcat

import "testing"

func TestModeString(t *testing.T) {
	tests := []struct {
		mode Mode
		want string
	}{
		{ModeStream, "stream"},
		{ModeDump, "dump"},
		{ModeClear, "clear"},
		{Mode(99), "unknown"},
	}
	for _, tt := range tests {
		if got := tt.mode.String(); got != tt.want {
			t.Errorf("Mode(%d).String() = %q, want %q", tt.mode, got, tt.want)
		}
	}
}

func TestOptionsMode(t *testing.T) {
	tests := []struct {
		name string
		opts Options
		want Mode
	}{
		{"default follows", Options{}, ModeStream},
		{"dump exits", Options{Dump: true}, ModeDump},
		{"clear wins", Options{Clear: true}, ModeClear},
		{"clear wins over dump", Options{Clear: true, Dump: true}, ModeClear},
		{"line limit implies dump", Options{Lines: 50}, ModeDump},
		{"level implies dump", Options{Level: "error"}, ModeDump},
		{"tag implies dump", Options{Tag: "VLC"}, ModeDump},
		{"format alone follows", Options{Format: "threadtime"}, ModeStream},
		{"pid alone follows", Options{Package: "1234"}, ModeStream},
		{"pid with tag dumps", Options{Package: "1234", Tag: "VLC"}, ModeDump},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.opts.Mode(); got != tt.want {
				t.Errorf("Options%+v.Mode() = %v, want %v", tt.opts, got, tt.want)
			}
		})
	}
}
