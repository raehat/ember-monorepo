package main

import (
	"context"
	"fmt"
	"log"

	"github.com/block-vision/sui-go-sdk/client"
	"github.com/block-vision/sui-go-sdk/models"
)

var (
	rpcURL        = "https://fullnode.testnet.sui.io"
	packageId     = "0x65f569a89e7c67696c56335e3d3443f6ce9683473d463cf1d1d4164e1062d948"
	managerObject = "0xb40d4084ab0fc68ffd184328cd94b1919cbf2ce52c61eb419c009a26e7d196dd"
	privateKey = "AKb3Z9xT4cWqL9y2pR7mN1sE5dQ8hV0f3uB2JxK9LmP0"
	myAddress  = "0x8f3a2b9c7d1e5f4a6b2c9d0e1f3a5b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3"
)

// ------------------------------
// MAIN
// ------------------------------
func main() {
	fmt.Println("🚀 Listening for new loans on Sui...")

	go ListenNewLoansAndFulfill()

	select {} // keep running
}

// ------------------------------
// LISTEN FOR NEW LOANS EVENT
// ------------------------------
func ListenNewLoansAndFulfill() {
	cli := client.NewSuiClient(rpcURL)

	filter := models.SuiEventFilter{
		MoveEventType: fmt.Sprintf("%s::EmberManager::LoanCreated", packageId),
	}

	sub, err := cli.SubscribeEvent(context.Background(), filter)
	if err != nil {
		log.Fatal("❌ Failed to subscribe to loan events:", err)
	}

	fmt.Println("🟢 Waiting for new loans...")

	for {
		select {
		case evt := <-sub:
			loanId := uint64(evt.ParsedJson["loan_id"].(float64))
			borrower := evt.ParsedJson["borrower"].(string)
			amount := uint64(evt.ParsedJson["amount"].(float64))

			fmt.Println("📢 New Loan Detected")
			fmt.Println("   Loan ID:", loanId)
			fmt.Println("   Borrower:", borrower)
			fmt.Println("   Amount:", amount)

			go FulfillLoan(loanId)
		}
	}
}

// ------------------------------
// FULFILL LOAN
// ------------------------------
func FulfillLoan(loanId uint64) {
	fmt.Println("💸 Fulfilling Loan:", loanId)

	cli := client.NewSuiClient(rpcURL)

	signer, err := models.NewSignerFromPrivateKey(privateKey)
	if err != nil {
		log.Println("❌ Signer error:", err)
		return
	}

	// TODO: Replace with real lender hash secrets
	hashSecretsLender := make([][]byte, 14)
	for i := 0; i < 14; i++ {
		hashSecretsLender[i] = []byte(fmt.Sprintf("secret_%d", i))
	}

	// Fetch largest SUI coin to fund loan
	coins, err := cli.GetCoins(context.Background(), myAddress, "0x2::sui::SUI", nil, 10)
	if err != nil || len(coins.Data) == 0 {
		log.Println("❌ No SUI coins found")
		return
	}

	coinObject := coins.Data[0].CoinObjectId

	tx := models.MoveCallRequest{
		PackageObjectId: packageId,
		Module:          "EmberManager",
		Function:        "fulfill_loan",
		TypeArguments:   []string{},
		Arguments: []interface{}{
			managerObject,
			loanId,
			myAddress,
			hashSecretsLender,
			coinObject,
		},
		GasBudget: 60000000,
	}

	resp, err := cli.MoveCall(context.Background(), signer, tx)
	if err != nil {
		log.Println("❌ Fulfill failed:", err)
		return
	}

	fmt.Println("✅ Loan Fulfilled TX:", resp.Digest)
}
