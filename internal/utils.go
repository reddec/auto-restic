package internal

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

func isAlreadyInitialized(err string) bool {
	return contains(err, "already initialized", "already exists")
}

func isNoSnapshot(err string) bool {
	return contains(err, "no snapshot found")
}

type binary string

func (b binary) exec(global context.Context, args ...string) error {
	return b.invoke(global, nil, args...)
}

func (b binary) invoke(global context.Context, out io.Writer, args ...string) error {
	if out == nil {
		out = os.Stderr
	}

	ctx, cancel := context.WithCancel(global)
	defer cancel()

	cmd := exec.CommandContext(ctx, string(b), args...)
	cmd.Stdout = out
	cmd.Stderr = out
	cmd.Env = os.Environ()
	SetFlags(cmd)
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("start %s: %w", b, err)
	}

	go func() {
		<-ctx.Done()
		cmd.Process.Kill()
		Reap(cmd.Process.Pid)
	}()

	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("%s %s: %w", b, strings.Join(args, " "), err)
	}
	return nil
}

func newLimitedBuffer(limit int, out io.Writer) *limitedTeeBuffer {
	return &limitedTeeBuffer{
		limit:   limit,
		wrapped: out,
	}
}

type limitedTeeBuffer struct {
	buffer  bytes.Buffer
	limit   int
	wrapped io.Writer
}

func (ltb *limitedTeeBuffer) Write(p []byte) (n int, err error) {
	left := ltb.limit - ltb.buffer.Len()
	if left > 0 {
		_, _ = ltb.buffer.Write(p[:min(len(p), left)])
	}
	return ltb.wrapped.Write(p)
}

func min(a, b int) int {
	if b < a {
		return b
	}
	return a
}

func contains(value string, options ...string) bool {
	for _, opt := range options {
		if strings.Contains(value, opt) {
			return true
		}
	}
	return false
}
