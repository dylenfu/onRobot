package core

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/contracts/native"
	"github.com/ethereum/go-ethereum/contracts/native/plt"
	"github.com/ethereum/go-ethereum/contracts/native/utils"
	"github.com/palettechain/onRobot/config"
	"github.com/palettechain/onRobot/pkg/eth"
	"github.com/palettechain/onRobot/pkg/log"
	"math/big"
)

func PLTDeployPLTWrap() (succeed bool) {
	lockProxy := common.HexToAddress(native.PLTContractAddress)
	chainId := new(big.Int).SetUint64(config.Conf.CrossChain.PaletteSideChainID)
	ccAdmCli := getPaletteCli(pltCTypeInvoker)
	owner := ccAdmCli.Address()

	contractAddr, err := ccAdmCli.DeployPLTWrapper(owner, lockProxy, chainId)
	if err != nil {
		log.Errorf("deploy wrap on palette failed, err: %s", err.Error())
		return
	}

	//if err := config.Conf.CrossChain.StorePalettePLTWrapper(contractAddr); err != nil {
	//	log.Error("store palette wrapper failed")
	//	return
	//}

	log.Infof("deploy wrap %s on palette success!", contractAddr.Hex())
	return true
}

func PLTWrapperLock() (succeed bool) {
	var params struct {
		To        common.Address
		Amount    int
		ID        uint64
	}

	if err := config.LoadParams("PLT-Wrap-Lock.json", &params); err != nil {
		log.Error(err)
		return
	}

	asset := common.HexToAddress(native.PLTContractAddress)
	ethAsset := config.Conf.CrossChain.EthereumPLTAsset
	amount := plt.MultiPLT(params.Amount)
	toChainID := config.Conf.CrossChain.EthereumSideChainID
	if err := config.LoadParams("PLT-Wrap-Lock.json", &params); err != nil {
		log.Error(err)
		return
	}
	//cli := getPaletteCli(pltCTypeAdmin)
	cli := getPaletteCli(pltCTypeInvoker)
	from := cli.Address()
	to := params.To
	wrapAddr := config.Conf.CrossChain.PalettePLTWrapper
	invoker := eth.NewEInvoker(
		config.Conf.CrossChain.EthereumSideChainID,
		config.Conf.CrossChain.EthereumRPCUrl,
		customLoadAccount(from),
	)
	fee := big.NewInt(0)
	id := new(big.Int).SetUint64(params.ID)

	// approve
	logsplit()
	log.Info("approve wrapper allowance...")
	spender := wrapAddr
	if _, err := cli.PLTApprove(spender, amount); err != nil {
		log.Errorf("approve wrapper failed: %v", err)
		return
	}
	curAmt, err := cli.PLTAllowance(from, spender, "latest")
	if err != nil {
		log.Errorf("get plt allowance failed, err: %v")
		return
	}
	if curAmt.Cmp(amount) >= 0 {
		log.Infof("(owner, spender) (%s, %s) allowance %d enough", cli.Address().Hex(), spender.Hex(), plt.PrintUPLT(curAmt))
	} else {
		log.Errorf("allowance not enough, current allowance %d", plt.PrintUPLT(curAmt))
		return
	}

	//// try to wrap lock
	//logsplit()
	//log.Info("try to use wrapper contract to lock...")
	//if _, err := cli.PLTWrapLock(wrapAddr, asset, params.To, toChainID, amount, fee, id); err != nil {
	//	log.Errorf("palette wrapper lock failed, err: %v", err)
	//	return false
	//}
	//
	//log.Infof("wrap lock success!")

	// lock
	logsplit()
	log.Infof("lock plt on palette......")
	fromBalanceBeforeLockOnPalette, err := cli.BalanceOf(from, "latest")
	if err != nil {
		log.Error(err)
		return
	}
	toBalanceBeforeLockOnEth, err := invoker.PLTBalanceOf(ethAsset, to)
	if err != nil {
		log.Error(err)
		return
	}
	if _, err := cli.PLTWrapLock(wrapAddr, asset, to, toChainID, amount, fee, id); err != nil {
		log.Errorf("palette wrapper lock failed, err: %v", err)
		return false
	}

	logsplit()
	log.Info("check balance on both of palette chain and ethereum chain...")
	for i := 0; i < 100; i++ {
		fromBalanceAfterLockOnPalette, err := cli.BalanceOf(from, "latest")
		if err != nil {
			log.Error(err)
			return
		}
		toBalanceAfterLockOnEth, err := invoker.PLTBalanceOf(ethAsset, to)
		if err != nil {
			log.Error(err)
			return
		}

		log.Infof("palette %s: balance before lock [%d], balance after lock [%d]",
			params.To.Hex(),
			plt.PrintUPLT(fromBalanceBeforeLockOnPalette),
			plt.PrintUPLT(fromBalanceAfterLockOnPalette),
		)
		log.Infof("ethereum %s: balance before lock [%d], balance after lock [%d]",
			from.Hex(),
			plt.PrintUPLT(toBalanceBeforeLockOnEth),
			plt.PrintUPLT(toBalanceAfterLockOnEth),
		)

		subFrom := utils.SafeSub(fromBalanceBeforeLockOnPalette, fromBalanceAfterLockOnPalette)
		subTo := utils.SafeSub(toBalanceAfterLockOnEth, toBalanceBeforeLockOnEth)
		zero := big.NewInt(0)
		if new(big.Int).Sub(subFrom, amount).Cmp(zero) == 0 && new(big.Int).Sub(subTo, amount).Cmp(zero) == 0 {
			log.Info("lock tx hash success!")
			break
		}
		logsplit()
		wait(1)
	}

	return true
}

func PLTWrapperUnpackLockEvent() (succeed bool) {
	var params struct {
		Hash string
	}

	if err := config.LoadParams("PLT-Lock-Event.json", &params); err != nil {
		log.Error(err)
		return
	}

	hash := common.HexToHash(params.Hash)
	cli := getPaletteCli(pltCTypeInvoker)
	fromAsset, fromAddr, toAsset, toAddr, toChainID, amount, err := cli.GetPaletteLockEvent(hash)
	if err != nil {
		log.Errorf("failed to get lock proxy event, %v", err)
		return false
	}
	log.Infof("get lock event success, fromAsset %s, fromAddr %s, toAsset %s, toAddr %s, toChainID %d, amount %s",
		fromAsset.Hash(), fromAddr.Hex(), toAsset.Hex(), toAddr.Hex(), toChainID, amount.String())
	return true
}

func PLTWrapperUnpackUnlockEvent() (succeed bool) {
	var params struct {
		Hash string
	}

	if err := config.LoadParams("PLT-UnLock-Event.json", &params); err != nil {
		log.Error(err)
		return
	}

	hash := common.HexToHash(params.Hash)
	cli := getPaletteCli(pltCTypeInvoker)
	toAddr, toAsset, amount, err := cli.GetPaletteUnlockEvent(hash)
	if err != nil {
		log.Errorf("failed to get lock proxy event, %v", err)
		return false
	}
	log.Infof("get unlock event success, toAddr %s, toAsset %s, amount %s", toAddr.Hex(), toAsset.Hex(), amount.String())
	return true
}
