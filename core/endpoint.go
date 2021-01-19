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
	frame.Tool.RegMethod("grep", Grep)

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

	// palette side chain environment
	frame.Tool.RegMethod("polyHeight", PolyHeight)
	frame.Tool.RegMethod("plt-deploy", PLTDeployCrossChainContract)
	frame.Tool.RegMethod("plt-ownership", PLTTransferOwnerShip)
	frame.Tool.RegMethod("plt-upgradeECCM", PLTUpgradeECCM)
	frame.Tool.RegMethod("plt-registerSideChain", PLTRegisterSideChain)
	frame.Tool.RegMethod("plt-updateSideChain", PLTUpdateSideChain)
	frame.Tool.RegMethod("plt-quitSideChain", PLTQuitSideChain)
	frame.Tool.RegMethod("plt-approveRegisterSideChain", PLTApproveRegisterSideChain)
	frame.Tool.RegMethod("plt-approveUpdateSideChain", PLTApproveUpdateSideChain)
	frame.Tool.RegMethod("plt-approveQuitSideChain", PLTApproveQuitSideChain)
	frame.Tool.RegMethod("plt-ccmp", PLTSetCCMP)
	frame.Tool.RegMethod("plt-bindProxy", PLTBindProxy)
	frame.Tool.RegMethod("plt-bindAsset", PLTBindAsset)
	frame.Tool.RegMethod("plt-syncGenesis", SyncGenesis)
	frame.Tool.RegMethod("plt-changePaletteBookKeeper", PLTChangeBookKeepers)
	frame.Tool.RegMethod("plt-changePolyBookKeeper", PolyChangeBookKeepers)

	// ethereum side chain environment
	frame.Tool.RegMethod("eth-deploy-eccd", ETHDeployECCD)
	frame.Tool.RegMethod("eth-deploy-eccm", ETHDeployECCM)
	frame.Tool.RegMethod("eth-deploy-ccmp", ETHDeployCCMP)
	frame.Tool.RegMethod("eth-ownership", ETHTransferOwnership)
	frame.Tool.RegMethod("eth-registerSideChain", ETHRegisterSideChain)
	frame.Tool.RegMethod("eth-approveRegisterSideChain", ETHApproveRegisterSideChain)
	frame.Tool.RegMethod("eth-deploy-plt", ETHDeployPLTAsset)
	frame.Tool.RegMethod("eth-deploy-plt-proxy", ETHDeployPLTProxy)
	frame.Tool.RegMethod("eth-bind-plt-proxy", ETHBindPLTProxy)
	frame.Tool.RegMethod("eth-plt-ccmp", ETHSetPLTCCMP)
	frame.Tool.RegMethod("eth-deploy-nft-asset", ETHDeployNFTAsset)
	frame.Tool.RegMethod("eth-deploy-nft-proxy", ETHDeployNFTProxy)
	frame.Tool.RegMethod("eth-nft-ccmp", ETHSetNFTCCMP)
	frame.Tool.RegMethod("eth-bind-nft-proxy", ETHBindNFTProxy)

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

	// evm test
	frame.Tool.RegMethod("test-deploy1", TestDeploy1)
	frame.Tool.RegMethod("test-deploy2", TestDeploy2)
	frame.Tool.RegMethod("test-evm1", TestEVM1)
	frame.Tool.RegMethod("test-evm2", TestEVM2)
}
