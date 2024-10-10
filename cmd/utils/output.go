package utils

import (
	"github.com/manifoldco/promptui"
)

func PromptBool(msg string) (bool, error) {
	prompt := promptui.Prompt{
		Label:     msg,
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
