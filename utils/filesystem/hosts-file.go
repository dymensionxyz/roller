package filesystem

import (
	"fmt"
	"os/exec"

	"github.com/pterm/pterm"
)

func UpdateHostsFile(addr, host string) error {
	// Check if the entry already exists
	pterm.Info.Printf("adding %s to hosts file\n", host)
	pterm.Debug.Println(
		"this is necessary to access the rollapp endpoint on your local machine from the docker" +
			" container",
	)
	checkCmd := exec.Command("grep", "-q", host, "/etc/hosts")
	err := checkCmd.Run()

	if err == nil {
		// Entry already exists
		fmt.Printf("Entry for %s already exists in /etc/hosts\n", host)
		return nil
	}

	// Append the new entry
	appendCmd := exec.Command(
		"sudo",
		"sh",
		"-c",
		fmt.Sprintf("echo '%s %s' >> /etc/hosts", addr, host),
	)
	output, err := appendCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to update hosts file: %v - %s", err, string(output))
	}

	fmt.Printf("Added %s to /etc/hosts\n", host)
	return nil
}
