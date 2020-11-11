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

func DumpHashList(hashlist []common.Hash) error {
	for _, hash := range hashlist {
		if err := admcli.DumpEventLog(hash); err != nil {
			log.Errorf("failed to dump receipt, hash %s, [%v]", hash.Hex(), err)
			return err
		}
	}
	return nil
}
