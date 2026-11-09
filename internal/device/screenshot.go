package device

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mutedfalcontv/tv/internal/adb"
)

// PNG signature: 89 50 4E 47 0D 0A 1A 0A.
var pngSignature = []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}

// DefaultScreenshotName returns the default output path for a capture taken at
// the current time, e.g. "screencap-1700000000.png".
func DefaultScreenshotName() string {
	return fmt.Sprintf("screencap-%d.png", time.Now().Unix())
}

// CaptureScreenshot grabs the current screen from the TV as PNG bytes. It uses
// "adb exec-out screencap -p", which bypasses the device PTY so binary data is
// not mangled by shell/terminal newline translation (the classic Windows
// corruption where every \n becomes \r\n).
func CaptureScreenshot(cfg *adb.Config, r adb.Runner) ([]byte, error) {
	raw, err := r.ExecOut(cfg, "screencap", "-p")
	if err != nil {
		return nil, fmt.Errorf("failed to capture screen: %w", err)
	}
	return NormalizePNG(raw)
}

// NormalizePNG validates that data is a PNG and repairs CRLF corruption if the
// capture still went through a text-mode pipe. A translated payload has an extra
// \r inserted before every \n, including inside the signature. When corruption
// is detected, each \r immediately preceding \n is stripped to undo the
// insertion without touching genuine binary bytes.
func NormalizePNG(data []byte) ([]byte, error) {
	if bytes.HasPrefix(data, pngSignature) {
		return data, nil
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("empty screenshot data")
	}
	fixed := make([]byte, 0, len(data))
	for i := 0; i < len(data); i++ {
		if data[i] == '\r' && i+1 < len(data) && data[i+1] == '\n' {
			continue
		}
		fixed = append(fixed, data[i])
	}
	if bytes.HasPrefix(fixed, pngSignature) {
		return fixed, nil
	}
	head := data
	if len(head) > 8 {
		head = head[:8]
	}
	return nil, fmt.Errorf("invalid PNG data: got %x, want %x", head, pngSignature)
}

// WriteScreenshot writes PNG data to path, creating parent directories.
func WriteScreenshot(path string, data []byte) error {
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return os.WriteFile(path, data, 0644)
}
