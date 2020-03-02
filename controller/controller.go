package controller

import (
	"encoding/json"
	"hypherledgertest2/model"
	"log"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

// Controller is ...
type Controller struct {
}

// NewController ...
func NewController() *Controller {
	return &Controller{}
}

// CheckErr is ...
func CheckErr(err error, errMessage string) {
	if err != nil {
		log.Fatalln(errMessage)
	}
}

// Init is ...
func (cc *Controller) Init(stub shim.ChaincodeStubInterface, params []string) sc.Response {
	tokenName, symbol, owner, amount := params[0], params[1], params[2], params[3]

	// check amount is unsigned int
	amountUint, err := strconv.ParseUint(amount, 10, 64)
	CheckErr(err, "amount must be a number or cannot be negative")

	// tokenName, symbol, owner cannot be empty
	if len(tokenName) == 0 || len(symbol) == 0 || len(owner) == 0 {
		return shim.Error("tokenName, symbol, owner cannont be empty")
	}

	// make meta data
	erc20 := model.ERC20Metadata{
		Name:        tokenName,
		Symbol:      symbol,
		Owner:       owner,
		TotalSupply: amountUint}

	erc20Bytes, err := json.Marshal(erc20)
	CheckErr(err, "failed to Marshal erc20")

	// save token to database
	err = stub.PutState(tokenName, erc20Bytes)
	CheckErr(err, "failed to PutState erc20")

	// save owner's balance
	err = stub.PutState(owner, []byte(amount))
	CheckErr(err, "failed to PutState erc20")

	// response
	return shim.Success(nil)
}
