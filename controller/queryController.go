package controller

import (
	"encoding/json"
	"fmt"
	"hypherledgertest2/model"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

// TotalSupply is query function
// params - tokenName
// Returns the amount of token in the ledge
func (cc *Controller) TotalSupply(stub shim.ChaincodeStubInterface, params []string) sc.Response {
	if len(params) != 1 {
		return shim.Error("the number of params must be one")
	}

	tokenName := params[0]

	erc20 := model.ERC20Metadata{}
	erc20Bytes, err := stub.GetState(tokenName) // tokenName아 없으면 erc20Bytes, err에 nil이 들어감
	CheckErr(err, `failed to  stub.GetState("tokenName")`)
	if erc20Bytes == nil {
		return shim.Error("erc20Bytes is nil")
	}

	err = json.Unmarshal(erc20Bytes, &erc20)
	CheckErr(err, `failed to json.Unmarshal(erc20Bytes, &erc20)`)
	totalBalance := erc20.TotalSupply

	totalBalanceBytes, err := json.Marshal(totalBalance)
	CheckErr(err, "failed to json.Marshal(totalBalance)")

	fmt.Println(tokenName + "', total supply is" + string(totalBalanceBytes))

	return shim.Success(totalBalanceBytes)
}

// BalanceOf is query function
// params - address
// Returns the amount of tokens owned by the addresss
func (cc *Controller) BalanceOf(stub shim.ChaincodeStubInterface, params []string) sc.Response {
	if len(params) != 1 {
		return shim.Error("the number of params must be one")
	}

	address := params[0]

	balanceByte, err := stub.GetState(address)
	CheckErr(err, `stub.GetState("owner")`)
	if balanceByte == nil {
		return shim.Error("balanceByte is nil")
	}

	fmt.Println(address + "'s, balance is " + string(balanceByte))
	return shim.Success(balanceByte)
}

// Allowance is a query function /
// params - owner's address, spender's address /
// Returns the remaining amount of token to invoke {transferFrom}.
func (cc *Controller) Allowance(stub shim.ChaincodeStubInterface, params []string) sc.Response {
	// check the number of the params is 2
	if len(params) != 2 {
		return shim.Error("the number of params must be two")
	}

	ownerAddress, spenderAddress := params[0], params[1]

	// create composite key
	approvalKey, err := stub.CreateCompositeKey("approval", []string{ownerAddress, spenderAddress})
	CheckErr(err, "failed to make a composite key for allowance")

	// get amount
	allowanceAmount, err := stub.GetState(approvalKey)
	CheckErr(err, "failed to get allowance amount from the ledger")
	if allowanceAmount == nil {
		allowanceAmount = []byte("0")
	}

	return shim.Success(allowanceAmount)
}

// ApprovalList is a query function.
// params - owner's address.
// Returns the approval list approved by owner.
func (cc *Controller) ApprovalList(stub shim.ChaincodeStubInterface, params []string) sc.Response {
	// check the number of the parameters is one
	if len(params) != 1 {
		return shim.Error("the number of params must be one")
	}

	ownerAddress := params[0]

	// get all approval list (format is iterator)
	approvalIter, err := stub.GetStateByPartialCompositeKey("approval", []string{ownerAddress})
	CheckErr(err, `failed to stub.GetStateByPartialCompositeKey("approval", []string{ownerAddress})`)

	// make slice for return value
	approvalSlice := []model.ApprovalEvent{}

	// iterator
	for approvalIter.HasNext() {
		approvalKeyValue, err := approvalIter.Next()
		CheckErr(err, `failed to approvalIter.Next()`)

		_, addresses, err := stub.SplitCompositeKey(approvalKeyValue.GetKey())
		CheckErr(err, `failed to stub.SplitCompositeKey(approvalKeyValue.GetKey())`)

		// - get spender address
		spenderAddress := addresses[1]

		// - get amount
		amount := approvalKeyValue.GetValue()
		if amount == nil {
			return shim.Error("amount does not exist in the ledger")
		}

		// - add approval result
		amountInt, err := strconv.Atoi(string(amount))
		CheckErr(err, `failed to strconv.Atoi(string(amount))`)

		approval := model.ApprovalEvent{Owner: ownerAddress, Spender: spenderAddress, Amount: amountInt}
		approvalSlice = append(approvalSlice, approval)
	}

	// convert approvalList to []byte for return
	approvalSliceByte, err := json.Marshal(approvalSlice)
	CheckErr(err, `failed to json.Marshal(approvalSlice)`)

	return shim.Success(approvalSliceByte)
}
