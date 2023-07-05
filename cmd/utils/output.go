package utils

import (
	"errors"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/dymensionxyz/roller/config"
	"github.com/olekukonko/tablewriter"
)

func PrintInsufficientBalancesIfAny(addressesData []NotFundedAddressData, config config.RollappConfig) {
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
		fmt.Println("💈 Please fund these addresses and try again.")
	}
	PrettifyErrorIfExists(errors.New("The following addresses have insufficient balance to perform this operation"),
		printAddresses)
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
