package rewards

import (
	"encoding/hex"
	"fmt"

	cosmossdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rewards [address]",
		Short: "temporary command to handle sequencer rewards address",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			pterm.Info.Println("not implemented")
		},
	}

	return cmd
}

func validateAddress(a string, prefix string) error {
	var addr []byte
	if len(a) == 0 {
		return fmt.Errorf("address cannot be empty")
	}

	// TODO: review
	// from cosmos sdk (https://github.com/cosmos/cosmos-sdk/blob/v0.46.16/client/debug/main.go#L203)
	var err error
	addr, err = hex.DecodeString(a)
	if err != nil {
		addr, err = cosmossdktypes.GetFromBech32(a, prefix)
		if err != nil {
			return fmt.Errorf("failed to decode address: %v", err)
		}
	}

	pterm.Info.Printf("%s (%X) is a valid address\n", cosmossdktypes.AccAddress(addr), addr)
	return nil
}
