package lightclient

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"github.com/pterm/pterm"
	"gopkg.in/yaml.v3"

	"github.com/dymensionxyz/roller/cmd/consts"
	datalayer "github.com/dymensionxyz/roller/data_layer"
	"github.com/dymensionxyz/roller/data_layer/celestia"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/keys"
	"github.com/dymensionxyz/roller/utils/roller"
	"github.com/dymensionxyz/roller/utils/sequencer"
)

// Initialize function initializes the celestia light client on a local machine and returns the
// KeyInfo of the created celestia address
func Initialize(env string, rollerData roller.RollappConfig) (*keys.KeyInfo, error) {
	if env != "mock" {
		pterm.Info.Println("initializing da light client")
		hd := rollerData.HubData
		raID := rollerData.RollappID
		kb := rollerData.KeyringBackend

		pterm.Info.Println("initializing da light node configuration")
		damanager := datalayer.NewDAManager(rollerData.DA.Backend, rollerData.Home, kb, rollerData.NodeType)
		mnemonic, err := damanager.InitializeLightNodeConfig()
		if err != nil {
			return nil, err
		}

		pterm.Info.Println("checking for registered sequencers")
		sequencers, err := sequencer.RegisteredRollappSequencersOnHub(
			rollerData.RollappID,
			rollerData.HubData,
		)
		if err != nil {
			return nil, err
		}

		var latestHeight string
		var latestBlockIdHash string

		pterm.Info.Println("retrieving latest block")

		for {
			latestHeight, latestBlockIdHash, err = celestia.GetLatestBlock(rollerData)
			if err == nil {
				break
			}
			newRPC, _ := pterm.DefaultInteractiveTextInput.WithDefaultText(
				"the provided Celestia RPC is not working, please enter another RPC Endpoint instead (you can be obtained in the following link https://docs.celestia.org/how-to-guides/mainnet#consensus-nodes)",
			).Show()

			rollerData.DA.RpcUrl = newRPC

			if err = roller.WriteConfig(rollerData); err != nil {
				return nil, err
			}
		}

		heightInt, err := strconv.Atoi(latestHeight)
		if err != nil {
			return nil, err
		}

		pterm.Info.Println("updating da light client configuration")
		celestiaConfigFilePath := filepath.Join(
			rollerData.Home,
			consts.ConfigDirName.DALightNode,
			"config.toml",
		)
		if len(sequencers.Sequencers) == 0 {
			pterm.Info.Println("no sequencers registered for the rollapp")
			pterm.Info.Println("using latest height for da-light-client configuration")

			pterm.Info.Printf(
				"da light client will be initialized at height %s, block hash %s",
				latestHeight,
				latestBlockIdHash,
			)

			err = UpdateConfig(celestiaConfigFilePath, latestBlockIdHash, heightInt)
			if err != nil {
				return nil, err
			}
		} else {
			pterm.Info.Println("checking for state update")
			cmd := exec.Command(
				consts.Executables.Dymension,
				"q",
				"rollapp",
				"state",
				raID,
				"--index",
				"1",
				"--node",
				hd.RpcUrl,
				"--chain-id",
				hd.ID,
			)

			pterm.Info.Println("checking for state update")
			out, err := bash.ExecCommandWithStdout(cmd)
			if err != nil {
				if strings.Contains(out.String(), "NotFound") {
					pterm.Info.Printf(
						"no state found for %s, da light client will be initialized with latest height\n",
						raID,
					)
					err = UpdateConfig(celestiaConfigFilePath, latestBlockIdHash, heightInt)
					if err != nil {
						return nil, err
					}
				} else {
					return nil, err
				}
			} else {
				pterm.Info.Println("state update found, extracting da height")

				var result RollappStateResponse
				if err := yaml.Unmarshal(out.Bytes(), &result); err != nil {
					pterm.Error.Println("failed to extract state update: ", err)
					return nil, err
				}

				h, err := celestia.ExtractHeightfromDAPath(result.StateInfo.DAPath)
				if err != nil {
					pterm.Error.Println("failed to extract height from state update da path: ", err)
					return nil, err
				}

				height, hash, err := celestia.GetBlockByHeight(h, rollerData)
				if err != nil {
					pterm.Error.Println("failed to retrieve DA height: ", err)
					return nil, err
				}

				heightInt, err := strconv.Atoi(height)
				if err != nil {
					return nil, err
				}

				pterm.Info.Printf(
					"the first %s state update has DA height of %s with hash %s\n",
					rollerData.RollappID,
					height,
					hash,
				)
				pterm.Info.Printf("updating %s \n", celestiaConfigFilePath)
				err = UpdateConfig(celestiaConfigFilePath, hash, heightInt)
				if err != nil {
					return nil, err
				}
			}
		}

		pterm.Info.Println("retrieving da address")
		daAddress, err := damanager.GetDAAccountAddress()
		if err != nil {
			return nil, err
		}

		if daAddress != nil {
			ki := &keys.KeyInfo{
				Name:     damanager.GetKeyName(),
				Address:  daAddress.Address,
				Mnemonic: mnemonic,
			}
			pterm.Success.Println("successfully initialized da light client")
			return ki, nil
		}
	}

	return nil, nil
}

func UpdateConfig(file, hash string, height int) error {
	// Read existing config
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	// Parse TOML into a map
	var config map[string]interface{}
	if err := toml.Unmarshal(data, &config); err != nil {
		return err
	}

	// Update DASer.SampleFrom
	if daser, ok := config["DASer"].(map[string]interface{}); ok {
		daser["SampleFrom"] = height
	}

	// Update Header.TrustedHash
	if header, ok := config["Header"].(map[string]interface{}); ok {
		header["TrustedHash"] = hash
	}

	// Write updated config
	f, err := os.Create(file)
	if err != nil {
		return err
	}

	// nolint:errcheck
	defer f.Close()

	encoder := toml.NewEncoder(f)
	encoder.SetIndentTables(true)

	if err := encoder.Encode(config); err != nil {
		return err
	}

	return nil
}
