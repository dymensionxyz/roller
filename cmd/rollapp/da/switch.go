package da

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/x/params/client/utils"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	datalayer "github.com/dymensionxyz/roller/data_layer"
	"github.com/dymensionxyz/roller/data_layer/avail"
	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
	"github.com/dymensionxyz/roller/utils/filesystem"
	"github.com/dymensionxyz/roller/utils/genesis"
	"github.com/dymensionxyz/roller/utils/gov"
	"github.com/dymensionxyz/roller/utils/rollapp"
	"github.com/dymensionxyz/roller/utils/roller"
	sequtils "github.com/dymensionxyz/roller/utils/sequencer"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "switch-da",
		Short: "Switch to another DA.",
		Long:  ``,
		Args:  cobra.MaximumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			envs := []string{"avail"} // TODO: support more DAs
			env, _ := pterm.DefaultInteractiveSelect.
				WithDefaultText("select the DA you want to switch").
				WithOptions(envs).
				Show()

			err := initconfig.AddFlags(cmd)
			if err != nil {
				pterm.Error.Println("failed to switch DA: ", err)
				return
			}

			home, err := filesystem.ExpandHomePath(
				cmd.Flag(initconfig.GlobalFlagNames.Home).Value.String(),
			)
			if err != nil {
				pterm.Error.Println("failed to switch DA: ", err)
				return
			}

			rollappConfig, err := roller.LoadConfig(home)
			if err != nil {
				pterm.Error.Println("failed to load roller config: ", err)
				return
			}

			if strings.ToLower(string(rollappConfig.DA.Backend)) == env {
				pterm.Info.Println("You are switching to the same DA as previous!")
				return
			}

			var dalayer datalayer.DataLayer
			switch env {
			case "avail":
				raResp, err := rollapp.GetMetadataFromChain(rollappConfig.RollappID, rollappConfig.HubData)
				if err != nil {
					pterm.Error.Println("failed to get metadata from chain: ", err)
					return
				}

				drsVersion, err := genesis.GetDrsVersionFromGenesis(home, raResp)
				if err != nil {
					pterm.Error.Println("failed to get drs version from genesis: ", err)
					return
				}

				drsVersionInt, err := strconv.ParseInt(drsVersion, 10, 64)
				if err != nil {
					pterm.Error.Println("failed to get drs version from genesis: ", err)
					return
				}

				if !rollapp.IsDaConfigNewFormat(drsVersionInt, strings.ToLower(raResp.Rollapp.VmType)) {
					pterm.Error.Println("required Rollapp DRS version of at least", drsVersionInt)
					return
				}

				dalayer = avail.NewAvail(home)

			default:
				pterm.Error.Println("switch does not support da: ", env)
				return
			}

			submited, _ := pterm.DefaultInteractiveConfirm.WithDefaultText("Have you submitted to gov yet?").Show()
			if !submited {
				// Create Gov
				keyName, _ := pterm.DefaultInteractiveTextInput.WithDefaultText("Enter your key name").Show()
				keyring, _ := pterm.DefaultInteractiveTextInput.WithDefaultText("Enter your keyring-backend").Show()
				title, _ := pterm.DefaultInteractiveTextInput.WithDefaultText("Enter your Proposal Title").Show()
				description, _ := pterm.DefaultInteractiveTextInput.WithDefaultText("Enter your Proposal Description").Show()
				deposit, _ := pterm.DefaultInteractiveTextInput.WithDefaultText("Enter your Proposal Deposit").Show()
				newDAParam := json.RawMessage(fmt.Sprintf(`"%s"`, env))
				txHash, err := gov.ParamChangeProposal(home, keyName, keyring,
					&utils.ParamChangeProposalJSON{
						Title:       title,
						Description: description,
						Changes: utils.ParamChangesJSON{
							utils.NewParamChangeJSON("rollappparams", "da", newDAParam),
						},
						Deposit: deposit + rollappConfig.Denom,
					})

				if err != nil {
					pterm.Error.Println("failed to submit proposal", err)
					return
				}
				pterm.Info.Println("Proposal Tx hash: ", txHash)
			}

			daConfig := dalayer.GetSequencerDAConfig(consts.NodeType.Sequencer)
			rollappConfig.DA.Backend = consts.DAType(env)

			dymintConfigPath := sequtils.GetDymintFilePath(home)

			pterm.Info.Println("updating dymint configuration")

			_ = tomlconfig.UpdateFieldInFile(
				dymintConfigPath,
				"da_layer",
				[]string{string(env)},
			)

			_ = tomlconfig.UpdateFieldInFile(
				dymintConfigPath,
				"da_config",
				daConfig,
			)

			if err := roller.WriteConfig(rollappConfig); err != nil {
				pterm.Error.Println("failed to write roller config", err)
				return
			}

			pterm.Info.Println("the config update process is complete! Now you need to restart the nodes before the proposal passes.")

		},
	}

	return cmd
}
