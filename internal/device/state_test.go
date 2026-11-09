package device

import (
	"testing"

	"github.com/mutedfalcontv/tv/internal/adb"
)

func TestParseForeground(t *testing.T) {
	tests := []struct {
		name         string
		dumpsys      string
		wantApp      string
		wantActivity string
	}{
		{
			name: "topResumedActivity modern",
			dumpsys: `ACTIVITY MANAGER ACTIVITIES (dumpsys activity activities)
  ResumedActivity: ActivityRecord{3a1b u0 com.google.android.youtube.tv/com.google.android.apps.youtube.tv.activity.ShellActivity t5}
  topResumedActivity=ActivityRecord{3a1b u0 com.google.android.youtube.tv/com.google.android.apps.youtube.tv.activity.ShellActivity t5}`,
			wantApp:      "com.google.android.youtube.tv",
			wantActivity: "com.google.android.apps.youtube.tv.activity.ShellActivity",
		},
		{
			name:         "mResumedActivity relative class",
			dumpsys:      `  mResumedActivity: ActivityRecord{9f3 u0 com.netflix.ninja/.MainActivity t4}`,
			wantApp:      "com.netflix.ninja",
			wantActivity: "com.netflix.ninja.MainActivity",
		},
		{
			name:         "mFocusedApp fallback",
			dumpsys:      `  mFocusedApp=ActivityRecord{abc u0 org.videolan.vlc/.gui.video.VideoPlayerActivity t2}`,
			wantApp:      "org.videolan.vlc",
			wantActivity: "org.videolan.vlc.gui.video.VideoPlayerActivity",
		},
		{
			name:         "fully qualified activity",
			dumpsys:      `  mResumedActivity: ActivityRecord{abc u0 com.example.app/com.example.app.MainActivity t1}`,
			wantApp:      "com.example.app",
			wantActivity: "com.example.app.MainActivity",
		},
		{
			name:         "no activity markers",
			dumpsys:      "ACTIVITY MANAGER ACTIVITIES\n  no resumed activity here",
			wantApp:      "",
			wantActivity: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, activity := ParseForeground(tt.dumpsys)
			if app != tt.wantApp {
				t.Errorf("app = %q, want %q", app, tt.wantApp)
			}
			if activity != tt.wantActivity {
				t.Errorf("activity = %q, want %q", activity, tt.wantActivity)
			}
		})
	}
}

func TestParseAudio(t *testing.T) {
	tests := []struct {
		name       string
		dumpsys    string
		wantVolume int
		wantMuted  bool
	}{
		{
			name: "modern stream blocks",
			dumpsys: `Audio Manager state:
  Stream volumes (device: index)
  - STREAM_VOICE_CALL:
      Muted: false
      Min: 1
      Max: 6
      Current: 4 (speaker): 4, 4
  - STREAM_MUSIC:
      Muted: false
      Min: 0
      Max: 15
      Current: 10 (speaker): 10, 10
  - STREAM_ALARM:
      Muted: true
      Min: 0
      Max: 7
      Current: 6 (speaker): 6, 6`,
			wantVolume: 10,
			wantMuted:  false,
		},
		{
			name: "music muted",
			dumpsys: `  - STREAM_MUSIC:
      Muted: true
      Current: 0 (speaker): 0`,
			wantVolume: 0,
			wantMuted:  true,
		},
		{
			name: "legacy no dash",
			dumpsys: `STREAM_MUSIC:
   Muted: false
   Min: 0
   Max: 15
   Current: 7 (speaker): 7`,
			wantVolume: 7,
			wantMuted:  false,
		},
		{
			name:       "missing music stream",
			dumpsys:    "Ringer mode: NORMAL\n  Muted: true\n  Current: 99",
			wantVolume: 0,
			wantMuted:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			volume, muted := ParseAudio(tt.dumpsys)
			if volume != tt.wantVolume {
				t.Errorf("volume = %d, want %d", volume, tt.wantVolume)
			}
			if muted != tt.wantMuted {
				t.Errorf("muted = %v, want %v", muted, tt.wantMuted)
			}
		})
	}
}

func TestParsePower(t *testing.T) {
	tests := []struct {
		name            string
		dumpsys         string
		wantScreenOn    bool
		wantWakefulness string
	}{
		{
			name:            "awake display on",
			dumpsys:         "POWER MANAGER (dumpsys power)\n  mWakefulness=Awake\n  mWakefulnessChanging=false\n  Display Power: state=ON",
			wantScreenOn:    true,
			wantWakefulness: "Awake",
		},
		{
			name:            "asleep display off",
			dumpsys:         "  mWakefulness=Asleep\n  mWakefulnessChanging=false\n  Display Power: state=OFF",
			wantScreenOn:    false,
			wantWakefulness: "Asleep",
		},
		{
			name:            "dozing without display line",
			dumpsys:         "  mWakefulness=Dozing\n  mWakefulnessChanging=true",
			wantScreenOn:    false,
			wantWakefulness: "Dozing",
		},
		{
			name:            "awake without display line",
			dumpsys:         "  mWakefulness=Awake",
			wantScreenOn:    true,
			wantWakefulness: "Awake",
		},
		{
			name:            "no fields",
			dumpsys:         "POWER MANAGER (dumpsys power)",
			wantScreenOn:    false,
			wantWakefulness: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			screenOn, wakefulness := ParsePower(tt.dumpsys)
			if screenOn != tt.wantScreenOn {
				t.Errorf("screenOn = %v, want %v", screenOn, tt.wantScreenOn)
			}
			if wakefulness != tt.wantWakefulness {
				t.Errorf("wakefulness = %q, want %q", wakefulness, tt.wantWakefulness)
			}
		})
	}
}

func TestGetState_Connected(t *testing.T) {
	m := &adb.MockRunner{
		DevicesOut: "List of devices attached\n192.168.2.3:5555\tdevice\n",
		ShellResponses: map[string]string{
			"dumpsys activity activities": "  mResumedActivity: ActivityRecord{9f3 u0 com.netflix.ninja/.MainActivity t4}",
			"dumpsys audio":               "  - STREAM_MUSIC:\n      Muted: false\n      Current: 12 (speaker): 12, 12",
			"dumpsys power":               "  mWakefulness=Awake\n  Display Power: state=ON",
		},
	}
	cfg := &adb.Config{TVIP: "192.168.2.3:5555"}

	st := GetState(cfg, m)

	if !st.Connected {
		t.Error("Connected = false, want true")
	}
	if st.TVIP != cfg.TVIP {
		t.Errorf("TVIP = %q, want %q", st.TVIP, cfg.TVIP)
	}
	if st.ForegroundApp != "com.netflix.ninja" {
		t.Errorf("ForegroundApp = %q, want com.netflix.ninja", st.ForegroundApp)
	}
	if st.ForegroundActivity != "com.netflix.ninja.MainActivity" {
		t.Errorf("ForegroundActivity = %q, want com.netflix.ninja.MainActivity", st.ForegroundActivity)
	}
	if st.Volume != 12 {
		t.Errorf("Volume = %d, want 12", st.Volume)
	}
	if st.Muted {
		t.Error("Muted = true, want false")
	}
	if !st.ScreenOn {
		t.Error("ScreenOn = false, want true")
	}
	if st.Wakefulness != "Awake" {
		t.Errorf("Wakefulness = %q, want Awake", st.Wakefulness)
	}
}

func TestGetState_Disconnected(t *testing.T) {
	m := &adb.MockRunner{
		DevicesOut: "List of devices attached\n192.168.2.3:5555\toffline\n",
		ShellResponses: map[string]string{
			"dumpsys power": "  mWakefulness=Awake",
		},
	}
	cfg := &adb.Config{TVIP: "192.168.2.3:5555"}

	st := GetState(cfg, m)

	if st.Connected {
		t.Error("Connected = true, want false")
	}
	if st.Wakefulness != "" {
		t.Errorf("Wakefulness = %q, want empty (no dumpsys when disconnected)", st.Wakefulness)
	}
	if m.LastShellCmd != "" {
		t.Errorf("LastShellCmd = %q, want no shell invocation", m.LastShellCmd)
	}
}

func TestGetState_DevicesError(t *testing.T) {
	m := &adb.MockRunner{DevicesErr: errBoom{}}
	cfg := &adb.Config{TVIP: "192.168.2.3:5555"}

	st := GetState(cfg, m)

	if st.Connected {
		t.Error("Connected = true, want false on devices error")
	}
}

type errBoom struct{}

func (errBoom) Error() string { return "boom" }
