package core

import (
	"crypto/ecdsa"
	"fmt"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/palettechain/onRobot/config"
	"github.com/palettechain/onRobot/pkg/log"
	"github.com/palettechain/onRobot/pkg/sdk"
	"github.com/palettechain/onRobot/pkg/shell"
)

var admcli *sdk.Client

func initialize() {
	baseUrl := config.Conf.Nodes[0].RPCAddr()
	key := config.AdminKey
	admcli = sdk.NewSender(baseUrl, key)
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

func DumpHashList(hashlist []common.Hash, mark string) error {
	for _, hash := range hashlist {
		if err := admcli.DumpEventLog(hash); err != nil {
			log.Errorf("failed to dump receipt, hash %s, [%v]", hash.Hex(), err)
			return err
		}
	}
	log.Infof("dump %s event log success", mark)
	log.Info("------------------------------------------------------------------")
	return nil
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
	args := make([]string, 2)
	args[0] = "false"
	args[1] = "127.0.0.1"

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

	args := make([]string, 3)
	args[0] = config.Conf.Environment.LocalWorkspace
	args[1] = config.Conf.Environment.RemoteWorkspace
	iplist := config.Conf.IpList()
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
	for _, ip := range iplist {
		shell.Exec(shRemoteBuild, ip)
	}
}

// args:
// nodeIdx=$0;
// isRemote=$1;
// workspace=$2;
// currentIp=$3;
func execInitNode(node *config.Node) {
	args := make([]string, 4)
	args[0] = fmt.Sprintf("%d", node.Index)
	args[1] = "false"
	args[2] = config.Conf.Environment.LocalWorkspace
	args[3] = "127.0.0.1"

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
	args := make([]string, 8)
	args[0] = "false"
	args[1] = fmt.Sprintf("%d", config.Conf.Environment.LogLevel)
	args[2] = fmt.Sprintf("%d", config.Conf.Environment.NetworkID)
	args[3] = "127.0.0.1"
	args[4] = fmt.Sprintf("%d", node.Index)
	args[5] = node.NodeDir
	args[6] = node.RPCPort
	args[7] = node.P2PPort

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
	args := make([]string, 3)
	args[0] = "false"
	args[1] = fmt.Sprintf("%d", node.Index)
	args[2] = "127.0.0.1"

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
	args := make([]string, 4)
	args[0] = "false"
	args[1] = fmt.Sprintf("%d", node.Index)
	args[2] = config.Conf.Environment.LocalWorkspace
	args[3] = "127.0.0.1"

	if config.Conf.Environment.Remote {
		args[0] = "true"
		args[2] = config.Conf.Environment.RemoteWorkspace
		args[3] = node.Host
	}

	shell.Exec(shClearNode, args...)
}
