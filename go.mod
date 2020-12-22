module github.com/palettechain/onRobot

go 1.13

require (
	github.com/btcsuite/btcd v0.20.1-beta
	github.com/btcsuite/btcutil v1.0.2
	github.com/cmars/basen v0.0.0-20150613233007-fe3947df716e // indirect
	github.com/ethereum/go-ethereum v1.9.15
	github.com/jinzhu/gorm v1.9.12
	github.com/ontio/ontology-crypto v1.0.9
	github.com/polynetwork/eth-contracts v0.0.0-20200903021827-c9212e419943
	github.com/polynetwork/poly v0.0.1
	github.com/polynetwork/poly-go-sdk v0.0.0-20200817120957-365691ad3493
	github.com/stretchr/testify v1.6.1
	github.com/tyler-smith/go-bip39 v1.0.2
	launchpad.net/gocheck v0.0.0-20140225173054-000000000087 // indirect
)

replace (
	github.com/ethereum/go-ethereum v1.9.15 => /Users/dylen/workspace/gohome/src/github.com/palettechain/palette
	github.com/polynetwork/poly v0.0.1 => /Users/dylen/workspace/gohome/src/github.com/zouxyan/poly
)
