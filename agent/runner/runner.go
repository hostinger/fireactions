package runner

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	// DefaultRunnerDir is the default directory where the GitHub runner is
	// installed.
	DefaultRunnerDir = "/opt/runner"
)

var (
	// ErrNotConfigured is returned when the GitHub runner is attepmted to be
	// started without being configured.
	ErrNotConfigured = fmt.Errorf("runner is not configured")
)

// Runner represents the actual GitHub runner.
type Runner struct {
	Name          string
	URL           string
	WorkDir       string
	Labels        []string
	Ephemeral     bool
	Replace       bool
	DisableUpdate bool

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

// WithEphemeral sets the ephemeral flag of the GitHub runner.
func WithEphemeral(ephemeral bool) Opt {
	f := func(r *Runner) {
		r.Ephemeral = ephemeral
	}

	return f
}

// WithReplace sets the replace flag of the GitHub runner.
func WithReplace(replace bool) Opt {
	f := func(r *Runner) {
		r.Replace = replace
	}

	return f
}

// WithDisableUpdate sets the disable update flag of the GitHub runner.
func WithDisableUpdate(disableUpdate bool) Opt {
	f := func(r *Runner) {
		r.DisableUpdate = disableUpdate
	}

	return f
}

// WithWorkDir sets the work directory of the GitHub runner.
func WithWorkDir(workDir string) Opt {
	f := func(r *Runner) {
		r.WorkDir = workDir
	}

	return f
}

// New creates a new Runner.
func New(name, url string, labels []string, opts ...Opt) *Runner {
	r := &Runner{
		Name:          name,
		URL:           url,
		Labels:        labels,
		WorkDir:       "/home/runner/work",
		Ephemeral:     true,
		DisableUpdate: true,
		Replace:       true,
		runCmd:        nil,
		stdout:        os.Stdout,
		stderr:        os.Stderr,
		fatalErr:      nil,
		exitCh:        make(chan struct{}),
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// Run starts the GitHub runner. This requires the GitHub runner to be configured first.
// If the GitHub runner is already running, this is a no-op.
func (r *Runner) Run(ctx context.Context) error {
	if !r.IsConfigured() {
		return ErrNotConfigured
	}

	if r.IsRunning() {
		return nil
	}

	runCmd := exec.CommandContext(ctx, filepath.Join(DefaultRunnerDir, "run.sh"))
	runCmd.Stdout = r.stdout
	runCmd.Stderr = r.stderr
	err := setCommandUser(runCmd, "runner")
	if err != nil {
		return fmt.Errorf("setCommandUser: %w", err)
	}

	if err := runCmd.Start(); err != nil {
		close(r.exitCh)
		return fmt.Errorf("run.sh: %w", err)
	}

	go func() {
		err := runCmd.Wait()
		if err != nil {
			r.fatalErr = fmt.Errorf("run.sh: %w", err)
		}

		close(r.exitCh)
	}()

	r.runCmd = runCmd
	return nil
}

// Configure configures the GitHub runner. This requires the GitHub
// organisation self-hosted runner remove token.
func (r *Runner) Configure(ctx context.Context, token string) error {
	if r.IsConfigured() {
		return nil
	}

	configArgs := []string{"--url", r.URL, "--name", r.Name, "--work", r.WorkDir,
		"--labels", strings.Join(r.Labels, ","), "--token", token, "--unattended", "--no-default-labels"}

	if r.Ephemeral {
		configArgs = append(configArgs, "--ephemeral")
	}

	if r.Replace {
		configArgs = append(configArgs, "--replace")
	}

	if r.DisableUpdate {
		configArgs = append(configArgs, "--disableupdate")
	}

	configCmd := exec.CommandContext(ctx, filepath.Join(DefaultRunnerDir, "config.sh"), configArgs...)
	configCmd.Stdout = r.stdout
	configCmd.Stderr = r.stderr
	err := setCommandUser(configCmd, "runner")
	if err != nil {
		return fmt.Errorf("setCommandUser: %w", err)
	}

	return configCmd.Run()
}

// Unconfigure unconfigures the GitHub runner. This requires the GitHub
// organisation self-hosted runner registration token.
func (r *Runner) Unconfigure(ctx context.Context, token string) error {
	if !r.IsConfigured() {
		return nil
	}

	configCmd := exec.CommandContext(ctx, filepath.Join(DefaultRunnerDir, "config.sh"), "remove", "--token", token)
	configCmd.Stdout = r.stdout
	configCmd.Stderr = r.stderr
	err := setCommandUser(configCmd, "runner")
	if err != nil {
		return fmt.Errorf("setCommandUser: %w", err)
	}

	return configCmd.Run()
}

// IsConfigured returns true if the GitHub runner is configured (i.e. the
// .runner file exists).
func (r *Runner) IsConfigured() bool {
	_, err := os.Stat(filepath.Join(DefaultRunnerDir, ".runner"))
	if err != nil && os.IsNotExist(err) {
		return false
	}

	return true
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
