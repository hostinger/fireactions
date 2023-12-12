package runner

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
)

const (
	// DefaultRunnerDir is the default directory where the GitHub runner is
	// installed.
	DefaultRunnerDir = "/opt/runner"
)

// Runner represents the actual GitHub runner.
type Runner struct {
	config   string
	runCmd   *exec.Cmd
	exitCh   chan struct{}
	fatalErr error
	stdout   io.Writer
	stderr   io.Writer
}

// Opt is a functional option for Runner.
type Opt func(r *Runner)

// WithStdout sets the writer to which the GitHub runner writes its stdout.
func WithStdout(stdout io.Writer) Opt {
	f := func(r *Runner) {
		r.stdout = stdout
	}

	return f
}

// WithStderr sets the writer to which the GitHub runner writes its stderr.
func WithStderr(stderr io.Writer) Opt {
	f := func(r *Runner) {
		r.stderr = stderr
	}

	return f
}

// New creates a new Runner.
func New(config string, opts ...Opt) *Runner {
	r := &Runner{
		config:   config,
		runCmd:   nil,
		stdout:   os.Stdout,
		stderr:   os.Stderr,
		fatalErr: nil,
		exitCh:   make(chan struct{}),
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// Start starts the GitHub runner. This requires the GitHub runner to be configured first.
// If the GitHub runner is already running, this is a no-op.
func (r *Runner) Start(ctx context.Context) error {
	if r.IsRunning() {
		return nil
	}

	runCmd := exec.CommandContext(ctx, filepath.Join(DefaultRunnerDir, "run.sh"), "--jitconfig", r.config)
	runCmd.Stdout = r.stdout
	runCmd.Stderr = r.stderr

	owner, err := user.Lookup("runner")
	if err != nil {
		return fmt.Errorf("lookup: %w", err)
	}

	uid, err := strconv.Atoi(owner.Uid)
	if err != nil {
		return fmt.Errorf("owner: uid: atoi: %w", err)
	}

	group, err := user.LookupGroup("docker")
	if err != nil {
		return fmt.Errorf("group: lookup: %w", err)
	}

	gid, err := strconv.Atoi(group.Gid)
	if err != nil {
		return fmt.Errorf("group: gid: atoi: %w", err)
	}

	runCmd.SysProcAttr = &syscall.SysProcAttr{Credential: &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid)}}
	runCmd.Env = append(
		runCmd.Env,
		"DEBIAN_FRONTEND=noninteractive",
		fmt.Sprintf("PATH=%s", "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"),
		fmt.Sprintf("LOGNAME=%s", owner.Username),
		fmt.Sprintf("HOME=%s", owner.HomeDir),
		fmt.Sprintf("USER=%s", owner.Username),
		fmt.Sprintf("UID=%d", uid),
		fmt.Sprintf("GID=%d", gid),
	)

	if err := runCmd.Start(); err != nil {
		return err
	}

	startCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	for {
		select {
		case <-time.After(100 * time.Millisecond):
		case <-startCtx.Done():
			return startCtx.Err()
		}

		if _, err := os.Stat(filepath.Join(DefaultRunnerDir, ".runner")); err != nil {
			continue
		}

		break
	}

	go func() {
		err := runCmd.Wait()
		if err != nil {
			r.fatalErr = err
		}

		r.exitCh <- struct{}{}
	}()

	r.runCmd = runCmd
	return nil
}

// IsRunning returns true if the GitHub runner is running.
func (r *Runner) IsRunning() bool {
	if r.runCmd == nil || r.runCmd.Process == nil {
		return false
	}

	select {
	case <-r.exitCh:
		return false
	default:
	}

	return true
}

// Wait waits for the GitHub runner to exit and returns the error returned by the
// GitHub runner (if any). If the GitHub runner is not running, this is a no-op.
func (r *Runner) Wait(ctx context.Context) error {
	if !r.IsRunning() {
		return nil
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-r.exitCh:
		return r.fatalErr
	}
}

// Stop stops the GitHub runner. If the GitHub runner is not running, this is a no-op.
func (r *Runner) Stop(ctx context.Context) error {
	if !r.IsRunning() {
		return nil
	}

	return r.runCmd.Process.Kill()
}
