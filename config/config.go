package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/palettechain/onRobot/pkg/files"
	"github.com/palettechain/onRobot/pkg/sdk"
	xtime "github.com/palettechain/onRobot/pkg/time"
)

const (
	paramsDir   = "params"
	keystoreDir = "keystore"
)

var (
	Conf          = new(Config)
	AdminKey *keystore.Key
)

type Config struct {
	Environment       *Env
	DefaultPassphrase string
	AdminAccount      string
	Accounts          []string
	GasLimit          uint64
	DeployGasLimit    uint64
	BlockPeriod       xtime.Duration
	Nodes             []*Node
}

type Node struct {
	Address string
	Nodekey string
}

type Env struct {
	Workspace    string
	NodeIdxStart int
	NodeNum      int
	NetworkID    int
	StartRPCPort int
	StartP2PPort int
	LogLevel     int
	RPCAddress   string
}

func Init(path string) {
	if err := LoadConfig(path, Conf); err != nil {
		panic(err)
	}

	sdk.Init(Conf.GasLimit, Conf.DeployGasLimit, time.Duration(Conf.BlockPeriod))

	AdminKey = LoadAccount(Conf.AdminAccount)
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
	filePath := files.FullPath(Conf.Environment.Workspace, paramsDir, fileName)
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
