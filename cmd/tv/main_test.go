package main

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/mutedfalcontv/tv/internal/player"
)

func TestParseGlobalFlags(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantArgs    []string
		wantJSON    bool
		wantNoInput bool
	}{
		{
			name:     "no globals",
			args:     []string{"app", "list"},
			wantArgs: []string{"app", "list"},
		},
		{
			name:     "json before command",
			args:     []string{"--json", "app", "list"},
			wantArgs: []string{"app", "list"},
			wantJSON: true,
		},
		{
			name:     "json after command",
			args:     []string{"app", "list", "--json"},
			wantArgs: []string{"app", "list"},
			wantJSON: true,
		},
		{
			name:     "json equals form",
			args:     []string{"--json=true", "app", "list"},
			wantArgs: []string{"app", "list"},
			wantJSON: true,
		},
		{
			name:     "json explicitly disabled",
			args:     []string{"--json=false", "app", "list"},
			wantArgs: []string{"app", "list"},
		},
		{
			name:        "no-input",
			args:        []string{"--no-input", "play", "https://example.com/v.mp4"},
			wantArgs:    []string{"play", "https://example.com/v.mp4"},
			wantNoInput: true,
		},
		{
			name:        "both flags",
			args:        []string{"-json", "play", "--no-input", "x"},
			wantArgs:    []string{"play", "x"},
			wantJSON:    true,
			wantNoInput: true,
		},
		{
			name:        "single dash forms",
			args:        []string{"-no-input=true", "logs", "-d", "-json=true"},
			wantArgs:    []string{"logs", "-d"},
			wantJSON:    true,
			wantNoInput: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotArgs, gotOpts := parseGlobalFlags(tt.args)
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("args = %v, want %v", gotArgs, tt.wantArgs)
			}
			if gotOpts.JSON != tt.wantJSON {
				t.Errorf("JSON = %v, want %v", gotOpts.JSON, tt.wantJSON)
			}
			if gotOpts.NoInput != tt.wantNoInput {
				t.Errorf("NoInput = %v, want %v", gotOpts.NoInput, tt.wantNoInput)
			}
		})
	}
}

func TestParsePackages(t *testing.T) {
	tests := []struct {
		name string
		out  string
		want []string
	}{
		{
			name: "typical output",
			out:  "package:org.videolan.vlc\npackage:com.brouken.player\n",
			want: []string{"com.brouken.player", "org.videolan.vlc"},
		},
		{
			name: "empty output",
			out:  "",
			want: []string{},
		},
		{
			name: "ignores noise",
			out:  "warning: something\npackage:com.example.app\n",
			want: []string{"com.example.app"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parsePackages(tt.out)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parsePackages() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSplitLines(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []string
	}{
		{"multiple", "a\nb\n", []string{"a", "b"}},
		{"no trailing newline", "a\nb", []string{"a", "b"}},
		{"empty", "", []string{}},
		{"single", "one\n", []string{"one"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitLines(tt.in)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("splitLines() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsTerminal_PipeIsNotTerminal(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Pipe() error: %v", err)
	}
	defer r.Close()
	defer w.Close()

	if isTerminal(r) {
		t.Error("isTerminal(pipe) = true, want false")
	}
}

func TestNoPlayerMessage(t *testing.T) {
	tests := []struct {
		name    string
		players []player.Info
		want    []string
	}{
		{
			name: "lists available players",
			players: []player.Info{
				{Name: "VLC", PackageName: "org.videolan.vlc"},
				{Name: "Just Player", PackageName: "com.brouken.player"},
			},
			want: []string{
				"No player specified and no default player is set.",
				"Available players:",
				"VLC                  org.videolan.vlc",
				"Just Player          com.brouken.player",
				"tv player default",
			},
		},
		{
			name:    "no players found",
			players: nil,
			want: []string{
				"No player specified and no default player is set.",
				"No video players were found on the TV.",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := noPlayerMessage(tt.players)
			for _, want := range tt.want {
				if !strings.Contains(got, want) {
					t.Errorf("message missing %q:\n%s", want, got)
				}
			}
		})
	}
}
