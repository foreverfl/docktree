package client

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"
)

const (
	spawnReadyTimeout = 5 * time.Second
	spawnPollInterval = 100 * time.Millisecond
)

// Spawn fork-execs `selfPath daemon-run` as a detached session, writes
// its pid to pidpath, redirects stdout/stderr to logpath, and polls the
// socket until the daemon answers Ping. Returns the child pid on success.
func Spawn(selfPath, sockpath, pidpath, logpath string) (int, error) {
	logFile, err := os.OpenFile(logpath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return 0, fmt.Errorf("open log: %w", err)
	}

	c := exec.Command(selfPath, "daemon-run")
	c.Stdout = logFile
	c.Stderr = logFile
	// Setsid detaches the daemon into its own session so it survives the
	// parent shell.
	c.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	if err := c.Start(); err != nil {
		_ = logFile.Close()
		return 0, fmt.Errorf("start daemon: %w", err)
	}
	// Parent owns the fd until Start succeeds; once handed off, the child
	// keeps it open via dup.
	_ = logFile.Close()

	pid := c.Process.Pid
	if err := os.WriteFile(pidpath, []byte(strconv.Itoa(pid)), 0o644); err != nil {
		return pid, fmt.Errorf("write pid: %w", err)
	}
	// Release so the OS reaps the child instead of us.
	if err := c.Process.Release(); err != nil {
		return pid, fmt.Errorf("release: %w", err)
	}

	deadline := time.Now().Add(spawnReadyTimeout)
	for time.Now().Before(deadline) {
		if err := Ping(sockpath); err == nil {
			return pid, nil
		} else if !errors.Is(err, ErrNotRunning) {
			return pid, err
		}
		time.Sleep(spawnPollInterval)
	}
	return pid, fmt.Errorf("daemon failed to become ready within %s. see %s", spawnReadyTimeout, logpath)
}
