package core

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/contracts/native/plt"
	"github.com/palettechain/onRobot/config"
	"github.com/palettechain/onRobot/pkg/log"
	"github.com/palettechain/onRobot/pkg/sdk"
)

func AddValidators() (succeed bool) {
	var params struct {
		InitAmount int
	}

	if err := config.LoadParams("AddValidators.json", &params); err != nil {
		log.Error(err)
		return
	}

	nodes := config.Conf.ValidatorNodes()
	balances := make([]int, len(nodes))

	// check and add balance before stake
	for i, node := range nodes {
		data, err := admcli.BalanceOf(node.StakeAddr(), "latest")
		if err != nil {
			log.Error("failed to check %s balance", node.NodeAddr().Hex())
			return
		}
		balances[i] = int(plt.PrintUPLT(data))
		log.Infof("%s balance before stake %d", node.NodeAddr().Hex(), plt.PrintUPLT(data))
	}
	depoistHashList := make([]common.Hash, 0)
	for i, balance := range balances {
		if balance < params.InitAmount {
			addAmount := params.InitAmount - balance
			node := nodes[i]
			hash, err := admcli.PLTTransfer(node.StakeAddr(), plt.MultiPLT(addAmount))
			if err != nil {
				log.Errorf("failed to deposit to node %s, amount %d", node.NodeAddr().Hex(), addAmount)
				return
			} else {
				depoistHashList = append(depoistHashList, hash)
				balances[i] += addAmount
			}
		}
	}
	wait(2)
	if err := DumpHashList(depoistHashList, "deposit for validator"); err != nil {
		log.Error(err)
		return
	}

	// stake and dump event log
	stakeHashList := make([]common.Hash, 0)
	log.Infof("validators stake at block %d", admcli.GetBlockNumber())
	for i, node := range nodes {
		nodecli := sdk.NewSender(node.RPCAddr(), node.StakePrivateKey())
		stkAmt := balances[i]
		hash, err := nodecli.Stake(node.NodeAddr(), node.StakeAddr(), plt.MultiPLT(stkAmt), false)
		if err != nil {
			log.Error("failed to stake for validator %s stake account %s amount %d", node.NodeAddr().Hex(), node.StakeAddr().Hex(), stkAmt)
			return
		}
		stakeHashList = append(stakeHashList, hash)
	}
	wait(2)
	if err := DumpHashList(stakeHashList, "stake"); err != nil {
		return
	}
	wait(2 * config.Conf.RewardEffectivePeriod)

	// check balance after stake
	for _, node := range nodes {
		data, err := admcli.BalanceOf(node.StakeAddr(), "latest")
		if err != nil {
			log.Error("failed to check %s's balance after stake, err :%v", node.NodeAddr().Hex(), err)
			return
		}
		log.Infof("%s balance after stake %d", node.NodeAddr().Hex(), plt.PrintUPLT(data))
	}

	// admin add validators
	adminAddValidatorHashList := make([]common.Hash, 0)
	for _, node := range nodes {
		hash, err := admcli.AddValidator(node.NodeAddr(), node.StakeAddr(), false)
		if err != nil {
			log.Errorf("failed to add validator %s, hash %s, [%v]", node.NodeAddr().Hex(), hash.Hex(), err)
			return
		}
		adminAddValidatorHashList = append(adminAddValidatorHashList, hash)
	}
	log.Infof("add validator at block %d", admcli.GetBlockNumber())
	wait(2)
	if err := DumpHashList(adminAddValidatorHashList, "admin add validators"); err != nil {
		return
	}

	log.Infof("check pending validators at block %d", admcli.GetBlockNumber())
	expectNodeAddrList := make([]common.Address, len(nodes))
	for i, node := range nodes {
		expectNodeAddrList[i] = node.NodeAddr()
	}
	validators := admcli.GetAllValidators("latest")
	if !HasAddrs(validators, expectNodeAddrList) {
		log.Error("validators not pending, check palette log")
		return
	}
	log.Infof("check pending validators success")

	// check validators
	wait(config.Conf.RewardEffectivePeriod)
	log.Infof("check effective validators at block %d", admcli.GetBlockNumber())
	effectiveValidators := admcli.GetEffectiveValidators("latest")
	if !HasAddrs(effectiveValidators, expectNodeAddrList) {
		log.Error("validators not effective, check palette log")
		return
	}
	for _, v := range effectiveValidators {
		log.Infof("add validator %s success", v.Hex())
	}

	return true
}

func GetValidators() bool {
	baseUrl := config.Conf.Nodes[0].RPCAddr()
	admcli = sdk.NewSender(baseUrl, config.AdminKey)
	effectiveValidators := admcli.GetEffectiveValidators("latest")
	for _, v := range effectiveValidators {
		log.Infof("validator %s", v.Hex())
	}
	return true
}

func DelValidators() bool {
	return true
}

func Reward() (succeed bool) {
	var params struct {
		RewardBlocks                   int
		ExpectRewardPoolAmount         int
		ExpectRewardAmountPerValidator int
	}

	if err := config.LoadParams("Reward.json", &params); err != nil {
		log.Error(err)
		return
	}
	rewardPool := common.HexToAddress(config.Conf.BaseRewardPool)
	nodes := config.Conf.ValidatorNodes()
	addrs := append(nodes.StakeAccounts(), rewardPool)

	// check balance before reward
	blkBeforeCheckReward := admcli.GetBlockNumber()
	log.Infof("check balance before testing reward at block %d", blkBeforeCheckReward)
	balancesBeforeCheckReward, err := getBalances(addrs, BlockNumber2Hex(blkBeforeCheckReward))
	if err != nil {
		log.Error("failed to check balance before testing, err: %v", err)
		return
	}

	// waiting for blocks
	wait(params.RewardBlocks)

	// check balance after reward
	blkAfterCheckReward := blkBeforeCheckReward + uint64(params.RewardBlocks)
	log.Infof("check balance after testing reward at block %d", blkAfterCheckReward)
	balancesAfterCheckReward, err := getBalances(addrs, BlockNumber2Hex(blkAfterCheckReward))
	if err != nil {
		log.Error("failed to check balance after testing, err: %v", err)
		return
	}

	res, err := subBalanceMap(balancesBeforeCheckReward, balancesAfterCheckReward)
	if err != nil {
		log.Error(err)
		return
	}
	var expect int
	for addr, v := range res {
		if addr == rewardPool {
			expect = params.ExpectRewardPoolAmount
		} else {
			expect = params.ExpectRewardAmountPerValidator
		}

		if v != expect {
			log.Errorf("%s reward amount, expect %d = actual %d", addr.Hex(), expect, v)
			return
		} else {
			log.Infof("%s reward amount %d", addr.Hex(), v)
		}
	}

	return true
}

// 节点代理用户质押一定数量的PLT
func Delegate() bool {
	return true
}

// propose and vote, proposal type:
// 1: mint price,
// 2: gas fee,
// 3: reward period
func Proposal() (succeed bool) {
	var params struct {
		ProposerNodeIndex int
		ProposalType      uint8
		ProposalValue     int
		VoteNodeIndexList []int
	}

	if err := config.LoadParams("Proposal.json", &params); err != nil {
		log.Error(err)
		return
	}

	log.Infof("check proposal params......")
	// check proposal type
	if params.ProposalType == 0 || params.ProposalType > 3 {
		log.Errorf("invalid proposal type %d", params.ProposalType)
		return
	}
	// check proposal value
	if params.ProposalValue <= 0 {
		log.Errorf("invalid proposal value %d", params.ProposalValue)
		return
	}
	proposalValue := plt.MultiPLT(params.ProposalValue)

	log.Infof("check validator authority......")
	// get validators
	nodes, err := getAndCheckValidator(append(params.VoteNodeIndexList, params.ProposerNodeIndex))
	if err != nil {
		log.Error(err)
		return
	}
	proposerNode := nodes[0]
	voteNodes := nodes[1:]

	log.Infof("propose new proposal......")
	// proposer send proposal
	proposerCli := sdk.NewSender(config.Conf.Nodes[0].RPCAddr(), proposerNode.PrivateKey())
	hash, err := proposerCli.Propose(params.ProposalType, proposalValue)
	if err != nil {
		log.Errorf("%s failed to propose, err %v", proposerNode.NodeAddr().Hex(), err)
		return
	}
	wait(2)
	proposalID, proposal, err := admcli.GetProposalFromReceipt(hash)
	if err != nil {
		log.Error(err)
		return
	} else {
		log.Infof("proposalID %s, proposer %s, proposal type %d, value %v, end block %d",
			proposalID.Hex(), proposerNode.NodeAddr().Hex(), proposal.ProposalType, proposalValue, proposal.EndBlock.Uint64())
	}

	log.Infof("votinng......")
	// vote and dump hash list
	voteHashList := make([]common.Hash, 0)
	for _, voteNode := range voteNodes {
		voteNodeCli := sdk.NewSender(config.Conf.Nodes[0].RPCAddr(), voteNode.PrivateKey())
		hash, err := voteNodeCli.Vote(proposalID)
		if err != nil {
			log.Error(err)
			return
		}
		voteHashList = append(voteHashList, hash)
		log.Infof("%s vote to proposalID %s, hash %s", voteNode.NodeAddr().Hex(), proposalID.Hex(), hash.Hex())
	}
	wait(2)
	if err := DumpHashList(voteHashList, "vote"); err != nil {
		log.Error(err)
		return
	}

	wait(config.Conf.RewardEffectivePeriod)

	// check proposal status
	proposal, err = admcli.GetProposal(proposalID, "latest")
	if err != nil {
		log.Error(err)
		return
	}
	if proposal.Passed != true {
		log.Errorf("proposal %s should be passed", proposalID.Hex())
		return
	}

	// check global params
	data, err := admcli.GetGlobalParams(params.ProposalType, "latest")
	if err != nil {
		log.Error(err)
		return
	}

	expect, actual := params.ProposalValue, int(plt.PrintUPLT(data))
	if expect != actual {
		log.Errorf("proposal failed to set global params, expect %d, actual %d", expect, actual)
	} else {
		log.Infof("global params changed to %d", actual)
	}
	return true
}

func GlobalParams() (succeed bool) {
	var params struct {
		ProposalType uint8
	}

	if err := config.LoadParams("GlobalParams.json", &params); err != nil {
		log.Error(err)
		return
	}

	log.Infof("check proposal type......")
	if params.ProposalType == 0 || params.ProposalType > 3 {
		log.Errorf("invalid proposal type %d", params.ProposalType)
		return
	}

	value, err := admcli.GetGlobalParams(params.ProposalType, "latest")
	if err != nil {
		log.Error(err)
		return
	}

	log.Infof("global params [%d]:value %d", params.ProposalType, plt.PrintUPLT(value))
	return true
}
