package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mutedfalcontv/tv/internal/adb"
	"github.com/mutedfalcontv/tv/internal/config"
	"github.com/mutedfalcontv/tv/internal/output"
)

func doctorCmd(args []string) {
	fs := newFlagSet("doctor")
	parseFlags(fs, args)

	var checks []output.DoctorCheck
	allOK := true
	add := func(c output.DoctorCheck) {
		checks = append(checks, c)
		if c.Status == "fail" {
			allOK = false
		}
	}

	if p, err := exec.LookPath("adb"); err != nil {
		add(output.DoctorCheck{Name: "adb", Status: "fail", Message: "ADB not found on PATH", Fix: "Install Android Platform Tools and ensure `adb` is on PATH."})
	} else {
		out, err := exec.Command(p, "version").CombinedOutput()
		msg := strings.TrimSpace(string(out))
		if i := strings.Index(msg, "\n"); i >= 0 {
			msg = strings.TrimSpace(msg[:i])
		}
		if err != nil || msg == "" {
			msg = "found at " + p
		}
		add(output.DoctorCheck{Name: "adb", Status: "ok", Message: msg})
	}

	cfg, cfgErr := config.Load()
	cfgPath, _ := config.Path()
	if cfgErr != nil {
		add(output.DoctorCheck{Name: "config", Status: "fail", Message: fmt.Sprintf("failed to load config: %v", cfgErr), Fix: fmt.Sprintf("Fix or remove %s.", cfgPath)})
		cfg = &config.Config{TVIP: "192.168.2.3:5555"}
	} else {
		msg := "loaded " + cfgPath
		if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
			msg = "using defaults (no config file yet at " + cfgPath + ")"
		}
		add(output.DoctorCheck{Name: "config", Status: "ok", Message: msg})
	}

	ac := &adb.Config{TVIP: cfg.TVIP, Device: cfg.Device, IP: cfg.IP}
	target := ac.ConnectTarget()
	if strings.TrimSpace(target) == "" {
		add(output.DoctorCheck{Name: "target", Status: "warn", Message: "no TV target configured", Fix: "Run `tv config set ip <tv-ip:5555>` or `tv discover` / `tv pair`."})
	} else {
		add(output.DoctorCheck{Name: "target", Status: "ok", Message: "target " + target})
	}

	r := &adb.RealRunner{}
	devOut, devErr := r.Devices()
	serial := ac.Serial()
	if devErr != nil {
		add(output.DoctorCheck{Name: "adb_server", Status: "fail", Message: fmt.Sprintf("cannot query devices: %v", devErr), Fix: "Ensure ADB works (`adb devices`)."})
	} else {
		devices := adb.ParseDevices(devOut)
		connected, unauth := false, false
		for _, d := range devices {
			if serial == "" || d.Serial == serial {
				if d.State == "unauthorized" {
					unauth = true
				}
				if d.State == "device" {
					connected = true
				}
			}
		}
		switch {
		case connected:
			add(output.DoctorCheck{Name: "device", Status: "ok", Message: "device connected (" + serial + ")"})
		case unauth:
			add(output.DoctorCheck{Name: "device", Status: "fail", Message: "device unauthorized", Fix: "Accept the ADB authorization prompt on the TV, then run `tv adb connect`."})
		case serial != "":
			add(output.DoctorCheck{Name: "device", Status: "fail", Message: fmt.Sprintf("target %s not connected", serial), Fix: "Run `tv adb connect`, or enable Wireless debugging and use `tv discover` / `tv pair`."})
		case len(devices) == 0:
			add(output.DoctorCheck{Name: "device", Status: "warn", Message: "no ADB devices attached", Fix: "Connect the TV (`tv adb connect`) or use `tv discover` / `tv pair`."})
		default:
			add(output.DoctorCheck{Name: "device", Status: "ok", Message: fmt.Sprintf("%d device(s) attached", len(devices))})
		}
	}

	if strings.TrimSpace(cfg.DefaultPlayer) != "" {
		add(output.DoctorCheck{Name: "default_player", Status: "ok", Message: cfg.DefaultPlayer})
	} else {
		add(output.DoctorCheck{Name: "default_player", Status: "warn", Message: "no default player set", Fix: "Run `tv player list` and `tv player default <name>` (optional)."})
	}

	if dir, err := config.Dir(); err == nil {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			add(output.DoctorCheck{Name: "config_dir", Status: "fail", Message: fmt.Sprintf("cannot create %s: %v", dir, err)})
		} else {
			probe := filepath.Join(dir, ".tv-doctor-write")
			if err := os.WriteFile(probe, []byte("ok"), 0o644); err != nil {
				add(output.DoctorCheck{Name: "config_dir", Status: "fail", Message: "not writable: " + dir})
			} else {
				_ = os.Remove(probe)
				add(output.DoctorCheck{Name: "config_dir", Status: "ok", Message: dir})
			}
		}
	}

	result := output.DoctorResult{OK: allOK, Checks: checks}
	output.Result(result, func() {
		for _, c := range checks {
			icon := map[string]string{"ok": "OK", "warn": "WARN", "fail": "FAIL"}[c.Status]
			if icon == "" {
				icon = strings.ToUpper(c.Status)
			}
			fmt.Printf("[%s] %-14s %s\n", icon, c.Name, c.Message)
			if c.Fix != "" && c.Status != "ok" {
				fmt.Printf("      fix: %s\n", c.Fix)
			}
		}
		if allOK {
			fmt.Println("\nAll checks passed.")
		} else {
			fmt.Println("\nSome checks failed. See fixes above.")
		}
	})
}
