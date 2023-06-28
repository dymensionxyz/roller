package initconfig

import (
	"fmt"
	"github.com/dymensionxyz/roller/cmd/utils"
)

func printInitOutput(addresses []utils.AddressData, rollappId string) {
	fmt.Printf("💈 RollApp '%s' configuration files have been successfully generated on your local machine. Congratulations!\n\n", rollappId)
	utils.PrintAddresses(addresses)
	fmt.Printf("\n🔔 Please fund these addresses to register and run the rollapp.\n")
}
