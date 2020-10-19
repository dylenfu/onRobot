package sdk

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/contracts/native"
	"github.com/ethereum/go-ethereum/contracts/native/governance"
	"github.com/ethereum/go-ethereum/contracts/native/plt"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palettechain/onRobot/pkg/log"
)

var (
	PLTABI, GovernanceABI abi.ABI
	PLTAddress            = common.HexToAddress(native.PLTContractAddress)
	GovernanceAddress     = common.HexToAddress(native.GovernanceContractAddress)

	gasLimit, deployGasLimit uint64
	blockPeriod              time.Duration
)

const (
	gasPrice = 0
)

func Init(_gasLimit, _deployGasLimit uint64, _blockPeriod time.Duration) {
	PLTABI = plt.GetABI()
	GovernanceABI = governance.GetABI()

	gasLimit = _gasLimit
	deployGasLimit = _deployGasLimit
	blockPeriod = _blockPeriod
}

func (c *Client) GetNonce(address string) uint64 {
	var raw string

	if err := c.Call(
		&raw,
		"eth_getTransactionCount",
		address,
		"latest",
	); err != nil {
		panic(fmt.Errorf("failed to get nonce: [%v]", err))
	}

	without0xStr := strings.Replace(raw, "0x", "", -1)
	bigNonce, _ := new(big.Int).SetString(without0xStr, 16)
	return bigNonce.Uint64()
}

func (c *Client) SendTransaction(contractAddr common.Address, payload []byte) (common.Hash, error) {
	addr := c.Address()

	nonce := c.GetNonce(addr.Hex())
	if c.currentNonce > nonce {
		nonce = c.currentNonce
	}

	tx := types.NewTransaction(
		nonce,
		contractAddr,
		big.NewInt(0),
		gasLimit,
		big.NewInt(gasPrice),
		payload,
	)

	signedTx, err := c.SignTransaction(tx)
	if err != nil {
		return common.Hash{}, err
	}
	c.currentNonce += 1
	return c.SendRawTransaction(signedTx)
}

func (c *Client) SendTransactionAndDumpEvent(contract common.Address, payload []byte) error {
	hash, err := c.SendTransaction(contract, payload)
	if err != nil {
		return err
	}
	time.Sleep(blockPeriod)
	return c.DumpEventLog(hash)
}

func (c *Client) RepeatSendTransactionAndDumpEvent(contract common.Address, payload []byte, repeat int) error {
	hashList := make([]common.Hash, repeat)

	for i := 0; i < repeat; i++ {
		hash, err := c.SendTransaction(contract, payload)
		if err != nil {
			return err
		}
		hashList[i] = hash
	}

	time.Sleep(blockPeriod)

	for i := 0; i < repeat; i++ {
		if err := c.DumpEventLog(hashList[i]); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) SignTransaction(tx *types.Transaction) (string, error) {

	signer := types.HomesteadSigner{}
	signedTx, err := types.SignTx(
		tx,
		signer,
		c.Key.PrivateKey,
	)
	if err != nil {
		return "", fmt.Errorf("failed to sign tx: [%v]", err)
	}

	bz, err := rlp.EncodeToBytes(signedTx)
	if err != nil {
		return "", fmt.Errorf("failed to rlp encode bytes: [%v]", err)
	}
	return "0x" + hex.EncodeToString(bz), nil
}

func (c *Client) SendRawTransaction(signedTx string) (common.Hash, error) {
	var result common.Hash
	if err := c.Client.Call(&result, "eth_sendRawTransaction", signedTx); err != nil {
		return result, fmt.Errorf("failed to send raw transaction: [%v]", err)
	}

	return result, nil
}

func (c *Client) Address() common.Address {
	return c.Key.Address
}

func (c *Client) DumpEventLog(hash common.Hash) error {
	raw := &types.Receipt{}

	if err := c.Call(raw, "eth_getTransactionReceipt", hash.Hex()); err != nil {
		return fmt.Errorf("failed to get nonce: [%v]", err)
	}

	for _, event := range raw.Logs {
		log.Infof("eventlog address %s", event.Address.Hex())
		log.Infof("eventlog data %s", new(big.Int).SetBytes(event.Data).String())
		for i, topic := range event.Topics {
			log.Infof("eventlog topic[%d] %s", i, topic.String())
		}
	}
	return nil
}

func (c *Client) CallContract(caller, contractAddr common.Address, payload []byte, blockNum string) ([]byte, error) {
	var res hexutil.Bytes
	arg := ethereum.CallMsg{
		From: caller,
		To:   &contractAddr,
		Data: payload,
	}
	err := c.CallContext(context.Background(), &res, "eth_call", toCallArg(arg), blockNum)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func toCallArg(msg ethereum.CallMsg) interface{} {
	arg := map[string]interface{}{
		"from": msg.From,
		"to":   msg.To,
	}
	if len(msg.Data) > 0 {
		arg["data"] = hexutil.Bytes(msg.Data)
	}
	if msg.Value != nil {
		arg["value"] = (*hexutil.Big)(msg.Value)
	}
	if msg.Gas != 0 {
		arg["gas"] = hexutil.Uint64(msg.Gas)
	}
	if msg.GasPrice != nil {
		arg["gasPrice"] = (*hexutil.Big)(msg.GasPrice)
	}
	return arg
}
