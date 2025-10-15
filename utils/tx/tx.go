package tx

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	cometclient "github.com/cometbft/cometbft/rpc/client/http"
	comettypes "github.com/cometbft/cometbft/types"
	"github.com/gorilla/websocket"
	"github.com/pterm/pterm"
)

type TxResult struct {
	Success   bool
	Code      uint32
	Log       string
	GasWanted int64
	GasUsed   int64
	Source    string // "websocket" or "api"
}

func MonitorTransaction(endpoint, txHash string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := validateEndpoint(endpoint); err != nil {
		newEndpoint, _ := pterm.DefaultInteractiveTextInput.WithDefaultText(
			"The provided endpoint is not working. Please enter another endpoint:",
		).Show()
		endpoint = newEndpoint
	}

	spinner, _ := pterm.DefaultSpinner.WithText(
		fmt.Sprintf("Monitoring transaction %s", pterm.FgYellow.Sprint(txHash)),
	).Start()

	resultChan := make(chan TxResult, 2)
	errChan := make(chan error, 2)
	var wg sync.WaitGroup

	if strings.HasPrefix(endpoint, "http") {
		wg.Add(2)
		go func() {
			defer wg.Done()
			if err := monitorViaWebSocket(ctx, endpoint, txHash, resultChan); err != nil {
				errChan <- fmt.Errorf("websocket error: %v", err)
			}
		}()
		go func() {
			defer wg.Done()
			if err := monitorViaAPI(ctx, endpoint, txHash, resultChan); err != nil {
				errChan <- fmt.Errorf("api error: %v", err)
			}
		}()
	} else if strings.HasPrefix(endpoint, "ws") {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := monitorViaRawWebSocket(ctx, endpoint, txHash, resultChan); err != nil {
				errChan <- fmt.Errorf("websocket error: %v", err)
			}
		}()
	} else {
		return fmt.Errorf("unsupported endpoint type: %s", endpoint)
	}

	go func() {
		wg.Wait()
		close(resultChan)
		close(errChan)
	}()

	// Process results
	var errors []error
	for {
		select {
		case result, ok := <-resultChan:
			if ok && result.Success {
				spinner.Success(fmt.Sprintf("Transaction succeeded via %s!", result.Source))
				pterm.Info.Printf("Gas wanted: %d, Gas used: %d\n", result.GasWanted, result.GasUsed)
				cancel()
				return nil
			} else if ok && !result.Success {
				spinner.Fail(fmt.Sprintf("Transaction failed via %s", result.Source))
				cancel()
				return fmt.Errorf("transaction failed with code %d: %s", result.Code, result.Log)
			}
		case err := <-errChan:
			if err != nil {
				errors = append(errors, err)
			}
		case <-ctx.Done():
			spinner.Fail("Timeout waiting for transaction")
			return fmt.Errorf("timeout waiting for transaction")
		}

		if len(errors) == 2 || (strings.HasPrefix(endpoint, "ws") && len(errors) == 1) {
			spinner.Fail("All monitoring methods failed")
			return fmt.Errorf("all monitoring methods failed: %v", errors)
		}

		if resultChan == nil && errChan == nil {
			break
		}
	}

	spinner.Fail("Transaction monitoring completed without result")
	return fmt.Errorf("transaction monitoring completed without finding transaction")
}

func validateEndpoint(endpoint string) error {
	if strings.HasPrefix(endpoint, "http") {
		return WaitForRPCStatus(fmt.Sprintf("%s/status", endpoint))
	} else if strings.HasPrefix(endpoint, "ws") {
		c, _, err := websocket.DefaultDialer.Dial(endpoint, nil)
		if err != nil {
			return err
		}
		c.Close()
		return nil
	}
	return fmt.Errorf("unsupported endpoint type")
}

func monitorViaWebSocket(ctx context.Context, rpcURL, txHash string, resultChan chan<- TxResult) error {
	client, err := cometclient.New(rpcURL, "/websocket")
	if err != nil {
		return fmt.Errorf("error creating client: %v", err)
	}

	if err := client.Start(); err != nil {
		return fmt.Errorf("error starting client: %v", err)
	}
	defer client.Stop()

	txBytes, err := hex.DecodeString(txHash)
	if err != nil {
		return fmt.Errorf("error decoding txHash: %v", err)
	}

	query := fmt.Sprintf("tm.event='Tx' AND tx.hash='%X'", txBytes)
	subscription, err := client.Subscribe(ctx, "tx-monitor", query, 100)
	if err != nil {
		return fmt.Errorf("error subscribing: %v", err)
	}
	defer client.Unsubscribe(ctx, "tx-monitor", query)

	for {
		select {
		case event := <-subscription:
			txEvent, ok := event.Data.(comettypes.EventDataTx)
			if !ok {
				continue
			}

			result := TxResult{
				Success:   txEvent.Result.Code == 0,
				Code:      txEvent.Result.Code,
				Log:       txEvent.Result.Log,
				GasWanted: txEvent.Result.GasWanted,
				GasUsed:   txEvent.Result.GasUsed,
				Source:    "websocket",
			}
			resultChan <- result
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func monitorViaRawWebSocket(ctx context.Context, wsURL, txHash string, resultChan chan<- TxResult) error {
	c, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to websocket: %v", err)
	}
	defer c.Close()

	txBytes, err := hex.DecodeString(txHash)
	if err != nil {
		return fmt.Errorf("error decoding txHash: %v", err)
	}

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

	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-ctx.Done():
				c.Close()
				return
			case <-done:
				return
			}
		}
	}()

	for {
		var response map[string]interface{}
		if err := c.ReadJSON(&response); err != nil {
			close(done)
			return fmt.Errorf("error reading from websocket: %v", err)
		}

		if result, ok := response["result"].(map[string]interface{}); ok {
			if data, ok := result["data"].(map[string]interface{}); ok {
				if value, ok := data["value"].(map[string]interface{}); ok {
					if txResult, ok := value["TxResult"].(map[string]interface{}); ok {
						code := uint32(0)
						if c, ok := txResult["code"].(float64); ok {
							code = uint32(c)
						}

						log := ""
						if l, ok := txResult["log"].(string); ok {
							log = l
						}

						gasWanted := int64(0)
						if gw, ok := txResult["gas_wanted"].(float64); ok {
							gasWanted = int64(gw)
						}

						gasUsed := int64(0)
						if gu, ok := txResult["gas_used"].(float64); ok {
							gasUsed = int64(gu)
						}

						result := TxResult{
							Success:   code == 0,
							Code:      code,
							Log:       log,
							GasWanted: gasWanted,
							GasUsed:   gasUsed,
							Source:    "websocket",
						}
						close(done)
						resultChan <- result
						return nil
					}
				}
			}
		}

		select {
		case <-ctx.Done():
			close(done)
			return ctx.Err()
		default:
			// nolint:gosec
		}
	}
}

func monitorViaAPI(ctx context.Context, rpcURL, txHash string, resultChan chan<- TxResult) error {
	txHash = strings.ToUpper(txHash)
	if !strings.HasPrefix(txHash, "0X") {
		txHash = "0X" + txHash
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	client := &http.Client{Timeout: 10 * time.Second}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			url := fmt.Sprintf("%s/tx?hash=%s", rpcURL, txHash)
			resp, err := client.Get(url)
			if err != nil {
				continue
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				continue
			}

			var txResponse struct {
				Result struct {
					TxResult struct {
						Code      uint32 `json:"code"`
						Log       string `json:"log"`
						GasWanted string `json:"gas_wanted"`
						GasUsed   string `json:"gas_used"`
					} `json:"tx_result"`
				} `json:"result"`
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				continue
			}

			if err := json.Unmarshal(body, &txResponse); err != nil {
				// Try alternative response format if standard format fails
				var altResponse struct {
					Result struct {
						Code      uint32 `json:"code"`
						Log       string `json:"log"`
						GasWanted string `json:"gas_wanted"`
						GasUsed   string `json:"gas_used"`
					} `json:"result"`
				}
				if err2 := json.Unmarshal(body, &altResponse); err2 == nil {
					// Use alternative format
					txResponse.Result.TxResult.Code = altResponse.Result.Code
					txResponse.Result.TxResult.Log = altResponse.Result.Log
					txResponse.Result.TxResult.GasWanted = altResponse.Result.GasWanted
					txResponse.Result.TxResult.GasUsed = altResponse.Result.GasUsed
				} else {
					continue
				}
			}

			gasWanted := int64(0)
			if gw, err := fmt.Sscanf(txResponse.Result.TxResult.GasWanted, "%d", &gasWanted); err == nil && gw == 1 {
			}

			gasUsed := int64(0)
			if gu, err := fmt.Sscanf(txResponse.Result.TxResult.GasUsed, "%d", &gasUsed); err == nil && gu == 1 {
			}

			if txResponse.Result.TxResult.GasWanted != "" || txResponse.Result.TxResult.GasUsed != "" {
				result := TxResult{
					Success:   txResponse.Result.TxResult.Code == 0,
					Code:      txResponse.Result.TxResult.Code,
					Log:       txResponse.Result.TxResult.Log,
					GasWanted: gasWanted,
					GasUsed:   gasUsed,
					Source:    "api",
				}
				resultChan <- result
				return nil
			}
		}
	}
}

func WaitForRPCStatus(url string) error {
	timeout := time.After(20 * time.Second)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	spinner, _ := pterm.DefaultSpinner.Start("checking rpc status")

	for {
		select {
		case <-timeout:
			spinner.Fail("Timeout: Failed to receive expected response within 20 seconds")
			return fmt.Errorf("timeout")
		case <-ticker.C:
			// nolint:gosec
			_, err := http.Get(url)
			if err != nil {
				fmt.Printf("Error making request: %v\n", err)
				continue
			}
			spinner.Success("RPC endpoint is healthy")
			return nil
		}
	}
}
