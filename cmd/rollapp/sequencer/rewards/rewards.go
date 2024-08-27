package rewards

import (
    "encoding/hex"
    "encoding/json"
    "fmt"
    "os/exec"
    "path/filepath"

    cosmossdktypes "github.com/cosmos/cosmos-sdk/types"
    "github.com/pterm/pterm"
    "github.com/spf13/cobra"

    initconfig "github.com/dymensionxyz/roller/cmd/config/init"
    "github.com/dymensionxyz/roller/cmd/consts"
    "github.com/dymensionxyz/roller/cmd/utils"
    globalutils "github.com/dymensionxyz/roller/utils"
    "github.com/dymensionxyz/roller/utils/bash"
    "github.com/dymensionxyz/roller/utils/config/tomlconfig"
    "github.com/dymensionxyz/roller/utils/rollapp"
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

            home, err := globalutils.ExpandHomePath(cmd.Flag(utils.FlagNames.Home).Value.String())

            if err != nil {
                pterm.Error.Println("failed to expand home directory")
                return
            }

            rollerCfg, err := tomlconfig.LoadRollerConfig(home)
            if err != nil {
                return
            }

            getRaCmd := rollapp.GetRollappCmd(rollerCfg.RollappID, rollerCfg.HubData)
            var raResponse rollapp.ShowRollappResponse
            out, err := bash.ExecCommandWithStdout(getRaCmd)
            if err != nil {
                pterm.Error.Println("failed to get rollapp: ", err)
                return
            }

            err = json.Unmarshal(out.Bytes(), &raResponse)
            if err != nil {
                pterm.Error.Println("failed to unmarshal", err)
                return
            }

            bech32Prefix = raResponse.Rollapp.Bech32Prefix

            var address string
            if len(args) != 0 {
                address = args[0]
            } else {
                address, _ = pterm.DefaultInteractiveTextInput.WithDefaultText(
                    "Sequencer reward address (press enter to create a new wallet)",
                ).Show()

                if address == "" {
                    kc := utils.KeyConfig{
                        Dir:         consts.ConfigDirName.RollappSequencerKeys,
                        ID:          consts.KeysIds.RollappSequencerReward,
                        ChainBinary: consts.Executables.RollappEVM,
                        Type:        consts.EVM_ROLLAPP,
                    }

                    address, err = utils.GetAddressBinary(kc, home)
                    if err != nil {
                        pterm.Error.Printf("failed to retrieve %s: %v", kc.ID, err)
                        return
                    }

                    if address == "" {
                        pterm.Info.Println("existing reward wallet not found, creating new")
                        ki, err := initconfig.CreateAddressBinary(kc, home)
                        if err != nil {
                            pterm.Error.Println("failed to create wallet", err)
                            return
                        }

                        ki.Print(utils.WithName(), utils.WithMnemonic())
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
                pterm.Error.Printf("address %s is invalid: %v", address, err)
                return
            }

            raSequencers, err := sequencer.RegisteredRollappSequencers(raResponse.Rollapp.RollappId)
            if err != nil {
                pterm.Error.Println("failed to retrieve RollApp sequencers: ", err)
            }

            if len(raSequencers.Sequencers) == 0 {
                pterm.Info.Println("no sequencers registered, registering")

                privValidatorKeyPath := filepath.Join(
                    home,
                    consts.ConfigDirName.RollappSequencerKeys,
                    consts.KeysIds.RollappSequencerPrivValidator,
                )
                impPrivValKeyCmd := exec.Command(
                    consts.Executables.RollappEVM,
                    "tx", "sequencer", "unsafe-import-cons-key",
                    consts.KeysIds.RollappSequencerPrivValidator, privValidatorKeyPath,
                )
                _, err = bash.ExecCommandWithStdout(impPrivValKeyCmd)
                if err != nil {
                    pterm.Error.Println("failed to import sequencer key", err)
                }

                regSeqCmd := exec.Command(
                    consts.Executables.RollappEVM,
                    "tx",
                    "sequencer",
                    "create-sequencer",
                    consts.KeysIds.RollappSequencerPrivValidator,
                    "--from",
                    "rollapp",
                    "--gas-prices",
                    fmt.Sprintf("100000000000a%s", raResponse.Rollapp.Metadata.TokenSymbol),
                )
                fmt.Println(regSeqCmd.String())

                txHash, err := bash.ExecCommandWithInput(regSeqCmd)
                if err != nil {
                    pterm.Error.Println("failed to update sequencer: ", err)
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
                    "--gas-prices", "100000000000aRUN",
                )

                fmt.Println(updSeqCmd.String())

                uTxHash, err := bash.ExecCommandWithInput(updSeqCmd)
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
