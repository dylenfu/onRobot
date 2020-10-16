package internal

import (
	"math/rand"
	"time"

	core "github.com/palettechain/onRobot/pkg/frame"
)

func Endpoint() {
	rand.Seed(time.Now().UnixNano())

	core.Tool.RegGCFunc(gc)

	core.Tool.RegMethod("demo", Demo)
	core.Tool.RegMethod("reset", ResetNetwork)

	// plt
	core.Tool.RegMethod("pltBalanceOf", PLTBalanceOf)
}
