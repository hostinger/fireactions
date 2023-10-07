package preflight

import (
	"errors"
	"testing"
)

func TestFirecrackerCheck_Pass(t *testing.T) {
	check := NewFirecrackerCheck()
	check.lookPathFunc = func(file string) (string, error) {
		return "/bin/firecracker", nil
	}

	err := check.Check()
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestFirecrackerCheck_Fail(t *testing.T) {
	check := NewFirecrackerCheck()
	check.lookPathFunc = func(file string) (string, error) {
		return "", errors.New("not found")
	}

	err := check.Check()
	if err == nil {
		t.Errorf("expected error, got: %v", err)
	}
}

func TestFirecrackerCheck_Name(t *testing.T) {
	check := NewFirecrackerCheck()
	if check.Name() != "firecracker" {
		t.Errorf("expected 'firecracker', got: %s", check.Name())
	}
}
