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
	shStartSyncNode = "reset_sync_node.sh"
	shStopSyncNode  = "stop_sync_node.sh"
)

func Demo() bool {
	log.Info("Hello, Palette chain")
	return true
}

func BlockNumber() bool {
	client = sdk.NewSender(config.Conf.BaseRPCUrl, config.AdminKey)
	blockNumber := client.GetBlockNumber()
	log.Infof("current block number %d", blockNumber)
	return true
}

func Nonce() (succeed bool) {
	var params struct {
		Address string
	}
	if err := config.LoadParams("GetNonce.json", &params); err != nil {
		log.Error(err)
		return
	}

	client = sdk.NewSender(config.Conf.BaseRPCUrl, config.AdminKey)
	nonce := client.GetNonce(params.Address)
	log.Infof("%s nonce is %d", params.Address, nonce)
	return true
}

// --------------------------------
// genesis node management
// --------------------------------

// 清理数据，重启节点，并给每个账户一定的初始PLT(100000)
func ResetNetwork() bool {
	StopNetwork()
	ClearNetwork()
	shell.Exec(shReset)
	wait(1)

	var params struct {
		UserInitAmount int
	}

	if err := config.LoadParams("Reset.json", &params); err != nil {
		log.Error(err)
		return false
	}

	client = sdk.NewSender(config.Conf.BaseRPCUrl, config.AdminKey)
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
	NewNodeUrl           string
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

	client = sdk.NewSender(config.Conf.BaseRPCUrl, config.AdminKey)

	// transfer and check balance
	amount := plt.TestMultiPLT(params.ValidatorInitAmount)
	for i := params.ValidatorsIndexStart; i <= params.ValidatorsIndexEnd; i++ {
		to := config.Conf.Nodes[i].Addr()
		if hash, err := client.PLTTransfer(to, amount); err != nil {
			log.Errorf("url %s, transfer to node%d %s amount %d, hash %s, err %v", config.Conf.BaseRPCUrl, i, to.Hex(), params.ValidatorInitAmount, hash.Hex(), err)
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

	// sync blocks
	newNodeClient := sdk.NewSender(params.NewNodeUrl, config.AdminKey)
	for {
		oldNodeBlockHeight := client.GetBlockNumber()
		newNodeBlockHeight := newNodeClient.GetBlockNumber()

		log.Infof("sync block, old node %s block height %d, new node %s block height %d",
			client.Url(), oldNodeBlockHeight, newNodeClient.Url(), newNodeBlockHeight)

		if newNodeBlockHeight >= oldNodeBlockHeight {
			break
		}
		wait(1)
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

// --------------------------------
// restart genesis nodes and validators
// --------------------------------
func RestartNetwork() bool {
	StopNetwork()
	StartNetwork()
	StartValidators()
	return true
}

// --------------------------------
// start sync node
// --------------------------------
func StartSyncNode() bool {
	shell.Exec(shStartSyncNode)
	wait(3)

	url := "http://127.0.0.1:22010"
	client := sdk.NewSender(url, config.AdminKey)

	// check current block number
	for i := 0;i<10;i++ {
		currentBlock := client.GetBlockNumber()
		wait(1)
		log.Infof("sync node check current block %d", currentBlock)
	}

	// check admin nonce
	nonce := client.GetNonce(config.AdminKey.Address.Hex())
	if nonce == 0 {
		log.Error("sync node check admin nonce failed")
		return false
	}
	log.Infof("sync node check admin nonce %d", nonce)

	// check PLT total supply
	totalSupply, err := client.PLTTotalSupply("latest")
	if err != nil {
		log.Errorf("sync node check total supply failed, [%v]", err)
		return false
	} else {
		log.Infof("sync node check total supply %d", utils.UnsafeDiv(totalSupply, plt.OnePLT))
	}

	return true
}

func StopSyncNode() bool {
	shell.Exec(shStopSyncNode)
	return true
}

func gc() {
	client = nil
	config.Conf = config.BakConf.DeepCopy()
}

func wait(nBlock int) {
	time.Sleep(time.Duration(config.Conf.BlockPeriod) * time.Duration(nBlock))
}
