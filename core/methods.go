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
	"github.com/palettechain/onRobot/pkg/shell"
)

var admcli *sdk.Client

const (
	shInit          = "init_nodes.sh"
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
	admcli = sdk.NewSender(config.Conf.BaseRPCUrl, config.AdminKey)
	blockNumber := admcli.GetBlockNumber()
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

	admcli = sdk.NewSender(config.Conf.BaseRPCUrl, config.AdminKey)
	nonce := admcli.GetNonce(params.Address)
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

	admcli = sdk.NewSender(config.Conf.BaseRPCUrl, config.AdminKey)
	for _, account := range config.Conf.Accounts {
		amount := plt.MultiPLT(params.UserInitAmount)
		to := common.HexToAddress(account)
		if _, err := admcli.PLTTransfer(to, amount); err != nil {
			log.Errorf("transfer to %s err %v", to.Hex(), err)
			return false
		}
	}

	wait(1)

	for _, account := range config.Conf.Accounts {
		owner := common.HexToAddress(account)
		data, err := admcli.BalanceOf(owner, "latest")
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
	wait(1)
	params := loadValidatorsConfig()
	config.Conf.ResetEnv(params.ValidatorsIndexStart, params.ValidatorsNumber)
	shell.Exec(shStopNodes)
	shell.Exec(shClearNodes)
	shell.Exec(shReset)

	wait(1)

	admcli = sdk.NewSender(config.Conf.BaseRPCUrl, config.AdminKey)

	// transfer and check balance
	amount := plt.MultiPLT(params.ValidatorInitAmount)
	for i := params.ValidatorsIndexStart; i <= params.ValidatorsIndexEnd; i++ {
		to := config.Conf.Nodes[i].StakeAddr()
		if hash, err := admcli.PLTTransfer(to, amount); err != nil {
			log.Errorf("url %s, transfer to node%d %s amount %d, hash %s, err %v", config.Conf.BaseRPCUrl, i, to.Hex(), params.ValidatorInitAmount, hash.Hex(), err)
			return false
		}
	}

	wait(1)

	for i := params.ValidatorsIndexStart; i <= params.ValidatorsIndexEnd; i++ {
		nodeAcc := config.Conf.Nodes[i].NodeAddr()
		stkAcc := config.Conf.Nodes[i].StakeAddr()
		nodeAmt, err := admcli.BalanceOf(nodeAcc, "latest")
		if err != nil {
			log.Errorf("query balanceOf %s err %v", nodeAcc.Hex(), err)
			return false
		}
		stkAmt, err := admcli.BalanceOf(stkAcc, "latest")
		if err != nil {
			log.Errorf("query balanceOf %s err %v", stkAcc.Hex(), err)
			return false
		}
		if nodeAmt.Cmp(big.NewInt(0)) > 0 {
			log.Errorf("validator init PLT used for staking, validator address used for reward")
			return false
		}
		log.Infof("%s init balance %d", nodeAcc.Hex(), plt.PrintUPLT(stkAmt))
	}

	// sync blocks
	newNodeClient := sdk.NewSender(params.NewNodeUrl, config.AdminKey)
	for {
		oldNodeBlockHeight := admcli.GetBlockNumber()
		newNodeBlockHeight := newNodeClient.GetBlockNumber()

		log.Infof("sync block, old node %s block height %d, new node %s block height %d",
			admcli.Url(), oldNodeBlockHeight, newNodeClient.Url(), newNodeBlockHeight)

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
	for i := 0; i < 3; i++ {
		currentBlock := client.GetBlockNumber()
		wait(1)
		log.Infof("sync node check current block %d", currentBlock)
	}

	// check admin nonce
	nonce := client.GetNonce(config.AdminAddr.Hex())
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
	admcli = nil
	config.Conf = config.BakConf.DeepCopy()
}

func wait(nBlock int) {
	time.Sleep(time.Duration(config.Conf.BlockPeriod) * time.Duration(nBlock))
}
