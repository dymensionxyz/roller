package filesystem

import (
	"fmt"
	"time"

	"github.com/briandowns/spinner"
	"github.com/pterm/pterm"
)

type OutputHandler struct {
	NoOutput bool
	spinner  *spinner.Spinner
}

func GetLoadingSpinner() *spinner.Spinner {
	return spinner.New(spinner.CharSets[9], 100*time.Millisecond)
}

func NewOutputHandler(noOutput bool) *OutputHandler {
	if noOutput {
		return &OutputHandler{
			NoOutput: noOutput,
		}
	}
	return &OutputHandler{
		NoOutput: noOutput,
		spinner:  GetLoadingSpinner(),
	}
}

func (o *OutputHandler) DisplayMessage(msg string) {
	if !o.NoOutput {
		fmt.Println(msg)
	}
}

func (o *OutputHandler) StartSpinner(suffix string) {
	if !o.NoOutput {
		o.spinner.Suffix = suffix
		o.spinner.Restart()
	}
}

func (o *OutputHandler) StopSpinner() {
	if !o.NoOutput {
		o.spinner.Stop()
	}
}

func (o *OutputHandler) PromptOverwriteConfig(home string) (bool, error) {
	if o.NoOutput {
		return true, nil
	}

	shouldOverwrite, _ := pterm.DefaultInteractiveConfirm.WithDefaultText(
		fmt.Sprintf("Directory %s is not empty. Do you want to overwrite it?", home),
	).WithDefaultValue(false).Show()

	return shouldOverwrite, nil
}
