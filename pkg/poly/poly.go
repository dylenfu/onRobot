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
	putils "github.com/polynetwork/poly/native/service/utils"
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
) error {

	if txhash, err := c.sdk.Native.Hs.SyncGenesisHeader(
		chainID,
		genesisHeader,
		c.accArr,
	); err != nil {
		if strings.Contains(err.Error(), "had been initialized") {
			log.Info("eth already synced")
			return nil
		}
		return err
	} else {
		return c.WaitPolyTx(txhash)
	}
}

var sideChainRouter = putils.PALETTE_ROUTER

const (
	sideChainName        = "palette"
	sideChainBlockToWait = 1
)

func (c *PolyClient) RegisterSideChain(chainID uint64, eccdAddr common.Address) error {
	acc := c.GetSideChainOwner()
	eccd, err := hex.DecodeString(strings.Replace(eccdAddr.Hex(), "0x", "", 1))
	if err != nil {
		return fmt.Errorf("failed to decode eccd address, err: %s", err)
	}

	if txhash, err := c.sdk.Native.Scm.RegisterSideChain(
		acc.Address,
		chainID,
		sideChainRouter,
		sideChainName,
		sideChainBlockToWait,
		eccd,
		acc,
	); err != nil {
		if strings.Contains(err.Error(), "already registered") {
			log.Infof("palette chain %d already registered", chainID)
			return nil
		}
		if strings.Contains(err.Error(), "already requested") {
			log.Infof("palette chain %d already requested", chainID)
			return nil
		}
		return err
	} else {
		return c.WaitPolyTx(txhash)
	}
}

func (c *PolyClient) QuitSideChain(chainID uint64) error {
	acc := c.GetSideChainOwner()
	txhash, err := c.sdk.Native.Scm.QuitSideChain(chainID, acc)
	if err != nil {
		return err
	}
	return c.WaitPolyTx(txhash)
}

func (c *PolyClient) UpdateSideChain(chainID uint64, eccdAddr common.Address) error {
	acc := c.GetSideChainOwner()
	eccd, err := hex.DecodeString(strings.Replace(eccdAddr.Hex(), "0x", "", 1))
	if err != nil {
		return fmt.Errorf("failed to decode eccd address")
	}

	if txhash, err := c.sdk.Native.Scm.UpdateSideChain(
		acc.Address,
		chainID,
		sideChainRouter,
		sideChainName,
		sideChainBlockToWait,
		eccd,
		acc,
	); err != nil {
		return err
	} else {
		return c.WaitPolyTx(txhash)
	}
}

func (c *PolyClient) ApproveRegisterSideChain(chainID uint64) error {
	var (
		txhash pcm.Uint256
		err    error
	)
	for i, acc := range c.accArr {
		txhash, err = c.sdk.Native.Scm.ApproveRegisterSideChain(chainID, acc)
		if err != nil {
			return fmt.Errorf("no%d - failed to approve %d: %v", i, chainID, err)
		}
		log.Infof("No%d: successful to approve register side chain %d: ( acc: %s, txhash: %s )",
			i, chainID, acc.Address.ToHexString(), txhash.ToHexString())
	}
	return c.WaitPolyTx(txhash)
}

func (c *PolyClient) ApproveQuitSideChain(chainID uint64) error {
	var (
		txhash pcm.Uint256
		err    error
	)
	for i, acc := range c.accArr {
		txhash, err = c.sdk.Native.Scm.ApproveQuitSideChain(chainID, acc)
		if err != nil {
			return fmt.Errorf("no%d - failed to approve %d: %v", i, chainID, err)
		}
		log.Infof("No%d: successful to approve quit side chain %d: ( acc: %s, txhash: %s )",
			i, chainID, acc.Address.ToHexString(), txhash.ToHexString())
	}
	return c.WaitPolyTx(txhash)
}

func (c *PolyClient) ApproveUpdateSideChain(chainID uint64) error {
	var (
		txhash pcm.Uint256
		err    error
	)
	for i, acc := range c.accArr {
		txhash, err = c.sdk.Native.Scm.ApproveUpdateSideChain(chainID, acc)
		if err != nil {
			return fmt.Errorf("no%d - failed to approve %d: %v", i, chainID, err)
		}
		log.Infof("No%d: successful to approve update side chain %d: ( acc: %s, txhash: %s )",
			i, chainID, acc.Address.ToHexString(), txhash.ToHexString())
	}
	return c.WaitPolyTx(txhash)
}

func (c *PolyClient) GetSideChainOwner() *polysdk.Account {
	return c.accArr[0]
}

func (c *PolyClient) GetBlockByHeight(height uint32) (*ptyp.Block, error) {
	return c.sdk.GetBlockByHeight(height)
}

func (c *PolyClient) GetCurrentBlockHeight() (uint32, error) {
	return c.sdk.GetCurrentBlockHeight()
}

func GetBookeeper(block *ptyp.Block) ([]keypair.PublicKey, error) {
	info := new(vconfig.VbftBlockInfo)//&vconfig.VbftBlockInfo{}
	info.NewChainConfig = new(vconfig.ChainConfig)
	if err := json.Unmarshal(block.Header.ConsensusPayload, info); err != nil {
		return nil, fmt.Errorf("failed to unmarshal consensus payload, err: %s", err)
	}

	if info == nil {
		log.Info("info is nil")
	}
	if info.NewChainConfig == nil {
		log.Info("new chain config is nil")
	}
	bookkeepers := make([]keypair.PublicKey, 0 )
	for _, peer := range info.NewChainConfig.Peers {
		log.Infof("poly peer index %d id %s", peer.Index, peer.ID)
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

func (c *PolyClient) WaitPolyTx(hash pcm.Uint256) error {
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
				hash.ToHexString(), 300)
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

// todo
func AddrConvertUp(src pcm.Address) common.Address {
	return common.BytesToAddress(src[:])
}

func AddrConvertDown(src common.Address) (pcm.Address, error) {
	return pcm.AddressFromHexString(src.Hex())
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
