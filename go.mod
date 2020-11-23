module github.com/palettechain/onRobot

go 1.13

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/blang/semver v3.5.1+incompatible
	github.com/btcsuite/btcd v0.0.0-20171128150713-2e60448ffcc6
	github.com/btcsuite/btcutil v0.0.0-20190425235716-9e5f4b9a998d
	github.com/ethereum/go-ethereum v1.0.0
	github.com/hashicorp/golang-lru v0.5.4
	github.com/ipfs/go-log v1.0.4
	github.com/jinzhu/gorm v1.9.12
	github.com/scylladb/go-set v1.0.2
	github.com/stretchr/testify v1.6.1
	github.com/tyler-smith/go-bip39 v1.0.1-0.20181017060643-dbb3b84ba2ef
)

replace (
	github.com/coreos/etcd v0.0.1 => github.com/polynetwork/coreos-etcd v0.0.1
	github.com/coreos/go-semver v0.0.1 => github.com/polynetwork/coreos-semver v0.0.1
	github.com/coreos/go-systemd v0.0.1 => github.com/polynetwork/coreos-systemd v0.0.1
	github.com/coreos/pkg v0.0.1 => github.com/polynetwork/coreos-pkg v0.0.1
	github.com/ethereum/go-ethereum v1.0.0 => /Users/dylen/workspace/gohome/src/github.com/palettechain/palette
)
