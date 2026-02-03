package tail

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/fsnotify/fsnotify"
)

const (
	// DefaultMaxLineSize is the default maximum size for a single line.
	DefaultMaxLineSize = 1024 * 1024 // 1MB
)

// Config specifies how a file should be tailed.
type Config struct {
	// Follow continues tailing the file as new lines are written.
	Follow bool
	// Location specifies where to start reading from.
	// If nil, starts from the beginning.
	Location *SeekInfo
	// MaxLineSize is the maximum size of a single line.
	// If 0, uses DefaultMaxLineSize.
	MaxLineSize int
	// ReOpen reopens the file if it's rotated or recreated.
	ReOpen bool
}

// SeekInfo specifies where to seek in the file.
type SeekInfo struct {
	Offset int64
	Whence int // io.SeekStart, io.SeekCurrent, or io.SeekEnd
}

// Line represents a line from the tailed file.
type Line struct {
	Text string
	Err  error
}

// Tail represents an active file tail.
type Tail struct {
	Lines    chan *Line
	filename string
	config   Config
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	file     *os.File
	err      error
	errMu    sync.RWMutex
}

// TailFile starts tailing a file.
func TailFile(filename string, config Config) (*Tail, error) {
	if config.MaxLineSize == 0 {
		config.MaxLineSize = DefaultMaxLineSize
	}

	ctx, cancel := context.WithCancel(context.Background())

	t := &Tail{
		Lines:    make(chan *Line),
		filename: filename,
		config:   config,
		ctx:      ctx,
		cancel:   cancel,
	}

	// Start the tail goroutine (all file operations happen inside)
	t.wg.Add(1)
	go t.run()

	return t, nil
}

// openFile opens the file for reading.
func (t *Tail) openFile() error {
	file, err := os.Open(t.filename)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	t.file = file
	return nil
}

// seekToPosition seeks to the configured position in the file.
func (t *Tail) seekToPosition() error {
	if t.config.Location == nil {
		return nil
	}

	if t.config.Location.Whence == io.SeekEnd && t.config.Location.Offset < 0 {
		// Tail n lines from the end
		return t.tailLines(int(-t.config.Location.Offset))
	}

	_, err := t.file.Seek(t.config.Location.Offset, t.config.Location.Whence)
	return err
}

// tailLines seeks to approximately n lines from the end.
func (t *Tail) tailLines(n int) error {
	stat, err := t.file.Stat()
	if err != nil {
		return err
	}

	size := stat.Size()
	if size == 0 {
		return nil
	}

	// Start reading from an estimated position
	// Average line length assumption: 100 bytes
	estimatedOffset := size - int64(n*100)
	if estimatedOffset < 0 {
		estimatedOffset = 0
	}

	if _, err := t.file.Seek(estimatedOffset, io.SeekStart); err != nil {
		return err
	}

	// Read all remaining lines and keep only the last n
	scanner := bufio.NewScanner(t.file)
	scanner.Buffer(make([]byte, 0, 64*1024), t.config.MaxLineSize)

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	// Keep only last n lines
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}

	// Send the historical lines
	for _, line := range lines {
		select {
		case <-t.ctx.Done():
			return t.ctx.Err()
		case t.Lines <- &Line{Text: line}:
		}
	}

	return nil
}

// run is the main tail loop.
func (t *Tail) run() {
	defer t.wg.Done()
	defer close(t.Lines)
	defer t.cleanup()

	// Open the file
	if err := t.openFile(); err != nil {
		t.setErr(err)
		return
	}

	// Seek to the requested position
	if err := t.seekToPosition(); err != nil {
		t.setErr(err)
		return
	}

	scanner := bufio.NewScanner(t.file)
	scanner.Buffer(make([]byte, 0, 64*1024), t.config.MaxLineSize)

	if !t.config.Follow {
		// Non-follow mode: just read the file
		t.readLines(scanner)
		return
	}

	// Follow mode: read existing lines then watch for new ones
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		t.setErr(fmt.Errorf("create watcher: %w", err))
		return
	}
	defer watcher.Close()

	if err := watcher.Add(t.filename); err != nil {
		t.setErr(fmt.Errorf("watch file: %w", err))
		return
	}

	for {
		// Read any available lines
		for scanner.Scan() {
			select {
			case <-t.ctx.Done():
				return
			case t.Lines <- &Line{Text: scanner.Text()}:
			}
		}

		if err := scanner.Err(); err != nil {
			select {
			case <-t.ctx.Done():
				return
			case t.Lines <- &Line{Err: err}:
			}
		}

		// Wait for file changes
		select {
		case <-t.ctx.Done():
			return
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				// New data written, create a new scanner to pick up the changes
				// The old scanner is at EOF and won't read new data
				scanner = bufio.NewScanner(t.file)
				scanner.Buffer(make([]byte, 0, 64*1024), t.config.MaxLineSize)
				continue
			}
			if event.Op&fsnotify.Remove == fsnotify.Remove || event.Op&fsnotify.Rename == fsnotify.Rename {
				if t.config.ReOpen {
					// File was removed/renamed, try to reopen
					if err := t.reopenFile(watcher); err != nil {
						t.setErr(err)
						return
					}
					scanner = bufio.NewScanner(t.file)
					scanner.Buffer(make([]byte, 0, 64*1024), t.config.MaxLineSize)
				} else {
					return
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			select {
			case <-t.ctx.Done():
				return
			case t.Lines <- &Line{Err: err}:
			}
		}
	}
}

// readLines reads all lines and sends them to the channel.
func (t *Tail) readLines(scanner *bufio.Scanner) {
	for scanner.Scan() {
		select {
		case <-t.ctx.Done():
			return
		case t.Lines <- &Line{Text: scanner.Text()}:
		}
	}

	if err := scanner.Err(); err != nil {
		select {
		case <-t.ctx.Done():
			return
		case t.Lines <- &Line{Err: err}:
		}
	}
}

// reopenFile closes the current file and opens it again.
func (t *Tail) reopenFile(watcher *fsnotify.Watcher) error {
	if t.file != nil {
		t.file.Close()
	}

	// Remove the old watch
	watcher.Remove(t.filename)

	// Wait for file to be recreated
	for {
		select {
		case <-t.ctx.Done():
			return t.ctx.Err()
		default:
			if err := t.openFile(); err == nil {
				// File opened successfully
				if err := watcher.Add(t.filename); err != nil {
					return fmt.Errorf("watch reopened file: %w", err)
				}
				return nil
			}
			// File doesn't exist yet, keep waiting
		}
	}
}

// Stop stops tailing the file.
func (t *Tail) Stop() {
	t.cancel()
}

// Wait waits for the tail to finish.
func (t *Tail) Wait() {
	t.wg.Wait()
}

// Cleanup stops tailing and waits for cleanup.
func (t *Tail) Cleanup() {
	t.Stop()
	t.Wait()
}

// Err returns any error that occurred during tailing.
func (t *Tail) Err() error {
	t.errMu.RLock()
	defer t.errMu.RUnlock()
	return t.err
}

// setErr sets the error.
func (t *Tail) setErr(err error) {
	if err == nil || errors.Is(err, context.Canceled) {
		return
	}
	t.errMu.Lock()
	if t.err == nil {
		t.err = err
	}
	t.errMu.Unlock()
}

// cleanup closes the file.
func (t *Tail) cleanup() {
	if t.file != nil {
		t.file.Close()
	}
}
