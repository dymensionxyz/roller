package utils

import (
	"errors"
	"fmt"
	"github.com/briandowns/spinner"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/config"
	"github.com/olekukonko/tablewriter"
	"math/big"
	"os"
	"strings"
	"time"
)

func PrintInsufficientBalancesIfAny(addressesData []NotFundedAddressData, config config.RollappConfig) {
	if len(addressesData) == 0 {
		return
	}
	keyIdToNetwork := map[string]string{
		consts.KeysIds.HubSequencer:     config.HubData.DisplayName + " Hub",
		consts.KeysIds.HubRelayer:       config.HubData.DisplayName + " Hub",
		consts.KeysIds.DALightNode:      consts.DefaultCelestiaNetwork,
		consts.KeysIds.RollappRelayer:   config.RollappID,
		consts.KeysIds.RollappSequencer: config.RollappID,
	}
	printAddresses := func() {
		data := make([][]string, len(addressesData))
		for i, addressData := range addressesData {
			data[i] = []string{
				addressData.KeyName,
				addressData.Address,
				addressData.CurrentBalance.String() + addressData.Denom,
				addressData.RequiredBalance.String() + addressData.Denom,
				strings.Title(keyIdToNetwork[addressData.KeyName]),
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
}

func GetLoadingSpinner() *spinner.Spinner {
	return spinner.New(spinner.CharSets[9], 100*time.Millisecond)
}
