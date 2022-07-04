// +build windows

package jsEngine

import (
	"os/exec"
	"syscall"
)

func setCommandParameters(command *exec.Cmd) *exec.Cmd {
	command.SysProcAttr = &syscall.SysProcAttr{ // start command to own process (to prevent ctrl+c signal from parent)
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
	return command
}
