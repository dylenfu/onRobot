package sdk

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
)

type Client struct {
	*rpc.Client
	url          string
	Key          *keystore.Key
	currentNonce uint64
}

func NewSender(url string, key *keystore.Key) *Client {
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
	return c.Key.Address
}

func dialNode(url string) *rpc.Client {
	client, err := rpc.Dial(url)
	if err != nil {
		panic(fmt.Sprintf("failed to dial geth rpc [%v]", err))
	}

	return client
}
