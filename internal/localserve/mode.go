package localserve

import "time"

// Mode is how a local media file is served to the TV.
//
//	ModeBackground  a detached child process serves for BackgroundTTL; the CLI exits
//	ModeForeground  this process serves for a fixed timeout, then exits
//	ModeForever     this process serves until interrupted
//
// ModeBackground is the only mode that outlives the CLI process. ModeForeground
// and ModeForever both block, so they differ only in how long they block.
type Mode int

const (
	// BackgroundTTL is how long a detached file server keeps serving.
	BackgroundTTL = time.Hour

	ModeBackground Mode = iota
	ModeForeground
	ModeForever
)

// ModeFor derives the serve mode from the CLI flags that select it.
func ModeFor(forever bool, timeout time.Duration) Mode {
	switch {
	case forever:
		return ModeForever
	case timeout > 0:
		return ModeForeground
	default:
		return ModeBackground
	}
}

// Detached reports whether the server must run in a separate process, because
// this command returns before serving finishes.
func (m Mode) Detached() bool { return m == ModeBackground }

// TTL is how long the server runs. Zero (in the foreground modes) means "until
// interrupted"; ModeBackground uses BackgroundTTL.
func (m Mode) TTL(timeout time.Duration) time.Duration {
	switch m {
	case ModeForeground:
		return timeout
	case ModeBackground:
		return BackgroundTTL
	default:
		return 0
	}
}

func (m Mode) String() string {
	switch m {
	case ModeBackground:
		return "background"
	case ModeForeground:
		return "foreground"
	case ModeForever:
		return "forever"
	default:
		return "unknown"
	}
}

// ParseMode maps a serve-mode word back to its typed value. It accepts the
// values emitted in PlayResult.ServeMode and in the detached child's --mode
// flag.
func ParseMode(raw string) Mode {
	switch raw {
	case "background":
		return ModeBackground
	case "foreground":
		return ModeForeground
	case "forever":
		return ModeForever
	default:
		return Mode(-1)
	}
}

// Valid reports whether m is one of the known modes.
func (m Mode) Valid() bool {
	switch m {
	case ModeBackground, ModeForeground, ModeForever:
		return true
	default:
		return false
	}
}
