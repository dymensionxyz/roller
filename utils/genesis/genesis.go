package genesis

import "github.com/dymensionxyz/roller/cmd/utils"

type AppState struct {
	Bank Bank `json:"bank"`
}

type Bank struct {
	Supply []utils.Balance `json:"supply"`
}
