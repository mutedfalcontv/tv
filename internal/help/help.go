// Package help exposes a machine-readable description of the tv command tree.
package help

import (
	"encoding/json"

	"github.com/mutedfalcontv/tv/internal/remote"
)

// Version is reported by the JSON help document.
const Version = "0.1.0"

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
	Commands    []Command `json:"commands"`
}

// Tree returns the full command tree.
func Tree() Document {
	return Document{
		Name:    "tv",
		Version: Version,
		Description: "Android TV CLI and HTTP API. Controls an Android TV over ADB " +
			"(connect, manage apps, send remote keys, type text, play URLs, read logs).",
		Commands: []Command{
			{
				Name:        "adb",
				Description: "ADB connection management",
				Usage:       "tv adb <connect|disconnect>",
				Subcommands: []Command{
					{Name: "connect", Description: "Connect to the configured TV", Usage: "tv adb connect"},
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
				Description: "Show configuration (TV IP and default player)",
				Usage:       "tv config",
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
				Name:        "play",
				Description: "Play a URL on TV using a video player",
				Usage:       "tv play [options] <url>",
				Arguments: []Argument{
					{Name: "url", Description: "Media URL to open (stream or file)", Required: true},
				},
				Flags: []Flag{
					{Name: "-player", Type: "string", Description: "Player name or package name (defaults to configured default player)"},
					{Name: "-ip", Type: "string", Description: "TV IP address (defaults to configured TV IP)"},
				},
			},
			{
				Name:        "player",
				Description: "List installed video players or set the default one",
				Usage:       "tv player <list|default> [options] [name]",
				Flags: []Flag{
					{Name: "-ip", Type: "string", Description: "TV IP address (defaults to configured TV IP)"},
				},
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
				Name:        "serve",
				Description: "Start the TV Remote UI web server and HTTP API",
				Usage:       "tv serve [options]",
				Flags: []Flag{
					{Name: "-port", Type: "int", Default: "8080", Description: "HTTP server port"},
					{Name: "-host", Type: "string", Default: "0.0.0.0", Description: "HTTP server host"},
					{Name: "-no-open", Type: "bool", Default: "false", Description: "Do not open the browser on start"},
				},
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
