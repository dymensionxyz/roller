package tx

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	cometclient "github.com/cometbft/cometbft/rpc/client/http"
	comettypes "github.com/cometbft/cometbft/types"
	"github.com/pterm/pterm"
)

func MonitorTransaction(wsURL, txHash string) error {
	for {
		_, err := http.Get(fmt.Sprintf("%s/status", wsURL))
		if err == nil {
			fmt.Println("✅ RPC is working!")
			break
		}
		fmt.Printf("❌ RPC %s is not responding: %v\n", wsURL, err)

		newRPC, _ := pterm.DefaultInteractiveTextInput.WithDefaultText(
			"the provided Hub RPC is not working, please enter another RPC Endpoint instead (you can be obtained in the following link https://blastapi.io/chains/dymension)",
		).Show()

		wsURL = newRPC
	}

	// Create a new client
	client, err := cometclient.New(wsURL, "/websocket")
	if err != nil {
		return fmt.Errorf("error creating client: %v", err)
	}

	// Start the client
	err = client.Start()
	if err != nil {
		return fmt.Errorf("error starting client: %v", err)
	}

	// nolint errcheck
	defer client.Stop()

	// Convert txHash string to bytes
	txBytes, err := hex.DecodeString(txHash)
	if err != nil {
		return fmt.Errorf("error decoding txHash: %v", err)
	}

	// Create a query to filter transactions
	query := fmt.Sprintf("tm.event='Tx' AND tx.hash='%X'", txBytes)

	// Subscribe to the query
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	// nolint errcheck
	defer cancel()

	subscription, err := client.Subscribe(ctx, "tx-monitor", query, 100)
	if err != nil {
		return fmt.Errorf("error subscribing: %v", err)
	}
	// nolint errcheck
	defer client.Unsubscribe(ctx, "tx-monitor", query)

	fmt.Println("Monitoring transaction:", txHash)

	spinner, _ := pterm.DefaultSpinner.WithText(
		fmt.Sprintf(
			"waiting for tx with hash %s to finalize",
			pterm.FgYellow.Sprint(txHash),
		),
	).Start()

	// Listen for events
	for {
		select {
		case event := <-subscription:
			txEvent, ok := event.Data.(comettypes.EventDataTx)
			if !ok {
				fmt.Println("Received non-tx event")
				continue
			}

			if txEvent.Result.Code == 0 {
				spinner.Success("transaction succeeded")
				pterm.Info.Printf(
					"Gas wanted: %d, Gas used: %d\n",
					txEvent.Result.GasWanted,
					txEvent.Result.GasUsed,
				)
				return nil
			} else {
				j, _ := json.MarshalIndent(txEvent.Result, "", " ")
				fmt.Println(string(j))

				return fmt.Errorf("transaction failed with code %d: %v", txEvent.Result.Code, txEvent.Result.Log)
			}
		case <-time.After(5 * time.Minute):
			return fmt.Errorf("timeout waiting for transaction")

		case <-ctx.Done():
			return fmt.Errorf("context cancelled")
		}
	}
}
