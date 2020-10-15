package internal

import (
	"math/rand"
	"time"

	core "github.com/palettechain/onRobot/pkg/frame"
)

func init() {
	rand.Seed(time.Now().UnixNano())

	core.Tool.RegGCFunc(reset)

	core.Tool.RegMethod("demo", Demo)
}
