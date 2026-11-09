package main

import (
	"flag"
	"fmt"

	"github.com/mutedfalcontv/tv/internal/adb"
	"github.com/mutedfalcontv/tv/internal/device"
	"github.com/mutedfalcontv/tv/internal/output"
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
	fs.Parse(args)

	if err := adb.EnsureConnected(cfg, r); err != nil {
		output.FailWith(fmt.Sprintf("Error: %v", err), err)
	}

	data, err := device.CaptureScreenshot(cfg, r)
	if err != nil {
		output.FailWith(fmt.Sprintf("Error: %v", err), err)
	}

	path := *outPath
	if path == "" {
		path = device.DefaultScreenshotName()
	}
	if err := device.WriteScreenshot(path, data); err != nil {
		output.FailWith(fmt.Sprintf("Error writing screenshot: %v", err), err)
	}

	output.Result(screenshotResult{Path: path, Bytes: len(data)}, func() {
		fmt.Printf("Saved screenshot: %s (%d bytes)\n", path, len(data))
	})
}
