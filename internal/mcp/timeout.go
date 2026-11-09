package mcp

import (
	"fmt"
	"time"

	"github.com/mutedfalcontv/tv/internal/adb"
)

// timeoutRunner bounds every ADB call so a hung device cannot stall a tool.
type timeoutRunner struct {
	inner   adb.Runner
	timeout time.Duration
}

type callResult[T any] struct {
	v   T
	err error
}

func callWithTimeout[T any](d time.Duration, fn func() (T, error)) (T, error) {
	ch := make(chan callResult[T], 1)
	go func() {
		v, err := fn()
		ch <- callResult[T]{v: v, err: err}
	}()
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case r := <-ch:
		return r.v, r.err
	case <-timer.C:
		var zero T
		return zero, fmt.Errorf("adb call timed out after %s", d)
	}
}

func (t *timeoutRunner) Devices() (string, error) {
	return callWithTimeout(t.timeout, t.inner.Devices)
}

func (t *timeoutRunner) Connect(ip string) (string, error) {
	return callWithTimeout(t.timeout, func() (string, error) { return t.inner.Connect(ip) })
}

func (t *timeoutRunner) Disconnect(ip string) (string, error) {
	return callWithTimeout(t.timeout, func() (string, error) { return t.inner.Disconnect(ip) })
}

func (t *timeoutRunner) Shell(cfg *adb.Config, cmd string) (string, error) {
	return callWithTimeout(t.timeout, func() (string, error) { return t.inner.Shell(cfg, cmd) })
}

func (t *timeoutRunner) ShellWithStderr(cfg *adb.Config, cmd string) (string, error) {
	return callWithTimeout(t.timeout, func() (string, error) { return t.inner.ShellWithStderr(cfg, cmd) })
}

func (t *timeoutRunner) ExecOut(cfg *adb.Config, args ...string) ([]byte, error) {
	return callWithTimeout(t.timeout, func() ([]byte, error) { return t.inner.ExecOut(cfg, args...) })
}

func (t *timeoutRunner) MDNSServices() (string, error) {
	return callWithTimeout(t.timeout, t.inner.MDNSServices)
}

func (t *timeoutRunner) Install(cfg *adb.Config, apkPath string) (string, error) {
	return callWithTimeout(t.timeout, func() (string, error) { return t.inner.Install(cfg, apkPath) })
}

func (t *timeoutRunner) Pair(addr, code string) (string, error) {
	return callWithTimeout(t.timeout, func() (string, error) { return t.inner.Pair(addr, code) })
}
