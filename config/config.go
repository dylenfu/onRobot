package config

import (
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sort"
	"time"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/palettechain/onRobot/pkg/encode"
	"github.com/palettechain/onRobot/pkg/files"
	"github.com/palettechain/onRobot/pkg/sdk"
)

const (
	testCaseDir = "cases"
	keystoreDir = "keystore"
	setupDir    = "setup"
)

var (
	Conf, BakConf = new(Config), new(Config)
	AdminKey      *keystore.Key
)

type Config struct {
	Environment       *Env
	DefaultPassphrase string
	AdminAccount      string
	Accounts          []string
	GasLimit          uint64
	DeployGasLimit    uint64
	BlockPeriod       encode.Duration
	EffectivePeriod   int // 区块奖励周期/参数生效周期
	Nodes             []*Node
}

func (c *Config) DeepCopy() *Config {
	cp := new(Config)
	enc, err := json.Marshal(c)
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(enc, cp); err != nil {
		panic(err)
	}
	return cp
}

func (c *Config) AllNodeAddressList() []string {
	list := make([]string, len(c.Nodes))
	for i, node := range c.Nodes {
		list[i] = node.Address
	}
	return list
}

func (c *Config) ResetEnv(nodeIdxStart, nodeNum int) {
	c.Environment.NodeIdxStart = nodeIdxStart
	c.Environment.NodeNum = nodeNum
}

type Node struct {
	Index      int
	Address    string
	Nodekey    string
}

func (n *Node) PrivateKey() *ecdsa.PrivateKey{
	key, err := hex.DecodeString(n.Nodekey)
	if err != nil {
		panic(err)
	}
	privKey, err := crypto.ToECDSA(key)
	if err != nil {
		panic(err)
	}
	return privKey
}

func (n *Node) Addr() common.Address {
	return common.HexToAddress(n.Address)
}

type Env struct {
	Workspace    string
	NodeIdxStart int
	NodeNum      int
	NetworkID    int
	StartRPCPort int
	StartP2PPort int
	LogLevel     int
}

func Init(path string) {
	if err := LoadConfig(path, Conf); err != nil {
		panic(err)
	}

	// sort nodes with node index
	sort.Slice(Conf.Nodes, func(i, j int) bool {
		return Conf.Nodes[i].Index < Conf.Nodes[j].Index
	})

	// load nodes privateKey
	sdk.Init(Conf.GasLimit, Conf.DeployGasLimit, time.Duration(Conf.BlockPeriod))

	AdminKey = LoadAccount(Conf.AdminAccount)
	BakConf = Conf.DeepCopy()
}

func LoadConfig(filepath string, ins interface{}) error {
	data, err := files.ReadFile(filepath)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, ins)
	if err != nil {
		return fmt.Errorf("json.Unmarshal TestConfig:%s error:%s", data, err)
	}
	return nil
}

func LoadParams(fileName string, data interface{}) error {
	filePath := files.FullPath(Conf.Environment.Workspace, testCaseDir, fileName)
	bz, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	return json.Unmarshal(bz, data)
}

func LoadAccount(keyhex string) *keystore.Key {
	filepath := files.FullPath(Conf.Environment.Workspace, keystoreDir, keyhex)
	keyJson, err := ioutil.ReadFile(filepath)
	if err != nil {
		panic(fmt.Errorf("failed to read file: [%v]", err))
	}

	key, err := keystore.DecryptKey(keyJson, Conf.DefaultPassphrase)
	if err != nil {
		panic(fmt.Errorf("failed to decrypt keyjson: [%v]", err))
	}

	return key
}

func ShellPath(fileName string) string {
	return files.FullPath(Conf.Environment.Workspace, "", fileName)
}

func GenesisNodeNumber() int {
	filepath := files.FullPath(Conf.Environment.Workspace, setupDir, "static-nodes.json")
	keyJson, err := ioutil.ReadFile(filepath)
	if err != nil {
		panic(fmt.Errorf("failed to read file: [%v]", err))
	}

	var nodes []string
	if err := json.Unmarshal(keyJson, &nodes); err != nil {
		panic(fmt.Errorf("failed to unmarshal static-nodes.json: [%v]", err))
	}

	return len(nodes)
}
