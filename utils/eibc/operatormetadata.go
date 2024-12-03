package eibc

import (
	"encoding/json"

	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
)

// EibcOperatorMetadata struct represents the metadata for the eibc operator which is associated with the
// eibc group and delegation policy
type EibcOperatorMetadata struct {
	Moniker           string                     `json:"moniker"`
	Description       string                     `json:"description"`
	ContactDetails    EibcOperatorContactDetails `json:"contact_details"`
	FeeShare          float64                    `json:"fee_share"`
	SupportedRollapps []string                   `json:"supported_rollapps"`
}

// EibcOperatorContactDetails struct represents the contact details for the eibc operator
type EibcOperatorContactDetails struct {
	X        string `json:"x"`
	Website  string `json:"website"`
	Telegram string `json:"telegram"`
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
	moniker, _ := pterm.DefaultInteractiveTextInput.WithDefaultText(
		"provide a moniker for the eibc operator",
	).Show()

	metadata := &EibcOperatorMetadata{
		Moniker:           moniker,
		ContactDetails:    EibcOperatorContactDetails{},
		FeeShare:          consts.DefaultEibcOperatorFeeShare,
		SupportedRollapps: raIDs,
	}

	shouldFillOptionalFields, _ := pterm.DefaultInteractiveConfirm.WithDefaultText(
		"Would you also like to fill optional metadata for your sequencer?",
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
