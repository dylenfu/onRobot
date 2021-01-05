package core

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/contracts/native/plt"
	"github.com/ethereum/go-ethereum/contracts/native/utils"
	"github.com/palettechain/onRobot/config"
	"github.com/palettechain/onRobot/pkg/log"
	"github.com/palettechain/onRobot/pkg/sdk"
)

// 在palette native plt合约mint一定量的PLT token到某个已经存在的用户地址
func PLTMint() (succeed bool) {
	var p struct {
		To    common.Address
		Value int
	}
	if err := config.LoadParams("Mint.json", &p); err != nil {
		log.Error(err)
		return
	}
	amount := plt.MultiPLT(p.Value)

	toBalanceBeforeTrans, err := admcli.BalanceOf(p.To, "latest")
	if err != nil {
		log.Error(err)
		return
	}
	tx, err := admcli.PLTMint(p.To, amount)
	if err != nil {
		log.Error(err)
		return
	}
	wait(2)
	if err := admcli.DumpEventLog(tx); err != nil {
		log.Error(err)
		return
	}

	toBalanceAfterTrans, err := admcli.BalanceOf(p.To, "latest")
	if err != nil {
		log.Error(err)
		return
	}

	if utils.SafeSub(toBalanceAfterTrans, toBalanceBeforeTrans).Cmp(amount) != 0 {
		log.Errorf("wrong mint value %s and should be %s",
			utils.SafeSub(toBalanceAfterTrans, toBalanceBeforeTrans), p.Value)
	}
	return true
}

// 在palette native plt合约烧毁合约PLT总供应量对应的amount
func PLTBurn() (succeed bool) {
	var p struct {
		Value int
	}
	if err := config.LoadParams("Burn.json", &p); err != nil {
		log.Error(err)
		return
	}
	amount := plt.MultiPLT(p.Value)

	toBalanceBeforeTrans, err := admcli.BalanceOf(admcli.Address(), "latest")
	if err != nil {
		log.Error(err)
		return
	}
	if toBalanceBeforeTrans.Cmp(amount) == -1 {
		log.Error(err)
		return
	}

	tx, err := admcli.PLTBurn(amount)
	if err != nil {
		log.Error(err)
		return
	}
	wait(2)
	if err := admcli.DumpEventLog(tx); err != nil {
		log.Error(err)
		return
	}

	toBalanceAfterTrans, err := admcli.BalanceOf(admcli.Address(), "latest")
	if err != nil {
		log.Error(err)
		return
	}

	if utils.SafeSub(toBalanceBeforeTrans, toBalanceAfterTrans).Cmp(amount) != 0 {
		log.Errorf("wrong mint value %s and should be %s",
			utils.SafeSub(toBalanceAfterTrans, toBalanceBeforeTrans), p.Value)
	}

	return true
}

func PLTLock() (succeed bool) {
	var params struct {
		AccountIndex int
		Amount       int
	}
	sideChainID := uint64(config.Conf.CrossChain.SideChainID)

	if err := config.LoadParams("Lock.json", &params); err != nil {
		log.Error(err)
		return
	}

	if params.AccountIndex > len(config.Conf.Accounts)-1 {
		log.Errorf("account index out of range")
		return
	}

	user := config.Conf.Accounts[params.AccountIndex]
	privKey := config.LoadAccount(user)
	baseUrl := config.Conf.Nodes[0].RPCAddr()
	userAddr := common.HexToAddress(user)
	bindTo := common.HexToAddress(user) // lock to self
	cli := sdk.NewSender(baseUrl, privKey)
	amount := plt.MultiPLT(params.Amount)

	// prepare balance
	{
		logsplit()
		log.Infof("prepare test account balance...")
		hash, err := admcli.PLTTransfer(userAddr, amount)
		if err != nil {
			log.Error(err)
			return
		}
		wait(2)
		if err := admcli.DumpEventLog(hash); err != nil {
			log.Error(err)
			return
		}
	}

	// lock plt
	{
		logsplit()
		log.Infof("lock PLT...")
		balanceBeforeLock, err := cli.BalanceOf(userAddr, "latest")
		if err != nil {
			log.Error(err)
			return
		}

		hash, err := cli.Lock(sideChainID, bindTo, amount)
		if err != nil {
			log.Errorf("failed to call `lock` err: %v", err)
			return
		}
		wait(2)
		if err := cli.DumpEventLog(hash); err != nil {
			log.Error("failed to dump `lock` event hash %s, err: %v", hash.Hex(), err)
			return
		}

		balanceAfterLock, err := cli.BalanceOf(userAddr, "latest")
		if err != nil {
			log.Error(err)
			return
		}

		subAmount := utils.SafeSub(balanceBeforeLock, balanceAfterLock)
		if subAmount.Cmp(amount) != 0 {
			log.Errorf("balance before lock %d, after lock %d, the sub amount should be %d",
				plt.PrintUPLT(balanceBeforeLock), plt.PrintUPLT(balanceAfterLock), plt.PrintUPLT(amount))
			return
		} else {
			log.Infof("balance before lock %d, after lock %d, the sub amount is %d",
				plt.PrintUPLT(balanceBeforeLock), plt.PrintUPLT(balanceAfterLock), plt.PrintUPLT(subAmount))
		}
	}

	// waiting for unlock
	{
		logsplit()
		log.Infof("unlock PLT...")

		var (
			balanceBeforeUnlock,
			balanceAfterUnlock *big.Int
		)
		for i := 0; i < 100; i++ {
			balance, err := cli.BalanceOf(bindTo, "latest")
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
					plt.PrintUPLT(balanceBeforeUnlock), plt.PrintUPLT(balanceAfterUnlock), plt.PrintUPLT(subAmount))
				break
			}
			time.Sleep(3 * time.Second)
		}
	}

	// return plt
	{
		hash, err := cli.PLTTransfer(common.HexToAddress(config.Conf.AdminAccount), amount)
		if err != nil {
			log.Infof("transfer back PLT to admin err: %s", err)
			return true
		}
		_ = cli.DumpEventLog(hash)
		balance, err := cli.BalanceOf(userAddr, "latest")
		if err != nil {
			log.Infof("check balance after unlock err: %s", err)
		} else {
			log.Infof("balance after unlock %d", plt.PrintUPLT(balance))
		}
	}

	return true
}
