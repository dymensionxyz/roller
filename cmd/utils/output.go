package utils

import (
	"errors"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"math/big"
	"os"
)

func PrintInsufficientBalancesIfAny(addressesData []NotFundedAddressData) {
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
			}
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Name", "Address", "Current Balance", "Required Balance"})
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
}
