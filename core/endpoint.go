package core

import (
	"math/rand"
	"time"

	"github.com/palettechain/onRobot/pkg/frame"
)

func Endpoint() {
	rand.Seed(time.Now().UnixNano())

	frame.Tool.RegGCFunc(gc)

	frame.Tool.RegMethod("demo", Demo)
	frame.Tool.RegMethod("reset", ResetNetwork)

	// plt
	frame.Tool.RegMethod("totalSupply", PLTTotalSupply)
	frame.Tool.RegMethod("decimal", PLTDecimal)
	frame.Tool.RegMethod("adminBalance", PLTBalanceOf)
	frame.Tool.RegMethod("governanceBalance", GovernanceBalance)
	frame.Tool.RegMethod("balanceOf", PLTBalanceOf)
	frame.Tool.RegMethod("transfer", PLTTransfer)
}
