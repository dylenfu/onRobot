package core

import (
	"math/big"

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

func PLTTransfer() bool {
	var params struct {
		RpcUrl string
		From   string
		To     string
		Amount int64
	}

	if err := loadParams("PLTTransfer.json", &params); err != nil {
		log.Error(err)
		return false
	}

	key := loadAccount(params.From)
	client := sdk.NewSender(params.RpcUrl, key)
	to := common.HexToAddress(params.To)
	amount := utils.SafeMul(big.NewInt(params.Amount), plt.OnePLT)

	// balance before transfer
	fromBalanceBeforeTrans, err := client.BalanceOf(key.Address, "latest")
	if err != nil {
		log.Error(err)
		return false
	}
	toBalanceBeforeTrans, err := client.BalanceOf(to, "latest")
	if err != nil {
		log.Error(err)
		return false
	}
	if fromBalanceBeforeTrans.Cmp(amount) < 0 {
		log.Errorf("%s balance not enough %d", params.From, utils.UnsafeDiv(fromBalanceBeforeTrans, plt.OnePLT))
		return false
	}

	// transfer and waiting for commit
	hash, err := client.PLTTransfer(to, amount)
	if err != nil {
		log.Error(err)
		return false
	}
	wait(1)
	if err := client.DumpEventLog(hash); err != nil {
		log.Error(err)
		return false
	}

	// balance after transfer
	fromBalanceAfterTrans, err := client.BalanceOf(key.Address, "latest")
	if err != nil {
		log.Error(err)
		return false
	}
	toBalanceAfterTrans, err := client.BalanceOf(to, "latest")
	if err != nil {
		log.Error(err)
		return false
	}

	// expect sum
	if utils.SafeAdd(toBalanceBeforeTrans, amount).Cmp(toBalanceAfterTrans) != 0 {
		log.Errorf("dst balance before transfer %d, balance after transfer %d, amount %d",
			utils.UnsafeDiv(toBalanceBeforeTrans, plt.OnePLT),
			utils.UnsafeDiv(toBalanceAfterTrans, plt.OnePLT),
			params.Amount,
		)
		return false
	}
	if utils.SafeSub(fromBalanceBeforeTrans, amount).Cmp(fromBalanceAfterTrans) != 0 {
		log.Errorf("src balance before transfer %d, balance after transfer %d, amount %d",
			utils.UnsafeDiv(fromBalanceAfterTrans, plt.OnePLT),
			utils.UnsafeDiv(fromBalanceAfterTrans, plt.OnePLT),
			params.Amount,
		)
		return false
	}

	return true
}

func PLTApprove() bool {
	var params struct {
		RpcUrl string
		Owner string
		Spender string
		Amount int
		BlockNum string
	}

	if err := loadParams("PLTApprove", &params); err != nil {
		log.Error(err)
		return false
	}

	key := loadAccount(params.Owner)
	client := sdk.NewSender(params.RpcUrl, key)

	owner := key.Address
	spender := common.HexToAddress(params.Spender)
	amount := plt.TestMultiPLT(params.Amount)

	// allowance before approve
	allowanceBeforeApprove, err := client.PLTAllowance(owner, spender, params.BlockNum)
	if err != nil {
		log.Error(err)
		return false
	}

	hash, err := client.PLTApprove(spender, amount)
	if err != nil {
		log.Error(err)
		return false
	}
	wait(1)
	if err := client.DumpEventLog(hash); err != nil {
		log.Error(err)
		return false
	}

	// allowance after approve
	allowanceAfterApprove, err := client.PLTAllowance(owner, spender, params.BlockNum)
	if err != nil {
		log.Error(err)
		return false
	}

	if allowanceAfterApprove.Cmp(utils.SafeAdd(allowanceBeforeApprove, amount)) != 0 {
		log.Errorf("owner %s, spender %s, allowance before approve %d, allowance after approve %d, amount %d",
			owner.Hex(), spender.Hex(),
			utils.UnsafeDiv(allowanceBeforeApprove, plt.OnePLT),
			utils.UnsafeDiv(allowanceAfterApprove, plt.OnePLT),
			utils.UnsafeDiv(amount, plt.OnePLT))
	}

	return true
}