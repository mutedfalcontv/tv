package serve

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"github.com/mutedfalcontv/tv/internal/logcat"

	gorilla "github.com/gorilla/websocket"
)

type logFrame struct {
	Line      string `json:"line"`
	Timestamp string `json:"timestamp"`
}

type filterMsg struct {
	Type  string `json:"type"`
	Level string `json:"level,omitempty"`
	Tag   string `json:"tag,omitempty"`
	PID   int    `json:"pid,omitempty"`
}

// handleWSLogs streams logcat over a WebSocket.
//
// Filter messages ({type:"filter",level,tag,pid}) mutate the connection's
// logcat options under a mutex. Because `adb logcat` arguments cannot change
// for a running process, an incoming filter restarts logcat with the new
// arguments instead of mutating args that were already used to start it. The
// previous implementation built args once before the reader goroutine ran, so
// filters never applied and the read was a data race.
func (s *Server) handleWSLogs(w http.ResponseWriter, r *http.Request) {
	upgrader := gorilla.Upgrader{CheckOrigin: s.checkOrigin}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WS upgrade error: %v", err)
		return
	}
	defer conn.Close()

	adbPath, err := exec.LookPath("adb")
	if err != nil {
		conn.WriteJSON(logFrame{Line: "ADB not found on PATH", Timestamp: time.Now().Format(time.RFC3339)})
		return
	}

	var mu sync.Mutex
	opts := initialLogOptions(r)

	restart := make(chan struct{}, 1)
	clientClosed := make(chan struct{})

	go func() {
		defer close(clientClosed)
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			var f filterMsg
			if json.Unmarshal(msg, &f) != nil || f.Type != "filter" {
				continue
			}
			mu.Lock()
			if f.Level != "" {
				opts.Level = f.Level
			}
			if f.Tag != "" {
				opts.Tag = f.Tag
			}
			if f.PID > 0 {
				opts.Package = strconv.Itoa(f.PID)
			}
			mu.Unlock()
			select {
			case restart <- struct{}{}:
			default:
			}
		}
	}()

	for {
		mu.Lock()
		cur := opts
		mu.Unlock()

		logArgs := logcat.BuildArgs(cur)
		cmd := exec.Command(adbPath, append([]string{"-s", s.config.Serial(), "logcat"}, logArgs...)...)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			conn.WriteJSON(logFrame{Line: fmt.Sprintf("Error: %v", err), Timestamp: time.Now().Format(time.RFC3339)})
			return
		}
		cmd.Stderr = cmd.Stdout

		if err := cmd.Start(); err != nil {
			conn.WriteJSON(logFrame{Line: fmt.Sprintf("Error: %v", err), Timestamp: time.Now().Format(time.RFC3339)})
			return
		}

		lines := make(chan string)
		stopScan := make(chan struct{})
		go func() {
			defer close(lines)
			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				select {
				case lines <- scanner.Text():
				case <-stopScan:
					return
				}
			}
		}()

		restarted := false
		for !restarted {
			select {
			case line, ok := <-lines:
				if !ok {
					// logcat exited on its own (adb error/disconnect); stop.
					close(stopScan)
					cmd.Wait()
					return
				}
				frame := logFrame{Line: line, Timestamp: time.Now().Format(time.RFC3339)}
				if err := conn.WriteJSON(frame); err != nil {
					close(stopScan)
					cmd.Process.Kill()
					cmd.Wait()
					return
				}
			case <-restart:
				restarted = true
			case <-clientClosed:
				close(stopScan)
				cmd.Process.Kill()
				cmd.Wait()
				return
			}
		}

		close(stopScan)
		cmd.Process.Kill()
		cmd.Wait()
	}
}

// initialLogOptions reads filters supplied as query parameters at connection
// time, matching the filter message fields.
func initialLogOptions(r *http.Request) logcat.Options {
	q := r.URL.Query()
	var opts logcat.Options
	opts.Level = q.Get("level")
	opts.Tag = q.Get("tag")
	if pid := q.Get("pid"); pid != "" {
		if _, err := strconv.Atoi(pid); err == nil {
			opts.Package = pid
		}
	}
	return opts
}
