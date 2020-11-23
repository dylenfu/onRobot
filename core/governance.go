package core

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/contracts/native/governance"
	"github.com/ethereum/go-ethereum/contracts/native/plt"
	"github.com/ethereum/go-ethereum/contracts/native/utils"
	"github.com/palettechain/onRobot/config"
	"github.com/palettechain/onRobot/pkg/log"
	"github.com/palettechain/onRobot/pkg/sdk"
)

// 管理员添加共识节点:
// 1.根据配置文件中使用node5,node6,node7作为validators
// 2.检查stakeAccount余额，余额不足initAmount则通过管理员转账补齐，确保3个节点的总余额大于1亿(投票所需最小票额)
// 3.节点质押所有余额initAmount，两个周期后，dump stake event
// 4.检查质押后余额，应该都为0
// 5.管理员添加共识节点,等待一个周期后查询有效节点，并比较
// 6.在日志中观察挖矿节点是否包含这3个节点.
func AddValidators() (succeed bool) {
	var (
		params struct {
			InitAmount int
		}

		nodes              = config.Conf.ValidatorNodes()
		balances           = make([]int, len(nodes))
		expectNodeAddrList = make([]common.Address, len(nodes))
		err                error
	)

	if err = config.LoadParams("AddValidators.json", &params); err != nil {
		log.Error(err)
		return
	}

	// check balance before stake
	{
		for i, node := range nodes {
			data, err := admcli.BalanceOf(node.StakeAddr(), "latest")
			if err != nil {
				log.Error("failed to check %s balance", node.NodeAddr().Hex())
				return
			}
			balances[i] = int(plt.PrintUPLT(data))
			log.Infof("%s balance before stake %d", node.NodeAddr().Hex(), plt.PrintUPLT(data))
		}
	}

	// deposit and dump event log
	{
		depositHashList := make([]common.Hash, 0)
		for i, balance := range balances {
			if balance < params.InitAmount {
				addAmount := params.InitAmount - balance
				node := nodes[i]
				hash, err := admcli.PLTTransfer(node.StakeAddr(), plt.MultiPLT(addAmount))
				if err != nil {
					log.Errorf("failed to deposit to node %s, amount %d", node.NodeAddr().Hex(), addAmount)
					return
				} else {
					depositHashList = append(depositHashList, hash)
					balances[i] += addAmount
				}
			}
		}
		wait(2)
		if err := DumpHashList(depositHashList, "deposit for validator"); err != nil {
			log.Error(err)
			return
		}
	}

	// stake and dump event log
	{
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
	}

	wait(2 * config.Conf.RewardEffectivePeriod)

	// check balance after stake
	{
		for _, node := range nodes {
			data, err := admcli.BalanceOf(node.StakeAddr(), "latest")
			if err != nil {
				log.Error("failed to check %s's balance after stake, err :%v", node.NodeAddr().Hex(), err)
				return
			}
			log.Infof("%s balance after stake %d", node.NodeAddr().Hex(), plt.PrintUPLT(data))
		}
	}

	// admin add validators and dump event logs
	{
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
	}

	// check pending validators
	{
		log.Infof("check pending validators at block %d", admcli.GetBlockNumber())
		for i, node := range nodes {
			expectNodeAddrList[i] = node.NodeAddr()
		}
		validators := admcli.GetAllValidators("latest")
		if !HasAddrs(validators, expectNodeAddrList) {
			log.Error("validators not pending, check palette log")
			return
		}
		log.Infof("check pending validators success")
	}

	wait(config.Conf.RewardEffectivePeriod)

	// check effective validators
	{
		log.Infof("check effective validators at block %d", admcli.GetBlockNumber())
		effectiveValidators := admcli.GetEffectiveValidators("latest")
		if !HasAddrs(effectiveValidators, expectNodeAddrList) {
			log.Error("validators not effective, check palette log")
			return
		}
		for _, v := range effectiveValidators {
			log.Infof("add validator %s success", v.Hex())
		}
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

// 分润:
// 1.在已经完成addValidators的情况下，测试分润结果
// 2.查询分润前账户余额
// 3.等待`rewardBlocks`，比如分润周期为5，`rewardBlocks`为12，横跨2个周期，那么真实的分润应该是每个区块的分润*10，也有可能因为等待时间过程横跨3个周期。
// 4.比较expect和分润前后余额差额
func Reward() (succeed bool) {
	var (
		params struct {
			RewardBlocks                   int
			ExpectRewardPoolAmount         float64
			ExpectRewardAmountPerValidator float64
		}

		rewardPool = common.HexToAddress(config.Conf.BaseRewardPool)
		nodes      = config.Conf.ValidatorNodes()
		addrs      = append(nodes.StakeAccounts(), rewardPool)

		blkBeforeCheckReward, blkAfterCheckReward           uint64
		balancesBeforeCheckReward, balancesAfterCheckReward map[common.Address]float64

		err error
	)

	if err = config.LoadParams("Reward.json", &params); err != nil {
		log.Error(err)
		return
	}

	// check balance before reward
	{
		blkBeforeCheckReward = admcli.GetBlockNumber()
		log.Infof("check balance before testing reward at block %d", blkBeforeCheckReward)
		if balancesBeforeCheckReward, err = getBalances(addrs, BlockNumber2Hex(blkBeforeCheckReward)); err != nil {
			log.Error("failed to check balance before testing, err: %v", err)
			return
		}
	}

	// waiting for blocks
	wait(params.RewardBlocks)

	// check balance after reward
	{
		blkAfterCheckReward = blkBeforeCheckReward + uint64(params.RewardBlocks)
		log.Infof("check balance after testing reward at block %d", blkAfterCheckReward)
		if balancesAfterCheckReward, err = getBalances(addrs, BlockNumber2Hex(blkAfterCheckReward)); err != nil {
			log.Error("failed to check balance after testing, err: %v", err)
			return
		}
	}

	// check expect
	{
		res, err := subBalanceMap(balancesBeforeCheckReward, balancesAfterCheckReward)
		if err != nil {
			log.Error(err)
			return
		}
		var expect float64
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
	}

	return true
}

// 分润动作不允许外部调用，palette中通过在miner.worker中对proposer的交易进行过滤实现这一屏蔽功能
// 构造一笔reward交易，选择任意一个validator发送交易，交易内容包含一个不正确的blockNum，观察交易后
// 的blockNum是否正确
func FakeReward() (succeed bool) {
	curBlkNo := admcli.GetBlockNumber()
	blockNum := new(big.Int).SetUint64(curBlkNo + 100)
	validators := config.Conf.ValidatorNodes().Validators()
	cli := admcli.Reset(config.Conf.ValidatorNodes()[0].PrivateKey())

	if hash, err := cli.Reward(validators, blockNum); err != nil {
		log.Error(err)
		return
	} else {
		log.Infof("fake reward tx hash %s", hash.Hex())
	}

	for i := 0; i < 3; i++ {
		latestBlk, err := cli.GetRewardRecordBlock("latest")
		if err != nil {
			log.Error(err)
			return
		}

		log.Infof("current block number %d, latest record block %d", admcli.GetBlockNumber(), latestBlk)
		wait(1)
	}
	return true
}

// 节点代理用户质押:
// 1. admin给用户充值
// 2. 检查节点是否为可用节点
// 2. fans delegate
// 3. 等待一定周期查询奖励
// 4. 取消质押
// 5. 等待一定周期查询奖励应该为0
func Delegate() (succeed bool) {
	var params struct {
		Fans []struct {
			Address   string
			Amount    int
			NodeIndex int
		}
		WaitBlock int
	}

	if err := config.LoadParams("Delegate.json", &params); err != nil {
		log.Error(err)
		return
	}

	checkBalance := func(mark string) map[common.Address]float64 {
		res := make(map[common.Address]float64)
		log.Infof("check balance %s", mark)
		for _, fan := range params.Fans {
			addr := common.HexToAddress(fan.Address)
			data, err := admcli.BalanceOf(addr, "latest")
			if err != nil {
				log.Errorf("%s check balance err: %v", fan.Address, err)
				return nil
			}
			value := plt.PrintFPLT(utils.DecimalFromBigInt(data))
			res[addr] = value
			log.Infof("%s balance %s %f", addr.Hex(), mark, value)
		}
		return res
	}

	checkStkAmt := func(mark string) {
		for _, fan := range params.Fans {
			node := config.Conf.Nodes[fan.NodeIndex]
			stkAcc := common.HexToAddress(fan.Address)
			value := admcli.GetStakeAmount(node.NodeAddr(), stkAcc, "latest")
			stkAmt := plt.PrintFPLT(utils.DecimalFromBigInt(value))
			log.Infof("%s stake amount %s %f", fan.Address, mark, stkAmt)
		}
	}

	// fans transfer back to admin
	clients := make(map[string]*sdk.Client)
	for _, fan := range params.Fans {
		clients[fan.Address] = sdk.NewSender(config.Conf.Nodes[0].RPCAddr(), config.LoadAccount(fan.Address))
	}

	// admin batch transfer to fans
	{
		balanceBeforeDeposit := checkBalance("before deposit")
		log.Infof("admin deposit to fans")
		hs := make([]common.Hash, 0)
		for _, fan := range params.Fans {
			to := common.HexToAddress(fan.Address)
			curAmt := balanceBeforeDeposit[to]
			if curAmt > float64(fan.Amount) {
				continue
			}
			amt := plt.MultiPLT(fan.Amount - int(curAmt))
			if hash, err := admcli.PLTTransfer(to, amt); err != nil {
				log.Errorf("admin transfer to %s failed, err: %v", to.Hex(), err)
				return
			} else {
				hs = append(hs, hash)
			}
		}
		wait(2)
		if err := DumpHashList(hs, "admin deposit"); err != nil {
			log.Errorf("dump hash list for delegate admin transfer err, %v", err)
			return
		}
		checkBalance("after deposit")
	}

	// fans delegate
	{
		hs := make([]common.Hash, 0)
		for _, fan := range params.Fans {
			cli := clients[fan.Address]
			node := config.Conf.Nodes[fan.NodeIndex]
			amt := plt.MultiPLT(fan.Amount)
			hash, err := cli.Stake(node.NodeAddr(), node.StakeAddr(), amt, false)
			if err != nil {
				log.Errorf("%s stake to node%d failed, hash %s, err: %v", fan.Address, node.Index, hash.Hex(), err)
				return
			} else {
				log.Infof("%s stake to node%d, hash %s", fan.Address, node.Index, hash.Hex())
			}
			hs = append(hs, hash)
		}

		wait(2)

		if err := DumpHashList(hs, "delegate"); err != nil {
			log.Error(err)
			return
		}
	}

	// waiting for reward
	{
		log.Infof("waiting for delegate taking effective......")
		wait(config.Conf.RewardEffectivePeriod * 2 + 2)

		checkStkAmt("after delegate")

		logsplit()
		for i:=0;i<params.WaitBlock;i++ {
			checkBalance("after delegate")
			wait(1)
		}
	}

	// revoke delegate
	{
		logsplit()
		log.Infof("revoke delegate")
		hs := make([]common.Hash, 0)
		for _, fan := range params.Fans {
			cli := clients[fan.Address]
			node := config.Conf.Nodes[fan.NodeIndex]
			amt := plt.MultiPLT(fan.Amount)
			hash, err := cli.Stake(node.NodeAddr(), node.StakeAddr(), amt, true)
			if err != nil {
				log.Errorf("%s revoke stake failed, hash %s, err: %v", fan.Address, hash.Hex(), err)
				return
			} else {
				log.Infof("%s revoke stake, hash %s", fan.Address, hash.Hex())
			}
			hs = append(hs, hash)
		}

		wait(2)
		if err := DumpHashList(hs, "revoke delegate"); err != nil {
			log.Error(err)
			return
		}
	}

	{
		log.Infof("waiting for revoke delegate taking effective......")
		wait(config.Conf.RewardEffectivePeriod + 2)

		checkStkAmt("after revoke delegate")

		for _, fan := range params.Fans {
			node := config.Conf.Nodes[fan.NodeIndex]
			value := admcli.GetStakeAmount(node.NodeAddr(), node.StakeAddr(), "latest")
			stkAmt := plt.PrintFPLT(utils.DecimalFromBigInt(value))
			log.Infof("%s stake amount %f", fan.Address, stkAmt)
		}
		for i:=0;i<params.WaitBlock;i++ {
			checkBalance("after revoke delegate")
			wait(1)
		}
	}

	// transfer back PLT to admin
	{
		logsplit()
		log.Infof("transfer back PLT to admin account")
		hs := make([]common.Hash, 0)
		to := config.AdminAddr
		for _, fan := range params.Fans {
			cli := clients[fan.Address]
			amt := plt.MultiPLT(fan.Amount)
			if hash, err := cli.PLTTransfer(to, amt); err != nil {
				log.Errorf("admin transfer to %s failed, err: %v", to.Hex(), err)
				return
			} else {
				hs = append(hs, hash)
			}
		}
		wait(2)
		if err := DumpHashList(hs, ""); err != nil {
			log.Errorf("dump hash list for delegate admin transfer err, %v", err)
			return
		}

		checkBalance("after testing")
	}

	return true
}

func ShowDelegateAmount() (succeed bool) {
	var params []struct {
		Address   string
		NodeIndex int
	}

	if err := config.LoadParams("ShowDelegate.json", &params); err != nil {
		log.Error(err)
		return
	}

	for _, fan := range params {
		owner := common.HexToAddress(fan.Address)
		validator := config.Conf.AllNodes()[fan.NodeIndex].NodeAddr()
		data := admcli.GetStakeAmount(owner, validator, "latest")
		value := plt.PrintFPLT(utils.DecimalFromBigInt(data))
		log.Infof("fan %s stake to %s %f PLT", owner.Hex(), validator.Hex(), value)
	}

	return true
}

// 1.检查节点路径是否存在，如果是remote需要
func DelValidators() (succeed bool) {
	var (
		params struct {
			InitAmount int
		}

		nodes              = config.Conf.ValidatorNodes()
		balances           = make([]int, len(nodes))
		expectNodeAddrList = make([]common.Address, len(nodes))
		err                error
	)

	if err = config.LoadParams("AddValidators.json", &params); err != nil {
		log.Error(err)
		return
	}

	// check balance before stake
	{
		for i, node := range nodes {
			data, err := admcli.BalanceOf(node.StakeAddr(), "latest")
			if err != nil {
				log.Error("failed to check %s balance", node.NodeAddr().Hex())
				return
			}
			balances[i] = int(plt.PrintUPLT(data))
			log.Infof("%s balance before stake %d", node.NodeAddr().Hex(), plt.PrintUPLT(data))
		}
	}

	// deposit and dump event log
	{
		depositHashList := make([]common.Hash, 0)
		for i, balance := range balances {
			if balance < params.InitAmount {
				addAmount := params.InitAmount - balance
				node := nodes[i]
				hash, err := admcli.PLTTransfer(node.StakeAddr(), plt.MultiPLT(addAmount))
				if err != nil {
					log.Errorf("failed to deposit to node %s, amount %d", node.NodeAddr().Hex(), addAmount)
					return
				} else {
					depositHashList = append(depositHashList, hash)
					balances[i] += addAmount
				}
			}
		}
		wait(2)
		if err := DumpHashList(depositHashList, "deposit for validator"); err != nil {
			log.Error(err)
			return
		}
	}

	// stake and dump event log
	{
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
	}

	wait(2 * config.Conf.RewardEffectivePeriod)

	// check balance after stake
	{
		for _, node := range nodes {
			data, err := admcli.BalanceOf(node.StakeAddr(), "latest")
			if err != nil {
				log.Error("failed to check %s's balance after stake, err :%v", node.NodeAddr().Hex(), err)
				return
			}
			log.Infof("%s balance after stake %d", node.NodeAddr().Hex(), plt.PrintUPLT(data))
		}
	}

	// admin add validators and dump event logs
	{
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
	}

	// check pending validators
	{
		log.Infof("check pending validators at block %d", admcli.GetBlockNumber())
		for i, node := range nodes {
			expectNodeAddrList[i] = node.NodeAddr()
		}
		validators := admcli.GetAllValidators("latest")
		if !HasAddrs(validators, expectNodeAddrList) {
			log.Error("validators not pending, check palette log")
			return
		}
		log.Infof("check pending validators success")
	}

	wait(config.Conf.RewardEffectivePeriod)

	// check effective validators
	{
		log.Infof("check effective validators at block %d", admcli.GetBlockNumber())
		effectiveValidators := admcli.GetEffectiveValidators("latest")
		if !HasAddrs(effectiveValidators, expectNodeAddrList) {
			log.Error("validators not effective, check palette log")
			return
		}
		for _, v := range effectiveValidators {
			log.Infof("add validator %s success", v.Hex())
		}
	}

	return true
}

// 提案修改全局参数，使用配置文件中的3个validator来做这件事情，该测试方案只用于mintPrice和gasFee，rewardPeriod另外测试.
// 1.node5提案修改某个全局参数，
// 2.提案前需要检查其是否为可用的validator。
// 3.提案后根据hash查询receipt，并打印log；
// 4.voteNodeIndexList代表的validator进行投票，需要满足stakeAmount 2/3原则。
// 5.等待一个分润周期后检查proposal状态，必须为passed，并查询global params是否正常变更。
func Proposal() (succeed bool) {
	var (
		params struct {
			ProposerNodeIndex int
			ProposalType      uint8
			ProposalValue     int
			VoteNodeIndexList []int
		}

		proposalValue *big.Int
		proposerNode  *config.Node
		voteNodes     config.Nodes

		proposalID common.Address
		proposal   *governance.MethodGetProposalOutput

		err error
	)

	{
		log.Infof("check proposal params......")
		if err = config.LoadParams("Proposal.json", &params); err != nil {
			log.Error(err)
			return
		}

		// check proposal type
		if params.ProposalType == 0 || params.ProposalType > 2 {
			log.Errorf("invalid proposal type %d", params.ProposalType)
			return
		}
		// check proposal value
		if params.ProposalValue < 0 {
			log.Errorf("invalid proposal value %d", params.ProposalValue)
			return
		}
		proposalValue = plt.MultiPLT(params.ProposalValue)
	}

	// get and check validators
	{
		log.Infof("get and check validator authority......")
		nodes, err := getAndCheckValidator(append(params.VoteNodeIndexList, params.ProposerNodeIndex))
		if err != nil {
			log.Error(err)
			return
		}
		proposerNode = nodes[0]
		voteNodes = nodes[1:]
	}

	// proposer send proposal
	{
		var hash common.Hash
		log.Infof("propose new proposal......")
		proposerCli := sdk.NewSender(config.Conf.Nodes[0].RPCAddr(), proposerNode.PrivateKey())

		if hash, err = proposerCli.Propose(params.ProposalType, proposalValue); err != nil {
			log.Errorf("%s failed to propose, err %v", proposerNode.NodeAddr().Hex(), err)
			return
		}
		wait(2)

		if proposalID, proposal, err = admcli.GetProposalFromReceipt(hash); err != nil {
			log.Error(err)
			return
		} else {
			log.Infof("proposalID %s, hash %s, proposer %s, proposal type %d, value %v, end block %d",
				proposalID.Hex(), hash.Hex(), proposerNode.NodeAddr().Hex(), proposal.ProposalType, proposalValue, proposal.EndBlock.Uint64())
		}
	}

	// vote and dump hash list
	{
		log.Infof("voting......")
		voteHashList := make([]common.Hash, 0)
		var hash common.Hash
		for _, voteNode := range voteNodes {
			voteNodeCli := sdk.NewSender(config.Conf.Nodes[0].RPCAddr(), voteNode.PrivateKey())
			if hash, err = voteNodeCli.Vote(proposalID); err != nil {
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
	}

	wait(config.Conf.RewardEffectivePeriod)

	// check proposal status
	{
		if proposal, err = admcli.GetProposal(proposalID, "latest"); err != nil {
			log.Errorf("failed to get proposal, err %v", err)
			return
		}
		if proposal.Passed != true {
			log.Errorf("proposal %s should be passed", proposalID.Hex())
			return
		}
	}

	// check global params
	{
		data, err := admcli.GetGlobalParams(params.ProposalType, "latest")
		if err != nil {
			log.Error("failed to get global params, err %v", err)
			return
		}

		expect, actual := params.ProposalValue, int(plt.PrintUPLT(data))
		if expect != actual {
			log.Errorf("proposal failed to set global params, expect %d, actual %d", expect, actual)
		} else {
			log.Infof("global params changed to %d", actual)
		}
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
