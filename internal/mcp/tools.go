package mcp

import (
	"context"
	"encoding/base64"
	"fmt"
	"sort"
	"strings"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/mutedfalcontv/tv/internal/adb"
	"github.com/mutedfalcontv/tv/internal/device"
	"github.com/mutedfalcontv/tv/internal/logcat"
	"github.com/mutedfalcontv/tv/internal/player"
	"github.com/mutedfalcontv/tv/internal/remote"
)

const (
	defaultLogLines = 100
	maxLogLines     = 1000
)

// ---- argument types ----

type statusIn struct{}

type connectIn struct{}

type listAppsIn struct{}

type packageIn struct {
	Package string `json:"package" jsonschema:"Android package name, e.g. com.netflix.ninja"`
}

type pressKeyIn struct {
	Key string `json:"key" jsonschema:"Remote key name, e.g. up, down, ok, home, back, play, volup"`
}

type typeTextIn struct {
	Text string `json:"text" jsonschema:"Text to type into the focused field"`
}

type playIn struct {
	URL    string `json:"url" jsonschema:"Media URL to play"`
	Player string `json:"player,omitempty" jsonschema:"Player name or package; defaults to the configured player"`
}

type playersIn struct{}

type logsIn struct {
	Level string `json:"level,omitempty" jsonschema:"Minimum log level: v, d, i, w, e, f"`
	Tag   string `json:"tag,omitempty" jsonschema:"Filter by log tag"`
	Lines int    `json:"lines,omitempty" jsonschema:"Number of recent lines (1-1000, default 100)"`
}

type screenshotIn struct {
	Path string `json:"path,omitempty" jsonschema:"Save the PNG to this path; when omitted, the image is returned as base64"`
}

// ---- output types ----

type statusOut struct {
	Connected     bool   `json:"connected"`
	TVIP          string `json:"tv_ip"`
	DefaultPlayer string `json:"default_player"`
}

type connectOut struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
	TVIP    string `json:"tv_ip"`
}

type listAppsOut struct {
	Count    int      `json:"count"`
	Packages []string `json:"packages"`
}

type packageActionOut struct {
	OK      bool   `json:"ok"`
	Action  string `json:"action"`
	Package string `json:"package"`
}

type pressKeyOut struct {
	OK  bool   `json:"ok"`
	Key string `json:"key"`
}

type typeTextOut struct {
	OK   bool   `json:"ok"`
	Text string `json:"text"`
}

type playOut struct {
	OK     bool   `json:"ok"`
	URL    string `json:"url"`
	Player string `json:"player"`
}

type playerInfo struct {
	Name    string `json:"name"`
	Package string `json:"package"`
}

type playersOut struct {
	Count   int          `json:"count"`
	Players []playerInfo `json:"players"`
}

type logsOut struct {
	Count     int      `json:"count"`
	Truncated bool     `json:"truncated"`
	Lines     []string `json:"lines"`
}

type screenshotOut struct {
	MimeType string `json:"mime_type"`
	Bytes    int    `json:"bytes"`
	Path     string `json:"path,omitempty"`
	Base64   string `json:"base64,omitempty"`
}

// ---- registration ----

func (s *Server) registerTools() {
	mcpsdk.AddTool(s.mcp, &mcpsdk.Tool{
		Name:        "tv_status",
		Description: "Report whether the TV is connected, its IP, and the configured default player.",
	}, s.handleStatus)

	mcpsdk.AddTool(s.mcp, &mcpsdk.Tool{
		Name:        "tv_connect",
		Description: "Connect to the TV over ADB using the configured IP.",
	}, s.handleConnect)

	mcpsdk.AddTool(s.mcp, &mcpsdk.Tool{
		Name:        "tv_disconnect",
		Description: "Disconnect the current ADB connection to the TV.",
	}, s.handleDisconnect)

	mcpsdk.AddTool(s.mcp, &mcpsdk.Tool{
		Name:        "tv_list_apps",
		Description: "List installed Android packages on the TV.",
	}, s.handleListApps)

	mcpsdk.AddTool(s.mcp, &mcpsdk.Tool{
		Name:        "tv_launch_app",
		Description: "Launch an app by Android package name.",
	}, s.handleLaunchApp)

	mcpsdk.AddTool(s.mcp, &mcpsdk.Tool{
		Name:        "tv_kill_app",
		Description: "Force-stop an app by Android package name.",
	}, s.handleKillApp)

	mcpsdk.AddTool(s.mcp, &mcpsdk.Tool{
		Name:        "tv_press_key",
		Description: "Send a remote key press to the TV.",
	}, s.handlePressKey)

	mcpsdk.AddTool(s.mcp, &mcpsdk.Tool{
		Name:        "tv_type_text",
		Description: "Type text into the focused input field on the TV.",
	}, s.handleTypeText)

	mcpsdk.AddTool(s.mcp, &mcpsdk.Tool{
		Name:        "tv_play",
		Description: "Open a media URL on the TV using a video player.",
	}, s.handlePlay)

	mcpsdk.AddTool(s.mcp, &mcpsdk.Tool{
		Name:        "tv_players",
		Description: "List video players detected on the TV.",
	}, s.handlePlayers)

	mcpsdk.AddTool(s.mcp, &mcpsdk.Tool{
		Name:        "tv_logs",
		Description: "Return a bounded snapshot of TV logs. Never streams.",
	}, s.handleLogs)

	mcpsdk.AddTool(s.mcp, &mcpsdk.Tool{
		Name:        "tv_screenshot",
		Description: "Capture a PNG screenshot of the TV, saving to a path or returning base64.",
	}, s.handleScreenshot)
}

// ---- handlers ----

func (s *Server) handleStatus(_ context.Context, _ *mcpsdk.CallToolRequest, _ statusIn) (*mcpsdk.CallToolResult, statusOut, error) {
	r := s.runnerWithTimeout()
	connected := adb.EnsureConnected(s.cfg, r) == nil
	return nil, statusOut{
		Connected:     connected,
		TVIP:          s.cfg.Serial(),
		DefaultPlayer: s.cfg.DefaultPlayer,
	}, nil
}

func (s *Server) handleConnect(_ context.Context, _ *mcpsdk.CallToolRequest, _ connectIn) (*mcpsdk.CallToolResult, connectOut, error) {
	r := s.runnerWithTimeout()
	msg, err := r.Connect(s.cfg.ConnectTarget())
	if err != nil {
		return nil, connectOut{}, fmt.Errorf("connect failed: %w", err)
	}
	return nil, connectOut{OK: true, Message: msg, TVIP: s.cfg.Serial()}, nil
}

func (s *Server) handleDisconnect(_ context.Context, _ *mcpsdk.CallToolRequest, _ connectIn) (*mcpsdk.CallToolResult, connectOut, error) {
	r := s.runnerWithTimeout()
	msg, err := r.Disconnect(s.cfg.ConnectTarget())
	if err != nil {
		return nil, connectOut{}, fmt.Errorf("disconnect failed: %w", err)
	}
	return nil, connectOut{OK: true, Message: msg, TVIP: s.cfg.Serial()}, nil
}

func (s *Server) handleListApps(_ context.Context, _ *mcpsdk.CallToolRequest, _ listAppsIn) (*mcpsdk.CallToolResult, listAppsOut, error) {
	r := s.runnerWithTimeout()
	if err := adb.EnsureConnected(s.cfg, r); err != nil {
		return nil, listAppsOut{}, err
	}
	out, err := r.Shell(s.cfg, "pm list packages")
	if err != nil {
		return nil, listAppsOut{}, err
	}
	pkgs := parsePackages(out)
	return nil, listAppsOut{Count: len(pkgs), Packages: pkgs}, nil
}

func (s *Server) handleLaunchApp(_ context.Context, _ *mcpsdk.CallToolRequest, in packageIn) (*mcpsdk.CallToolResult, packageActionOut, error) {
	if err := validatePackage(in.Package); err != nil {
		return nil, packageActionOut{}, err
	}
	r := s.runnerWithTimeout()
	if err := adb.EnsureConnected(s.cfg, r); err != nil {
		return nil, packageActionOut{}, err
	}
	if _, err := r.ShellWithStderr(s.cfg, "monkey -p "+in.Package+" -c android.intent.category.LAUNCHER 1"); err != nil {
		return nil, packageActionOut{}, fmt.Errorf("launch failed: %w", err)
	}
	return nil, packageActionOut{OK: true, Action: "launch", Package: in.Package}, nil
}

func (s *Server) handleKillApp(_ context.Context, _ *mcpsdk.CallToolRequest, in packageIn) (*mcpsdk.CallToolResult, packageActionOut, error) {
	if err := validatePackage(in.Package); err != nil {
		return nil, packageActionOut{}, err
	}
	r := s.runnerWithTimeout()
	if err := adb.EnsureConnected(s.cfg, r); err != nil {
		return nil, packageActionOut{}, err
	}
	if _, err := r.ShellWithStderr(s.cfg, "am force-stop "+in.Package); err != nil {
		return nil, packageActionOut{}, fmt.Errorf("kill failed: %w", err)
	}
	return nil, packageActionOut{OK: true, Action: "kill", Package: in.Package}, nil
}

func (s *Server) handlePressKey(_ context.Context, _ *mcpsdk.CallToolRequest, in pressKeyIn) (*mcpsdk.CallToolResult, pressKeyOut, error) {
	r := s.runnerWithTimeout()
	if err := adb.EnsureConnected(s.cfg, r); err != nil {
		return nil, pressKeyOut{}, err
	}
	if err := remote.Press(s.cfg, r, in.Key); err != nil {
		return nil, pressKeyOut{}, err
	}
	return nil, pressKeyOut{OK: true, Key: in.Key}, nil
}

func (s *Server) handleTypeText(_ context.Context, _ *mcpsdk.CallToolRequest, in typeTextIn) (*mcpsdk.CallToolResult, typeTextOut, error) {
	if in.Text == "" {
		return nil, typeTextOut{}, fmt.Errorf("text must not be empty")
	}
	r := s.runnerWithTimeout()
	if err := adb.EnsureConnected(s.cfg, r); err != nil {
		return nil, typeTextOut{}, err
	}
	if err := remote.Type(s.cfg, r, in.Text); err != nil {
		return nil, typeTextOut{}, err
	}
	return nil, typeTextOut{OK: true, Text: in.Text}, nil
}

func (s *Server) handlePlay(_ context.Context, _ *mcpsdk.CallToolRequest, in playIn) (*mcpsdk.CallToolResult, playOut, error) {
	if strings.TrimSpace(in.URL) == "" {
		return nil, playOut{}, fmt.Errorf("url must not be empty")
	}
	r := s.runnerWithTimeout()
	if err := adb.EnsureConnected(s.cfg, r); err != nil {
		return nil, playOut{}, err
	}
	pkg := in.Player
	if pkg == "" {
		pkg = s.cfg.DefaultPlayer
	}
	if pkg == "" {
		return nil, playOut{}, fmt.Errorf("no player specified and no default player configured")
	}
	if !strings.Contains(pkg, ".") {
		resolved, err := player.Resolve(s.cfg, r, pkg)
		if err != nil {
			return nil, playOut{}, err
		}
		pkg = resolved
	}
	if err := adb.ValidatePackage(pkg); err != nil {
		return nil, playOut{}, err
	}
	if err := player.PlayOnTV(s.cfg, r, pkg, in.URL); err != nil {
		return nil, playOut{}, err
	}
	return nil, playOut{OK: true, URL: in.URL, Player: pkg}, nil
}

func (s *Server) handlePlayers(_ context.Context, _ *mcpsdk.CallToolRequest, _ playersIn) (*mcpsdk.CallToolResult, playersOut, error) {
	r := s.runnerWithTimeout()
	detected, err := player.Detect(s.cfg, r)
	if err != nil {
		return nil, playersOut{}, err
	}
	out := playersOut{Count: len(detected), Players: []playerInfo{}}
	for _, p := range detected {
		out.Players = append(out.Players, playerInfo{Name: p.Name, Package: p.PackageName})
	}
	return nil, out, nil
}

func (s *Server) handleLogs(_ context.Context, _ *mcpsdk.CallToolRequest, in logsIn) (*mcpsdk.CallToolResult, logsOut, error) {
	lines, truncated := normalizeLines(in.Lines)
	r := s.runnerWithTimeout()
	if err := adb.EnsureConnected(s.cfg, r); err != nil {
		return nil, logsOut{}, err
	}
	opts := logcat.Options{
		Level: in.Level,
		Tag:   in.Tag,
		Lines: lines,
		Dump:  true,
	}
	raw, err := logcat.RunOnce(s.cfg, r, opts)
	if err != nil {
		return nil, logsOut{}, err
	}
	outLines := splitLogLines(raw)
	if len(outLines) > lines {
		truncated = true
		outLines = outLines[len(outLines)-lines:]
	}
	return nil, logsOut{Count: len(outLines), Truncated: truncated, Lines: outLines}, nil
}

func (s *Server) handleScreenshot(_ context.Context, _ *mcpsdk.CallToolRequest, in screenshotIn) (*mcpsdk.CallToolResult, screenshotOut, error) {
	r := s.runnerWithTimeout()
	if err := adb.EnsureConnected(s.cfg, r); err != nil {
		return nil, screenshotOut{}, err
	}
	data, err := device.CaptureScreenshot(s.cfg, r)
	if err != nil {
		return nil, screenshotOut{}, err
	}
	if len(data) == 0 {
		return nil, screenshotOut{}, fmt.Errorf("screencap returned no data")
	}
	out := screenshotOut{MimeType: "image/png", Bytes: len(data)}
	if in.Path != "" {
		if err := device.WriteScreenshot(in.Path, data); err != nil {
			return nil, screenshotOut{}, fmt.Errorf("writing screenshot: %w", err)
		}
		out.Path = in.Path
		return nil, out, nil
	}
	out.Base64 = base64.StdEncoding.EncodeToString(data)
	return nil, out, nil
}

// ---- helpers ----

func normalizeLines(n int) (int, bool) {
	if n <= 0 {
		return defaultLogLines, false
	}
	if n > maxLogLines {
		return maxLogLines, true
	}
	return n, false
}

func parsePackages(out string) []string {
	var pkgs []string
	for _, line := range strings.Split(out, "\n") {
		pkg, ok := strings.CutPrefix(strings.TrimSpace(line), "package:")
		if !ok {
			continue
		}
		if pkg = strings.TrimSpace(pkg); pkg != "" {
			pkgs = append(pkgs, pkg)
		}
	}
	sort.Strings(pkgs)
	return pkgs
}

func splitLogLines(raw string) []string {
	lines := []string{}
	for _, line := range strings.Split(raw, "\n") {
		if strings.TrimSpace(line) != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

func validatePackage(pkg string) error {
	if pkg == "" {
		return fmt.Errorf("package must not be empty")
	}
	for _, r := range pkg {
		switch {
		case r >= 'a' && r <= 'z',
			r >= 'A' && r <= 'Z',
			r >= '0' && r <= '9',
			r == '.', r == '_':
		default:
			return fmt.Errorf("invalid package name %q", pkg)
		}
	}
	return nil
}
