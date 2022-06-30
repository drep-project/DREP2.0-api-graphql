/*
Package rpc implements bridge to Opera full node API interface.

We recommend using local IPC for fast and the most efficient inter-process communication between the API server
and an Opera/Opera node. Any remote RPC connection will work, but the performance may be significantly degraded
by extra networking overhead of remote RPC calls.

You should also consider security implications of opening Opera RPC interface for a remote access.
If you considering it as your deployment strategy, you should establish encrypted channel between the API server
and Opera RPC interface with connection limited to specified endpoints.

We strongly discourage opening Opera RPC interface for unrestricted Internet access.
*/
package rpc

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"math/big"
	"strings"
	"time"
)

// maxAcceptedGasPrice defines max accepted gas price, everything above invokes additional check.
var maxAcceptedGasPrice = big.NewInt(1_000_000_000_000_000_000)

// solidityCallFunctionName is the identifier for a call to Solidity function name() view returns(string)
var solidityCallFunctionName = common.Hex2Bytes("06fdde03")

// GasPrice pulls the current amount of WEI for single Gas.
func (ftm *FtmBridge) GasPrice() (hexutil.Big, error) {
	// keep track of the operation
	ftm.log.Debugf("checking current gas price")

	// call for data
	var price hexutil.Big
	var try uint8
	for {
		err := ftm.rpc.Call(&price, "ftm_gasPrice")
		if err != nil {
			ftm.log.Error("current gas price could not be obtained")
			return price, err
		}

		// did we get an acceptable price
		if price.ToInt().Cmp(maxAcceptedGasPrice) <= 0 {
			break
		}

		// all attempts failed
		if try > 2 {
			ftm.log.Errorf("can not get reasonable gas price from the backend server")
			return price, fmt.Errorf("gas price not available, please try again later")
		}

		try++
		time.Sleep(100 * time.Millisecond)
	}

	return price, nil
}

// GasEstimate calculates the estimated amount of Gas required to perform
// transaction described by the input params.
func (ftm *FtmBridge) GasEstimate(trx *struct {
	From  *common.Address
	To    *common.Address
	Value *hexutil.Big
	Data  *string
}) (*hexutil.Uint64, error) {
	// keep track of the operation
	ftm.log.Debugf("calling for gas amount estimation")

	var val hexutil.Uint64
	err := ftm.rpc.Call(&val, "ftm_estimateGas", trx)
	if err != nil {
		// missing required argument? incompatibility between old and new RPC API
		if strings.Contains(err.Error(), "missing value") {
			return ftm.GasEstimateWithBlock(trx)
		}

		// return error
		ftm.log.Errorf("can not estimate gas; %s", err.Error())
		return nil, err
	}

	return &val, nil
}

// GasEstimateWithBlock calculates the estimated amount of Gas required to perform
// transaction described by the input params with specifying the block on which the calculation
// should happen (new RPC API compatibility).
// @TODO Replace the old gas estimate call once the API gets upgraded on all nodes.
func (ftm *FtmBridge) GasEstimateWithBlock(trx *struct {
	From  *common.Address
	To    *common.Address
	Value *hexutil.Big
	Data  *string
}) (*hexutil.Uint64, error) {
	// keep track of the operation
	ftm.log.Debugf("calling for gas amount estimation with block details")

	var val hexutil.Uint64
	err := ftm.rpc.Call(&val, "ftm_estimateGas", trx, BlockTypeLatest)
	if err != nil {
		// return error
		ftm.log.Errorf("can not estimate gas; %s", err.Error())
		return nil, err
	}

	return &val, nil
}

// TokenNameAttempt tries to extract token name from the contract on the given address.
// We assume to be able to call Solidity: function name() view returns(string)
func (ftm *FtmBridge) TokenNameAttempt(adr *common.Address) (string, error) {
	// call the function on the contract address
	data, err := ftm.eth.CallContract(context.Background(), ethereum.CallMsg{
		From: common.Address{},
		To:   adr,
		Data: solidityCallFunctionName,
	}, nil)
	if err != nil {
		return "", err
	}

	return parseAbiString(data)
}

// parseAbiString decodes a string in ABI format, if possible.
func parseAbiString(data []byte) (string, error) {
	// we expect string in response => offset + string length + string body padded to 32 bytes (at least 3x32 bytes)
	if nil == data || len(data) < 96 {
		return "", fmt.Errorf("abi encoded string expected, only %d bytes received", len(data))
	}

	// read the offset of the actual string
	bigOffset := new(big.Int).SetBytes(data[:32])
	if bigOffset.BitLen() > 63 {
		return "", fmt.Errorf("string offset larger than int64: %v", bigOffset)
	}

	// the string data starts with the length, so add another 32 bytes to skip it
	offset := int(bigOffset.Add(bigOffset, common.Big32).Uint64())

	// how long is the string?
	bigLength := new(big.Int).SetBytes(data[offset-32 : offset])
	if bigLength.BitLen() > 63 {
		return "", fmt.Errorf("string length larger than int64: %v", bigLength)
	}
	length := int(bigLength.Uint64())

	return string(data[offset : offset+length]), nil
}
