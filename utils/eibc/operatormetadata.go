package eibc

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pterm/pterm"
	"gopkg.in/yaml.v3"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/tx"
)

// TODO: refactor everything here, it'll be a lot easier and cleaner to use io.Reader/Writer + []byte

// EibcOperatorMetadata struct represents the metadata for the eibc operator which is associated with the
// eibc group and delegation policy
type EibcOperatorMetadata struct {
	Moniker           string                     `json:"moniker"            yaml:"moniker"`
	Description       string                     `json:"description"        yaml:"description"`
	ContactDetails    EibcOperatorContactDetails `json:"contact_details"    yaml:"contact_details"`
	PolicyAddress     string                     `json:"policy_address"     yaml:"policy_address"`
	FeeShare          float64                    `json:"fee_share"          yaml:"fee_share"`
	SupportedRollapps []string                   `json:"supported_rollapps" yaml:"supported_rollapps"`
}

// EibcOperatorContactDetails struct represents the contact details for the eibc operator
type EibcOperatorContactDetails struct {
	X        string `json:"x"        yaml:"x"`
	Website  string `json:"website"  yaml:"website"`
	Telegram string `json:"telegram" yaml:"telegram"`
}

// ToBytes converts EibcOperatorMetadata to []byte
func (m *EibcOperatorMetadata) ToBytes() ([]byte, error) {
	return json.Marshal(m)
}

// WithDescription sets the description and returns the modified EibcOperatorMetadata
func (m *EibcOperatorMetadata) WithDescription(description string) *EibcOperatorMetadata {
	m.Description = description
	return m
}

// WithX sets the X (Twitter) contact detail and returns the modified EibcOperatorMetadata
func (m *EibcOperatorMetadata) WithX(x string) *EibcOperatorMetadata {
	m.ContactDetails.X = x
	return m
}

// WithWebsite sets the website contact detail and returns the modified EibcOperatorMetadata
func (m *EibcOperatorMetadata) WithWebsite(website string) *EibcOperatorMetadata {
	m.ContactDetails.Website = website
	return m
}

// WithTelegram sets the telegram contact detail and returns the modified EibcOperatorMetadata
func (m *EibcOperatorMetadata) WithTelegram(telegram string) *EibcOperatorMetadata {
	m.ContactDetails.Telegram = telegram
	return m
}

// WithFeeShare sets the fee share and returns the modified EibcOperatorMetadata
func (m *EibcOperatorMetadata) WithFeeShare(feeShare float64) *EibcOperatorMetadata {
	m.FeeShare = feeShare
	return m
}

// WithSupportedRollapps sets the supported rollapps and returns the modified EibcOperatorMetadata
func (m *EibcOperatorMetadata) WithSupportedRollapps(rollapps []string) *EibcOperatorMetadata {
	m.SupportedRollapps = rollapps
	return m
}

// NewEibcOperatorMetadata function creates a new EibcOperatorMetadata instance	with the provided rollapp ID
// and prompts the user for the required information
func NewEibcOperatorMetadata(raIDs []string) *EibcOperatorMetadata {
	pterm.Info.Println(
		"the information provided below will be associated with the eibc group and delegation policy",
	)

	var moniker string
	for {
		moniker, _ = pterm.DefaultInteractiveTextInput.WithDefaultText(
			"provide a moniker for the eibc operator",
		).Show()

		if strings.TrimSpace(moniker) != "" {
			break
		}
	}

	metadata := &EibcOperatorMetadata{
		Moniker:           moniker,
		ContactDetails:    EibcOperatorContactDetails{},
		FeeShare:          consts.DefaultEibcOperatorFeeShare,
		SupportedRollapps: raIDs,
	}

	shouldFillOptionalFields, _ := pterm.DefaultInteractiveConfirm.WithDefaultText(
		"Would you also like to fill optional metadata for your eibc operator?",
	).Show()

	if shouldFillOptionalFields {
		description, _ := pterm.DefaultInteractiveTextInput.WithDefaultText(
			"provide a description for the eibc operator (leave empty to skip)",
		).Show()

		x, _ := pterm.DefaultInteractiveTextInput.WithDefaultText(
			"provide a link to your X (leave empty to skip)",
		).Show()
		website, _ := pterm.DefaultInteractiveTextInput.WithDefaultText(
			"provide a link to your website (leave empty to skip)",
		).Show()
		telegram, _ := pterm.DefaultInteractiveTextInput.WithDefaultText(
			"provide a link to your telegram (leave empty to skip)",
		).Show()

		if description != "" {
			metadata.WithDescription(description)
		}

		if x != "" {
			metadata.WithX(x)
		}

		if website != "" {
			metadata.WithWebsite(website)
		}

		if telegram != "" {
			metadata.WithTelegram(telegram)
		}
	}

	return metadata
}

func getUpdateEibcOperatorMetadataCmd(
	eibcHome string,
	adminAddr, groupID, metadata string,
	hd consts.HubData,
) *exec.Cmd {
	cmd := exec.Command(
		consts.Executables.Dymension,
		"tx",
		"group",
		"update-group-metadata",
		adminAddr,
		groupID,
		metadata,
		"--home",
		eibcHome,
		"--node",
		hd.RpcUrl,
		"--chain-id",
		hd.ID,
		"--keyring-backend",
		"test",
		"--fees",
		fmt.Sprintf("%d%s", consts.DefaultTxFee, consts.Denoms.Hub),
		"-y",
	)

	return cmd
}

func UpdateEibcOperatorMetadata(home, metadata string, hd consts.HubData) error {
	eibcHome := filepath.Join(home, consts.ConfigDirName.Eibc)
	kc := GetKeyConfig()
	ki, err := kc.Info(home)
	if err != nil {
		return err
	}

	gid, err := GetGroups(eibcHome, ki.Address, hd)
	if err != nil {
		return err
	}

	c := getUpdateEibcOperatorMetadataCmd(eibcHome, ki.Address, gid.Groups[0].ID, metadata, hd)

	out, err := bash.ExecCommandWithStdout(c)
	if err != nil {
		pterm.Error.Println("failed to create group: ", err)
		return err
	}

	txHash, err := bash.ExtractTxHash(out.String())
	if err != nil {
		pterm.Error.Println("failed to extract tx hash: ", err)
		return err
	}

	err = tx.MonitorTransaction(hd.RpcUrl, txHash)
	if err != nil {
		return err
	}

	return nil
}

func EibcOperatorMetadataFromChain(
	home string,
	hd consts.HubData,
) (*EibcOperatorMetadata, error) {
	eibcHome := filepath.Join(home, consts.ConfigDirName.Eibc)
	kc := GetKeyConfig()

	ki, err := kc.Info(home)
	if err != nil {
		return nil, err
	}

	pol, err := GetGroups(eibcHome, ki.Address, hd)
	if err != nil {
		return nil, err
	}

	if pol.Groups[0].Metadata != "" {
		metadataB64 := pol.Groups[0].Metadata
		var m EibcOperatorMetadata
		metadata, err := base64.StdEncoding.DecodeString(metadataB64)
		if err != nil {
			pterm.Warning.Println("not base 64 decodeable")
			m.Moniker = pol.Groups[0].Metadata

			return &m, nil
		}

		err = json.Unmarshal(metadata, &m)
		if err != nil {
			return nil, err
		}
		return &m, nil
	}

	raIDs, err := LoadSupportedRollapps(filepath.Join(eibcHome, "config.yaml"))
	if err != nil {
		return nil, err
	}
	m := NewEibcOperatorMetadata(raIDs)
	return m, nil
}

// UpdateGroupSupportedRollapps function updates the supported rollapps list in the onchain metadata
// of group and group-policy and returns an error if any
func UpdateGroupSupportedRollapps(eibcConfigPath string, cfg Config, home string) error {
	rspn, _ := pterm.DefaultSpinner.Start("updating eibc operator metadata")
	rspn.UpdateText("retrieving updated supported rollapp list")
	raIDs, err := LoadSupportedRollapps(eibcConfigPath)
	if err != nil {
		pterm.Error.Println("failed to load supported rollapps: ", err)
		return err
	}
	hd, err := cfg.HubDataFromHubRpc(eibcConfigPath)
	if err != nil {
		pterm.Error.Println("failed to retrieve hub data: ", err)
		return err
	}

	rspn.UpdateText("retrieving existing eibc operator metadata")
	metadata, err := EibcOperatorMetadataFromChain(home, *hd)
	if err != nil {
		pterm.Error.Println("failed to retrieve eibc operator metadata: ", err)
		return err
	}
	rspn.UpdateText("updating supported rollapp list")
	metadata.SupportedRollapps = raIDs

	mb, err := metadata.ToBytes()
	if err != nil {
		pterm.Error.Println("failed to generate eibc operator metadata: ", err)
		return err
	}
	mbs := base64.StdEncoding.EncodeToString(mb)

	rspn.UpdateText("pushing changes to chain")
	err = UpdateEibcOperatorMetadata(home, mbs, *hd)
	if err != nil {
		pterm.Error.Println("failed to update eibc operator metadata: ", err)
		return err
	}
	rspn.Success("operator metadata updated, new metadata:")
	ym, err := yaml.Marshal(metadata)
	if err != nil {
		pterm.Error.Println("failed to marshal eibc operator metadata: ", err)
		return err
	}
	fmt.Println(string(ym))
	return nil
}

// UpdateGroupSupportedRollapps function updates the supported rollapps list in the onchain metadata
// of group and group-policy and returns an error if any
func UpdateGroupOperatorMinFee(
	eibcConfigPath string,
	feeShare float64,
	cfg Config,
	home string,
) error {
	rspn, _ := pterm.DefaultSpinner.Start("updating eibc operator metadata")
	rspn.UpdateText("retrieving updated supported rollapp list")
	hd, err := cfg.HubDataFromHubRpc(eibcConfigPath)
	if err != nil {
		pterm.Error.Println("failed to retrieve hub data: ", err)
		return err
	}

	rspn.UpdateText("retrieving existing eibc operator metadata")
	metadata, err := EibcOperatorMetadataFromChain(home, *hd)
	if err != nil {
		pterm.Error.Println("failed to retrieve eibc operator metadata: ", err)
		return err
	}

	rspn.UpdateText("updating supported rollapp list")
	metadata.FeeShare = feeShare

	mb, err := metadata.ToBytes()
	if err != nil {
		pterm.Error.Println("failed to generate eibc operator metadata: ", err)
		return err
	}
	mbs := base64.StdEncoding.EncodeToString(mb)

	rspn.UpdateText("pushing changes to chain")
	err = UpdateEibcOperatorMetadata(home, mbs, *hd)
	if err != nil {
		pterm.Error.Println("failed to update eibc operator metadata: ", err)
		return err
	}
	rspn.Success("operator metadata updated, new metadata:")
	ym, err := yaml.Marshal(metadata)
	if err != nil {
		pterm.Error.Println("failed to marshal eibc operator metadata: ", err)
		return err
	}
	fmt.Println(string(ym))
	return nil
}
