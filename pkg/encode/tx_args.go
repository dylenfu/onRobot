package encode

import (
	"encoding/hex"
	"fmt"
	"io"
	"math/big"

	"github.com/polynetwork/poly/common"
)

type TxArgs struct {
	ToAssetHash []byte
	ToAddress   []byte
	Amount      *big.Int
}

func (args *TxArgs) Serialization() []byte {
	sink := common.NewZeroCopySink(nil)
	sink.WriteVarBytes(args.ToAssetHash)
	sink.WriteVarBytes(args.ToAddress)
	raw, _ := PadFixedBytes(args.Amount, 32)
	sink.WriteBytes(raw)
	return sink.Bytes()
}

func (args *TxArgs) Deserialization(raw []byte) error {
	source := common.NewZeroCopySource(raw)
	assetHash, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("Args.Deserialization NextVarBytes AssetHash error:%s", io.ErrUnexpectedEOF)
	}

	toAddress, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("Args.Deserialization NextVarBytes ToAddress error:%s", io.ErrUnexpectedEOF)
	}

	value, eof := source.NextBytes(32)
	if eof {
		return fmt.Errorf("Args.Deserialization NextBytes Value error:%s", io.ErrUnexpectedEOF)
	}

	amt, err := UnpadFixedBytes(value, 32)
	if err != nil {
		return fmt.Errorf("faield to get amount: %v", err)
	}

	args.ToAssetHash = assetHash
	args.ToAddress = toAddress
	args.Amount = amt
	return nil
}

func UnpadFixedBytes(paddedBs []byte, intBsLen int) (*big.Int, error) {
	if len(paddedBs) != intBsLen {
		return nil, fmt.Errorf("UnpadFixedBytes only support 32 bytes value, but got:%s", hex.EncodeToString(paddedBs))
	}
	nonZeroPos := intBsLen - 1
	for i := nonZeroPos; i >= 0; i-- {
		p := paddedBs[i]
		if p != 0x0 {
			nonZeroPos = i
			break
		}
	}
	if nonZeroPos == intBsLen-1 && paddedBs[intBsLen-1]>>7 == 1 {
		return nil, fmt.Errorf("UnpadFixedBytes only support 32 bytes nonnegative value, but got:%s", hex.EncodeToString(paddedBs))
	}

	return big.NewInt(0).SetBytes(ToArrayReverse(paddedBs[:nonZeroPos+1])), nil
}

func ToArrayReverse(arr []byte) []byte {
	l := len(arr)
	x := make([]byte, 0)
	for i := l - 1; i >= 0; i-- {
		x = append(x, arr[i])
	}
	return x
}

func PadFixedBytes(bigint *big.Int, intBsLen int) ([]byte, error) {
	ret := make([]byte, intBsLen)
	if bigint.Cmp(big.NewInt(0)) < 0 {
		return nil, fmt.Errorf("PadFixedBytes doesnot support negative big.Int, but got:%s", bigint.String())
	}
	bigBs := bigint.Bytes()
	if len(bigBs) > intBsLen || (len(bigBs) == intBsLen && bigBs[0]>>7 == 1) {
		return nil, fmt.Errorf("PadFixedBytes only support maximum 2**255-1 big.Int, but got:%s", bigint.String())
	}
	copy(ret[:len(bigBs)], make([]byte, len(bigBs)))
	copy(ret[intBsLen-len(bigBs):], bigBs)
	return ToArrayReverse(ret), nil
}
