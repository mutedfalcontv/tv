package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/mutedfalcontv/tv/internal/adb"
	"github.com/mutedfalcontv/tv/internal/device"
)

func stateCmd(args []string) {
	cfg := loadAdbConfig()
	r := &adb.RealRunner{}

	fs := flag.NewFlagSet("state", flag.ExitOnError)
	jsonOut := fs.Bool("json", false, "Emit JSON state")
	ip := fs.String("ip", cfg.TVIP, "TV IP address")
	fs.Parse(args)
	cfg.TVIP = *ip

	st := device.GetState(cfg, r)

	if *jsonOut {
		enc := json.NewEncoder(os.Stdout)
		if err := enc.Encode(st); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

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
