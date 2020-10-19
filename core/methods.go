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

func Demo() bool {
	log.Info("Hello, Palette chain")
	return true
}

// 清理数据，重启节点，并给每个账户一定的初始PLT(100000)
func ResetNetwork() bool {
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

	wait(2)

	client := sdk.NewSender(params.RpcUrl, config.AdminKey)
	amount := plt.TestMultiPLT(params.InitAmount)
	for _, account := range config.Conf.Accounts {
		to := common.HexToAddress(account)
		hash, err := client.PLTTransfer(to, amount)
		if err != nil {
			log.Error("transfer to %s err %v", to.Hex(), err)
			return false
		}
		// for nonce increasing
		wait(1)
		if err := client.DumpEventLog(hash); err != nil {
			log.Error("dump %s event log err %v", hash.Hex(), err)
			return false
		}
	}

	wait(1)

	for _, account := range config.Conf.Accounts {
		owner := common.HexToAddress(account)
		data, err := client.BalanceOf(owner, "latest")
		if err != nil {
			log.Error("query balanceOf %s err %v", owner.Hex(), err)
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

func gc() {
}

func wait(nBlock int) {
	time.Sleep(time.Duration(config.Conf.BlockPeriod) * time.Duration(nBlock))
}
