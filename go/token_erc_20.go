package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// TokenERC20Chaincode implements a simple ERC20 token on Hyperledger Fabric
type TokenERC20Chaincode struct {
}

// Token represents an ERC20 token
type Token struct {
	Name     string            `json:"name"`
	Symbol   string            `json:"symbol"`
	Total    uint64            `json:"total"`
	Decimals uint8             `json:"decimals"`
	Balance  map[string]uint64 `json:"balance"`
}

func (t *TokenERC20Chaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

// Initialize the token with name, symbol, total supply, and decimals
func (t *TokenERC20Chaincode) Initialize(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// Check the number of arguments
	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expected 4: name, symbol, total supply, decimals")
	}

	// Retrieve information from the arguments
	name := args[0]
	symbol := args[1]
	totalSupply, err := strconv.ParseUint(args[2], 10, 64)
	if err != nil {
		return shim.Error(fmt.Sprintf("Invalid total supply: %s", err))
	}
	decimals, err := strconv.ParseUint(args[3], 10, 8)
	if err != nil {
		return shim.Error(fmt.Sprintf("Invalid decimals: %s", err))
	}

	// Initialize the token
	token := Token{
		Name:     name,
		Symbol:   symbol,
		Total:    totalSupply,
		Decimals: uint8(decimals),
		Balance:  make(map[string]uint64),
	}

	// Get information of the transaction creator
	creator, err := stub.GetCreator()
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to get transaction creator information: %s", err))
	}

	// Set total supply to the balance of the transaction creator
	token.Balance[hex.EncodeToString(creator)] = totalSupply

	// Save the token state to the ledger
	tokenJSON, err := json.Marshal(token)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to encode token: %s", err))
	}
	err = stub.PutState("token", tokenJSON)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to save state: %s", err))
	}

	return shim.Success(nil)
}

func (t *TokenERC20Chaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()

	switch function {
	case "Initialize":
		return t.Initialize(stub, args)
	case "Mint":
		return t.Mint(stub, args)
	case "ClientAccountBalance":
		return t.ClientAccountBalance(stub)
	case "ClientAccountID":
		return t.ClientAccountID(stub)
	case "transfer":
		return t.transfer(stub, args)
	case "balanceOf":
		return t.balanceOf(stub, args)
	case "name":
		return t.name(stub)
	case "symbol":
		return t.symbol(stub)
	case "totalSupply":
		return t.totalSupply(stub)
	}
	return shim.Error("Invalid function name")
}

// Mint creates new tokens and adds them to the minter's account balance
// This function triggers a Transfer event
func (t *TokenERC20Chaincode) Mint(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// Check number of arguments
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1: amount")
	}

	// Parse amount
	amount, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		return shim.Error(fmt.Sprintf("Invalid amount: %s", err))
	}

	// Load token state
	tokenJSON, err := stub.GetState("token")
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to get token: %s", err))
	}
	var token Token
	err = json.Unmarshal(tokenJSON, &token)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to unmarshal token: %s", err))
	}

	// Add amount to total supply and minter's balance
	creator, err := stub.GetCreator()
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to get creator: %s", err))
	}
	creatorHex := hex.EncodeToString(creator)
	token.Total += amount
	token.Balance[creatorHex] += amount

	// Update token state
	tokenJSON, err = json.Marshal(token)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to marshal token: %s", err))
	}
	err = stub.PutState("token", tokenJSON)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to put state: %s", err))
	}

	// Trigger Transfer event
	err = stub.SetEvent("Transfer", []byte(fmt.Sprintf("Minted %d tokens to %s", amount, creatorHex)))
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to set event: %s", err))
	}

	return shim.Success(nil)
}

// ClientAccountBalance retrieves the account balance of the client's account
func (t *TokenERC20Chaincode) ClientAccountBalance(stub shim.ChaincodeStubInterface) pb.Response {
	// Get client ID
	clientID, err := stub.GetCreator()
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to get client ID: %s", err))
	}

	// Load token state
	tokenJSON, err := stub.GetState("token")
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to get token: %s", err))
	}
	var token Token
	err = json.Unmarshal(tokenJSON, &token)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to unmarshal token: %s", err))
	}
	// Get balance of client ID
	clientIDHex := hex.EncodeToString(clientID)
	balance, exists := token.Balance[clientIDHex]
	if !exists {
		// Initialize balance to 0 if client ID does not exist in map
		balance = 0
		token.Balance[clientIDHex] = balance
		tokenJSON, err := json.Marshal(token)
		if err != nil {
			return shim.Error(fmt.Sprintf("Failed to marshal token: %s", err))
		}
		err = stub.PutState("token", tokenJSON)
		if err != nil {
			return shim.Error(fmt.Sprintf("Failed to update token state: %s", err))
		}
	}

	return shim.Success([]byte(fmt.Sprintf("%d", balance)))
}

// ClientAccountID retrieves the client account ID
func (t *TokenERC20Chaincode) ClientAccountID(stub shim.ChaincodeStubInterface) pb.Response {
	// Get client ID
	clientID, err := stub.GetCreator()
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to get client ID: %s", err))
	}

	// encodedClientID := base64.StdEncoding.EncodeToString(clientID)
	addressHex := hex.EncodeToString([]byte(clientID))
	return shim.Success([]byte(addressHex))
}

// Transfer transfers tokens from client account to recipient account
// recipient account must be a valid clientID as returned by the ClientID() function
// This function triggers a Transfer event
func (t *TokenERC20Chaincode) transfer(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// Check number of arguments
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2: to address and amount")
	}

	// Parse amount
	amount, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		return shim.Error(fmt.Sprintf("Invalid amount: %s", err))
	}

	// Load token state
	tokenJSON, err := stub.GetState("token")
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to get token: %s", err))
	}
	var token Token
	err = json.Unmarshal(tokenJSON, &token)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to unmarshal token: %s", err))
	}

	// Deduct amount from sender's balance
	sender, err := stub.GetCreator()
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to get creator: %s", err))
	}
	senderBalance := token.Balance[string(sender)]
	if senderBalance < amount {
		return shim.Error("Insufficient balance")
	}
	token.Balance[string(sender)] -= amount

	// Add amount to receiver's balance
	receiver := args[0]
	token.Balance[receiver] += amount

	// Update token state
	tokenJSON, err = json.Marshal(token)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to marshal token: %s", err))
	}
	err = stub.PutState("token", tokenJSON)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to put state: %s", err))
	}

	return shim.Success(nil)
}

// BalanceOf returns the balance of the given account
func (t *TokenERC20Chaincode) balanceOf(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// Check number of arguments
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1: to address")
	}
	address := args[0]
	if address == "" {
		return shim.Error("Address argument must be a non-empty string")
	}

	// Load token state
	tokenJSON, err := stub.GetState("token")
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to get token: %s", err))
	}
	if tokenJSON == nil {
		return shim.Error("Token state does not exist")
	}

	var token Token
	err = json.Unmarshal(tokenJSON, &token)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to unmarshal token: %s", err))
	}

	// Get balance of specified address
	balance, exists := token.Balance[address]
	if !exists {
		return shim.Error(fmt.Sprintf("No balance found for address: %s", address))
	}

	return shim.Success([]byte(fmt.Sprintf("%d", balance)))
}

// Name returns a descriptive name for fungible tokens in this contract
// returns {String} Returns the name of the token
func (t *TokenERC20Chaincode) name(stub shim.ChaincodeStubInterface) pb.Response {
	// Load token state
	tokenJSON, err := stub.GetState("token")
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to get token: %s", err))
	}
	var token Token
	err = json.Unmarshal(tokenJSON, &token)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to unmarshal token: %s", err))
	}

	return shim.Success([]byte(token.Name))
}

// Symbol returns an abbreviated name for fungible tokens in this contract.
// returns {String} Returns the symbol of the token
func (t *TokenERC20Chaincode) symbol(stub shim.ChaincodeStubInterface) pb.Response {
	// Load token state
	tokenJSON, err := stub.GetState("token")
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to get token: %s", err))
	}
	var token Token
	err = json.Unmarshal(tokenJSON, &token)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to unmarshal token: %s", err))
	}

	return shim.Success([]byte(token.Symbol))
}

// TotalSupply returns the total token supply
func (t *TokenERC20Chaincode) totalSupply(stub shim.ChaincodeStubInterface) pb.Response {
	// Load token state
	tokenJSON, err := stub.GetState("token")
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to get token: %s", err))
	}
	var token Token
	err = json.Unmarshal(tokenJSON, &token)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to unmarshal token: %s", err))
	}

	return shim.Success([]byte(fmt.Sprintf("%d", token.Total)))
}

func main() {
	err := shim.Start(new(TokenERC20Chaincode))
	if err != nil {
		fmt.Printf("Error starting TokenERC20Chaincode: %s", err)
	}
}