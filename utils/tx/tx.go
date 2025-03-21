package tx

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	cometclient "github.com/cometbft/cometbft/rpc/client/http"
	comettypes "github.com/cometbft/cometbft/types"
	"github.com/gorilla/websocket"
	"github.com/pterm/pterm"
)

func MonitorTransaction(wsURL, txHash string) error {
	if strings.HasPrefix(wsURL, "http") {
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
	} else {
		for {
			c, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if err != nil {
				fmt.Printf("❌ WebSocket %s is not responding\n", wsURL)

				newWS, _ := pterm.DefaultInteractiveTextInput.WithDefaultText(
					"WebSocket is not working. Please enter a new WebSocket URL (wss:// or ws://):",
				).Show()

				wsURL = newWS
				continue
			}
			defer c.Close()
			fmt.Println("✅ WebSocket is working!")

			txBytes, err := hex.DecodeString(txHash)
			if err != nil {
				return fmt.Errorf("error decoding txHash: %v", err)
			}

			// Create a query to filter transactions
			query := fmt.Sprintf("tm.event='Tx' AND tx.hash='%X'", txBytes)

			subscribeMsg := map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      "1",
				"method":  "subscribe",
				"params":  []interface{}{query},
			}

			if err := c.WriteJSON(subscribeMsg); err != nil {
				return fmt.Errorf("failed to send subscribe request: %v", err)
			}

			fmt.Println("Monitoring transaction:", txHash)

			spinner, _ := pterm.DefaultSpinner.WithText(
				fmt.Sprintf("Waiting for transaction %s to finalize...", pterm.FgYellow.Sprint(txHash)),
			).Start()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			for {
				select {
				case <-ctx.Done():
					return fmt.Errorf("timeout waiting for transaction")

				default:
					var event struct {
						DataTx comettypes.EventDataTx `json:"result"`
					}

					_, message, err := c.ReadMessage()
					if err != nil {
						return fmt.Errorf("error reading from WebSocket: %v", err)
					}

					if err := json.Unmarshal(message, &event); err != nil {
						fmt.Println("⚠️ Error parsing response:", err)
						continue
					}

					if event.DataTx.Result.Code == 0 {
						spinner.Success("✅ Transaction succeeded!")
						fmt.Printf("Gas Wanted: %d, Gas Used: %d\n", event.DataTx.Result.GasWanted, event.DataTx.Result.GasUsed)
						return nil
					} else {
						fmt.Printf("❌ Transaction failed: %s\n", event.DataTx.Result.Log)
						return fmt.Errorf("transaction failed with code %d: %v", event.DataTx.Result.Code, event.DataTx.Result.Log)
					}
				}
			}
		}
	}
}
