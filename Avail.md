## Instructions to Run Avail as a Data Availability (DA) Layer

To register Rollapp using Rollapp-EVM with Avail as the DA layer, follow the instructions provided in the [vitwit/rollapp-evm](https://github.com/vitwit/rollapp-evm/tree/fix_daconfig) repository.


#### Steps to Run the Roller:

Clone the roller repository, switch to the branch with Avail support, and build the project:
```bash
git clone https://github.com/vitwit/roller.git
git fetch && git checkout v1.9.0-vw-main
make build && make install
```

Follow the steps in the (official doc)https://github.com/vitwit/roller/tree/v1.9.0-vw-main, with the additional adjustments specified below.

init the rollapp

```bash 
./build/roller rollapp init
```
After executing this command modify the `avail.toml` file with appropriate values. This file is located at `$HOME/.roller/da-light-node/avail.toml`. Use the following configuration as an example:
```bash
AccAddress = ""
AppID = 1n
Mnemonic = "bottom drive obey lake curtain smoke basket hold race lonely fit walk//Alice"
Root = ""
RpcEndpoint = "ws://127.0.0.1:9944"
```

setup the rollapp

```bash 
./build/roller rollapp setup
```
After executing the above command verify and replace the fields in the `dymint.toml` file, which is located at `$HOME/.roller/rollapp/config/dymint.toml`. The configuration should look like the example below:
```bash
batch_acceptance_attempts = "5"
batch_acceptance_timeout = "2m0s"
batch_submit_bytes = 500000
batch_submit_time = "30s"
block_batch_size = "500"
block_time = "0.2s"
da_config = "{\"seed\": \"bottom drive obey lake curtain smoke basket hold race lonely fit walk//Alice\", \"api_url\": \"ws://127.0.0.1:9944\", \"app_id\": 1, \"tip\":0}"
da_layer = "avail"
dym_account_name = "hub_sequencer"
gas_prices = "2000000000adym"
keyring_backend = "test"
keyring_home_dir = "/root/.roller/hub-keys"
max_idle_time = "20s"
max_proof_time = "5s"
max_skew_time = "168h0m0s"
p2p_advertising_enabled = "false"
p2p_blocksync_block_request_interval = "30s"
p2p_blocksync_enabled = "true"
p2p_bootstrap_nodes = ""
p2p_bootstrap_retry_time = "30s"
p2p_gossip_cache_size = 50
p2p_listen_address = "/ip4/0.0.0.0/tcp/26656"
p2p_persistent_nodes = ""
retry_attempts = "10"
retry_max_delay = "10s"
retry_min_delay = "1s"
settlement_gas_fees = ""
settlement_gas_limit = 0
settlement_gas_prices = "20000000000adym"
settlement_layer = "dymension"
settlement_node_address = "http://localhost:36657"

[db]
  in_memory = false
  sync_writes = true

[instrumentation]
  prometheus = true
  prometheus_listen_addr = ":2112"

```

After making the above changes, you can run the following commands to register and start the services, including the DA light client:
```bash
./build/roller rollapp services load

```

Once all previous steps are complete, start the roller with:

```bash
./build/roller rollapp start
```

These are the required changes to verify before starting the Rollapp. Please follow the instructions above carefully before starting roller.

**Note**: If you encounter the issue of a negative registration fee when starting the roller, update the registration_fee field in the erc20 params section of your genesis.json file as shown below:

```json 
"erc20": {
  "params": {
    "enable_erc20": true,
    "enable_evm_hook": true,
    "registration_fee": "1000000000000000000"
  },
  "token_pairs": []
}
```