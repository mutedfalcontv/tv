package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/mutedfalcontv/tv/internal/adb"
	"github.com/mutedfalcontv/tv/internal/config"
	tvmcp "github.com/mutedfalcontv/tv/internal/mcp"
)

func mcpCmd(args []string) {
	fs := flag.NewFlagSet("mcp", flag.ExitOnError)
	ip := fs.String("ip", "", "TV IP address (overrides config)")
	timeout := fs.Duration("timeout", tvmcp.DefaultTimeout, "Per-ADB-call timeout")
	fs.Parse(args)

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %v (using defaults)\n", err)
		cfg = &config.Config{TVIP: "192.168.2.3:5555"}
	}
	if *ip != "" {
		cfg.TVIP = *ip
	}

	adbCfg := &adb.Config{TVIP: cfg.TVIP, DefaultPlayer: cfg.DefaultPlayer}
	server := tvmcp.New(adbCfg, &adb.RealRunner{})
	server.SetTimeout(*timeout)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := server.Run(ctx); err != nil && ctx.Err() == nil {
		fmt.Fprintf(os.Stderr, "MCP server error: %v\n", err)
		os.Exit(1)
	}
}
