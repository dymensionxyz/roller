package initconfig

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

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

func prepareDirectory(path string) (bool, error) {
	isNotEmpty, err := dirNotEmpty(path)
	if err != nil {
		return false, err
	}
	if !isNotEmpty {
		return true, nil
	}
	prompt := promptui.Prompt{
		Label:     fmt.Sprintf("Directory %s is not empty. Do you want to overwrite", path),
		IsConfirm: true,
	}
	_, err = prompt.Run()
	if err != nil {
		if err == promptui.ErrAbort {
			return false, nil
		}
		return false, err
	}
	tempDir, err := ioutil.TempDir("", "config_backup")
	if err != nil {
		return false, err
	}
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return false, err
	}
	for _, file := range files {
		src := filepath.Join(path, file.Name())
		dst := filepath.Join(tempDir, file.Name())
		err = os.Rename(src, dst)
		if err != nil {
			return false, err
		}
	}
	return true, nil
}
