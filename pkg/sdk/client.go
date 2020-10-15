package sdk

import (
	"fmt"
	"io/ioutil"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/rpc"
)

type PaletteClient struct {
	*rpc.Client
	Admin        *keystore.Key
	TestAccounts []*keystore.Key
}

func NewPaletteClient(c *Config) *PaletteClient {
	client, err := rpc.Dial(c.Rpc)
	if err != nil {
		panic(fmt.Errorf("failed to dial geth rpc: [%v]", err))
	}

	admin := loadAccount(c.Admin)
	testAccounts := make([]*keystore.Key, len(c.TestAccounts))
	for i := 0; i < len(c.TestAccounts); i++ {
		testAccounts[i] = loadAccount(c.TestAccounts[i])
	}

	return &PaletteClient{
		Client:       client,
		Admin:        admin,
		TestAccounts: testAccounts,
	}
}

func loadAccount(c *Account) *keystore.Key {
	keyJson, err := ioutil.ReadFile(c.KeyFile)
	if err != nil {
		panic(fmt.Errorf("failed to read file: [%v]", err))
	}

	key, err := keystore.DecryptKey(keyJson, c.Passphrase)
	if err != nil {
		panic(fmt.Errorf("failed to decrypt keyjson: [%v]", err))
	}

	return key
}
