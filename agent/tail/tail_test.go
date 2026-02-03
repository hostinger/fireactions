package tail

import (
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestTailFile_NonFollow(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	// Write test data
	content := "line1\nline2\nline3\n"
	if err := os.WriteFile(logFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Tail the file
	tail, err := TailFile(logFile, Config{Follow: false})
	if err != nil {
		t.Fatalf("TailFile failed: %v", err)
	}
	defer tail.Cleanup()

	// Read all lines
	var lines []string
	for line := range tail.Lines {
		if line.Err != nil {
			t.Fatalf("Line error: %v", line.Err)
		}
		lines = append(lines, line.Text)
	}

	expected := []string{"line1", "line2", "line3"}
	if len(lines) != len(expected) {
		t.Errorf("Expected %d lines, got %d", len(expected), len(lines))
	}

	for i, line := range lines {
		if line != expected[i] {
			t.Errorf("Line %d: expected %q, got %q", i, expected[i], line)
		}
	}
}

func TestTailFile_Tail(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	// Write test data
	content := "line1\nline2\nline3\nline4\nline5\n"
	if err := os.WriteFile(logFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Tail last 2 lines
	tail, err := TailFile(logFile, Config{
		Follow: false,
		Location: &SeekInfo{
			Offset: -2,
			Whence: io.SeekEnd,
		},
	})
	if err != nil {
		t.Fatalf("TailFile failed: %v", err)
	}
	defer tail.Cleanup()

	// Read all lines
	var lines []string
	for line := range tail.Lines {
		if line.Err != nil {
			t.Fatalf("Line error: %v", line.Err)
		}
		lines = append(lines, line.Text)
	}

	expected := []string{"line4", "line5"}
	if len(lines) != len(expected) {
		t.Errorf("Expected %d lines, got %d", len(expected), len(lines))
	}

	for i, line := range lines {
		if line != expected[i] {
			t.Errorf("Line %d: expected %q, got %q", i, expected[i], line)
		}
	}
}

func TestTailFile_Follow(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	// Write initial data
	f, err := os.Create(logFile)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString("line1\n"); err != nil {
		t.Fatal(err)
	}
	f.Sync()

	// Start tailing
	tail, err := TailFile(logFile, Config{Follow: true})
	if err != nil {
		t.Fatalf("TailFile failed: %v", err)
	}
	defer tail.Cleanup()

	// Read first line
	line := <-tail.Lines
	if line.Err != nil {
		t.Fatalf("Line error: %v", line.Err)
	}
	if line.Text != "line1" {
		t.Errorf("Expected 'line1', got %q", line.Text)
	}

	// Write more data
	if _, err := f.WriteString("line2\n"); err != nil {
		t.Fatal(err)
	}
	f.Sync()

	// Read second line (with timeout)
	select {
	case line := <-tail.Lines:
		if line.Err != nil {
			t.Fatalf("Line error: %v", line.Err)
		}
		if line.Text != "line2" {
			t.Errorf("Expected 'line2', got %q", line.Text)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for new line")
	}

	f.Close()
}

func TestTailFile_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	// Create empty file
	if err := os.WriteFile(logFile, []byte{}, 0644); err != nil {
		t.Fatal(err)
	}

	// Tail the file
	tail, err := TailFile(logFile, Config{Follow: false})
	if err != nil {
		t.Fatalf("TailFile failed: %v", err)
	}
	defer tail.Cleanup()

	// Should get no lines
	var count int
	for range tail.Lines {
		count++
	}

	if count != 0 {
		t.Errorf("Expected 0 lines from empty file, got %d", count)
	}
}

func TestTailFile_Stop(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	// Create file
	if err := os.WriteFile(logFile, []byte("line1\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Start tailing
	tail, err := TailFile(logFile, Config{Follow: true})
	if err != nil {
		t.Fatalf("TailFile failed: %v", err)
	}

	// Read first line
	<-tail.Lines

	// Stop tailing
	tail.Stop()

	// Channel should close
	select {
	case _, ok := <-tail.Lines:
		if ok {
			t.Error("Expected channel to be closed")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for channel to close")
	}
}
