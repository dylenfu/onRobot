package core

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/contracts/native/nft"
	"github.com/ethereum/go-ethereum/contracts/native/utils"
	"github.com/palettechain/onRobot/config"
	"github.com/palettechain/onRobot/pkg/eth"
	"github.com/palettechain/onRobot/pkg/log"
	"github.com/palettechain/onRobot/pkg/sdk"
)

// 在palette上lock，ethereum上unlock
// 1. palette validator[0] mint token to user `from`, 合约是validators[0]部署的，只有他有权限mint给相关用户
// 2. lock之前并不需要授权给nft proxy，因为safeTransferFrom本身是将nft转账给proxy，to只会打包到data args中传过去
// 3. 在两条链上检查余额
func NFTLock() (succeed bool) {
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
	valcli := getPaletteCli(pltCTypeInvoker)
	owner := valcli.Address()
	asset := params.PLTNFTAsset
	from := params.From
	to := params.To
	token := new(big.Int).SetUint64(params.TokenID)
	proxy := config.Conf.CrossChain.PaletteNFTProxy
	sideChainID := config.Conf.CrossChain.EthereumSideChainID
	amount := big.NewInt(1)

	// generate new sender
	baseUrl := config.Conf.Nodes[0].RPCAddr()
	privKey := customLoadAccount(from)
	cli := sdk.NewSender(baseUrl, privKey)
	ethInvoker := getEthereumCli(ethCTypeInvoker)

	// mint or transfer ownership
	{
		logsplit()
		log.Infof("mint if token not exist or ownership is not user `from`......")
		preOwner, err := valcli.NFTTokenOwner(asset, token, "latest")
		if preOwner != utils.EmptyAddress && preOwner != from {
			if _, err := valcli.NFTTransferFrom(asset, owner, from, token); err != nil {
				log.Errorf("transfer nft ownership err: %s", err.Error())
				return
			} else {
				log.Infof("%s transfer token%d's ownership to %s on asset %s", owner.Hex(), token.Uint64(), from.Hex(), asset.Hex())
			}
		}
		if err != nil && err.Error() == nft.NOT_VALID_NFT {
			if _, err := valcli.NFTMint(asset, from, token, params.Uri); err != nil {
				log.Errorf("mint token on palette err: %s", err.Error())
				return
			} else {
				log.Infof("%s mint token%d to %s on asset %s, uri is %s", owner.Hex(), token.Uint64(), from.Hex(), asset.Hex(), params.Uri)
			}
		}

		// check ownership
		curOwner, err := cli.NFTTokenOwner(asset, token, "latest")
		if err != nil {
			log.Error(err)
			return
		}
		if curOwner != from {
			log.Errorf("token%d current owner %s!=%s", token.Uint64(), curOwner.Hex(), from.Hex())
		} else {
			log.Infof("token%d current owner is %s", token.Uint64(), from.Hex())
		}
	}

	// lock
	logsplit()
	log.Info("lock token.....")
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

	hash, err := cli.NFTSafeTransferFrom(asset, from, proxy, token, to, sideChainID)
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

	return true
}

func NFTUnLock() (succeed bool) {
	var params struct {
		From        common.Address
		To          common.Address
		PLTNFTAsset common.Address
		ETHNFTAsset common.Address
		TokenID     uint64
	}

	if err := config.LoadParams("NFT-UnLock.json", &params); err != nil {
		log.Error(err)
		return
	}

	// cross chain params
	asset := params.ETHNFTAsset
	targetAsset := params.PLTNFTAsset
	from := params.From
	to := params.To
	token := new(big.Int).SetUint64(params.TokenID)
	proxy := config.Conf.CrossChain.EthereumNFTProxy
	targetSideChainID := config.Conf.CrossChain.PaletteSideChainID
	amount := big.NewInt(1)
	cli := getPaletteCli(pltCTypeCustomer)

	invoker := eth.NewEInvoker(
		config.Conf.CrossChain.EthereumSideChainID,
		config.Conf.CrossChain.EthereumRPCUrl,
		customLoadAccount(from),
	)

	// check ownership
	curOwner, err := invoker.NFTOwner(asset, token)
	if err != nil {
		log.Error(err)
		return
	}
	if curOwner != from {
		log.Errorf("token%d current owner %s!=%s", token.Uint64(), curOwner.Hex(), from.Hex())
	} else {
		log.Infof("token%d current owner is %s", token.Uint64(), from.Hex())
	}

	// please make sure that eth account's balance is enough for gas fee.

	// lock
	logsplit()
	log.Info("unlock token.....")
	fromBalanceBeforeLockOnEthereum, err := invoker.NFTBalance(asset, from)
	if err != nil {
		log.Error(err)
		return
	}
	toBalanceBeforeLockOnPalette, err := cli.NFTBalance(targetAsset, to, "latest")
	if err != nil {
		log.Error(err)
		return
	}

	hash, err := invoker.NFTSafeTransferFrom(asset, from, proxy, token, to, targetSideChainID)
	if err != nil {
		log.Error(err)
		return
	}

	for i := 0; i < 100; i++ {
		fromBalanceAfterLockOnEthereum, err := invoker.NFTBalance(asset, from)
		if err != nil {
			log.Error(err)
			return
		}
		toBalanceAfterLockOnPalette, err := cli.NFTBalance(targetAsset, to, "latest")
		if err != nil {
			log.Error(err)
			return
		}

		log.Infof("ethereum %s: balance before lock [%d], balance after lock [%d]",
			params.From.Hex(),
			fromBalanceBeforeLockOnEthereum.Uint64(),
			fromBalanceAfterLockOnEthereum.Uint64(),
		)
		log.Infof("palette %s: balance before lock [%d], balance after lock [%d]",
			params.To.Hex(),
			toBalanceBeforeLockOnPalette.Uint64(),
			toBalanceAfterLockOnPalette.Uint64(),
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
