package serve

import (
	"encoding/json"
	"testing"
)

func TestOpenAPIDocumentIsValid(t *testing.T) {
	var doc struct {
		OpenAPI string                     `json:"openapi"`
		Paths   map[string]json.RawMessage `json:"paths"`
	}
	if err := json.Unmarshal(openapiJSON, &doc); err != nil {
		t.Fatalf("openapi.json is not valid JSON: %v", err)
	}
	if doc.OpenAPI != "3.0.3" {
		t.Errorf("openapi = %q, want 3.0.3", doc.OpenAPI)
	}
}

func TestOpenAPIDocumentsRoutes(t *testing.T) {
	var doc struct {
		Paths map[string]json.RawMessage `json:"paths"`
	}
	if err := json.Unmarshal(openapiJSON, &doc); err != nil {
		t.Fatalf("openapi.json is not valid JSON: %v", err)
	}
	want := []string{
		"/api/status",
		"/api/status/keys",
		"/api/remote/press",
		"/api/remote/type",
		"/api/apps",
		"/api/apps/launch",
		"/api/apps/kill",
		"/api/players",
		"/api/player/default",
		"/api/play",
		"/api/logs",
		"/api/config",
		"/api/openapi.json",
		"/ws/logs",
	}
	for _, path := range want {
		if _, ok := doc.Paths[path]; !ok {
			t.Errorf("missing path in OpenAPI: %s", path)
		}
	}
}
