package config

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"os"
	"regexp"
	"strings"

	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/filesystem"
)

const (
	lowerCase   = "abcdefghijklmnopqrstuvwxyz"
	upperCase   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numbers     = "0123456789"
	specialChar = "!@#$%^&*()_-+={}[/?]"
)

func CreateCustomHubData() (*consts.HubData, error) {
	opts := []string{"from-file", "manual"}
	opt, _ := pterm.DefaultInteractiveSelect.WithDefaultText(
		"select how you want to provide the hub data",
	).WithOptions(opts).Show()

	var hd consts.HubData

	switch opt {
	case "from-file":
		return createCustomHubDataFromFile()
	case "manual":
		hd = createCustomHubDataManually()
	}

	return &hd, nil
}

func createCustomHubDataManually() consts.HubData {
	id, _ := pterm.DefaultInteractiveTextInput.WithDefaultText("provide hub chain id").Show()
	rpcUrl, _ := pterm.DefaultInteractiveTextInput.WithDefaultText(
		"provide hub rpc endpoint (including port, example: http://dym.dev:26657)",
	).Show()
	restUrl, _ := pterm.DefaultInteractiveTextInput.WithDefaultText(
		"provide hub rest api endpoint (including port, example: http://dym.dev:1318)",
	).Show()
	gasPrice, _ := pterm.DefaultInteractiveTextInput.WithDefaultText("provide gas price").
		WithDefaultValue("2000000000").Show()

	id = strings.TrimSpace(id)
	rpcUrl = strings.TrimSpace(rpcUrl)
	restUrl = strings.TrimSpace(restUrl)
	gasPrice = strings.TrimSpace(gasPrice)

	hd := consts.HubData{
		Environment:   "custom",
		ApiUrl:        restUrl,
		ID:            id,
		RpcUrl:        rpcUrl,
		ArchiveRpcUrl: rpcUrl,
		GasPrice:      gasPrice,
		DaNetwork:     consts.CelestiaTestnet,
	}
	return hd
}

func createCustomHubDataFromFile() (*consts.HubData, error) {
	pterm.Info.Printf("provide a path to a json file that has the following structure")
	fmt.Println(`
{
  "id": "<hub-id>",
  "rpcUrl": "<hub-rpc-endpoint>",
  "apiUrl": "<hub-rest-endpoint>",
}`)
	path, _ := pterm.DefaultInteractiveTextInput.WithDefaultText("").Show()
	for len(path) == 0 {
		path, _ = pterm.DefaultInteractiveTextInput.WithDefaultText(
			"provide a path to a json file that has the following structure",
		).Show()
	}
	ep, err := filesystem.ExpandHomePath(path)
	if err != nil {
		return nil, err
	}
	path = ep

	jsonFile, err := os.Open(path)
	if err != nil {
		log.Fatalf("Failed to open JSON file: %s", err)
	}
	//nolint:errcheck
	defer jsonFile.Close()

	// Read file content into a byte slice
	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		log.Fatalf("Failed to read JSON file: %s", err)
	}

	// Unmarshal the byte slice into the Config struct
	var hd consts.HubData
	err = json.Unmarshal(byteValue, &hd)
	if err != nil {
		log.Fatalf("Failed to unmarshal JSON: %s", err)
	}
	hd.DaNetwork = consts.CelestiaTestnet
	hd.GasPrice = "2000000000"
	hd.ArchiveRpcUrl = hd.RpcUrl
	hd.Environment = "custom"

	return &hd, nil
}

func generateSecurePassword() ([]byte, error) {
	charSet := lowerCase + upperCase + numbers + specialChar
	password := make([]byte, 24)

	for i := range password {
		randNum, err := rand.Int(rand.Reader, big.NewInt(int64(len(charSet))))
		if err != nil {
			return nil, err
		}
		password[i] = charSet[randNum.Int64()]
	}

	return password, nil
}

func WritePasswordToFile(path string) error {
	p, err := generateSecurePassword()
	if err != nil {
		return err
	}

	err = os.WriteFile(path, p, 0o755)
	if err != nil {
		return err
	}

	return nil
}

func IsValidURL(url string) bool {
	regex := `^(https?:\/\/)?([\da-z\.-]+)\.([a-z\.]{2,6})([\/\w \.-]*)*\/?$`
	re := regexp.MustCompile(regex)
	return re.MatchString(url)
}
