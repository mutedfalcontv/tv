package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/mutedfalcontv/tv/internal/adb"
	"github.com/mutedfalcontv/tv/internal/config"
	tvmcp "github.com/mutedfalcontv/tv/internal/mcp"
	"github.com/mutedfalcontv/tv/internal/output"
)

func mcpCmd(args []string) {
	fs := newFlagSet("mcp")
	timeout := fs.Duration("timeout", tvmcp.DefaultTimeout, "Per-ADB-call timeout")
	parseFlags(fs, args)

	cfg, err := config.Load()
	if err != nil {
		output.Warn(fmt.Errorf("%v (using defaults)", err))
		cfg = &config.Config{TVIP: "192.168.2.3:5555"}
	}

	adbCfg := newAdbConfig(cfg)
	server := tvmcp.New(adbCfg, &adb.RealRunner{Timeout: *timeout})
	server.SetTimeout(*timeout)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := server.Run(ctx); err != nil && ctx.Err() == nil {
		output.FailWith(fmt.Sprintf("MCP server error: %v", err), err)
	}
}
