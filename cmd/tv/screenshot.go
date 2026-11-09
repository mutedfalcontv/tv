package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/mutedfalcontv/tv/internal/adb"
	"github.com/mutedfalcontv/tv/internal/device"
)

type screenshotResult struct {
	Path  string `json:"path"`
	Bytes int    `json:"bytes"`
}

func screenshotCmd(args []string) {
	cfg := loadAdbConfig()
	r := &adb.RealRunner{}

	fs := flag.NewFlagSet("screenshot", flag.ExitOnError)
	outPath := fs.String("o", "", "Output PNG path (default ./screencap-<unix-ts>.png)")
	jsonOut := fs.Bool("json", false, "Emit JSON {path, bytes}")
	ip := fs.String("ip", cfg.TVIP, "TV IP address")
	fs.Parse(args)
	cfg.TVIP = *ip

	if err := adb.EnsureConnected(cfg, r); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	data, err := device.CaptureScreenshot(cfg, r)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	path := *outPath
	if path == "" {
		path = device.DefaultScreenshotName()
	}
	if err := device.WriteScreenshot(path, data); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing screenshot: %v\n", err)
		os.Exit(1)
	}

	if *jsonOut {
		enc := json.NewEncoder(os.Stdout)
		if err := enc.Encode(screenshotResult{Path: path, Bytes: len(data)}); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}
	fmt.Printf("Saved screenshot: %s (%d bytes)\n", path, len(data))
}
