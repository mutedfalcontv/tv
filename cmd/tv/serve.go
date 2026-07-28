package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mutedfalcontv/tv/internal/adb"
	"github.com/mutedfalcontv/tv/internal/config"
	"github.com/mutedfalcontv/tv/internal/serve"
)

func serveCmd(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	port := fs.Int("port", 8080, "HTTP server port")
	host := fs.String("host", "0.0.0.0", "HTTP server host")
	noOpen := fs.Bool("no-open", false, "Don't open browser")
	fs.Parse(args)

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %v (using defaults)\n", err)
		cfg = &config.Config{TVIP: "192.168.2.3:5555"}
	}

	adbCfg := &adb.Config{TVIP: cfg.TVIP, DefaultPlayer: cfg.DefaultPlayer}
	runner := &adb.RealRunner{}

	addr := fmt.Sprintf("%s:%d", *host, *port)
	server := serve.New(adbCfg, runner)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if !*noOpen {
		url := fmt.Sprintf("http://localhost:%d", *port)
		openBrowser(url)
	}

	go func() {
		if err := server.Start(addr); err != nil {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			stop()
		}
	}()

	<-ctx.Done()
	fmt.Println("\nShutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(shutdownCtx)
}

func openBrowser(url string) {
	switch {
	case isWindows():
		execCommand("rundll32", "url.dll,FileProtocolHandler", url)
	case isMac():
		execCommand("open", url)
	default:
		execCommand("xdg-open", url)
	}
}
