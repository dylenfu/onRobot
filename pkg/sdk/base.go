package sdk

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palettechain/onRobot/pkg/log"
)

func (c *PaletteClient) Balance(address string) (*big.Int, error) {
	var raw string

	if err := c.Call(
		&raw,
		"eth_getBalance",
		address,
		"latest",
	); err != nil {
		return nil, fmt.Errorf("faild to get balance [%v]", err)
	}

	src, _ := new(big.Int).SetString(raw, 0)
	data := new(big.Int).Div(src, OneEth)

	return data, nil
}

func (c *PaletteClient) GetNonce(address string) (uint64, error) {
	var raw string

	if err := c.Call(
		&raw,
		"eth_getTransactionCount",
		address,
		"latest",
	); err != nil {
		return 0, fmt.Errorf("failed to get nonce: [%v]", err)
	}

	without0xStr := strings.Replace(raw, "0x", "", -1)
	bigNonce, _ := new(big.Int).SetString(without0xStr, 16)
	return bigNonce.Uint64(), nil
}

func (c *PaletteClient) SignTransaction(key *ecdsa.PrivateKey, tx *types.Transaction) (string, error) {

	signer := types.HomesteadSigner{}
	signedTx, err := types.SignTx(
		tx,
		signer,
		key,
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

func (c *PaletteClient) SendRawTransaction(signedTx string) (common.Hash, error) {
	var result common.Hash
	if err := c.Client.Call(&result, "eth_sendRawTransaction", signedTx); err != nil {
		return result, fmt.Errorf("failed to send raw transaction: [%v]", err)
	}

	return result, nil
}

func (c *PaletteClient) AdminAddress() common.Address {
	return c.Admin.Address
}

func (c *PaletteClient) DumpEventLog(hash common.Hash) error {
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

func (c *PaletteClient) CallContract(from, to common.Address, payload []byte, blockNum string) ([]byte, error) {
	var res hexutil.Bytes
	arg := ethereum.CallMsg{
		From: from,
		To:   &to,
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

func TransferETH(nonce uint64, toAddress string, value *big.Int) *types.Transaction {
	return types.NewTransaction(
		nonce,
		common.HexToAddress(toAddress),
		value,
		GasMin,
		big.NewInt(GasPrice),
		nil,
	)
}
