package runnermanager

import (
	"fmt"
	"os"
	"path/filepath"
)

func setupHostname(rootPath string, hostname string) error {
	err := os.WriteFile(filepath.Join(rootPath, "etc", "hostname"), []byte(fmt.Sprintf("%s\n", hostname)), 0644)
	if err != nil {
		return fmt.Errorf("error writing %s: %v", filepath.Join(rootPath, "etc", "hostname"), err)
	}

	return nil
}

func setupDNS(rootPath string) error {
	resolvConfPath := filepath.Join(rootPath, "etc", "resolv.conf")

	var err error
	err = os.Remove(resolvConfPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("error removing %s: %v", resolvConfPath, err)
	}

	err = os.Symlink("../proc/net/pnp", resolvConfPath)
	if err != nil {
		return fmt.Errorf("error creating symlink /etc/resolv.conf -> ../proc/net/pnp: %v", err)
	}

	return nil
}
