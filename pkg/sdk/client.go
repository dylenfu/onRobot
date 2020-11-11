package sdk

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rpc"
)

type Client struct {
	*rpc.Client
	url          string
	Key          *ecdsa.PrivateKey
	currentNonce uint64
}

func NewSender(url string, key *ecdsa.PrivateKey) *Client {
	return &Client{
		url:    url,
		Client: dialNode(url),
		Key:    key,
	}
}

func (c *Client) Url() string {
	return c.url
}

func (c *Client) Address() common.Address {
	return crypto.PubkeyToAddress(c.Key.PublicKey)
}

func dialNode(url string) *rpc.Client {
	client, err := rpc.Dial(url)
	if err != nil {
		panic(fmt.Sprintf("failed to dial geth rpc [%v]", err))
	}

	return client
}
