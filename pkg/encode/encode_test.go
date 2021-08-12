package encode

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/contracts/native/utils"
	"github.com/stretchr/testify/assert"
)

func TestDurationText(t *testing.T) {
	var (
		input  = []byte("10s")
		output = time.Second * 10
		d      Duration
	)
	if err := d.UnmarshalText(input); err != nil {
		t.FailNow()
	}
	if int64(output) != int64(d) {
		t.FailNow()
	}

	enc, err := d.MarshalText()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, input, enc)
}

func TestConcurrentMapReadAndMapWrite(t *testing.T) {
	m := make(map[int]string)
	for i := 0; i < 3; i++ {
		m[i] = fmt.Sprintf("data%d", i)
	}

	//for i := 0;i<100;i++ {
	//	go t.Logf(m[2][:])
	//}
	for i := 0; i < 60; i++ {
		go func(idx int) {
			m[2] = fmt.Sprintf("data%d", idx)
		}(i)
	}
	time.Sleep(1 * time.Second)
}

func TestSimple(t *testing.T) {
	const EthCrossChainManagerABI = `[{"inputs":[{"internalType":"address","name":"_eccd","type":"address"}],"payable":false,"stateMutability":"nonpayable","type":"constructor"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"uint256","name":"height","type":"uint256"},{"indexed":false,"internalType":"bytes","name":"rawHeader","type":"bytes"}],"name":"ChangeBookKeeperEvent","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"sender","type":"address"},{"indexed":false,"internalType":"bytes","name":"txId","type":"bytes"},{"indexed":false,"internalType":"address","name":"proxyOrAssetContract","type":"address"},{"indexed":false,"internalType":"uint64","name":"toChainId","type":"uint64"},{"indexed":false,"internalType":"bytes","name":"toContract","type":"bytes"},{"indexed":false,"internalType":"bytes","name":"rawdata","type":"bytes"}],"name":"CrossChainEvent","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"uint256","name":"height","type":"uint256"},{"indexed":false,"internalType":"bytes","name":"rawHeader","type":"bytes"}],"name":"InitGenesisBlockEvent","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"previousOwner","type":"address"},{"indexed":true,"internalType":"address","name":"newOwner","type":"address"}],"name":"OwnershipTransferred","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"account","type":"address"}],"name":"Paused","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"account","type":"address"}],"name":"Unpaused","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"uint64","name":"fromChainID","type":"uint64"},{"indexed":false,"internalType":"bytes","name":"toContract","type":"bytes"},{"indexed":false,"internalType":"bytes","name":"crossChainTxHash","type":"bytes"},{"indexed":false,"internalType":"bytes","name":"fromChainTxHash","type":"bytes"}],"name":"VerifyHeaderAndExecuteTxEvent","type":"event"},{"constant":true,"inputs":[],"name":"EthCrossChainDataAddress","outputs":[{"internalType":"address","name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"internalType":"bytes","name":"rawHeader","type":"bytes"},{"internalType":"bytes","name":"pubKeyList","type":"bytes"},{"internalType":"bytes","name":"sigList","type":"bytes"}],"name":"changeBookKeeper","outputs":[{"internalType":"bool","name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"uint64","name":"toChainId","type":"uint64"},{"internalType":"bytes","name":"toContract","type":"bytes"},{"internalType":"bytes","name":"method","type":"bytes"},{"internalType":"bytes","name":"txData","type":"bytes"}],"name":"crossChain","outputs":[{"internalType":"bool","name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"bytes","name":"rawHeader","type":"bytes"},{"internalType":"bytes","name":"pubKeyList","type":"bytes"}],"name":"initGenesisBlock","outputs":[{"internalType":"bool","name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"isOwner","outputs":[{"internalType":"bool","name":"","type":"bool"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"owner","outputs":[{"internalType":"address","name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[],"name":"pause","outputs":[{"internalType":"bool","name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"paused","outputs":[{"internalType":"bool","name":"","type":"bool"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[],"name":"renounceOwnership","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"newOwner","type":"address"}],"name":"transferOwnership","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[],"name":"unpause","outputs":[{"internalType":"bool","name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"newEthCrossChainManagerAddress","type":"address"}],"name":"upgradeToNew","outputs":[{"internalType":"bool","name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"bytes","name":"proof","type":"bytes"},{"internalType":"bytes","name":"rawHeader","type":"bytes"},{"internalType":"bytes","name":"headerProof","type":"bytes"},{"internalType":"bytes","name":"curRawHeader","type":"bytes"},{"internalType":"bytes","name":"headerSig","type":"bytes"}],"name":"verifyHeaderAndExecuteTx","outputs":[{"internalType":"bool","name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"}]`
	inputStr := "0xd450e04c00000000000000000000000000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000000000180000000000000000000000000000000000000000000000000000000000000026000000000000000000000000000000000000000000000000000000000000002a000000000000000000000000000000000000000000000000000000000000002e000000000000000000000000000000000000000000000000000000000000000b2b120479b395e2f9819ee0ddb622bf0f736becdf8ec26bb5d7e467ee2f6e2931985400a00000000000000010001001434d4a23a1fc0c694f0d74ddaf9d8d564cfe2d430020000000000000014250e76987d838a75310c34bf422ea9f1ac4cc90606756e6c6f636b4a14956f47f50a910163d8bf957cf5846d573e7f87ca14c8a65fadf0e0ddaf421f28feab69bf6e2e589963f69f3cac80b974e7758200000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000be000000000200000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000003c95acd6d2114c224e8536140abb85ce777f6a900c3d77dac953dd1ace36548b00000000000000000000000000000000000000000000000000000000000000000000000000ca9a3b0200000000000000010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000415abe67dd3d88b053e160d0caa70f004801687cd2e70d7be7961cf49d9c0d9c6b6aee14666726d7789c66e23482810a55a01340f2cff8833ffecaeb807279fa1e0100000000000000000000000000000000000000000000000000000000000000"
	ab, err := abi.JSON(strings.NewReader(EthCrossChainManagerABI))
	assert.NoError(t, err)

	type CrossChainInput struct {
		Proof        []byte
		RawHeader    []byte
		HeaderProof  []byte
		CurRawHeader []byte
		HeaderSig    []byte
	}
	payload, err := hexutil.Decode(inputStr)
	assert.NoError(t, err)
	input := new(CrossChainInput)
	err = utils.UnpackMethod(ab, "verifyHeaderAndExecuteTx", input, payload)
	assert.NoError(t, err)

	t.Logf("proof is %s", hexutil.Encode(input.Proof))
	t.Logf("rawHeader is %s", hexutil.Encode(input.RawHeader))
	t.Logf("headerProof is %s", hexutil.Encode(input.HeaderProof))
	t.Logf("curRawHeader is %s", hexutil.Encode(input.RawHeader))
	t.Logf("headerSig is %s", hexutil.Encode(input.HeaderSig))
	//data := "250e76987d838a75310c34bf422ea9f1ac4cc906"
	//enc, err := hexutil.Decode(data)
	//assert.NoError(t, err)
}
