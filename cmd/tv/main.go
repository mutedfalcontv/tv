package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/mutedfalcontv/tv/internal/adb"
	"github.com/mutedfalcontv/tv/internal/config"
	"github.com/mutedfalcontv/tv/internal/discovery"
	"github.com/mutedfalcontv/tv/internal/help"
	"github.com/mutedfalcontv/tv/internal/logcat"
	"github.com/mutedfalcontv/tv/internal/output"
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
	{"devices", "List connected ADB devices", devicesCmd},
	{"discover", "Discover TVs via mDNS (wireless debugging)", discoverCmd},
	{"install", "Install an APK on the TV", installCmd},
	{"logs", "View and filter TV logs (logcat)", logsCmd},
	{"mcp", "Start stdio MCP server for AI agents", mcpCmd},
	{"pair", "Pair with a TV over wireless ADB", pairCmd},
	{"play", "Play a URL or local file on TV", playCmd},
	{"player", "List/set default video player", playerCmd},
	{"remote", "TV remote control", remoteCmd},
	{"screenshot", "Capture the TV screen to a PNG", screenshotCmd},
	{"serve", "Start TV Remote UI web server", serveCmd},
	{"state", "Show live TV state from dumpsys", stateCmd},
}

// noInput is true when --no-input was passed or stdin is not a terminal.
var noInput bool

// globalDevice and globalIP come from the --device/--ip global flags.
var (
	globalDevice string
	globalIP     string
)

func main() {
	args, opts := parseGlobalFlags(os.Args[1:])
	output.JSON = opts.JSON
	noInput = opts.NoInput || !isTerminal(os.Stdin)

	if len(args) == 0 {
		if output.JSON {
			output.FailMsg("no command specified")
		}
		printUsage(os.Stderr)
		os.Exit(1)
	}

	sub := args[0]
	rest := args[1:]

	// Internal subcommand used by `tv play <local-file>` to keep serving in a
	// detached process after the parent command exits.
	if sub == "__serve-file" {
		serveFileChildCmd(rest)
		return
	}

	switch sub {
	case "help", "-h", "--help":
		if output.JSON {
			printHelpJSON()
			return
		}
		printUsage(os.Stdout)
		return
	}

	for _, c := range commands {
		if c.name == sub {
			c.run(rest)
			return
		}
	}
	if output.JSON {
		output.Failf("unknown command: %s", sub)
	}
	fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", sub)
	printUsage(os.Stderr)
	os.Exit(1)
}

type globalOptions struct {
	JSON    bool
	NoInput bool
	Device  string
	IP      string
}

// parseGlobalFlags extracts global flags from anywhere in the argument list so
// that both `tv --json app list` and `tv app list --json` work. It also
// consumes --device/--ip used to target a specific device.
func parseGlobalFlags(args []string) ([]string, globalOptions) {
	var opts globalOptions
	out := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--json" || a == "-json" || a == "--json=true" || a == "-json=true":
			opts.JSON = true
		case a == "--json=false" || a == "-json=false":
			opts.JSON = false
		case a == "--no-input" || a == "-no-input" || a == "--no-input=true" || a == "-no-input=true":
			opts.NoInput = true
		case a == "--no-input=false" || a == "-no-input=false":
			opts.NoInput = false
		case a == "--device" || a == "-device":
			if i+1 < len(args) {
				opts.Device = args[i+1]
				globalDevice = args[i+1]
				i++
				continue
			}
			out = append(out, a)
		case a == "--ip" || a == "-ip":
			if i+1 < len(args) {
				opts.IP = args[i+1]
				globalIP = args[i+1]
				i++
				continue
			}
			out = append(out, a)
		default:
			if v, ok := strings.CutPrefix(a, "--device="); ok {
				opts.Device = v
				globalDevice = v
				continue
			}
			if v, ok := strings.CutPrefix(a, "-device="); ok {
				opts.Device = v
				globalDevice = v
				continue
			}
			if v, ok := strings.CutPrefix(a, "--ip="); ok {
				opts.IP = v
				globalIP = v
				continue
			}
			if v, ok := strings.CutPrefix(a, "-ip="); ok {
				opts.IP = v
				globalIP = v
				continue
			}
			out = append(out, a)
		}
	}
	return out, opts
}

// isTerminal reports whether f is attached to a terminal. Anything else
// (pipes, files, CI environments) activates non-interactive mode.
func isTerminal(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func newFlagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	return fs
}

func parseFlags(fs *flag.FlagSet, args []string) {
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(0)
		}
		output.Failf("%v", err)
	}
}

func printHelpJSON() {
	data, err := help.JSON()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	os.Stdout.Write(data)
	fmt.Println()
}

func printUsage(w io.Writer) {
	fmt.Fprint(w, `tv — Android TV CLI tools

Usage:
  tv [global options] <command> [options] [args]

Global options:
  --json             Emit one JSON value on stdout; errors are JSON on stderr
  --no-input         Never prompt (automatic when stdin is not a TTY)
  --device <serial>  Target a specific device (adb -s <serial>)
  --ip <address>     Target a specific device IP address

Commands:
`)
	for _, c := range commands {
		fmt.Fprintf(w, "  %-15s %s\n", c.name, c.desc)
	}
	fmt.Fprint(w, `
Notes:
  tv play <local-file> serves the file over HTTP and starts a detached server
  (1h TTL), prints the URL, and returns immediately. Use --serve-forever to
  keep serving in the foreground or --serve-timeout <duration> to serve for a
  fixed time.

  With --json, commands print exactly one JSON value. The exception is
  streaming logcat (tv logs with no dump/lines/filter), which prints one JSON
  object per line; use -d for a single {"logs":...} object.
`)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Run 'tv help --json' for machine-readable help.")
}

func adbCmd(args []string) {
	cfg := loadAdbConfig()
	fs := newFlagSet("adb")
	parseFlags(fs, args)
	rest := fs.Args()
	if len(rest) < 1 {
		output.FailMsg("Usage: tv adb <connect|disconnect> [address]")
	}
	r := &adb.RealRunner{}
	target := cfg.ConnectTarget()
	if len(rest) >= 2 {
		target = rest[1]
	}
	switch rest[0] {
	case "connect":
		if !output.JSON {
			fmt.Printf("Connecting to %s...\n", target)
		}
		result, err := r.Connect(target)
		if err != nil {
			output.FailWith(fmt.Sprintf("Error: %v", err), err)
		}
		output.Result(output.AdbResult{Action: "connect", IP: target, Output: result}, func() {
			fmt.Println(result)
		})
	case "disconnect":
		if !output.JSON {
			fmt.Printf("Disconnecting from %s...\n", target)
		}
		result, err := r.Disconnect(target)
		if err != nil {
			output.FailWith(fmt.Sprintf("Error: %v", err), err)
		}
		output.Result(output.AdbResult{Action: "disconnect", IP: target, Output: result}, func() {
			fmt.Println(result)
		})
	default:
		output.FailMsg(fmt.Sprintf("unknown adb subcommand: %s", rest[0]))
	}
}

func appCmd(args []string) {
	if len(args) < 1 {
		output.FailMsg("Usage: tv app <list|launch|kill> [package]")
	}

	cfg := loadAdbConfig()
	r := &adb.RealRunner{}

	switch args[0] {
	case "list":
		if err := adb.EnsureConnected(cfg, r); err != nil {
			output.FailWith(fmt.Sprintf("Error: %v", err), err)
		}
		out, err := r.Shell(cfg, "pm list packages")
		if err != nil {
			output.FailWith(fmt.Sprintf("Error: %v", err), err)
		}
		output.Result(output.AppListResult{Packages: parsePackages(out)}, func() {
			fmt.Println(out)
		})
	case "launch":
		if len(args) < 2 {
			output.FailMsg("Usage: tv app launch <package>")
		}
		if err := adb.EnsureConnected(cfg, r); err != nil {
			output.FailWith(fmt.Sprintf("Error: %v", err), err)
		}
		_, err := r.ShellWithStderr(cfg, "monkey -p "+args[1]+" -c android.intent.category.LAUNCHER 1")
		if err != nil {
			output.FailWith(fmt.Sprintf("Error launching %s: %v", args[1], err), err)
		}
		output.Result(output.AppActionResult{Action: "launch", Package: args[1]}, func() {
			fmt.Printf("Launched %s\n", args[1])
		})
	case "kill":
		if len(args) < 2 {
			output.FailMsg("Usage: tv app kill <package>")
		}
		if err := adb.EnsureConnected(cfg, r); err != nil {
			output.FailWith(fmt.Sprintf("Error: %v", err), err)
		}
		_, err := r.ShellWithStderr(cfg, "am force-stop "+args[1])
		if err != nil {
			output.FailWith(fmt.Sprintf("Error killing %s: %v", args[1], err), err)
		}
		output.Result(output.AppActionResult{Action: "kill", Package: args[1]}, func() {
			fmt.Printf("Killed %s\n", args[1])
		})
	default:
		output.FailMsg(fmt.Sprintf("unknown app subcommand: %s", args[0]))
	}
}

func installCmd(args []string) {
	cfg := loadAdbConfig()
	r := &adb.RealRunner{}

	fs := newFlagSet("install")
	parseFlags(fs, args)

	remaining := fs.Args()
	if len(remaining) < 1 {
		output.FailMsg("Usage: tv install [options] <apk-file>")
	}
	apk := remaining[0]
	if _, err := os.Stat(apk); err != nil {
		output.FailWith(fmt.Sprintf("Error: cannot read APK: %v", err), fmt.Errorf("cannot read APK: %w", err))
	}
	if err := adb.EnsureConnected(cfg, r); err != nil {
		output.FailWith(fmt.Sprintf("Error: %v", err), err)
	}
	out, err := r.Install(cfg, apk)
	if err != nil {
		output.FailWith(fmt.Sprintf("Error installing %s: %v", apk, err), fmt.Errorf("failed to install %s: %w", apk, err))
	}
	output.Result(output.InstallResult{APK: apk, Installed: true, Output: out}, func() {
		fmt.Printf("Installed %s\n", apk)
		if out != "" {
			fmt.Println(out)
		}
	})
}

func configCmd(args []string) {
	cfg, err := config.Load()
	if err != nil {
		output.FailWith(fmt.Sprintf("Error: %v", err), err)
	}
	configFile := ""
	if dir, derr := config.Dir(); derr == nil {
		configFile = filepath.Join(dir, "config.json")
	}
	output.Result(output.ConfigResult{
		ConfigFile:    configFile,
		TVIP:          cfg.TVIP,
		DefaultPlayer: cfg.DefaultPlayer,
	}, func() {
		fmt.Printf("Config file: ~/.config/tv/config.json\n")
		fmt.Printf("TV IP:        %s\n", cfg.TVIP)
		if cfg.DefaultPlayer != "" {
			fmt.Printf("Default player: %s\n", cfg.DefaultPlayer)
		} else {
			fmt.Println("Default player: (none)")
		}
	})
}

func playerCmd(args []string) {
	cfg := loadAdbConfig()
	r := &adb.RealRunner{}

	fs := newFlagSet("player")
	parseFlags(fs, args)

	remaining := fs.Args()
	if len(remaining) < 1 {
		output.FailMsg("Usage: tv player <list|default> [options] [name]")
	}

	switch remaining[0] {
	case "list":
		players, err := player.Detect(cfg, r)
		if err != nil {
			output.FailWith(fmt.Sprintf("Error: %v", err), err)
		}
		result := output.PlayersResult{
			Players:       make([]output.PlayerInfo, 0, len(players)),
			DefaultPlayer: cfg.DefaultPlayer,
		}
		for _, p := range players {
			result.Players = append(result.Players, output.PlayerInfo{Name: p.Name, Package: p.PackageName})
		}
		output.Result(result, func() {
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
		})
	case "default":
		if len(remaining) < 2 {
			output.FailMsg("Usage: tv player default <name>")
		}
		name := remaining[1]
		pkg, err := player.Resolve(cfg, r, name)
		if err != nil {
			output.FailWith(fmt.Sprintf("Error: %v", err), err)
		}
		conf, err := config.Load()
		if err != nil {
			output.FailWith(fmt.Sprintf("Error: %v", err), err)
		}
		conf.DefaultPlayer = pkg
		if err := conf.Save(); err != nil {
			output.FailWith(fmt.Sprintf("Error saving config: %v", err), err)
		}
		output.Result(output.PlayerDefaultResult{DefaultPlayer: pkg}, func() {
			fmt.Printf("Default player set to: %s\n", pkg)
		})
	default:
		output.FailMsg(fmt.Sprintf("unknown player subcommand: %s", remaining[0]))
	}
}

func playCmd(args []string) {
	cfg := loadAdbConfig()
	r := &adb.RealRunner{}

	fs := newFlagSet("play")
	playerFlag := fs.String("player", cfg.DefaultPlayer, "Player name or package name")
	serveForever := fs.Bool("serve-forever", false, "Keep serving a local file until interrupted")
	serveTimeout := fs.Duration("serve-timeout", 0, "Serve a local file for this duration, e.g. 30m or 1h")
	parseFlags(fs, args)

	remaining := fs.Args()
	if len(remaining) < 1 {
		output.FailMsg(`Usage: tv play [options] <url|local-file>
  -player <name>          Player to use
  -ip <address>           TV IP address
  -serve-forever          Serve a local file in the foreground until interrupted
  -serve-timeout <dur>    Serve a local file in the foreground for a duration`)
	}

	target := remaining[0]
	playerPkg := *playerFlag
	if playerPkg == "" {
		players, err := player.Detect(cfg, r)
		if err != nil {
			output.FailWith(fmt.Sprintf("Error: %v", err), err)
		}
		output.FailMsg(noPlayerMessage(players))
	}
	if !strings.Contains(playerPkg, ".") {
		pkg, err := player.Resolve(cfg, r, playerPkg)
		if err != nil {
			output.FailWith(fmt.Sprintf("Error: %v", err), err)
		}
		playerPkg = pkg
	}
	if err := adb.EnsureConnected(cfg, r); err != nil {
		output.FailWith(fmt.Sprintf("Error: %v", err), err)
	}

	if fi, err := os.Stat(target); err == nil && !fi.IsDir() {
		abs, err := filepath.Abs(target)
		if err != nil {
			output.FailWith(fmt.Sprintf("Error: %v", err), err)
		}
		playLocal(cfg, r, playerPkg, abs, *serveForever, *serveTimeout)
		return
	}

	if !output.JSON {
		fmt.Printf("Playing on %s via %s...\n", cfg.TVIP, playerPkg)
	}
	if err := player.PlayOnTV(cfg, r, playerPkg, target); err != nil {
		output.FailWith(fmt.Sprintf("Error: %v", err), err)
	}
	output.Result(output.PlayResult{
		URL:    target,
		Player: playerPkg,
		TVIP:   cfg.TVIP,
		Played: true,
	}, func() {
		fmt.Println("Done.")
	})
}

func remoteCmd(args []string) {
	cfg := loadAdbConfig()
	r := &adb.RealRunner{}

	if len(args) < 1 {
		output.FailMsg(`Usage: tv remote <key>
Keys: up, down, left, right, ok, home, back, menu, power,
      input, info, subtitle, volup, voldown, mute,
      play, pause, stop, ff, rew, next, prev,
      chup, chdown, 0-9, type <text>`)
	}

	cmd := args[0]
	if cmd == "type" {
		if len(args) < 2 {
			output.FailMsg("Usage: tv remote type <text>")
		}
		if err := adb.EnsureConnected(cfg, r); err != nil {
			output.FailWith(fmt.Sprintf("Error: %v", err), err)
		}
		if err := remote.Type(cfg, r, args[1]); err != nil {
			output.FailWith(fmt.Sprintf("Error: %v", err), err)
		}
		output.Result(output.RemoteResult{Text: args[1], Sent: true}, func() {
			fmt.Printf("Typed: %s\n", args[1])
		})
		return
	}
	if err := adb.EnsureConnected(cfg, r); err != nil {
		output.FailWith(fmt.Sprintf("Error: %v", err), err)
	}
	if err := remote.Press(cfg, r, cmd); err != nil {
		output.FailWith(fmt.Sprintf("Error: %v", err), err)
	}
	keycode, _ := remote.Keycode(cmd)
	output.Result(output.RemoteResult{Key: cmd, Keycode: keycode, Sent: true}, func() {
		fmt.Printf("Sent key: %s\n", cmd)
	})
}

func logsCmd(args []string) {
	cfg := loadAdbConfig()
	r := &adb.RealRunner{}
	var opts logcat.Options

	fs := newFlagSet("logs")
	packageFlag := fs.String("p", "", "Filter by package name")
	levelFlag := fs.String("l", "", "Log level: v/d/i/w/e/f")
	tagFlag := fs.String("t", "", "Filter by log tag")
	linesFlag := fs.Int("n", 0, "Number of recent lines")
	formatFlag := fs.String("v", "", "Log format: brief, threadtime, etc.")
	clearFlag := fs.Bool("c", false, "Clear log buffer")
	dumpFlag := fs.Bool("d", false, "Dump and exit (no follow)")
	parseFlags(fs, args)

	opts.Level = *levelFlag
	opts.Tag = *tagFlag
	opts.Lines = *linesFlag
	opts.Format = *formatFlag
	opts.Clear = *clearFlag
	opts.Dump = *dumpFlag

	if len(fs.Args()) > 0 && fs.Args()[0] == "dump" {
		opts.Dump = true
	}

	if err := adb.EnsureConnected(cfg, r); err != nil {
		output.FailWith(fmt.Sprintf("Error: %v", err), err)
	}

	result := output.LogsResult{Lines: []string{}}
	if *packageFlag != "" {
		pkg := *packageFlag
		if !strings.Contains(pkg, ".") {
			var err error
			pkg, err = player.Resolve(cfg, r, pkg)
			if err != nil {
				output.FailWith(fmt.Sprintf("Error: %v", err), err)
			}
		}
		pid, err := logcat.ResolvePID(cfg, r, pkg)
		if err != nil {
			output.FailWith(fmt.Sprintf("Error: %v", err), err)
		}
		opts.Package = strconv.Itoa(pid)
		result.Package = pkg
		result.PID = pid
		if !output.JSON {
			fmt.Printf("Filtering by %s (PID %d)\n", pkg, pid)
		}
	}

	if opts.Clear {
		_, err := logcat.RunOnce(cfg, r, opts)
		if err != nil {
			output.FailWith(fmt.Sprintf("Error: %v", err), err)
		}
		output.Result(output.ClearResult{Cleared: true}, func() {
			fmt.Println("Log buffer cleared.")
		})
		return
	}

	if opts.Dump || opts.Lines > 0 || opts.Level != "" || opts.Tag != "" {
		out, err := logcat.RunOnce(cfg, r, opts)
		if err != nil {
			output.FailWith(fmt.Sprintf("Error: %v", err), err)
		}
		result.Logs = out
		result.Lines = splitLines(out)
		output.Result(result, func() {
			fmt.Println(out)
		})
		return
	}

	if output.JSON {
		w := output.NewJSONLineWriter(os.Stdout)
		err := logcat.RunStreamTo(cfg, opts, w, os.Stderr)
		_ = w.Flush()
		if err != nil {
			output.FailWith(fmt.Sprintf("Error: %v", err), err)
		}
		return
	}

	fmt.Println("Streaming logs (Ctrl-C to stop)...")
	if err := logcat.RunStream(cfg, opts); err != nil {
		output.FailWith(fmt.Sprintf("Error: %v", err), err)
	}
}

// noPlayerMessage explains how to pick a player and lists what is installed.
func noPlayerMessage(players []player.Info) string {
	var msg strings.Builder
	msg.WriteString("No player specified and no default player is set.\n")
	if len(players) == 0 {
		msg.WriteString("No video players were found on the TV.")
		return msg.String()
	}
	msg.WriteString("Available players:\n")
	for _, p := range players {
		fmt.Fprintf(&msg, "  %-20s %s\n", p.Name, p.PackageName)
	}
	msg.WriteString("Set one with 'tv player default <name>' or pass -player <name>.")
	return msg.String()
}

func loadAdbConfig() *adb.Config {
	c, err := config.Load()
	if err != nil {
		output.Warn(fmt.Errorf("config load failed: %v (using defaults)", err))
		c = &config.Config{TVIP: "192.168.2.3:5555"}
	}
	return newAdbConfig(c)
}

// parsePackages extracts package names from `pm list packages` output.
func parsePackages(out string) []string {
	packages := make([]string, 0)
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if pkg, ok := strings.CutPrefix(line, "package:"); ok && pkg != "" {
			packages = append(packages, pkg)
		}
	}
	sort.Strings(packages)
	return packages
}

// splitLines turns logcat output into a non-nil slice of lines.
func splitLines(out string) []string {
	out = strings.TrimRight(out, "\n")
	if out == "" {
		return []string{}
	}
	return strings.Split(out, "\n")
}

func isWindows() bool {
	return len(os.Getenv("SYSTEMROOT")) > 0
}

func isMac() bool {
	if len(os.Getenv("HOME")) == 0 {
		return false
	}
	_, err := exec.LookPath("open")
	return err == nil
}

func execCommand(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Start()
}

func devicesCmd(args []string) {
	fs := newFlagSet("devices")
	parseFlags(fs, args)

	r := &adb.RealRunner{}
	out, err := r.Devices()
	if err != nil {
		output.FailWith(fmt.Sprintf("Error: %v", err), err)
	}

	devices := adb.ParseDevices(out)
	if devices == nil {
		devices = []adb.Device{}
	}

	output.Result(devices, func() {
		if len(devices) == 0 {
			fmt.Println("No devices attached.")
			return
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "SERIAL\tSTATE\tMODEL\tPRODUCT\tDEVICE\tTRANSPORT")
		for _, d := range devices {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", d.Serial, d.State, d.Model, d.Product, d.Device, d.Transport)
		}
		w.Flush()
	})
}

func discoverCmd(args []string) {
	fs := newFlagSet("discover")
	kindFlag := fs.String("kind", "", "Filter by kind: pairing, connect, legacy")
	parseFlags(fs, args)

	d := &discovery.ADBDiscovery{Runner: &adb.RealRunner{}}
	services, err := d.Services()
	if err != nil {
		output.FailWith(fmt.Sprintf("Error: %v", err), err)
	}

	visible := make([]discovery.Service, 0, len(services))
	for _, s := range services {
		if *kindFlag == "" || string(s.Kind) == *kindFlag {
			visible = append(visible, s)
		}
	}

	output.Result(visible, func() {
		if len(visible) == 0 {
			fmt.Println("No wireless debugging services found. Enable Wireless debugging on the TV (Android 11+).")
			return
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "KIND\tADDRESS\tNAME\tTYPE")
		for _, s := range visible {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", s.Kind, s.Address, s.Name, s.Type)
		}
		w.Flush()
		for _, s := range visible {
			if s.Kind == discovery.KindPairing {
				fmt.Printf("\nPair with:    tv pair %s <code>\n", s.Address)
			}
		}
		for _, s := range visible {
			if s.Kind == discovery.KindConnect {
				fmt.Printf("Connect with: tv adb connect %s\n", s.Address)
			}
		}
	})
}

func pairCmd(args []string) {
	fs := newFlagSet("pair")
	connectAddr := fs.String("connect", "", "Address to connect to after pairing (defaults to the pairing address)")
	parseFlags(fs, args)
	rest := fs.Args()

	if len(rest) < 2 {
		output.FailMsg("Usage: tv pair <ip:port> <code>")
	}
	addr, code := rest[0], rest[1]
	r := &adb.RealRunner{}

	if !output.JSON {
		fmt.Printf("Pairing with %s...\n", addr)
	}
	out, err := adb.PairDevice(r, addr, code)
	if err != nil {
		output.FailWith(fmt.Sprintf("Error: %v", err), err)
	}

	target := addr
	if *connectAddr != "" {
		target = *connectAddr
	}
	if !output.JSON {
		fmt.Printf("Connecting to %s...\n", target)
	}
	connectOut, err := r.Connect(target)
	if err != nil || connectFailed(connectOut) {
		output.FailWith(fmt.Sprintf("Paired successfully, but connecting to %s failed: %s", target, strings.TrimSpace(connectOut)), fmt.Errorf("connect failed: %s", strings.TrimSpace(connectOut)))
	}
	output.Result(map[string]any{
		"paired":    strings.TrimSpace(out),
		"target":    target,
		"connected": strings.TrimSpace(connectOut),
	}, func() {
		fmt.Println(connectOut)
	})
}

func connectFailed(out string) bool {
	low := strings.ToLower(out)
	for _, marker := range []string{"failed", "cannot", "unable", "refused"} {
		if strings.Contains(low, marker) {
			return true
		}
	}
	return false
}

func newAdbConfig(c *config.Config) *adb.Config {
	cfg := &adb.Config{
		TVIP:          c.TVIP,
		Device:        c.Device,
		IP:            c.IP,
		DefaultPlayer: c.DefaultPlayer,
	}
	if globalDevice != "" {
		cfg.Device = globalDevice
	}
	if globalIP != "" {
		cfg.IP = globalIP
	}
	return cfg
}
