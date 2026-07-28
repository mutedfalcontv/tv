package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/mutedfalcontv/tv/internal/adb"
	"github.com/mutedfalcontv/tv/internal/config"
	"github.com/mutedfalcontv/tv/internal/logcat"
	"github.com/mutedfalcontv/tv/internal/player"
	"github.com/mutedfalcontv/tv/internal/remote"
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
	{"logs", "View and filter TV logs (logcat)", logsCmd},
	{"play", "Play a URL on TV", playCmd},
	{"player", "List/set default video player", playerCmd},
	{"remote", "TV remote control", remoteCmd},
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
	fmt.Fprint(os.Stderr, `tv — Android TV CLI tools

Usage:
  tv <command> [options] [args]

Commands:
`)
	for _, c := range commands {
		fmt.Fprintf(os.Stderr, "  %-15s %s\n", c.name, c.desc)
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

func remoteCmd(args []string) {
	cfg := loadAdbConfig()
	r := &adb.RealRunner{}

	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: tv remote <key>")
		fmt.Fprintln(os.Stderr, "Keys: up, down, left, right, ok, home, back, menu, power,")
		fmt.Fprintln(os.Stderr, "      input, info, subtitle, volup, voldown, mute,")
		fmt.Fprintln(os.Stderr, "      play, pause, stop, ff, rew, next, prev,")
		fmt.Fprintln(os.Stderr, "      chup, chdown, 0-9, type <text>")
		os.Exit(1)
	}

	cmd := args[0]
	if cmd == "type" {
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: tv remote type <text>")
			os.Exit(1)
		}
		if err := adb.EnsureConnected(cfg, r); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if err := remote.Type(cfg, r, args[1]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Typed: %s\n", args[1])
		return
	}
	if err := adb.EnsureConnected(cfg, r); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if err := remote.Press(cfg, r, cmd); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Sent key: %s\n", cmd)
}

func logsCmd(args []string) {
	cfg := loadAdbConfig()
	r := &adb.RealRunner{}
	var opts logcat.Options

	fs := flag.NewFlagSet("logs", flag.ExitOnError)
	packageFlag := fs.String("p", "", "Filter by package name")
	levelFlag := fs.String("l", "", "Log level: v/d/i/w/e/f")
	tagFlag := fs.String("t", "", "Filter by log tag")
	linesFlag := fs.Int("n", 0, "Number of recent lines")
	formatFlag := fs.String("v", "", "Log format: brief, threadtime, etc.")
	clearFlag := fs.Bool("c", false, "Clear log buffer")
	dumpFlag := fs.Bool("d", false, "Dump and exit (no follow)")
	fs.Parse(args)

	opts.Level = *levelFlag
	opts.Tag = *tagFlag
	opts.Lines = *linesFlag
	opts.Format = *formatFlag
	opts.Clear = *clearFlag
	opts.Dump = *dumpFlag

	if err := adb.EnsureConnected(cfg, r); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if *packageFlag != "" {
		pkg := *packageFlag
		if !strings.Contains(pkg, ".") {
			var err error
			pkg, err = player.Resolve(cfg, r, pkg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		}
		pid, err := logcat.ResolvePID(cfg, r, pkg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Filtering by %s (PID %d)\n", pkg, pid)
	}

	if opts.Clear {
		_, err := logcat.RunOnce(cfg, r, opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Log buffer cleared.")
		return
	}

	if opts.Dump || opts.Lines > 0 || opts.Level != "" || opts.Tag != "" {
		out, err := logcat.RunOnce(cfg, r, opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(out)
		return
	}

	fmt.Println("Streaming logs (Ctrl-C to stop)...")
	if err := logcat.RunStream(cfg, opts); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func loadAdbConfig() *adb.Config {
	c, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: config load failed: %v (using defaults)\n", err)
		return &adb.Config{TVIP: "192.168.2.3:5555"}
	}
	return &adb.Config{TVIP: c.TVIP, DefaultPlayer: c.DefaultPlayer}
}
