package core

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/contracts/native"
	"github.com/ethereum/go-ethereum/contracts/native/plt"
	"github.com/ethereum/go-ethereum/contracts/native/utils"
	"github.com/palettechain/onRobot/config"
	"github.com/palettechain/onRobot/pkg/log"
	"github.com/palettechain/onRobot/pkg/sdk"
)

func TotalSupply() (succeed bool) {
	var params struct {
		Expect uint64
	}

	if err := config.LoadParams("TotalSupply.json", &params); err != nil {
		log.Error(err)
		return
	}

	client = sdk.NewSender(config.Conf.BaseRPCUrl, config.AdminKey)
	totalSupply, err := client.PLTTotalSupply("latest")
	if err != nil {
		log.Error(err)
		return
	}

	actual := utils.UnsafeDiv(totalSupply, plt.OnePLT)
	if actual != params.Expect {
		log.Errorf("totalSupply expect %d actually %d", params.Expect, actual)
		return
	}

	log.Infof("totalSupply %d", utils.UnsafeDiv(totalSupply, plt.OnePLT))

	return true
}

func Decimal() (succeed bool) {
	var params struct {
		Expect uint64
	}

	if err := config.LoadParams("Decimal.json", &params); err != nil {
		log.Error(err)
		return
	}

	client = sdk.NewSender(config.Conf.BaseRPCUrl, config.AdminKey)
	actual, err := client.PLTDecimals()
	if err != nil {
		log.Error(err)
		return
	}

	if params.Expect != actual {
		log.Errorf("decimal expect %d actual %d", params.Expect, actual)
		return
	}

	log.Infof("decimal %d", actual)

	return true
}

func AdminBalance() (succeed bool) {
	var params struct {
		BlockNum string
		Expect   uint64
	}

	if err := config.LoadParams("AdminBalance.json", &params); err != nil {
		log.Error(err)
		return
	}
	client = sdk.NewSender(config.Conf.BaseRPCUrl, config.AdminKey)

	balance, err := client.BalanceOf(config.AdminKey.Address, params.BlockNum)
	if err != nil {
		log.Error(err)
		return
	}

	actual := utils.UnsafeDiv(balance, plt.OnePLT)
	if actual != params.Expect {
		log.Errorf("balance expect %d actually %d", params.Expect, actual)
		return
	}

	log.Infof("balance %d")

	return true
}

func GovernanceBalance() (succeed bool) {
	var params struct {
		BlockNum string
		Expect   uint64
	}

	if err := config.LoadParams("GovernanceBalance.json", &params); err != nil {
		log.Error(err)
		return
	}

	client = sdk.NewSender(config.Conf.BaseRPCUrl, config.AdminKey)
	owner := common.HexToAddress(native.GovernanceContractAddress)
	balance, err := client.BalanceOf(owner, params.BlockNum)
	if err != nil {
		log.Error(err)
		return
	}

	actual := utils.UnsafeDiv(balance, plt.OnePLT)
	if actual != params.Expect {
		log.Errorf("balance expect %d actually %d", params.Expect, actual)
		return
	}

	log.Infof("balance %d", utils.UnsafeDiv(balance, plt.OnePLT))

	return true
}

func BalanceOf() (succeed bool) {
	var params struct {
		Owner    string
		BlockNum string
	}

	if err := config.LoadParams("BalanceOf.json", &params); err != nil {
		log.Error(err)
		return
	}

	client = sdk.NewSender(config.Conf.BaseRPCUrl, config.AdminKey)
	owner := common.HexToAddress(params.Owner)
	balance, err := client.BalanceOf(owner, params.BlockNum)
	if err != nil {
		log.Error(err)
		return
	}

	log.Infof("balance %d", utils.UnsafeDiv(balance, plt.OnePLT))

	return true
}

func Transfer() (succeed bool) {
	var params struct {
		From   string
		To     string
		Amount int64
	}

	if err := config.LoadParams("Transfer.json", &params); err != nil {
		log.Error(err)
		return
	}

	key := config.LoadAccount(params.From)
	client = sdk.NewSender(config.Conf.BaseRPCUrl, key)
	to := common.HexToAddress(params.To)
	amount := utils.SafeMul(big.NewInt(params.Amount), plt.OnePLT)

	// balance before transfer
	fromBalanceBeforeTrans, err := client.BalanceOf(key.Address, "latest")
	if err != nil {
		log.Error(err)
		return
	}
	toBalanceBeforeTrans, err := client.BalanceOf(to, "latest")
	if err != nil {
		log.Error(err)
		return
	}
	if fromBalanceBeforeTrans.Cmp(amount) < 0 {
		log.Errorf("%s balance not enough %d", params.From, utils.UnsafeDiv(fromBalanceBeforeTrans, plt.OnePLT))
		return
	}

	// transfer and waiting for commit
	hash, err := client.PLTTransfer(to, amount)
	if err != nil {
		log.Error(err)
		return
	}
	wait(1)
	if err := client.DumpEventLog(hash); err != nil {
		log.Error(err)
		return
	}

	// balance after transfer
	fromBalanceAfterTrans, err := client.BalanceOf(key.Address, "latest")
	if err != nil {
		log.Error(err)
		return
	}
	toBalanceAfterTrans, err := client.BalanceOf(to, "latest")
	if err != nil {
		log.Error(err)
		return
	}

	// expect sum
	if utils.SafeAdd(toBalanceBeforeTrans, amount).Cmp(toBalanceAfterTrans) != 0 {
		log.Errorf("dst balance before transfer %d, balance after transfer %d, amount %d",
			utils.UnsafeDiv(toBalanceBeforeTrans, plt.OnePLT),
			utils.UnsafeDiv(toBalanceAfterTrans, plt.OnePLT),
			params.Amount,
		)
		return
	}
	if utils.SafeSub(fromBalanceBeforeTrans, amount).Cmp(fromBalanceAfterTrans) != 0 {
		log.Errorf("src balance before transfer %d, balance after transfer %d, amount %d",
			utils.UnsafeDiv(fromBalanceAfterTrans, plt.OnePLT),
			utils.UnsafeDiv(fromBalanceAfterTrans, plt.OnePLT),
			params.Amount,
		)
		return
	}

	return true
}

func Approve() (succeed bool) {
	var params struct {
		Owner   string
		Spender string
		Amount  int
	}

	if err := config.LoadParams("Approve.json", &params); err != nil {
		log.Error(err)
		return
	}

	key := config.LoadAccount(params.Owner)
	client = sdk.NewSender(config.Conf.BaseRPCUrl, key)

	owner := key.Address
	spender := common.HexToAddress(params.Spender)
	amount := plt.TestMultiPLT(params.Amount)

	// allowance before approve
	allowanceBeforeApprove, err := client.PLTAllowance(owner, spender, "latest")
	if err != nil {
		log.Error(err)
		return
	}

	hash, err := client.PLTApprove(spender, amount)
	if err != nil {
		log.Error(err)
		return
	}
	wait(1)
	if err := client.DumpEventLog(hash); err != nil {
		log.Error(err)
		return
	}

	// allowance after approve
	allowanceAfterApprove, err := client.PLTAllowance(owner, spender, "latest")
	if err != nil {
		log.Error(err)
		return
	}

	if allowanceAfterApprove.Cmp(amount) != 0 {
		log.Errorf("owner %s, spender %s, allowance before approve %d, allowance after approve %d, amount %d",
			owner.Hex(), spender.Hex(),
			utils.UnsafeDiv(allowanceBeforeApprove, plt.OnePLT),
			utils.UnsafeDiv(allowanceAfterApprove, plt.OnePLT),
			utils.UnsafeDiv(amount, plt.OnePLT))
		return
	}

	return true
}
