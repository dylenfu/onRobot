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
	polycm "github.com/polynetwork/poly/common"
	vconfig "github.com/polynetwork/poly/consensus/vbft/config"
	polytype "github.com/polynetwork/poly/core/types"
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

// client的账户列表就是poly共识节点账户列表，可以通过注册和取消账户的方式实现bookKeeper的变更
func (c *PolyClient) RegNode(node *polysdk.Account) error {
	validators := c.accArr
	peer := vconfig.PubkeyID(node.PublicKey)

	if err := c.RegisterCandidate(peer, validators[0]); err != nil {
		return err
	} else {
		log.Infof("register %s success!", peer)
	}
	if err := c.ApproveCandidate(peer, validators); err != nil {
		return err
	} else {
		log.Infof("approve %s success", peer)
	}
	return c.CommitPolyDpos(validators)
}

func (c *PolyClient) QuitNode(acc *polysdk.Account) error {
	peer := vconfig.PubkeyID(acc.PublicKey)

	txhash, err := c.sdk.Native.Nm.QuitNode(peer, acc)
	if err != nil {
		return fmt.Errorf("failed to quit %s: %v", acc.Address.ToBase58(), err)
	}
	if err := c.WaitPolyTx(txhash); err != nil {
		return err
	}
	return c.CommitPolyDpos(c.accArr)
}

func (c *PolyClient) SyncGenesisBlock(
	selfChainID uint64,
	genesisHeader []byte,
) error {

	for idx, acc := range c.accArr {
		log.Infof("acc-%d %s", idx, acc.Address.ToBase58())
	}

	if txhash, err := c.sdk.Native.Hs.SyncGenesisHeader(
		selfChainID,
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

const sideChainBlockToWait = 1

func (c *PolyClient) RegisterSideChain(
	chainID uint64,
	eccdAddr common.Address,
	sideChainRouter uint64,
	sideChainName string,
) error {

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

func (c *PolyClient) UpdateSideChain(
	chainID uint64,
	eccdAddr common.Address,
	sideChainRouter uint64,
	sideChainName string,
) error {
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
		txhash polycm.Uint256
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
		txhash polycm.Uint256
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
		txhash polycm.Uint256
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

func (c *PolyClient) RegisterCandidate(peer string, validator *polysdk.Account) error {
	txHash, err := c.sdk.Native.Nm.RegisterCandidate(peer, validator)
	if err != nil {
		if strings.Contains(err.Error(), "already") {
			log.Warnf("candidate %s already registered: %v", peer, err)
			return nil
		}
		return fmt.Errorf("sendTransaction error: %v", err)
	}
	return c.WaitPolyTx(txHash)
}

func (c *PolyClient) ApproveCandidate(peer string, validators []*polysdk.Account) error {
	var (
		txhash polycm.Uint256
		err    error
	)

	for index, validator := range validators {
		txhash, err = c.sdk.Native.Nm.ApproveCandidate(peer, validator)
		if err != nil {
			return fmt.Errorf("node-%d sendTransaction error: %v", index, err)
		}
		log.Infof("node-%d approve %s", index, peer)
	}

	return c.WaitPolyTx(txhash)
}

func (c *PolyClient) CommitPolyDpos(accArr []*polysdk.Account) error {
	txhash, err := c.sdk.Native.Nm.CommitDpos(accArr)
	if err != nil {
		return err
	}
	return c.WaitPolyTx(txhash)
}

func (c *PolyClient) GetSideChainOwner() *polysdk.Account {
	return c.accArr[0]
}

func (c *PolyClient) GetBlockByHeight(height uint32) (*polytype.Block, error) {
	return c.sdk.GetBlockByHeight(height)
}

func (c *PolyClient) GetCurrentBlockHeight() (uint32, error) {
	return c.sdk.GetCurrentBlockHeight()
}

func GetBookeeper(block *polytype.Block) ([]keypair.PublicKey, error) {
	info := new(vconfig.VbftBlockInfo) //&vconfig.VbftBlockInfo{}
	info.NewChainConfig = new(vconfig.ChainConfig)
	if err := json.Unmarshal(block.Header.ConsensusPayload, info); err != nil {
		return nil, fmt.Errorf("failed to unmarshal consensus payload, err: %s", err)
	}

	if info.NewChainConfig == nil {
		return nil, fmt.Errorf("new chain config is nil")
	}

	bookkeepers := make([]keypair.PublicKey, 0)
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

func (c *PolyClient) WaitPolyTx(hash polycm.Uint256) error {
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

func (c *PolyClient) removeAccount(accIndex int) {
	list := make([]*polysdk.Account, 0)
	for i, v := range c.accArr {
		if i != accIndex {
			list = append(list, v)
		}
	}
	c.accArr = list
}

func HashConvertUp(src polycm.Uint256) common.Hash {
	return common.BytesToHash(src[:])
}

func HashConvertDown(src common.Hash) (polycm.Uint256, error) {
	return polycm.Uint256ParseFromBytes(src[:])
}

// todo
func AddrConvertUp(src polycm.Address) common.Address {
	return common.BytesToAddress(src[:])
}

func AddrConvertDown(src common.Address) (polycm.Address, error) {
	return polycm.AddressFromHexString(src.Hex())
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
