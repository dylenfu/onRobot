package config

import (
	"bufio"
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
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/palettechain/onRobot/pkg/dao"
	"github.com/palettechain/onRobot/pkg/encode"
	"github.com/palettechain/onRobot/pkg/files"
	"github.com/palettechain/onRobot/pkg/log"
	"github.com/palettechain/onRobot/pkg/poly"
	"github.com/palettechain/onRobot/pkg/sdk"
	polysdk "github.com/polynetwork/poly-go-sdk"
)

const (
	testCaseDir     = "cases"
	keystoreDir     = "keystore"
	setupDir        = "setup"
	polyKeystoreDir = "poly_keystore"
	ethKeystoreDir  = "eth_keystore"
	dataDir         = "leveldb"
	envName         = "ONROBOT"
)

type pwdSessionType byte

const (
	pwdSessionUnknown pwdSessionType = iota
	pwdSessionETH
	pwdSessionPLT
	pwdSessionPoly
)

var (
	Conf, BakConf                = new(Config), new(Config)
	AdminKey, CrossChainAdminKey *ecdsa.PrivateKey
	ConfigFilePath               string
	ethPwdSession                = make(map[common.Address]string)
	pltPwdSession                = make(map[common.Address]string)
)

type Config struct {
	Environment            *Env
	Network                *Network
	Rpc                    string
	DefaultPassphrase      string
	AdminAccount           common.Address
	CrossChainAdminAccount common.Address
	BaseRewardPool         common.Address
	Accounts               []common.Address
	GasLimit               uint64
	DeployGasLimit         uint64
	BlockPeriod            encode.Duration
	RewardEffectivePeriod  int // 区块奖励周期/参数生效周期
	Nodes                  []*Node
	CrossChain             *CrossChainConfig
	FinalOwner             *FinalOwner
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
	acc := common.HexToAddress(n.StakeAccount)
	enc, err := readWalletFile(keystoreDir, acc)
	if err != nil {
		panic(fmt.Sprintf("load keystore err %v", err))
	}
	if ks, err := repeatDecrypt(enc, acc, Conf.DefaultPassphrase, pwdSessionPLT); err != nil {
		panic(fmt.Sprintf("decrypt key %s err %v", n.StakeAccount, err))
	} else {
		n.sapk = ks.PrivateKey
	}
}

func (n *Node) NodeDirPath() string {
	n.once.Do(n.init)
	data := fmt.Sprintf("node%d", n.Index)
	nodedir := path.Join(Conf.Environment.WorkSpace(), data)
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

var (
	envOnce sync.Once
	env     string
)

func (e *Env) WorkSpace() string {
	envOnce.Do(func() {
		env = os.Getenv(envName)
	})
	return path.Join(e.LocalWorkspace, env)
}

type Network struct {
	NodeIndexStart    int
	GenesisNodeNumber int
	ValidatorsNumber  int
}

func Init(filepath string) {
	ConfigFilePath = filepath
	err := LoadConfig(ConfigFilePath, Conf)
	if err != nil {
		panic(err)
	}

	// init leveldb
	dir := path.Join(Conf.Environment.WorkSpace(), dataDir)
	dao.NewDao(dir)

	// sort nodes with node index
	sort.Slice(Conf.Nodes, func(i, j int) bool {
		return Conf.Nodes[i].Index < Conf.Nodes[j].Index
	})

	// load nodes privateKey
	sdk.Init(Conf.GasLimit, Conf.DeployGasLimit, time.Duration(Conf.BlockPeriod))

	AdminKey, err = LoadPaletteAccount(Conf.AdminAccount)
	if err != nil {
		panic(err)
	}

	CrossChainAdminKey, err = LoadPaletteAccount(Conf.CrossChainAdminAccount)
	if err != nil {
		panic(err)
	}

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
		PaletteECCD          common.Address
		PaletteECCM          common.Address
		PaletteCCMP          common.Address
		PaletteNFTProxy      common.Address
		PalettePLTWrapper    common.Address
		PaletteNFTWrapper    common.Address

		// ethereum side chain configuration
		EthereumSideChainID   uint64
		EthereumSideChainName string
		EthereumECCD          common.Address
		EthereumECCM          common.Address
		EthereumCCMP          common.Address
		EthereumPLTAsset      common.Address
		EthereumPLTProxy      common.Address
		EthereumNFTProxy      common.Address
		EthereumPLTWrapper    common.Address
		EthereumNFTWrapper    common.Address

		// ethereum node rpc and account
		EthereumRPCUrl          string
		EthereumAccount         common.Address
		EthereumAccountPassword string
		EthereumOwner           common.Address
		EthereumOwnerPassword   string
	}

	type XConfig struct {
		Environment            *Env
		Network                *Network
		DefaultPassphrase      string
		Rpc                    string
		AdminAccount           common.Address
		CrossChainAdminAccount common.Address
		BaseRewardPool         common.Address
		Accounts               []common.Address
		GasLimit               uint64
		DeployGasLimit         uint64
		BlockPeriod            encode.Duration
		RewardEffectivePeriod  int // 区块奖励周期/参数生效周期
		Nodes                  []*Node
		CrossChain             *XCrossChainConfig
		FinalOwner             *FinalOwner
	}

	x := new(XConfig)
	x.Environment = c.Environment
	x.Network = c.Network
	x.DefaultPassphrase = c.DefaultPassphrase
	x.Rpc = c.Rpc
	x.AdminAccount = c.AdminAccount
	x.CrossChainAdminAccount = c.CrossChainAdminAccount
	x.BaseRewardPool = c.BaseRewardPool
	x.GasLimit = c.GasLimit
	x.DeployGasLimit = c.DeployGasLimit
	x.BlockPeriod = c.BlockPeriod
	x.RewardEffectivePeriod = c.RewardEffectivePeriod
	x.Nodes = c.Nodes
	x.FinalOwner = c.FinalOwner
	x.Accounts = make([]common.Address, 0)
	for _, acc := range c.Accounts {
		x.Accounts = append(x.Accounts, acc)
	}

	xc := new(XCrossChainConfig)
	xc.PolyAccountDefaultPassphrase = c.CrossChain.PolyAccountDefaultPassphrase
	xc.PolyRPCAddress = c.CrossChain.PolyRPCAddress

	// poly side chain configuration
	xc.PaletteSideChainID = c.CrossChain.PaletteSideChainID
	xc.PaletteSideChainName = c.CrossChain.PaletteSideChainName
	xc.PaletteECCD = c.CrossChain.PaletteECCD
	xc.PaletteECCM = c.CrossChain.PaletteECCM
	xc.PaletteCCMP = c.CrossChain.PaletteCCMP
	xc.PaletteNFTProxy = c.CrossChain.PaletteNFTProxy

	// ethereum side chain configuration
	xc.EthereumSideChainID = c.CrossChain.EthereumSideChainID
	xc.EthereumSideChainName = c.CrossChain.EthereumSideChainName
	xc.EthereumECCD = c.CrossChain.EthereumECCD
	xc.EthereumECCM = c.CrossChain.EthereumECCM
	xc.EthereumCCMP = c.CrossChain.EthereumCCMP
	xc.EthereumPLTAsset = c.CrossChain.EthereumPLTAsset
	xc.EthereumPLTProxy = c.CrossChain.EthereumPLTProxy
	xc.EthereumNFTProxy = c.CrossChain.EthereumNFTProxy

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
	filePath := files.FullPath(Conf.Environment.WorkSpace(), testCaseDir, fileName)
	bz, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	return json.Unmarshal(bz, data)
}

func LoadPaletteAccount(address common.Address) (*ecdsa.PrivateKey, error) {
	enc, err := readWalletFile(keystoreDir, address)
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

	key, err := repeatDecrypt(enc, address, Conf.DefaultPassphrase, pwdSessionPLT)
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
	PalettePLTWrapper    common.Address
	PaletteNFTWrapper    common.Address

	// ethereum side chain configuration
	EthereumSideChainID   uint64
	EthereumSideChainName string
	EthereumECCD          common.Address
	EthereumECCM          common.Address
	EthereumCCMP          common.Address
	EthereumPLTAsset      common.Address
	EthereumPLTProxy      common.Address
	EthereumNFTProxy      common.Address
	EthereumPLTWrapper    common.Address
	EthereumNFTWrapper    common.Address

	// ethereum node rpc and account
	EthereumRPCUrl          string
	EthereumAccount         common.Address
	EthereumAccountPassword string
	EthereumOwner           common.Address
	EthereumOwnerPassword   string
}

func (c *CrossChainConfig) LoadPolyAccountList() []*polysdk.Account {

	list := make([]*polysdk.Account, 0)

	dir := path.Join(Conf.Environment.WorkSpace(), polyKeystoreDir)
	fs, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	fmt.Println("fs length ", len(fs))
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

func (c *CrossChainConfig) LoadCurrentBookKeeperBytes() []byte {
	accList := c.LoadPolyAccountList()
	keepers := []keypair.PublicKey{}
	for _, v := range accList {
		keepers = append(keepers, v.PublicKey)
	}
	sink, _ := poly.AssemblePubKeyList(keepers)
	return sink.Bytes()
}

func (c *CrossChainConfig) LoadPolyTestCaseAccount(filename string) (*polysdk.Account, error) {
	filePath := files.FullPath(Conf.Environment.WorkSpace(), testCaseDir, filename)
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
	return c.CustomLoadEthAccount(c.EthereumAccount, c.EthereumAccountPassword)
}

func (c *CrossChainConfig) LoadETHOwner() (*ecdsa.PrivateKey, error) {
	return c.CustomLoadEthAccount(c.EthereumOwner, c.EthereumOwnerPassword)
}

func (c *CrossChainConfig) CustomLoadEthAccount(acc common.Address, pwd string) (*ecdsa.PrivateKey, error) {
	enc, err := readWalletFile(ethKeystoreDir, acc)
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

	key, err := repeatDecrypt(enc, acc, pwd, pwdSessionETH)
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

func (c *CrossChainConfig) StorePalettePLTWrapper(addr common.Address) error {
	c.PalettePLTWrapper = addr
	return SaveConfig(Conf)
}

func (c *CrossChainConfig) StorePaletteNFTWrapper(addr common.Address) error {
	c.PaletteNFTWrapper = addr
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
	if err == nil {
		return acc, nil
	}

	reader := bufio.NewReader(os.Stdin)
	for i := 0; i < 10; i++ {
		curPwd, err := reader.ReadString('\n')
		if err != nil {
			log.Infof("input error, try it again......")
			continue
		}
		curPwd = strings.Trim(curPwd, " ")
		curPwd = strings.Trim(curPwd, "\r")
		curPwd = strings.Trim(curPwd, "\n")
		if acc, err := wallet.GetDefaultAccount([]byte(curPwd)); err == nil {
			return acc, nil
		} else {
			log.Infof("password invalid, err %s, try it again......", err.Error())
		}
	}

	return acc, nil
}

func LoadContract(fileName string, data interface{}) error {
	filePath := files.FullPath(Conf.Environment.WorkSpace(), setupDir, fileName)
	keyJson, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	return json.Unmarshal(keyJson, data)
}

func ShellPath(fileName string) string {
	return files.FullPath(Conf.Environment.WorkSpace(), "", fileName)
}

func GenesisNodeNumber() int {
	filepath := files.FullPath(Conf.Environment.WorkSpace(), setupDir, "static-nodes.json")
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

func repeatDecrypt(enc []byte, account common.Address, pwd string, typ pwdSessionType) (key *keystore.Key, err error) {
	if existPwd, err := getPwdSession(account, typ); err == nil {
		return keystore.DecryptKey(enc, existPwd)
	}

	if key, err = keystore.DecryptKey(enc, pwd); err == nil {
		_ = setPwdSession(account, pwd, typ)
		return
	}

	log.Infof("please input password for ethereum account %s", account.Hex())

	reader := bufio.NewReader(os.Stdin)
	var curPwd string
	for i := 0; i < 10; i++ {
		curPwd, err = reader.ReadString('\n')
		if err != nil {
			log.Infof("input error, try it again......")
			continue
		}
		curPwd = strings.Trim(curPwd, " ")
		curPwd = strings.Trim(curPwd, "\r")
		curPwd = strings.Trim(curPwd, "\n")
		if key, err = keystore.DecryptKey(enc, curPwd); err == nil {
			_ = setPwdSession(account, curPwd, typ)
			return
		} else {
			log.Infof("password invalid, err %s, try it again......", err.Error())
		}
	}
	return
}

func readWalletFile(storeDir string, acc common.Address) (enc []byte, err error) {
	dir := path.Join(Conf.Environment.WorkSpace(), storeDir)
	normalAddr := path.Join(dir, acc.Hex())
	lowerAddr := path.Join(dir, strings.ToLower(acc.Hex()))
	if enc, err = ioutil.ReadFile(normalAddr); err != nil {
		if enc, err = ioutil.ReadFile(lowerAddr); err != nil {
			return nil, fmt.Errorf("failed to read file: [%v]", err)
		}
	}
	return
}

func setPwdSession(acc common.Address, pwd string, typ pwdSessionType) error {
	return dao.SavePwd(byte(typ), acc.Bytes(), []byte(pwd))
}

func getPwdSession(acc common.Address, typ pwdSessionType) (string, error) {
	bz, err := dao.GetPwd(byte(typ), acc.Bytes())
	if err != nil {
		return "", err
	}
	return string(bz), nil
}
