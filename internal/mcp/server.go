package mcp

import (
	"context"
	"time"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/mutedfalcontv/tv/internal/adb"
)

const (
	ServerName    = "tv"
	ServerVersion = "1.0.0"

	// DefaultTimeout bounds each individual ADB call made by a tool handler.
	DefaultTimeout = 30 * time.Second
)

const serverInstructions = `Tools for controlling an Android TV over ADB.
Start with tv_status to confirm the TV is reachable, then use tv_connect if needed.
Apps use Android package names (e.g. com.netflix.ninja). Use tv_list_apps or
tv_players to discover installed packages before launching or killing them.
Remote keys accepted by tv_press_key: up, down, left, right, ok, home, back,
menu, power, input, info, subtitle, volup, voldown, mute, play, pause, stop,
ff, rew, next, prev, chup, chdown, and 0-9.
tv_logs always returns a bounded snapshot; it never streams.`

// Server exposes Android TV control as MCP tools over a mockable adb.Runner.
type Server struct {
	cfg     *adb.Config
	runner  adb.Runner
	timeout time.Duration
	mcp     *mcpsdk.Server
}

// New builds an MCP server that wraps cfg and runner.
func New(cfg *adb.Config, runner adb.Runner) *Server {
	s := &Server{
		cfg:     cfg,
		runner:  runner,
		timeout: DefaultTimeout,
	}
	s.mcp = mcpsdk.NewServer(
		&mcpsdk.Implementation{Name: ServerName, Version: ServerVersion},
		&mcpsdk.ServerOptions{Instructions: serverInstructions},
	)
	s.registerTools()
	return s
}

// SetTimeout overrides the per-ADB-call timeout. Zero keeps the default.
func (s *Server) SetTimeout(d time.Duration) {
	if d > 0 {
		s.timeout = d
	}
}

// MCP returns the underlying SDK server, primarily for embedding and tests.
func (s *Server) MCP() *mcpsdk.Server { return s.mcp }

// Run serves MCP over stdin/stdout until ctx is cancelled or the client exits.
func (s *Server) Run(ctx context.Context) error {
	return s.mcp.Run(ctx, &mcpsdk.StdioTransport{})
}

// runner returns the configured runner wrapped with a per-call timeout.
func (s *Server) runnerWithTimeout() adb.Runner {
	return &timeoutRunner{inner: s.runner, timeout: s.timeout}
}
