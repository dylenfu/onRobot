package sdk

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/palettechain/palette/common"
	"github.com/palettechain/palette/crypto"
	"github.com/palettechain/palette/rpc"
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

func (c *Client) Reset(key *ecdsa.PrivateKey) *Client {
	c.Key = key
	return c
}

func dialNode(url string) *rpc.Client {
	client, err := rpc.Dial(url)
	if err != nil {
		panic(fmt.Sprintf("failed to dial geth rpc [%v]", err))
	}

	return client
}
