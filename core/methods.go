package core

import (
	"time"

	"github.com/palettechain/onRobot/pkg/log"
	"github.com/palettechain/onRobot/pkg/shell"
)

func Demo() bool {
	log.Info("Hello, Palette chain")
	return true
}

func ResetNetwork() bool {
	var params struct {
		ShellPath string
	}

	if err := loadParams("Reset.json", &params); err != nil {
		log.Error(err)
		return false
	}

	shellPath := shellPath(params.ShellPath)
	shell.Exec(shellPath)
	return true
}

func gc() {
}

func wait(nBlock int) {
	time.Sleep(time.Duration(config.BlockPeriod) * time.Duration(nBlock))
}
