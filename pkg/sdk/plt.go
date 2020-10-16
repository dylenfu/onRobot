package sdk

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/contracts/native/plt"
	"github.com/ethereum/go-ethereum/contracts/native/utils"
)

func (c *Client) BalanceOf(owner common.Address, blockNum string) (*big.Int, error) {
	payload, err := c.packPLT(plt.MethodBalanceOf, owner)
	if err != nil {
		return nil, err
	}

	enc, err := c.callPLT(payload, blockNum)
	if err != nil {
		return nil, err
	}

	output := new(plt.MethodBalanceOfOutput)
	if err := c.unpackPLT(plt.MethodBalanceOf, output, enc); err != nil {
		return nil, err
	}

	return output.Balance, nil
}

func (c *Client) PLTTotalSupply(blockNum string) (*big.Int, error) {
	payload, err := c.packPLT(plt.MethodTotalSupply)
	if err != nil {
		return nil, err
	}

	enc, err := c.callPLT(payload, blockNum)
	if err != nil {
		return nil, err
	}

	output := new(plt.MethodTotalSupplyOutput)
	if err := c.unpackPLT(plt.MethodTotalSupply, output, enc); err != nil {
		return nil, err
	}
	return output.Supply, nil
}

func (c *Client) PLTDecimals() (uint64, error) {
	payload, err := c.packPLT(plt.MethodDecimals)
	if err != nil {
		return 0, err
	}

	enc, err := c.callPLT(payload, "latest")
	if err != nil {
		return 0, err
	}

	output := new(plt.MethodDecimalsOutput)
	if err := c.unpackPLT(plt.MethodDecimals, output, enc); err != nil {
		return 0, err
	}
	return output.Decimal.Uint64(), nil
}

func (c *Client) PLTTransfer(to common.Address, amount *big.Int) (common.Hash, error) {
	payload, err := c.packPLT(plt.MethodTransfer, to, amount)
	if err != nil {
		return common.Hash{}, err
	}
	return c.sendPLTTx(payload)
}

func (c *Client) PLTTransferFrom(from, to common.Address, amount *big.Int) (common.Hash, error) {
	payload, err := utils.PackMethod(PLTABI, plt.MethodTransferFrom, from, to, amount)
	if err != nil {
		return common.Hash{}, err
	}
	return c.sendPLTTx(payload)
}

func (c *Client) PLTApprove(spender common.Address, amount *big.Int) (common.Hash, error) {
	payload, err := utils.PackMethod(PLTABI, plt.MethodApprove, spender, amount)
	if err != nil {
		return common.Hash{}, err
	}
	return c.sendPLTTx(payload)
}

func (c *Client) packPLT(method string, args ...interface{}) ([]byte, error) {
	return utils.PackMethod(PLTABI, method, args...)
}
func (c *Client) unpackPLT(method string, output interface{}, enc []byte) error {
	return utils.UnpackOutputs(PLTABI, method, output, enc)
}
func (c *Client) sendPLTTx(payload []byte) (common.Hash, error) {
	return c.SendTransaction(PLTAddress, payload)
}
func (c *Client) callPLT(payload []byte, blockNum string) ([]byte, error) {
	return c.CallContract(c.Address(), PLTAddress, payload, blockNum)
}