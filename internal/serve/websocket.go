package serve

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"time"

	"github.com/mutedfalcontv/tv/internal/logcat"

	gorilla "github.com/gorilla/websocket"
)

var upgrader = gorilla.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

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

func (s *Server) handleWSLogs(w http.ResponseWriter, r *http.Request) {
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

	var opts logcat.Options
	opts.Dump = false

	go func() {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			var f filterMsg
			if json.Unmarshal(msg, &f) == nil && f.Type == "filter" {
				if f.Level != "" {
					opts.Level = f.Level
				}
				if f.Tag != "" {
					opts.Tag = f.Tag
				}
				if f.PID > 0 {
					opts.Package = fmt.Sprintf("%d", f.PID)
				}
			}
		}
	}()

	logArgs := logcat.BuildArgs(opts)
	cmd := exec.Command(adbPath, append([]string{"-s", s.config.TVIP, "logcat"}, logArgs...)...)
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
	defer cmd.Wait()

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		frame := logFrame{Line: line, Timestamp: time.Now().Format(time.RFC3339)}
		if err := conn.WriteJSON(frame); err != nil {
			return
		}
	}
}
