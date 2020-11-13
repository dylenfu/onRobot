package core

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/palettechain/onRobot/pkg/log"
)

func PrivKey2Addr(pk *ecdsa.PrivateKey) common.Address {
	return crypto.PubkeyToAddress(pk.PublicKey)
}

func DumpHashList(hashlist []common.Hash, mark string) error {
	for _, hash := range hashlist {
		if err := admcli.DumpEventLog(hash); err != nil {
			log.Errorf("failed to dump receipt, hash %s, [%v]", hash.Hex(), err)
			return err
		}
	}
	log.Infof("dump %s event log success", mark)
	log.Info("------------------------------------------------------------------")
	return nil
}

func HasAddrs(src, dst []common.Address) bool {
	contain := func(addr common.Address, list []common.Address) bool {
		for _,v := range list {
			if addr == v {
				return true
			}
		}

		return false
	}

	for _, da := range dst {
		if !contain(da, src) {
			return false
		}
	}

	return true
}