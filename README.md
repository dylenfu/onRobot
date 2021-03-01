# onRobot

palette测试工具, 包含:
 * PLT token的相关测试
 * Ethereum 转账等简单工具型测试
 * Palette chain部署测试
 * 远程部署工具型测试
 * 跨链测试
 
## 环境设置:
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
├── eth_keystore
│   └── 0x83***0f
├── keystore
│   ├── 0x11***fb
│   ├── ......
│   └── 0xf7***82
├── poly_keystore
│   ├── wallet1.dat
│   ├── ......
│   └── wallet4.dat
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
其中:
 * cases目录下包含具体测试需要的参数
 * eth_keystore包含测试需要的以太账户地址，这里统一规范keystore文件为0x开头
 * keystore包含测试需要的palette账户地址
 * poly_keystore包含测试需要的poly账户地址，用于侧链注册/授权
 * `make prepare`会将config/local.json以及所有scripts下的shell文件拷贝到build/local对应的工作目录下
 * `make compile`会将项目编译到该工作目录

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

## 远程构建
remoteBuild                                         // 远程构建: 拉取git代码，编译
remoteSetup                                         // scp上传setup目录到远程机器, 前提是在远程机器上已设置过pub key
	
## 节点管理部分
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

## palette链上常用查询	
blockNumber                                         // 查询palette当前高度
nonce                                               // 查看palette上某个账户当前nonce
	
## PLT部分
totalSupply                                         // 查询palette上PLT总供应量
name                                                // 查询palette上PLT对应合约名称
decimal                                             // 查询palette上PLT decimal精度
adminBalance                                        // 查询palette上管理员账户PLT余额    
governanceBalance                                   // 查询palette上治理合约地址PLT余额
balanceOf                                           // 查询palette上某个账户地址PLT余额
transfer                                            // 在palette上转账PLT
approve                                             // 在palette上授权PLT给某个账户
deposit                                             // palette管理员账户给所有palette测试账户地址充值PLT

# 治理部分	
addValidators                                       // 在palette上添加多个validators(质押&管理员添加节点)
getValidators                                       // 查询palette上所有validators   
reward                                              // 查看分润情况
delegate                                            // validator代理用户质押
showDelegate                                        // 查询validator代理用户质押
proposal                                            // validator提案修改全局参数
globalParams                                        // 查看全局参数
stakeAmount                                         // 查看质押数量
	
// palette 跨链部分
polyHeight                                          // 查看poly高度
plt-deploy-eccd                                     // 在palette上部署eccd合约    
plt-deploy-eccm                                     // 在palette上部署eccm合约
plt-deploy-ccmp                                     // 在palette上部署ccmp合约
plt-eccd-ownership                                  // 在palette上转移eccd合约所有权到eccm    
plt-eccm-ownership                                  // 在palette上转移eccm合约所有权到ccmp
plt-ccmp-ownership                                  // 在palette上转移ccmp合约所有权给最终账户
plt-nft-proxy-ownership                             // 在palette上转移NFT lock proxy合约所有权给最终账户        
plt-cross-chain-admin-ownership                     // 在palette上转移genesis的cross chain admin所有权给最终账户
plt-registerSideChain                               // 在poly上注册palette侧链
plt-approveRegisterSideChain                        // 在poly上授权palette侧链
plt-bind-plt-proxy                                  // 在palette上绑定以太上PLT lock proxy地址
plt-bind-plt-asset                                  // 在palette上绑定以太上PLT 资产合约
plt-plt-ccmp                                        // 在palette的PLT lock proxy设置ccmp地址
plt-deploy-nft-proxy                                // 在palette上部署NFT lock proxy合约
plt-bind-nft-proxy                                  // 在palette上绑定以太NFT lock proxy合约    
plt-bind-nft-asset                                  // 在palette上绑定以太NFT 资产合约    
plt-nft-ccmp                                        // 在palette的NFT lock proxy设置ccmp地址
plt-sync-plt-genesis                                // 同步palette区块头到poly
plt-sync-poly-genesis                               // 同步poly区块头到palette        

// ethereum side chain environment
eth-deploy-eccd                                     // 在以太上部署eccd合约
eth-deploy-eccm                                     // 在以太上部署eccm合约
eth-deploy-ccmp                                     // 在以太上部署ccmp合约
eth-eccd-ownership                                  // 在以太上转移eccd合约所有权到eccm    
eth-eccm-ownership                                  // 在以太上转移eccm合约所有权到ccmp
eth-ccmp-ownership                                  // 在以太上转移ccmp合约所有权给最终账户    
eth-registerSideChain                               // 在poly上注册以太侧链        
eth-approveRegisterSideChain                        // 在poly上授权以太侧链    
eth-deploy-plt                                      // 在以太上部署PLT资产合约
eth-deploy-plt-proxy                                // 在以太上部署PLT lock proxy合约
eth-bind-plt-proxy                                  // 在以太上绑定palette的PLT lock proxy合约        
eth-bind-plt-asset                                  // 在以太上绑定palette的PLT资产合约
eth-plt-ccmp                                        // 在以太上设置PLT lock proxy合约的ccmp地址
eth-deploy-nft-asset                                // 在以太上部署NFT 资产合约
eth-deploy-nft-proxy                                // 在以太上部署NFT lock proxy合约
eth-nft-ccmp                                        // 在以太上设置NFT lock proxy合约的ccmp地址    
eth-bind-nft-proxy                                  // 在以太上绑定palette NFT lock proxy合约
eth-bind-nft-asset                                  // 在以太上绑定palette NFT 资产合约
eth-sync-eth-genesis                                // 在以太上同步区块头到poly
eth-sync-poly-genesis                               // 在poly上同步区块头到以太坊
eth-plt-asset-ownership                             // 在以太上转移PLT资产合约所有权到最终账户
eth-plt-proxy-ownership                             // 在以太上转移PLT lock proxy合约所有权到最终账户
eth-nft-proxy-ownership                             // 在以太上转移NFT lock proxy合约所有权到最终账户
eth-plt-mint-gov                                    // 从以太上跨链3.4亿PLT到palette治理合约地址, 具体数字可以修改
eth-plt-mint-admin                                  // 从以太上跨链5亿PLT到palette管理员地址(主网没有这个必要，只测试网测试使用)
eth-plt-total-supply                                // 查询一台上PLT总供应量
eth-plt-balance                                     // 查询以太上某个账户的PLT余额
eth-plt-transfer                                    // 在以太上实现PLT转账
eth-eth-transfer                                    // 纯以太坊转账    

// plt cross chain
plt-lock                                            // PLT从palette跨链到以太
plt-unlock                                          // PLT从以太跨链到palette

// nft
plt-deploy-nft-asset                                // 在palette上部署NFT资产合约
nft-transfer                                        // 在palette上转账NFT
nft-balance                                         // 在palette上查询某个账户NFT余额
nft-token-owner                                     // 在palette上查询某个token的owner
nft-set-uri                                         // 在palette上设置某个NFT资产的base uri

// nft cross chain
nft-lock                                            // NFT从palette跨链到以太
nft-unlock                                          // NFT从以太上跨链到palette
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

##测试参数
一部分测试(非部署/管理节点或合约相关)需要用到单的测试参数，具体如下:

1. `nonce`: GetNonce.json
```dtd
{
  "Address": "0x6a**7c"
}
``` 
需要填充地址

2. `deposit`: Deposit.json
 ```dtd
{
  "Amount": 1000.0
}
```
浮点数，单位为PLT(在所有的测试中，均无需处理decimal，代码会整理成包含精度的bigInt)

3.`balanceOf`: PLT-Balance.json
```dtd
{
  "Owner": "0x0000000000000000000000000000000000000103",
  "BlockNum": "latest"
}
```
BlockNum为查询的块高度，hex表示或者为latest

4.`transfer`: PLT-Transfer.json
```dtd
{
  "From": "0xf3**5f",
  "To": "0x5c**44",
  "Amount": 100
}
```

5.`approve`: PLT-Approve.json
```dtd
{
  "Owner": "0x2cd**f7",
  "Spender": "0x2f**7d",
  "Amount": 120
}
```
spender为授权对象.

6.`addValidators`: AddValidators.json
```dtd
{
  "InitAmount": 50000000
}
```
InitAmount为质押量，测试过程中，如果账户余额不足，会从admin账户转账到多个validators(config.json中配置)，质押并等待，直到成功添加。

7.`reward`: Reward.json
```dtd
{
  "RewardBlocks": 12,
  "ExpectRewardPoolAmount": 24.0,
  "ExpectRewardAmountPerValidator": 32.0
}
```
分润测试，rewardBlocks为等待区块数，`ExpectRewardPoolAmount`为奖励池所得到的基础奖励，`ExpectRewardAmountPerValidator`为每个validator得到的奖励，该测试一般在addValidators后测试。

8.`delegate`: Delegate.json
```dtd
{
  "Fans": [
    {
      "Address": "0x4c**f5",
      "Amount": 1000,
      "NodeIndex": 5
    },
    {
      "Address": "0x2c**f7",
      "Amount": 500,
      "NodeIndex": 5
    }
  ],
  "WaitBlock": 12
}
```
`Address`为fans地址，`NodeIndex`为代理质押的节点下标，比如说，网络中8个节点，5个位genesis节点，3个为validators节点，node5则是第一个validator节点。<br>
如果用户余额不足，程序会从admin账户中转账到该测试地址.

9.`showDelegate`: ShowDelegate.json
```dtd
[
  {
    "Address": "0xDC**83",
    "NodeIndex": 5
  }
]
```
批量查询fans代理质押数量

10.`proposal`: Proposal.json
```dtd
{
  "ProposerNodeIndex": 5,
  "ProposalType": 2,
  "ProposalValue": 0,
  "VoteNodeIndexList": [
    6,
    7
  ]
}
```
`proposerNodeIndex`为提案节点下标，`proposalType`为提案类型: 1为mint nft手续费，2为部署NFT合约gas fee, 2位分润周期(该测试只测前两种类型).<br>
`proposalValue`为参数值，假设提案手续费费率为21.72%,则该值为2172, 系统传入参数后会* 10000. `voteNodeIndexList`为投票节点index列表

11.`globalParams`: GlobalParams.json
```dtd
{
  "ProposalType": 2
}
``` 
根据全局参数类型，查询全局参数.

12.`plt-deploy-nft-asset`: NFT-Deploy.json
```dtd
{
  "Name":"JpDigitalCat01",
  "Symbol":"JDC-01"
}
```
在palette上部署NFT合约，参数分别为NFT的名称和symbol

13.`eth-deploy-nft-asset`: 在以太上部署NFT合约，参数同上

14.`plt-bind-nft-asset`: BindNFTAsset.json
{
  "EthereumNFTAsset": "0x60**6b",
  "PaletteNFTAsset": "0x0000000000000000000000000000000000001002"
}
绑定nft资产合约到palette，`ethereumNFTAsset`为NFT在ethereum上的合约地址，`paletteNFTAsset`为NFT在palette上地址

15.`eth-bind-nft-asset`, 绑定nft资产合约到以太坊，参数同上

16.`eth-plt-mint-gov`: ETH-PLT-Mint-Gov.json
```dtd
{
  "Amount": 340000000
}
```
从以太上mint一定数量的PLT到palette治理合约

17.`eth-plt-mint-admin`: ETH-PLT-Mint-Admin.json
```dtd
{
  "Amount": 500000000
}
``` 
从以太上mint一定数量的PLT到palette的admin账户

18.`nft-transfer`: NFT-Transfer
```dtd
{
  "Asset": "0x0000000000000000000000000000000000001001",
  "TokenID": 3,
  "To": "0x2c**f7"
}
```
palette上的NFT转账, asset为NFT资产地址，tokenID为NFT id, `to`为目标地址

19.`nft-balance`: NFT-Balance.json
```dtd
{
  "Asset": "0x0000000000000000000000000000000000001001",
  "User": "0x6a**27c"
}
```
查询palette上某NFT资产上，某用户的资产余额

20.`nft-token-owner`: NFT-Owner.json
```dtd
{
  "Asset": "0x0000000000000000000000000000000000001002",
  "TokenID": 1
}
```
查询palette上NFT token所有者

21.`nft-set-uri`: SetAssetUri.json
```dtd
{
  "List": [
    "0x0000000000000000000000000000000000001001",
    "0x0000000000000000000000000000000000001002",
    "0x0000000000000000000000000000000000001003",
    "0x0000000000000000000000000000000000001004",
    "0x0000000000000000000000000000000000001005"
  ],
  "Storage": "http://127.0.0.1:10060/minio/"
}
```
批量设置NFT资产合约地址uri，format: http://127.0.0.1:10060/minio/1001(去除前面的0x00000...).

22.`plt-lock`: PLT-Lock.json
```dtd
{
  "From": "0x2c**f7",
  "To": "0x4c**f5",
  "Amount": 1
}
```
palette上`from`账户跨链一定量PLT资产到ethereum上`to`地址

23.`plt-unlock`: PLT-UnLock.json
```dtd
{
  "From": "0x4c**f5",
  "To": "0x2c**f7",
  "Amount": 1
}
```
以太上`from`账户跨链一定量PLT资产到palette上`to`地址

24.`nft-lock`: NFT-Lock.json
```dtd
{
  "From": "0x2c**f7",
  "To": "0x4c**f5",
  "PLTNFTAsset": "0x0000000000000000000000000000000000001002",
  "ETHNFTAsset": "0x60**6b",
  "TokenID": 1,
  "Uri": "cat1.jpeg"
}
```
palette上`from`账户跨链一定量NFT资产到ethereum上`to`地址, 如果token不存在，则在palette上mint该token, 过程中如果没有授权，程序会授权给proxy。

25.`nft-unlock`: NFT-UnLock.json
```dtd
{
  "From": "0x4c**f5",
  "To": "0x2c**f7",
  "PLTNFTAsset": "0x0000000000000000000000000000000000001002",
  "ETHNFTAsset": "0x6**6b",
  "TokenID": 1
}
```
以太上`from`账户跨链一定量PLT资产到palette上`to`地址