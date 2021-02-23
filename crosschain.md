# cross chain environment preparing

## prepare on ethereum chain

1. deploy eccd, eccm, ccmp contracts and record in config.json
```bash
make robot t=eth-deploy-eccd
make robot t=eth-deploy-eccm
make robot t=eth-deploy-ccmp
```

2. transfer eccd ownership to eccm, transfer eccm ownship to ccmp
```bash
make robot t=eth-eccd-ownership
make robot t=eth-eccm-ownership
```

3. register ethereum as an new side chain on poly chain
```bash
make robot t=eth-registerSideChain
make robot t=eth-approveRegisterSideChain
```

4. sync genesis header
```bash
make robot t=eth-sync-eth-genesis
make robot t=eth-sync-poly-genesis
```

5. deploy plt asset
```bash
make robot t=eth-deploy-plt
```

6. deploy proxy contracts and set manager proxy, and record in config.json
```bash
make robot t=eth-deploy-plt-proxy
make robot t=eth-plt-ccmp

make robot t=eth-deploy-nft-proxy
make robot t=eth-nft-ccmp
```

7. mint plt from `plt asset contract` owner to `plt proxy`
```bash
make robot t=eth-plt-transfer
```

## prepare on palette chain
1. deploy eccd, eccm, ccmp contracts and record in config.json
```bash
make robot t=plt-deploy-eccd
make robot t=plt-deploy-eccm
make robot t=plt-deploy-ccmp
```

2. transfer eccd ownership to eccm, transfer eccm ownership to ccmp.
```bash
make robot t=plt-eccd-ownership
make robot t=plt-eccm-ownership
```

3. set plt manager proxy with ccmp
```bash
make robot t=plt-plt-ccmp
```

4. register side chain id to poly chain and approve it with 4 poly validators' wallet file.
```bash
make robot t=plt-registerSideChain
make robot t=plt-approveRegisterSideChain
```

5. deploy nft proxy
```bash
make robot t=plt-deploy-nft-proxy
```

6. set nft manager proxy with ccmp
```bash
make robot t=plt-nft-ccmp
```

7. sync palette header to palette chain and store poly book keepers in the palette chain
```bash
make robot t=plt-sync-plt-genesis
make robot t=plt-sync-poly-genesis
```

## bind proxies and PLT asset on ethereum chain
```bash
make robot t=eth-bind-plt-proxy
make robot t=eth-bind-nft-proxy
make robot t=eth-bind-plt-asset
```

## bind proxies and PLT asset on palette chain
```bash
make robot t=plt-bind-plt-proxy
make robot t=plt-bind-nft-proxy
make robot t=plt-bind-plt-asset
```

## deploy and bind nft asset on both of palette and ethereum
1. deploy nft contract on both of palette and ethereum, record them in cases/BindNFTAsset.json
```bash
make robot t=plt-deploy-nft-asset
make robot t=eth-deploy-nft-asset
```   

2. bind plt and nft asset
```bash
make robot t=plt-bind-nft-asset
make robot t=eth-bind-nft-asset
```