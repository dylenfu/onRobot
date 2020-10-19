package core

import (
	"math/rand"
	"time"

	"github.com/palettechain/onRobot/pkg/frame"
)

func Endpoint() {
	rand.Seed(time.Now().UnixNano())

	// gc function
	frame.Tool.RegGCFunc(gc)

	// demo
	frame.Tool.RegMethod("demo", Demo)

	// network
	frame.Tool.RegMethod("reset", ResetNetwork)
	frame.Tool.RegMethod("start", StartNetwork)

	// plt
	frame.Tool.RegMethod("totalSupply", TotalSupply)
	frame.Tool.RegMethod("decimal", Decimal)
	frame.Tool.RegMethod("adminBalance", BalanceOf)
	frame.Tool.RegMethod("governanceBalance", GovernanceBalance)
	frame.Tool.RegMethod("balanceOf", BalanceOf)
	frame.Tool.RegMethod("transfer", Transfer)
	frame.Tool.RegMethod("approve", Approve)

	// governance
	frame.Tool.RegMethod("addValidator", AddValidator)
	frame.Tool.RegMethod("delValidator", DelValidator)
	frame.Tool.RegMethod("reward", Reward)
	frame.Tool.RegMethod("stake", Stake)
	frame.Tool.RegMethod("propose", Propose)
	frame.Tool.RegMethod("vote", Vote)
}
