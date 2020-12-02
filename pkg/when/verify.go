package when

import (
	"os/exec"
)

func IsInstalled() bool {
	_, err := exec.Command("/bin/sh", "-c", "which when").CombinedOutput()

	return err == nil
}
