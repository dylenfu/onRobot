package core

import (
	"github.com/palettechain/onRobot/config"
	"github.com/palettechain/onRobot/pkg/log"
	"github.com/palettechain/onRobot/pkg/sdk"
)

func AddValidator() (succeed bool) {
	var params struct {
		RpcUrl string
	}

	// add validators
	sv := loadValidatorsConfig()
	start, end, num := sv.ValidatorsIndexStart, sv.ValidatorsIndexEnd, sv.ValidatorsNumber

	client = sdk.NewSender(params.RpcUrl, config.AdminKey)
	for i := start; i <= end; i++ {
		node := config.Conf.Nodes[i]
		if _, err := client.AddValidator(node.Addr(), false); err != nil {
			log.Errorf("failed to add validator, [%v]", err)
			return
		}
	}
	wait(1)

	wait(config.Conf.EffectivePeriod)

	// check validators
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

func DelValidator() bool {
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
