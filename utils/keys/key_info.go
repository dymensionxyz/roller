package keys

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/roller"
)

// KeyInfo struct stores information about a generated wallet
type KeyInfo struct {
	Name          string `json:"name"`
	Address       string `json:"address"`
	Mnemonic      string `json:"mnemonic"`
	PubKey        string `json:"pubkey"`
	PrintName     bool
	PrintMnemonic bool
	PrintPubKey   bool
}

type KeyInfoOption func(*KeyInfo)

func WithName() KeyInfoOption {
	return func(opts *KeyInfo) {
		opts.PrintName = true
	}
}

func WithMnemonic() KeyInfoOption {
	return func(opts *KeyInfo) {
		opts.PrintMnemonic = true
	}
}

func WithPubKey() KeyInfoOption {
	return func(opts *KeyInfo) {
		opts.PrintPubKey = true
	}
}

func All(rollappConfig roller.RollappConfig, hd consts.HubData) ([]KeyInfo, error) {
	var aki []KeyInfo

	// relayer
	rkc := KeyConfig{
		ChainBinary:    consts.Executables.Dymension,
		ID:             consts.KeysIds.HubRelayer,
		Dir:            filepath.Join(consts.ConfigDirName.Relayer, "keys", hd.ID),
		KeyringBackend: consts.SupportedKeyringBackends.Test,
	}
	rki, err := rkc.Info(rollappConfig.Home)
	if err != nil {
		return nil, err
	}
	aki = append(aki, *rki)

	// sequencer
	seqKc := KeyConfig{
		Dir:            consts.ConfigDirName.HubKeys,
		ID:             consts.KeysIds.HubSequencer,
		ChainBinary:    consts.Executables.Dymension,
		Type:           consts.SDK_ROLLAPP,
		KeyringBackend: rollappConfig.KeyringBackend,
	}
	j, _ := json.Marshal(seqKc)
	pterm.Info.Println(string(j))
	seqKi, err := seqKc.Info(rollappConfig.Home)
	if err != nil {
		return nil, err
	}
	aki = append(aki, *seqKi)
	uhd, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	// eibc - only if directory exists
	eibcDir := path.Join(uhd, consts.ConfigDirName.Eibc)
	if _, err := os.Stat(eibcDir); err == nil {
		eibcKc := KeyConfig{
			Dir:            consts.ConfigDirName.Eibc,
			ID:             consts.KeysIds.Eibc,
			ChainBinary:    consts.Executables.Dymension,
			Type:           consts.SDK_ROLLAPP,
			KeyringBackend: consts.SupportedKeyringBackends.Test,
		}
		eibcKi, err := eibcKc.Info(rollappConfig.Home)
		if err != nil {
			return nil, err
		}
		aki = append(aki, *eibcKi)
	} else {
		pterm.Error.Println("failed to get eibc key", err)
	}

	return aki, nil
}

func (ki *KeyInfo) Print(o ...KeyInfoOption) {
	for _, opt := range o {
		opt(ki)
	}

	if ki.PrintName {
		pterm.DefaultBasicText.Println(pterm.LightGreen(ki.Name))
	}

	fmt.Printf("\t%s\n", ki.Address)

	if ki.PrintPubKey {
		fmt.Printf("\t%s\n", ki.PubKey)
	}
	if ki.PrintMnemonic {
		fmt.Printf("\t%s\n", ki.Mnemonic)
		fmt.Println()
		fmt.Println(pterm.LightYellow("💡 save this information and keep it safe"))
	}

	fmt.Println()
}

func PrintAddressesWithTitle(addresses []KeyInfo) {
	pterm.DefaultSection.WithIndentCharacter("🔑").Println("Addresses")
	for _, address := range addresses {
		address.Print(WithMnemonic(), WithName())
	}
}
