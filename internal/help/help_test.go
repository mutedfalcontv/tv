package help

import (
	"encoding/json"
	"testing"
)

func TestJSON_IsValidDocument(t *testing.T) {
	data, err := JSON()
	if err != nil {
		t.Fatalf("JSON() unexpected error: %v", err)
	}
	var doc Document
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("JSON() produced invalid JSON: %v", err)
	}
	if doc.Name != "tv" {
		t.Errorf("Name = %q, want tv", doc.Name)
	}
	if len(doc.Commands) == 0 {
		t.Fatal("Commands is empty")
	}
}

func TestTree_ContainsAllCommands(t *testing.T) {
	doc := Tree()
	got := make(map[string]bool)
	for _, c := range doc.Commands {
		got[c.Name] = true
	}
	for _, want := range []string{
		"adb", "app", "config", "devices", "discover", "help", "install",
		"logs", "mcp", "pair", "play", "player", "remote", "screenshot",
		"serve", "state",
	} {
		if !got[want] {
			t.Errorf("missing command: %s", want)
		}
	}
}

func TestTree_RemoteHasKeyChoices(t *testing.T) {
	doc := Tree()
	for _, c := range doc.Commands {
		if c.Name != "remote" {
			continue
		}
		if len(c.Arguments) == 0 || len(c.Arguments[0].Choices) == 0 {
			t.Fatal("remote key argument has no choices")
		}
		return
	}
	t.Fatal("remote command not found")
}
