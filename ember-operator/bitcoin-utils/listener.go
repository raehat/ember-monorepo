package main

import (
	"context"
	"fmt"
	"log"

	"github.com/block-vision/sui-go-sdk/client"
	"github.com/block-vision/sui-go-sdk/models"

	"github.com/btcsuite/btcd/btcec"
)

var aggregator string

// ------------------------------
// CONFIG
// ------------------------------
var (
	rpcURL      = "https://fullnode.testnet.sui.io"
	packageId   = "0x65f569a89e7c67696c56335e3d3443f6ce9683473d463cf1d1d4164e1062d948"
	mySuiAddress = "0xfed87af74e52f83d49babfd788ac87458d58e4a0b56a6945379e8a3dde8787d8"
)

// ------------------------------
// LISTEN OPERATOR SELECTED EVENT
// ------------------------------
func ListenOperatorSelectedSui(
	rpcURL string,
	packageId string,
	mySuiAddress string,
) {
	cli := client.NewSuiClient(rpcURL)

	filter := models.SuiEventFilter{
		MoveEventType: fmt.Sprintf("%s::EmberManager::OperatorSelected", packageId),
	}

	sub, err := cli.SubscribeEvent(context.Background(), filter)
	if err != nil {
		log.Fatal("❌ Failed to subscribe to Sui events:", err)
	}

	fmt.Println("🟢 Listening for OperatorSelected events on Sui...")

	for {
		select {
		case evt := <-sub:
			selected := evt.ParsedJson["selected_operator"].(string)

			fmt.Println("📢 Operator Selected:", selected)

			if selected == mySuiAddress {
				// Selected → run DKG
				privateKey, _ := btcec.NewPrivateKey(btcec.S256())
				fmt.Println("🔐 Private Key:", privateKey.D)

				shares := SplitSecret(privateKey.D, 5, 3)
				fmt.Println("📦 Shares:", shares)

				// TODO: Send shares to other operators if needed
			} else {
				aggregator = selected
				fmt.Println("🧠 Aggregator set to:", aggregator)
			}
		}
	}
}

// ------------------------------
// FETCH OPERATOR IPs (VIEW CALL)
// ------------------------------
func FetchOperatorIPsSui(
	rpcURL string,
	packageId string,
) (map[string]string, error) {

	cli := client.NewSuiClient(rpcURL)

	resp, err := cli.CallMoveView(context.Background(), models.MoveViewRequest{
		PackageObjectId: packageId,
		Module:          "EmberManager",
		Function:        "get_all_operators",
	})
	if err != nil {
		return nil, err
	}

	operators := resp.Results[0].([]interface{})
	result := make(map[string]string)

	for _, op := range operators {
		addr := op.(string)

		ipResp, err := cli.CallMoveView(context.Background(), models.MoveViewRequest{
			PackageObjectId: packageId,
			Module:          "EmberManager",
			Function:        "get_operator_ip",
			Arguments:       []interface{}{addr},
		})
		if err != nil {
			continue
		}

		ip := ipResp.Results[0].(string)
		result[addr] = ip
	}

	return result, nil
}

// ------------------------------
// FETCH LAST SELECTED OPERATOR
// ------------------------------
func FetchLastSelectedOperatorSui(
	rpcURL string,
	packageId string,
) (string, error) {

	cli := client.NewSuiClient(rpcURL)

	resp, err := cli.CallMoveView(context.Background(), models.MoveViewRequest{
		PackageObjectId: packageId,
		Module:          "EmberManager",
		Function:        "get_last_selected_operator",
	})
	if err != nil {
		return "", err
	}

	return resp.Results[0].(string), nil
}
