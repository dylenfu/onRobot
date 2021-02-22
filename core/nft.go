package core

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/contracts/native/utils"
	"github.com/palettechain/onRobot/config"
	"github.com/palettechain/onRobot/pkg/log"
	"github.com/palettechain/onRobot/pkg/sdk"
	"strings"

	//polycm "github.com/polynetwork/poly/common"
	"math/big"
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
	balanceBeforeMint, err := valcli.NFTBalance(params.Asset, owner, "latest")
	if err != nil {
		log.Error(err)
		return
	}

	// mint
	if _, err := valcli.NFTMint(params.Asset, owner, token, params.Uri); err != nil {
		log.Error(err)
		return
	}

	balanceAfterMint, err := valcli.NFTBalance(params.Asset, owner, "latest")
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
	if _, err := valcli.NFTBurn(params.Asset, token); err != nil {
		log.Error(err)
		return
	}

	return true
}

func NFTSetUri() (succeed bool) {
	var params struct {
		List    []common.Address
		Storage string
	}

	if err := config.LoadParams("SetAssetUri.json", &params); err != nil {
		log.Error(err)
		return
	}
	if !strings.HasSuffix(params.Storage, "/") {
		params.Storage += "/"
	}

	getSuffix := func(src common.Address) string {
		num := new(big.Int).SetBytes(src.Bytes()).Uint64()
		return fmt.Sprintf("%x/", num)
	}

	rpc := admcli.Url()
	for _, asset := range params.List {
		owner, _ := admcli.NFTAssetOwner(asset, "latest")
		if owner == utils.EmptyAddress {
			continue
		}

		cli := sdk.NewSender(rpc, customLoadAccount(owner.Hex()))

		suffix := getSuffix(asset)
		uri := params.Storage + suffix
		hash, err := cli.NFTSetBaseUri(asset, uri)
		if err != nil {
			log.Error(err)
			return
		}
		uri, _ = admcli.NFTGetBaseUri(asset, "latest")
		log.Infof("hash %s, asset: %s, owner: %s, uri: %s", hash.Hex(), asset.Hex(), owner.Hex(), uri)

		for i := 0; i < 10; i++ {
			token := big.NewInt(int64(i))
			if uri, _ := admcli.NFTTokenURI(asset, token, "latest"); uri != "" {
				log.Infof("token uri %s", uri)
			}
		}
	}
	return true
}

func NFTTransfer() (succeed bool) {
	var params struct {
		Asset   common.Address
		TokenID uint64
		To      common.Address
	}

	if err := config.LoadParams("NFT-Transfer.json", &params); err != nil {
		log.Error(err)
		return
	}

	// validator transfer to someone
	asset := params.Asset
	token := new(big.Int).SetUint64(params.TokenID)
	owner := valcli.Address()
	if _, err := valcli.NFTTransferFrom(asset, owner, params.To, token); err != nil {
		log.Error(err)
		return
	}

	//return true
	// transfer back to validator
	return nftTransferBack(asset, token, params.To)
}

func NFTBalance() (succeed bool) {
	var params struct {
		Asset common.Address
		User  common.Address
	}

	if err := config.LoadParams("NFT-Balance.json", &params); err != nil {
		log.Error(err)
		return
	}

	num, err := valcli.NFTBalance(params.Asset, params.User, "latest")
	if err != nil {
		log.Error(err)
		return
	}
	log.Infof("%s balance %d", params.User.Hex(), num.Uint64())
	return true
}

func NFTTokenOwner() (succeed bool) {
	var params struct {
		Asset   common.Address
		TokenID uint64
		Proxy   common.Address
	}

	if err := config.LoadParams("NFT-Lock.json", &params); err != nil {
		log.Error(err)
		return
	}

	tokenID := new(big.Int).SetUint64(params.TokenID)
	owner, err := valcli.NFTTokenOwner(params.Asset, tokenID, "latest")
	if err != nil {
		log.Error(err)
		return
	}

	log.Infof("asset %s, token %d, owner %s", params.Asset.Hex(), tokenID.Uint64(), owner.Hex())
	return true
}

func nftTransferBack(asset common.Address, tokenID *big.Int, from common.Address) (succeed bool) {
	url := valcli.Url()
	cli := sdk.NewSender(url, customLoadAccount(from.Hex()))
	to := valcli.Address()

	if _, err := cli.NFTTransferFrom(asset, from, to, tokenID); err != nil {
		log.Error(err)
		return
	}

	return true
}

func nftBalance(asset, user common.Address) (*big.Int, error) {
	return valcli.NFTBalance(asset, user, "latest")
}
