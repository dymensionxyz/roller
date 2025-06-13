package solana

import (
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"

	cosmossdkmath "cosmossdk.io/math"
	cosmossdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/keys"
	"github.com/dymensionxyz/roller/utils/roller"
	"github.com/pterm/pterm"
)

const (
	requiredSOL = 1 // Minimum SOL required for DA operations
)

type Solana struct {
	Root        string
	PrivateKey  string
	Address     string
	RpcEndpoint string
	ChainID     string
	ApiUrl      string
	GrpcAddress string
	Network     string
}

func NewSolana(home string) *Solana {
	root := filepath.Join(home, consts.ConfigDirName.DALightNode)
	cfgPath := filepath.Join(root, "config.toml")

	var solanaConfig Solana
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		solanaConfig = Solana{
			Root:        root,
			RpcEndpoint: "https://api.mainnet-beta.solana.com",
			ChainID:     "mainnet-beta",
			ApiUrl:      "https://api.mainnet-beta.solana.com",
			GrpcAddress: "https://api.mainnet-beta.solana.com",
			Network:     "mainnet-beta",
		}

		// Check for mnemonic in environment variable
		if mnemonic := os.Getenv("SOLANA_MNEMONIC"); mnemonic != "" {
			solanaConfig.PrivateKey = mnemonic
			// TODO: Generate address from mnemonic
			solanaConfig.Address = "YOUR_SOLANA_ADDRESS"
		} else {
			// Prompt user for wallet details
			pterm.Info.Println("Please provide your Solana wallet details")

			importExisting, _ := pterm.DefaultInteractiveConfirm.
				WithDefaultValue(false).
				WithDefaultText("Do you want to import an existing wallet?").
				Show()

			if importExisting {
				mnemonic, _ := pterm.DefaultInteractiveTextInput.
					WithDefaultText("Enter your wallet mnemonic").
					Show()
				solanaConfig.PrivateKey = mnemonic
				// TODO: Generate address from mnemonic
				solanaConfig.Address = "YOUR_SOLANA_ADDRESS"
			} else {
				// TODO: Generate new wallet
				solanaConfig.PrivateKey = "YOUR_GENERATED_PRIVATE_KEY"
				solanaConfig.Address = "YOUR_GENERATED_ADDRESS"
			}
		}

		// Check if wallet is funded
		for {
			proceed, _ := pterm.DefaultInteractiveConfirm.WithDefaultValue(false).
				WithDefaultText(
					"press 'y' when the wallet is funded",
				).Show()

			if !proceed {
				pterm.Error.Println("Solana address needs to be funded")
				continue
			}

			balance, err := solanaConfig.getBalance()
			if err != nil {
				pterm.Println("Error getting balance:", err)
				continue
			}

			if balance.Cmp(big.NewInt(0)) > 0 {
				pterm.Println("Wallet funded with balance:", balance)
				break
			}
			pterm.Error.Println("Solana wallet needs to be funded")
		}

		err := writeConfigToTOML(cfgPath, solanaConfig)
		if err != nil {
			panic(err)
		}
	} else {
		solanaConfig = loadConfigFromTOML(cfgPath)
	}
	return &solanaConfig
}

func (s *Solana) InitializeLightNodeConfig() (string, error) {
	return "", nil
}

func (s *Solana) GetDAAccountAddress() (*keys.KeyInfo, error) {
	return &keys.KeyInfo{
		Address: s.Address,
	}, nil
}

func (s *Solana) GetRootDirectory() string {
	return s.Root
}

func (s *Solana) CheckDABalance() ([]keys.NotFundedAddressData, error) {
	balance, err := s.getBalance()
	if err != nil {
		return nil, fmt.Errorf("failed to get DA balance: %w", err)
	}

	exp := new(big.Int).Exp(big.NewInt(10), big.NewInt(9), nil) // SOL has 9 decimals
	required := new(big.Int).Mul(big.NewInt(requiredSOL), exp)
	if required.Cmp(balance) > 0 {
		return []keys.NotFundedAddressData{
			{
				KeyName:         s.GetKeyName(),
				Address:         s.Address,
				CurrentBalance:  balance,
				RequiredBalance: required,
				Denom:           "SOL",
				Network:         string(consts.Solana),
			},
		}, nil
	}
	return nil, nil
}

func (s *Solana) GetStartDACmd() *exec.Cmd {
	return nil
}

func (s *Solana) GetDAAccData(cfg roller.RollappConfig) ([]keys.AccountData, error) {
	balance, err := s.getBalance()
	if err != nil {
		return nil, err
	}

	return []keys.AccountData{
		{
			Address: s.Address,
			Balance: cosmossdktypes.Coin{
				Denom:  "SOL",
				Amount: cosmossdkmath.NewIntFromBigInt(balance),
			},
		},
	}, nil
}

func (s *Solana) GetSequencerDAConfig(_ string) string {
	return `{"endpoint":"http://barcelona:8899","keypath_env":"SOLANA_KEYPATH","program_address":"5cfjxBnFMoqdbZXTMHaoXfQm7obMpYMnkT681sRd95Qo"}`
}

func (s *Solana) SetRPCEndpoint(rpc string) {
	s.RpcEndpoint = rpc
}

func (s *Solana) GetLightNodeEndpoint() string {
	return ""
}

func (s *Solana) GetNetworkName() string {
	return "solana"
}

func (s *Solana) GetStatus(c roller.RollappConfig) string {
	return "Active"
}

func (s *Solana) GetKeyName() string {
	return "solana"
}

func (s *Solana) GetNamespaceID() string {
	return ""
}

func (s *Solana) GetAppID() uint32 {
	return 0
}

func (s *Solana) getBalance() (*big.Int, error) {
	// TODO: Implement Solana balance check
	// This is a placeholder that returns a mock balance
	return big.NewInt(0), nil
}

func (s *Solana) GetPrivateKey() (string, error) {
	return s.PrivateKey, nil
}

func (s *Solana) SetMetricsEndpoint(endpoint string) {
	// Not implemented for Solana
}
