package device

import (
	"bytes"
	"testing"

	"github.com/mutedfalcontv/tv/internal/adb"
)

func validPNG(extra ...byte) []byte {
	data := make([]byte, 0, len(pngSignature)+len(extra))
	data = append(data, pngSignature...)
	return append(data, extra...)
}

// crlfCorrupt simulates text-mode translation inserting \r before every \n.
func crlfCorrupt(data []byte) []byte {
	out := make([]byte, 0, len(data))
	for _, b := range data {
		if b == '\n' {
			out = append(out, '\r')
		}
		out = append(out, b)
	}
	return out
}

func TestNormalizePNG(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		want    []byte
		wantErr bool
	}{
		{
			name: "valid signature unchanged",
			data: validPNG(0x00, 0x0d, 0x0a, 0xff),
			want: validPNG(0x00, 0x0d, 0x0a, 0xff),
		},
		{
			name: "repairs crlf corrupted signature",
			data: crlfCorrupt(validPNG(0x00, 0x0d, 0x0a, 0xff)),
			want: validPNG(0x00, 0x0d, 0x0a, 0xff),
		},
		{
			name:    "empty payload",
			data:    nil,
			wantErr: true,
		},
		{
			name:    "not a png",
			data:    []byte("this is not a png"),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizePNG(tt.data)
			if tt.wantErr {
				if err == nil {
					t.Fatal("NormalizePNG() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("NormalizePNG() unexpected error: %v", err)
			}
			if !bytes.Equal(got, tt.want) {
				t.Errorf("NormalizePNG() = %x, want %x", got, tt.want)
			}
		})
	}
}

func TestCaptureScreenshot(t *testing.T) {
	want := validPNG(0x01, 0x02, 0x03)
	m := &adb.MockRunner{ExecOutOut: want}
	cfg := &adb.Config{TVIP: "192.168.2.3:5555"}

	got, err := CaptureScreenshot(cfg, m)
	if err != nil {
		t.Fatalf("CaptureScreenshot() unexpected error: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("CaptureScreenshot() = %x, want %x", got, want)
	}
	if len(m.LastExecOutArgs) != 2 || m.LastExecOutArgs[0] != "screencap" || m.LastExecOutArgs[1] != "-p" {
		t.Errorf("ExecOut args = %v, want [screencap -p]", m.LastExecOutArgs)
	}
}

func TestCaptureScreenshot_ExecOutError(t *testing.T) {
	m := &adb.MockRunner{ExecOutErr: errBoom{}}
	cfg := &adb.Config{TVIP: "192.168.2.3:5555"}

	if _, err := CaptureScreenshot(cfg, m); err == nil {
		t.Fatal("CaptureScreenshot() expected error, got nil")
	}
}

func TestDefaultScreenshotName(t *testing.T) {
	name := DefaultScreenshotName()
	if len(name) < len("screencap-.png") || name[:10] != "screencap-" {
		t.Errorf("DefaultScreenshotName() = %q, want screencap-<ts>.png", name)
	}
	if name[len(name)-4:] != ".png" {
		t.Errorf("DefaultScreenshotName() = %q, want .png suffix", name)
	}
}
