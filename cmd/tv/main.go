package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/mutedfalcontv/tv/pkg/adb"
	"github.com/mutedfalcontv/tv/pkg/config"
	"github.com/mutedfalcontv/tv/pkg/player"
)

type command struct {
	name string
	desc string
	run  func(args []string)
}

var commands = []command{
	{"adb", "ADB connection management", adbCmd},
	{"app", "App management (list, launch, kill)", appCmd},
	{"config", "Show configuration", configCmd},
	{"play", "Play a URL on TV", playCmd},
	{"player", "List/set default video player", playerCmd},
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
	sub := os.Args[1]
	for _, c := range commands {
		if c.name == sub {
			c.run(os.Args[2:])
			return
		}
	}
	fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", sub)
	printUsage()
	os.Exit(1)
}

func printUsage() {
	fmt.Println(`tv — Android TV CLI tools

Usage:
  tv <command> [options] [args]

Commands:`)
	for _, c := range commands {
		fmt.Printf("  %-15s %s\n", c.name, c.desc)
	}
}

func adbCmd(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: tv adb <connect|disconnect>")
		os.Exit(1)
	}
	r := &adb.RealRunner{}
	ip := adb.GetTVIP()
	switch args[0] {
	case "connect":
		fmt.Printf("Connecting to %s...\n", ip)
		result, err := r.Connect(ip)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(result)
	case "disconnect":
		fmt.Printf("Disconnecting from %s...\n", ip)
		result, err := r.Disconnect(ip)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(result)
	default:
		fmt.Fprintf(os.Stderr, "unknown adb subcommand: %s\n", args[0])
		os.Exit(1)
	}
}

func appCmd(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: tv app <list|launch|kill> [package]")
		os.Exit(1)
	}

	cfg := loadAdbConfig()
	r := &adb.RealRunner{}

	switch args[0] {
	case "list":
		if err := adb.EnsureConnected(cfg, r); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		out, err := r.Shell(cfg, "pm list packages")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(out)
	case "launch":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: tv app launch <package>")
			os.Exit(1)
		}
		if err := adb.EnsureConnected(cfg, r); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		_, err := r.ShellWithStderr(cfg, "monkey -p "+args[1]+" -c android.intent.category.LAUNCHER 1")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error launching %s: %v\n", args[1], err)
			os.Exit(1)
		}
		fmt.Printf("Launched %s\n", args[1])
	case "kill":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: tv app kill <package>")
			os.Exit(1)
		}
		if err := adb.EnsureConnected(cfg, r); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		_, err := r.ShellWithStderr(cfg, "am force-stop "+args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error killing %s: %v\n", args[1], err)
			os.Exit(1)
		}
		fmt.Printf("Killed %s\n", args[1])
	default:
		fmt.Fprintf(os.Stderr, "unknown app subcommand: %s\n", args[0])
		os.Exit(1)
	}
}

func configCmd(args []string) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Config file: ~/.config/tv/config.json\n")
	fmt.Printf("TV IP:        %s\n", cfg.TVIP)
	if cfg.DefaultPlayer != "" {
		fmt.Printf("Default player: %s\n", cfg.DefaultPlayer)
	} else {
		fmt.Println("Default player: (none)")
	}
}

func playerCmd(args []string) {
	cfg := loadAdbConfig()
	r := &adb.RealRunner{}

	fs := flag.NewFlagSet("player", flag.ExitOnError)
	ip := fs.String("ip", cfg.TVIP, "TV IP address")
	fs.Parse(args)
	cfg.TVIP = *ip

	remaining := fs.Args()
	if len(remaining) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: tv player <list|default> [options] [name]")
		os.Exit(1)
	}

	switch remaining[0] {
	case "list":
		players, err := player.Detect(cfg, r)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if len(players) == 0 {
			fmt.Println("No video players found on TV.")
			return
		}
		fmt.Println("Installed video players:")
		for _, p := range players {
			mark := " "
			if cfg.DefaultPlayer == p.PackageName {
				mark = "*"
			}
			fmt.Printf("  %s %-20s %s\n", mark, p.Name, p.PackageName)
		}
	case "default":
		if len(remaining) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: tv player default <name>")
			os.Exit(1)
		}
		name := remaining[1]
		pkg, err := player.Resolve(cfg, r, name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		conf, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		conf.DefaultPlayer = pkg
		if err := conf.Save(); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Default player set to: %s\n", pkg)
	default:
		fmt.Fprintf(os.Stderr, "unknown player subcommand: %s\n", remaining[0])
		os.Exit(1)
	}
}

func playCmd(args []string) {
	cfg := loadAdbConfig()
	r := &adb.RealRunner{}

	fs := flag.NewFlagSet("play", flag.ExitOnError)
	playerFlag := fs.String("player", cfg.DefaultPlayer, "Player name or package name")
	ip := fs.String("ip", cfg.TVIP, "TV IP address")
	fs.Parse(args)
	cfg.TVIP = *ip

	remaining := fs.Args()
	if len(remaining) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: tv play [options] <url>")
		fmt.Fprintln(os.Stderr, "  -player <name>   Player to use")
		fmt.Fprintln(os.Stderr, "  -ip <address>    TV IP address")
		os.Exit(1)
	}

	url := remaining[0]
	playerPkg := *playerFlag
	if playerPkg == "" {
		fmt.Fprintln(os.Stderr, "No default player set. Use 'tv player default <name>' or 'tv play -player <name> <url>'")
		os.Exit(1)
	}
	if !strings.Contains(playerPkg, ".") {
		pkg, err := player.Resolve(cfg, r, playerPkg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		playerPkg = pkg
	}
	if err := adb.EnsureConnected(cfg, r); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Playing on %s via %s...\n", cfg.TVIP, playerPkg)
	if err := player.PlayOnTV(cfg, r, playerPkg, url); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Done.")
}

func loadAdbConfig() *adb.Config {
	c, err := config.Load()
	if err != nil {
		return &adb.Config{TVIP: "192.168.2.3:5555"}
	}
	return &adb.Config{TVIP: c.TVIP, DefaultPlayer: c.DefaultPlayer}
}
