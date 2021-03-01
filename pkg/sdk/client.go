package sdk

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

type Client struct {
	*rpc.Client
	backend      *ethclient.Client
	url          string
	Key          *ecdsa.PrivateKey
	currentNonce uint64
}

func NewSender(url string, key *ecdsa.PrivateKey) *Client {
	cli := dialNode(url)
	return &Client{
		url:     url,
		Client:  cli,
		Key:     key,
		backend: ethclient.NewClient(cli),
	}
}

func (c *Client) Url() string {
	return c.url
}

func (c *Client) Address() common.Address {
	return PubKey2Address(c.Key.PublicKey)
}

func (c *Client) Reset(key *ecdsa.PrivateKey) *Client {
	c.Key = key
	return c
}

func PubKey2Address(pub ecdsa.PublicKey) common.Address {
	return crypto.PubkeyToAddress(pub)
}

func dialNode(url string) *rpc.Client {
	client, err := rpc.Dial(url)
	if err != nil {
		panic(fmt.Sprintf("failed to dial geth rpc [%v]", err))
	}

	return client
}
