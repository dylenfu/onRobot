package core

import (
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/contracts/native/plt"
	"github.com/ethereum/go-ethereum/contracts/native/utils"
	"github.com/palettechain/onRobot/config"
	"github.com/palettechain/onRobot/pkg/log"
	"github.com/palettechain/onRobot/pkg/sdk"
)

func Demo() bool {
	//if err := config.SaveConfig(config.Conf); err != nil {
	//	log.Error(err)
	//	return false
	//}
	log.Info("it is a demo")
	return true
}

func BlockNumber() bool {
	cli := getPaletteCli(pltCTypeCustomer)
	blockNumber := cli.GetBlockNumber()
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

	cli := getPaletteCli(pltCTypeCustomer)
	nonce := cli.GetNonce(params.Address)
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
	admcli := getPaletteCli(pltCTypeAdmin)
	for _, to := range accounts {
		if _, err := admcli.PLTTransfer(to, amount.BigInt()); err != nil {
			log.Errorf("failed to deposit to %s, amount %f, err: %v", to.Hex(), plt.PrintFPLT(amount), err)
			return
		}
	}
	return true
}

type EVMTestOps struct {
	Object  string `json:"object"`
	ABI     string `json:"abi"`
	Address string `json:"address"`
}

// 只有validator拥有部署solidity合约的权限，在调用该方法前，先调用addValidators
func DeploySafetyContract() (succeed bool) {
	var params EVMTestOps

	if err := config.LoadParams("evm1.json", &params); err != nil {
		log.Error(err)
		return
	}

	if _, _, err := deployContract(params.ABI, params.Object); err != nil {
		log.Errorf("failed to deploy contract, err: %v", err)
		return
	}

	return true
}

func Safety() (succeed bool) {
	var params EVMTestOps

	logsplit()
	log.Info("prepare to safety test, load safety contract abi...")
	if err := config.LoadParams("evm1.json", &params); err != nil {
		log.Error(err)
		return
	}

	type TransferInput struct {
		To    common.Address
		Value *big.Int
	}

	abiJs, err := abi.JSON(strings.NewReader(params.ABI))
	if err != nil {
		log.Errorf("failed to read abj json string, err: %v", err)
		return
	}

	to := common.HexToAddress("0xecce5f1346afee82990cccc52fe521005bd54ff0")
	contract := common.HexToAddress(params.Address)
	amount := plt.MultiPLT(1)
	admcli := getPaletteCli(pltCTypeAdmin)

	//
	logsplit()
	log.Infof("prepare balance, admin account transfer %d PLT to delegate caller(safety contract)...", plt.PrintUPLT(amount))
	{
		if _, err := admcli.PLTTransfer(contract, amount); err != nil {
			log.Errorf("failed to transfer PLT to contract, err: %v", err)
			return
		}
		balance, err := admcli.BalanceOf(contract, "latest")
		if err != nil {
			log.Errorf("failed to get balance of contract, err: %v", err)
			return
		}
		log.Infof("admin account transfer prepare PLT to contract success! and contract %s balance is %d", contract.Hex(), plt.PrintUPLT(balance))
	}
	wait(2)

	logsplit()
	log.Info("check balance before safety test...")
	dstBalanceBeforeTest, err := admcli.BalanceOf(to, "latest")
	if err != nil {
		log.Errorf("failed to get dest account balance before transfer, err: %v", err)
		return
	} else {
		log.Infof("dest account balance before safety test %d", plt.PrintUPLT(dstBalanceBeforeTest))
	}
	admBalanceBeforeTest, err := admcli.BalanceOf(admcli.Address(), "latest")
	if err != nil {
		log.Errorf("failed to get admin balance before transfer, err: %v", err)
		return
	} else {
		log.Infof("admin balance before safety test %d", plt.PrintUPLT(admBalanceBeforeTest))
	}
	contractBalanceBeforeTest, err := admcli.BalanceOf(contract, "latest")
	if err != nil {
		log.Errorf("failed to get contract balance before transfer, err: %v", err)
	} else {
		log.Infof("contract balance before safety test %d", plt.PrintUPLT(contractBalanceBeforeTest))
	}

	//
	logsplit()
	log.Info("safety testing, try to call nativeTransfer which will delegate call transfer method of plt contract...")
	enc, err := utils.PackMethod(abiJs, "nativeTransfer", to, amount)
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
	if err := admcli.WaitTransaction(hash); err != nil {
		log.Error(err)
		return
	}

	logsplit()
	log.Info("check balance after test...")
	dstBalanceAfterTest, err := admcli.BalanceOf(to, "latest")
	if err != nil {
		log.Errorf("failed to get dest account balance after transfer, err: %v", err)
		return
	} else {
		log.Infof("dest account balance after test %d", plt.PrintUPLT(dstBalanceAfterTest))
	}
	admBalanceAfterTest, err := admcli.BalanceOf(admcli.Address(), "latest")
	if err != nil {
		log.Errorf("failed to get admin balance after transfer, err: %v", err)
		return
	} else {
		log.Infof("admin balance after test %d", plt.PrintUPLT(admBalanceAfterTest))
	}
	contractBalanceAfterTest, err := admcli.BalanceOf(contract, "latest")
	if err != nil {
		log.Errorf("failed to get contract balance after transfer, err: %v", err)
		return
	} else {
		log.Infof("contract balance after test %d", plt.PrintUPLT(contractBalanceAfterTest))
	}

	// dest account should received enough PLT
	if utils.SafeSub(dstBalanceAfterTest, dstBalanceBeforeTest).Cmp(amount) != 0 {
		log.Errorf("dest account balance before test %d, balance after test %d, amount %d is not correct",
			plt.PrintUPLT(dstBalanceBeforeTest), plt.PrintUPLT(dstBalanceAfterTest), plt.PrintUPLT(amount))
		return
	} else {
		log.Info("dest account balance is correct")
	}
	// admin account balance should be less
	if utils.SafeSub(admBalanceBeforeTest, admBalanceAfterTest).Cmp(amount) != 0 {
		log.Errorf("admin account balance before transfer %d, balance after test %d",
			plt.PrintUPLT(admBalanceBeforeTest), plt.PrintUPLT(admBalanceAfterTest))
	} else {
		log.Info("admin account balance is correct")
	}
	if contractBalanceBeforeTest.Cmp(contractBalanceAfterTest) != 0 {
		log.Errorf("contract account balance before test %d, balance after test %d",
			plt.PrintUPLT(contractBalanceBeforeTest), plt.PrintUPLT(contractBalanceAfterTest))
	} else {
		log.Info("contract account balance is correct")
	}

	return true
}

func TestDeploy2() (succeed bool) {
	var params EVMTestOps

	if err := config.LoadParams("evm2.json", &params); err != nil {
		log.Error(err)
		return
	}

	if _, _, err := deployContract(params.ABI, params.Object); err != nil {
		log.Errorf("failed to deploy contract, err: %v", err)
		return
	}

	return true
}

//nativeTransfer(address _to, uint _value)
func TestEVM2() (succeed bool) {
	var params EVMTestOps

	if err := config.LoadParams("evm2.json", &params); err != nil {
		log.Error(err)
		return
	}

	abiJs, err := abi.JSON(strings.NewReader(params.ABI))
	if err != nil {
		log.Errorf("failed to read abj json string, err: %v", err)
		return
	}
	contract := common.HexToAddress(params.Address)
	enc, err := utils.PackMethod(abiJs, "simpleRequire")
	if err != nil {
		log.Errorf("failed to pack `nativeTransfer`, err: %v", err)
		return
	}

	admcli := getPaletteCli(pltCTypeAdmin)
	hash, err := admcli.SendTransaction(contract, enc)
	if err != nil {
		log.Errorf("failed to send transaction to new deployed contract, err: %v", err)
		return
	} else {
		log.Infof("send tx %s success", hash.Hex())
	}

	return true
}

func DumpBlock() (succeed bool) {
	var params struct {
		Blocks []uint64
	}
	if err := config.LoadParams("DumpBlock.json", &params); err != nil {
		log.Error(err)
		return
	}

	cli := getPaletteCli(pltCTypeCustomer)
	for _, block := range params.Blocks {
		_ = cli.DumpBlock(block)
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
