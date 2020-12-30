package sdk

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/contracts/native/plt"
	"github.com/ethereum/go-ethereum/contracts/native/utils"
	"github.com/palettechain/onRobot/pkg/log"
	"github.com/polynetwork/eth-contracts/go_abi/eccm_abi"
	"github.com/polynetwork/eth-contracts/go_abi/eccmp_abi"
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

func (c *Client) PLTMint(to common.Address, val *big.Int) (common.Hash, error) {
	payload, err := c.packPLT(plt.MethodMint, to, val)
	if err != nil {
		return utils.EmptyHash, err
	}
	return c.sendPLTTx(payload)
}

func (c *Client) PLTBurn(val *big.Int) (common.Hash, error) {
	payload, err := c.packPLT(plt.MethodBurn, val)
	if err != nil {
		return utils.EmptyHash, err
	}
	return c.sendPLTTx(payload)
}

func (c *Client) PLTGetCCMP(blockNum string) (common.Address, error) {
	payload, err := c.packPLT(plt.MethodGetManagerProxy)
	if err != nil {
		return utils.EmptyAddress, err
	}

	enc, err := c.callPLT(payload, blockNum)
	if err != nil {
		return utils.EmptyAddress, err
	}

	var proxy common.Address
	if err := c.unpackPLT(plt.MethodGetManagerProxy, &proxy, enc); err != nil {
		return utils.EmptyAddress, err
	}

	return proxy, nil
}

func (c *Client) PLTSetCCMP(ccmp common.Address) (common.Hash, error) {
	payload, err := c.packPLT(plt.MethodSetManagerProxy, ccmp)
	if err != nil {
		return utils.EmptyHash, err
	}
	return c.sendPLTTx(payload)
}

func (c *Client) BindProxy(chainID uint64, proxy common.Address) (common.Hash, error) {
	payload, err := c.packPLT(plt.MethodBindProxy, chainID, proxy.Bytes())
	if err != nil {
		return utils.EmptyHash, err
	}

	return c.sendPLTTx(payload)
}

func (c *Client) GetBindProxy(chainID uint64, blockNum string) (common.Address, error) {
	payload, err := c.packPLT(plt.MethodGetBindedProxy, chainID)
	if err != nil {
		return utils.EmptyAddress, err
	}

	enc, err := c.callPLT(payload, blockNum)
	if err != nil {
		return utils.EmptyAddress, err
	}

	var proxy []byte
	if err := c.unpackPLT(plt.MethodGetBindedProxy, &proxy, enc); err != nil {
		return utils.EmptyAddress, err
	}

	return common.BytesToAddress(proxy), nil
}

func (c *Client) BindAsset(chainID uint64, asset common.Address) (common.Hash, error) {
	payload, err := c.packPLT(plt.MethodBindAsset, chainID, asset.Bytes())
	if err != nil {
		return utils.EmptyHash, err
	}

	return c.sendPLTTx(payload)
}

func (c *Client) GetBindAsset(chainID uint64, blockNum string) (common.Address, error) {
	payload, err := c.packPLT(plt.MethodGetBindedAsset, chainID)
	if err != nil {
		return utils.EmptyAddress, err
	}

	enc, err := c.callPLT(payload, blockNum)
	if err != nil {
		return utils.EmptyAddress, err
	}

	var asset []byte
	if err := c.unpackPLT(plt.MethodGetBindedAsset, &asset, enc); err != nil {
		return utils.EmptyAddress, err
	}

	return common.BytesToAddress(asset), nil
}

func (c *Client) Lock(chainID uint64, dstAddr common.Address, amount *big.Int) (common.Hash, error) {
	payload, err := c.packPLT(plt.MethodLock, chainID, dstAddr.Bytes(), amount)
	if err != nil {
		return utils.EmptyHash, err
	}
	return c.sendPLTTx(payload)
}

func (c *Client) InitGenesisBlock(eccmAddr common.Address, rawHdr, publickeys []byte) (common.Hash, error) {
	eccm, err := eccm_abi.NewEthCrossChainManager(eccmAddr, c.backend)
	if err != nil {
		return utils.EmptyHash, fmt.Errorf("new EthCrossChainManager err: %s", err)
	}

	auth := c.getBindAuth()
	tx, err := eccm.InitGenesisBlock(auth, rawHdr, publickeys)
	if err != nil {
		return utils.EmptyHash, fmt.Errorf("call eccm InitGenesisBlock err: %s", err)
	}

	c.WaitTransaction(tx.Hash())
	return tx.Hash(), nil
}

func (c *Client) PauseCCMP(ccmpAddr common.Address) (common.Hash, error) {
	ccmp, err := eccmp_abi.NewEthCrossChainManagerProxy(ccmpAddr, c.backend)
	if err != nil {
		return utils.EmptyHash, fmt.Errorf("new EthCrossChainManagerProxy err: %s", err)
	}
	//eccm, err := eccm_abi.NewEthCrossChainManager(oldeccmAddr, c.backend)
	//if err != nil {
	//	return utils.EmptyHash, fmt.Errorf("call eccd paused err: %s", err)
	//}

	//callOpts := c.getCallOpts()
	//ccmpPaused, err := ccmp.Paused(callOpts)
	//if err != nil {
	//	return utils.EmptyHash, fmt.Errorf("call ccmp paused err: %s", err)
	//}
	//eccmPaused, err := eccm.Paused(callOpts)
	//if err != nil {
	//	return utils.EmptyHash, fmt.Errorf("call eccm paused err: %s", err)
	//}

	auth := c.getBindAuth()
	tx, err := ccmp.PauseEthCrossChainManager(auth)
	if err != nil {
		return utils.EmptyHash, fmt.Errorf("call ccmp pause err: %s", err)
	}

	return tx.Hash(), nil
}

func (c *Client) UnPauseCCMP(ccmpAddr common.Address) (common.Hash, error) {
	ccmp, err := eccmp_abi.NewEthCrossChainManagerProxy(ccmpAddr, c.backend)
	if err != nil {
		return utils.EmptyHash, fmt.Errorf("new EthCrossChainManagerProxy err: %s", err)
	}

	auth := c.getBindAuth()
	tx, err := ccmp.UnpauseEthCrossChainManager(auth)
	if err != nil {
		return utils.EmptyHash, fmt.Errorf("call ccmp unpause err: %s", err)
	}

	return tx.Hash(), nil
}

func (c *Client) UpgradeECCM(newEccmAddr, ccmpAddr common.Address) (common.Hash, error) {
	ccmp, err := eccmp_abi.NewEthCrossChainManagerProxy(ccmpAddr, c.backend)
	if err != nil {
		return utils.EmptyHash, fmt.Errorf("new EthCrossChainManagerProxy err: %s", err)
	}

	auth := c.getBindAuth()
	tx, err := ccmp.UpgradeEthCrossChainManager(auth, newEccmAddr)
	if err != nil {
		return utils.EmptyHash, fmt.Errorf("call upgradeEthCrossChainManager err: %s", err)
	}

	return tx.Hash(), nil
}

//func (c *Client) WaitTransaction(txhash common.Hash) error {
//	for {
//		time.Sleep(time.Second * 2)
//		receipt, err := c.GetReceipt(txhash)
//		if err != nil {
//			log.Errorf("failed to call TransactionByHash: %v", err)
//			continue
//		}
//		if receipt == nil {
//			continue
//		}
//		if receipt.Status > 0 {
//			c.DumpEventLog(txhash)
//			break
//		} else {
//			return fmt.Errorf("tx %s failed", txhash.Hex())
//		}
//	}
//	return nil
//}

func (self *Client) WaitTransaction(hash common.Hash) {
	for {
		time.Sleep(time.Second * 1)
		_, ispending, err := self.backend.TransactionByHash(context.Background(), hash)
		if err != nil {
			log.Errorf("failed to call TransactionByHash: %v", err)
			continue
		}
		if ispending == true {
			continue
		} else {
			break
		}
	}
}

//func (c *Client) UnLock(args *plt.TxArgs, srcContract common.Address, chainID uint64) (common.Hash, error) {
//	var buf []byte
//	sink := pcom.NewZeroCopySink(buf)
//	args.Serialization(sink)
//	payload, err := c.packPLT(plt.MethodUnlock, buf, srcContract.Bytes(), chainID)
//	if err != nil {
//		return utils.EmptyHash, err
//	}
//	return c.sendPLTTx(payload)
//}

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
