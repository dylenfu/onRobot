package core

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/contracts/native/utils"
	"github.com/palettechain/onRobot/config"
	"github.com/palettechain/onRobot/pkg/log"
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
	sideChainID := uint64(config.Conf.CrossChain.EthereumSideChainID)
	from := valcli.Address()
	to := from
	proxy := config.Conf.CrossChain.PaletteNFTProxy
	token := new(big.Int).SetUint64(params.TokenID)

	// mint
	if params.NeedMint {
		log.Info("mint token")
		owner := valcli.Address()
		hash, err := valcli.NFTMint(params.Asset, owner, token, params.Uri)
		if err != nil {
			log.Error(err)
			return
		}

		wait(3)
		if err := valcli.DumpEventLog(hash); err != nil {
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
		hash, err := valcli.NFTSafeTransferFrom(asset, from, proxy, token, to, sideChainID)
		if err != nil {
			log.Error(err)
			return
		}

		wait(2)
		if err := valcli.DumpEventLog(hash); err != nil {
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

	// waiting for unlock
	{
		logsplit()
		log.Infof("unlock nft...")

		var (
			balanceBeforeUnlock,
			balanceAfterUnlock *big.Int
		)
		for i := 0; i < 10000; i++ {
			balance, err := nftBalance(asset, from)
			if err != nil {
				log.Error(err)
				return
			}
			if i == 0 {
				balanceBeforeUnlock = balance
				log.Infof("waiting for unlock")
			} else if balance.Cmp(balanceBeforeUnlock) > 0 {
				balanceAfterUnlock = balance
				subAmount := utils.SafeSub(balanceAfterUnlock, balanceBeforeUnlock)
				log.Infof("balance before unlock %d, after unlock %d, the sub amount is %d",
					balanceBeforeUnlock.Uint64(), balanceAfterUnlock.Uint64(), subAmount.Uint64())
				break
			}
			time.Sleep(3 * time.Second)
		}
	}

	return true
}
