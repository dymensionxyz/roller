package cronjobs

import (
	"fmt"
	"os/exec"
	"strings"
)

func Add(schedule, command string) error {
	cronEntry := fmt.Sprintf("%s %s\n", schedule, command)

	getCurrentCmd := exec.Command("crontab", "-l")
	currentCrontab, err := getCurrentCmd.Output()
	if err != nil {
		currentCrontab = []byte{}
	}

	newCrontab := string(currentCrontab) + cronEntry

	setCrontabCmd := exec.Command("crontab", "-")
	setCrontabCmd.Stdin = strings.NewReader(newCrontab)
	return setCrontabCmd.Run()
}
