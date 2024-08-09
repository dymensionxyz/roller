package utils

import (
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/manifoldco/promptui"
	"github.com/olekukonko/tablewriter"
	"github.com/pterm/pterm"
)

func PrintInsufficientBalancesIfAny(
	addressesData []NotFundedAddressData,
) {
	if len(addressesData) == 0 {
		return
	}

	printAddresses := func() {
		data := make([][]string, len(addressesData))
		for i, addressData := range addressesData {
			data[i] = []string{
				addressData.KeyName,
				addressData.Address,
				addressData.CurrentBalance.String() + addressData.Denom,
				addressData.RequiredBalance.String() + addressData.Denom,
				addressData.Network,
			}
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Name", "Address", "Current", "Required", "Network"})
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		table.SetBorder(false)
		table.AppendBulk(data)
		fmt.Println()
		table.Render()
		fmt.Println()
	}

	pterm.DefaultSection.WithIndentCharacter("🔔").
		Println("Please fund the addresses below to register and run the sequencer.")
	printAddresses()

	proceed, _ := pterm.DefaultInteractiveConfirm.WithDefaultValue(true).
		WithDefaultText(
			"press enter when funded",
		).Show()
	if !proceed {
		pterm.Info.Println("exiting")
		return
	}
}

type NotFundedAddressData struct {
	KeyName         string
	Address         string
	CurrentBalance  *big.Int
	RequiredBalance *big.Int
	Denom           string
	Network         string
}

func GetLoadingSpinner() *spinner.Spinner {
	return spinner.New(spinner.CharSets[9], 100*time.Millisecond)
}

type OutputHandler struct {
	NoOutput bool
	spinner  *spinner.Spinner
}

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
