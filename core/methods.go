package core

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/contracts/native/plt"
	"github.com/ethereum/go-ethereum/contracts/native/utils"
	"github.com/palettechain/onRobot/config"
	"github.com/palettechain/onRobot/pkg/log"
	"github.com/palettechain/onRobot/pkg/sdk"
	"github.com/palettechain/onRobot/pkg/shell"
)

var client *sdk.Client

func Demo() bool {
	log.Info("Hello, Palette chain")
	return true
}

// 清理数据，重启节点，并给每个账户一定的初始PLT(100000)
func ResetNetwork() bool {
	StopNetwork()
	ClearNetwork()

	var params struct {
		RpcUrl     string
		ShellPath  string
		InitAmount int
	}

	if err := config.LoadParams("Reset.json", &params); err != nil {
		log.Error(err)
		return false
	}

	shellPath := config.ShellPath(params.ShellPath)
	shell.Exec(shellPath)

	wait(1)

	client = sdk.NewSender(params.RpcUrl, config.AdminKey)
	amount := plt.TestMultiPLT(params.InitAmount)

	accounts := config.Conf.Accounts
	nodeAccounts := config.Conf.AllNodeAddressList()
	accounts = append(accounts, nodeAccounts...)
	for _, account := range accounts {
		to := common.HexToAddress(account)
		if _, err := client.PLTTransfer(to, amount); err != nil {
			log.Errorf("transfer to %s err %v", to.Hex(), err)
			return false
		}
	}

	wait(1)

	for _, account := range accounts {
		owner := common.HexToAddress(account)
		data, err := client.BalanceOf(owner, "latest")
		if err != nil {
			log.Errorf("query balanceOf %s err %v", owner.Hex(), err)
			return false
		}

		log.Infof("%s init balance %d", account, utils.UnsafeDiv(data, plt.OnePLT))
	}

	return true
}

func StartNetwork() bool {
	var params struct {
		ShellPath string
	}

	if err := config.LoadParams("Start.json", &params); err != nil {
		log.Error(err)
		return false
	}

	shellPath := config.ShellPath(params.ShellPath)
	shell.Exec(shellPath)
	return true
}

func StopNetwork() bool {
	var params struct {
		ShellPath string
	}

	if err := config.LoadParams("Stop.json", &params); err != nil {
		log.Error(err)
		return false
	}

	shellPath := config.ShellPath(params.ShellPath)
	shell.Exec(shellPath)
	return true
}

func ClearNetwork() bool {
	var params struct {
		ShellPath string
	}

	if err := config.LoadParams("Clear.json", &params); err != nil {
		log.Error(err)
		return false
	}

	shellPath := config.ShellPath(params.ShellPath)
	shell.Exec(shellPath)
	return true
}

func gc() {
	client = nil
}

func wait(nBlock int) {
	time.Sleep(time.Duration(config.Conf.BlockPeriod) * time.Duration(nBlock))
}
