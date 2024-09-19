package types

import (
	"github.com/dymensionxyz/roller/utils/config"
	queryutils "github.com/dymensionxyz/roller/utils/query"
	"github.com/spf13/cobra"
)

type Config struct {
	CobraCmd *cobra.Command // used to retrieve flags if needed

	// start info
	IP   string
	Port uint16

	// application config
	HubQueryClients *queryutils.HubQueryClients
	RollerHome      string
	RollappConfig   config.RollappConfig

	WhaleAccount string
}
