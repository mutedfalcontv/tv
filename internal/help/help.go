// Package help exposes a machine-readable description of the tv command tree.
package help

import (
	"encoding/json"

	"github.com/mutedfalcontv/tv/internal/remote"
)

// Version is reported by the JSON help document.
// Overridden at build time via -ldflags "-X github.com/mutedfalcontv/tv/internal/help.Version=<tag>".
var Version = "dev"

// Flag describes a command-line flag.
type Flag struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Default     string `json:"default,omitempty"`
	Description string `json:"description"`
}

// Argument describes a positional argument.
type Argument struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Required    bool     `json:"required"`
	Choices     []string `json:"choices,omitempty"`
}

// Command describes a command or subcommand.
type Command struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Usage       string     `json:"usage"`
	Arguments   []Argument `json:"arguments,omitempty"`
	Flags       []Flag     `json:"flags,omitempty"`
	Subcommands []Command  `json:"subcommands,omitempty"`
}

// Document is the root of the JSON help output.
type Document struct {
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	Description string    `json:"description"`
	GlobalFlags []Flag    `json:"global_flags,omitempty"`
	Commands    []Command `json:"commands"`
}

// Tree returns the full command tree.
func Tree() Document {
	return Document{
		Name:    "tv",
		Version: Version,
		Description: "Android TV CLI and HTTP API. Controls an Android TV over ADB " +
			"(connect, manage apps, send remote keys, type text, play URLs, read logs, " +
			"capture screenshots, inspect state).",
		GlobalFlags: []Flag{
			{Name: "--json", Type: "bool", Default: "false", Description: "Emit exactly one JSON value on stdout; errors are JSON on stderr"},
			{Name: "--no-input", Type: "bool", Default: "false", Description: "Never prompt (automatic when stdin is not a TTY)"},
			{Name: "--device", Type: "string", Description: "Target a specific ADB device serial (adb -s <serial>)"},
			{Name: "--ip", Type: "string", Description: "Target a specific device IP address"},
		},
		Commands: []Command{
			{
				Name:        "adb",
				Description: "ADB connection management",
				Usage:       "tv adb <connect|disconnect> [address]",
				Subcommands: []Command{
					{Name: "connect", Description: "Connect to the configured TV (or an explicit address)", Usage: "tv adb connect [address]"},
					{Name: "disconnect", Description: "Disconnect from the configured TV", Usage: "tv adb disconnect"},
				},
			},
			{
				Name:        "app",
				Description: "App management (list, launch, kill)",
				Usage:       "tv app <list|launch|kill> [package]",
				Subcommands: []Command{
					{Name: "list", Description: "List installed packages", Usage: "tv app list"},
					{
						Name:        "launch",
						Description: "Launch an app by package name",
						Usage:       "tv app launch <package>",
						Arguments:   []Argument{{Name: "package", Description: "Android package name", Required: true}},
					},
					{
						Name:        "kill",
						Description: "Force-stop an app by package name",
						Usage:       "tv app kill <package>",
						Arguments:   []Argument{{Name: "package", Description: "Android package name", Required: true}},
					},
				},
			},
			{
				Name:        "config",
				Description: "Show configuration (TV IP/device, default player)",
				Usage:       "tv config",
			},
			{
				Name:        "devices",
				Description: "List connected ADB devices",
				Usage:       "tv devices",
			},
			{
				Name:        "discover",
				Description: "Discover TVs via mDNS (wireless debugging)",
				Usage:       "tv discover [options]",
				Flags: []Flag{
					{Name: "-kind", Type: "string", Description: "Filter by kind: pairing, connect, legacy"},
				},
			},
			{
				Name:        "install",
				Description: "Install an APK on the TV",
				Usage:       "tv install <apk-file>",
				Arguments:   []Argument{{Name: "apk-file", Description: "Path to the APK to install", Required: true}},
			},
			{
				Name:        "logs",
				Description: "View and filter TV logs (logcat)",
				Usage:       "tv logs [options] [dump]",
				Arguments: []Argument{
					{Name: "dump", Description: "Dump the current buffer and exit instead of streaming", Required: false},
				},
				Flags: []Flag{
					{Name: "-p", Type: "string", Description: "Filter by package name (resolved to PID)"},
					{Name: "-l", Type: "string", Description: "Log level: v, d, i, w, e, f (or names)"},
					{Name: "-t", Type: "string", Description: "Filter by log tag"},
					{Name: "-n", Type: "int", Description: "Number of recent lines (implies dump mode)"},
					{Name: "-v", Type: "string", Description: "Log format: brief, threadtime, ..."},
					{Name: "-c", Type: "bool", Default: "false", Description: "Clear the log buffer and exit"},
					{Name: "-d", Type: "bool", Default: "false", Description: "Dump the buffer and exit (no follow)"},
				},
			},
			{
				Name:        "mcp",
				Description: "Start a stdio MCP server exposing TV tools to AI agents",
				Usage:       "tv mcp [options]",
				Flags: []Flag{
					{Name: "-timeout", Type: "duration", Default: "30s", Description: "Per-ADB-call timeout"},
				},
			},
			{
				Name:        "pair",
				Description: "Pair with a TV over wireless ADB (Android 11+)",
				Usage:       "tv pair <ip:port> <code>",
				Arguments: []Argument{
					{Name: "address", Description: "Pairing address from `tv discover`", Required: true},
					{Name: "code", Description: "6-digit pairing code shown on the TV", Required: true},
				},
				Flags: []Flag{
					{Name: "-connect", Type: "string", Description: "Address to connect to after pairing (defaults to the pairing address)"},
				},
			},
			{
				Name:        "play",
				Description: "Play a URL or local file on TV using a video player",
				Usage:       "tv play [options] <url|local-file>",
				Arguments: []Argument{
					{Name: "url", Description: "Media URL or local file path", Required: true},
				},
				Flags: []Flag{
					{Name: "-player", Type: "string", Description: "Player name or package name (defaults to configured default player)"},
					{Name: "-serve-forever", Type: "bool", Default: "false", Description: "Serve a local file until interrupted"},
					{Name: "-serve-timeout", Type: "duration", Default: "0", Description: "Serve a local file for this duration"},
				},
			},
			{
				Name:        "player",
				Description: "List installed video players or set the default one",
				Usage:       "tv player <list|default> [name]",
				Subcommands: []Command{
					{Name: "list", Description: "List detected video players and mark the default", Usage: "tv player list"},
					{
						Name:        "default",
						Description: "Set the default player by name or package",
						Usage:       "tv player default <name>",
						Arguments:   []Argument{{Name: "name", Description: "Player display name or package name", Required: true}},
					},
				},
			},
			{
				Name: "remote",
				Description: "TV remote control. Sends a named key or types text. " +
					"Available keys are listed in the key argument choices.",
				Usage: "tv remote <key> | tv remote type <text>",
				Arguments: []Argument{
					{
						Name:        "key",
						Description: "Named remote key to press",
						Required:    true,
						Choices:     remote.Keys(),
					},
				},
				Subcommands: []Command{
					{
						Name:        "type",
						Description: "Type text into the focused field",
						Usage:       "tv remote type <text>",
						Arguments:   []Argument{{Name: "text", Description: "Text to type", Required: true}},
					},
				},
			},
			{
				Name:        "screenshot",
				Description: "Capture the TV screen to a PNG",
				Usage:       "tv screenshot [options]",
				Flags: []Flag{
					{Name: "-o", Type: "string", Description: "Output PNG path (default ./screencap-<unix-ts>.png)"},
				},
			},
			{
				Name:        "serve",
				Description: "Start the TV Remote UI web server and HTTP API",
				Usage:       "tv serve [options]",
				Flags: []Flag{
					{Name: "-port", Type: "int", Default: "8080", Description: "HTTP server port"},
					{Name: "-host", Type: "string", Default: "127.0.0.1", Description: "HTTP server host (use 0.0.0.0 to expose)"},
					{Name: "-token", Type: "string", Description: "Require this bearer token on /api/* and /ws/*"},
					{Name: "-cors", Type: "string", Description: "Comma-separated origins allowed for CORS"},
					{Name: "-no-open", Type: "bool", Default: "false", Description: "Do not open the browser on start"},
				},
			},
			{
				Name:        "state",
				Description: "Show live TV state from dumpsys (foreground app, volume, screen)",
				Usage:       "tv state",
			},
			{
				Name:        "help",
				Description: "Show help. Use --json for machine-readable output.",
				Usage:       "tv help [--json]",
				Flags: []Flag{
					{Name: "--json", Type: "bool", Default: "false", Description: "Emit the command tree as JSON"},
				},
			},
		},
	}
}

// JSON renders the command tree as indented JSON.
func JSON() ([]byte, error) {
	return json.MarshalIndent(Tree(), "", "  ")
}
