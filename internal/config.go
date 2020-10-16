package internal

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/palettechain/onRobot/pkg/files"
	"github.com/palettechain/onRobot/pkg/sdk"
	"github.com/palettechain/onRobot/pkg/shell"
	xtime "github.com/palettechain/onRobot/pkg/time"
)

const (
	paramsDir   = "params"
	keystoreDir = "keystore"
)

var config = new(Config)

type Config struct {
	Env               string
	Workspace         string
	DefaultPassphrase string
	GasLimit          json.Number
	DeployGasLimit    json.Number
	BlockPeriod       xtime.Duration
}

func Init(path string) {
	if err := loadConfig(path, config); err != nil {
		panic(err)
	}

	shell.Init(config.Env, config.Workspace)

	gasLimit, _ := config.GasLimit.Int64()
	deployGasLimit, _ := config.DeployGasLimit.Int64()
	sdk.Init(uint64(gasLimit), uint64(deployGasLimit), time.Duration(config.BlockPeriod))
}

func loadConfig(filepath string, ins interface{}) error {
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

func loadParams(fileName string, data interface{}) error {
	filePath := files.FullPath(config.Workspace, paramsDir, fileName)
	bz, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	return json.Unmarshal(bz, data)
}

func loadAccount(keyhex string) *keystore.Key {
	filepath := files.FullPath(config.Workspace, keystoreDir, keyhex)
	keyJson, err := ioutil.ReadFile(filepath)
	if err != nil {
		panic(fmt.Errorf("failed to read file: [%v]", err))
	}

	key, err := keystore.DecryptKey(keyJson, config.DefaultPassphrase)
	if err != nil {
		panic(fmt.Errorf("failed to decrypt keyjson: [%v]", err))
	}

	return key
}

func shellPath(fileName string) string {
	return files.FullPath(config.Workspace, "", fileName)
}
