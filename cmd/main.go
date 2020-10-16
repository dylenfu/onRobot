package main

import (
	"flag"
	"math/rand"
	"strings"
	"time"

	"github.com/palettechain/onRobot/internal"
	//_ "github.com/palettechain/onRobot/internal"
	core "github.com/palettechain/onRobot/pkg/frame"
	"github.com/palettechain/onRobot/pkg/log"
)

var (
	loglevel   int    // log level [1: debug, 2: info]
	configpath string //config file
	Methods    string //Methods list in cmdline
)

func init() {
	flag.StringVar(&configpath, "config", "config.json", "configpath of palette-tool")
	flag.StringVar(&Methods, "t", "connect", "methods to run. use ',' to split methods")
	flag.IntVar(&loglevel, "loglevel", 2, "loglevel [1: debug, 2: info]")

	flag.Parse()
}

func main() {
	rand.Seed(time.Now().UnixNano())
	defer time.Sleep(time.Second)

	log.InitLog(loglevel, log.Stdout)
	internal.Init(configpath)
	internal.Endpoint()

	methods := make([]string, 0)
	if Methods != "" {
		methods = strings.Split(Methods, ",")
	}

	core.Tool.Start(methods)
}
