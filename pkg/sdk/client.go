package sdk

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/rpc"
)

type Client struct {
	*rpc.Client
	Key *keystore.Key
}

func NewSender(url string, key *keystore.Key) *Client {
	return &Client{
		Client: dialNode(url),
		Key:    key,
	}
}

func dialNode(url string) *rpc.Client {
	client, err := rpc.Dial(url)
	if err != nil {
		panic(fmt.Sprintf("failed to dial geth rpc [%v]", err))
	}

	return client
}
