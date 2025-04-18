package bnb

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/big"
	"os/exec"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/errorhandling"
	"github.com/dymensionxyz/roller/utils/keys"
	"github.com/dymensionxyz/roller/utils/roller"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pterm/pterm"
)

const (
	ConfigFileName      = "bnb.toml"
	MnemonicEntropySize = 256
	requiredAVL         = 1
)

type Bnb struct {
	Root        string
	PrivateKey  string
	Address     string
	RpcEndpoint string
	ChainID     uint32
}

func (b *Bnb) GetPrivateKey() (string, error) {
	return b.PrivateKey, nil
}

func (b *Bnb) SetMetricsEndpoint(endpoint string) {
}

func NewBnb(root string) *Bnb {
	var daNetwork string

	rollerData, err := roller.LoadConfig(root)
	errorhandling.PrettifyErrorIfExists(err)

	cfgPath := GetCfgFilePath(root)
	bnbConfig, err := loadConfigFromTOML(cfgPath)

	if err != nil {
		if rollerData.HubData.Environment == "mainnet" {
			daNetwork = string(consts.BnbMainnet)
			bnbConfig.ChainID = 56
		} else {
			daNetwork = string(consts.BnbTestnet)
			bnbConfig.ChainID = 97
		}

		daData, exists := consts.DaNetworks[daNetwork]
		if !exists {
			pterm.Error.Printf("DA network configuration not found for: %s", daNetwork)
			return &bnbConfig
		}

		bnbConfig.RpcEndpoint = daData.RpcUrl
		bnbConfig.Root = root

		useExistingBnbWallet, _ := pterm.DefaultInteractiveConfirm.WithDefaultText(
			"would you like to import an existing Bnb wallet?",
		).Show()

		if useExistingBnbWallet {
			bnbConfig.PrivateKey, _ = pterm.DefaultInteractiveTextInput.WithDefaultText(
				"> Enter your hex private key",
			).Show()
			privateKey, err := crypto.HexToECDSA(bnbConfig.PrivateKey)
			if err != nil {
				pterm.Error.Println("failed to parse private key from hex", err)
			}
			publicKey := privateKey.Public().(*ecdsa.PublicKey)
			address := crypto.PubkeyToAddress(*publicKey).Hex()

			bnbConfig.Address = address
		} else {
			privateKey, err := crypto.GenerateKey()
			if err != nil {
				pterm.Error.Println("failed to generate new private key", err)
			}

			privateKeyBytes := crypto.FromECDSA(privateKey)
			privateKeyHex := hex.EncodeToString(privateKeyBytes)
			bnbConfig.PrivateKey = privateKeyHex

			fmt.Printf("\t%s\n", bnbConfig.PrivateKey)
			fmt.Println()
			fmt.Println(pterm.LightYellow("💡 save this information and keep it safe"))

			publicKey := privateKey.Public().(*ecdsa.PublicKey)
			address := crypto.PubkeyToAddress(*publicKey).Hex()
			bnbConfig.Address = address
		}

		pterm.DefaultSection.WithIndentCharacter("🔔").Println("Please fund your bnb addresses below")
		pterm.DefaultBasicText.Println(pterm.LightGreen(bnbConfig.Address))

		for {
			proceed, _ := pterm.DefaultInteractiveConfirm.WithDefaultValue(false).
				WithDefaultText(
					"press 'y' when the wallet is funded",
				).Show()

			if !proceed {
				pterm.Error.Println("BNB addr needs to be funded!")
				continue
			}

			balance, err := bnbConfig.getBalance()
			if err != nil {
				pterm.Println("Error getting balance:", err)
				continue
			}

			if balance.Cmp(big.NewInt(0)) > 0 {
				pterm.Println("Wallet funded with balance:", balance)
				break
			}
			pterm.Error.Println("BNB wallet needs to be funded!")
		}

		err = writeConfigToTOML(cfgPath, bnbConfig)
		if err != nil {
			panic(err)
		}
	}
	return &bnbConfig
}

func (b *Bnb) InitializeLightNodeConfig() (string, error) {
	return "", nil
}

func (b *Bnb) GetDAAccountAddress() (*keys.KeyInfo, error) {
	return nil, nil
}

func (b *Bnb) GetRootDirectory() string {
	return b.Root
}

func (b *Bnb) CheckDABalance() ([]keys.NotFundedAddressData, error) {
	return nil, nil
}

func (b *Bnb) GetStartDACmd() *exec.Cmd {
	return nil
}

func (b *Bnb) GetDAAccData(cfg roller.RollappConfig) ([]keys.AccountData, error) {
	return nil, nil
}

func (b *Bnb) GetSequencerDAConfig(_ string) string {
	return fmt.Sprintf(
		`{"endpoint": "%s", "chain_id": %d, "timeout": 5000000000, "private_key_hex": "%s"}`,
		b.RpcEndpoint,
		b.ChainID,
		b.PrivateKey,
	)
}

func (b *Bnb) SetRPCEndpoint(rpc string) {
	b.RpcEndpoint = rpc
}

func (b *Bnb) GetLightNodeEndpoint() string {
	return ""
}

func (b *Bnb) GetNetworkName() string {
	return "bnb"
}

func (b *Bnb) GetStatus(c roller.RollappConfig) string {
	return "Active"
}

func (b *Bnb) GetKeyName() string {
	return "bnb"
}

func (b *Bnb) GetNamespaceID() string {
	return ""
}

func (b *Bnb) GetAppID() uint32 {
	return 0
}

func (b *Bnb) getBalance() (*big.Int, error) {
	client, err := ethclient.Dial(b.RpcEndpoint)
	if err != nil {
		return nil, err
	}
	balance, err := client.BalanceAt(context.Background(), common.HexToAddress(b.Address), nil)
	if err != nil {
		return nil, err
	}

	return balance, nil
}
