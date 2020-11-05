package core

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/palettechain/onRobot/config"
	"github.com/palettechain/onRobot/pkg/log"
	"github.com/palettechain/onRobot/pkg/sdk"
	"math/big"
	"strconv"
)

// 检查数据一致性(重要):
// 5个创世验证节点启动后，再添加3个新的验证节点，轮询这8个节点，比较其查询所得的lastRewardBlock是否一致。
func Consistency() (succeed bool) {
	var params struct {
		UrlList []string
	}

	log.Infof("-----------------1")
	if err := config.LoadParams("Consistency.json", &params); err != nil {
		log.Error(err)
		return
	}
	log.Infof("-----------------2")
	clients := make([]*sdk.Client, len(params.UrlList))
	for i := 0; i < len(params.UrlList); i++ {
		clients[i] = sdk.NewSender(params.UrlList[i], config.AdminKey)
	}
	log.Infof("-----------------3")
	queryBlkNo := int64(config.Conf.RewardEffectivePeriod + 2)
	queryBlkHex := "0x" + strconv.FormatInt(queryBlkNo, 16)
	lastRdBlk := big.NewInt(0)
	for i := 0; i < len(params.UrlList); i++ {
		data, err := clients[i].GetRewardRecordBlock(queryBlkHex)
		if err != nil {
			log.Error(err)
			return
		}
		if i == 0 {
			lastRdBlk = data
			continue
		}
		if lastRdBlk.Cmp(data) != 0 {
			log.Errorf("%d query result %d, %s query result %d", clients[0].Url(), lastRdBlk.Uint64(), clients[i].Url(), data.Uint64())
		}
	}
	log.Infof("last reward block %d", lastRdBlk.Uint64())
	return true
}

func AddValidators() (succeed bool) {
	sv := loadValidatorsConfig()
	start, end, num := sv.ValidatorsIndexStart, sv.ValidatorsIndexEnd, sv.ValidatorsNumber

	client = sdk.NewSender(config.Conf.BaseRPCUrl, config.AdminKey)
	newClient := sdk.NewSender(sv.NewNodeUrl, config.AdminKey)

	// send transactions and dump receipt
	hashList := make([]common.Hash, 0)
	for i := start; i <= end; i++ {
		node := config.Conf.Nodes[i]
		hash, err := client.AddValidator(node.NodeAddr(), false)
		if err != nil {
			log.Errorf("failed to add validator %s, hash %s, [%v]", node.NodeAddr().Hex(), hash.Hex(), err)
			return
		}
		hashList = append(hashList, hash)
	}
	wait(1)
	for _, hash := range hashList {
		if err := client.DumpEventLog(hash); err != nil {
			log.Errorf("failed to dump receipt, hash %s, [%v]", hash.Hex(), err)
			return
		}
		if err := newClient.DumpEventLog(hash); err != nil {
			log.Errorf("failed to dump receipt, hash %s, [%v]", hash.Hex(), err)
			return
		}
		log.Infof("------------------------------------------------------")
	}

	// check validators
	wait(config.Conf.RewardEffectivePeriod)
	validators := client.GetValidators()
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
