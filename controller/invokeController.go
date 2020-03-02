package controller

import (
	"encoding/json"
	"fmt"
	"hypherledgertest2/model"
	"hypherledgertest2/util"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

// Transfer is invoke function that moves amount token /
// from the caller's address to recipient /
// params - caller's address, recipient's address, amount of token.
func (cc *Controller) Transfer(stub shim.ChaincodeStubInterface, params []string) sc.Response {
	// check a number of params is 3
	if len(params) != 3 {
		return shim.Error("the number of params must be three")
	}

	callerAddress, recipientAddress, transferedMoney := params[0], params[1], params[2]

	// check amount is integer & positive
	transferedMoneyInt, err := util.ConverToPositive(transferedMoney, "transferedMoney")
	if err != nil {
		return shim.Error(err.Error())
	}

	// get caller amount
	callerAmountBytes, err := stub.GetState(callerAddress)
	CheckErr(err, "failed to stub.GetState(callerAddress)")
	if callerAmountBytes == nil {
		return shim.Error("callerAmountBytes does not exist in the DB")
	}

	callerAmountInt, err := strconv.Atoi(string(callerAmountBytes))
	CheckErr(err, "failed to strconv.Atoi(string(callerAmountBytes))")

	// check callerReuslt transferedResult is positive
	if callerAmountInt < transferedMoneyInt {
		return shim.Error("caller's amount must be over the transfered money")
	}

	// get recipient amount
	recipientAmountBytes, err := stub.GetState(recipientAddress)
	CheckErr(err, "failed to stub.GetState(recipientAddress)")
	if recipientAmountBytes == nil {
		recipientAmountBytes = []byte("0")
	}

	recipientAmountInt, err := strconv.Atoi(string(recipientAmountBytes))
	CheckErr(err, "failed to strconv.Atoi(string(recipientAmountBytes))")

	// calculate amount
	callerResult := callerAmountInt - transferedMoneyInt
	recipientResult := recipientAmountInt + transferedMoneyInt

	// save the caller's & recipient's amount
	callerResultBytes, err := json.Marshal(callerResult)
	CheckErr(err, "failed to json.Marshal(callerResult)")

	err = stub.PutState(callerAddress, callerResultBytes)
	CheckErr(err, "failed to stub.PutState(callerAddress, callerResultBytes)")

	recipientResultBytes, err := json.Marshal(recipientResult)
	CheckErr(err, "failed to json.Marshal(recipientResult)")

	err = stub.PutState(recipientAddress, recipientResultBytes)
	CheckErr(err, "failed to stub.PutState(recipientAddress, recipientResultBytes)")

	// emit transfer event
	transferedEvent := model.TransferedEvent{
		Sender:          callerAddress,
		Recipient:       recipientAddress,
		TransferedMoney: transferedMoney}

	transferedEventBytes, err := json.Marshal(transferedEvent)
	CheckErr(err, "failed to json.Marshal(transferedEvent)")

	err = stub.SetEvent("transferEvent", transferedEventBytes)
	CheckErr(err, `failed to stub.SetEvent("transferEvent", transferedEventBytes)`)

	fmt.Println(callerAddress + `sent ` + transferedMoney + ` to ` + recipientAddress)

	return shim.Success([]byte("Transfer Success"))
}

// Approve is invoke function that Sets amount as the allowance /
// of spender over the owner tokens /
// params - owner's address, spender's address, amount of token.
func (cc *Controller) Approve(stub shim.ChaincodeStubInterface, params []string) sc.Response {
	// check the number of params is three
	if len(params) != 3 {
		return shim.Error("the number of params must be three")
	}

	ownerAddress, spenderAddress, amount := params[0], params[1], params[2]

	// check amount is integer & positive
	amountInt, err := util.ConverToPositive(amount, "approveAmount")
	if err != nil {
		return shim.Error(err.Error())
	}

	// create composite key for allowance: approval/owner/spender
	approvalKey, err := stub.CreateCompositeKey("approval", []string{ownerAddress, spenderAddress})
	CheckErr(err, "failed to make a composit key for approval")

	// save the allowance amount
	err = stub.PutState(approvalKey, []byte(amount))
	CheckErr(err, "failed to stub.PutState(approvalKey, []byte(amount))")

	// emit approval event
	approvalEvent := model.ApprovalEvent{Owner: ownerAddress, Spender: spenderAddress, Amount: amountInt}
	approvalEventByte, err := json.Marshal(approvalEvent)
	CheckErr(err, "failed to json.Marshal(approvalEvent)")

	err = stub.SetEvent("approvalEvent", approvalEventByte)

	return shim.Success([]byte("allowance success"))
}

// TransferFrom is a invoke function that Moves amount of tokens from sender(owner) to recipient /
// using allowance of spender /
// parmas - owner's address, spender's address, recipient's address, amount of token.
func (cc *Controller) TransferFrom(stub shim.ChaincodeStubInterface, params []string) sc.Response {
	// check the number of parmas is 4
	if len(params) != 4 {
		return shim.Error("the number of params must be four")
	}

	ownerAddress, spenderAddress, recipientAddress, amount := params[0], params[1], params[2], params[3]

	// check amount is integer & positive
	amountInt, err := util.ConverToPositive(amount, "TransferedAmount")
	if err != nil {
		return shim.Error(err.Error())
	}

	spenderAllowanceResponse := cc.Allowance(stub, []string{ownerAddress, spenderAddress})
	if spenderAllowanceResponse.Status >= 400 {
		return shim.Error(`failed to cc.allowance([]string{ownerAddress, spenderAddress}), err: ` + spenderAllowanceResponse.GetMessage())
	}

	// get allowance of spender and recipient
	recipientAllowanceResponse := cc.Allowance(stub, []string{ownerAddress, recipientAddress})
	if recipientAllowanceResponse.Status >= 400 {
		return shim.Error(`failed to cc.allowance([]string{ownerAddress, spenderAddress}), err: ` + recipientAllowanceResponse.GetMessage())
	}

	// convert allowance response paylaod to allowance data(int)
	spenderAllowanceStr := string(spenderAllowanceResponse.GetPayload())
	spenderAllowanceInt, err := strconv.Atoi(spenderAllowanceStr)
	CheckErr(err, `failed to strconv.Atoi(spenderAllowanceStr)`)

	recipientAllowanceStr := string(recipientAllowanceResponse.GetPayload())
	recipientAllowanceInt, err := strconv.Atoi(recipientAllowanceStr)
	CheckErr(err, `failed to strconv.Atoi(recipientAllowanceStr)`)

	// transfer from owner to recipient
	transferResponse := cc.Transfer(stub, []string{ownerAddress, recipientAddress, amount})
	if transferResponse.Status >= 400 {
		return shim.Error(`failed to cc.transfer([]string{spenderAddress, recipientAddress, amount}), err: ` + transferResponse.GetMessage())
	}

	// decrease & increase allowance amount
	spenderAllowanceInt -= amountInt
	recipientAllowanceInt += amountInt

	// approve amount of tokens transfered
	spenderAllowanceStr = strconv.Itoa(spenderAllowanceInt)
	recipientAllowanceStr = strconv.Itoa(recipientAllowanceInt)

	Res := cc.Approve(stub, []string{ownerAddress, spenderAddress, spenderAllowanceStr})
	if Res.Status >= 400 {
		return shim.Error(`failed to cc.approve(ownerAddress, spenderAddress, spenderAllowanceStr), err: ` + Res.GetMessage())
	}

	Res = cc.Approve(stub, []string{ownerAddress, recipientAddress, recipientAllowanceStr})
	if Res.Status >= 400 {
		return shim.Error(`failed to cc.approve(ownerAddress, recipientAddress, recipientAllowanceStr), err: ` + Res.GetMessage())
	}

	return shim.Success([]byte("transferFrom func success"))
}

// TransferFromOther is an invoke function that invokes transferFrom in different chaincode /
// params - chaincodeName, senderAddress, recipientAddress, amount
func (cc *Controller) TransferFromOther(stub shim.ChaincodeStubInterface, params []string) sc.Response {
	// check the number of parmas is 4
	if len(params) != 5 {
		return shim.Error("the number of params must be five")
	}

	chaincodeName, ownerAddress, senderAddress, recipientAddress, amount := params[0], params[1], params[2], params[3], params[4]

	// make arguments
	args := [][]byte{[]byte("transferFrom"), []byte(ownerAddress), []byte(senderAddress), []byte(recipientAddress), []byte(amount)}

	// get channel
	channelID := stub.GetChannelID()

	// invoke transferFrom in another chaincode
	invokeResponse := stub.InvokeChaincode(chaincodeName, args, channelID)
	if invokeResponse.GetStatus() >= 400 {
		return shim.Error(`failed to stub.InvokeChaincode(chaincodeName, args, channelID), err: ` + invokeResponse.GetMessage())
	}

	return shim.Success([]byte("transferFrom in other token success"))
}

// IncreaseAllowance is invoke function that increases spender's allowance by owner /
// params - owner's address, spender's address, amount of increase.
func (cc *Controller) IncreaseAllowance(stub shim.ChaincodeStubInterface, params []string) sc.Response {
	// check the number of params if three
	if len(params) != 3 {
		return shim.Error("the number of params must be three")
	}

	ownerAddress, targetAddress, amount := params[0], params[1], params[2]

	// check amount is integer & positive
	amountInt, err := util.ConverToPositive(amount, "IncreaseAmount")
	if err != nil {
		return shim.Error(err.Error())
	}

	// get allowance
	allowanceResponse := cc.Allowance(stub, []string{ownerAddress, targetAddress})
	if allowanceResponse.Status >= 400 {
		return shim.Error(`failed to get allowanceResponse, err: ` + allowanceResponse.GetMessage())
	}

	// convert allowance response payload to allowance data
	allowanceInt, err := strconv.Atoi(string(allowanceResponse.GetPayload()))
	CheckErr(err, `failed to strconv.Atoi(string(allowanceResponse.GetPayload()))`)

	// increase allowance
	allowanceInt += amountInt

	// call approve
	allowanceStr := strconv.Itoa(allowanceInt)
	approveResponse := cc.Approve(stub, []string{ownerAddress, targetAddress, allowanceStr})
	if approveResponse.Status >= 400 {
		return shim.Error(`failed to get approveResponse, err: ` + approveResponse.GetMessage())
	}

	return shim.Success([]byte("increaseAllowance func success"))
}

// DecreaseAllowance is invoke function that increases spender's allowance by owner /
// params - owner's address, spender's address, amount of decrease.
func (cc *Controller) DecreaseAllowance(stub shim.ChaincodeStubInterface, params []string) sc.Response {
	// check the number of params if three
	if len(params) != 3 {
		return shim.Error("the number of params must be three")
	}

	ownerAddress, targetAddress, amount := params[0], params[1], params[2]

	// check amount is integer & positive
	amountInt, err := util.ConverToPositive(amount, "descreaseAmount")
	if err != nil {
		return shim.Error(err.Error())
	}

	// get allowance
	allowanceResponse := cc.Allowance(stub, []string{ownerAddress, targetAddress})
	if allowanceResponse.Status >= 400 {
		return shim.Error(`failed to get allowanceResponse, err: ` + allowanceResponse.GetMessage())
	}

	// convert allowance response payload to allowance data
	allowanceInt, err := strconv.Atoi(string(allowanceResponse.GetPayload()))
	CheckErr(err, `failed to strconv.Atoi(string(allowanceResponse.GetPayload()))`)

	// decrease allowance
	allowanceInt -= amountInt

	// call approve
	allowanceStr := strconv.Itoa(allowanceInt)
	approveResponse := cc.Approve(stub, []string{ownerAddress, targetAddress, allowanceStr})
	if approveResponse.Status >= 400 {
		return shim.Error(`failed to get approveResponse, err: ` + approveResponse.GetMessage())
	}

	return shim.Success([]byte("decreaseAllowance func success"))
}

// Mint is ...
func (cc *Controller) Mint(stub shim.ChaincodeStubInterface, params []string) sc.Response {
	return shim.Success(nil)
}

// Burn is ...
func (cc *Controller) Burn(stub shim.ChaincodeStubInterface, params []string) sc.Response {
	return shim.Success(nil)
}
