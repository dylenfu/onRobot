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

const (
	shReset         = "reset.sh"
	shStartNodes    = "start_nodes.sh"
	shStopNodes     = "stop_nodes.sh"
	shStopAllNodes  = "stop_all_nodes.sh"
	shClearNodes    = "clear_nodes.sh"
	shClearAllNodes = "clear_all_nodes.sh"
)

func Demo() bool {
	log.Info("Hello, Palette chain")
	return true
}

// 清理数据，重启节点，并给每个账户一定的初始PLT(100000)
func ResetNetwork() bool {
	StopNetwork()
	ClearNetwork()
	shell.Exec(shReset)
	wait(1)

	var params struct {
		RpcUrl         string
		UserInitAmount int
	}

	if err := config.LoadParams("Reset.json", &params); err != nil {
		log.Error(err)
		return false
	}

	client = sdk.NewSender(params.RpcUrl, config.AdminKey)
	for _, account := range config.Conf.Accounts {
		amount := plt.TestMultiPLT(params.UserInitAmount)
		to := common.HexToAddress(account)
		if _, err := client.PLTTransfer(to, amount); err != nil {
			log.Errorf("transfer to %s err %v", to.Hex(), err)
			return false
		}
	}

	wait(1)

	for _, account := range config.Conf.Accounts {
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

// start genesis nodes
func StartNetwork() bool {
	shell.Exec(shStartNodes)
	return true
}

// stop all nodes
func StopNetwork() bool {
	shell.Exec(shStopAllNodes)
	return true
}

// clear all nodes
func ClearNetwork() bool {
	shell.Exec(shClearAllNodes)
	return true
}

// --------------------------------
// validators management
// --------------------------------

type ValidatorsConfig struct {
	RpcUrl               string
	ValidatorsIndexStart int
	ValidatorsIndexEnd   int
	ValidatorsNumber     int
	ValidatorInitAmount  int
}

func loadValidatorsConfig() *ValidatorsConfig {
	params := new(ValidatorsConfig)
	if err := config.LoadParams("InitValidators.json", params); err != nil {
		log.Fatalf("failed to load `startValidators` params, [%v]", err)
	}
	params.ValidatorsIndexEnd = params.ValidatorsIndexStart + params.ValidatorsNumber - 1
	return params
}

func InitValidators() (succeed bool) {
	params := loadValidatorsConfig()
	config.Conf.ResetEnv(params.ValidatorsIndexStart, params.ValidatorsNumber)
	shell.Exec(shStopNodes)
	shell.Exec(shClearNodes)
	shell.Exec(shReset)

	wait(1)

	client = sdk.NewSender(params.RpcUrl, config.AdminKey)

	amount := plt.TestMultiPLT(params.ValidatorInitAmount)
	for i := params.ValidatorsIndexStart; i <= params.ValidatorsIndexEnd; i++ {
		to := config.Conf.Nodes[i].Addr()
		if _, err := client.PLTTransfer(to, amount); err != nil {
			log.Errorf("transfer to %s err %v", to.Hex(), err)
			return false
		}
	}

	wait(1)

	for i := params.ValidatorsIndexStart; i <= params.ValidatorsIndexEnd; i++ {
		owner := config.Conf.Nodes[i].Addr()
		data, err := client.BalanceOf(owner, "latest")
		if err != nil {
			log.Errorf("query balanceOf %s err %v", owner.Hex(), err)
			return false
		}

		log.Infof("%s init balance %d", owner.Hex(), utils.UnsafeDiv(data, plt.OnePLT))
	}

	return true
}

func StartValidators() bool {
	params := loadValidatorsConfig()
	config.Conf.ResetEnv(params.ValidatorsIndexStart, params.ValidatorsNumber)
	shell.Exec(shStartNodes)
	return true
}

func StopValidators() bool {
	params := loadValidatorsConfig()
	config.Conf.ResetEnv(params.ValidatorsIndexStart, params.ValidatorsNumber)
	shell.Exec(shStopNodes)
	return true
}

func ClearValidators() bool {
	params := loadValidatorsConfig()
	config.Conf.ResetEnv(params.ValidatorsIndexStart, params.ValidatorsNumber)
	shell.Exec(shClearNodes)
	return true
}

func gc() {
	client = nil
	config.Conf = config.BakConf.DeepCopy()
}

func wait(nBlock int) {
	time.Sleep(time.Duration(config.Conf.BlockPeriod) * time.Duration(nBlock))
}
