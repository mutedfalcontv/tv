package device

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/mutedfalcontv/tv/internal/adb"
)

// State is the live device state derived from dumpsys. It is the stable schema
// emitted by "tv state --json" and GET /api/state.
type State struct {
	Connected          bool   `json:"connected"`
	TVIP               string `json:"tv_ip"`
	ForegroundApp      string `json:"foreground_app,omitempty"`
	ForegroundActivity string `json:"foreground_activity,omitempty"`
	Volume             int    `json:"volume"`
	Muted              bool   `json:"muted"`
	ScreenOn           bool   `json:"screen_on"`
	Wakefulness        string `json:"wakefulness,omitempty"`
}

// GetState probes the TV and returns its current state. Connectivity is derived
// from "adb devices"; the dumpsys-derived fields are filled in only when a
// device is connected. Individual dumpsys failures leave their fields at zero
// values rather than failing the whole call.
func GetState(cfg *adb.Config, r adb.Runner) State {
	st := State{TVIP: cfg.TVIP}
	devices, err := r.Devices()
	if err != nil {
		return st
	}
	st.Connected = adbConnected(devices, cfg.TVIP)
	if !st.Connected {
		return st
	}
	if out, err := r.Shell(cfg, "dumpsys activity activities"); err == nil {
		st.ForegroundApp, st.ForegroundActivity = ParseForeground(out)
	}
	if out, err := r.Shell(cfg, "dumpsys audio"); err == nil {
		st.Volume, st.Muted = ParseAudio(out)
	}
	if out, err := r.Shell(cfg, "dumpsys power"); err == nil {
		st.ScreenOn, st.Wakefulness = ParsePower(out)
	}
	return st
}

// adbConnected reports whether ip appears in "adb devices" output in the
// "device" state (as opposed to "offline" or "unauthorized").
func adbConnected(devices, ip string) bool {
	for _, line := range strings.Split(devices, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == ip {
			return fields[1] == "device"
		}
	}
	return false
}

var componentRe = regexp.MustCompile(`([A-Za-z0-9_][A-Za-z0-9_.]*)/([A-Za-z0-9_.$]+)`)

// ParseForeground extracts the foreground package and activity from
// "dumpsys activity activities". It looks at topResumedActivity, the global
// ResumedActivity, mResumedActivity, and finally mFocusedApp, in that order.
func ParseForeground(dumpsys string) (pkg, activity string) {
	lines := strings.Split(dumpsys, "\n")
	markers := []string{
		"topResumedActivity=",
		"topResumedActivity:",
		"ResumedActivity:",
		"mResumedActivity:",
		"mFocusedApp=",
	}
	for _, marker := range markers {
		for _, line := range lines {
			idx := strings.Index(line, marker)
			if idx < 0 {
				continue
			}
			if m := componentRe.FindStringSubmatch(line[idx+len(marker):]); m != nil {
				return resolveComponent(m[1], m[2])
			}
		}
	}
	return "", ""
}

func resolveComponent(pkg, activity string) (string, string) {
	if strings.HasPrefix(activity, ".") {
		activity = pkg + activity
	}
	return pkg, activity
}

// ParseAudio extracts the media stream volume and muted flag from
// "dumpsys audio". It scans the STREAM_MUSIC block, where "Current:" carries the
// volume index and "Muted:" carries the mute state.
func ParseAudio(dumpsys string) (volume int, muted bool) {
	inMusic := false
	for _, line := range strings.Split(dumpsys, "\n") {
		if name, ok := streamHeader(strings.TrimSpace(line)); ok {
			if inMusic && name != "STREAM_MUSIC" {
				break
			}
			inMusic = name == "STREAM_MUSIC"
			continue
		}
		if !inMusic {
			continue
		}
		if v, ok := parseAfterColonInt(line, "Current:"); ok {
			volume = v
		}
		if v, ok := parseAfterColonBool(line, "Muted:"); ok {
			muted = v
		}
	}
	return volume, muted
}

// streamHeader recognizes lines like "- STREAM_MUSIC:" or "STREAM_MUSIC:".
func streamHeader(trimmed string) (string, bool) {
	t := strings.TrimPrefix(trimmed, "- ")
	if !strings.HasPrefix(t, "STREAM_") || !strings.HasSuffix(t, ":") {
		return "", false
	}
	return strings.TrimSuffix(t, ":"), true
}

func parseAfterColonInt(line, key string) (int, bool) {
	idx := strings.Index(line, key)
	if idx < 0 {
		return 0, false
	}
	rest := strings.TrimSpace(line[idx+len(key):])
	fields := strings.Fields(rest)
	if len(fields) == 0 {
		return 0, false
	}
	v, err := strconv.Atoi(fields[0])
	if err != nil {
		return 0, false
	}
	return v, true
}

func parseAfterColonBool(line, key string) (bool, bool) {
	idx := strings.Index(line, key)
	if idx < 0 {
		return false, false
	}
	rest := strings.ToLower(strings.TrimSpace(line[idx+len(key):]))
	switch {
	case strings.HasPrefix(rest, "true"):
		return true, true
	case strings.HasPrefix(rest, "false"):
		return false, true
	}
	return false, false
}

// ParsePower extracts screen power and wakefulness from "dumpsys power".
// Wakefulness comes from mWakefulness (Awake/Asleep/Dozing); screen on/off comes
// from the "Display Power: state=ON|OFF" line when available, falling back to
// wakefulness == Awake.
func ParsePower(dumpsys string) (screenOn bool, wakefulness string) {
	displayState := ""
	for _, line := range strings.Split(dumpsys, "\n") {
		trimmed := strings.TrimSpace(line)
		if wakefulness == "" && strings.HasPrefix(trimmed, "mWakefulness=") {
			wakefulness = strings.TrimSpace(strings.TrimPrefix(trimmed, "mWakefulness="))
		}
		if displayState == "" && strings.Contains(line, "Display Power:") {
			displayState = valueAfter(line, "state=")
		}
	}
	switch strings.ToUpper(displayState) {
	case "ON":
		return true, wakefulness
	case "OFF":
		return false, wakefulness
	}
	return strings.EqualFold(wakefulness, "Awake"), wakefulness
}

func valueAfter(line, key string) string {
	idx := strings.Index(line, key)
	if idx < 0 {
		return ""
	}
	rest := strings.TrimSpace(line[idx+len(key):])
	fields := strings.Fields(rest)
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}
