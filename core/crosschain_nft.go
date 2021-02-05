package core

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/contracts/native/nft"
	"github.com/ethereum/go-ethereum/contracts/native/utils"
	"github.com/palettechain/onRobot/config"
	"github.com/palettechain/onRobot/pkg/log"
	"github.com/palettechain/onRobot/pkg/sdk"
)

// 在palette上lock，ethereum上unlock
// 1. palette validator[0] mint token to user `from`, 合约是validators[0]部署的，只有他有权限mint给相关用户
// 2. 用户授权给nft proxy合约
// 3. lock
// 4. 在两条链上检查余额
func NFTLock() (succeed bool) {
	var params struct {
		From     common.Address
		To       common.Address
		Asset    common.Address
		NFTAsset common.Address
		TokenID  uint64
		Uri      string
	}

	if err := config.LoadParams("NFT-Lock.json", &params); err != nil {
		log.Error(err)
		return
	}

	// cross chain params
	owner := valcli.Address()
	asset := params.Asset
	from := params.From
	to := params.To
	token := new(big.Int).SetUint64(params.TokenID)
	proxy := config.Conf.CrossChain.PaletteNFTProxy
	sideChainID := config.Conf.CrossChain.EthereumSideChainID
	amount := big.NewInt(1)

	// generate new sender
	baseUrl := config.Conf.Nodes[0].RPCAddr()
	privKey := config.LoadAccount(from.Hex())
	cli := sdk.NewSender(baseUrl, privKey)

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
	toBalanceBeforeLockOnEthereum, err := ethInvoker.NFTBalance(params.NFTAsset, to)
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
		toBalanceAfterLockOnEthereum, err := ethInvoker.NFTBalance(params.NFTAsset, to)
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
		Asset    common.Address
		TokenID  uint64
		Uri      string
		NeedMint bool
	}

	if err := config.LoadParams("NFT-Lock.json", &params); err != nil {
		log.Error(err)
		return
	}

	// 这个流程的大致轮廓是 palette nft -> palette nft proxy
	// 假设validator A depoly了一个nft合约，同时将token mint给了自己
	// safe transfer的时候
	asset := params.Asset
	sideChainID := config.Conf.CrossChain.EthereumSideChainID
	from := valcli.Address()
	to := from
	proxy := config.Conf.CrossChain.PaletteNFTProxy
	token := new(big.Int).SetUint64(params.TokenID)

	// mint
	if params.NeedMint {
		log.Info("mint token")
		owner := valcli.Address()
		if _, err := valcli.NFTMint(params.Asset, owner, token, params.Uri); err != nil {
			log.Error(err)
			return
		}
		// check owner
		actualOwner, err := valcli.NFTTokenOwner(asset, token, "latest")
		if err != nil {
			log.Error(err)
			return
		}
		if actualOwner != owner {
			log.Error("expect owner %s actual %s", owner.Hex(), actualOwner.Hex())
		}

		// check uri
		actualUri, err := valcli.NFTTokenURI(asset, token, "latest")
		if err != nil {
			log.Error(err)
			return
		}
		if actualUri != params.Uri {
			log.Errorf("expect uri %s, actual %s", params.Uri, actualUri)
			return
		}

		// check balance
		actualBalance, err := valcli.NFTBalance(asset, owner, "latest")
		if err != nil {
			log.Error(err)
			return
		}

		log.Infof("%s asset %s balance %d after mint, uri %s, nft proxy %s",
			owner.Hex(), asset.Hex(), actualBalance.Uint64(), actualUri, proxy.Hex())
	}

	// lock
	{
		logsplit()
		log.Info("lock token")

		balanceBeforeLock, err := nftBalance(asset, from)
		if err != nil {
			log.Error(err)
			return
		}
		if _, err := valcli.NFTSafeTransferFrom(asset, from, proxy, token, to, sideChainID); err != nil {
			log.Error(err)
			return
		}

		balanceAfterLock, err := nftBalance(asset, from)
		if err != nil {
			log.Error(err)
			return
		}

		log.Infof("balance before lock %d, balance after lock %d", balanceBeforeLock.Uint64(), balanceAfterLock.Uint64())
	}

	return true
}

//func NFTUnlock() (succeed bool) {
//	var params = struct {
//		UnlockTo     common.Address
//		Asset        common.Address
//		Proof        string
//		RawHeader    string
//		HeaderProof  string
//		CurRawHeader string
//		HeaderSig    string
//	}{}
//
//	if err := config.LoadParams("NFT-Unlock.json", &params); err != nil {
//		log.Error(err)
//		return
//	}
//
//	_b, _ := admcli.NFTBalance(params.Asset, params.UnlockTo, "latest")
//	balanceBeforeUnlock := _b.Uint64()
//
//	proof, _ := hexutil.Decode(params.Proof)
//	rawHeader, _ := hexutil.Decode(params.RawHeader)
//	headerProof, _ := hexutil.Decode(params.HeaderProof)
//	curRawHeader, _ := hexutil.Decode(params.CurRawHeader)
//	headerSig, _ := hexutil.Decode(params.HeaderSig)
//
//	eccm := config.Conf.CrossChain.EthereumECCM
//	hash, err := ethInvoker.VerifyAndExecuteTx(
//		eccm,
//		proof,
//		rawHeader,
//		headerProof,
//		curRawHeader,
//		headerSig,
//	)
//	if err != nil {
//		log.Error(err)
//		return
//	}
//
//	for i := 0; i < 10000; i++ {
//		_b, err = admcli.NFTBalance(params.Asset, params.UnlockTo, "latest")
//		if err != nil {
//			log.Error(err)
//			return
//		}
//		balanceAfterUnlock := _b.Uint64()
//		if balanceAfterUnlock > balanceBeforeUnlock {
//			subAmount := balanceAfterUnlock - balanceBeforeUnlock
//			log.Infof("balance before unlock %d, after unlock %d, the sub amount is %d, eth hash %s",
//				balanceBeforeUnlock, balanceAfterUnlock, subAmount, hash.Hex())
//			break
//		}
//		time.Sleep(3 * time.Second)
//	}
//
//	return true
//}
