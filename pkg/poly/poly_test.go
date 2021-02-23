package poly

import (
	"github.com/palettechain/onRobot/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPolyClient_GetCurrentBlockHeight(t *testing.T) {
	rpc := "http://106.75.226.11:40436"
	configpath := "/Users/dylen/software/onRobot/build/target/config.json"
	config.Init(configpath)
	cli, err := NewPolyClient(rpc, nil)
	assert.NoError(t, err)

	height, err := cli.GetCurrentBlockHeight()
	assert.NoError(t, err)

	t.Logf("height %d", height)
}
