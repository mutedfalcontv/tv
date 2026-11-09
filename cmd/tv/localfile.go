package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/mutedfalcontv/tv/internal/adb"
	"github.com/mutedfalcontv/tv/internal/localserve"
	"github.com/mutedfalcontv/tv/internal/output"
	"github.com/mutedfalcontv/tv/internal/player"
)

// defaultBackgroundServeTTL is how long a detached file server keeps running.
const defaultBackgroundServeTTL = time.Hour

func playLocal(cfg *adb.Config, r adb.Runner, playerPkg, file string, forever bool, timeout time.Duration) {
	if forever || timeout > 0 {
		playLocalForeground(cfg, r, playerPkg, file, forever, timeout)
		return
	}

	serveURL, err := spawnFileServer(file, cfg.TVIP, defaultBackgroundServeTTL)
	if err != nil {
		output.FailWith(fmt.Sprintf("Error starting file server: %v", err), err)
	}
	if !output.JSON {
		fmt.Printf("Serving %s at %s\n", file, serveURL)
		fmt.Printf("Playing on %s via %s...\n", cfg.TVIP, playerPkg)
	}
	if err := player.PlayOnTV(cfg, r, playerPkg, serveURL); err != nil {
		output.FailWith(fmt.Sprintf("Error: %v", err), err)
	}
	output.Result(output.PlayResult{
		URL:        serveURL,
		Player:     playerPkg,
		TVIP:       cfg.TVIP,
		Played:     true,
		LocalFile:  file,
		ServeURL:   serveURL,
		ServeMode:  "background",
		TTLSeconds: int(defaultBackgroundServeTTL.Seconds()),
	}, func() {
		fmt.Println("Done.")
	})
}

func playLocalForeground(cfg *adb.Config, r adb.Runner, playerPkg, file string, forever bool, timeout time.Duration) {
	srv, err := localserve.Start(file, cfg.TVIP, 0)
	if err != nil {
		output.FailWith(fmt.Sprintf("Error starting file server: %v", err), err)
	}
	if !output.JSON {
		fmt.Printf("Serving %s at %s\n", file, srv.URL)
		fmt.Printf("Playing on %s via %s...\n", cfg.TVIP, playerPkg)
	}
	if err := player.PlayOnTV(cfg, r, playerPkg, srv.URL); err != nil {
		shutdownServer(srv)
		output.FailWith(fmt.Sprintf("Error: %v", err), err)
	}

	mode := "foreground"
	ttl := 0
	if forever {
		mode = "forever"
	} else {
		ttl = int(timeout.Seconds())
	}
	output.Result(output.PlayResult{
		URL:        srv.URL,
		Player:     playerPkg,
		TVIP:       cfg.TVIP,
		Played:     true,
		LocalFile:  file,
		ServeURL:   srv.URL,
		ServeMode:  mode,
		TTLSeconds: ttl,
	}, func() {
		if forever {
			fmt.Println("Serving forever (Ctrl-C to stop)...")
		} else {
			fmt.Printf("Serving for %s (Ctrl-C to stop)...\n", timeout)
		}
	})

	if forever {
		waitForShutdown(0)
	} else {
		waitForShutdown(timeout)
	}
	shutdownServer(srv)
	if !output.JSON {
		fmt.Println("Done.")
	}
}

func waitForShutdown(timeout time.Duration) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if timeout <= 0 {
		<-ctx.Done()
		return
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-ctx.Done():
	case <-timer.C:
	}
}

func shutdownServer(s *localserve.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = s.Shutdown(ctx)
}

// spawnFileServer starts a detached copy of this binary that serves file, then
// returns the URL once the child has bound its listener.
func spawnFileServer(file, tvIP string, ttl time.Duration) (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	cmd := exec.Command(exe, "__serve-file",
		"--file", file,
		"--tv-ip", tvIP,
		"--ttl", ttl.String(),
	)
	cmd.Stdin = nil
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return "", err
	}
	scanner := bufio.NewScanner(stdout)
	if !scanner.Scan() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		msg := stderr.String()
		if msg == "" {
			return "", fmt.Errorf("background server exited before reporting a URL")
		}
		return "", fmt.Errorf("background server failed: %s", msg)
	}
	serveURL := scanner.Text()
	_ = stdout.Close()
	if err := cmd.Process.Release(); err != nil {
		return "", err
	}
	return serveURL, nil
}

// serveFileChildCmd is the internal detached-server subcommand.
func serveFileChildCmd(args []string) {
	fs := newFlagSet("__serve-file")
	file := fs.String("file", "", "File to serve")
	tvIP := fs.String("tv-ip", "", "TV address used to choose the advertised interface")
	ttl := fs.Duration("ttl", defaultBackgroundServeTTL, "How long to keep serving")
	parseFlags(fs, args)

	if *file == "" {
		fmt.Fprintln(os.Stderr, "missing --file")
		os.Exit(1)
	}
	srv, err := localserve.Start(*file, *tvIP, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "serve-file: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(srv.URL)

	waitForShutdown(*ttl)
	shutdownServer(srv)
}
