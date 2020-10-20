package core

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/palettechain/onRobot/config"
	"github.com/palettechain/onRobot/pkg/log"
	"github.com/palettechain/onRobot/pkg/sdk"
)

func AddValidators() (succeed bool) {
	sv := loadValidatorsConfig()
	start, end, num := sv.ValidatorsIndexStart, sv.ValidatorsIndexEnd, sv.ValidatorsNumber

	client = sdk.NewSender(config.Conf.BaseRPCUrl, config.AdminKey)
	newClient := sdk.NewSender(sv.NewNodeUrl, config.AdminKey)

	// send transactions and dump receipt
	hashList := make([]common.Hash, 0)
	for i := start; i <= end; i++ {
		node := config.Conf.Nodes[i]
		hash, err := client.AddValidator(node.Addr(), false)
		if err != nil {
			log.Errorf("failed to add validator %s, hash %s, [%v]", node.Addr().Hex(), hash.Hex(), err)
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
	wait(config.Conf.EffectivePeriod)
	validators := client.GetValidators()
	if len(validators) != num {
		log.Error("validators not effective, check palette log")
		return
	}

	for _, v := range validators {
		exist := false
		for i := start; i <= end; i++ {
			nodeAddr := config.Conf.Nodes[i].Addr()
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
