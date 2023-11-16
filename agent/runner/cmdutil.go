package runner

import (
	"fmt"
	"os/exec"
	"os/user"
	"strconv"
	"syscall"
)

func setCommandUser(cmd *exec.Cmd, username string) error {
	user, err := user.Lookup(username)
	if err != nil {
		return fmt.Errorf("lookup: %w", err)
	}

	uid, err := strconv.Atoi(user.Uid)
	if err != nil {
		return fmt.Errorf("atoi: %w", err)
	}

	gid, err := strconv.Atoi(user.Gid)
	if err != nil {
		return fmt.Errorf("atoi: %w", err)
	}

	cmd.SysProcAttr = &syscall.SysProcAttr{Credential: &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid)}}
	cmd.Env = append(
		cmd.Env,
		fmt.Sprintf("PATH=%s", "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"),
		fmt.Sprintf("LOGNAME=%s", user.Username),
		fmt.Sprintf("HOME=%s", user.HomeDir),
		fmt.Sprintf("USER=%s", user.Username),
		fmt.Sprintf("UID=%d", uid),
		fmt.Sprintf("GID=%d", gid),
	)

	return nil
}
