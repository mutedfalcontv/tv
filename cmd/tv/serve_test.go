package main

import "testing"

func TestDisplayURL(t *testing.T) {
	tests := []struct {
		host string
		port int
		want string
	}{
		{"127.0.0.1", 8080, "http://127.0.0.1:8080"},
		{"localhost", 8080, "http://localhost:8080"},
		{"0.0.0.0", 8080, "http://localhost:8080"},
		{"::", 8080, "http://localhost:8080"},
		{"192.168.2.3", 9000, "http://192.168.2.3:9000"},
		{"::1", 8080, "http://[::1]:8080"},
	}
	for _, tt := range tests {
		if got := displayURL(tt.host, tt.port); got != tt.want {
			t.Errorf("displayURL(%q, %d) = %q, want %q", tt.host, tt.port, got, tt.want)
		}
	}
}
