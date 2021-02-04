package core

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/palettechain/onRobot/config"
	"github.com/palettechain/onRobot/pkg/log"
	"math/big"
)

func NFTLock() (succeed bool) {
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
