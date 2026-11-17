package player

import "testing"

func TestClassifyTarget(t *testing.T) {
	tests := []struct {
		name     string
		arg      string
		isFile   bool
		statOK   bool
		wantKind Kind
	}{
		{"existing file", "/media/movie.mkv", true, true, KindLocalFile},
		{"existing directory", "/media", false, true, KindURL},
		{"missing path is a url", "https://example.com/a.m3u8", false, false, KindURL},
		{"missing local path falls through", "/nope/file.mkv", false, false, KindURL},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stat := func(string) (bool, bool) { return tt.isFile, tt.statOK }
			got := ClassifyTarget(tt.arg, stat)
			if got.Kind != tt.wantKind {
				t.Errorf("Kind = %v, want %v", got.Kind, tt.wantKind)
			}
			if tt.wantKind == KindLocalFile && got.Path != tt.arg {
				t.Errorf("Path = %q, want %q", got.Path, tt.arg)
			}
			if tt.wantKind == KindURL && got.URL != tt.arg {
				t.Errorf("URL = %q, want %q", got.URL, tt.arg)
			}
		})
	}
}

func TestKindIsLocal(t *testing.T) {
	if !KindLocalFile.IsLocal() {
		t.Error("KindLocalFile.IsLocal() = false, want true")
	}
	if KindURL.IsLocal() {
		t.Error("KindURL.IsLocal() = true, want false")
	}
}

func TestKindString(t *testing.T) {
	tests := map[Kind]string{
		KindURL:       "url",
		KindLocalFile: "local-file",
		Kind(99):      "unknown",
	}
	for kind, want := range tests {
		if got := kind.String(); got != want {
			t.Errorf("Kind(%d).String() = %q, want %q", kind, got, want)
		}
	}
}

func TestLooksLikeURL(t *testing.T) {
	if !LooksLikeURL("https://example.com/v.m3u8") {
		t.Error("LooksLikeURL(https://...) = false, want true")
	}
	if LooksLikeURL("/media/movie.mkv") {
		t.Error("LooksLikeURL(file) = true, want false")
	}
}
