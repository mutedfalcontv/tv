package logcat

import (
	"fmt"
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

func ResolvePID(cfg *adb.Config, r adb.Runner, pkg string) (int, error) {
	out, err := r.Shell(cfg, "ps -A | grep "+pkg)
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
	adbPath, err := exec.LookPath("adb")
	if err != nil {
		return fmt.Errorf("ADB not found on PATH")
	}
	logArgs := BuildArgs(opts)
	cmd := exec.Command(adbPath, append([]string{"-s", cfg.TVIP, "logcat"}, logArgs...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func RunOnce(cfg *adb.Config, r adb.Runner, opts Options) (string, error) {
	logArgs := BuildArgs(opts)
	cmd := "logcat " + strings.Join(logArgs, " ")
	return r.ShellWithStderr(cfg, cmd)
}
