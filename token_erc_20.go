
// // ClientAccountBalance retrieves the account balance of the client's account
// func (t *TokenERC20Chaincode) ClientAccountBalance(stub shim.ChaincodeStubInterface) pb.Response {
// 	// Get client ID
// 	clientIDResp := t.ClientAccountID(stub)
// 	if clientIDResp.Status != shim.OK {
// 		return shim.Error(fmt.Sprintf("Failed to get client ID: %s", clientIDResp.Message))
// 	}
// 	clientID := string(clientIDResp.Payload)

// 	// Load token state
// 	tokenJSON, err := stub.GetState("token")
// 	if err != nil {
// 		return shim.Error(fmt.Sprintf("Failed to get token: %s", err))
// 	}
// 	var token Token
// 	err = json.Unmarshal(tokenJSON, &token)
// 	if err != nil {
// 		return shim.Error(fmt.Sprintf("Failed to unmarshal token: %s", err))
// 	}

// 	// Get balance of client ID
// 	balance, exists := token.Balance[clientID]
// 	if !exists {
// 		return shim.Error(fmt.Sprintf("Balance not found for client ID: %s", clientID))
// 	}

// 	return shim.Success([]byte(strconv.FormatUint(balance, 10)))
// }

// // ClientAccountID retrieves the client account ID
// func (t *TokenERC20Chaincode) ClientAccountID(stub shim.ChaincodeStubInterface) pb.Response {
// 	// Get client ID
// 	clientID, err := stub.GetCreator()
// 	if err != nil {
// 		return shim.Error(fmt.Sprintf("Failed to get client ID: %s", err))
// 	}

// 	return shim.Success(clientID)
// }
