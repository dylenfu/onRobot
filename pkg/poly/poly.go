package poly

import (
	"bytes"
	"crypto/elliptic"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ontio/ontology-crypto/ec"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology-crypto/sm2"
	"github.com/palettechain/onRobot/pkg/log"
	polysdk "github.com/polynetwork/poly-go-sdk"
	pcm "github.com/polynetwork/poly/common"
	vconfig "github.com/polynetwork/poly/consensus/vbft/config"
	ptyp "github.com/polynetwork/poly/core/types"
)

type PolyClient struct {
	sdk    *polysdk.PolySdk
	accArr []*polysdk.Account
}

func NewPolyClient(rpcAddr string, accArr []*polysdk.Account) (*PolyClient, error) {
	sdk := polysdk.NewPolySdk()
	sdk.NewRpcClient().SetAddress(rpcAddr)
	hdr, err := sdk.GetHeaderByHeight(0)
	if err != nil {
		return nil, err
	}
	sdk.SetChainId(hdr.ChainID)
	return &PolyClient{
		sdk:    sdk,
		accArr: accArr,
	}, nil
}

func (c *PolyClient) SyncGenesisBlock(
	chainID uint64,
	genesisHeader []byte,
) (common.Hash, error) {
	txhash, err := c.sdk.Native.Hs.SyncGenesisHeader(chainID, genesisHeader, c.accArr)
	if err != nil {
		if strings.Contains(err.Error(), "had been initialized") {
			log.Info("eth already synced")
		} else {
			return common.Hash{}, err
		}
	}

	return HashConvertUp(txhash), nil
}

func (c *PolyClient) GetBlockByHeight(height uint32) (*ptyp.Block, error) {
	return c.sdk.GetBlockByHeight(height)
}

func (c *PolyClient) GetCurrentBlockHeight() (uint32, error) {
	return c.sdk.GetCurrentBlockHeight()
}

func GetBookeeper(block *ptyp.Block) ([]keypair.PublicKey, error) {
	info := &vconfig.VbftBlockInfo{}
	if err := json.Unmarshal(block.Header.ConsensusPayload, info); err != nil {
		return nil, fmt.Errorf("failed to unmarshal consensus payload, err: %s", err)
	}

	var bookkeepers []keypair.PublicKey
	for _, peer := range info.NewChainConfig.Peers {
		keystr, _ := hex.DecodeString(peer.ID)
		key, _ := keypair.DeserializePublicKey(keystr)
		bookkeepers = append(bookkeepers, key)
	}
	bookkeepers = keypair.SortPublicKeys(bookkeepers)

	return bookkeepers, nil
}

func AssembleNoCompressBookeeper(bookeepers []keypair.PublicKey) []byte {
	publickeys := make([]byte, 0)
	for _, key := range bookeepers {
		publickeys = append(publickeys, GetOntNoCompressKey(key)...)
	}
	return publickeys
}

func (c *PolyClient) WaitPolyTx(txhash common.Hash) error {
	hash, err := HashConvertDown(txhash)
	if err != nil {
		return err
	}

	tick := time.NewTicker(100 * time.Millisecond)
	var h uint32
	startTime := time.Now()
	for range tick.C {
		h, _ = c.sdk.GetBlockHeightByTxHash(hash.ToHexString())
		curr, _ := c.sdk.GetCurrentBlockHeight()
		if h > 0 && curr > h {
			break
		}

		if startTime.Add(100 * time.Millisecond); startTime.Second() > 300 {
			return fmt.Errorf("tx( %s ) is not confirm for a long time ( over %d sec )",
				txhash, 300)
		}
	}

	return nil
}

func HashConvertUp(src pcm.Uint256) common.Hash {
	return common.BytesToHash(src[:])
}

func HashConvertDown(src common.Hash) (pcm.Uint256, error) {
	return pcm.Uint256ParseFromBytes(src[:])
}

func GetOntNoCompressKey(key keypair.PublicKey) []byte {
	var buf bytes.Buffer
	switch t := key.(type) {
	case *ec.PublicKey:
		switch t.Algorithm {
		case ec.ECDSA:
			// Take P-256 as a special case
			if t.Params().Name == elliptic.P256().Params().Name {
				return ec.EncodePublicKey(t.PublicKey, false)
			}
			buf.WriteByte(byte(0x12))
		case ec.SM2:
			buf.WriteByte(byte(0x13))
		}
		label, err := GetCurveLabel(t.Curve.Params().Name)
		if err != nil {
			panic(err)
		}
		buf.WriteByte(label)
		buf.Write(ec.EncodePublicKey(t.PublicKey, false))
	default:
		panic("err")
	}
	return buf.Bytes()
}

func GetCurveLabel(name string) (byte, error) {
	switch strings.ToUpper(name) {
	case strings.ToUpper(elliptic.P224().Params().Name):
		return 1, nil
	case strings.ToUpper(elliptic.P256().Params().Name):
		return 2, nil
	case strings.ToUpper(elliptic.P384().Params().Name):
		return 3, nil
	case strings.ToUpper(elliptic.P521().Params().Name):
		return 4, nil
	case strings.ToUpper(sm2.SM2P256V1().Params().Name):
		return 20, nil
	case strings.ToUpper(btcec.S256().Name):
		return 5, nil
	default:
		panic("err")
	}
}
