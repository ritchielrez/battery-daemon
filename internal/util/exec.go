// Package util contains simple functions, which may be
// needed by other packages.
//
// This package is only intended to be used by this project.
package util

import (
	"log"
	"os/exec"
	"strings"

	"github.com/cli/safeexec"
)

// RunCommand runs a command with the help of `safeexec` library.
// Retuns the command output as string.
// Prints an error if there is one and causes
// the program to exit out early.
func RunCommand(command string, args ...string) string {
	commandBin, err := safeexec.LookPath(command)
	if err != nil {
		log.Fatalf("Error finding command %s, error %v\n", command, err)
	}

	cmd := exec.Command(commandBin, args...)
	cmd_output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error runnning command %s, error %v\n", command, err)
	}

	return strings.TrimSpace(string(cmd_output))
}
