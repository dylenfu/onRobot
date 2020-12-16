module github.com/palettechain/onRobot

go 1.13

require (
	github.com/btcsuite/btcd v0.20.1-beta
	github.com/btcsuite/btcutil v1.0.2
	github.com/ethereum/go-ethereum v1.9.13
	github.com/jinzhu/gorm v1.9.12
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/naoina/go-stringutil v0.1.0 // indirect
	github.com/steakknife/hamming v0.0.0-20180906055917-c99c65617cd3 // indirect
	github.com/stretchr/testify v1.6.1
	github.com/tyler-smith/go-bip39 v1.0.1-0.20181017060643-dbb3b84ba2ef
	github.com/polynetwork/poly v0.0.0-20201022033008-b0240c68a6bc
)

replace (
	github.com/coreos/etcd v0.0.1 => github.com/polynetwork/coreos-etcd v0.0.1
	github.com/coreos/go-semver v0.0.1 => github.com/polynetwork/coreos-semver v0.0.1
	github.com/coreos/go-systemd v0.0.1 => github.com/polynetwork/coreos-systemd v0.0.1
	github.com/coreos/pkg v0.0.1 => github.com/polynetwork/coreos-pkg v0.0.1
	github.com/ethereum/go-ethereum v1.9.13 => /Users/dylen/workspace/gohome/src/github.com/palettechain/palette
)
