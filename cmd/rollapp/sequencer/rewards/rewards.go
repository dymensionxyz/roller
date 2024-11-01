package rewards

import (
	"encoding/hex"
	"fmt"
	"os/exec"
	"path/filepath"

	cosmossdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/filesystem"
	"github.com/dymensionxyz/roller/utils/keys"
	"github.com/dymensionxyz/roller/utils/rollapp"
	"github.com/dymensionxyz/roller/utils/roller"
	"github.com/dymensionxyz/roller/utils/sequencer"
	"github.com/dymensionxyz/roller/utils/tx"
)

func Cmd() *cobra.Command {
	var bech32Prefix string

	cmd := &cobra.Command{
		Use:   "rewards [address]",
		Short: "temporary command to handle sequencer rewards address",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := initconfig.AddFlags(cmd)
			if err != nil {
				pterm.Error.Println("failed to add flags")
				return
			}

			home, err := filesystem.ExpandHomePath(
				cmd.Flag(initconfig.GlobalFlagNames.Home).Value.String(),
			)
			if err != nil {
				pterm.Error.Println("failed to expand home directory")
				return
			}

			rollerCfg, err := roller.LoadConfig(home)
			if err != nil {
				return
			}

			raResponse, err := rollapp.GetMetadataFromChain(rollerCfg.RollappID, rollerCfg.HubData)
			if err != nil {
				pterm.Error.Println("failed to fetch rollapp information from hub: ", err)
				return
			}

			// check whether the address is imported
			pterm.Info.Println("checking whether the validator key is present in the keyring")
			privValidatorKeyPath := filepath.Join(
				home,
				consts.ConfigDirName.Rollapp,
				"config",
				"priv_validator_key.json",
			)

			pterm.Info.Println("importing the validator key")
			err = bash.ExecCommandWithInteractions(
				consts.Executables.RollappEVM,
				"tx",
				"sequencer",
				"unsafe-import-cons-key",
				consts.KeysIds.RollappSequencerPrivValidator,
				privValidatorKeyPath,
				"--keyring-backend",
				"test",
				"--keyring-dir",
				filepath.Join(home, consts.ConfigDirName.RollappSequencerKeys),
			)
			if err != nil {
				pterm.Error.Println("failed to import sequencer key", err)
			}

			// check for existing sequencer with the imported address

			// when the sequencer isn't registered go through the flow
			// of registering the sequencer and settings the reward address
			var address string
			bech32Prefix = raResponse.Rollapp.GenesisInfo.Bech32Prefix
			kc := keys.KeyConfig{
				Dir:            consts.ConfigDirName.RollappSequencerKeys,
				ID:             consts.KeysIds.RollappSequencerReward,
				ChainBinary:    consts.Executables.RollappEVM,
				Type:           consts.EVM_ROLLAPP,
				KeyringBackend: rollerCfg.KeyringBackend,
			}

			isKeyInKeyring, err := kc.IsInKeyring(home)
			if err != nil {
				pterm.Error.Printf("failed to check for %s: %v", kc.ID, err)
				return
			}

			if isKeyInKeyring {
				pterm.Info.Println("key already present in the keyring")
			}

			if len(args) != 0 {
				address = args[0]
			}

			if isKeyInKeyring {
				address, err = kc.Address(home)
				if err != nil {
					pterm.Error.Println("failed to get address", err)
				}
			}

			if !isKeyInKeyring && len(args) == 0 {
				address, _ = pterm.DefaultInteractiveTextInput.WithDefaultText(
					"Sequencer reward address (press enter to create a new wallet)",
				).Show()

				if address == "" {
					if !isKeyInKeyring {
						pterm.Info.Println("existing reward wallet not found, creating new")
						ki, err := kc.Create(home)
						if err != nil {
							pterm.Error.Println("failed to create wallet", err)
							return
						}

						ki.Print(keys.WithName(), keys.WithMnemonic())
						address = ki.Address
					}
				}
			}

			// Set the bech32 prefix for the SDK
			config := cosmossdktypes.GetConfig()
			config.SetBech32PrefixForAccount(bech32Prefix, bech32Prefix+"pub")
			config.SetBech32PrefixForValidator(bech32Prefix+"valoper", bech32Prefix+"valoperpub")
			config.SetBech32PrefixForConsensusNode(
				bech32Prefix+"valcons",
				bech32Prefix+"valconspub",
			)

			err = validateAddress(address, bech32Prefix)
			if err != nil {
				pterm.Error.Printf("address %s is invalid: %v\n", address, err)
				return
			}

			raSequencers, err := sequencer.RegisteredRollappSequencers(raResponse.Rollapp.RollappId)
			if err != nil {
				pterm.Error.Println("failed to retrieve RollApp sequencers: ", err)
			}

			if len(raSequencers.Sequencers) == 0 {
				pterm.Info.Println("no sequencers registered, registering")

				createSeqCmd := exec.Command(
					consts.Executables.RollappEVM,
					"tx",
					"sequencer",
					"create-sequencer",
					consts.KeysIds.RollappSequencerPrivValidator,
					"--from",
					"rollapp",
					"--gas-prices",
					fmt.Sprintf("100000000000a%s", raResponse.Rollapp.GenesisInfo.NativeDenom.Base),
					"--keyring-backend", "test",
					"--keyring-dir", filepath.Join(home, consts.ConfigDirName.RollappSequencerKeys),
				)
				fmt.Println(createSeqCmd.String())

				createSeqOut, err := bash.ExecCommandWithInput(
					createSeqCmd,
					"signatures",
				)
				if err != nil {
					pterm.Error.Println("failed to create sequencer: ", err)
					return
				}

				txHash, err := bash.ExtractTxHash(createSeqOut)
				if err != nil {
					return
				}

				err = tx.MonitorTransaction("http://localhost:26657", txHash)
				if err != nil {
					pterm.Error.Println("failed to update sequencer: ", err)
					return
				}

				updSeqCmd := exec.Command(
					consts.Executables.RollappEVM,
					"tx", "sequencer", "update-sequencer",
					address, "--keyring-backend", "test", "--node", "http://localhost:26657",
					"--chain-id", rollerCfg.RollappID,
					"--from", "rollapp",
					"--gas-prices",
					fmt.Sprintf("100000000000a%s", raResponse.Rollapp.GenesisInfo.NativeDenom.Base),
					"--keyring-backend", "test",
					"--keyring-dir", filepath.Join(home, consts.ConfigDirName.RollappSequencerKeys),
				)

				fmt.Println(updSeqCmd.String())

				uTxOutput, err := bash.ExecCommandWithInput(updSeqCmd, "signatures")
				if err != nil {
					pterm.Error.Println("failed to update sequencer: ", err)
					return
				}

				uTxHash, err := bash.ExtractTxHash(uTxOutput)
				if err != nil {
					pterm.Error.Println("failed to update sequencer: ", err)
					return
				}

				err = tx.MonitorTransaction("http://localhost:26657", uTxHash)
				if err != nil {
					pterm.Error.Println("failed to update sequencer: ", err)
					return
				}
			}
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
