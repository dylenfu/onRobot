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
	frame.Tool.RegMethod("init", InitAllNetwork)
	frame.Tool.RegMethod("start", StartAllNetwork)
	frame.Tool.RegMethod("stop", StopAllNetwork)
	frame.Tool.RegMethod("clear", ClearAllNetwork)
	frame.Tool.RegMethod("restart", ReStartAllNetwork)
	frame.Tool.RegMethod("reset", ResetAllNetwork)

	// spare nodes

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
	frame.Tool.RegMethod("getValidators", GetValidators)
	frame.Tool.RegMethod("reward", Reward)
	frame.Tool.RegMethod("fakeReward", FakeReward)
	frame.Tool.RegMethod("delegate", Delegate)
	frame.Tool.RegMethod("showDelegate", ShowDelegateAmount)
	frame.Tool.RegMethod("proposal", Proposal)
	frame.Tool.RegMethod("globalParams", GlobalParams)
	frame.Tool.RegMethod("spare", SpareNode)
	frame.Tool.RegMethod("delValidators", DelValidators)
	frame.Tool.RegMethod("period", RewardPeriod)
	frame.Tool.RegMethod("stakeAmount", StakeAmount)
	frame.Tool.RegMethod("dumpBlock", DumpBlock)

	// cross chain environment and deploy
	frame.Tool.RegMethod("polyHeight", PolyHeight)
	frame.Tool.RegMethod("deploy", DeployCrossChainContract)
	frame.Tool.RegMethod("upgradeECCM", UpgradeECCM)
	frame.Tool.RegMethod("registerSideChain", RegisterSideChain)
	frame.Tool.RegMethod("updateSideChain", UpdateSideChain)
	frame.Tool.RegMethod("quitSideChain", QuitSideChain)
	frame.Tool.RegMethod("approveRegisterSideChain", ApproveRegisterSideChain)
	frame.Tool.RegMethod("approveUpdateSideChain", ApproveUpdateSideChain)
	frame.Tool.RegMethod("approveQuitSideChain", ApproveQuitSideChain)
	frame.Tool.RegMethod("ccmp", SetCCMP)
	frame.Tool.RegMethod("bindProxy", BindProxy)
	frame.Tool.RegMethod("bindAsset", BindAsset)
	frame.Tool.RegMethod("getProof", GetProof)
	frame.Tool.RegMethod("syncGenesis", SyncGenesis)
	frame.Tool.RegMethod("changePaletteBookKeeper", ChangePaletteBookKeepers)
	frame.Tool.RegMethod("changePolyBookKeeper", ChangePolyBookKeepers)

	// evm
	//frame.Tool.RegMethod("deploy", Deploy)
	//frame.Tool.RegMethod("evm", EVM)

	// plt cross chain
	frame.Tool.RegMethod("plt-mint", PLTMint)
	frame.Tool.RegMethod("plt-burn", PLTBurn)
	frame.Tool.RegMethod("plt-lock", PLTLock)

	// nft
	frame.Tool.RegMethod("nft-deploy", NFTDeploy)
	frame.Tool.RegMethod("nft-transfer", NFTTransfer)
	frame.Tool.RegMethod("nft-balance", NFTBalance)
	frame.Tool.RegMethod("nft-token-owner", NFTTokenOwner)

	// nft cross chain
	frame.Tool.RegMethod("nft-mint", NFTMint)
	frame.Tool.RegMethod("nft-burn", NFTBurn)
	frame.Tool.RegMethod("nft-lock", NFTLock)
}
