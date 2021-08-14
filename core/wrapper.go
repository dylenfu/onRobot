package core

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/contracts/native"
	"github.com/ethereum/go-ethereum/contracts/native/nft"
	"github.com/ethereum/go-ethereum/contracts/native/plt"
	"github.com/ethereum/go-ethereum/contracts/native/utils"
	"github.com/palettechain/onRobot/config"
	"github.com/palettechain/onRobot/pkg/eth"
	"github.com/palettechain/onRobot/pkg/log"
	"github.com/palettechain/onRobot/pkg/sdk"
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
		From   common.Address
		To     common.Address
		Amount int
	}

	if err := config.LoadParams("PLT-Lock.json", &params); err != nil {
		log.Error(err)
		return
	}

	baseUrl := config.Conf.Nodes[0].RPCAddr()
	privKey := customLoadAccount(params.From)
	cli := sdk.NewSender(baseUrl, privKey)
	from := cli.Address()
	invoker := eth.NewEInvoker(
		config.Conf.CrossChain.EthereumSideChainID,
		config.Conf.CrossChain.EthereumRPCUrl,
		customLoadAccount(from),
	)

	wrapAddr := config.Conf.CrossChain.PalettePLTWrapper
	asset := common.HexToAddress(native.PLTContractAddress)
	ethAsset := config.Conf.CrossChain.EthereumPLTAsset
	amount := plt.MultiPLT(params.Amount)
	toChainID := config.Conf.CrossChain.EthereumSideChainID
	to := params.To
	fee := big.NewInt(1234500000)
	id := big.NewInt(1)

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
		log.Infof("ethereum %s: balance before lock [%s], balance after lock [%s]",
			from.Hex(),
			toBalanceBeforeLockOnEth.String(),
			toBalanceAfterLockOnEth.String(),
		)

		if fromBalanceBeforeLockOnPalette.Cmp(fromBalanceAfterLockOnPalette) < 0 &&
			toBalanceAfterLockOnEth.Cmp(toBalanceBeforeLockOnEth) > 0 {
			log.Info("lock tx hash success!")
			break
		}
		logsplit()
		wait(1)
	}

	return true
}

// 在palette上lock，ethereum上unlock
// 1. palette validator[0] mint token to user `from`, 合约是validators[0]部署的，只有他有权限mint给相关用户
// 2. lock之前并不需要授权给nft proxy，因为safeTransferFrom本身是将nft转账给proxy，to只会打包到data args中传过去
// 3. 在两条链上检查余额
func NFTWrapLock() (succeed bool) {
	var params struct {
		From        common.Address
		To          common.Address
		PLTNFTAsset common.Address
		ETHNFTAsset common.Address
		TokenID     uint64
		Uri         string
	}

	if err := config.LoadParams("NFT-Lock.json", &params); err != nil {
		log.Error(err)
		return
	}

	// cross chain params
	valcli := getPaletteCli(pltCTypeCrossChainAdmin)
	owner := valcli.Address()
	asset := params.PLTNFTAsset
	from := params.From
	to := params.To
	tokenID := new(big.Int).SetUint64(params.TokenID)
	feeToken := common.HexToAddress(native.PLTContractAddress)
	sideChainID := config.Conf.CrossChain.EthereumSideChainID
	amount := big.NewInt(1)
	wrap := config.Conf.CrossChain.PaletteNFTWrapper
	fee := big.NewInt(123450000000000)
	id := big.NewInt(1)

	// generate new sender
	baseUrl := config.Conf.Nodes[0].RPCAddr()
	privKey := customLoadAccount(from)
	cli := sdk.NewSender(baseUrl, privKey)
	ethInvoker := getEthereumCli(ethCTypeInvoker)

	// mint or transfer ownership
	{
		logsplit()
		log.Infof("mint if tokenID not exist or ownership is not user `from`......")
		preOwner, err := valcli.NFTTokenOwner(asset, tokenID, "latest")
		if preOwner != utils.EmptyAddress && preOwner != from {
			if _, err := valcli.NFTTransferFrom(asset, owner, from, tokenID); err != nil {
				log.Errorf("transfer nft ownership err: %s", err.Error())
				return
			} else {
				log.Infof("%s transfer tokenID%d's ownership to %s on asset %s", owner.Hex(), tokenID.Uint64(), from.Hex(), asset.Hex())
			}
		}
		if err != nil && err.Error() == nft.NOT_VALID_NFT {
			if _, err := valcli.NFTMint(asset, from, tokenID, params.Uri); err != nil {
				log.Errorf("mint tokenID on palette err: %s", err.Error())
				return
			} else {
				log.Infof("%s mint tokenID%d to %s on asset %s, uri is %s", owner.Hex(), tokenID.Uint64(), from.Hex(), asset.Hex(), params.Uri)
			}
		}

		// check ownership
		curOwner, err := cli.NFTTokenOwner(asset, tokenID, "latest")
		if err != nil {
			log.Error(err)
			return
		}
		if curOwner != from {
			log.Errorf("tokenID%d current owner %s!=%s", tokenID.Uint64(), curOwner.Hex(), from.Hex())
		} else {
			log.Infof("tokenID%d current owner is %s", tokenID.Uint64(), from.Hex())
		}
	}

	// approve
	logsplit()
	log.Info("approve wrapper allowance...")
	spender := wrap
	if _, err := cli.NFTTokenApprove(asset, spender, amount); err != nil {
		log.Errorf("approve wrapper failed: %v", err)
		return
	}
	cur, err := cli.NFTTokenAllowance(asset, tokenID, "latest")
	if err != nil {
		log.Errorf("get plt allowance failed, err: %v")
		return
	}
	if cur != spender {
		log.Errorf("nft approved expect %s, got %s", spender.Hex(), cur.Hex())
		return
	}

	// lock
	logsplit()
	log.Info("lock tokenID.....")
	fromBalanceBeforeLockOnPalette, err := cli.NFTBalance(asset, from, "latest")
	if err != nil {
		log.Error(err)
		return
	}
	toBalanceBeforeLockOnEthereum, err := ethInvoker.NFTBalance(params.ETHNFTAsset, to)
	if err != nil {
		log.Error(err)
		return
	}

	hash, err := cli.NFTWrapLock(wrap, asset, to, feeToken, sideChainID, tokenID, fee, id)
	if err != nil {
		log.Error(err)
		return
	}

	for i := 0; i < 100; i++ {
		fromBalanceAfterLockOnPalette, err := cli.NFTBalance(asset, from, "latest")
		if err != nil {
			log.Error(err)
			return
		}
		toBalanceAfterLockOnEthereum, err := ethInvoker.NFTBalance(params.ETHNFTAsset, to)
		if err != nil {
			log.Error(err)
			return
		}

		log.Infof("palette %s: balance before lock [%d], balance after lock [%d]",
			params.From.Hex(),
			fromBalanceBeforeLockOnPalette.Uint64(),
			fromBalanceAfterLockOnPalette.Uint64(),
		)
		log.Infof("ethereum %s: balance before lock [%d], balance after lock [%d]",
			params.To.Hex(),
			toBalanceBeforeLockOnEthereum.Uint64(),
			toBalanceAfterLockOnEthereum.Uint64(),
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

	uri, err := ethInvoker.NFTTokenUri(params.ETHNFTAsset, tokenID)
	if err != nil {
		log.Errorf("can not find uri")
		return
	}
	if uri != params.Uri {
		log.Errorf("expect uri %s, actual %s", params.Uri, uri)
		return
	}
	log.Infof("expect uri %s == actual uri %s", params.Uri, uri)
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
