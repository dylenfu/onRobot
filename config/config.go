package config

import (
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/palettechain/onRobot/pkg/encode"
	"github.com/palettechain/onRobot/pkg/files"
	"github.com/palettechain/onRobot/pkg/sdk"
	polysdk "github.com/polynetwork/poly-go-sdk"
)

const (
	testCaseDir     = "cases"
	keystoreDir     = "keystore"
	setupDir        = "setup"
	polyKeystoreDir = "poly_keystore"
	ethKeystoreDir  = "eth_keystore"
)

var (
	Conf, BakConf  = new(Config), new(Config)
	AdminKey       *ecdsa.PrivateKey
	AdminAddr      common.Address
	ConfigFilePath string
)

type Config struct {
	Environment           *Env
	Network               *Network
	DefaultPassphrase     string
	AdminAccount          string
	BaseRewardPool        string
	Accounts              []string
	GasLimit              uint64
	DeployGasLimit        uint64
	BlockPeriod           encode.Duration
	RewardEffectivePeriod int // 区块奖励周期/参数生效周期
	Nodes                 []*Node
	CrossChain            *CrossChainConfig
	FinalOwner            *FinalOwner
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

func (c *Config) IpList() []string {
	data := make(map[string]struct{})
	for _, v := range c.Nodes {
		data[v.Host] = struct{}{}
	}

	list := make([]string, 0)
	for host, _ := range data {
		list = append(list, host)
	}

	return list
}

func (c *Config) GenesisNodes() Nodes {
	start := c.Network.NodeIndexStart
	end := start + c.Network.GenesisNodeNumber - 1
	return c.getRangeNodes(start, end)
}

func (c *Config) ValidatorNodes() Nodes {
	genesisStart := c.Network.NodeIndexStart
	genesisEnd := genesisStart + c.Network.GenesisNodeNumber - 1

	start := genesisEnd + 1
	end := start + c.Network.ValidatorsNumber - 1
	return c.getRangeNodes(start, end)
}

func (c *Config) AllNodes() Nodes {
	start := c.Network.NodeIndexStart
	num := c.Network.GenesisNodeNumber + c.Network.ValidatorsNumber
	end := start + num - 1
	return c.getRangeNodes(start, end)
}

func (c *Config) SpareNodes() Nodes {
	end := c.Network.GenesisNodeNumber + c.Network.ValidatorsNumber
	return c.Nodes[end:]
}

func (c *Config) GetNodeByIndex(index int) *Node {
	for _, n := range c.Nodes {
		if n.Index == index {
			return n
		}
	}
	return nil
}

func (c *Config) getRangeNodes(start, end int) Nodes {
	list := make([]*Node, 0)
	for i := start; i <= end; i++ {
		list = append(list, c.Nodes[i])
	}
	return list
}

type Nodes []*Node

func (n Nodes) Validators() []common.Address {
	list := make([]common.Address, len(n))
	for i, node := range n {
		list[i] = node.NodeAddr()
	}
	return list
}

func (n Nodes) StakeAccounts() []common.Address {
	list := make([]common.Address, len(n))
	for i, node := range n {
		list[i] = node.StakeAddr()
	}
	return list
}

type Node struct {
	Index        int    `json:"Index"`
	Address      string `json:"Address"`
	NodeKey      string `json:"NodeKey"`
	StakeAccount string `json:"StakeAccount"`
	Host         string `json:"Host"`
	RPCPort      string `json:"RPCPort"`
	P2PPort      string `json:"P2PPort"`

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
	file := path.Join(Conf.Environment.LocalWorkspace, keystoreDir, n.StakeAccount)
	if bz, err = ioutil.ReadFile(file); err != nil {
		panic(fmt.Sprintf("load keystore err %v", err))
	}
	if ks, err := keystore.DecryptKey(bz, Conf.DefaultPassphrase); err != nil {
		panic(fmt.Sprintf("decrypt key %s err %v", n.StakeAccount, err))
	} else {
		n.sapk = ks.PrivateKey
	}
}

func (n *Node) NodeDirPath() string {
	n.once.Do(n.init)
	data := fmt.Sprintf("node%d", n.Index)
	nodedir := path.Join(Conf.Environment.LocalWorkspace, data)
	if Conf.Environment.Remote {
		nodedir = path.Join(Conf.Environment.RemoteWorkspace, data)
	}
	return nodedir
}

func (n *Node) PrivateKey() *ecdsa.PrivateKey {
	n.once.Do(n.init)
	return n.ndpk
}

func (n *Node) NodeAddr() common.Address {
	n.once.Do(n.init)
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

func (n *Node) RPCAddr() string {
	n.once.Do(n.init)
	return fmt.Sprintf("http://%s:%s", n.Host, n.RPCPort)
}

type Env struct {
	Remote          bool
	LocalWorkspace  string
	RemoteWorkspace string
	NetworkID       int
	LogLevel        int
	IpList          []string
	SSHPort         string
	RemoteGoPath    string
	NFTServer       string
}

type Network struct {
	NodeIndexStart    int
	GenesisNodeNumber int
	ValidatorsNumber  int
}

func Init(path string) {
	ConfigFilePath = path
	err := LoadConfig(ConfigFilePath, Conf)
	if err != nil {
		panic(err)
	}

	// sort nodes with node index
	sort.Slice(Conf.Nodes, func(i, j int) bool {
		return Conf.Nodes[i].Index < Conf.Nodes[j].Index
	})

	// load nodes privateKey
	sdk.Init(Conf.GasLimit, Conf.DeployGasLimit, time.Duration(Conf.BlockPeriod))

	AdminKey, err = LoadAccount(Conf.AdminAccount)
	if err != nil {
		panic(err)
	}
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

func SaveConfig(c *Config) error {
	type XCrossChainConfig struct {
		// poly account and node rpc url
		PolyAccountDefaultPassphrase string
		PolyRPCAddress               string

		// poly side chain configuration
		PaletteSideChainID   uint64
		PaletteSideChainName string
		PaletteECCD          string
		PaletteECCM          string
		PaletteCCMP          string
		PaletteNFTProxy      string

		// ethereum side chain configuration
		EthereumSideChainID   uint64
		EthereumSideChainName string
		EthereumECCD          string
		EthereumECCM          string
		EthereumCCMP          string
		EthereumPLTAsset      string
		EthereumPLTProxy      string
		EthereumNFTProxy      string

		// ethereum node rpc and account
		EthereumRPCUrl          string
		EthereumAccount         string
		EthereumAccountPassword string
		EthereumOwner           string
		EthereumOwnerPassword   string
	}

	type XConfig struct {
		Environment           *Env
		Network               *Network
		DefaultPassphrase     string
		AdminAccount          string
		BaseRewardPool        string
		Accounts              []string
		GasLimit              uint64
		DeployGasLimit        uint64
		BlockPeriod           encode.Duration
		RewardEffectivePeriod int // 区块奖励周期/参数生效周期
		Nodes                 []*Node
		CrossChain            *XCrossChainConfig
		FinalOwner            *FinalOwner
	}

	x := new(XConfig)
	x.Environment = c.Environment
	x.Network = c.Network
	x.DefaultPassphrase = c.DefaultPassphrase
	x.AdminAccount = c.AdminAccount
	x.BaseRewardPool = c.BaseRewardPool
	x.Accounts = c.Accounts
	x.GasLimit = c.GasLimit
	x.DeployGasLimit = c.DeployGasLimit
	x.BlockPeriod = c.BlockPeriod
	x.RewardEffectivePeriod = c.RewardEffectivePeriod
	x.Nodes = c.Nodes
	x.FinalOwner = c.FinalOwner

	xc := new(XCrossChainConfig)
	xc.PolyAccountDefaultPassphrase = c.CrossChain.PolyAccountDefaultPassphrase
	xc.PolyRPCAddress = c.CrossChain.PolyRPCAddress

	// poly side chain configuration
	xc.PaletteSideChainID = c.CrossChain.PaletteSideChainID
	xc.PaletteSideChainName = c.CrossChain.PaletteSideChainName
	xc.PaletteECCD = c.CrossChain.PaletteECCD.Hex()
	xc.PaletteECCM = c.CrossChain.PaletteECCM.Hex()
	xc.PaletteCCMP = c.CrossChain.PaletteCCMP.Hex()
	xc.PaletteNFTProxy = c.CrossChain.PaletteNFTProxy.Hex()

	// ethereum side chain configuration
	xc.EthereumSideChainID = c.CrossChain.EthereumSideChainID
	xc.EthereumSideChainName = c.CrossChain.EthereumSideChainName
	xc.EthereumECCD = c.CrossChain.EthereumECCD.Hex()
	xc.EthereumECCM = c.CrossChain.EthereumECCM.Hex()
	xc.EthereumCCMP = c.CrossChain.EthereumCCMP.Hex()
	xc.EthereumPLTAsset = c.CrossChain.EthereumPLTAsset.Hex()
	xc.EthereumPLTProxy = c.CrossChain.EthereumPLTProxy.Hex()
	xc.EthereumNFTProxy = c.CrossChain.EthereumNFTProxy.Hex()

	// ethereum node rpc and account
	xc.EthereumRPCUrl = c.CrossChain.EthereumRPCUrl
	xc.EthereumAccount = c.CrossChain.EthereumAccount
	xc.EthereumAccountPassword = c.CrossChain.EthereumAccountPassword
	xc.EthereumOwner = c.CrossChain.EthereumOwner
	xc.EthereumOwnerPassword = c.CrossChain.EthereumOwnerPassword
	x.CrossChain = xc

	enc, err := json.Marshal(x)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(ConfigFilePath, enc, os.ModePerm)
}

func LoadParams(fileName string, data interface{}) error {
	filePath := files.FullPath(Conf.Environment.LocalWorkspace, testCaseDir, fileName)
	bz, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	return json.Unmarshal(bz, data)
}

func LoadAccount(address string) (*ecdsa.PrivateKey, error) {
	address = strings.ToLower(address)
	filepath := files.FullPath(Conf.Environment.LocalWorkspace, keystoreDir, address)
	keyJson, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: [%v]", err)
	}

	key, err := keystore.DecryptKey(keyJson, Conf.DefaultPassphrase)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt keyjson: [%v]", err)
	}

	return key.PrivateKey, nil
}

type CrossChainConfig struct {
	// poly account and node rpc url
	PolyAccountDefaultPassphrase string
	PolyRPCAddress               string

	// poly side chain configuration
	PaletteSideChainID   uint64
	PaletteSideChainName string
	PaletteECCD          common.Address
	PaletteECCM          common.Address
	PaletteCCMP          common.Address
	PaletteNFTProxy      common.Address

	// ethereum side chain configuration
	EthereumSideChainID   uint64
	EthereumSideChainName string
	EthereumECCD          common.Address
	EthereumECCM          common.Address
	EthereumCCMP          common.Address
	EthereumPLTAsset      common.Address
	EthereumPLTProxy      common.Address
	EthereumNFTProxy      common.Address

	// ethereum node rpc and account
	EthereumRPCUrl          string
	EthereumAccount         string
	EthereumAccountPassword string
	EthereumOwner           string
	EthereumOwnerPassword   string
}

func (c *CrossChainConfig) LoadPolyAccountList() []*polysdk.Account {

	list := make([]*polysdk.Account, 0)

	dir := path.Join(Conf.Environment.LocalWorkspace, polyKeystoreDir)

	fs, _ := ioutil.ReadDir(dir)
	for _, f := range fs {
		fullPath := path.Join(dir, f.Name())
		acc, err := c.LoadPolyAccount(fullPath)
		if err != nil {
			panic(err)
		}
		list = append(list, acc)
	}

	return list
}

func (c *CrossChainConfig) LoadPolyTestCaseAccount(filename string) (*polysdk.Account, error) {
	filePath := files.FullPath(Conf.Environment.LocalWorkspace, testCaseDir, filename)
	return c.LoadPolyAccount(filePath)
}

func (c *CrossChainConfig) LoadPolyAccount(path string) (*polysdk.Account, error) {
	polySDK := polysdk.NewPolySdk()
	pwd := []byte(c.PolyAccountDefaultPassphrase)

	acc, err := getPolyAccountByPassword(polySDK, path, pwd)
	if err != nil {
		return nil, fmt.Errorf("failed to get poly account, err: %s", err)
	}
	return acc, nil
}

func (c *CrossChainConfig) LoadETHAccount() (*ecdsa.PrivateKey, error) {
	dir := path.Join(Conf.Environment.LocalWorkspace, ethKeystoreDir)
	fullPath := path.Join(dir, c.EthereumAccount)
	return c.LoadAccountWithPathAndPwd(fullPath, c.EthereumAccountPassword)
}

func (c *CrossChainConfig) LoadETHOwner() (*ecdsa.PrivateKey, error) {
	dir := path.Join(Conf.Environment.LocalWorkspace, ethKeystoreDir)
	fullPath := path.Join(dir, c.EthereumOwner)
	return c.LoadAccountWithPathAndPwd(fullPath, c.EthereumOwnerPassword)
}

func (c *CrossChainConfig) LoadAccountWithPathAndPwd(path string, pwd string) (*ecdsa.PrivateKey, error) {
	enc, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if len(enc) <= 64 {
		bz, err := hex.DecodeString(string(enc))
		if err != nil {
			return nil, err
		}
		return crypto.ToECDSA(bz)
	}

	key, err := keystore.DecryptKey(enc, pwd)
	if err != nil {
		return nil, err
	}

	return key.PrivateKey, nil
}

func (c *CrossChainConfig) StorePaletteECCD(addr common.Address) error {
	c.PaletteECCD = addr
	return SaveConfig(Conf)
}

func (c *CrossChainConfig) StorePaletteECCM(addr common.Address) error {
	c.PaletteECCM = addr
	return SaveConfig(Conf)
}

func (c *CrossChainConfig) StorePaletteCCMP(addr common.Address) error {
	c.PaletteCCMP = addr
	return SaveConfig(Conf)
}

func (c *CrossChainConfig) StorePaletteNFTProxy(addr common.Address) error {
	c.PaletteNFTProxy = addr
	return SaveConfig(Conf)
}

func (c *CrossChainConfig) StoreEthereumECCD(addr common.Address) error {
	c.EthereumECCD = addr
	return SaveConfig(Conf)
}

func (c *CrossChainConfig) StoreEthereumECCM(addr common.Address) error {
	c.EthereumECCM = addr
	return SaveConfig(Conf)
}

func (c *CrossChainConfig) StoreEthereumCCMP(addr common.Address) error {
	c.EthereumCCMP = addr
	return SaveConfig(Conf)
}

func (c *CrossChainConfig) StoreEthereumNFTProxy(addr common.Address) error {
	c.EthereumNFTProxy = addr
	return SaveConfig(Conf)
}

func (c *CrossChainConfig) StoreEthereumPLTAsset(addr common.Address) error {
	c.EthereumPLTAsset = addr
	return SaveConfig(Conf)
}

func (c *CrossChainConfig) StoreEthereumPLTProxy(addr common.Address) error {
	c.EthereumPLTProxy = addr
	return SaveConfig(Conf)
}

type FinalOwner struct {
	PaletteFinalOwner  common.Address
	EthereumFinalOwner common.Address
}

func getPolyAccountByPassword(sdk *polysdk.PolySdk, path string, pwd []byte) (
	*polysdk.Account, error) {
	wallet, err := sdk.OpenWallet(path)
	if err != nil {
		return nil, fmt.Errorf("open wallet error: %v", err)
	}
	acc, err := wallet.GetDefaultAccount(pwd)
	if err != nil {
		return nil, fmt.Errorf("getDefaultAccount error: %v", err)
	}
	return acc, nil
}

func LoadContract(fileName string, data interface{}) error {
	filePath := files.FullPath(Conf.Environment.LocalWorkspace, setupDir, fileName)
	keyJson, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	return json.Unmarshal(keyJson, data)
}

func ShellPath(fileName string) string {
	return files.FullPath(Conf.Environment.LocalWorkspace, "", fileName)
}

func GenesisNodeNumber() int {
	filepath := files.FullPath(Conf.Environment.LocalWorkspace, setupDir, "static-nodes.json")
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
