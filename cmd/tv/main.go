package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/mutedfalcontv/tv/internal/adb"
	"github.com/mutedfalcontv/tv/internal/config"
	"github.com/mutedfalcontv/tv/internal/discovery"
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
	{"devices", "List connected ADB devices", devicesCmd},
	{"discover", "Discover TVs via mDNS (wireless debugging)", discoverCmd},
	{"logs", "View and filter TV logs (logcat)", logsCmd},
	{"pair", "Pair with a TV over wireless ADB", pairCmd},
	{"play", "Play a URL on TV", playCmd},
	{"player", "List/set default video player", playerCmd},
	{"remote", "TV remote control", remoteCmd},
	{"serve", "Start TV Remote UI web server", serveCmd},
}

var (
	globalDevice string
	globalIP     string
)

func main() {
	args := parseGlobalFlags(os.Args[1:])
	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}
	sub := args[0]
	for _, c := range commands {
		if c.name == sub {
			c.run(args[1:])
			return
		}
	}
	fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", sub)
	printUsage()
	os.Exit(1)
}

// parseGlobalFlags consumes --device/--ip that appear before the command.
func parseGlobalFlags(args []string) []string {
	for len(args) >= 2 {
		switch args[0] {
		case "--device", "-device":
			globalDevice = args[1]
			args = args[2:]
		case "--ip", "-ip":
			globalIP = args[1]
			args = args[2:]
		default:
			if v, ok := strings.CutPrefix(args[0], "--device="); ok {
				globalDevice = v
				args = args[1:]
				continue
			}
			if v, ok := strings.CutPrefix(args[0], "-device="); ok {
				globalDevice = v
				args = args[1:]
				continue
			}
			if v, ok := strings.CutPrefix(args[0], "--ip="); ok {
				globalIP = v
				args = args[1:]
				continue
			}
			if v, ok := strings.CutPrefix(args[0], "-ip="); ok {
				globalIP = v
				args = args[1:]
				continue
			}
			return args
		}
	}
	return args
}

func printUsage() {
	fmt.Fprint(os.Stderr, `tv — Android TV CLI tools

Usage:
  tv [global flags] <command> [options] [args]

Global flags (before the command):
  --device <serial>   Target a specific device (adb -s <serial>)
  --ip <address>      Target a specific device IP address

Commands:
`)
	for _, c := range commands {
		fmt.Fprintf(os.Stderr, "  %-15s %s\n", c.name, c.desc)
	}
}

func adbCmd(args []string) {
	cfg := loadAdbConfig()
	fs := flag.NewFlagSet("adb", flag.ExitOnError)
	addDeviceFlags(fs, cfg)
	fs.Parse(args)
	rest := fs.Args()

	if len(rest) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: tv adb <connect|disconnect> [address]")
		os.Exit(1)
	}
	r := &adb.RealRunner{}
	target := cfg.ConnectTarget()
	if len(rest) >= 2 {
		target = rest[1]
	}
	switch rest[0] {
	case "connect":
		fmt.Printf("Connecting to %s...\n", target)
		result, err := r.Connect(target)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(result)
	case "disconnect":
		fmt.Printf("Disconnecting from %s...\n", target)
		result, err := r.Disconnect(target)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(result)
	default:
		fmt.Fprintf(os.Stderr, "unknown adb subcommand: %s\n", rest[0])
		os.Exit(1)
	}
}

func appCmd(args []string) {
	cfg := loadAdbConfig()
	fs := flag.NewFlagSet("app", flag.ExitOnError)
	addDeviceFlags(fs, cfg)
	fs.Parse(args)
	rest := fs.Args()

	if len(rest) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: tv app <list|launch|kill> [package]")
		os.Exit(1)
	}

	r := &adb.RealRunner{}

	switch rest[0] {
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
		if len(rest) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: tv app launch <package>")
			os.Exit(1)
		}
		if err := adb.EnsureConnected(cfg, r); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		_, err := r.ShellWithStderr(cfg, "monkey -p "+rest[1]+" -c android.intent.category.LAUNCHER 1")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error launching %s: %v\n", rest[1], err)
			os.Exit(1)
		}
		fmt.Printf("Launched %s\n", rest[1])
	case "kill":
		if len(rest) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: tv app kill <package>")
			os.Exit(1)
		}
		if err := adb.EnsureConnected(cfg, r); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		_, err := r.ShellWithStderr(cfg, "am force-stop "+rest[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error killing %s: %v\n", rest[1], err)
			os.Exit(1)
		}
		fmt.Printf("Killed %s\n", rest[1])
	default:
		fmt.Fprintf(os.Stderr, "unknown app subcommand: %s\n", rest[0])
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
	if cfg.Device != "" {
		fmt.Printf("Device serial: %s\n", cfg.Device)
	}
	if cfg.IP != "" {
		fmt.Printf("Device IP:     %s\n", cfg.IP)
	}
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
	addDeviceFlags(fs, cfg)
	fs.Parse(args)

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
	addDeviceFlags(fs, cfg)
	fs.Parse(args)

	remaining := fs.Args()
	if len(remaining) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: tv play [options] <url>")
		fmt.Fprintln(os.Stderr, "  -player <name>   Player to use")
		fmt.Fprintln(os.Stderr, "  -device <serial> Target device serial")
		fmt.Fprintln(os.Stderr, "  -ip <address>    Target device IP address")
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
	fmt.Printf("Playing on %s via %s...\n", cfg.Serial(), playerPkg)
	if err := player.PlayOnTV(cfg, r, playerPkg, url); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Done.")
}

func remoteCmd(args []string) {
	cfg := loadAdbConfig()
	r := &adb.RealRunner{}

	fs := flag.NewFlagSet("remote", flag.ExitOnError)
	addDeviceFlags(fs, cfg)
	fs.Parse(args)
	rest := fs.Args()

	if len(rest) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: tv remote <key>")
		fmt.Fprintln(os.Stderr, "Keys: up, down, left, right, ok, home, back, menu, power,")
		fmt.Fprintln(os.Stderr, "      input, info, subtitle, volup, voldown, mute,")
		fmt.Fprintln(os.Stderr, "      play, pause, stop, ff, rew, next, prev,")
		fmt.Fprintln(os.Stderr, "      chup, chdown, 0-9, type <text>")
		os.Exit(1)
	}

	cmd := rest[0]
	if cmd == "type" {
		if len(rest) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: tv remote type <text>")
			os.Exit(1)
		}
		if err := adb.EnsureConnected(cfg, r); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if err := remote.Type(cfg, r, rest[1]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Typed: %s\n", rest[1])
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
	addDeviceFlags(fs, cfg)
	fs.Parse(args)

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
		opts.Package = strconv.Itoa(pid)
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

func devicesCmd(args []string) {
	fs := flag.NewFlagSet("devices", flag.ExitOnError)
	jsonFlag := fs.Bool("json", false, "Output JSON")
	fs.Parse(args)

	r := &adb.RealRunner{}
	out, err := r.Devices()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	devices := adb.ParseDevices(out)
	if devices == nil {
		devices = []adb.Device{}
	}

	if *jsonFlag {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(devices); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

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
}

func discoverCmd(args []string) {
	fs := flag.NewFlagSet("discover", flag.ExitOnError)
	jsonFlag := fs.Bool("json", false, "Output JSON")
	kindFlag := fs.String("kind", "", "Filter by kind: pairing, connect, legacy")
	fs.Parse(args)

	d := &discovery.ADBDiscovery{Runner: &adb.RealRunner{}}
	services, err := d.Services()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	visible := make([]discovery.Service, 0, len(services))
	for _, s := range services {
		if *kindFlag == "" || string(s.Kind) == *kindFlag {
			visible = append(visible, s)
		}
	}

	if *jsonFlag {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(visible); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

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
}

func pairCmd(args []string) {
	fs := flag.NewFlagSet("pair", flag.ExitOnError)
	connectAddr := fs.String("connect", "", "Address to connect to after pairing (defaults to the pairing address)")
	fs.Parse(args)
	rest := fs.Args()

	if len(rest) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: tv pair <ip:port> <code>")
		fmt.Fprintln(os.Stderr, "  Find the pairing address with 'tv discover'.")
		os.Exit(1)
	}
	addr, code := rest[0], rest[1]
	r := &adb.RealRunner{}

	fmt.Printf("Pairing with %s...\n", addr)
	out, err := adb.PairDevice(r, addr, code)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(out)

	target := addr
	if *connectAddr != "" {
		target = *connectAddr
	}
	fmt.Printf("Connecting to %s...\n", target)
	connectOut, err := r.Connect(target)
	if err != nil || connectFailed(connectOut) {
		fmt.Fprintf(os.Stderr, "Paired successfully, but connecting to %s failed: %s\n", target, strings.TrimSpace(connectOut))
		fmt.Fprintln(os.Stderr, "Android TV uses a separate port for connections. Run 'tv discover' and use 'tv adb connect <connect-address>'.")
		os.Exit(1)
	}
	fmt.Println(connectOut)
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

func addDeviceFlags(fs *flag.FlagSet, cfg *adb.Config) {
	fs.StringVar(&cfg.Device, "device", cfg.Device, "Target device serial (adb -s <serial>)")
	fs.StringVar(&cfg.IP, "ip", cfg.IP, "Target device IP address")
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

func loadAdbConfig() *adb.Config {
	c, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: config load failed: %v (using defaults)\n", err)
		c = &config.Config{TVIP: "192.168.2.3:5555"}
	}
	return newAdbConfig(c)
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
