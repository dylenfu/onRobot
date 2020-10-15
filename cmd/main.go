package main

import (
	"flag"
	"github.com/palettechain/onRobot/pkg/log"
	"math/rand"
	"strings"
	"time"

	_ "github.com/palettechain/onRobot/internal"
	"github.com/palettechain/onRobot/internal/conf"
	core "github.com/palettechain/onRobot/pkg/frame"
)

var (
	configpath           string //config file
	LogConfig            string //Log config file
	TestCaseConfig       string // Test case file dir
	WalletConfig         string // Wallet path
	TransferWalletConfig string // Transfer wallet path
	Methods              string //Methods list in cmdline
	loglevel             int
	logDir               string
)

func init() {
	flag.StringVar(&configpath, "config", "target/robot/config.json", "configpath of palette-tool")
	flag.StringVar(&TestCaseConfig, "params", "target/robot/params", "Test params")
	flag.StringVar(&WalletConfig, "wallet", "target/robot/wallet.dat", "Wallet path")
	flag.StringVar(&TransferWalletConfig, "transfer", "target/robot/transfer_wallet.dat", "Transfer wallet path")
	flag.StringVar(&Methods, "t", "connect", "methods to run. use ',' to split methods")
	flag.IntVar(&loglevel, "loglevel", 2, "loglevel [1: debug, 2: info]")
	flag.StringVar(&logDir, "logdir", "target/node/log", "set log dir")

	flag.Parse()
}

func main() {
	rand.Seed(time.Now().UnixNano())
	conf.SetParamsDir(TestCaseConfig)
	defer time.Sleep(time.Second)

	log.InitLog(loglevel, log.Stdout)
	conf.Init(configpath)

	methods := make([]string, 0)
	if Methods != "" {
		methods = strings.Split(Methods, ",")
	}

	core.Tool.Start(methods)
}
