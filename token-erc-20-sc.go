package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

// Define key names for options
const (
	nameKey       = "name"
	symbolKey     = "symbol"
	decimalsKey   = "decimals"
	totalSupplyKey = "totalSupply"
)

// Define SmartContract structure
type SmartContract struct {
}

// Define Token structure
type Token struct {
	Name        string `json:"name"`
	Symbol      string `json:"symbol"`
	Decimals    int    `json:"decimals"`
	TotalSupply int    `json:"totalSupply"`
}

// Define event structure
type Event struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Value int    `json:"value"`
}

// Init initializes chaincode
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

// Invoke - Our entry point for Invocations
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {
	function, args := APIstub.GetFunctionAndParameters()
	switch function {
	case "Initialize":
		return s.Initialize(APIstub, args)
	case "Mint":
		return s.Mint(APIstub, args)
	case "Burn":
		return s.Burn(APIstub, args)
	case "Transfer":
		return s.Transfer(APIstub, args)
	case "BalanceOf":
		return s.BalanceOf(APIstub, args)
  case "ClientAccountBalance":
		return s.ClientAccountBalance(APIstub, args)
	case "ClientAccountID":
		return s.ClientAccountID(APIstub, args)
	case "TotalSupply":
		return s.TotalSupply(APIstub, args)
	case "Approve":
		return s.Approve(APIstub, args)
	case "Allowance":
		return s.Allowance(APIstub, args)
	case "TransferFrom":
		return s.TransferFrom(APIstub, args)
	default:
		return shim.Error("Invalid function name")
	}
}

// Initialize initializes the token's state (name, symbol, decimals, totalSupply)
func (s *SmartContract) Initialize(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}

	name := args[0]
	symbol := args[1]
	decimals, err := strconv.Atoi(args[2])
	if err != nil {
		return shim.Error("Invalid decimals. Expecting a numeric string")
	}
	totalSupply, err := strconv.Atoi(args[3])
	if err != nil {
		return shim.Error("Invalid total supply. Expecting a numeric string")
	}

	token := Token{
		Name:        name,
		Symbol:      symbol,
		Decimals:    decimals,
		TotalSupply: totalSupply,
	}

	tokenBytes, err := json.Marshal(token)
	if err != nil {
		return shim.Error("Failed to marshal token data")
	}

	err = APIstub.PutState(nameKey, []byte(name))
	if err != nil {
		return shim.Error("Failed to set token name")
	}

	err = APIstub.PutState(symbolKey, []byte(symbol))
	if err != nil {
		return shim.Error("Failed to set token symbol")
	}

	err = APIstub.PutState(decimalsKey, []byte(strconv.Itoa(decimals)))
	if err != nil {
		return shim.Error("Failed to set token decimals")
	}

	err = APIstub.PutState(totalSupplyKey, []byte(strconv.Itoa(totalSupply)))
	if err != nil {
		return shim.Error("Failed to set token total supply")
	}

	return shim.Success(tokenBytes)
}

// Mint creates new tokens and adds them to minter's account balance
func (s *SmartContract) Mint(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	minter := args[0]
	amount, err := strconv.Atoi(args[1])
	if err != nil {
		return shim.Error("Invalid amount. Expecting a numeric string")
	}

	balanceBytes, err := APIstub.GetState(minter)
	if err != nil {
		return shim.Error(err.Error())
	}
	var balance int
	if balanceBytes == nil {
		balance = 0
	} else {
		balance, _ = strconv.Atoi(string(balanceBytes))
	}
  
	// Mint tokens
	balance += amount

	// Update state with new balance
	err = APIstub.PutState(minter, []byte(strconv.Itoa(balance)))
	if err != nil {
		return shim.Error(err.Error())
	}

	// Emit Transfer event
	eventData := Event{From: "", To: minter, Value: amount}
	eventBytes, err := json.Marshal(eventData)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = APIstub.SetEvent("Transfer", eventBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

// Burn redeems tokens from the minter's account balance
func (s *SmartContract) Burn(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	minter := args[0]
	amount, err := strconv.Atoi(args[1])
	if err != nil {
		return shim.Error("Invalid amount. Expecting a numeric string")
	}

	balanceBytes, err := APIstub.GetState(minter)
	if err != nil {
		return shim.Error(err.Error())
	}
	if balanceBytes == nil {
		return shim.Error("Account not found")
	}
	balance, _ := strconv.Atoi(string(balanceBytes))

	if balance < amount {
		return shim.Error("Insufficient balance")
	}

	balance -= amount

	err = APIstub.PutState(minter, []byte(strconv.Itoa(balance)))
	if err != nil {
		return shim.Error(err.Error())
	}

	eventData := Event{From: minter, To: "", Value: amount}
	eventBytes, err := json.Marshal(eventData)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = APIstub.SetEvent("Transfer", eventBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}


// Transfer transfers tokens from client account to recipient account
func (s *SmartContract) Transfer(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	from := args[0]
	to := args[1]
	amount, err := strconv.Atoi(args[2])
	if err != nil {
		return shim.Error("Invalid amount. Expecting a numeric string")
	}

	fromBalanceBytes, err := APIstub.GetState(from)
	if err != nil {
		return shim.Error(err.Error())
	}
	if fromBalanceBytes == nil {
		return shim.Error("Sender account not found")
	}
	fromBalance, _ := strconv.Atoi(string(fromBalanceBytes))

	toBalanceBytes, err := APIstub.GetState(to)
	if err != nil {
		return shim.Error(err.Error())
	}
	var toBalance int
	if toBalanceBytes == nil {
		toBalance = 0
	} else {
		toBalance, _ = strconv.Atoi(string(toBalanceBytes))
	}

	if fromBalance < amount {
		return shim.Error("Insufficient balance")
	}

	fromBalance -= amount
	toBalance += amount

	err = APIstub.PutState(from, []byte(strconv.Itoa(fromBalance)))
	if err != nil {
		return shim.Error(err.Error())
	}

	err = APIstub.PutState(to, []byte(strconv.Itoa(toBalance)))
	if err != nil {
		return shim.Error(err.Error())
	}

	eventData := Event{From: from, To: to, Value: amount}
	eventBytes, err := json.Marshal(eventData)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = APIstub.SetEvent("Transfer", eventBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

// BalanceOf returns the balance of the given account
func (s *SmartContract) BalanceOf(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	account := args[0]

	balanceBytes, err := APIstub.GetState(account)
	if err != nil {
		return shim.Error(err.Error())
	}
	if balanceBytes == nil {
		return shim.Error("Account not found")
	}

	return shim.Success(balanceBytes)
}

// ClientAccountBalance returns the balance of the account invoking the transaction
func (s *SmartContract) ClientAccountBalance(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	// Get the ID of the client invoking the transaction
	clientID, err := s.GetClientAccountID(APIstub)
	if err != nil {
		return shim.Error(err.Error())
	}

	// Get the balance of the client account
	balanceBytes, err := APIstub.GetState(clientID)
	if err != nil {
		return shim.Error("Failed to get account balance: " + err.Error())
	}
	if balanceBytes == nil {
		return shim.Error("Account not found")
	}

	return shim.Success(balanceBytes)
}

// ClientAccountID returns the ID of the client invoking the transaction
func (s *SmartContract) ClientAccountID(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	// Get the ID of the client invoking the transaction
	clientID, err := s.GetClientAccountID(APIstub)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte(clientID))
}

// GetClientAccountID is a helper function to get the ID of the client invoking the transaction
func (s *SmartContract) GetClientAccountID(APIstub shim.ChaincodeStubInterface) (string, error) {
	// Get the identity of the client
	clientIdentity, err := APIstub.GetCreator()
	if err != nil {
		return "", fmt.Errorf("Failed to get client identity: %s", err.Error())
	}

	// Extract the client ID from the identity
	clientID := string(clientIdentity)
	return clientID, nil
}

// TotalSupply returns the total supply of tokens
func (s *SmartContract) TotalSupply(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	totalSupplyBytes, err := APIstub.GetState(totalSupplyKey)
	if err != nil {
		return shim.Error("Failed to get total supply")
	}
	if totalSupplyBytes == nil {
		return shim.Error("Total supply not set")
	}
	return shim.Success(totalSupplyBytes)
}

// Approve allows `spender` to withdraw from `owner`'s account, multiple times, up to the `amount`.
func (s *SmartContract) Approve(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	owner := args[0]
	spender := args[1]
	amount, err := strconv.Atoi(args[2])
	if err != nil {
		return shim.Error("Invalid amount. Expecting a numeric string")
	}

	allowanceKey := allowancePrefix + owner + spender

	err = APIstub.PutState(allowanceKey, []byte(strconv.Itoa(amount)))
	if err != nil {
		return shim.Error("Failed to set allowance")
	}

	return shim.Success(nil)
}

// Allowance returns the amount which `spender` is still allowed to withdraw from `owner`.
func (s *SmartContract) Allowance(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	owner := args[0]
	spender := args[1]
	allowanceKey := allowancePrefix + owner + spender

	allowanceBytes, err := APIstub.GetState(allowanceKey)
	if err != nil {
		return shim.Error("Failed to get allowance")
	}
	if allowanceBytes == nil {
		return shim.Error("Allowance not found")
	}
	return shim.Success(allowanceBytes)
}

// TransferFrom transfers `amount` tokens from `from` to `to` using the allowance mechanism.
func (s *SmartContract) TransferFrom(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}

	from := args[0]
	to := args[1]
	amount, err := strconv.Atoi(args[2])
	if err != nil {
		return shim.Error("Invalid amount. Expecting a numeric string")
	}
	owner := args[3]

	allowanceKey := allowancePrefix + owner + from

	allowanceBytes, err := APIstub.GetState(allowanceKey)
	if err != nil {
		return shim.Error("Failed to get allowance")
	}
	if allowanceBytes == nil {
		return shim.Error("Allowance not found")
	}

	allowance, _ := strconv.Atoi(string(allowanceBytes))
	if allowance < amount {
		return shim.Error("Insufficient allowance")
	}

	fromBalanceBytes, err := APIstub.GetState(from)
	if err != nil {
		return shim.Error(err.Error())
	}
	if fromBalanceBytes == nil {
		return shim.Error("Sender account not found")
	}
	fromBalance, _ := strconv.Atoi(string(fromBalanceBytes))

	if fromBalance < amount {
		return shim.Error("Insufficient balance")
	}

	toBalanceBytes, err := APIstub.GetState(to)
	if err != nil {
		return shim.Error(err.Error())
	}
	var toBalance int
	if toBalanceBytes == nil {
		toBalance = 0
	} else {
		toBalance, _ = strconv.Atoi(string(toBalanceBytes))
	}

	fromBalance -= amount
	toBalance += amount
	allowance -= amount

	err = APIstub.PutState(from, []byte(strconv.Itoa(fromBalance)))
	if err != nil {
		return shim.Error(err.Error())
	}

	err = APIstub.PutState(to, []byte(strconv.Itoa(toBalance)))
	if err != nil {
		return shim.Error(err.Error())
	}

	err = APIstub.PutState(allowanceKey, []byte(strconv.Itoa(allowance)))
	if err != nil {
		return shim.Error(err.Error())
	}

	eventData := Event{From: from, To: to, Value: amount}
	eventBytes, err := json.Marshal(eventData)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = APIstub.SetEvent("Transfer", eventBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func main() {
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error starting Smart Contract: %s", err)
	}
}
