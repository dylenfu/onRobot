package config

import (
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"sort"
	"sync"
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
	AdminKey      *ecdsa.PrivateKey
	AdminAddr     common.Address
)

type Config struct {
	Environment           *Env
	BaseRPCUrl            string
	DefaultPassphrase     string
	AdminAccount          string
	BaseRewardPool        string
	Accounts              []string
	GasLimit              uint64
	DeployGasLimit        uint64
	BlockPeriod           encode.Duration
	RewardEffectivePeriod int // 区块奖励周期/参数生效周期
	Nodes                 []*Node
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
	Index        int    `json:"Index"`
	Address      string `json:"Address"`
	NodeKey      string `json:"NodeKey"`
	StakeAccount string `json:"StakeAccount"`

	once       sync.Once
	ndpk, sapk *ecdsa.PrivateKey
}

func (n *Node) init() {

	// load node private key
	bz, err := hex.DecodeString(n.NodeKey)
	if err != nil {
		panic(err)
	}
	if pk, err := crypto.ToECDSA(bz); err != nil {
		panic(err)
	} else {
		n.ndpk = pk
	}

	// load node stake account private key
	file := path.Join(Conf.Environment.Workspace, keystoreDir, n.StakeAccount)
	if bz, err = ioutil.ReadFile(file); err != nil {
		panic(fmt.Sprintf("load keystore err %v", err))
	}
	if ks, err := keystore.DecryptKey(bz, Conf.DefaultPassphrase); err != nil {
		panic(fmt.Sprintf("decrypt key %s err %v", n.StakeAccount, err))
	} else {
		n.sapk = ks.PrivateKey
	}
}

func (n *Node) PrivateKey() *ecdsa.PrivateKey {
	n.once.Do(n.init)
	return n.ndpk
}

func (n *Node) NodeAddr() common.Address {
	return common.HexToAddress(n.Address)
}

func (n *Node) StakePrivateKey() *ecdsa.PrivateKey {
	n.once.Do(n.init)
	return n.sapk
}

func (n *Node) StakeAddr() common.Address {
	n.once.Do(n.init)
	return common.HexToAddress(n.StakeAccount)
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
	AdminAddr = crypto.PubkeyToAddress(AdminKey.PublicKey)
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

func LoadAccount(keyhex string) *ecdsa.PrivateKey {
	filepath := files.FullPath(Conf.Environment.Workspace, keystoreDir, keyhex)
	keyJson, err := ioutil.ReadFile(filepath)
	if err != nil {
		panic(fmt.Errorf("failed to read file: [%v]", err))
	}

	key, err := keystore.DecryptKey(keyJson, Conf.DefaultPassphrase)
	if err != nil {
		panic(fmt.Errorf("failed to decrypt keyjson: [%v]", err))
	}

	return key.PrivateKey
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
