package sdk

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Rpc          string
	Admin        *Account
	TestAccounts []*Account
}

type Account struct {
	KeyFile    string
	Passphrase string
}

func GenerateConfig(dir string) *Config {
	cfg := &Config{}

	if _, err := toml.DecodeFile(dir, cfg); err != nil {
		panic(fmt.Sprintf("failed to decode config file: [%v]", err))
	}

	return cfg
}
