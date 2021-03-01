package core

import (
	"bytes"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/contracts/native/plt"
	"github.com/ethereum/go-ethereum/contracts/native/utils"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/palettechain/onRobot/config"
	"github.com/palettechain/onRobot/pkg/eth"
	"github.com/palettechain/onRobot/pkg/log"
	"github.com/palettechain/onRobot/pkg/sdk"
	"github.com/palettechain/onRobot/pkg/shell"
)

type cliType uint8

const (
	pltCTypeCustomer cliType = iota
	pltCTypeInvoker
	pltCTypeAdmin
	pltCTypeCrossChainAdmin
	ethCTypeInvoker
	ethCTypeOwner
)

func getPaletteCli(typ cliType) (cli *sdk.Client) {
	url := config.Conf.Rpc
	switch typ {
	case pltCTypeCustomer:
		cli = sdk.NewSender(url, nil)
	case pltCTypeInvoker:
		node := config.Conf.ValidatorNodes()[0]
		cli = sdk.NewSender(url, node.PrivateKey())
	case pltCTypeAdmin:
		cli = sdk.NewSender(url, config.AdminKey)
	case pltCTypeCrossChainAdmin:
		cli = sdk.NewSender(url, config.CrossChainAdminKey)
	}
	return
}

func getEthereumCli(typ cliType) (cli *eth.EthInvoker) {
	chainID := config.Conf.CrossChain.EthereumSideChainID
	url := config.Conf.CrossChain.EthereumRPCUrl
	switch typ {
	case ethCTypeInvoker:
		if priv, err := config.Conf.CrossChain.LoadETHAccount(); err == nil {
			cli = eth.NewEInvoker(chainID, url, priv)
		} else {
			panic(fmt.Sprintf("load eth account err: %s", err.Error()))
		}
	case ethCTypeOwner:
		if priv, err := config.Conf.CrossChain.LoadETHOwner(); err == nil {
			cli = eth.NewEInvoker(chainID, url, priv)
		} else {
			panic(fmt.Sprintf("load eth owner err: %s", err.Error()))
		}
	}
	return
}

func gc() {
	//admcli = nil
	//config.Conf = config.BakConf.DeepCopy()
}

func wait(nBlock int) {
	time.Sleep(time.Duration(config.Conf.BlockPeriod) * time.Duration(nBlock))
}

func BlockNumber2Hex(data uint64) string {
	str := strconv.FormatInt(int64(data), 16)
	return "0x" + str
}

func PrivKey2Addr(pk *ecdsa.PrivateKey) common.Address {
	return crypto.PubkeyToAddress(pk.PublicKey)
}

func deployContract(abiJson, objectCode string, params ...interface{}) (common.Address, *bind.BoundContract, error) {
	node := config.Conf.ValidatorNodes()[0]
	cli := sdk.NewSender(node.RPCAddr(), node.PrivateKey())
	return cli.DeployContract(abiJson, objectCode, params...)
}

func logsplit() {
	log.Info("------------------------------------------------------------------")
}

func HasAddrs(src, dst []common.Address) bool {
	contain := func(addr common.Address, list []common.Address) bool {
		for _, v := range list {
			if addr == v {
				return true
			}
		}

		return false
	}

	for _, da := range dst {
		if !contain(da, src) {
			return false
		}
	}

	return true
}

func getBalances(cli *sdk.Client, list []common.Address, curBlkNoHex string) (map[common.Address]float64, error) {
	balancesMap := make(map[common.Address]float64)
	for _, addr := range list {
		data, err := cli.BalanceOf(addr, curBlkNoHex)
		if err != nil {
			return nil, err
		}
		balance := plt.PrintFPLT(utils.DecimalFromBigInt(data))
		balancesMap[addr] = balance
		log.Infof("%s balance %f PLT", addr.Hex(), balance)
	}
	return balancesMap, nil
}

func subBalanceMap(m1, m2 map[common.Address]float64) (map[common.Address]float64, error) {
	res := make(map[common.Address]float64)
	for addr, v1 := range m1 {
		v2, exist := m2[addr]
		if !exist {
			return nil, fmt.Errorf("missing check %s's balance after reward", addr.Hex())
		}
		res[addr] = v2 - v1
	}
	return res, nil
}

func getAndCheckValidator(cli *sdk.Client, nodeIndexList []int) (config.Nodes, error) {
	nodes := make(config.Nodes, 0)
	for _, nodeIndex := range nodeIndexList {
		node := config.Conf.GetNodeByIndex(nodeIndex)
		if node == nil {
			return nil, fmt.Errorf("failed to get validator node %d", nodeIndex)
		}
		if !cli.CheckValidator(node.NodeAddr(), "latest") {
			return nil, fmt.Errorf("%s is not valid validator", node.NodeAddr().Hex())
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

func calculateGasFee(invoker *eth.EthInvoker, gasLimit uint64) (*big.Int, error) {
	gasPrice, err := invoker.SuggestGasPrice()
	if err != nil {
		return nil, err
	}
	return utils.SafeMul(gasPrice, new(big.Int).SetUint64(gasLimit)), nil
}

func prepareEth(invoker *eth.EthInvoker, to common.Address, amount *big.Int) error {
	balanceBeforeTransfer, err := invoker.ETHBalance(to)
	if err != nil {
		return err
	}
	if balanceBeforeTransfer.Cmp(amount) >= 0 {
		return nil
	}
	if bytes.Equal(invoker.Address().Bytes(), to.Bytes()) {
		return fmt.Errorf("balance not enough")
	}
	hash, err := invoker.TransferETH(to, amount)
	if err != nil {
		return err
	}
	balanceAfterTransfer, err := invoker.ETHBalance(to)
	if err != nil {
		return err
	}
	if utils.SafeSub(balanceAfterTransfer, balanceBeforeTransfer).Cmp(amount) == 0 {
		log.Infof("prepare %s ETH %d success, tx hash %s", to.Hex(), amount, hash.Hex())
		return nil
	}
	log.Infof("prepare %s balance incorrect, balance before transfer %s, balance after transfer %s, txhash %s",
		to.Hex(),
		balanceBeforeTransfer.String(),
		balanceAfterTransfer.String(),
		hash.Hex(),
	)
	return nil
}

func prepareAllowance(invoker *eth.EthInvoker, owner, spender common.Address, amount *big.Int) error {
	asset := config.Conf.CrossChain.EthereumPLTAsset
	curAmt, err := invoker.PLTAllowance(asset, owner, spender)
	if err != nil {
		return err
	}
	if curAmt.Cmp(amount) >= 0 {
		log.Infof("(owner, spender) (%s, %s) allowance %d enough", owner.Hex(), spender.Hex(), plt.PrintUPLT(amount))
		return nil
	}

	hash, err := invoker.PLTApprove(asset, spender, amount)
	if err != nil {
		return err
	}
	log.Infof("(owner, spender) (%s, %s) approve %d success, ethereum tx hash %s", owner.Hex(), spender.Hex(), plt.PrintUPLT(amount), hash.Hex())
	return nil
}

func customLoadAccount(addr common.Address) *ecdsa.PrivateKey {
	acc, err := config.LoadPaletteAccount(addr)
	if err == nil {
		return acc
	}

	for _, node := range config.Conf.Nodes {
		if bytes.Equal(node.NodeAddr().Bytes(), addr.Bytes()) {
			return node.PrivateKey()
		}
		if bytes.Equal(node.StakeAddr().Bytes(), addr.Bytes()) {
			return node.StakePrivateKey()
		}
	}

	return nil
}

///////////////////////////////////////////////////////////////////////////////////////
//
// exec shell scripts
//
///////////////////////////////////////////////////////////////////////////////////////

const (
	shGrep        = "grep.sh"
	shRemoteSetup = "remote_setup.sh"
	shRemoteBuild = "remote_build.sh"
	shInit        = "init_node.sh"
	shStartNode   = "start_node.sh"
	shStopNode    = "stop_node.sh"
	shClearNode   = "clear_node.sh"
)

// args:
// isRemote=$0
// currentIp=$1
func execGrep() {
	args := make([]string, 3)
	args[0] = "false"
	args[1] = "127.0.0.1"
	args[2] = config.Conf.Environment.SSHPort

	if !config.Conf.Environment.Remote {
		shell.Exec(shGrep, args...)
		return
	}

	iplist := config.Conf.IpList()
	args[0] = "true"
	for _, ip := range iplist {
		args[1] = ip
		shell.Exec(shGrep, args...)
	}
}

// args:
// localWorkspace=$0;
// remoteWorkspace=$1;
// currentIp=$2;
func execRemoteSetup() {
	if !config.Conf.Environment.Remote {
		return
	}

	args := make([]string, 4)
	args[0] = config.Conf.Environment.WorkSpace()
	args[1] = config.Conf.Environment.RemoteWorkspace
	iplist := config.Conf.IpList()
	args[3] = config.Conf.Environment.SSHPort
	for _, ip := range iplist {
		args[2] = ip
		shell.Exec(shRemoteSetup, args...)
	}
}

// args:
// currentIp=$0;
func execRemoteBuild() {
	if !config.Conf.Environment.Remote {
		return
	}

	iplist := config.Conf.IpList()
	port := config.Conf.Environment.SSHPort
	gopath := config.Conf.Environment.RemoteGoPath
	for _, ip := range iplist {
		shell.Exec(shRemoteBuild, ip, port, gopath)
	}
}

// args:
// nodeIdx=$0;
// isRemote=$1;
// workspace=$2;
// currentIp=$3;
func execInitNode(node *config.Node) {
	args := make([]string, 5)
	args[0] = fmt.Sprintf("%d", node.Index)
	args[1] = "false"
	args[2] = config.Conf.Environment.WorkSpace()
	args[3] = "127.0.0.1"
	args[4] = config.Conf.Environment.SSHPort

	if config.Conf.Environment.Remote {
		args[1] = "true"
		args[2] = config.Conf.Environment.RemoteWorkspace
		args[3] = node.Host
	}

	shell.Exec(shInit, args...)
}

// args:
// isRemote=$1;
// logLevel=$2;
// networkID=$3;
// currentIp=$4;
// nodeIndex=$5;
// nodeDir=$6;
// rpcPort=$7;
// p2pPort=$8;
func execStartNode(node *config.Node) {
	args := make([]string, 9)
	args[0] = "false"
	args[1] = fmt.Sprintf("%d", config.Conf.Environment.LogLevel)
	args[2] = fmt.Sprintf("%d", config.Conf.Environment.NetworkID)
	args[3] = "127.0.0.1"
	args[4] = fmt.Sprintf("%d", node.Index)
	args[5] = node.NodeDirPath()
	args[6] = node.RPCPort
	args[7] = node.P2PPort
	args[8] = config.Conf.Environment.SSHPort

	if config.Conf.Environment.Remote {
		args[0] = "true"
		args[3] = node.Host
	}

	shell.Exec(shStartNode, args...)
}

// args:
// isRemote=$0;
// nodeIdx=$1;
// p2pPort=$2;
// currentIp=$3;
func execStopNode(node *config.Node) {
	args := make([]string, 4)
	args[0] = "false"
	args[1] = fmt.Sprintf("%d", node.Index)
	args[2] = "127.0.0.1"
	args[3] = config.Conf.Environment.SSHPort

	if config.Conf.Environment.Remote {
		args[0] = "true"
		args[2] = node.Host
	}

	shell.Exec(shStopNode, args...)
}

// args:
// isRemote=$1;
// nodeIdx=$2;
// workspace=$3;
// currentIp=$4;
func execClearNode(node *config.Node) {
	args := make([]string, 5)
	args[0] = "false"
	args[1] = fmt.Sprintf("%d", node.Index)
	args[2] = config.Conf.Environment.WorkSpace()
	args[3] = "127.0.0.1"
	args[4] = config.Conf.Environment.SSHPort

	if config.Conf.Environment.Remote {
		args[0] = "true"
		args[2] = config.Conf.Environment.RemoteWorkspace
		args[3] = node.Host
	}

	shell.Exec(shClearNode, args...)
}
