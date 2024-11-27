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

	"github.com/rs/zerolog"
)

const (
	defaultDir = "/opt/runner"
)

// Runner represents a virtual machine agent that's responsible for running
// the actual GitHub runner.
type Runner struct {
	config    string
	directory string
	owner     string
	group     string
	stdout    io.Writer
	stderr    io.Writer
	logger    *zerolog.Logger
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

// WithLogger sets the logger for the Runner.
func WithLogger(logger *zerolog.Logger) Opt {
	f := func(r *Runner) {
		r.logger = logger
	}

	return f
}

// WithDirectory sets the directory where the GitHub runner is located.
func WithDirectory(dir string) Opt {
	f := func(r *Runner) {
		r.directory = dir
	}

	return f
}

// WithOwner sets the owner of the GitHub runner.
func WithOwner(owner string) Opt {
	f := func(r *Runner) {
		r.owner = owner
	}

	return f
}

// WithGroup sets the group of the GitHub runner.
func WithGroup(group string) Opt {
	f := func(r *Runner) {
		r.group = group
	}

	return f
}

// New creates a new Runner.
func New(config string, opts ...Opt) *Runner {
	logger := zerolog.Nop()
	runner := &Runner{
		config:    config,
		directory: defaultDir,
		owner:     "runner",
		group:     "docker",
		stdout:    os.Stdout,
		stderr:    os.Stderr,
		logger:    &logger,
	}

	for _, opt := range opts {
		opt(runner)
	}

	return runner
}

// Start starts the GitHub runner. This requires the GitHub runner to be configured first.
// If the GitHub runner is already running, this is a no-op.
func (r *Runner) Run(ctx context.Context) error {
	r.logger.Info().Msgf("Starting GitHub runner")
	r.logger.Info().Msgf("Running command: %s", filepath.Join(defaultDir, "run.sh"))

	runCmd := exec.CommandContext(ctx, filepath.Join(defaultDir, "run.sh"), "--jitconfig", r.config)
	runCmd.Stdout = r.stdout
	runCmd.Stderr = r.stderr
	runCmd.Dir = defaultDir

	owner, err := user.Lookup(r.owner)
	if err != nil {
		return fmt.Errorf("lookup: %w", err)
	}

	uid, err := strconv.Atoi(owner.Uid)
	if err != nil {
		return fmt.Errorf("owner: uid: atoi: %w", err)
	}

	group, err := user.LookupGroup(r.group)
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
		fmt.Sprintf("PATH=%s", "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"),
		fmt.Sprintf("LOGNAME=%s", owner.Username),
		fmt.Sprintf("HOME=%s", owner.HomeDir),
		fmt.Sprintf("USER=%s", owner.Username),
		fmt.Sprintf("UID=%d", uid),
		fmt.Sprintf("GID=%d", gid),
	)

	return runCmd.Run()
}
