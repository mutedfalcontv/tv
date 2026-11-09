package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/mutedfalcontv/tv/internal/adb"
	"github.com/mutedfalcontv/tv/internal/config"
	"github.com/mutedfalcontv/tv/internal/serve"
)

func serveCmd(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	port := fs.Int("port", 8080, "HTTP server port")
	host := fs.String("host", "127.0.0.1", "HTTP server host (use 0.0.0.0 to expose on the network)")
	token := fs.String("token", "", "Require this bearer token on /api/* and /ws/* (default: config token)")
	cors := fs.String("cors", "", "Comma-separated origins allowed for CORS (default: none; loopback always allowed)")
	noOpen := fs.Bool("no-open", false, "Don't open browser")
	fs.Parse(args)

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %v (using defaults)\n", err)
		cfg = &config.Config{TVIP: "192.168.2.3:5555"}
	}

	authToken := *token
	if authToken == "" {
		authToken = cfg.Token
	}
	var corsOrigins []string
	if *cors != "" {
		for _, o := range strings.Split(*cors, ",") {
			if o = strings.TrimSpace(o); o != "" {
				corsOrigins = append(corsOrigins, o)
			}
		}
	} else {
		corsOrigins = cfg.CORSOrigins
	}

	if !isLoopbackHost(*host) {
		fmt.Fprintf(os.Stderr, "Warning: binding to non-loopback host %q — the server is reachable from the network\n", *host)
		if authToken == "" {
			fmt.Fprintln(os.Stderr, "Warning: no token set; anyone who can reach this host can control the TV. Consider --token.")
		}
	}

	adbCfg := &adb.Config{TVIP: cfg.TVIP, DefaultPlayer: cfg.DefaultPlayer}
	runner := &adb.RealRunner{}

	addr := fmt.Sprintf("%s:%d", *host, *port)
	server := serve.New(adbCfg, runner, serve.Options{Token: authToken, CORSOrigins: corsOrigins})

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if !*noOpen {
		openBrowser(fmt.Sprintf("http://localhost:%d", *port))
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

func isLoopbackHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	if ip := net.ParseIP(host); ip != nil {
		return ip.IsLoopback()
	}
	return false
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
