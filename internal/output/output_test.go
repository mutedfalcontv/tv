package output

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestEncodeTo_ErrorResponse(t *testing.T) {
	var buf bytes.Buffer
	if err := EncodeTo(&buf, ErrorResponse{Error: "boom"}); err != nil {
		t.Fatalf("EncodeTo() error: %v", err)
	}
	want := `{"error":"boom"}` + "\n"
	if got := buf.String(); got != want {
		t.Errorf("EncodeTo() = %q, want %q", got, want)
	}
}

func TestResultSchemas(t *testing.T) {
	tests := []struct {
		name string
		v    any
		want string
	}{
		{
			name: "adb",
			v:    AdbResult{Action: "connect", IP: "192.168.2.3:5555", Output: "connected"},
			want: `{"action":"connect","ip":"192.168.2.3:5555","output":"connected"}`,
		},
		{
			name: "app list empty",
			v:    AppListResult{Packages: []string{}},
			want: `{"packages":[]}`,
		},
		{
			name: "app action",
			v:    AppActionResult{Action: "launch", Package: "com.brouken.player"},
			want: `{"action":"launch","package":"com.brouken.player"}`,
		},
		{
			name: "config",
			v:    ConfigResult{ConfigFile: "/home/u/.config/tv/config.json", TVIP: "10.0.0.2:5555", DefaultPlayer: "org.videolan.vlc"},
			want: `{"config_file":"/home/u/.config/tv/config.json","tv_ip":"10.0.0.2:5555","default_player":"org.videolan.vlc"}`,
		},
		{
			name: "install",
			v:    InstallResult{APK: "app.apk", Installed: true},
			want: `{"apk":"app.apk","installed":true}`,
		},
		{
			name: "logs",
			v:    LogsResult{Package: "com.example", PID: 42, Logs: "line", Lines: []string{"line"}},
			want: `{"package":"com.example","pid":42,"logs":"line","lines":["line"]}`,
		},
		{
			name: "clear",
			v:    ClearResult{Cleared: true},
			want: `{"cleared":true}`,
		},
		{
			name: "play url",
			v:    PlayResult{URL: "https://example.com/v.mp4", Player: "org.videolan.vlc", TVIP: "10.0.0.2:5555", Played: true},
			want: `{"url":"https://example.com/v.mp4","player":"org.videolan.vlc","tv_ip":"10.0.0.2:5555","played":true}`,
		},
		{
			name: "play local",
			v: PlayResult{
				URL: "http://10.0.0.2:9000/v.mp4", Player: "org.videolan.vlc", TVIP: "10.0.0.2:5555",
				Played: true, LocalFile: "/tmp/v.mp4", ServeURL: "http://10.0.0.2:9000/v.mp4",
				ServeMode: "background", TTLSeconds: 3600,
			},
			want: `{"url":"http://10.0.0.2:9000/v.mp4","player":"org.videolan.vlc","tv_ip":"10.0.0.2:5555","played":true,"local_file":"/tmp/v.mp4","serve_url":"http://10.0.0.2:9000/v.mp4","serve_mode":"background","ttl_seconds":3600}`,
		},
		{
			name: "players",
			v: PlayersResult{
				Players:       []PlayerInfo{{Name: "VLC", Package: "org.videolan.vlc"}},
				DefaultPlayer: "org.videolan.vlc",
			},
			want: `{"players":[{"name":"VLC","package":"org.videolan.vlc"}],"default_player":"org.videolan.vlc"}`,
		},
		{
			name: "player default",
			v:    PlayerDefaultResult{DefaultPlayer: "org.videolan.vlc"},
			want: `{"default_player":"org.videolan.vlc"}`,
		},
		{
			name: "remote key",
			v:    RemoteResult{Key: "up", Keycode: "KEYCODE_DPAD_UP", Sent: true},
			want: `{"key":"up","keycode":"KEYCODE_DPAD_UP","sent":true}`,
		},
		{
			name: "remote type",
			v:    RemoteResult{Text: "hello", Sent: true},
			want: `{"text":"hello","sent":true}`,
		},
		{
			name: "serve",
			v:    ServeResult{URL: "http://localhost:8080", Host: "0.0.0.0", Port: 8080},
			want: `{"url":"http://localhost:8080","host":"0.0.0.0","port":8080}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.v)
			if err != nil {
				t.Fatalf("Marshal() error: %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("Marshal() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestJSONLineWriter(t *testing.T) {
	var buf bytes.Buffer
	w := NewJSONLineWriter(&buf)
	if _, err := w.Write([]byte("hello\nwor")); err != nil {
		t.Fatalf("Write() error: %v", err)
	}
	if _, err := w.Write([]byte("ld\n")); err != nil {
		t.Fatalf("Write() error: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("Flush() error: %v", err)
	}
	want := "{\"line\":\"hello\"}\n{\"line\":\"world\"}\n"
	if got := buf.String(); got != want {
		t.Errorf("JSON lines = %q, want %q", got, want)
	}
}

func TestJSONLineWriter_FlushPartial(t *testing.T) {
	var buf bytes.Buffer
	w := NewJSONLineWriter(&buf)
	if _, err := w.Write([]byte("tail")); err != nil {
		t.Fatalf("Write() error: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("Flush() error: %v", err)
	}
	want := "{\"line\":\"tail\"}\n"
	if got := buf.String(); got != want {
		t.Errorf("JSON lines = %q, want %q", got, want)
	}
}
