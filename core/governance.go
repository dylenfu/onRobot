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
	sv := loadValidatorsConfig()
	start, end, num := sv.ValidatorsIndexStart, sv.ValidatorsIndexEnd, sv.ValidatorsNumber

	admcli = sdk.NewSender(config.Conf.BaseRPCUrl, config.AdminKey)

	// stake
	stakeAmt := plt.MultiPLT(sv.ValidatorInitAmount)
	hashList := make([]common.Hash, 0)
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
	wait(1)
	if err := DumpHashList(hashList); err != nil {
		return
	}

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

	// admin add balance
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
	wait(1)
	if err := DumpHashList(hashList); err != nil {
		return
	}

	// check validators
	wait(config.Conf.RewardEffectivePeriod)
	validators := admcli.GetAllValidators("latest")
	if len(validators) != num {
		log.Error("validators not effective, check palette log")
		return
	}

	for _, v := range validators {
		exist := false
		for i := start; i <= end; i++ {
			nodeAddr := config.Conf.Nodes[i].NodeAddr()
			if nodeAddr == v {
				exist = true
				goto check
			}
		}
	check:
		if !exist {
			log.Errorf("validator %s not exist in params", v.Hex())
			return
		}
	}

	return true
}

func DelValidators() bool {
	return true
}

func Reward() bool {
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
