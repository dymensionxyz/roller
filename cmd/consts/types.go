package consts

type HubData struct {
	Environment   string `toml:"environment"     json:"environment"`
	ApiUrl        string `toml:"api_url"         json:"apiUrl"`
	ID            string `toml:"id"              json:"id"`
	RpcUrl        string `toml:"rpc_url"         json:"rpcUrl"`
	WsUrl         string `toml:"ws_url"          json:"wsUrl"`
	ArchiveRpcUrl string `toml:"archive_rpc_url" json:"archiveRpcUrl"`
	GasPrice      string `toml:"gas_price"       json:"gasPrice"`
}

type RollappData = struct {
	ID       string `toml:"id"        yaml:"id"`
	ApiUrl   string `toml:"api_url"`
	RpcUrl   string `toml:"rpc_url"`
	GasPrice string `toml:"gas_price"`
	Denom    string `toml:"denom"`
}

type DaData = struct {
	Backend DAType    `toml:"backend"`
	ID      DaNetwork `toml:"id"`
	ApiUrl  string    `toml:"api_url"`
	RpcUrl  string    `toml:"rpc_url"`
	// TODO: combine CurrentStateNode and StateNodes
	CurrentStateNode string   `toml:"current_state_node"`
	StateNodes       []string `toml:"state_nodes"`
	GasPrice         string   `toml:"gas_price"`
}
