package mcp

import (
	"context"
	"testing"
	"time"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

func connectTestClient(t *testing.T, s *Server) *mcpsdk.ClientSession {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	serverTransport, clientTransport := mcpsdk.NewInMemoryTransports()
	if _, err := s.MCP().Connect(ctx, serverTransport, nil); err != nil {
		t.Fatalf("server connect: %v", err)
	}

	client := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "test-client", Version: "0.0.0"}, nil)
	cs, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	t.Cleanup(func() { cs.Close() })
	return cs
}

func TestMCP_RoundTripPressKey(t *testing.T) {
	m := connectedMock()
	s := newTestServer(m)
	cs := connectTestClient(t, s)

	res, err := cs.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name:      "tv_press_key",
		Arguments: map[string]any{"key": "down"},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if res.IsError {
		t.Fatalf("tool returned error: %+v", res.Content)
	}
	if m.LastShellCmd != "input keyevent KEYCODE_DPAD_DOWN" {
		t.Errorf("LastShellCmd = %q", m.LastShellCmd)
	}
}

func TestMCP_RoundTripInvalidArguments(t *testing.T) {
	s := newTestServer(connectedMock())
	cs := connectTestClient(t, s)

	// Missing required "package" field: the SDK validates against the input schema.
	res, err := cs.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name:      "tv_launch_app",
		Arguments: map[string]any{"package": "bad pkg"},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if !res.IsError {
		t.Fatal("expected IsError for invalid package")
	}
}

func TestMCP_ListsExpectedTools(t *testing.T) {
	s := newTestServer(connectedMock())
	cs := connectTestClient(t, s)

	res, err := cs.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}
	want := []string{
		"tv_status", "tv_connect", "tv_disconnect", "tv_list_apps",
		"tv_launch_app", "tv_kill_app", "tv_press_key", "tv_type_text",
		"tv_play", "tv_players", "tv_logs", "tv_screenshot",
	}
	got := map[string]bool{}
	for _, tool := range res.Tools {
		got[tool.Name] = true
	}
	for _, name := range want {
		if !got[name] {
			t.Errorf("missing tool %q", name)
		}
	}
}
