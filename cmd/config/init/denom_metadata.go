package initconfig

import (
	"fmt"
	"strings"
)

type BankDenomMetadata struct {
	Base        string                  `json:"base"`
	DenomUnits  []BankDenomUnitMetadata `json:"denom_units"`
	Description string                  `json:"description"`
	Display     string                  `json:"display"`
	Name        string                  `json:"name"`
	Symbol      string                  `json:"symbol"`
}

type BankDenomUnitMetadata struct {
	Aliases  []string `json:"aliases"`
	Denom    string   `json:"denom"`
	Exponent uint     `json:"exponent"`
}

func getBankDenomMetadata(denom string, decimals uint) []BankDenomMetadata {
	displayDenom := denom[1:]

	metadata := []BankDenomMetadata{
		{
			Base: denom,
			DenomUnits: []BankDenomUnitMetadata{
				{
					Aliases:  []string{},
					Denom:    denom,
					Exponent: 0,
				},
				{
					Aliases:  []string{},
					Denom:    displayDenom,
					Exponent: decimals,
				},
			},
			Description: fmt.Sprintf("Denom metadata for %s (%s)", displayDenom, denom),
			Display:     displayDenom,
			Name:        displayDenom,
			Symbol:      strings.ToUpper(displayDenom),
		},
	}
	return metadata
}

//
// func createTokenMetadaJSON(metadataPath string, denom string, decimals uint) error {
// 	metadata := getBankDenomMetadata(denom, decimals)
// 	json, err := json.Marshal(metadata)
// 	if err != nil {
// 		return err
// 	}
//
// 	file, err := os.Create(metadataPath)
// 	if err != nil {
// 		return err
// 	}
// 	_, err = file.WriteString(string(json))
// 	if err != nil {
// 		return err
// 	}
//
// 	return nil
// }
