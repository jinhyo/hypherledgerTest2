/*
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

// ERC20Chaincode is the definition of the chaincode structure.
type ERC20Chaincode struct {
}

// ERC20Metadata is the definition of the Token meta info
type ERC20Metadata struct {
	Name        string `json:"name"`
	Symbol      string `json:"symbol"`
	Owner       string `json:"owner"`
	TotalSupply uint64 `json:"totalsupply"`
}

// TransferedEvent is the log of the TransferedEvent
type TransferedEvent struct {
	Sender          string `json:"sender"`
	Recipient       string `json:"recipient"`
	TransferedMoney string `json:"transferedMoney"`
}

// ApprovalEvent is the log of the ApprovalEvent
type ApprovalEvent struct {
	Owner   string `json:"owner"`
	Spender string `json:"spender"`
	Amount  int    `json:"amount"`
}

func checkErr(err error, errMessage string) {
	if err != nil {
		log.Fatalln(errMessage)
	}
}

// Init is called when the chaincode is instantiated by the blockchain network.
// params : tokenName, symbol, owner(address), amount
func (cc *ERC20Chaincode) Init(stub shim.ChaincodeStubInterface) sc.Response {
	_, params := stub.GetFunctionAndParameters()
	fmt.Println("Init is called with params:", params)
	if len(params) != 4 {
		return shim.Error("incorrect number of the params")
	}

	tokenName, symbol, owner, amount := params[0], params[1], params[2], params[3]

	// check amount is unsigned int
	amountUint, err := strconv.ParseUint(amount, 10, 64)
	checkErr(err, "amount must be a number or cannot be negative")

	// tokenName, symbol, owner cannot be empty
	if len(tokenName) == 0 || len(symbol) == 0 || len(owner) == 0 {
		return shim.Error("tokenName, symbol, owner cannont be empty")
	}

	// make meta data
	erc20 := ERC20Metadata{
		Name:        tokenName,
		Symbol:      symbol,
		Owner:       owner,
		TotalSupply: amountUint}

	erc20Bytes, err := json.Marshal(erc20)
	checkErr(err, "failed to Marshal erc20")

	// save token to database
	err = stub.PutState(tokenName, erc20Bytes)
	checkErr(err, "failed to PutState erc20")

	// save owner's balance
	err = stub.PutState(owner, []byte(amount))
	checkErr(err, "failed to PutState erc20")

	// response
	return shim.Success(nil)
}

// Invoke is called as a result of an application request to run the chaincode.
func (cc *ERC20Chaincode) Invoke(stub shim.ChaincodeStubInterface) sc.Response {
	fcn, params := stub.GetFunctionAndParameters()

	switch fcn {
	case "totalSupply":
		return cc.totalSupply(stub, params)
	case "balanceOf":
		return cc.balanceOf(stub, params)
	case "transfer":
		return cc.transfer(stub, params)
	case "allowance":
		return cc.allowance(stub, params)
	case "approve":
		return cc.approve(stub, params)
	case "approvalList":
		return cc.approvalList(stub, params)
	case "transferFrom":
		return cc.transferFrom(stub, params)
	case "increaseAllowance":
		return cc.increaseAllowance(stub, params)
	case "decreaseAllowance":
		return cc.decreaseAllowance(stub, params)
	case "mint":
		return cc.mint(stub, params)
	case "burn":
		return cc.burn(stub, params)
	default:
		return sc.Response{Status: 404, Message: "404 Not Found", Payload: nil}
	}
}

// totalSuuply is query function
// params - tokenName
// Returns the amount of token in the ledge
func (cc *ERC20Chaincode) totalSupply(stub shim.ChaincodeStubInterface, params []string) sc.Response {
	if len(params) != 1 {
		return shim.Error("the number of params must be one")
	}

	tokenName := params[0]

	erc20 := ERC20Metadata{}
	erc20Bytes, err := stub.GetState(tokenName) // tokenName아 없으면 erc20Bytes, err에 nil이 들어감
	checkErr(err, `failed to  stub.GetState("tokenName")`)
	if erc20Bytes == nil {
		return shim.Error("erc20Bytes is nil")
	}

	err = json.Unmarshal(erc20Bytes, &erc20)
	checkErr(err, `failed to json.Unmarshal(erc20Bytes, &erc20)`)
	totalBalance := erc20.TotalSupply

	totalBalanceBytes, err := json.Marshal(totalBalance)
	checkErr(err, "failed to json.Marshal(totalBalance)")

	fmt.Println(tokenName + "', total supply is" + string(totalBalanceBytes))

	return shim.Success(totalBalanceBytes)
}

// balanceOf is query function
// params - address
// Returns the amount of tokens owned by the addresss
func (cc *ERC20Chaincode) balanceOf(stub shim.ChaincodeStubInterface, params []string) sc.Response {
	if len(params) != 1 {
		return shim.Error("the number of params must be one")
	}

	address := params[0]

	balanceByte, err := stub.GetState(address)
	checkErr(err, `stub.GetState("owner")`)
	if balanceByte == nil {
		return shim.Error("balanceByte is nil")
	}

	fmt.Println(address + "'s, balance is " + string(balanceByte))
	return shim.Success(balanceByte)
}

// transfer is invoke function that moves amount token
// from the caller's address to recipient
// params - caller's address, recipient's address, amount of token
func (cc *ERC20Chaincode) transfer(stub shim.ChaincodeStubInterface, params []string) sc.Response {
	// check a number of params is 3
	if len(params) != 3 {
		return shim.Error("the number of params must be three")
	}

	callerAddress, recipientAddress, transferedMoney := params[0], params[1], params[2]

	// check amount is integer & positive
	transferedMoneyInt, err := strconv.Atoi(transferedMoney)
	checkErr(err, "failed to strconv.Atoi(transferedMoney)")

	if transferedMoneyInt <= 0 {
		return shim.Error("transfered money must be more than zero")
	}

	// get caller amount
	callerAmountBytes, err := stub.GetState(callerAddress)
	checkErr(err, "failed to stub.GetState(callerAddress)")
	if callerAmountBytes == nil {
		return shim.Error("callerAmountBytes does not exist in the DB")
	}

	callerAmountInt, err := strconv.Atoi(string(callerAmountBytes))
	checkErr(err, "failed to strconv.Atoi(string(callerAmountBytes))")

	// check callerReuslt transferedResult is positive
	if callerAmountInt < transferedMoneyInt {
		return shim.Error("caller's amount must be over the transfered money")
	}

	// get recipient amount
	recipientAmountBytes, err := stub.GetState(recipientAddress)
	checkErr(err, "failed to stub.GetState(recipientAddress)")
	if recipientAmountBytes == nil {
		recipientAmountBytes = []byte("0")
	}

	recipientAmountInt, err := strconv.Atoi(string(recipientAmountBytes))
	checkErr(err, "failed to strconv.Atoi(string(recipientAmountBytes))")

	// calculate amount
	callerResult := callerAmountInt - transferedMoneyInt
	recipientResult := recipientAmountInt + transferedMoneyInt

	// save the caller's & recipient's amount
	callerResultBytes, err := json.Marshal(callerResult)
	checkErr(err, "failed to json.Marshal(callerResult)")

	err = stub.PutState(callerAddress, callerResultBytes)
	checkErr(err, "failed to stub.PutState(callerAddress, callerResultBytes)")

	recipientResultBytes, err := json.Marshal(recipientResult)
	checkErr(err, "failed to json.Marshal(recipientResult)")

	err = stub.PutState(recipientAddress, recipientResultBytes)
	checkErr(err, "failed to stub.PutState(recipientAddress, recipientResultBytes)")

	// emit transfer event
	transferedEvent := TransferedEvent{
		Sender:          callerAddress,
		Recipient:       recipientAddress,
		TransferedMoney: transferedMoney}

	transferedEventBytes, err := json.Marshal(transferedEvent)
	checkErr(err, "failed to json.Marshal(transferedEvent)")

	err = stub.SetEvent("transferEvent", transferedEventBytes)
	checkErr(err, `failed to stub.SetEvent("transferEvent", transferedEventBytes)`)

	fmt.Println(callerAddress + `sent ` + transferedMoney + ` to ` + recipientAddress)

	return shim.Success([]byte("Transfer Success"))
}

// allowance is query function
// params - owner's address, spender's address
// Returns the remaining amount of token to invoke {transferFrom}
func (cc *ERC20Chaincode) allowance(stub shim.ChaincodeStubInterface, params []string) sc.Response {
	// check the number of the params is 2
	if len(params) != 2 {
		return shim.Error("the number of params must be two")
	}

	ownerAddress, spenderAddress := params[0], params[1]

	// create composite key
	approvalKey, err := stub.CreateCompositeKey("approval", []string{ownerAddress, spenderAddress})
	checkErr(err, "failed to make a composite key for allowance")

	// get amount
	allowanceAmount, err := stub.GetState(approvalKey)
	checkErr(err, "failed to get allowance amount from the ledger")
	if allowanceAmount == nil {
		allowanceAmount = []byte("0")
	}

	return shim.Success(allowanceAmount)
}

// approve is invoke function that Sets amount as the allowance
// of spender over the owner tokens
// params - owner's address, spender's address, amount of token
func (cc *ERC20Chaincode) approve(stub shim.ChaincodeStubInterface, params []string) sc.Response {
	// check the number of params is three
	if len(params) != 3 {
		return shim.Error("the number of params must be three")
	}

	ownerAddress, spenderAddress, amount := params[0], params[1], params[2]

	// check amount is integer & positive
	amountInt, err := strconv.Atoi(amount)
	checkErr(err, "failed to strconv.Atoi(amount)")

	if amountInt <= 0 {
		return shim.Error("amount must be more than 0")
	}

	// create composite key for allowance: approval/owner/spender
	approvalKey, err := stub.CreateCompositeKey("approval", []string{ownerAddress, spenderAddress})
	checkErr(err, "failed to make a composit key for approval")

	// save the allowance amount
	err = stub.PutState(approvalKey, []byte(amount))
	checkErr(err, "failed to stub.PutState(approvalKey, []byte(amount))")

	// emit approval event
	approvalEvent := ApprovalEvent{Owner: ownerAddress, Spender: spenderAddress, Amount: amountInt}
	approvalEventByte, err := json.Marshal(approvalEvent)
	checkErr(err, "failed to json.Marshal(approvalEvent)")

	err = stub.SetEvent("approvalEvent", approvalEventByte)

	return shim.Success([]byte("allowance success"))
}

func (cc *ERC20Chaincode) approvalList(stub shim.ChaincodeStubInterface, params []string) sc.Response {
	// check the number of the parameters is one
	if len(params) != 1 {
		return shim.Error("the number of params must be one")
	}

	// get all approval list (format is iterator)

	// make slice for return value

	// iterator
	// - get spender address
	// - get amount
	// - add approval result

	// convert approvalList to []byte for return

}

func (cc *ERC20Chaincode) transferFrom(stub shim.ChaincodeStubInterface, params []string) sc.Response {
	return shim.Success(nil)
}

func (cc *ERC20Chaincode) increaseAllowance(stub shim.ChaincodeStubInterface, params []string) sc.Response {
	return shim.Success(nil)
}

func (cc *ERC20Chaincode) decreaseAllowance(stub shim.ChaincodeStubInterface, params []string) sc.Response {
	return shim.Success(nil)
}

func (cc *ERC20Chaincode) mint(stub shim.ChaincodeStubInterface, params []string) sc.Response {
	return shim.Success(nil)
}

func (cc *ERC20Chaincode) burn(stub shim.ChaincodeStubInterface, params []string) sc.Response {
	return shim.Success(nil)
}
