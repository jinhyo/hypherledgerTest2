/*
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

// ERC20Chaincode is the definition of the chaincode structure.
type ERC20Chaincode struct {
}

// Init is called when the chaincode is instantiated by the blockchain network.
func (cc *ERC20Chaincode) Init(stub shim.ChaincodeStubInterface) sc.Response {
	fcn, params := stub.GetFunctionAndParameters()
	fmt.Println("Init()", fcn, params)
	return shim.Success(nil)
}

// Invoke is called as a result of an application request to run the chaincode.
func (cc *ERC20Chaincode) Invoke(stub shim.ChaincodeStubInterface) sc.Response {
	// GetARgs
	args := stub.GetArgs()
	fmt.Println("GetArgs()-args:", args)

	for i, arg := range args {
		argStr := string(arg)
		fmt.Println("i:", i, "argStr:", argStr)
	}

	// GetStringArgs
	stringArgs := stub.GetStringArgs()
	fmt.Println(stringArgs)

	// GetArgsSlice
	argsSlice, _ := stub.GetArgsSlice()
	fmt.Println("GetArgsSlice():", argsSlice)
	fmt.Println("GetArgsSlice() string:", string(argsSlice))

	return shim.Success(nil)
}
