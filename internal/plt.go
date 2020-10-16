package internal

import (
	"github.com/ethereum/go-ethereum/contracts/native/plt"
	"github.com/ethereum/go-ethereum/contracts/native/utils"
	"github.com/palettechain/onRobot/pkg/log"
	"github.com/palettechain/onRobot/pkg/sdk"
)

func PLTBalanceOf() bool {
	var params struct {
		RpcUrl string
		Owner  string
	}

	if err := loadParams("PLTBalanceOf.json", &params); err != nil {
		log.Error(err)
		return false
	}
	key := loadAccount(params.Owner)
	client := sdk.NewSender(params.RpcUrl, key)

	balance := client.BalanceOf(key.Address, "latest")
	log.Infof("balance %d", utils.UnsafeDiv(balance, plt.OnePLT))

	return true
}
