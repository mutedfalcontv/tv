package remote

import (
	"fmt"
	"sort"
	"strings"

	"github.com/mutedfalcontv/tv/internal/adb"
)

var keyMap = map[string]string{
	"up":       "KEYCODE_DPAD_UP",
	"down":     "KEYCODE_DPAD_DOWN",
	"left":     "KEYCODE_DPAD_LEFT",
	"right":    "KEYCODE_DPAD_RIGHT",
	"ok":       "KEYCODE_DPAD_CENTER",
	"home":     "KEYCODE_HOME",
	"back":     "KEYCODE_BACK",
	"menu":     "KEYCODE_MENU",
	"power":    "KEYCODE_POWER",
	"input":    "KEYCODE_TV_INPUT",
	"info":     "KEYCODE_INFO",
	"subtitle": "KEYCODE_CAPTIONS",
	"volup":    "KEYCODE_VOLUME_UP",
	"voldown":  "KEYCODE_VOLUME_DOWN",
	"mute":     "KEYCODE_VOLUME_MUTE",
	"play":     "KEYCODE_MEDIA_PLAY",
	"pause":    "KEYCODE_MEDIA_PAUSE",
	"stop":     "KEYCODE_MEDIA_STOP",
	"ff":       "KEYCODE_MEDIA_FAST_FORWARD",
	"rew":      "KEYCODE_MEDIA_REWIND",
	"next":     "KEYCODE_MEDIA_NEXT",
	"prev":     "KEYCODE_MEDIA_PREVIOUS",
	"chup":     "KEYCODE_CHANNEL_UP",
	"chdown":   "KEYCODE_CHANNEL_DOWN",
	"0":        "KEYCODE_0",
	"1":        "KEYCODE_1",
	"2":        "KEYCODE_2",
	"3":        "KEYCODE_3",
	"4":        "KEYCODE_4",
	"5":        "KEYCODE_5",
	"6":        "KEYCODE_6",
	"7":        "KEYCODE_7",
	"8":        "KEYCODE_8",
	"9":        "KEYCODE_9",
}

func Press(cfg *adb.Config, r adb.Runner, key string) error {
	keycode, ok := keyMap[key]
	if !ok {
		return fmt.Errorf("unknown key: %s. Available keys: %v", key, Keys())
	}
	_, err := r.Shell(cfg, "input keyevent "+keycode)
	if err != nil {
		return fmt.Errorf("failed to send keyevent: %w", err)
	}
	return nil
}

func Type(cfg *adb.Config, r adb.Runner, text string) error {
	escaped := "'" + strings.ReplaceAll(text, "'", "'\\''") + "'"
	_, err := r.Shell(cfg, "input text "+escaped)
	if err != nil {
		return fmt.Errorf("failed to input text: %w", err)
	}
	return nil
}

func Keys() []string {
	var names []string
	for k := range keyMap {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}
