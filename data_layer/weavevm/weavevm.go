package weavevm

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"os/exec"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/data_layer/avail"
	"github.com/dymensionxyz/roller/utils/keys"
	"github.com/dymensionxyz/roller/utils/roller"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pterm/pterm"
)

const (
	ConfigFileName        = "weavevm.toml"
	privateKeySize        = 64
	ChainId        uint16 = 9496
)

type WeaveVM struct {
	PrivateKey  string
	RpcEndpoint string
	ChainId     int64
}

func (wv *WeaveVM) GetPrivateKey() (string, error) {
	return wv.PrivateKey, nil
}

func (wv *WeaveVM) SetMetricsEndpoint(endpoint string) {
}

func (wv *WeaveVM) GetStatus(c roller.RollappConfig) string {
	return "Active"
}

func (wv *WeaveVM) GetRootDirectory() string {
	return ""
}

func (wv *WeaveVM) GetNamespaceID() string {
	return ""
}

func NewWeaveVM(root string) *WeaveVM {

	var daNetwork string

	//rollerData, err := roller.LoadConfig(root)
	//errorhandling.PrettifyErrorIfExists(err)

	cfgPath := avail.GetCfgFilePath(root)
	weaveVmConfig, err := loadConfigFromTOML(cfgPath)
	weaveVmConfig.ChainId = int64(ChainId)

	if err != nil {
		daNetwork = string(consts.WeaveVMTestnet)

		useExistingWeaveVMWallet, _ := pterm.DefaultInteractiveConfirm.WithDefaultText(
			"would you like to import an existing WeaveVM wallet?",
		).Show()

		var privateKey *ecdsa.PrivateKey
		if useExistingWeaveVMWallet {
			weaveVmConfig.PrivateKey, _ = pterm.DefaultInteractiveTextInput.WithDefaultText(
				"> Enter your private key",
			).Show()
			// Getting public address from private key
			pKeyBytes, err := hexutil.Decode("0x" + weaveVmConfig.PrivateKey)
			if err != nil {
				panic(err)
			}
			// Convert the private key bytes to an ECDSA private key.
			privateKey, err = crypto.ToECDSA(pKeyBytes)
			if err != nil {
				panic(err)
			}
		} else {
			// Convert the private key bytes to an ECDSA private key.
			privateKey, err = crypto.GenerateKey()
			if err != nil {
				panic(err)
			}

			pkey := hex.EncodeToString(privateKey.D.Bytes())
			weaveVmConfig.PrivateKey = pkey
			fmt.Printf("\t%s %s\n", "generated private key", pkey)
			fmt.Println()
			fmt.Println(pterm.LightYellow("💡 save this information and keep it safe"))
		}

		publicKey := privateKey.Public()
		publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
		if !ok {
			panic(fmt.Errorf("error casting public key to ECDSA"))
		}
		pterm.DefaultSection.WithIndentCharacter("🔔").Println("Please fund your WeaveVM address below")
		pterm.DefaultBasicText.Println(pterm.LightGreen(crypto.PubkeyToAddress(*publicKeyECDSA)))

		proceed, _ := pterm.DefaultInteractiveConfirm.WithDefaultValue(false).
			WithDefaultText(
				"press 'y' when the wallets are funded",
			).Show()

		if !proceed {
			panic(fmt.Errorf("Avail addr need to be fund!"))
		}

		daData, exists := consts.DaNetworks[daNetwork]
		if !exists {
			panic(fmt.Errorf("DA network configuration not found for: %s", daNetwork))
		}
		weaveVmConfig.RpcEndpoint = daData.ApiUrl

		err = writeConfigToTOML(cfgPath, weaveVmConfig)
		if err != nil {
			panic(err)
		}

	}
	return &weaveVmConfig
}

func (wv *WeaveVM) GetDAAccountAddress() (*keys.KeyInfo, error) {
	return nil, nil
}

func (wv *WeaveVM) InitializeLightNodeConfig() (string, error) {
	return "", nil
}

func (wv *WeaveVM) CheckDABalance() ([]keys.NotFundedAddressData, error) {
	return []keys.NotFundedAddressData{}, nil
}

func (wv *WeaveVM) GetStartDACmd() *exec.Cmd {
	return nil
}

func (wv *WeaveVM) GetDAAccData(c roller.RollappConfig) ([]keys.AccountData, error) {
	return []keys.AccountData{}, nil
}

func (wv *WeaveVM) GetLightNodeEndpoint() string {
	return ""
}

func (wv *WeaveVM) GetSequencerDAConfig(nt string) string {
	return ""
}

func (wv *WeaveVM) SetRPCEndpoint(string) {
}

func (wv *WeaveVM) GetKeyName() string {
	return ""
}

func (wv *WeaveVM) GetNetworkName() string {
	return "local"
}

func (wv *WeaveVM) GetAppID() uint32 {
	return 0
}
