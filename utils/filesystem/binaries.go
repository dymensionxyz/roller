package filesystem

import (
	"os/exec"
)

func IsAvailable(binary string) bool {
	_, err := exec.LookPath(binary)
	return err == nil
}
