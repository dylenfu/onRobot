package core

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/contracts/native/plt"
	"github.com/palettechain/onRobot/config"
	"github.com/palettechain/onRobot/pkg/log"
	"github.com/palettechain/onRobot/pkg/sdk"
)

func AddValidators() (succeed bool) {
	var params struct {
		InitAmount int
	}

	if err := config.LoadParams("AddValidators.json", &params); err != nil {
		log.Error(err)
		return
	}

	nodes := config.Conf.ValidatorNodes()
	balances := make([]int, len(nodes))

	// check and add balance before stake
	for i, node := range nodes {
		data, err := admcli.BalanceOf(node.StakeAddr(), "latest")
		if err != nil {
			log.Error("failed to check %s balance", node.NodeAddr().Hex())
			return
		}
		balances[i] = int(plt.PrintUPLT(data))
		log.Infof("%s balance before stake %d", node.NodeAddr().Hex(), plt.PrintUPLT(data))
	}
	depoistHashList := make([]common.Hash, 0)
	for i, balance := range balances {
		if balance < params.InitAmount {
			addAmount := params.InitAmount - balance
			node := nodes[i]
			hash, err := admcli.PLTTransfer(node.StakeAddr(), plt.MultiPLT(addAmount))
			if err != nil {
				log.Errorf("failed to deposit to node %s, amount %d", node.NodeAddr().Hex(), addAmount)
				return
			} else {
				depoistHashList = append(depoistHashList, hash)
				balances[i] += addAmount
			}
		}
	}
	wait(2)
	if err := DumpHashList(depoistHashList, "deposit for validator"); err != nil {
		log.Error(err)
		return
	}

	// stake and dump event log
	stakeHashList := make([]common.Hash, 0)
	log.Infof("validators stake at block %d", admcli.GetBlockNumber())
	for i, node := range nodes {
		nodecli := sdk.NewSender(node.RPCAddr(), node.StakePrivateKey())
		stkAmt := balances[i]
		hash, err := nodecli.Stake(node.NodeAddr(), node.StakeAddr(), plt.MultiPLT(stkAmt), false)
		if err != nil {
			log.Error("failed to stake for validator %s stake account %s amount %d", node.NodeAddr().Hex(), node.StakeAddr().Hex(), stkAmt)
			return
		}
		stakeHashList = append(stakeHashList, hash)
	}
	wait(2)
	if err := DumpHashList(stakeHashList, "stake"); err != nil {
		return
	}
	wait(2 * config.Conf.RewardEffectivePeriod)

	// check balance after stake
	for _, node := range nodes {
		data, err := admcli.BalanceOf(node.StakeAddr(), "latest")
		if err != nil {
			log.Error("failed to check %s's balance after stake, err :%v", node.NodeAddr().Hex(), err)
			return
		}
		log.Infof("%s balance after stake %d", node.NodeAddr().Hex(), plt.PrintUPLT(data))
	}

	// admin add validators
	adminAddValidatorHashList := make([]common.Hash, 0)
	for _, node := range nodes {
		hash, err := admcli.AddValidator(node.NodeAddr(), node.StakeAddr(), false)
		if err != nil {
			log.Errorf("failed to add validator %s, hash %s, [%v]", node.NodeAddr().Hex(), hash.Hex(), err)
			return
		}
		adminAddValidatorHashList = append(adminAddValidatorHashList, hash)
	}
	log.Infof("add validator at block %d", admcli.GetBlockNumber())
	wait(2)
	if err := DumpHashList(adminAddValidatorHashList, "admin add validators"); err != nil {
		return
	}

	log.Infof("check pending validators at block %d", admcli.GetBlockNumber())
	expectNodeAddrList := make([]common.Address, len(nodes))
	for i, node := range nodes {
		expectNodeAddrList[i] = node.NodeAddr()
	}
	validators := admcli.GetAllValidators("latest")
	if !HasAddrs(validators, expectNodeAddrList) {
		log.Error("validators not pending, check palette log")
		return
	}
	log.Infof("check pending validators success")

	// check validators
	wait(config.Conf.RewardEffectivePeriod)
	log.Infof("check effective validators at block %d", admcli.GetBlockNumber())
	effectiveValidators := admcli.GetEffectiveValidators("latest")
	if !HasAddrs(effectiveValidators, expectNodeAddrList) {
		log.Error("validators not effective, check palette log")
		return
	}
	for _, v := range effectiveValidators {
		log.Infof("add validator %s success", v.Hex())
	}

	return true
}

func GetValidators() bool {
	baseUrl := config.Conf.Nodes[0].RPCAddr()
	admcli = sdk.NewSender(baseUrl, config.AdminKey)
	effectiveValidators := admcli.GetEffectiveValidators("latest")
	for _, v := range effectiveValidators {
		log.Infof("validator %s", v.Hex())
	}
	return true
}

func DelValidators() bool {
	return true
}

func Reward() (succeed bool) {
	var params struct {
		StakeAmountIsSame              bool
		RewardBlocks                   int
		ValidatorsIndexStart           int
		ValidatorsNumber               int
		ExpectRewardPoolAmount         int
		ExpectRewardAmountPerValidator int
	}

	if err := config.LoadParams("CheckReward.json", &params); err != nil {
		log.Error(err)
		return
	}
	baseUrl := config.Conf.Nodes[0].RPCAddr()
	admcli = sdk.NewSender(baseUrl, config.AdminKey)
	rewardPool := common.HexToAddress(config.Conf.BaseRewardPool)

	// check balance before reward
	start, end := params.ValidatorsIndexStart, params.ValidatorsIndexStart+params.ValidatorsNumber-1
	balancesBeforeCheckReward := make(map[common.Address]int)
	log.Infof("check balance before testing reward at block %d", admcli.GetBlockNumber())
	for i := start; i <= end; i++ {
		node := config.Conf.Nodes[i]
		addr := node.NodeAddr()
		if balance, err := admcli.BalanceOf(addr, "latest"); err != nil {
			log.Errorf("%s check balance err %v", addr.Hex(), err)
			return
		} else {
			balancesBeforeCheckReward[addr] = int(plt.PrintUPLT(balance))
		}
	}
	if balance, err := admcli.BalanceOf(rewardPool, "latest"); err != nil {
		log.Errorf("%s check balance err %v", rewardPool.Hex(), err)
		return
	} else {
		balancesBeforeCheckReward[rewardPool] = int(plt.PrintUPLT(balance))
	}

	// waiting for blocks
	wait(params.RewardBlocks + 2)

	// check balance after reward
	balancesAfterCheckReward := make(map[common.Address]int)
	log.Infof("check balance after testing reward at block %d, waited for %d blocks", admcli.GetBlockNumber(), params.RewardBlocks)
	for i := start; i <= end; i++ {
		node := config.Conf.Nodes[i]
		addr := node.NodeAddr()
		balance, err := admcli.BalanceOf(addr, "latest")
		if err != nil {
			log.Errorf("%s check balance err %v", addr.Hex(), err)
			return
		}
		balancesAfterCheckReward[addr] = int(plt.PrintUPLT(balance))
	}
	if balance, err := admcli.BalanceOf(rewardPool, "latest"); err != nil {
		log.Errorf("%s check balance err %v", rewardPool.Hex(), err)
		return
	} else {
		balancesAfterCheckReward[rewardPool] = int(plt.PrintUPLT(balance))
	}

	for addr, bBefRd := range balancesBeforeCheckReward {
		bAftRd, exist := balancesAfterCheckReward[addr]
		if !exist {
			log.Errorf("missing check %s's balance after reward", addr.Hex())
			return
		}
		amt := bAftRd - bBefRd
		expect := params.ExpectRewardAmountPerValidator
		if addr == rewardPool {
			expect = params.ExpectRewardPoolAmount
		}

		if amt != expect {
			log.Errorf("%s amount expect %d, actual %d", addr.Hex(), expect, amt)
		} else {
			log.Infof("%s reward amount %d", addr.Hex(), amt)
		}
	}

	return true
}

func Stake() bool {
	return true
}

func Propose() bool {
	return true
}

func Vote() bool {
	return true
}
