package preflight

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"golang.org/x/mod/semver"
)

var (
	// Checks is a list of preflight checks to be executed
	Checks = map[string]PreflightCheckFunc{
		"Firecracker binary exists in PATH":           CheckFirecrackerBinary,
		"Firecracker version is supported (>= 1.4.1)": CheckFirecrackerVersion,
		"GitHub API is reachable":                     CheckGitHubConnectivity,
		"Virtualization is supported (KVM)":           CheckVirtualization,
	}
)

// PreflightCheckFunc is a function that performs a preflight check
type PreflightCheckFunc func() (bool, error)

// CheckFirecrackerBinary checks if the Firecracker binary is available in PATH.
func CheckFirecrackerBinary() (bool, error) {
	_, err := exec.LookPath("firecracker")
	if err != nil {
		return false, nil
	}

	return true, nil
}

// CheckFirecrackerVersion checks if the Firecracker version is supported.
func CheckFirecrackerVersion() (bool, error) {
	cmd := exec.Command("firecracker", "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, err
	}

	version := strings.Split(string(output), "\n")[0]
	version = strings.Split(version, " ")[1]

	if !semver.IsValid(version) {
		return false, fmt.Errorf("unsupported or unknown version: %s", version)
	}

	if semver.Compare(version, "v1.4.1") == -1 {
		return false, fmt.Errorf("unsupported version: %s", version)
	}

	return true, nil
}

// CheckGitHubConnectivity checks if the GitHub API is reachable.
func CheckGitHubConnectivity() (bool, error) {
	client := http.Client{
		Timeout: 10 * time.Second,
	}

	response, err := client.Get("https://github.com")
	if err != nil {
		return false, err
	}

	if response.StatusCode != 200 {
		return false, fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	return true, nil
}

// CheckVirtualization checks if the virtualization is supported.
func CheckVirtualization() (bool, error) {
	_, err := os.Stat("/dev/kvm")
	if err != nil {
		return false, err
	}

	return true, nil
}
