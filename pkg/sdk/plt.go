package sdk

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/contracts/native/plt"
	"github.com/ethereum/go-ethereum/contracts/native/utils"
	"github.com/palettechain/onRobot/pkg/log"
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

func (c *Client) PLTName() (string, error) {
	payload, err := c.packPLT(plt.MethodName)
	if err != nil {
		return "", err
	}

	enc, err := c.callPLT(payload, "latest")
	if err != nil {
		return "", err
	}

	output := new(plt.MethodNameOutput)
	if err := c.unpackPLT(plt.MethodName, output, enc); err != nil {
		return "", err
	}
	return output.Name, nil
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
		return utils.EmptyHash, err
	}
	return c.sendPLTTx(payload)
}

func (c *Client) PLTTransferFrom(from, to common.Address, amount *big.Int) (common.Hash, error) {
	payload, err := c.packPLT(plt.MethodTransferFrom, from, to, amount)
	if err != nil {
		return utils.EmptyHash, err
	}
	return c.sendPLTTx(payload)
}

func (c *Client) PLTApprove(spender common.Address, amount *big.Int) (common.Hash, error) {
	payload, err := c.packPLT(plt.MethodApprove, spender, amount)
	if err != nil {
		return utils.EmptyHash, err
	}
	return c.sendPLTTx(payload)
}

func (c *Client) PLTAllowance(owner, spender common.Address, blockNum string) (*big.Int, error) {
	payload, err := c.packPLT(plt.MethodAllowance, owner, spender)
	if err != nil {
		return nil, err
	}

	enc, err := c.callPLT(payload, blockNum)
	if err != nil {
		return nil, err
	}

	output := new(plt.MethodAllowanceOutput)
	if err := c.unpackPLT(plt.MethodAllowance, output, enc); err != nil {
		return nil, err
	}

	return output.Value, nil
}

func (self *Client) WaitTransaction(hash common.Hash) error {
	for {
		time.Sleep(time.Second * 1)
		_, ispending, err := self.backend.TransactionByHash(context.Background(), hash)
		if err != nil {
			log.Errorf("failed to call TransactionByHash: %v", err)
			continue
		}
		if ispending == true {
			continue
		}

		if err := self.DumpEventLog(hash); err != nil {
			return err
		}
		break
	}
	return nil
}

func (c *Client) packPLT(method string, args ...interface{}) ([]byte, error) {
	return utils.PackMethod(PLTABI, method, args...)
}
func (c *Client) unpackPLT(method string, output interface{}, enc []byte) error {
	return utils.UnpackOutputs(PLTABI, method, output, enc)
}
func (c *Client) sendPLTTx(payload []byte) (common.Hash, error) {
	hash, err := c.SendTransaction(PLTAddress, payload)
	if err != nil {
		return utils.EmptyHash, err
	}
	if err := c.WaitTransaction(hash); err != nil {
		return utils.EmptyHash, err
	}
	return hash, nil
}
func (c *Client) callPLT(payload []byte, blockNum string) ([]byte, error) {
	return c.CallContract(c.Address(), PLTAddress, payload, blockNum)
}
