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
make robot t=eth-ownership
```

3. register ethereum as an new side chain on poly chain
```bash
make robot t=eth-registerSideChain
make robot t=eth-approveRegisterSideChain
```

4. deploy plt asset
```bash
make robot t=eth-deploy-plt
```

5. deploy proxy contracts and set manager proxy, and record in config.json
```bash
make robot t=eth-deploy-plt-proxy
make robot t=eth-plt-ccmp

make robot t=eth-deploy-nft-proxy
make robot t=eth-nft-ccmp
```

6. sync genesis header
```bash
make robot t=eth-syncGenesis
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
make robot t=plt-ownership
```

3. set plt manager proxy with ccmp
```bash
make robot t=plt-plt-ccmp
```

4. register side chain id to poly chain and approve it with 4 poly validators' wallet file.
```bash
make robot t=plt-registerSideChain,plt-approveRegisterSideChain
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
make robot t=plt-syncGenesis
```

## bind proxy on both of palette and ethereum
1. bind plt and nft proxy on ethereum chain
```bash
make robot t=eth-bind-plt-proxy
make robot t=eth-bind-nft-proxy
```

2. bind plt and nft proxy on palette chain
```bash
make robot t=plt-bind-plt-proxy
make robot t=plt-bind-nft-proxy
```

## bind asset on both of palette and ethereum
1. deploy nft contract on both of palette and ethereum, record them in cases/BindNFTAsset.json
```bash
make robot t=plt-deploy-nft-asset
make robot t=eth-deploy-nft-asset
```   

2. bind plt and nft asset
```bash
make robot t=plt-bind-plt-asset
make robot t=plt-bind-nft-asset
make robot t=eth-bind-plt-asset
make robot t=eth-bind-nft-asset
```