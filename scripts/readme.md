## create palette network with scratch

    this document described how to create a palette network with scratch.

# 1.admin and test account
```bash
geth --datadir=admin account new
```

Public address of the key:   0xf3A9d42C01635A585f1721463842F8936075105F 
Path of the secret key file: admin/keystore/UTC--2020-09-11T02-29-49.024005000Z--f3a9d42c01635a585f1721463842f8936075105f 

Public address of the key:   0x99E2A19CB2D4698Ee2A040E953ea5014a65FC218
Path of the secret key file: admin/keystore/UTC--2020-09-14T06-26-51.546662000Z--99e2a19cb2d4698ee2a040e953ea5014a65fc218

Public address of the key:   0xeCce5F1346aFEe82990cccC52FE521005bD54ff0
Path of the secret key file: admin/keystore/UTC--2020-09-14T06-27-00.280341000Z--ecce5f1346afee82990cccc52fe521005bd54ff0

Public address of the key:   0x2fFff236ff085B4D468B14C7b7b9fa1974A3bF7d
Path of the secret key file: admin/keystore/UTC--2020-09-14T06-27-08.318469000Z--2ffff236ff085b4d468b14c7b7b9fa1974a3bf7d

# 2.mkdir node dir
```bash
mkdir node0 node1 node2 node3 node4

mkdir -p node0/data/geth
mkdir -p node1/data/geth
mkdir -p node2/data/geth
mkdir -p node3/data/geth
mkdir -p node4/data/geth
```

# 3.gen nodekeys, static-node.json, genesis.json
```bash
mkdir setup
cd setup
../istanbul-tools/build/bin/istanbul setup --num 5 --nodes --quorum --save --verbose
```

```xml
validators
{
	"Address": "0xc095448424a5ecd5ca7ccdadfaad127a9d7e88ec",
	"Nodekey": "49e26aa4d60196153153388a24538c2693d65f0010a3a488c0c4c2b2a64b2de4",
	"NodeInfo": "enode://44e509103445d5e8fd290608308d16d08c739655d6994254e413bc1a067838564f7a32ed8fed182450ec2841856c0cc0cd313588a6e25002071596a7363e84b6@0.0.0.0:30303?discport=0"
}
{
	"Address": "0xd47a4e56e9262543db39d9203cf1a2e53735f834",
	"Nodekey": "9fc1723cff3bc4c11e903a53edb3b31c57b604bfc88a5d16cfec6a64fbf3141c",
	"NodeInfo": "enode://3884de29148505a8d862992e5721767d4b47ff52ffab4c2d2527182d812a6d95d2049e00b7c5579ca7b86b3dba8c935e742d2dfde9ae16abb5e3265e33a6d472@0.0.0.0:30303?discport=0"
}
{
	"Address": "0x258af48e28e4a6846e931ddff8e1cdf8579821e5",
	"Nodekey": "4b0c9b9d685db17ac9f295cb12f9d7d2369f5bf524b3ce52ce424031cafda1ae",
	"NodeInfo": "enode://c07fb7d48eac559a2483e249d27841c18c7ce5dbbbf2796a6963cc9cef27cabd2e1bc9c456a83f0777a98dfd6e7baf272739b9e5f8febf0077dc09509c2dfa48@0.0.0.0:30303?discport=0"
}
{
	"Address": "0x8c09d936a1b408d6e0afaa537ba4e06c4504a0ae",
	"Nodekey": "cc69b13ca2c5cd4d76bb881f6ad18d93bd947042c0f3a7adc80bdd17dac68210",
	"NodeInfo": "enode://ecac0ebe7224cfd04056c940605a4a9d4cb0367cf5819bf7e5502bf44f68bdd471a6b215c733f4a4ab6a1b417ec18b2e382e83d2e1a4d7936b437e8c047b41f5@0.0.0.0:30303?discport=0"
}
{
	"Address": "0xbfb558f0dceb07fbb09e1c283048b551a4310921",
	"Nodekey": "5555ebb339d3d5ed1efbf0ca96f5b145134e5ce8044fec693558056d268776ae",
	"NodeInfo": "enode://b838fa2387beb3a56aed86e447309f8844cb208387c63af64ad740729b5c0a276d97dc5409622775eb2bbc5cd3f880b42efa07314d0f66d7e0f5e1c4d0cee3f3@0.0.0.0:30303?discport=0"
}



static-nodes.json
[
	"enode://44e509103445d5e8fd290608308d16d08c739655d6994254e413bc1a067838564f7a32ed8fed182450ec2841856c0cc0cd313588a6e25002071596a7363e84b6@0.0.0.0:30303?discport=0",
	"enode://3884de29148505a8d862992e5721767d4b47ff52ffab4c2d2527182d812a6d95d2049e00b7c5579ca7b86b3dba8c935e742d2dfde9ae16abb5e3265e33a6d472@0.0.0.0:30303?discport=0",
	"enode://c07fb7d48eac559a2483e249d27841c18c7ce5dbbbf2796a6963cc9cef27cabd2e1bc9c456a83f0777a98dfd6e7baf272739b9e5f8febf0077dc09509c2dfa48@0.0.0.0:30303?discport=0",
	"enode://ecac0ebe7224cfd04056c940605a4a9d4cb0367cf5819bf7e5502bf44f68bdd471a6b215c733f4a4ab6a1b417ec18b2e382e83d2e1a4d7936b437e8c047b41f5@0.0.0.0:30303?discport=0",
	"enode://b838fa2387beb3a56aed86e447309f8844cb208387c63af64ad740729b5c0a276d97dc5409622775eb2bbc5cd3f880b42efa07314d0f66d7e0f5e1c4d0cee3f3@0.0.0.0:30303?discport=0"
]



genesis.json
{
    "config": {
        "chainId": 10,
        "homesteadBlock": 0,
        "eip150Block": 0,
        "eip150Hash": "0x0000000000000000000000000000000000000000000000000000000000000000",
        "eip155Block": 0,
        "eip158Block": 0,
        "byzantiumBlock": 0,
        "constantinopleBlock": 0,
        "istanbul": {
            "epoch": 30000,
            "policy": 0,
            "ceil2Nby3Block": 0
        },
        "txnSizeLimit": 64,
        "maxCodeSize": 0,
        "isQuorum": true
    },
    "nonce": "0x0",
    "timestamp": "0x5f5b7d29",
    "extraData": "0x0000000000000000000000000000000000000000000000000000000000000000f8aff86994c095448424a5ecd5ca7ccdadfaad127a9d7e88ec94d47a4e56e9262543db39d9203cf1a2e53735f83494258af48e28e4a6846e931ddff8e1cdf8579821e5948c09d936a1b408d6e0afaa537ba4e06c4504a0ae94bfb558f0dceb07fbb09e1c283048b551a4310921b8410000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000c0",
    "gasLimit": "0xe0000000",
    "difficulty": "0x1",
    "mixHash": "0x63746963616c2062797a616e74696e65206661756c7420746f6c6572616e6365",
    "coinbase": "0x0000000000000000000000000000000000000000",
    "alloc": {
        "258af48e28e4a6846e931ddff8e1cdf8579821e5": {
            "balance": "0x446c3b15f9926687d2c40534fdb564000000000000"
        },
        "8c09d936a1b408d6e0afaa537ba4e06c4504a0ae": {
            "balance": "0x446c3b15f9926687d2c40534fdb564000000000000"
        },
        "bfb558f0dceb07fbb09e1c283048b551a4310921": {
            "balance": "0x446c3b15f9926687d2c40534fdb564000000000000"
        },
        "c095448424a5ecd5ca7ccdadfaad127a9d7e88ec": {
            "balance": "0x446c3b15f9926687d2c40534fdb564000000000000"
        },
        "d47a4e56e9262543db39d9203cf1a2e53735f834": {
            "balance": "0x446c3b15f9926687d2c40534fdb564000000000000"
        }
    },
    "number": "0x0",
    "gasUsed": "0x0",
    "parentHash": "0x0000000000000000000000000000000000000000000000000000000000000000"
}

```

Notice: 
. [x] modify ip and port in setup/static-nodes.json 
. [x] modify genesis.json, add admin public address in config 

# 4.copy setup files in nodes
```bash
cp setup/genesis.json node0
cp setup/genesis.json node1
cp setup/genesis.json node2
cp setup/genesis.json node3
cp setup/genesis.json node4

cp setup/static-nodes.json node0/data/
cp setup/static-nodes.json node1/data/
cp setup/static-nodes.json node2/data/
cp setup/static-nodes.json node3/data/
cp setup/static-nodes.json node4/data/

cp setup/0/nodekey node0/data/geth
cp setup/1/nodekey node1/data/geth
cp setup/2/nodekey node2/data/geth
cp setup/3/nodekey node3/data/geth
cp setup/4/nodekey node4/data/geth
```

# 5.init geth node
```bash
cd node0
geth --datadir data init genesis.json

cd ../node1/
geth --datadir data init genesis.json

cd ../node2/
geth --datadir data init genesis.json

cd ../node3/
geth --datadir data init genesis.json

cd ../node4/
geth --datadir data init genesis.json
```

# 6.start up all nodes
```bash
cd node0
PRIVATE_CONFIG=ignore nohup geth --datadir data --nodiscover --istanbul.blockperiod 5 --syncmode full --mine --minerthreads 1 --verbosity 5 --networkid 10 --rpc --rpcaddr 0.0.0.0 --rpcport 22000 --rpcapi admin,db,eth,debug,miner,net,shh,txpool,personal,web3,quorum,istanbul --emitcheckpoints --port 30300 2>>node.log &

cd ../node1
PRIVATE_CONFIG=ignore nohup geth --datadir data --nodiscover --istanbul.blockperiod 5 --syncmode full --mine --minerthreads 1 --verbosity 5 --networkid 10 --rpc --rpcaddr 0.0.0.0 --rpcport 22001 --rpcapi admin,db,eth,debug,miner,net,shh,txpool,personal,web3,quorum,istanbul --emitcheckpoints --port 30301 2>>node.log &

cd ../node2
PRIVATE_CONFIG=ignore nohup geth --datadir data --nodiscover --istanbul.blockperiod 5 --syncmode full --mine --minerthreads 1 --verbosity 5 --networkid 10 --rpc --rpcaddr 0.0.0.0 --rpcport 22002 --rpcapi admin,db,eth,debug,miner,net,shh,txpool,personal,web3,quorum,istanbul --emitcheckpoints --port 30302 2>>node.log &

cd ../node3
PRIVATE_CONFIG=ignore nohup geth --datadir data --nodiscover --istanbul.blockperiod 5 --syncmode full --mine --minerthreads 1 --verbosity 5 --networkid 10 --rpc --rpcaddr 0.0.0.0 --rpcport 22003 --rpcapi admin,db,eth,debug,miner,net,shh,txpool,personal,web3,quorum,istanbul --emitcheckpoints --port 30303 2>>node.log &

cd ../node4
PRIVATE_CONFIG=ignore nohup geth --datadir data --nodiscover --istanbul.blockperiod 5 --syncmode full --mine --minerthreads 1 --verbosity 5 --networkid 10 --rpc --rpcaddr 0.0.0.0 --rpcport 22004 --rpcapi admin,db,eth,debug,miner,net,shh,txpool,personal,web3,quorum,istanbul --emitcheckpoints --port 30304 2>>node.log &
```

## 7. other nodes(node6~node10)
```bash
istanbul-tools/build/bin/istanbul setup --num 5 --nodes --quorum --save --verbose
```

```xml
validators
{
	"Address": "0xad3bf5ed640cc72f37bd21d64a65c3c756e9c88c",
	"Nodekey": "018c71d5e3b245117ffba0975e46129371473c6a1d231c5eddf7a8364d704846",
	"NodeInfo": "enode://d0ecfd09db6b1e4f59da7ebde8f6c3ea3ed09f06f5190477ae4ee528ec692fa858b2c5b7a84cff373b487e2caa493d24ed88952c1058ca617712c3a6d26770f1@0.0.0.0:30303?discport=0"
}
{
	"Address": "0x03ff6beb65feb5da87ca1b5468b3e95da767255e",
	"Nodekey": "c8d3e5e3fbc72898d1b90dedff34d6043fcbaaadeecd0bcb211a05c7c9a33af7",
	"NodeInfo": "enode://3cb4089265d269bc69f00fb1867fe11b14326ca9ce62456699165a2569ceb7be424efad114cd09e83312e5c57a8082ef4d77f35527f361210fa5943b237831f3@0.0.0.0:30303?discport=0"
}
{
	"Address": "0xc191f60e7e3633f46d01557508ec817c4a7c724b",
	"Nodekey": "e0f5429b336cb2c803383d0ef39cb0a0003d4d701c96a2e7b15e468740ed72f7",
	"NodeInfo": "enode://346015f9efe9477c1175192750b8db173958b60abf6712ae38e3ba30d3fdd4b2ac07ef3d659d094764a801ce94326dc926dafa2b4b36025861926f576d354be7@0.0.0.0:30303?discport=0"
}
{
	"Address": "0x8b0c92a3380d3527a649dfe18aefaba57ed82785",
	"Nodekey": "c124e7f77166ee5cd4ba490b838db0ee251d9d5a7ce64cbb3cababf8ae99bd37",
	"NodeInfo": "enode://840c5b12e405c30225d15f5e0fe1f2b23f5fbd54be643a1548eb3f30659812abb33588214d783d2cfd770e0b3ab2e34029494629ab800bddaa3868ee7da461b9@0.0.0.0:30303?discport=0"
}
{
	"Address": "0x9cdaddaa2ad2e1b13e8de08b8f1e209e6be0885e",
	"Nodekey": "1992e51194f5e1179c759881df0a2343e4f5cfc06575578d51cc641895b3a197",
	"NodeInfo": "enode://cc9c114355c56fefa3bab61418ff3e97e02d50f27d721aaf4bc109c9f20b4e99a6eeb062ac320154394de65d6306080a8e5fb07a45960144a277990f33b7b3e7@0.0.0.0:30303?discport=0"
}


static-nodes.json
[
	"enode://d0ecfd09db6b1e4f59da7ebde8f6c3ea3ed09f06f5190477ae4ee528ec692fa858b2c5b7a84cff373b487e2caa493d24ed88952c1058ca617712c3a6d26770f1@0.0.0.0:30303?discport=0",
	"enode://3cb4089265d269bc69f00fb1867fe11b14326ca9ce62456699165a2569ceb7be424efad114cd09e83312e5c57a8082ef4d77f35527f361210fa5943b237831f3@0.0.0.0:30303?discport=0",
	"enode://346015f9efe9477c1175192750b8db173958b60abf6712ae38e3ba30d3fdd4b2ac07ef3d659d094764a801ce94326dc926dafa2b4b36025861926f576d354be7@0.0.0.0:30303?discport=0",
	"enode://840c5b12e405c30225d15f5e0fe1f2b23f5fbd54be643a1548eb3f30659812abb33588214d783d2cfd770e0b3ab2e34029494629ab800bddaa3868ee7da461b9@0.0.0.0:30303?discport=0",
	"enode://cc9c114355c56fefa3bab61418ff3e97e02d50f27d721aaf4bc109c9f20b4e99a6eeb062ac320154394de65d6306080a8e5fb07a45960144a277990f33b7b3e7@0.0.0.0:30303?discport=0"
]
```

# validators
```dtd
{
	"Address": "0xc095448424a5ecd5ca7ccdadfaad127a9d7e88ec",
	"Nodekey": "49e26aa4d60196153153388a24538c2693d65f0010a3a488c0c4c2b2a64b2de4",
	"NodeInfo": "enode://44e509103445d5e8fd290608308d16d08c739655d6994254e413bc1a067838564f7a32ed8fed182450ec2841856c0cc0cd313588a6e25002071596a7363e84b6@0.0.0.0:30303?discport=0"
}
{
	"Address": "0xd47a4e56e9262543db39d9203cf1a2e53735f834",
	"Nodekey": "9fc1723cff3bc4c11e903a53edb3b31c57b604bfc88a5d16cfec6a64fbf3141c",
	"NodeInfo": "enode://3884de29148505a8d862992e5721767d4b47ff52ffab4c2d2527182d812a6d95d2049e00b7c5579ca7b86b3dba8c935e742d2dfde9ae16abb5e3265e33a6d472@0.0.0.0:30303?discport=0"
}
{
	"Address": "0x258af48e28e4a6846e931ddff8e1cdf8579821e5",
	"Nodekey": "4b0c9b9d685db17ac9f295cb12f9d7d2369f5bf524b3ce52ce424031cafda1ae",
	"NodeInfo": "enode://c07fb7d48eac559a2483e249d27841c18c7ce5dbbbf2796a6963cc9cef27cabd2e1bc9c456a83f0777a98dfd6e7baf272739b9e5f8febf0077dc09509c2dfa48@0.0.0.0:30303?discport=0"
}
{
	"Address": "0x8c09d936a1b408d6e0afaa537ba4e06c4504a0ae",
	"Nodekey": "cc69b13ca2c5cd4d76bb881f6ad18d93bd947042c0f3a7adc80bdd17dac68210",
	"NodeInfo": "enode://ecac0ebe7224cfd04056c940605a4a9d4cb0367cf5819bf7e5502bf44f68bdd471a6b215c733f4a4ab6a1b417ec18b2e382e83d2e1a4d7936b437e8c047b41f5@0.0.0.0:30303?discport=0"
}
{
	"Address": "0xbfb558f0dceb07fbb09e1c283048b551a4310921",
	"Nodekey": "5555ebb339d3d5ed1efbf0ca96f5b145134e5ce8044fec693558056d268776ae",
	"NodeInfo": "enode://b838fa2387beb3a56aed86e447309f8844cb208387c63af64ad740729b5c0a276d97dc5409622775eb2bbc5cd3f880b42efa07314d0f66d7e0f5e1c4d0cee3f3@0.0.0.0:30303?discport=0"
}
{
    "Address": "0x6a708455c8777630aac9d1e7702d13f7a865b27c",
    "Nodekey": "3d9c828244d3b2da70233a0a2aea7430feda17bded6edd7f0c474163802a431c",
    "NodeInfo": "enode://f5135ae0853af71f017a8ecb68e720b729ab92c7123c686e75b7487d4a57ae07dec951380b356246366391ed6cf36f5bcaf39b20c1049ba4a436330406b7b60c@0.0.0.0:30303?discport=0"
}
{
	"Address": "0xad3bf5ed640cc72f37bd21d64a65c3c756e9c88c",
	"Nodekey": "018c71d5e3b245117ffba0975e46129371473c6a1d231c5eddf7a8364d704846",
	"NodeInfo": "enode://d0ecfd09db6b1e4f59da7ebde8f6c3ea3ed09f06f5190477ae4ee528ec692fa858b2c5b7a84cff373b487e2caa493d24ed88952c1058ca617712c3a6d26770f1@0.0.0.0:30303?discport=0"
}
{
	"Address": "0x03ff6beb65feb5da87ca1b5468b3e95da767255e",
	"Nodekey": "c8d3e5e3fbc72898d1b90dedff34d6043fcbaaadeecd0bcb211a05c7c9a33af7",
	"NodeInfo": "enode://3cb4089265d269bc69f00fb1867fe11b14326ca9ce62456699165a2569ceb7be424efad114cd09e83312e5c57a8082ef4d77f35527f361210fa5943b237831f3@0.0.0.0:30303?discport=0"
}
{
	"Address": "0xc191f60e7e3633f46d01557508ec817c4a7c724b",
	"Nodekey": "e0f5429b336cb2c803383d0ef39cb0a0003d4d701c96a2e7b15e468740ed72f7",
	"NodeInfo": "enode://346015f9efe9477c1175192750b8db173958b60abf6712ae38e3ba30d3fdd4b2ac07ef3d659d094764a801ce94326dc926dafa2b4b36025861926f576d354be7@0.0.0.0:30303?discport=0"
}
{
	"Address": "0x8b0c92a3380d3527a649dfe18aefaba57ed82785",
	"Nodekey": "c124e7f77166ee5cd4ba490b838db0ee251d9d5a7ce64cbb3cababf8ae99bd37",
	"NodeInfo": "enode://840c5b12e405c30225d15f5e0fe1f2b23f5fbd54be643a1548eb3f30659812abb33588214d783d2cfd770e0b3ab2e34029494629ab800bddaa3868ee7da461b9@0.0.0.0:30303?discport=0"
}
{
	"Address": "0x9cdaddaa2ad2e1b13e8de08b8f1e209e6be0885e",
	"Nodekey": "1992e51194f5e1179c759881df0a2343e4f5cfc06575578d51cc641895b3a197",
	"NodeInfo": "enode://cc9c114355c56fefa3bab61418ff3e97e02d50f27d721aaf4bc109c9f20b4e99a6eeb062ac320154394de65d6306080a8e5fb07a45960144a277990f33b7b3e7@0.0.0.0:30303?discport=0"
}
```
