package main

import (
	"encoding/json"
	"log"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

// Define key names for options
const nameKey = "name"
const symbolKey = "symbol"
const decimalsKey = "decimals"
const totalSupplyKey = "totalSupply"

// Define objectType names for prefix
const allowancePrefix = "allowance"

// Define SmartContract structure
type SmartContract struct {
}

// event provides an organized struct for emitting events
type event struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Value int    `json:"value"`
}

// Init initializes chaincode
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success(nil)
}

// Invoke - Our entry point for Invocations
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) peer.Response {
	function, args := APIstub.GetFunctionAndParameters()
	switch function {
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
	case "Name":
		return s.Name(APIstub, args)
	case "Symbol":
		return s.Symbol(APIstub, args)
	case "Initialize":
		return s.Initialize(APIstub, args)
	default:
		return shim.Error("Invalid function name")
	}
}

// Mint creates new tokens and adds them to minter's account balance
// This function triggers a Transfer event
func (s *SmartContract) Mint(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	minter := args[0]
	amount, err := strconv.Atoi(args[1])
	if err != nil {
		return shim.Error("Invalid amount. Expecting a numeric string")
	}

	// Check if caller is authorized to mint tokens
	// (you may need to implement this authorization logic)

	// Get current balance of minter
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
	eventData := event{From: "", To: minter, Value: amount}
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
// This function triggers a Transfer event
func (s *SmartContract) Burn(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	minter := args[0]
	amount, err := strconv.Atoi(args[1])
	if err != nil {
		return shim.Error("Invalid amount. Expecting a numeric string")
	}

	// Check if caller is authorized to burn tokens
	// (you may need to implement this authorization logic)

	// Get current balance of minter
	balanceBytes, err := APIstub.GetState(minter)
	if err != nil {
		return shim.Error(err.Error())
	}
	if balanceBytes == nil {
		return shim.Error("Account not found")
	}
	balance, _ := strconv.Atoi(string(balanceBytes))

	// Ensure minter has enough tokens to burn
	if balance < amount {
		return shim.Error("Insufficient balance")
	}

	// Burn tokens
	balance -= amount

	// Update state with new balance
	err = APIstub.PutState(minter, []byte(strconv.Itoa(balance)))
	if err != nil {
		return shim.Error(err.Error())
	}

	// Emit Transfer event
	eventData := event{From: minter, To: "", Value: amount}
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
// recipient account must be a valid clientID as returned by the ClientID() function
// This function triggers a Transfer event
func (s *SmartContract) Transfer(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	from := args[0]
	to := args[1]
	amount, err := strconv.Atoi(args[2])
	if err != nil {
		return shim.Error("Invalid amount. Expecting a numeric string")
	}

	// Get balances of sender and recipient
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

	// Ensure sender has enough tokens to transfer
	if fromBalance < amount {
		return shim.Error("Insufficient balance")
	}

	// Transfer tokens
	fromBalance -= amount
	toBalance += amount

	// Update sender's balance
	err = APIstub.PutState(from, []byte(strconv.Itoa(fromBalance)))
	if err != nil {
		return shim.Error(err.Error())
	}

	// Update recipient's balance
	err = APIstub.PutState(to, []byte(strconv.Itoa(toBalance)))
	if err != nil {
		return shim.Error(err.Error())
	}

	// Emit Transfer event
	eventData := event{From: from, To: to, Value: amount}
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
func (s *SmartContract) BalanceOf(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {
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

// ClientAccountBalance returns the balance of the requesting client's account
func (s *SmartContract) ClientAccountBalance(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {
	// In this implementation, the requesting client's account is identified by its certificate
	// You may need to implement additional logic to identify clients in your actual implementation
	cert, err := APIstub.GetCreator()
	if err != nil {
		return shim.Error("Failed to get client's certificate")
	}
	clientID := string(cert)

	return s.BalanceOf(APIstub, []string{clientID})
}

// ClientAccountID returns the id of the requesting client's account
// In this implementation, the client account ID is the client's certificate
func (s *SmartContract) ClientAccountID(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {
	// In this implementation, the requesting client's account is identified by its certificate
	// You may need to implement additional logic to identify clients in your actual implementation
	cert, err := APIstub.GetCreator()
	if err != nil {
		return shim.Error("Failed to get client's certificate")
	}
	clientID := string(cert)

	return shim.Success([]byte(clientID))
}

// TotalSupply returns the total supply of tokens
func (s *SmartContract) TotalSupply(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {
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
// If this function is called again it overwrites the current allowance with the `amount`.
func (s *SmartContract) Approve(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {
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
func (s *SmartContract) Allowance(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {
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
// `amount` is then deducted from the caller’s allowance.
func (s *SmartContract) TransferFrom(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}

	owner := args[0]
	spender := args[1]
	to := args[2]
	amount, err := strconv.Atoi(args[3])
	if err != nil {
		return shim.Error("Invalid amount. Expecting a numeric string")
	}

	allowanceKey := allowancePrefix + owner + spender

	allowanceBytes, err := APIstub.GetState(allowanceKey)
	if err != nil {
		return shim.Error("Failed to get allowance")
	}
	if allowanceBytes == nil {
		return shim.Error("Allowance not found")
	}

	allowance, _ := strconv.Atoi(string(allowanceBytes))
	if allowance < amount {
		return shim.Error("Allowance exceeded")
	}

	// Get balances of owner and recipient
	fromBalanceBytes, err := APIstub.GetState(owner)
	if err != nil {
		return shim.Error(err.Error())
	}
	if fromBalanceBytes == nil {
		return shim.Error("Owner account not found")
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

	// Ensure owner has enough tokens to transfer
	if fromBalance < amount {
		return shim.Error("Insufficient balance")
	}

	// Transfer tokens
	fromBalance -= amount
	toBalance += amount

	// Update owner's balance
	err = APIstub.PutState(owner, []byte(strconv.Itoa(fromBalance)))
	if err != nil {
		return shim.Error(err.Error())
	}

	// Update recipient's balance
	err = APIstub.PutState(to, []byte(strconv.Itoa(toBalance)))
	if err != nil {
		return shim.Error(err.Error())
	}

	// Update spender's allowance
	allowance -= amount
	err = APIstub.PutState(allowanceKey, []byte(strconv.Itoa(allowance)))
	if err != nil {
		return shim.Error("Failed to update allowance")
	}

	// Emit Transfer event
	eventData := event{From: owner, To: to, Value: amount}
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

// Name returns the name of the token
func (s *SmartContract) Name(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {
	nameBytes, err := APIstub.GetState(nameKey)
	if err != nil {
		return shim.Error("Failed to get token name")
	}
	if nameBytes == nil {
		return shim.Error("Token name not set")
	}
	return shim.Success(nameBytes)
}

// Symbol returns the symbol of the token
func (s *SmartContract) Symbol(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {
	symbolBytes, err := APIstub.GetState(symbolKey)
	if err != nil {
		return shim.Error("Failed to get token symbol")
	}
	if symbolBytes == nil {
		return shim.Error("Token symbol not set")
	}
	return shim.Success(symbolBytes)
}

// Initialize initializes the token's state (name, symbol, decimals, totalSupply)
func (s *SmartContract) Initialize(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {
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

	return shim.Success(nil)
}

func main() {
	err := shim.Start(new(SmartContract))
	if err != nil {
		log.Fatalf("Error starting token-erc-20 chaincode: %v", err)
	}
}
