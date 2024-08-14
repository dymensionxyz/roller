package consts

type HubData = struct {
	API_URL         string `toml:"api_url"`
	ID              string `toml:"id"`
	RPC_URL         string `toml:"rpc_url"`
	ARCHIVE_RPC_URL string `toml:"archive_rpc_url"`
	GAS_PRICE       string `toml:"gas_price"`
}
