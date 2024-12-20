## Instructions to Run Avail as a Data Availability (DA) Layer

To build and test the roller with avail as DA follw these steps:

First, install all the necessary dependencies using the following command:

```sh
git clone https://github.com/vitwit/roller.git
cd roller
git fetch
git checkout v1.11.0-alpha-rc03-vw
```
Build the roller 
```bash
make build
```

This command builds the desired version of Roller and places the executable
in the `./build` directory.

To run Roller, use:

```bash
./build/roller
```

#### Steps to Run the Roller:

Follow the steps in the (official doc)https://github.com/vitwit/roller/tree/v1.11.0-alpha-rc03-vw, with the additional adjustments specified below.

init the rollapp

```bash 
./build/roller rollapp init
```
After executing this command modify the `avail.toml` file with appropriate values. This file is located at `$HOME/.roller/da-light-node/avail.toml`. Use the following configuration as an example:
```bash
AccAddress = ""
AppID = 1
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

These are the required changes to verify before starting the Rollapp. Please follow the instructions above carefully before starting the roller.

## Migrate RollApp to another server

To migrate the rollapp to another server follow these instructions

Compress the .roller folder:

```sh
sudo tar -cvf - .roller | lz4 > copyroller.tar.lz4
```

Copy copyroller.tar.lz4 to the new server and extract it:

```sh
scp copyroller.tar.lz4 user@new-server:/path/to/destination
```

```sh
lz4 -c -d copyroller.tar.lz4 | tar -x -C .
```

Modify the file `.roller/rollapp/config/dymint.toml` and change the user of the new server, if necessary:

```sh
keyring_home_dir = "/home/user/.roller/hub-keys"
```

Modify the file `.roller/roller.toml` and change the home path of the new server, if necessary:

```sh
home = "/home/user/.roller"
```

Migrate the dependencies downloaded locally during the roller init command. If the dependency versions differ, ensure that the correct versions are also migrated.

```sh
tar --absolute-names -czvf rollapp-binaries.tar.gz /usr/local/bin/dymd /usr/local/bin/rollapp-evm
```

```sh
 scp rollapp-binaries.tar.gz user@new-server:/path/to/destination
 ```

Extract the binaries on the new server:

 ```sh
 tar --absolute-names -xzvf rollapp-binaries.tar.gz -C /usr/local/bin/
 ```

Verify the updated versions:

 ```sh
 dymd version
 rollapp-evm version
 ```

**Note**: When running the RollApp, the Dymension node to which it is registered must also be migrated when moving from a local server to another server.
Compress .dymension folder:

```sh
sudo tar -cvf - .dymension | lz4 > copydymension.tar.lz4
```

Copy copydymension.tar.lz4 to new server and unzip it:

```sh
scp copydymension.tar.lz4 user@new-server:/path/to/destination
```

```sh
lz4 -c -d copydymension.tar.lz4 | tar -x -C .
```

```sh
dymd start
```

After making the necessary changes, you can start the RollApp using the following command:

```sh
./build/roller rollapp start
```

