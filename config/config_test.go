package config

import (
	"encoding/hex"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
)

func TestNodeKey(t *testing.T) {
	nodeKey := "3d9c828244d3b2da70233a0a2aea7430feda17bded6edd7f0c474163802a431c"

	bz, err := hex.DecodeString(nodeKey)
	assert.NoError(t, err)

	prikey, err := crypto.ToECDSA(bz)
	assert.NoError(t, err)

	t.Logf(prikey.X.String())
}

// test this case on with command of `go test -count=1 -v github.com/palettechain/onRobot/config -run TestEnv`
func TestEnv(t *testing.T) {
	data, ok := os.LookupEnv(envName)
	if !ok {
		t.Errorf("environment %s not settle", envName)
	}
	t.Log(data)
}
