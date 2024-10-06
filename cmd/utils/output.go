package utils

import (
	"errors"
	"fmt"
	"math/big"
	"os"

	"github.com/manifoldco/promptui"
	"github.com/olekukonko/tablewriter"
	"github.com/pterm/pterm"
)

func PrintInsufficientBalancesIfAny(
	addressesData []NotFundedAddressData,
) error {
	if len(addressesData) == 0 {
		return errors.New("no addresses to print")
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
		Println("Please fund the addresses below.")
	printAddresses()

	// TODO: to util
	proceed, _ := pterm.DefaultInteractiveConfirm.WithDefaultValue(false).
		WithDefaultText(
			"press 'y' when the wallets are funded",
		).Show()
	if !proceed {
		pterm.Info.Println("exiting")
		return errors.New("cancelled by user")
	}

	return nil
}

type NotFundedAddressData struct {
	KeyName         string
	Address         string
	CurrentBalance  *big.Int
	RequiredBalance *big.Int
	Denom           string
	Network         string
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
