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
	cli := getPaletteCli(pltCTypeCustomer)
	totalSupply, err := cli.PLTTotalSupply("latest")
	if err != nil {
		log.Error(err)
		return
	}

	actual := plt.PrintFPLT(utils.DecimalFromBigInt(totalSupply))
	log.Infof("totalSupply %f", actual)
	return true
}

func Decimal() (succeed bool) {
	cli := getPaletteCli(pltCTypeCustomer)
	data, err := cli.PLTDecimals()
	if err != nil {
		log.Error(err)
		return
	}

	log.Infof("decimal %d", data)

	return true
}

func Name() (succeed bool) {
	cli := getPaletteCli(pltCTypeCustomer)
	actual, err := cli.PLTName()
	if err != nil {
		log.Error(err)
		return
	}

	expect := "Palette Token"
	if actual != expect {
		log.Errorf("contract name expect %s actual %s", expect, actual)
		return
	}

	log.Infof("contract name %s", actual)

	return true
}

func AdminBalance() (succeed bool) {
	addr := config.Conf.AdminAccount
	cli := getPaletteCli(pltCTypeCustomer)
	balance, err := cli.BalanceOf(addr, "latest")
	if err != nil {
		log.Error(err)
		return
	}

	actual := plt.PrintUPLT(balance)
	log.Infof("admin %s balance %d", addr.Hex(), actual)

	return true
}

func GovernanceBalance() (succeed bool) {
	owner := common.HexToAddress(native.GovernanceContractAddress)
	cli := getPaletteCli(pltCTypeCustomer)
	balance, err := cli.BalanceOf(owner, "latest")
	if err != nil {
		log.Error(err)
		return
	}

	actual := plt.PrintUPLT(balance)
	log.Infof("governance %s balance %d", owner.Hex(), actual)

	return true
}

func BalanceOf() (succeed bool) {
	var params struct {
		Owner    common.Address
		BlockNum string
	}

	if err := config.LoadParams("PLT-Balance.json", &params); err != nil {
		log.Error(err)
		return
	}

	owner := params.Owner
	cli := getPaletteCli(pltCTypeCustomer)
	balance, err := cli.BalanceOf(owner, params.BlockNum)
	if err != nil {
		log.Error(err)
		return
	}

	log.Infof("balance %s", balance.String())
	return true
}

func Transfer() (succeed bool) {
	var params struct {
		From   common.Address
		To     common.Address
		Amount int64
	}

	if err := config.LoadParams("PLT-Transfer.json", &params); err != nil {
		log.Error(err)
		return
	}

	key := customLoadAccount(params.From)
	to := params.To
	amount := utils.SafeMul(big.NewInt(params.Amount), plt.OnePLT)
	admcli := getPaletteCli(pltCTypeAdmin)

	// balance before transfer
	fromBalanceBeforeTrans, err := admcli.BalanceOf(PrivKey2Addr(key), "latest")
	if err != nil {
		log.Error(err)
		return
	}
	toBalanceBeforeTrans, err := admcli.BalanceOf(to, "latest")
	if err != nil {
		log.Error(err)
		return
	}
	if fromBalanceBeforeTrans.Cmp(amount) < 0 {
		log.Errorf("%s balance not enough %d", params.From, utils.UnsafeDiv(fromBalanceBeforeTrans, plt.OnePLT))
		return
	}

	// transfer and waiting for commit
	if _, err := admcli.PLTTransfer(to, amount); err != nil {
		log.Error(err)
		return
	}

	// balance after transfer
	fromBalanceAfterTrans, err := admcli.BalanceOf(PrivKey2Addr(key), "latest")
	if err != nil {
		log.Error(err)
		return
	}
	toBalanceAfterTrans, err := admcli.BalanceOf(to, "latest")
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
		Owner   common.Address
		Spender common.Address
		Amount  int
	}

	if err := config.LoadParams("PLT-Approve.json", &params); err != nil {
		log.Error(err)
		return
	}

	baseUrl := config.Conf.Rpc
	key := customLoadAccount(params.Owner)
	cli := sdk.NewSender(baseUrl, key)

	owner := PrivKey2Addr(key)
	spender := params.Spender
	amount := plt.MultiPLT(params.Amount)

	// allowance before approve
	allowanceBeforeApprove, err := cli.PLTAllowance(owner, spender, "latest")
	if err != nil {
		log.Error(err)
		return
	}

	if _, err := cli.PLTApprove(spender, amount); err != nil {
		log.Error(err)
		return
	}

	// allowance after approve
	allowanceAfterApprove, err := cli.PLTAllowance(owner, spender, "latest")
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
