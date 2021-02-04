package core

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/contracts/native/plt"
	"github.com/ethereum/go-ethereum/contracts/native/utils"
	"github.com/palettechain/onRobot/config"
	"github.com/palettechain/onRobot/pkg/eth"
	"github.com/palettechain/onRobot/pkg/log"
	"github.com/palettechain/onRobot/pkg/sdk"
)

// 在palette native plt合约mint一定量的PLT token到某个已经存在的用户地址
func PLTMint() (succeed bool) {
	var params struct {
		To    common.Address
		Value int
	}
	if err := config.LoadParams("PLT-Mint.json", &params); err != nil {
		log.Error(err)
		return
	}
	amount := plt.MultiPLT(params.Value)

	toBalanceBeforeTrans, err := admcli.BalanceOf(params.To, "latest")
	if err != nil {
		log.Error(err)
		return
	}
	tx, err := admcli.PLTMint(params.To, amount)
	if err != nil {
		log.Error(err)
		return
	}
	wait(2)
	if err := admcli.DumpEventLog(tx); err != nil {
		log.Error(err)
		return
	}

	toBalanceAfterTrans, err := admcli.BalanceOf(params.To, "latest")
	if err != nil {
		log.Error(err)
		return
	}

	if utils.SafeSub(toBalanceAfterTrans, toBalanceBeforeTrans).Cmp(amount) != 0 {
		log.Errorf("wrong mint value %s and should be %s",
			utils.SafeSub(toBalanceAfterTrans, toBalanceBeforeTrans), params.Value)
	}
	return true
}

// 在palette native plt合约烧毁合约PLT总供应量对应的amount
func PLTBurn() (succeed bool) {
	var p struct {
		Value int
	}
	if err := config.LoadParams("PLT-Burn.json", &p); err != nil {
		log.Error(err)
		return
	}
	amount := plt.MultiPLT(p.Value)

	toBalanceBeforeTrans, err := admcli.BalanceOf(admcli.Address(), "latest")
	if err != nil {
		log.Error(err)
		return
	}
	if toBalanceBeforeTrans.Cmp(amount) == -1 {
		log.Error(err)
		return
	}

	tx, err := admcli.PLTBurn(amount)
	if err != nil {
		log.Error(err)
		return
	}
	wait(2)
	if err := admcli.DumpEventLog(tx); err != nil {
		log.Error(err)
		return
	}

	toBalanceAfterTrans, err := admcli.BalanceOf(admcli.Address(), "latest")
	if err != nil {
		log.Error(err)
		return
	}

	if utils.SafeSub(toBalanceBeforeTrans, toBalanceAfterTrans).Cmp(amount) != 0 {
		log.Errorf("wrong mint value %s and should be %s",
			utils.SafeSub(toBalanceAfterTrans, toBalanceBeforeTrans), p.Value)
	}

	return true
}

func PLTLock() (succeed bool) {
	var params struct {
		From   common.Address
		To     common.Address
		Amount int
	}
	if err := config.LoadParams("PLT-Lock.json", &params); err != nil {
		log.Error(err)
		return
	}

	baseUrl := config.Conf.Nodes[0].RPCAddr()
	privKey := config.LoadAccount(params.From.Hex())
	userAddr := params.From
	bindTo := params.To
	cli := sdk.NewSender(baseUrl, privKey)
	amount := plt.MultiPLT(params.Amount)
	targetSideChainID := config.Conf.CrossChain.EthereumSideChainID
	ethAsset := config.Conf.CrossChain.EthereumPLTAsset

	fromBalanceBeforeLockOnPalette, err := cli.BalanceOf(userAddr, "latest")
	if err != nil {
		log.Error(err)
		return
	}
	if fromBalanceBeforeLockOnPalette.Cmp(amount) < 0 {
		logsplit()
		log.Infof("prepare test account balance...")
		if _, err := admcli.PLTTransfer(userAddr, amount); err != nil {
			log.Error(err)
			return
		}
		fromBalanceBeforeLockOnPalette, _ = cli.BalanceOf(userAddr, "latest")
	}

	toBalanceBeforeLockOnEthereum, err := ethInvoker.PLTBalanceOf(ethAsset, bindTo)
	if err != nil {
		log.Error(err)
		return
	}

	logsplit()
	hash, err := cli.LockPLT(targetSideChainID, bindTo, amount)
	if err != nil {
		log.Errorf("failed to call `lock` err: %v", err)
		return
	} else {
		log.Infof("lock plt on palette, tx hash %s", hash.Hex())
	}

	logsplit()
	log.Info("check balance on both of palette chain and ethereum chain...")
	for i := 0; i < 100; i++ {
		fromBalanceAfterLockOnPalette, err := cli.BalanceOf(userAddr, "latest")
		if err != nil {
			log.Error(err)
			return
		}
		toBalanceAfterLockOnEthereum, err := ethInvoker.PLTBalanceOf(ethAsset, bindTo)
		if err != nil {
			log.Error(err)
			return
		}

		log.Infof("palette %s: balance before lock [%d], balance after lock [%d]",
			params.From.Hex(),
			plt.PrintUPLT(fromBalanceBeforeLockOnPalette),
			plt.PrintUPLT(fromBalanceAfterLockOnPalette),
		)
		log.Infof("ethereum %s: balance before lock [%d], balance after lock [%d]",
			params.To.Hex(),
			plt.PrintUPLT(toBalanceBeforeLockOnEthereum),
			plt.PrintUPLT(toBalanceAfterLockOnEthereum),
		)
		subFrom := utils.SafeSub(fromBalanceBeforeLockOnPalette, fromBalanceAfterLockOnPalette)
		subTo := utils.SafeSub(toBalanceAfterLockOnEthereum, toBalanceBeforeLockOnEthereum)
		zero := big.NewInt(0)
		if new(big.Int).Sub(subFrom, amount).Cmp(zero) == 0 && new(big.Int).Sub(subTo, amount).Cmp(zero) == 0 {
			log.Infof("lock tx hash %s success!", hash.Hex())
			break
		}
		logsplit()
		wait(1)
	}

	return true
}

// 以太坊lock对应到palette的unlock:
// 1.准备需要的eth作为gas
// 2.准备ethereum proxy需要的allowance
// 3.lock前查询from在以太上的余额
// 4.lock前查询to在palette上的余额
// 5.lock锁定
// 6.循环内查询lock后from在以太上的余额
// 7.循环内查询lock后to在palette上的余额
// 8.比较并判断是否成功
func PLTUnlock() (succeed bool) {
	var params struct {
		From   common.Address
		To     common.Address
		Amount int
	}
	if err := config.LoadParams("PLT-UnLock.json", &params); err != nil {
		log.Error(err)
		return
	}

	from := params.From
	to := params.To
	proxy := config.Conf.CrossChain.EthereumPLTProxy
	targetSideChainID := config.Conf.CrossChain.PaletteSideChainID
	asset := config.Conf.CrossChain.EthereumPLTAsset
	amount := plt.MultiPLT(params.Amount)

	invoker := eth.NewEInvoker(
		config.Conf.CrossChain.EthereumSideChainID,
		config.Conf.CrossChain.EthereumRPCUrl,
		config.LoadAccount(from.Hex()),
	)

	// prepare ETH for gas fee
	{
		logsplit()
		log.Infof("prepare eth gas fee......")
		gasLimit := 210000
		gasFee, err := calculateGasFee(invoker, uint64(gasLimit))
		if err != nil {
			log.Errorf("calculate gas fee err %s", err.Error())
		}
		amount := utils.SafeMul(gasFee, big.NewInt(2))
		if err := prepareEth(from, amount); err != nil {
			log.Errorf("prepare eth as gas failed, err: %s", err.Error())
			return
		}
	}

	// prepare allowance
	logsplit()
	log.Infof("prepare from account allowance for proxy......")
	if err := prepareAllowance(invoker, from, proxy, amount); err != nil {
		log.Error(err)
		return
	}

	// unlock
	logsplit()
	log.Infof("lock plt on ethereum......")
	fromBalanceBeforeLockOnEthereum, err := invoker.PLTBalanceOf(asset, from)
	if err != nil {
		log.Error(err)
		return
	}
	toBalanceBeforeLockOnPalette, err := admcli.BalanceOf(to, "latest")
	if err != nil {
		log.Error(err)
		return
	}
	hash, err := invoker.PLTLock(proxy, asset, targetSideChainID, to, amount)
	if err != nil {
		log.Error(err)
		return
	} else {
		log.Infof("lock plt on ethereum, tx hash %s", hash.Hex())
	}

	logsplit()
	log.Info("check balance on both of palette chain and ethereum chain...")
	for i := 0; i < 100; i++ {
		fromBalanceAfterLockOnEthereum, err := ethInvoker.PLTBalanceOf(asset, from)
		if err != nil {
			log.Error(err)
			return
		}
		toBalanceAfterLockOnPalette, err := admcli.BalanceOf(to, "latest")
		if err != nil {
			log.Error(err)
			return
		}

		log.Infof("ethereum %s: balance before lock [%d], balance after lock [%d]",
			params.From.Hex(),
			plt.PrintUPLT(fromBalanceBeforeLockOnEthereum),
			plt.PrintUPLT(fromBalanceAfterLockOnEthereum),
		)
		log.Infof("palette %s: balance before lock [%d], balance after lock [%d]",
			params.To.Hex(),
			plt.PrintUPLT(toBalanceBeforeLockOnPalette),
			plt.PrintUPLT(toBalanceAfterLockOnPalette),
		)
		subFrom := utils.SafeSub(fromBalanceBeforeLockOnEthereum, fromBalanceAfterLockOnEthereum)
		subTo := utils.SafeSub(toBalanceAfterLockOnPalette, toBalanceBeforeLockOnPalette)
		zero := big.NewInt(0)
		if new(big.Int).Sub(subFrom, amount).Cmp(zero) == 0 && new(big.Int).Sub(subTo, amount).Cmp(zero) == 0 {
			log.Infof("lock tx hash %s success!", hash.Hex())
			break
		}
		logsplit()
		wait(1)
	}

	return true
}
