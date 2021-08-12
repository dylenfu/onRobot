package sdk

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/contracts/native"
	"github.com/ethereum/go-ethereum/contracts/native/utils"
	"github.com/ethereum/go-ethereum/core/types"
	pltwpabi "github.com/palettechain/onRobot/pkg/plt_wrap_abi"
)

func (c *Client) DeployPLTWrapper(owner, lockProxy common.Address, chainId *big.Int) (common.Address, error) {
	auth := c.makeDeployAuth()
	addr, tx, _, err := pltwpabi.DeployPolyWrapper(auth, c.backend, owner, lockProxy, chainId)
	if err != nil {
		return utils.EmptyAddress, err
	}
	if err := c.WaitTransaction(tx.Hash()); err != nil {
		return utils.EmptyAddress, err
	}
	return addr, nil
}

func (c *Client) PLTWrapLock(wrapAddr, fromAsset, toAddr common.Address, toChainId uint64, amount, fee, id *big.Int) (common.Hash, error) {
	wrapper, err := pltwpabi.NewPolyWrapper(wrapAddr, c.backend)
	if err != nil {
		return utils.EmptyHash, err
	}

	auth := c.makeDeployAuth()
	tx, err := wrapper.Lock(auth, fromAsset, toChainId, toAddr.Bytes(), amount, fee, id)
	if err != nil {
		return utils.EmptyHash, err
	}

	if err := c.WaitTransaction(tx.Hash()); err != nil {
		return utils.EmptyHash, err
	}
	return tx.Hash(), nil
}

var (
	abiLockEvent   abi.Event
	abiUnLockEvent abi.Event
)

const (
	pltProxyAbiJsonStr = `[
	{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"fromAssetHash","type":"address"},{"indexed":false,"internalType":"address","name":"fromAddress","type":"address"},{"indexed":false,"internalType":"uint64","name":"toChainId","type":"uint64"},{"indexed":false,"internalType":"bytes","name":"toAssetHash","type":"bytes"},{"indexed":false,"internalType":"bytes","name":"toAddress","type":"bytes"},{"indexed":false,"internalType":"uint256","name":"amount","type":"uint256"}],"name":"lock","type":"event"},
	{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"toAssetHash","type":"address"},{"indexed":false,"internalType":"address","name":"toAddress","type":"address"},{"indexed":false,"internalType":"uint256","name":"amount","type":"uint256"}],"name":"unlock","type":"event"}
]`
)

type LockEvent struct {
	FromAssetHash common.Address
	FromAddress   common.Address
	ToChainId     uint64
	ToAssetHash   common.Address
	ToAddress     common.Address
	Amount        *big.Int
}

type UnlockEvent struct {
	ToAssetHash common.Address
	ToAddress   common.Address
	Amount      *big.Int
}

func init() {
	ab, err := abi.JSON(strings.NewReader(pltProxyAbiJsonStr))
	if err != nil {
		panic(err)
	} else {
		abiLockEvent = ab.Events["lock"]
		abiUnLockEvent = ab.Events["unlock"]
	}
}

func (c *Client) GetPaletteLockEvent(hash common.Hash) (fromAsset, fromAddress, toAsset, toAddress common.Address, chainID uint64, amount *big.Int, err error) {
	var (
		receipt   *types.Receipt
		event     *types.Log
		lockEvent *LockEvent
		proxyAddr = common.HexToAddress(native.PLTContractAddress)
	)

	if receipt, err = c.backend.TransactionReceipt(context.Background(), hash); err != nil {
		return
	}
	if length := len(receipt.Logs); length < 3 {
		err = fmt.Errorf("invalid receipt %s, logs length expect 3, got %d", hash.Hex(), length)
	}

	for _, e := range receipt.Logs {
		eid := common.BytesToHash(e.Topics[0][:])
		if eid != abiLockEvent.ID() {
			continue
		} else {
			event = e
		}
	}
	if event == nil {
		err = fmt.Errorf("can not find proxy unlock event")
		return
	}
	if event.Address != proxyAddr {
		err = fmt.Errorf("expect proxy addr %s, got %s", proxyAddr.Hex(), event.Address.Hex())
	}

	if lockEvent, err = unpackLockEvent(event.Data, abiLockEvent); err != nil {
		return
	}
	fromAsset = lockEvent.FromAssetHash
	fromAddress = lockEvent.FromAddress
	toAsset = lockEvent.ToAssetHash
	toAddress = lockEvent.ToAddress
	chainID = lockEvent.ToChainId
	amount = lockEvent.Amount
	return
}

func (c *Client) GetPaletteUnlockEvent(hash common.Hash) (toAddress, toAsset common.Address, amount *big.Int, err error) {
	var (
		receipt     *types.Receipt
		event       *types.Log
		unlockEvent *UnlockEvent
		proxyAddr   = common.HexToAddress(native.PLTContractAddress)
	)

	if receipt, err = c.backend.TransactionReceipt(context.Background(), hash); err != nil {
		return
	}
	if length := len(receipt.Logs); length < 3 {
		err = fmt.Errorf("invalid receipt %s, logs length expect 3, got %d", hash.Hex(), length)
	}

	for _, e := range receipt.Logs {
		eid := common.BytesToHash(e.Topics[0][:])
		if eid != abiUnLockEvent.ID() {
			continue
		} else {
			event = e
		}
	}
	if event == nil {
		err = fmt.Errorf("can not find proxy unlock event")
		return
	}
	if event.Address != proxyAddr {
		err = fmt.Errorf("expect proxy addr %s, got %s", proxyAddr.Hex(), event.Address.Hex())
	}

	if unlockEvent, err = unpackUnlockEvent(event.Data, abiUnLockEvent); err != nil {
		return
	}
	toAddress = unlockEvent.ToAddress
	toAsset = unlockEvent.ToAssetHash
	amount = unlockEvent.Amount
	return
}

func unpackLockEvent(enc []byte, ab abi.Event) (*LockEvent, error) {
	event := new(LockEvent)
	if err := ab.Inputs.Unpack(event, enc); err != nil {
		return nil, err
	}
	return event, nil
}

func unpackUnlockEvent(enc []byte, ab abi.Event) (*UnlockEvent, error) {
	event := new(UnlockEvent)
	if err := ab.Inputs.Unpack(event, enc); err != nil {
		return nil, err
	}
	return event, nil
}
