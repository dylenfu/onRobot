# onRobot

palette测试工具, 包含:
 * PLT token的相关测试
 * Ethereum 转账等简单工具型测试
 * Palette chain部署测试
 * 远程部署工具型测试
 * 跨链测试
 
#### 环境设置:
 * local: palette&ethereum&poly等链及relayer都在本地
 * test: 测试环境
 * prod: 主网环境
```bash
export ONROBOT=local
make prepare
```

## 使用方式
下载项目onRobot \
https://github.com/palettechain/onRobot.git

在cmd目录下找到需要运行的服务，比如robot，目录树如下:
```dtd
build/local/
├── cases
│   ├── AddValidators.json
│   ├── BindNFTAsset.json
│   ├── ChangePolyBookKeepers.json
│   ├── Consistency.json
│   ├── DelValidator.json
│   ├── Delegate.json
│   ├── Deposit.json
│   ├── DumpBlock.json
│   ├── ETH-ETH-Transfer.json
│   ├── ETH-PLT-Balance.json
│   ├── ETH-PLT-Mint-Admin.json
│   ├── ETH-PLT-Mint-Gov.json
│   ├── ETH-PLT-Transfer.json
│   ├── GetNonce.json
│   ├── GlobalParams.json
│   ├── NFT-Balance.json
│   ├── NFT-Deploy.json
│   ├── NFT-Lock.json
│   ├── NFT-Mint.json
│   ├── NFT-Transfer.json
│   ├── NFT-Unlock.json
│   ├── PLT-Approve.json
│   ├── PLT-Balance.json
│   ├── PLT-Burn.json
│   ├── PLT-Lock.json
│   ├── PLT-Mint.json
│   ├── PLT-Transfer.json
│   ├── PLT-UnLock.json
│   ├── PolyTx.json
│   ├── Proposal.json
│   ├── Reward.json
│   ├── RewardPeriod.json
│   ├── SetAssetUri.json
│   ├── ShowDelegate.json
│   ├── UpdateEccm.json
│   ├── evm1.json
│   ├── evm1.sol
│   ├── evm2.json
│   ├── evm2.sol
│   └── newpolynode.dat
├── clear_node.sh
├── config.json
├── eth_keystore
│   └── 0x83***0f
├── grep.sh
├── init_node.sh
├── keystore
│   ├── 0x11***fb
│   ├── ......
│   └── 0xf7***82
├── leveldb
│   ├── 000136.ldb
│   ├── ......
│   └── MANIFEST-000406
├── node0
├── ......
├── node7
├── poly_keystore
│   ├── wallet1.dat
│   ├── ......
│   └── wallet4.dat
├── remote_build.sh
├── remote_setup.sh
├── robot
├── setup
│   ├── genesis.json
│   ├── node0
│   │   └── nodekey
│   │── ......
│   ├── node9
│   │   └── nodekey
│   └── static-nodes.json
├── start_node.sh
└── stop_node.sh
```
其中，config.json是配置文件，`make prepare`会将config/local.json文件拷贝到build/local对应的工作目录下，cases目录下包含具体测试需要的参数

构建
```bash
make compile
```
运行
```bash
make robot t=demo
```
也支持批量测试
```bash
make robot t=name,totalSupply
```

## 测试用例
```dtd
demo                                

remoteBuild                                         // 远程构建: 拉取git代码，编译
remoteSetup                                         // scp上传setup目录到远程机器, 前提是在远程机器上已设置过pub key
	
initGenesis                                         // 初始化多个genesis节点
startGenesis                                        // 启动genesis节点
stopGenesis                                         // 关停genesis节点
clearGenesis                                        // 清空genesis节点所有数据，慎用
restartGenesis                                      // 重启genesis节点

initValidator                                       // 初始化多个validator节点
startValidator                                      // 启动validator节点
stopValidator                                       // 关停validator节点
clearValidator                                      // 清空validator节点，慎用
restartValidator                                    // 重启validator节点

init                                                // 初始化所有节点
start                                               // 启动所有节点
stop                                                // 关停节点
clear                                               // 清空所有节点数据，慎用
restart                                             // 重启所有节点
grep                                                // grep查看所有节点运行信息
	
blockNumber
nonce
consistency
deposit
	
totalSupply
name
decimal
adminBalance
governanceBalance
balanceOf
transfer
approve
	
addValidators
getValidators
reward
fakeReward
delegate
showDelegate
proposal
globalParams
spare
delValidators
period
stakeAmount
dumpBlock
	
	// palette side chain environment
polyHeight
plt-deploy-eccd
plt-deploy-eccm
plt-deploy-ccmp
plt-eccd-ownership
plt-eccm-ownership
plt-ccmp-ownership
plt-nft-proxy-ownership
plt-cross-chain-admin-ownership
plt-registerSideChain
plt-approveRegisterSideChain
plt-bind-plt-proxy
plt-bind-plt-asset
plt-plt-ccmp
plt-deploy-nft-proxy
plt-bind-nft-proxy
plt-bind-nft-asset
plt-nft-ccmp
plt-sync-plt-genesis
plt-sync-poly-genesis
plt-upgradeECCM
plt-changePaletteBookKeeper
plt-changePolyBookKeeper
plt-updateSideChain
plt-quitSideChain
plt-approveUpdateSideChain
plt-approveQuitSideChain

// ethereum side chain environment
eth-deploy-eccd
eth-deploy-eccm
eth-deploy-ccmp
eth-eccd-ownership
eth-eccm-ownership
eth-ccmp-ownership
eth-registerSideChain
eth-approveRegisterSideChain
eth-deploy-plt
eth-deploy-plt-proxy
eth-bind-plt-proxy
eth-bind-plt-asset
eth-plt-ccmp
eth-deploy-nft-asset
eth-deploy-nft-proxy
eth-nft-ccmp
eth-bind-nft-proxy
eth-bind-nft-asset
eth-sync-eth-genesis
eth-sync-poly-genesis
eth-plt-asset-ownership
eth-plt-proxy-ownership
eth-nft-proxy-ownership
eth-plt-mint-gov
eth-plt-mint-admin
eth-plt-total-supply
eth-plt-balance
eth-plt-transfer
eth-eth-transfer

// plt cross chain
plt-lock
plt-unlock

// nft
plt-deploy-nft-asset
nft-transfer
nft-balance
nft-token-owner
nft-set-uri

// nft cross chain
nft-lock
nft-unlock
```

## 配置文件及测试参数
```dtd
{
  "Environment":{
    "Remote":false,                                                     // 是否支持远程操作
    "LocalWorkspace":"/Users/**/software/crosschain/onRobot/build/",    // 本地工作目录
    "RemoteWorkspace":"/home/ubuntu/palette/",                          // 远程工作目录
    "NetworkID":101,                                                    // palette chain network id
    "LogLevel":4,                                                       // log等级
    "IpList":[                                                          // 远程工作机器ip列表

    ],
    "SSHPort":"22",                                                     // 远程通讯端口
    "RemoteGoPath":"",                                                  // 远程机器gopath    
    "NFTServer":""
  },
  "Network":{
    "NodeIndexStart":0,                                                 // palette节点下标，假设网络中共有10个节点，第一个节点的index为0，
    "GenesisNodeNumber":5,                                              // palette网络中，创世节点数量，如网络中总共10个节点，GenesisNodeNumber为5，则0~4这5个节点为创世节点
    "ValidatorsNumber":3                                                // palette网络中，validator节点数量，如网络中总共10个节点，ValidatorsNumber为3，则5~7这3个节点为validator节点
  },
  "DefaultPassphrase":"",                                               // 默认的palette账户密码，如为空则在后续测试中需要输入密码，密码会保存到leveldb    
  "Rpc": "http://127.0.0.1:22000",                                      // palette网络默认rpc访问地址
  "AdminAccount":"0xf3**5f",                                            // palette genesis文件中管理员账户
  "CrossChainAdminAccount":"0x83**0f",                                  // palette genesis文件中跨链管理员账户
  "BaseRewardPool":"0xa2**59",                                          // palette genesis文件中奖励池账户
  "Accounts":[                                                          // palette 后续测试账户列表
    "0x2c**f7",
    ......
    "0x7f**32"
  ],
  "GasLimit":2100000,                                                   // gasLimit默认值
  "DeployGasLimit":10000000000,                                         // 部署合约需要的gasLimit默认值
  "BlockPeriod":"7s",                                                   // palette出块时间，一般会大于正式网络出块时间1到2s
  "RewardEffectivePeriod":6,                                            // palette分润周期，一般会大于正式网络分润周期1到2个块
  "Nodes":[                                                             // 节点列表
    {
      "Index":0,                                                        // 节点在列表中下标
      "Address":"0xc0**ec",                                             // 节点地址, 如无需部署/管理节点，可以在此设置假地址，质押地址也是如此
      "NodeKey":"49e**e4",                                              // 节点私钥
      "StakeAccount":"0xb2**c8",                                        // 节点质押地址
      "Host":"127.0.0.1",                                               // 节点所在机器IP
      "RPCPort":"22000",                                                // 节点RPC通讯端口
      "P2PPort":"30300"                                                 // 节点P2P通讯端口
    },
    ......
    {
      "Index":10,
      "Address":"0x9c**5e",
      "NodeKey":"19**97",
      "StakeAccount":"0xc1**2f",
      "Host":"127.0.0.1",
      "RPCPort":"22010",
      "P2PPort":"30310"
    }
  ],
  "CrossChain":{
    "PolyAccountDefaultPassphrase":"4c**Qc",                            // poly账户密码, 必须填写，且多账户密码保持一致
    "PolyRPCAddress":"http://127.0.0.1:40336",                          // poly rpc地址
    "PaletteSideChainID":101,                                           // palette在poly上的侧链ID    
    "PaletteSideChainName":"palette",                                   // palette在poly上的侧链name
    "PaletteECCD":"0x51**8f",                                           // palette上部署的eccd合约地址，部署成功后，build/local/config.json文件中该字段会被修改，后续只需拷贝到config/local.json中即可
    "PaletteECCM":"0xa9**4f",                                           // palette上部署的eccm合约地址
    "PaletteCCMP":"0x2e**1f",                                           // palette上部署的ccmp合约地址
    "PaletteNFTProxy":"0x55**4c",                                       // palette上部署的NFT lock proxy合约地址
    "EthereumSideChainID":1,                                            // ethereum在poly上的侧链ID
    "EthereumSideChainName":"ethereum",                                 // ethereum在poly上的侧链name
    "EthereumECCD":"0xd6**fb",                                          // ethereum上部署的eccd合约地址   
    "EthereumECCM":"0xc6**ea",                                          // ethereum上部署的eccm合约地址
    "EthereumCCMP":"0xd9**b0",                                          // ethereum上部署的ccpm合约地址
    "EthereumPLTAsset":"0xbb**2c",                                      // ethereum上部署的PLT资产合约地址
    "EthereumPLTProxy":"0x8b**9e",                                      // ethereum上部署的PLT lock proxy合约地址
    "EthereumNFTProxy":"0x5e**e6",                                      // ethereum上部署的NFT lock proxy合约地址
    "EthereumRPCUrl":"http://127.0.0.1:8545",                           // ethereum rpc地址   
    "EthereumAccount":"0x83**0f",                                       // ethereum相关基本测试账户地址
    "EthereumAccountPassword":"",                                       // ethereum基本测试账户密码，可为空
    "EthereumOwner":"0x83**0f",                                         // ethereum合约部署owner地址
    "EthereumOwnerPassword":""
  },
  "FinalOwner":{
    "PaletteFinalOwner":"0x83**0f",                                     // palette合约所有权转移最终账户地址
    "EthereumFinalOwner":"0x94**2B"                                     // ethereum合约所有权转移最终账户地址
  }
}
```