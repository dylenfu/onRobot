package internal

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/contracts/native"
	"github.com/ethereum/go-ethereum/contracts/native/plt"
	"github.com/ethereum/go-ethereum/contracts/native/utils"
	"github.com/palettechain/onRobot/pkg/log"
	"github.com/palettechain/onRobot/pkg/sdk"
)

func PLTTotalSupply() bool {
	var params struct {
		RpcUrl string
		Expect uint64
	}

	if err := loadParams("PLTTotalSupply.json", &params); err != nil {
		log.Error(err)
		return false
	}

	client := sdk.NewSender(params.RpcUrl, adminKey)
	totalSupply, err := client.PLTTotalSupply("latest")
	if err != nil {
		log.Error(err)
		return false
	}

	actual := utils.UnsafeDiv(totalSupply, plt.OnePLT)
	if actual != params.Expect {
		log.Error("totalSupply expect %d actually %d", params.Expect, actual)
		return false
	}

	log.Infof("totalSupply %d", utils.UnsafeDiv(totalSupply, plt.OnePLT))

	return true
}

func PLTDecimal() bool {
	var params struct {
		RpcUrl string
		Expect uint64
	}

	if err := loadParams("PLTDecimal.json", &params); err != nil {
		log.Error(err)
		return false
	}

	client := sdk.NewSender(params.RpcUrl, adminKey)
	actual, err := client.PLTDecimals()
	if err != nil {
		log.Error(err)
		return false
	}

	if params.Expect != actual {
		log.Error("decimal expect %d actual %d", params.Expect, actual)
		return false
	}

	log.Infof("decimal %d", actual)

	return true
}

func AdminBalance() bool {
	var params struct {
		RpcUrl   string
		BlockNum string
		Expect   uint64
	}

	if err := loadParams("AdminBalance.json", &params); err != nil {
		log.Error(err)
		return false
	}
	client := sdk.NewSender(params.RpcUrl, adminKey)

	balance, err := client.BalanceOf(adminKey.Address, params.BlockNum)
	if err != nil {
		log.Error(err)
		return false
	}

	actual := utils.UnsafeDiv(balance, plt.OnePLT)
	if actual != params.Expect {
		log.Error("balance expect %d actually %d", params.Expect, actual)
		return false
	}

	log.Infof("balance %d")

	return true
}

func GovernanceBalance() bool {
	var params struct {
		RpcUrl   string
		BlockNum string
		Expect   uint64
	}

	if err := loadParams("GovernanceBalance.json", &params); err != nil {
		log.Error(err)
		return false
	}

	client := sdk.NewSender(params.RpcUrl, adminKey)
	owner := common.HexToAddress(native.GovernanceContractAddress)
	balance, err := client.BalanceOf(owner, params.BlockNum)
	if err != nil {
		log.Error(err)
		return false
	}

	actual := utils.UnsafeDiv(balance, plt.OnePLT)
	if actual != params.Expect {
		log.Error("balance expect %d actually %d", params.Expect, actual)
		return false
	}

	log.Infof("balance %d", utils.UnsafeDiv(balance, plt.OnePLT))

	return true
}

func PLTBalanceOf() bool {
	var params struct {
		RpcUrl   string
		Owner    string
		BlockNum string
	}

	if err := loadParams("PLTBalanceOf.json", &params); err != nil {
		log.Error(err)
		return false
	}

	client := sdk.NewSender(params.RpcUrl, adminKey)
	owner := common.HexToAddress(params.Owner)
	balance, err := client.BalanceOf(owner, params.BlockNum)
	if err != nil {
		log.Error(err)
		return false
	}

	log.Infof("balance %d", utils.UnsafeDiv(balance, plt.OnePLT))

	return true
}
