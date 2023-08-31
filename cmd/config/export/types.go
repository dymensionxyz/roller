package export

type IbcConfig struct {
	HubChannel    string   `json:"hubChannel"`
	Channel       string   `json:"channel"`
	Timeout       int      `json:"timeout"`
	EnabledTokens []string `json:"enabledTokens,omitempty"`
}

type EvmConfig struct {
	ChainId string `json:"chainId"`
	Rpc     string `json:"rpc"`
}

type NetworkType string

const (
	Hub     NetworkType = "Hub"
	RollApp NetworkType = "RollApp"
	Regular NetworkType = "Regular"
)

type DataAvailability string

const (
	Celestia DataAvailability = "Celestia"
	Avail    DataAvailability = "Avail"
)

type GasPriceSteps struct {
	Low     int `json:"low"`
	Average int `json:"average"`
	High    int `json:"high"`
}

type App struct {
	Name string `json:"name"`
	Url  string `json:"url"`
	Logo string `json:"logo"`
}

type Currency struct {
	DisplayDenom string `json:"displayDenom"`
	BaseDenom    string `json:"baseDenom"`
	Decimals     uint   `json:"decimals"`
	Logo         string `json:"logo"`
	CurrencyType string `json:"type"`
}

type NetworkJson struct {
	ChainId      string     `json:"chainId"`
	ChainName    string     `json:"chainName"`
	Rpc          string     `json:"rpc"`
	Rest         string     `json:"rest"`
	Bech32Prefix string     `json:"bech32Prefix"`
	Currencies   []Currency `json:"currencies"`
	//NativeCurrency            string           `json:"nativeCurrency"`
	//StakeCurrency             string           `json:"stakeCurrency"`
	//FeeCurrency               string           `json:"feeCurrency"`
	GasPriceSteps             *GasPriceSteps   `json:"gasPriceSteps,omitempty"`
	GasAdjustment             *float64         `json:"gasAdjustment,omitempty"`
	CoinType                  int              `json:"coinType"`
	ExplorerUrl               *string          `json:"explorerUrl,omitempty"`
	ExploreTxUrl              *string          `json:"exploreTxUrl,omitempty"`
	FaucetUrl                 string           `json:"faucetUrl,omitempty"`
	Website                   string           `json:"website,omitempty"`
	ValidatorsLogosStorageDir *string          `json:"validatorsLogosStorageDir,omitempty"`
	Logo                      string           `json:"logo"`
	Disabled                  *bool            `json:"disabled,omitempty"`
	Custom                    *bool            `json:"custom,omitempty"`
	Ibc                       IbcConfig        `json:"ibc"`
	Evm                       *EvmConfig       `json:"evm,omitempty"`
	Type                      NetworkType      `json:"type"`
	Da                        DataAvailability `json:"da,omitempty"`
	Apps                      []App            `json:"apps,omitempty"`
	Description               *string          `json:"description,omitempty"`
	IsValidator               *bool            `json:"isValidator,omitempty"`
	Analytics                 bool             `json:"analytics"`
}
