package core

import (
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/contracts/native/plt"
	"github.com/palettechain/onRobot/config"
	"github.com/palettechain/onRobot/pkg/log"
	"github.com/palettechain/onRobot/pkg/sdk"
)

// 检查数据一致性(重要): 轮询N个节点，比较其查询所得的lastRewardBlock是否一致。非验证节点同步速度可能会慢上几个块.
func Consistency() (succeed bool) {
	var params struct {
		UrlList []string
	}
	if err := config.LoadParams("Consistency.json", &params); err != nil {
		log.Error(err)
		return
	}

	clients := make([]*sdk.Client, len(params.UrlList))
	for i := 0; i < len(params.UrlList); i++ {
		clients[i] = sdk.NewSender(params.UrlList[i], config.AdminKey)
	}

	currentBlkNo := clients[0].GetBlockNumber() - 10

	var i, blkNo uint64 = 0, 10
	for i = currentBlkNo - blkNo; i < currentBlkNo; i++ {
		lastRdBlk := big.NewInt(0)
		lastRdProposer := common.Address{}
		queryBlkHex := "0x" + strconv.FormatInt(int64(i), 16)
		for i := 0; i < len(params.UrlList); i++ {
			rdBlk, err := clients[i].GetRewardRecordBlock(queryBlkHex)
			if err != nil {
				log.Error(err)
				return
			}
			rdProp, err := clients[i].GetLatestRewardProposer(queryBlkHex)
			if err != nil {
				log.Error(err)
				return
			}

			if i == 0 {
				lastRdBlk = rdBlk
				lastRdProposer = rdProp
				continue
			}
			if lastRdBlk.Cmp(rdBlk) != 0 {
				log.Errorf("%s query result %d, %s query result %d", clients[0].Url(), lastRdBlk.Uint64(), clients[i].Url(), rdBlk.Uint64())
			}
			if lastRdProposer != rdProp {
				log.Errorf("%s query result %s, %s query result %s", clients[0].Url(), lastRdProposer.Hex(), clients[i].Url(), rdProp.Hex())
			}
		}
		log.Infof("last reward block %d, last reward proposer %s", lastRdBlk.Uint64(), lastRdProposer.Hex())
	}

	return true
}

func AddValidators() (succeed bool) {
	wait(1)
	sv := loadValidatorsConfig()
	start, end, _ := sv.ValidatorsIndexStart, sv.ValidatorsIndexEnd, sv.ValidatorsNumber

	admcli = sdk.NewSender(config.Conf.BaseRPCUrl, config.AdminKey)
	nodeList := make([]common.Address, 0)

	// check balance before stake
	for i := start; i <= end; i++ {
		node := config.Conf.Nodes[i]
		nodeList = append(nodeList, node.NodeAddr())
		data, err := admcli.BalanceOf(node.StakeAddr(), "latest")
		if err != nil {
			log.Error("failed to stake for validator %s stake account %s amount %d", node.NodeAddr().Hex(), node.StakeAddr().Hex(), sv.ValidatorInitAmount)
			return
		}
		log.Infof("%s balance before stake %d", node.NodeAddr().Hex(), plt.PrintUPLT(data))
	}

	// stake and dump event log
	stakeAmt := plt.MultiPLT(sv.ValidatorInitAmount)
	hashList := make([]common.Hash, 0)
	log.Infof("validators stake at block %d", admcli.GetBlockNumber())
	for i := start; i <= end; i++ {
		node := config.Conf.Nodes[i]
		nodecli := sdk.NewSender(config.Conf.BaseRPCUrl, node.StakePrivateKey())
		hash, err := nodecli.Stake(node.NodeAddr(), node.StakeAddr(), stakeAmt, false)
		if err != nil {
			log.Error("failed to stake for validator %s stake account %s amount %d", node.NodeAddr().Hex(), node.StakeAddr().Hex(), sv.ValidatorInitAmount)
			return
		}
		hashList = append(hashList, hash)
	}
	wait(2)
	if err := DumpHashList(hashList, "stake"); err != nil {
		return
	}
	wait(2 * config.Conf.RewardEffectivePeriod)

	// check balance after stake
	for i := start; i <= end; i++ {
		node := config.Conf.Nodes[i]
		data, err := admcli.BalanceOf(node.StakeAddr(), "latest")
		if err != nil {
			log.Error("failed to stake for validator %s stake account %s amount %d", node.NodeAddr().Hex(), node.StakeAddr().Hex(), sv.ValidatorInitAmount)
			return
		}
		log.Infof("%s balance after stake %d", node.NodeAddr().Hex(), plt.PrintUPLT(data))
	}

	// admin add validators
	hashList = make([]common.Hash, 0)
	for i := start; i <= end; i++ {
		node := config.Conf.Nodes[i]
		hash, err := admcli.AddValidator(node.NodeAddr(), node.StakeAddr(), false)
		if err != nil {
			log.Errorf("failed to add validator %s, hash %s, [%v]", node.NodeAddr().Hex(), hash.Hex(), err)
			return
		}
		hashList = append(hashList, hash)
	}
	log.Infof("add validator at block %d", admcli.GetBlockNumber())
	wait(2)
	if err := DumpHashList(hashList, "addValidator"); err != nil {
		return
	}

	log.Infof("check pending validators at block %d", admcli.GetBlockNumber())
	validators := admcli.GetAllValidators("latest")
	if !HasAddrs(validators, nodeList) {
		log.Error("validators not pending, check palette log")
		return
	}
	log.Infof("check pending validators success")

	// check validators
	wait(config.Conf.RewardEffectivePeriod)
	log.Infof("check effective validators at block %d", admcli.GetBlockNumber())
	effectiveValidators := admcli.GetEffectiveValidators("latest")
	if !HasAddrs(effectiveValidators, nodeList) {
		log.Error("validators not effective, check palette log")
		return
	}
	for _, v := range effectiveValidators {
		log.Infof("add validator %s success", v.Hex())
	}

	return true
}

func GetValidators() bool {
	admcli = sdk.NewSender(config.Conf.BaseRPCUrl, config.AdminKey)
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

	admcli = sdk.NewSender(config.Conf.BaseRPCUrl, config.AdminKey)
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
