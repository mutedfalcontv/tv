package main

import (
	"flag"
	"fmt"

	"github.com/mutedfalcontv/tv/internal/adb"
	"github.com/mutedfalcontv/tv/internal/device"
	"github.com/mutedfalcontv/tv/internal/output"
)

func stateCmd(args []string) {
	cfg := loadAdbConfig()
	r := &adb.RealRunner{}

	fs := flag.NewFlagSet("state", flag.ExitOnError)
	fs.Parse(args)

	st := device.GetState(cfg, r)

	output.Result(st, func() {
		fmt.Printf("TV:         %s\n", st.TVIP)
		fmt.Printf("Connected:  %s\n", yesNo(st.Connected))
		if !st.Connected {
			return
		}
		if st.ForegroundApp != "" {
			fmt.Printf("Foreground: %s (%s)\n", st.ForegroundApp, st.ForegroundActivity)
		} else {
			fmt.Println("Foreground: (unknown)")
		}
		fmt.Printf("Volume:     %d\n", st.Volume)
		fmt.Printf("Muted:      %s\n", yesNo(st.Muted))
		fmt.Printf("Screen:     %s (%s)\n", onOff(st.ScreenOn), st.Wakefulness)
	})
}

func yesNo(v bool) string {
	if v {
		return "yes"
	}
	return "no"
}

func onOff(v bool) string {
	if v {
		return "on"
	}
	return "off"
}
