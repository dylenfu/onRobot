package core

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/contracts/native/utils"
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/contracts/native/plt"
	"github.com/palettechain/onRobot/config"
	"github.com/palettechain/onRobot/pkg/log"
	"github.com/palettechain/onRobot/pkg/sdk"
)

func BlockNumber() bool {
	blockNumber := admcli.GetBlockNumber()
	log.Infof("current block number %d", blockNumber)
	return true
}

func Nonce() (succeed bool) {
	var params struct {
		Address string
	}
	if err := config.LoadParams("GetNonce.json", &params); err != nil {
		log.Error(err)
		return
	}

	nonce := admcli.GetNonce(params.Address)
	log.Infof("%s nonce is %d", params.Address, nonce)
	return true
}

// 检查数据一致性(重要):
// 轮询N个节点，比较其查询所得的lastRewardBlock是否一致。非验证节点同步速度可能会慢上几个块.
func Consistency() (succeed bool) {
	var params struct {
		NodeIndex []int
	}
	if err := config.LoadParams("Consistency.json", &params); err != nil {
		log.Error(err)
		return
	}

	nodes := make([]*config.Node, 0)
	for _, v := range params.NodeIndex {
		node := config.Conf.Nodes[v]
		nodes = append(nodes, node)
	}

	clients := make([]*sdk.Client, len(nodes))
	for i, node := range nodes {
		clients[i] = sdk.NewSender(node.RPCAddr(), config.AdminKey)
	}

	currentBlkNo := clients[0].GetBlockNumber() - 10

	var i, blkNo uint64 = 0, 10
	for i = currentBlkNo - blkNo; i < currentBlkNo; i++ {
		lastRdBlk := big.NewInt(0)
		lastRdProposer := common.Address{}
		queryBlkHex := "0x" + strconv.FormatInt(int64(i), 16)
		for j, client := range clients {
			rdBlk, err := client.GetRewardRecordBlock(queryBlkHex)
			if err != nil {
				log.Error(err)
				return
			}
			rdProp, err := client.GetLatestRewardProposer(queryBlkHex)
			if err != nil {
				log.Error(err)
				return
			}

			if j == 0 {
				lastRdBlk = rdBlk
				lastRdProposer = rdProp
				continue
			}
			if lastRdBlk.Cmp(rdBlk) != 0 {
				log.Errorf("%s query result %d, %s query result %d", clients[0].Url(), lastRdBlk.Uint64(), clients[i].Url(), rdBlk.Uint64())
			}
			if lastRdProposer != rdProp {
				log.Errorf("%s query result %s, %s query result %s", clients[0].Url(), lastRdProposer.Hex(), clients[i].Url(), rdProp.Hex())
			}
		}
		log.Infof("last reward block %d, last reward proposer %s", lastRdBlk.Uint64(), lastRdProposer.Hex())
	}

	return true
}

// 准备测试需要的一定量PLT
func Deposit() (succeed bool) {
	var params struct {
		Amount float64
	}
	if err := config.LoadParams("Deposit.json", &params); err != nil {
		log.Error(err)
		return
	}

	amount := plt.MultiFloatPLT(params.Amount)
	accounts := config.Conf.Accounts
	hashList := make([]common.Hash, len(accounts))
	for i, acc := range accounts {
		to := common.HexToAddress(acc)
		hash, err := admcli.PLTTransfer(to, amount.BigInt())
		if err != nil {
			log.Errorf("failed to deposit to %s, amount %f, err: %v", to.Hex(), plt.PrintFPLT(amount), err)
			return
		}
		hashList[i] = hash
	}

	wait(2)

	if err := DumpHashList(hashList, "deposit"); err != nil {
		return
	}

	return true
}

type EVMTestOps struct {
	Object    string `json:"object"`
	Opcodes   string `json:"opcodes"`
	SourceMap string `json:"sourceMap"`
	ABI       string `json:"abi"`
	Address   string `json:"address"`
}

// 只有validator拥有部署solidity合约的权限，在调用该方法前，先调用addValidators
func Deploy() (succeed bool) {
	var params EVMTestOps

	if err := config.LoadParams("evm.json", &params); err != nil {
		log.Error(err)
		return
	}

	node := config.Conf.ValidatorNodes()[0]
	cli := sdk.NewSender(node.RPCAddr(), node.PrivateKey())
	// send transaction
	if err := cli.DeployContract(params.ABI, params.Object); err != nil {
		log.Errorf("failed to deploy contract, err: %v", err)
		return
	}

	return true
}

//nativeTransfer(address _to, uint _value)
func EVM() (succeed bool) {
	var params EVMTestOps

	if err := config.LoadParams("evm.json", &params); err != nil {
		log.Error(err)
		return
	}

	type TransferInput struct {
		To common.Address
		Value   *big.Int
	}

	abiJs, err := abi.JSON(strings.NewReader(params.ABI))
	if err != nil {
		log.Errorf("failed to read abj json string, err: %v", err)
		return
	}

	to := common.HexToAddress("0xecce5f1346afee82990cccc52fe521005bd54ff0")
	contract := common.HexToAddress(params.Address)
	amount := plt.MultiPLT(1)

	// transfer plt to contract
	{
		if _, err := admcli.PLTTransfer(contract, amount); err != nil {
			log.Errorf("failed to transfer PLT to contract, err: %v", err)
			return
		}
		wait(1)
		balance, err := admcli.BalanceOf(contract, "latest")
		if err != nil {
			log.Errorf("failed to get balance of contract, err: %v", err)
			return
		}
		log.Infof("contract %s balance %d", contract.Hex(), plt.PrintUPLT(balance))
	}

	b1, err := admcli.BalanceOf(to, "latest")
	if err != nil {
		log.Errorf("failed to get balance before transfer, err: %v", err)
		return
	}

	enc, err := utils.PackMethod(abiJs, "nativeTransfer",  to, amount)
	if err != nil {
		log.Errorf("failed to pack `nativeTransfer`, err: %v", err)
		return
	}

	hash, err := admcli.SendTransaction(contract, enc)
	if err != nil {
		log.Errorf("failed to send transaction to new deployed contract, err: %v", err)
		return
	} else {
		log.Infof("send tx %s success", hash.Hex())
	}

	wait(2)

	if err := admcli.DumpEventLog(hash); err != nil {
		log.Errorf("failed to dump tx %s, err: %v", hash.Hex(), err)
		return
	}

	b2, err := admcli.BalanceOf(to, "latest")
	if err != nil {
		log.Errorf("failed to get balance after transfer, err: %v", err)
		return
	}

	if utils.SafeSub(b2, b1).Cmp(amount) != 0 {
		log.Errorf("balance before transfer %d, balance after transfer %d, amount %d is not correct",
			plt.PrintUPLT(b1), plt.PrintUPLT(b2), plt.PrintUPLT(amount))
	}

	return true
}

func RemoteBuild() (succeed bool) {
	execRemoteBuild()
	return true
}

func RemoteSetup() (succeed bool) {
	execRemoteSetup()
	return true
}
