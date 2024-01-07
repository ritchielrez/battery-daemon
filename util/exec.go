// This package contains simple functions,  which may be
// needed by other packages.
package util

import (
	"log"
	"os/exec"
	"strings"

	"github.com/cli/safeexec"
)

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
