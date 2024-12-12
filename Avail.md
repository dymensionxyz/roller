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
da_config = "{\"seed\": \"bottom drive obey lake curtain smoke basket hold race lonely fit walk//Alice\", \"api_url\": \"ws://127.0.0.1:9944\", \"app_id\": 1, \"tip\":0}"
max_proof_time = "1s"
max_idle_time = "2s"
batch_submit_time = "30s"
batch_acceptance_attempts = "5"
batch_acceptance_timeout = "2m0s"
batch_submit_bytes = 500000
batch_submit_max_time = "1h0m0s"
da_layer = "avail"
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