package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
)

var (
	suiAddress      string
	privateKey      string // Sui uses Ed25519 private keys
	suiRPCUrl       = "http://localhost:9000"
	contractObject  = "0xb40d4084ab0fc68ffd184328cd94b1919cbf2ce52c61eb419c009a26e7d196dd" 
)

type SuiTransfer struct {
	Sender    string `json:"sender"`
	Recipient string `json:"recipient"`
	Amount    uint64 `json:"amount"`
}

// Simple JSON-RPC request structure
type RPCRequest struct {
	Jsonrpc string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      int           `json:"id"`
}

type RPCResponse struct {
	Jsonrpc string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result"`
	Error   interface{}     `json:"error,omitempty"`
	ID      int             `json:"id"`
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "✅ Sui Address: %s\n", suiAddress)
}

func suiAddressHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(suiAddress))
}

// Example: call a Move function on Sui
func registerHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("📤 Registering operator on Sui...")

	go func() {
		err := registerOperator()
		if err != nil {
			log.Printf("❌ registerOperator failed: %v\n", err)
		}
	}()

	w.Write([]byte("📤 Registering operator on Sui..."))
}

func registerOperator() error {
	// For demo: we’ll call a dummy Move function using RPC
	reqBody := RPCRequest{
		Jsonrpc: "2.0",
		Method:  "sui_executeTransactionBlock", // Use the appropriate Move function
		Params: []interface{}{
			map[string]interface{}{
				"sender": suiAddress,
				"txBlocks": []interface{}{
					map[string]interface{}{
						"kind": "moveCall",
						"data": map[string]interface{}{
							"packageObjectId": contractObject,
							"module":          "ember_manager",
							"function":        "register_operator",
							"typeArguments":   []interface{}{},
							"arguments":       []interface{}{},
						},
					},
				},
			},
		},
		ID: 1,
	}

	body, _ := json.Marshal(reqBody)
	resp, err := http.Post(suiRPCUrl, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to call Sui RPC: %v", err)
	}
	defer resp.Body.Close()

	var rpcResp RPCResponse
	json.NewDecoder(resp.Body).Decode(&rpcResp)

	if rpcResp.Error != nil {
		return fmt.Errorf("Sui RPC error: %v", rpcResp.Error)
	}

	log.Printf("📤 registerOperator response: %s\n", rpcResp.Result)
	return nil
}

func main() {
	// Demo private key (ed25519) and address
	privateKey = "MC4CAQAwBQYDK2VwBCIEIP8vYc1bZr2c9JqN3eZ0yQmF0lH6xvN3bQ2k7p9s2u1"
	suiAddress = "0x65f569a89e7c67696c56335e3d3443f6ce9683473d463cf1d1d4164e1062d948"

	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port

	fmt.Printf("✅ Sui Address: %s\n", suiAddress)
	fmt.Printf("🌐 Running at http://localhost:%d\n", port)

	http.HandleFunc("/", helloHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/sui-address", suiAddressHandler)

	log.Fatal(http.Serve(listener, nil))
}
