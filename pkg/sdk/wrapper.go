package sdk

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/contracts/native/utils"
	wrap_abi "github.com/palettechain/onRobot/pkg/plt_wrap_abi"
)

func (c *Client) DeployPaletteWrapper(owner, feeToken common.Address, chainId *big.Int) (common.Address, error) {
	auth := c.makeDeployAuth()
	addr, tx, _, err := wrap_abi.DeployPolyWrapper(auth, c.backend, owner, feeToken, chainId)
	if err != nil {
		return utils.EmptyAddress, err
	}
	if err := c.WaitTransaction(tx.Hash()); err != nil {
		return utils.EmptyAddress, err
	}
	return addr, nil
}

func (c *Client) PaletteWrapSetLockProxy(wrapAddr, proxyAddr common.Address) (common.Hash, error) {
	wrapper, err := wrap_abi.NewPolyWrapper(wrapAddr, c.backend)
	if err != nil {
		return utils.EmptyHash, err
	}

	auth := c.makeDeployAuth()
	tx, err := wrapper.SetLockProxy(auth, proxyAddr)
	if err != nil {
		return utils.EmptyHash, err
	}

	if err := c.WaitTransaction(tx.Hash()); err != nil {
		return utils.EmptyHash, err
	}
	return tx.Hash(), nil
}

func (c *Client) GetPaletteWrapLockProxy(wrapAddr common.Address) (common.Address, error) {
	wrapper, err := wrap_abi.NewPolyWrapper(wrapAddr, c.backend)
	if err != nil {
		return utils.EmptyAddress, err
	}

	return wrapper.LockProxy(nil)
}

func (c *Client) PaletteWrapLock(wrapAddr, fromAsset, toAddr common.Address, toChainId uint64, amount, fee, id *big.Int) (common.Hash, error) {
	wrapper, err := wrap_abi.NewPolyWrapper(wrapAddr, c.backend)
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
