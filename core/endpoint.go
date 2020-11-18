package core

import (
	"math/rand"
	"time"

	"github.com/palettechain/onRobot/pkg/frame"
)

func Endpoint() {
	rand.Seed(time.Now().UnixNano())
	initialize()

	// gc function
	frame.Tool.RegGCFunc(gc)

	// remote construct
	frame.Tool.RegMethod("remoteBuild", RemoteBuild)
	frame.Tool.RegMethod("remoteSetup", RemoteSetup)

	// genesis network
	frame.Tool.RegMethod("initGenesis", InitGenesisNetwork)
	frame.Tool.RegMethod("startGenesis", StartGenesisNetwork)
	frame.Tool.RegMethod("stopGenesis", StopGenesisNetwork)
	frame.Tool.RegMethod("clearGenesis", ClearGenesisNetwork)
	frame.Tool.RegMethod("restartGenesis", ReStartGenesisNetwork)
	frame.Tool.RegMethod("resetGenesis", ResetGenesisNetwork)

	// validator network
	frame.Tool.RegMethod("initValidator", InitValidatorNetwork)
	frame.Tool.RegMethod("startValidator", StartValidatorNetwork)
	frame.Tool.RegMethod("stopValidator", StopValidatorNetwork)
	frame.Tool.RegMethod("clearValidator", ClearValidatorNetwork)
	frame.Tool.RegMethod("restartValidator", ReStartValidatorNetwork)
	frame.Tool.RegMethod("resetValidator", ResetValidatorNetwork)

	// total network
	frame.Tool.RegMethod("start", StartAllNetwork)
	frame.Tool.RegMethod("stop", StopAllNetwork)
	frame.Tool.RegMethod("clear", ClearAllNetwork)
	frame.Tool.RegMethod("restart", ReStartAllNetwork)

	// sync node
	//frame.Tool.RegMethod("startSyncNode", StartSyncNode)
	//frame.Tool.RegMethod("stopSyncNode", StopSyncNode)

	// uncle
	frame.Tool.RegMethod("blockNumber", BlockNumber)
	frame.Tool.RegMethod("nonce", Nonce)
	frame.Tool.RegMethod("consistency", Consistency)
	frame.Tool.RegMethod("deposit", Deposit)

	// plt
	frame.Tool.RegMethod("totalSupply", TotalSupply)
	frame.Tool.RegMethod("name", Name)
	frame.Tool.RegMethod("decimal", Decimal)
	frame.Tool.RegMethod("adminBalance", AdminBalance)
	frame.Tool.RegMethod("governanceBalance", GovernanceBalance)
	frame.Tool.RegMethod("balanceOf", BalanceOf)
	frame.Tool.RegMethod("transfer", Transfer)
	frame.Tool.RegMethod("approve", Approve)

	// governance
	frame.Tool.RegMethod("addValidators", AddValidators)
	frame.Tool.RegMethod("delValidators", DelValidators)
	frame.Tool.RegMethod("getValidators", GetValidators)
	frame.Tool.RegMethod("reward", Reward)
	frame.Tool.RegMethod("delegate", Delegate)
	frame.Tool.RegMethod("proposal", Proposal)
	frame.Tool.RegMethod("globalParams", GlobalParams)
}
