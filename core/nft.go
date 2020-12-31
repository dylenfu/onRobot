package core

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/contracts/native/utils"
	"github.com/palettechain/onRobot/config"
	"github.com/palettechain/onRobot/pkg/log"
	"github.com/palettechain/onRobot/pkg/sdk"
	//polycm "github.com/polynetwork/poly/common"
	"math/big"
	"time"
)

func NFTDeploy() (succeed bool) {
	var params struct {
		Name   string
		Symbol string
	}

	if err := config.LoadParams("NFT-Deploy.json", &params); err != nil {
		log.Error(err)
		return
	}

	_, addr, err := valcli.NFTDeploy(params.Name, params.Symbol)
	if err != nil {
		log.Error(err)
		return
	}

	log.Infof("deploy nft %s success, address %s", params.Symbol, addr.Hex())
	return true
}

func NFTMint() (succeed bool) {
	var params struct {
		Asset   common.Address
		TokenID uint64
		Uri     string
	}

	if err := config.LoadParams("NFT-Mint.json", &params); err != nil {
		log.Error(err)
		return
	}

	owner := valcli.Address()
	token := new(big.Int).SetUint64(params.TokenID)
	balanceBeforeMint, err := nftBalance(params.Asset, owner)
	if err != nil {
		log.Error(err)
		return
	}

	{
		hash, err := valcli.NFTMint(params.Asset, owner, token, params.Uri)
		if err != nil {
			log.Error(err)
			return
		}

		wait(2)
		if err := valcli.DumpEventLog(hash); err != nil {
			log.Error(err)
			return
		}
	}

	balanceAfterMint, err := nftBalance(params.Asset, owner)
	if err != nil {
		log.Error(err)
		return
	}

	subAmt := utils.SafeSub(balanceAfterMint, balanceBeforeMint).Uint64()
	if subAmt == 0 {
		log.Errorf("balance before mint %d, balance after mint %d, sub amount should be %d",
			balanceBeforeMint.Uint64(), balanceAfterMint.Uint64(), subAmt)
	}
	return true
}

func NFTBurn() (succeed bool) {
	var params struct {
		Asset   common.Address
		TokenID uint64
		Uri     string
	}

	if err := config.LoadParams("NFT-Mint.json", &params); err != nil {
		log.Error(err)
		return
	}

	token := new(big.Int).SetUint64(params.TokenID)
	hash, err := valcli.NFTBurn(params.Asset, token)
	if err != nil {
		log.Error(err)
		return
	}

	wait(2)
	if err := valcli.DumpEventLog(hash); err != nil {
		log.Error(err)
		return
	}

	return true
}

func NFTTransfer() (succeed bool) {
	var params struct {
		Asset          common.Address
		TokenID        uint64
		ToAccountIndex int
	}

	if err := config.LoadParams("NFT-Transfer.json", &params); err != nil {
		log.Error(err)
		return
	}

	// validator transfer to someone
	asset := params.Asset
	token := new(big.Int).SetUint64(params.TokenID)
	to := common.HexToAddress(config.Conf.Accounts[params.ToAccountIndex])
	{
		owner := valcli.Address()
		hash, err := valcli.NFTTransferFrom(asset, owner, to, token)
		if err != nil {
			log.Error(err)
			return
		}

		wait(2)
		if err := valcli.DumpEventLog(hash); err != nil {
			log.Error(err)
			return
		}
	}

	// transfer back to validator
	return nftTransferBack(asset, token, to)
}

func NFTLock() (succeed bool) {
	var params struct {
		Asset      common.Address
		TokenID    uint64
		Proxy common.Address
	}

	if err := config.LoadParams("NFT-Lock.json", &params); err != nil {
		log.Error(err)
		return
	}

	// 这个流程的大致轮廓是 palette nft -> palette nft proxy
	// 假设validator A depoly了一个nft合约，同时将token mint给了自己
	// safe transfer的时候
	asset := params.Asset
	crossChainID := uint64(config.Conf.Environment.NetworkID)
	from := valcli.Address()
	to := params.Proxy
	proxy := params.Proxy
	token := new(big.Int).SetUint64(params.TokenID)

	// lock
	{
		balanceBeforeLock, err := nftBalance(asset, from)
		if err != nil {
			log.Error(err)
			return
		}
		hash, err := valcli.NFTSafeTransferFrom(asset, from, proxy, token, to, crossChainID)
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
		for i := 0; i < 100; i++ {
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

func nftTransferBack(asset common.Address, tokenID *big.Int, from common.Address) (succeed bool) {
	url := valcli.Url()
	acc := config.LoadAccount(from.Hex())
	cli := sdk.NewSender(url, acc)
	to := valcli.Address()

	hash, err := cli.NFTTransferFrom(asset, from, to, tokenID)
	if err != nil {
		log.Error(err)
		return
	}

	wait(2)

	if err := cli.DumpEventLog(hash); err != nil {
		log.Error(err)
		return
	}

	return true
}

func nftBalance(asset, user common.Address) (*big.Int, error) {
	return valcli.NFTBalance(asset, user, "latest")
}
