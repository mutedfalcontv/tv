package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/mutedfalcontv/tv/internal/adb"
	"github.com/mutedfalcontv/tv/internal/config"
	"github.com/mutedfalcontv/tv/internal/output"
	"github.com/mutedfalcontv/tv/internal/serve"
)

func serveCmd(args []string) {
	fs := newFlagSet("serve")
	port := fs.Int("port", 8080, "HTTP server port")
	host := fs.String("host", "127.0.0.1", "HTTP server host (use 0.0.0.0 to expose on the network)")
	token := fs.String("token", "", "Require this bearer token on /api/* and /ws/* (default: config token)")
	cors := fs.String("cors", "", "Comma-separated origins allowed for CORS (default: none; loopback always allowed)")
	allowNoAuth := fs.Bool("allow-no-auth", false, "Allow binding to non-loopback hosts without a token (unsafe)")
	noOpen := fs.Bool("no-open", false, "Don't open browser")
	parseFlags(fs, args)

	cfg, err := config.Load()
	if err != nil {
		output.Warn(fmt.Errorf("%v (using defaults)", err))
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
			if *allowNoAuth {
				fmt.Fprintln(os.Stderr, "Warning: --allow-no-auth set; anyone who can reach this host can control the TV.")
			} else {
				output.FailMsg(fmt.Sprintf("refusing to bind to non-loopback host %q without a token. Set --token / config token, or pass --allow-no-auth to override (unsafe).", *host))
			}
		}
	}

	adbCfg := newAdbConfig(cfg)
	runner := &adb.RealRunner{}

	addr := fmt.Sprintf("%s:%d", *host, *port)
	server := serve.New(adbCfg, runner, serve.Options{Token: authToken, CORSOrigins: corsOrigins})

	// Bind before reporting readiness so --json never claims success for an
	// unavailable port.
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		output.FailWith(fmt.Sprintf("Error: cannot listen on %s: %v", addr, err), err)
	}
	actualPort := ln.Addr().(*net.TCPAddr).Port
	url := displayURL(*host, actualPort)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if output.JSON {
		output.Emit(output.ServeResult{URL: url, Host: *host, Port: actualPort})
	}
	if !*noOpen && !noInput && !output.JSON {
		openBrowser(url)
	}

	go func() {
		if err := server.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			output.FailWith(fmt.Sprintf("Server error: %v", err), err)
		}
	}()

	<-ctx.Done()
	if !output.JSON {
		fmt.Println("\nShutting down...")
	}
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

// displayURL returns the URL to print/open. Wildcard binds are reachable via
// localhost; a specific bind address is used as-is.
func displayURL(host string, port int) string {
	h := host
	if h == "" || h == "0.0.0.0" || h == "::" || h == "[::]" {
		h = "localhost"
	}
	return "http://" + net.JoinHostPort(h, strconv.Itoa(port))
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
