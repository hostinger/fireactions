package runner

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/rs/zerolog"
)

const (
	defaultDir = "/opt/runner"
)

// RunnerState represents the current state of the runner.
type RunnerState string

const (
	StateStarting  RunnerState = "Starting"  // Runner process is starting
	StateIdle      RunnerState = "Idle"      // Runner is listening for jobs
	StateRunning   RunnerState = "Running"   // Runner is executing a job
	StateCompleted RunnerState = "Completed" // Runner completed a job
	StateExited    RunnerState = "Exited"    // Runner process exited
	StateError     RunnerState = "Error"     // Runner process exited with error
)

// Runner manages the GitHub Actions runner process.
type Runner struct {
	config    string
	directory string
	owner     string
	group     string
	stdout    io.Writer
	stderr    io.Writer
	logger    *zerolog.Logger

	stateMu sync.RWMutex
	state   RunnerState
}

// Opt is a functional option for Runner.
type Opt func(r *Runner)

// WithStdout sets the stdout writer.
func WithStdout(stdout io.Writer) Opt {
	f := func(r *Runner) {
		r.stdout = stdout
	}

	return f
}

// WithStderr sets the stderr writer.
func WithStderr(stderr io.Writer) Opt {
	f := func(r *Runner) {
		r.stderr = stderr
	}

	return f
}

// WithLogger sets the logger.
func WithLogger(logger *zerolog.Logger) Opt {
	f := func(r *Runner) {
		r.logger = logger
	}

	return f
}

// WithDirectory sets the runner directory.
func WithDirectory(dir string) Opt {
	f := func(r *Runner) {
		r.directory = dir
	}

	return f
}

// WithOwner sets the runner process owner.
func WithOwner(owner string) Opt {
	f := func(r *Runner) {
		r.owner = owner
	}

	return f
}

// WithGroup sets the runner process group.
func WithGroup(group string) Opt {
	f := func(r *Runner) {
		r.group = group
	}

	return f
}

// New creates a new Runner.
func New(config string, opts ...Opt) *Runner {
	logger := zerolog.Nop()
	r := &Runner{
		config:    config,
		directory: defaultDir,
		owner:     "runner",
		group:     "docker",
		stdout:    os.Stdout,
		stderr:    os.Stderr,
		logger:    &logger,
		state:     StateStarting,
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// GetState returns the current state of the runner.
func (r *Runner) GetState() RunnerState {
	r.stateMu.RLock()
	defer r.stateMu.RUnlock()
	return r.state
}

// setState updates the runner state and notifies listeners.
func (r *Runner) setState(state RunnerState) {
	r.stateMu.Lock()
	oldState := r.state
	r.state = state
	r.stateMu.Unlock()

	if oldState != state {
		r.logger.Info().Msgf("Runner state changed: %s -> %s", oldState, state)
	}
}

// GetVersion retrieves the GitHub Actions runner version.
func (r *Runner) GetVersion() (string, error) {
	cmd := exec.Command(filepath.Join(r.directory, "run.sh"), "--version")
	cmd.Dir = r.directory
	cmd.Env = []string{
		"RUNNER_ALLOW_RUNASROOT=1",
	}

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("run.sh --version: %w", err)
	}

	// Parse the first line as the version number
	// Output format:
	// 2.331.0
	// Runner listener exit with 0 return code, stop the service, no retry needed.
	// Exiting runner...
	lines := strings.Split(string(output), "\n")
	if len(lines) == 0 {
		return "", fmt.Errorf("empty version output")
	}

	version := strings.TrimSpace(lines[0])
	if version == "" {
		return "", fmt.Errorf("empty version string")
	}

	return version, nil
}

// Run starts the GitHub Actions runner and blocks until it exits.
func (r *Runner) Run(ctx context.Context) error {
	r.logger.Info().Msg("Starting GitHub Actions runner")

	sanitizedConfig := strings.TrimSpace(r.config)
	if strings.ContainsAny(sanitizedConfig, ";|&$`(){}[]*?~<>^!\n\r") {
		return fmt.Errorf("invalid characters in config: %q", sanitizedConfig)
	}

	runCmd := exec.CommandContext(ctx, filepath.Join(r.directory, "run.sh"), "--jitconfig", sanitizedConfig)
	runCmd.Dir = r.directory

	stdoutPipe, err := runCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("create stdout pipe: %w", err)
	}

	stderrPipe, err := runCmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("create stderr pipe: %w", err)
	}

	owner, err := user.Lookup(r.owner)
	if err != nil {
		return fmt.Errorf("lookup owner: %w", err)
	}

	uid, err := strconv.Atoi(owner.Uid)
	if err != nil {
		return fmt.Errorf("parse uid: %w", err)
	}

	group, err := user.LookupGroup(r.group)
	if err != nil {
		return fmt.Errorf("lookup group: %w", err)
	}

	gid, err := strconv.Atoi(group.Gid)
	if err != nil {
		return fmt.Errorf("parse gid: %w", err)
	}

	runCmd.SysProcAttr = &syscall.SysProcAttr{
		Credential: &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid)},
	}

	runCmd.Env = append(
		runCmd.Env,
		fmt.Sprintf("PATH=%s", "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"),
		fmt.Sprintf("LOGNAME=%s", owner.Username),
		fmt.Sprintf("HOME=%s", owner.HomeDir),
		fmt.Sprintf("USER=%s", owner.Username),
		fmt.Sprintf("UID=%d", uid),
		fmt.Sprintf("GID=%d", gid),
	)

	if err := runCmd.Start(); err != nil {
		r.setState(StateError)
		return fmt.Errorf("start runner: %w", err)
	}

	go r.pipeToLogger(stdoutPipe, *r.logger, zerolog.InfoLevel)
	go r.pipeToLogger(stderrPipe, *r.logger, zerolog.ErrorLevel)

	err = runCmd.Wait()
	if err != nil {
		r.setState(StateError)
		return err
	}

	r.setState(StateExited)
	return nil
}

// pipeToLogger reads from stdout and logs each line, detecting when runner reaches idle state.
func (r *Runner) pipeToLogger(reader io.Reader, logger zerolog.Logger, level zerolog.Level) {
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		logger.WithLevel(level).Msg(line)

		// Detect state changes based on log messages
		if strings.Contains(line, "Listening for Jobs") {
			r.setState(StateIdle)
		} else if strings.Contains(line, "Running job:") {
			r.setState(StateRunning)
		} else if strings.Contains(line, "Job") && strings.Contains(line, "completed with result:") {
			r.setState(StateCompleted)
		}
	}
}
