package avail

import (
	"fmt"

	daPallet "github.com/availproject/avail-go-sdk/metadata/pallets/data_availability"
	prim "github.com/availproject/avail-go-sdk/primitives"
	"github.com/availproject/avail-go-sdk/sdk"
)

func CreateAppID(rpcURL, seedPhrase, rollappID string) (appID uint32, err error) {
	api, err := sdk.NewSDK(rpcURL)
	if err != nil {
		fmt.Printf("cannot create api:%v", err)
	}

	//If the appID corresponding to the rollappID already exists, it will be reused.
	blockStorage, err := api.Client.StorageAt(prim.None[prim.H256]())
	acc, err := sdk.Account.NewKeyPair(seedPhrase)

	// Fetch Map Storage
	storage := daPallet.StorageAppKeys{}
	val, _ := storage.Fetch(&blockStorage, []byte(rollappID))

	if val.IsSome() {
		value := val.Unwrap().Value
		fmt.Println("AppId: ", value.AppId, "Owner: ", value.Owner.ToHuman())
		return value.AppId, nil
	}

	fmt.Println("Creating Application key...")
	tx := api.Tx.DataAvailability.CreateApplicationKey([]byte(rollappID))
	// Transaction Execution
	res, err := tx.ExecuteAndWatchInclusion(acc, sdk.NewTransactionOptions())
	if err != nil {
		return 0, fmt.Errorf("failed to submit and execute the transaction %v", err)
	}
	if !res.IsSuccessful().IsSome() {
		return 0, fmt.Errorf("failed to determine transaction status")
	}
	if !res.IsSuccessful().Unwrap() {
		return 0, fmt.Errorf("transaction failed")
	}
	if !res.IsSuccessful().IsSome() {
		return 0, fmt.Errorf("failed to decode events")
	}

	events := res.Events.UnsafeUnwrap()
	eventMyb := sdk.EventFindFirst(events, daPallet.EventApplicationKeyCreated{})
	event := eventMyb.UnsafeUnwrap().UnsafeUnwrap()

	appID = event.Id

	return appID, nil
}
