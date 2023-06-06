package initconfig

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/manifoldco/promptui"
)

func dirNotEmpty(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return false, err
	}
	files, err := ioutil.ReadDir(path)
	return len(files) > 0, err
}

func promptOverwriteConfig(home string) (bool, error) {
	prompt := promptui.Prompt{
		Label:     fmt.Sprintf("Directory %s is not empty. Do you want to overwrite", home),
		IsConfirm: true,
	}
	_, err := prompt.Run()
	if err != nil {
		if err == promptui.ErrAbort {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
