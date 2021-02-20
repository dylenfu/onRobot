# cross chain environment preparing

## prepare on ethereum chain

1. 在以太坊上部署PLT资产合约
需要注意的是，该合约owner应该是某特定账户，配置在config.json中的ethereumOwner
```bash
make robot t=eth-deploy-plt
```

2. deploy proxy contracts and set manager proxy, and record in config.json
使用第一步用到的ethereumOwner账户部署PLT和NFT的proxy合约，注意，这两个proxy不同于其他以太lock proxy合约, 需单独部署
```bash
make robot t=eth-deploy-plt-proxy
make robot t=eth-plt-ccmp

make robot t=eth-deploy-nft-proxy
make robot t=eth-nft-ccmp
```

## prepare on palette chain
1. deploy eccd, eccm, ccmp contracts and record in config.json
使用palette admin账户在palette链上部署eccd,eccm,ccmp合约
```bash
make robot t=plt-deploy-eccd
make robot t=plt-deploy-eccm
make robot t=plt-deploy-ccmp
```

2. transfer eccd ownership to eccm, transfer eccm ownership to ccmp.
使用palette admin账户转移eccd使用权到eccm, 转移eccm使用权到ccmp
```bash
make robot t=plt-ownership
```

3. set plt manager proxy with ccmp
使用palette admin账户在PLT(同时也是proxy)合约写入ccmp地址
```bash
make robot t=plt-plt-ccmp
```

4. deploy nft proxy
使用admin账户注册nft的特定proxy合约
```bash
make robot t=plt-deploy-nft-proxy
```

5. set nft manager proxy with ccmp
使用admin账户绑定ccmp和nft proxy
```bash
make robot t=plt-nft-ccmp
```

6. register side chain id to poly chain and approve it with 4 poly validators' wallet file.
使用poly账户注册palette侧链
```bash
make robot t=plt-registerSideChain
make robot t=plt-approveRegisterSideChain
```

7. sync palette header to palette chain and store poly book keepers in the palette chain
palette和poly chain同步区块头
```bash
make robot t=plt-syncGenesis
```

## bind proxies and PLT asset on ethereum chain
使用ethereumOwner账户在以太坊上绑定plt、nft各自对应的proxy以及plt资产合约
```bash
make robot t=eth-bind-plt-proxy
make robot t=eth-bind-nft-proxy
make robot t=eth-bind-plt-asset
```

## bind proxies and PLT asset on palette chain
使用admin账户在palette上绑定plt、nft各自对应的proxy以及plt资产合约
```bash
make robot t=plt-bind-plt-proxy
make robot t=plt-bind-nft-proxy
make robot t=plt-bind-plt-asset
```

## mint plt from ethereum to palette
使用ethereumOwner账户跨链3.4亿到palette governance合约
```bash
make robot t=eth-plt-mint-gov
```
如果是在本地，还需要跨链3亿plt到palette admin账户，用于后续validators等测试
```bash
make robot t=eth-plt-mint-admin
```
这里注意，上述两个testcase都需要在case对应的json文件指定amount, 以防一次性出错.

## deploy and bind nft asset on both of palette and ethereum
跨链流程，使用任意账户在以太坊和palette上部署1对1的nft合约，并在两条链上相互绑定
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