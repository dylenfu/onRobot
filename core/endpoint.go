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

	// basic
	frame.Tool.RegMethod("demo", Demo)
	frame.Tool.RegMethod("blockNumber", BlockNumber)
	frame.Tool.RegMethod("nonce", Nonce)

	// network
	frame.Tool.RegMethod("reset", ResetNetwork)
	frame.Tool.RegMethod("start", StartNetwork)
	frame.Tool.RegMethod("stop", StopNetwork)
	frame.Tool.RegMethod("clear", ClearNetwork)
	frame.Tool.RegMethod("restart", RestartNetwork)

	// sync node
	frame.Tool.RegMethod("startSyncNode", StartSyncNode)
	frame.Tool.RegMethod("stopSyncNode", StopSyncNode)

	// plt
	frame.Tool.RegMethod("totalSupply", TotalSupply)
	frame.Tool.RegMethod("name", Name)
	frame.Tool.RegMethod("decimal", Decimal)
	frame.Tool.RegMethod("adminBalance", BalanceOf)
	frame.Tool.RegMethod("governanceBalance", GovernanceBalance)
	frame.Tool.RegMethod("balanceOf", BalanceOf)
	frame.Tool.RegMethod("transfer", Transfer)
	frame.Tool.RegMethod("approve", Approve)

	// validators manage
	frame.Tool.RegMethod("initValidators", InitValidators)
	frame.Tool.RegMethod("startValidators", StartValidators)
	frame.Tool.RegMethod("stopValidators", StopValidators)
	frame.Tool.RegMethod("clearValidators", ClearValidators)

	// governance
	frame.Tool.RegMethod("consistency", Consistency)
	frame.Tool.RegMethod("addValidators", AddValidators)
	frame.Tool.RegMethod("delValidators", DelValidators)
	frame.Tool.RegMethod("reward", Reward)
	frame.Tool.RegMethod("stake", Stake)
	frame.Tool.RegMethod("propose", Propose)
	frame.Tool.RegMethod("vote", Vote)
}
