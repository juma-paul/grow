package executor

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

// Snapshot is a single observation of a CPython list's internal state.
type Snapshot struct {
	Tag  string `json:"tag"`
	Len  int    `json:"len"`
	Cap  int    `json:"cap"`
	Addr uint64 `json:"addr"`
}

// RunCPython executes a Python script in a subprocess and collects
// the JSON-line snapshots it prints to stdout. The script must use
// the _snap() helper (injected by the wrapper template) to emit
// snapshots. Returns on completion, error, or timeout.
func RunCPython(script string, timeout time.Duration) ([]Snapshot, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "python3", "-u", "-c", script)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("stdout pipe: %w", err)
	}

	var stderrBuf []byte
	cmd.Stderr = nil // captured below

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start python3: %w", err)
	}

	var snapshots []Snapshot
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 0, 64*1024), 256*1024)

	for scanner.Scan() {
		var snap Snapshot
		if err := json.Unmarshal(scanner.Bytes(), &snap); err != nil {
			continue
		}
		snapshots = append(snapshots, snap)
	}
	if scanErr := scanner.Err(); scanErr != nil {
		return nil, fmt.Errorf("reading stdout: %w", scanErr)
	}

	stderrScanner := bufio.NewScanner(stderrPipe)
	for stderrScanner.Scan() {
		stderrBuf = append(stderrBuf, stderrScanner.Bytes()...)
		stderrBuf = append(stderrBuf, '\n')
	}
	_ = stderrScanner.Err()

	err = cmd.Wait()
	if ctx.Err() == context.DeadlineExceeded {
		return snapshots, fmt.Errorf("python3 timed out after %s", timeout)
	}
	if err != nil {
		msg := string(stderrBuf)
		if msg == "" {
			msg = err.Error()
		}
		return snapshots, fmt.Errorf("python3 error: %s", msg)
	}

	return snapshots, nil
}
