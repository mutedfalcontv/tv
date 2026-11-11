package logcat

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/mutedfalcontv/tv/internal/adb"
)

type Options struct {
	Package string
	Level   string
	Tag     string
	Lines   int
	Format  string
	Clear   bool
	Dump    bool
}

var levels = map[string]string{
	"v": "*:V", "d": "*:D", "i": "*:I", "w": "*:W", "e": "*:E", "f": "*:F",
	"verbose": "*:V", "debug": "*:D", "info": "*:I",
	"warning": "*:W", "error": "*:E", "fatal": "*:F",
}

func BuildArgs(opts Options) []string {
	var args []string

	if opts.Clear {
		args = append(args, "-b", "all", "-c")
		return args
	}

	if opts.Dump {
		args = append(args, "-d")
	}

	if opts.Lines > 0 {
		args = append(args, "-t", strconv.Itoa(opts.Lines))
	}

	if opts.Package != "" {
		args = append(args, "--pid="+opts.Package)
	}

	if opts.Tag != "" {
		args = append(args, "-s", opts.Tag+":*")
	}

	if opts.Level != "" {
		if l, ok := levels[strings.ToLower(opts.Level)]; ok {
			args = append(args, l)
		}
	}

	format := opts.Format
	if format == "" {
		format = "brief"
	}
	args = append(args, "-v", format)

	return args
}

// SafeField reports whether v is safe to pass to logcat as a tag or format
// value without shell interpretation.
func SafeField(v string) bool {
	for _, r := range v {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9',
			r == '_', r == '-', r == '.', r == ':':
		default:
			return false
		}
	}
	return true
}

// Validate rejects options that would be unsafe or invalid once RunOnce joins
// them into an `adb shell` command string.
func Validate(opts Options) error {
	if opts.Tag != "" && !SafeField(opts.Tag) {
		return fmt.Errorf("invalid log tag %q", opts.Tag)
	}
	if opts.Format != "" && !SafeField(opts.Format) {
		return fmt.Errorf("invalid log format %q", opts.Format)
	}
	if opts.Package != "" {
		if _, err := strconv.Atoi(opts.Package); err != nil {
			return fmt.Errorf("package filter must be a numeric PID, got %q", opts.Package)
		}
	}
	return nil
}

func ResolvePID(cfg *adb.Config, r adb.Runner, pkg string) (int, error) {
	if err := adb.ValidatePackage(pkg); err != nil {
		return 0, err
	}
	// pidof resolves the process by package name directly and avoids matching
	// the `grep` process that `ps -A | grep <pkg>` can return.
	if out, err := r.Shell(cfg, "pidof "+pkg); err == nil {
		for _, field := range strings.Fields(out) {
			if pid, err := strconv.Atoi(field); err == nil {
				return pid, nil
			}
		}
	}
	// Fallback for adb/Android versions without pidof. -w stops a package from
	// matching an unrelated process, and the final grep drops its own process.
	out, err := r.Shell(cfg, "ps -A | grep -w "+pkg+" | grep -v grep")
	if err != nil {
		return 0, fmt.Errorf("failed to list processes: %w", err)
	}
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			pid, err := strconv.Atoi(fields[1])
			if err == nil {
				return pid, nil
			}
		}
	}
	return 0, fmt.Errorf("process not found for package: %s", pkg)
}

func RunStream(cfg *adb.Config, opts Options) error {
	return RunStreamTo(cfg, opts, os.Stdout, os.Stderr)
}

// RunStreamTo streams logcat output to the supplied writers.
func RunStreamTo(cfg *adb.Config, opts Options, stdout, stderr io.Writer) error {
	adbPath, err := exec.LookPath("adb")
	if err != nil {
		return fmt.Errorf("ADB not found on PATH")
	}
	logArgs := BuildArgs(opts)
	cmd := exec.Command(adbPath, append([]string{"-s", cfg.Serial(), "logcat"}, logArgs...)...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}

// onceArgs returns the logcat arguments for a bounded, non-streaming read.
// It always sets Dump so the underlying `adb shell logcat` exits instead of
// streaming forever (RunOnce is used by one-shot callers such as GET /api/logs).
func onceArgs(opts Options) []string {
	opts.Dump = true
	return BuildArgs(opts)
}

func RunOnce(cfg *adb.Config, r adb.Runner, opts Options) (string, error) {
	if err := Validate(opts); err != nil {
		return "", err
	}
	logArgs := onceArgs(opts)
	cmd := "logcat " + strings.Join(logArgs, " ")
	return r.ShellWithStderr(cfg, cmd)
}
