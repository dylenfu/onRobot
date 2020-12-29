package core

import (
	"bytes"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/contracts/native/plt"
	"github.com/ethereum/go-ethereum/contracts/native/utils"
	"github.com/palettechain/onRobot/config"
	"github.com/palettechain/onRobot/pkg/log"
	"github.com/palettechain/onRobot/pkg/poly"
	"github.com/palettechain/onRobot/pkg/sdk"
)

type DeployContractParams struct {
	Abi    string `json:"Abi"`
	Object string `json:"Object"`
}

func DeployTest() (succeed bool) {
	params := new(DeployContractParams)

	eccd, _, _ ,err := config.ReadContracts()
	if err != nil {
		log.Error(err)
		return
	}

	chainID := uint64(config.Conf.Environment.NetworkID)
	if err := config.LoadParams("DeployTest.json", params); err != nil {
		log.Error(err)
		return
	}

	addr, _, err := deployContract(params.Abi, params.Object, eccd, chainID)
	if err != nil {
		log.Errorf("failed to deploy test contract, err: %v", err)
		return
	}

	wait(2)
	log.Infof("new contract %s", addr.Hex())

	return true
}

func DeployCrossChainContract() (succeed bool) {
	eccdFileName := "ECCD-raw.json"
	eccmFileName := "ECCM-raw.json"
	ecmpFileName := "ECCMP-raw.json"
	eccdParams := new(DeployContractParams)
	eccmParams := new(DeployContractParams)
	ecmpParams := new(DeployContractParams)

	if err := config.LoadContract(eccdFileName, eccdParams); err != nil {
		log.Errorf("failed to load contract %s, err: %v", eccdFileName, err)
		return
	}
	eccdAddr, eccd, err := deployContract(eccdParams.Abi, eccdParams.Object)
	if err != nil {
		log.Errorf("failed to deploy eccd contract, err: %v", err)
		return
	}
	wait(3)

	if err := config.LoadContract(eccmFileName, eccmParams); err != nil {
		log.Errorf("failed to load contract %s, err: %v", eccmFileName, err)
		return
	}
	otherChainID := uint64(config.Conf.Environment.NetworkID)
	eccmAddr, eccm, err := deployContract(eccmParams.Abi, eccmParams.Object, eccdAddr, otherChainID)
	if err != nil {
		log.Errorf("failed to deploy eccm contract, err: %v", err)
		return
	}
	wait(3)

	if err := config.LoadContract(ecmpFileName, ecmpParams); err != nil {
		log.Errorf("failed to load contract %s, err: %v", ecmpFileName, err)
		return
	}
	ccmpAddr, _, err := deployContract(ecmpParams.Abi, ecmpParams.Object, eccmAddr)
	if err != nil {
		log.Errorf("failed to deploy ecmp contract, err: %v", err)
		return
	}
	wait(3)

	node := config.Conf.ValidatorNodes()[0]
	cli := sdk.NewSender(node.RPCAddr(), node.PrivateKey())
	auth := bind.NewKeyedTransactor(node.PrivateKey())
	auth.GasLimit = 21000

	// eccd contract transfer ownership
	{
		logsplit()
		auth.Nonce = new(big.Int).SetUint64(cli.GetNonce(cli.Address().Hex()))
		log.Info("eccd transferOwnership")
		tx, err := eccd.Transact(auth, "transferOwnership", eccmAddr)
		if err != nil {
			log.Error(err)
			return
		}
		wait(3)
		if err := admcli.DumpEventLog(tx.Hash()); err != nil {
			log.Error(err)
			return
		}
	}

	// eccm contract transfer ownership
	{
		logsplit()
		log.Info("eccm transferOwnership")
		auth.Nonce = new(big.Int).SetUint64(cli.GetNonce(cli.Address().Hex()))
		tx, err := eccm.Transact(auth, "transferOwnership", ccmpAddr)
		if err != nil {
			log.Error(err)
			return
		}
		wait(3)
		if err := admcli.DumpEventLog(tx.Hash()); err != nil {
			log.Error(err)
			return
		}
	}

	// record contracts address
	{
		logsplit()
		if err := config.RecordContractAddress(eccdAddr, eccmAddr, ccmpAddr); err != nil {
			log.Error(err)
			return
		}
		log.Infof(" {\n\teccd: %s\n\teccm: %s\n\tccmp: %s\n}", eccdAddr.Hex(), eccmAddr.Hex(), ccmpAddr.Hex())
	}

	return true
}

// 在palette native plt合约mint一定量的PLT token到某个已经存在的用户地址
func Mint() (succeed bool) {
	var p struct {
		To    common.Address
		Value int
	}
	if err := config.LoadParams("Mint.json", &p); err != nil {
		log.Error(err)
		return
	}
	amount := plt.MultiPLT(p.Value)

	toBalanceBeforeTrans, err := admcli.BalanceOf(p.To, "latest")
	if err != nil {
		log.Error(err)
		return
	}
	tx, err := admcli.PLTMint(p.To, amount)
	if err != nil {
		log.Error(err)
		return
	}
	wait(2)
	if err := admcli.DumpEventLog(tx); err != nil {
		log.Error(err)
		return
	}

	toBalanceAfterTrans, err := admcli.BalanceOf(p.To, "latest")
	if err != nil {
		log.Error(err)
		return
	}

	if utils.SafeSub(toBalanceAfterTrans, toBalanceBeforeTrans).Cmp(amount) != 0 {
		log.Errorf("wrong mint value %s and should be %s",
			utils.SafeSub(toBalanceAfterTrans, toBalanceBeforeTrans), p.Value)
	}
	return true
}

// 在palette native plt合约烧毁合约PLT总供应量对应的amount
func Burn() (succeed bool) {
	var p struct {
		Value int
	}
	if err := config.LoadParams("Burn.json", &p); err != nil {
		log.Error(err)
		return
	}
	amount := plt.MultiPLT(p.Value)

	toBalanceBeforeTrans, err := admcli.BalanceOf(admcli.Address(), "latest")
	if err != nil {
		log.Error(err)
		return
	}
	if toBalanceBeforeTrans.Cmp(amount) == -1 {
		log.Error(err)
		return
	}

	tx, err := admcli.PLTBurn(amount)
	if err != nil {
		log.Error(err)
		return
	}
	wait(2)
	if err := admcli.DumpEventLog(tx); err != nil {
		log.Error(err)
		return
	}

	toBalanceAfterTrans, err := admcli.BalanceOf(admcli.Address(), "latest")
	if err != nil {
		log.Error(err)
		return
	}

	if utils.SafeSub(toBalanceBeforeTrans, toBalanceAfterTrans).Cmp(amount) != 0 {
		log.Errorf("wrong mint value %s and should be %s",
			utils.SafeSub(toBalanceAfterTrans, toBalanceBeforeTrans), p.Value)
	}

	return true
}

// 在palette合约部署成功后由三本合约:
// eccd: 管理epoch
// eccm: 管理跨链转账
// ccmp: 记录eccm地址及升级等
// 加入跨链事件从poly回到palette，事件流如下:
// relayer:
// 1. 执行palette eccm合约的verifyProofAndExecuteTx，这个方法会进入到palette native PLT合约的unlock方法
// 2. palette native PLT unlock 取出ccmp地址，并进入该合约查询eccm地址，比较从relayer过来的eccm地址与该地址是否匹配
// 3. 进入unlock资金逻辑
func SetCCMP() (succeed bool) {
	_, _, ccmp, err := config.ReadContracts()
	if err != nil {
		log.Error("read ccmp contract err")
		return
	} else {
		log.Infof("ccmp contract addr %s", ccmp.Hex())
	}

	tx, err := admcli.PLTSetCCMP(ccmp)
	if err != nil {
		log.Error(err)
		return
	}
	log.Infof("tx hash: %s", tx.Hex())
	wait(2)
	proxy, err := admcli.PLTGetCCMP("latest")
	if err != nil {
		log.Error(err)
		return
	}
	if !bytes.Equal(proxy.Bytes(), ccmp.Bytes()) {
		log.Errorf("wrong ccmp: should be %s but get %s", ccmp.Hex(), proxy.Hex())
		return
	}
	return true
}

// 在palette native合约上记录以太坊localProxy地址,
// 这里我们将实现palette->poly->palette的循环，不走ethereum，那么proxy就直接是plt地址，
// asset的地址也是palette plt地址
func BindProxy() (succeed bool) {
	var params struct {
		Proxy   common.Address
	}
	// todo: this is from palette to palette
	chainID := uint64(config.Conf.Environment.NetworkID)

	if err := config.LoadParams("BindProxy.json", &params); err != nil {
		log.Error(err)
		return
	}

	// bind proxy
	{
		log.Infof("bind proxy...")
		tx, err := admcli.BindProxy(chainID, params.Proxy)
		if err != nil {
			log.Error(err)
			return
		}
		log.Infof("tx hash: %s", tx.Hex())
		wait(2)

		if err := admcli.DumpEventLog(tx); err != nil {
			log.Error(err)
			return
		}
	}

	// get and compare proxy
	{
		log.Infof("get bind proxy...")
		proxy, err := admcli.GetBindProxy(chainID, "latest")
		if err != nil {
			log.Error(err)
			return
		}

		if !bytes.Equal(proxy.Bytes(), params.Proxy.Bytes()) {
			log.Errorf("wrong proxy: expect  %s but get %s", params.Proxy.Hex(), proxy.Hex())
			return
		} else {
			log.Infof("bind proxy success %s", proxy.Hex())
		}
	}

	return true
}

// 在palette native合约上记录以太坊erc20资产地址
func BindAsset() (succeed bool) {
	var params struct {
		Asset   common.Address
	}
	// todo: this is from palette to palette
	chainID := uint64(config.Conf.Environment.NetworkID)

	if err := config.LoadParams("BindAsset.json", &params); err != nil {
		log.Error(err)
		return
	}

	// bind asset
	{
		log.Infof("bind asset...")
		tx, err := admcli.BindAsset(chainID, params.Asset)
		if err != nil {
			log.Error(err)
			return
		}
		wait(2)
		if err := admcli.DumpEventLog(tx); err != nil {
			log.Error(err)
			return
		}
	}

	// get and compare asset
	{
		log.Infof("get bind asset...")
		asset, err := admcli.GetBindAsset(chainID, "latest")
		if err != nil {
			log.Error(err)
			return
		}
		if asset != params.Asset {
			log.Errorf("asset err, expect %s, actual %s", params.Asset.Hex(), asset.Hex())
			return
		} else {
			log.Infof("get asset %s success", asset.Hex())
		}
	}

	return true
}

func Lock() (succeed bool) {
	var params struct {
		AccountIndex int
		Proxy        common.Address
		Asset        common.Address
		BindTo       common.Address
		Amount       int
	}
	// todo: this is from palette to palette
	chainID := uint64(config.Conf.Environment.NetworkID)

	if err := config.LoadParams("Lock.json", &params); err != nil {
		log.Error(err)
		return
	}

	if params.AccountIndex > len(config.Conf.Accounts)-1 {
		log.Errorf("account index out of range")
		return
	}

	user := config.Conf.Accounts[params.AccountIndex]
	privKey := config.LoadAccount(user)
	baseUrl := config.Conf.Nodes[0].RPCAddr()
	userAddr := common.HexToAddress(user)
	cli := sdk.NewSender(baseUrl, privKey)
	amount := plt.MultiPLT(params.Amount)

	// prepare balance
	{
		logsplit()
		log.Infof("prepare test account balance...")
		hash, err := admcli.PLTTransfer(userAddr, amount)
		if err != nil {
			log.Error(err)
			return
		}
		wait(2)
		if err := admcli.DumpEventLog(hash); err != nil {
			log.Error(err)
			return
		}
	}

	// lock plt
	{
		logsplit()
		log.Infof("lock PLT...")
		balanceBeforeLock, err := cli.BalanceOf(userAddr, "latest")
		if err != nil {
			log.Error(err)
			return
		}

		hash, err := cli.Lock(chainID, params.BindTo, amount)
		if err != nil {
			log.Errorf("failed to call `lock` err: %v", err)
			return
		}
		wait(2)
		if err := cli.DumpEventLog(hash); err != nil {
			log.Error("failed to dump `lock` event hash %s, err: %v", hash.Hex(), err)
			return
		}

		balanceAfterLock, err := cli.BalanceOf(userAddr, "latest")
		if err != nil {
			log.Error(err)
			return
		}

		subAmount := utils.SafeSub(balanceBeforeLock, balanceAfterLock)
		if subAmount.Cmp(amount) != 0 {
			log.Errorf("balance before lock %d, after lock %d, the sub amount should be %d",
				plt.PrintUPLT(balanceBeforeLock), plt.PrintUPLT(balanceAfterLock), plt.PrintUPLT(amount))
			return
		} else {
			log.Infof("balance before lock %d, after lock %d, the sub amount should be %d",
				plt.PrintUPLT(balanceBeforeLock), plt.PrintUPLT(balanceAfterLock), plt.PrintUPLT(amount))
		}
	}

	// waiting for unlock
	{
		for i := 0; i < 50; i++ {
			balance, err := cli.BalanceOf(userAddr, "latest")
			if err != nil {
				log.Error(err)
				return
			}
			log.Infof("waiting for unlock, balance %d", plt.PrintUPLT(balance))
			time.Sleep(5 * time.Second)
		}
	}
	// unlock
	//{
	//	logsplit()
	//	log.Infof("unlock PLT...")
	//	balanceBeforeUnLock, err := cli.BalanceOf(userAddr, "latest")
	//	if err != nil {
	//		log.Error(err)
	//		return
	//	}
	//
	//	args := &plt.TxArgs{
	//		ToAssetHash: []byte{},
	//		ToAddress:   params.BindTo.Bytes(),
	//		Amount:      amount,
	//	}
	//	hash, err := cli.UnLock(args, params.BindTo, params.ChainID)
	//	if err != nil {
	//		log.Error(err)
	//		return
	//	}
	//	wait(2)
	//	if err := cli.DumpEventLog(hash); err != nil {
	//		log.Error(err)
	//		return
	//	}
	//
	//	balanceAfterUnLock, err := cli.BalanceOf(userAddr, "latest")
	//	if err != nil {
	//		log.Error(err)
	//		return
	//	}
	//
	//	subAmount := utils.SafeSub(balanceAfterUnLock, balanceBeforeUnLock)
	//	if subAmount.Cmp(amount) != 0 {
	//		log.Errorf("balance before unlock %d, after unlock %d, the sub amount should be %d",
	//			plt.PrintUPLT(balanceBeforeUnLock), plt.PrintUPLT(balanceAfterUnLock), plt.PrintUPLT(amount))
	//		return
	//	}
	//}
	return true
}

// 同步palette区块头到poly链上
// 1. 环境准备，palette cli: 使用任意palette签名者对应的cli, poly cli: 必须是poly验证节点的validators作为多签地址
// 2. 获取palette当前块高的区块头, 并使用json序列化为bytes
// 3. 使用poly cli同步第二步的bytes以及palette network id到poly native管理合约,
//	  这笔交易发出后等待poly当前块高超过交易块高, 作为落账的判断条件
// 4. 获取poly当前块高作为写入palette管理合约的genesis块高，获取对应的block，将block header及block book keeper
//    序列化，提交到palette管理合约
func SyncGenesis() (succeed bool) {

	// 1. prepare
	polyRPC := config.Conf.Poly.RPCAddress
	polyValidators := config.Conf.Poly.LoadPolyAccountList()
	polyCli, err := poly.NewPolyClient(polyRPC, polyValidators)
	if err != nil {
		log.Errorf("failed to generate poly client, err: %s", err)
		return
	} else {
		log.Infof("generate poly client success!")
	}

	// 2. get palette current block header
	logsplit()
	cli := admcli
	curr, hdr, err := cli.GetCurrentBlockHeader()
	if err != nil {
		log.Errorf("failed to get block header, err: %s", err)
		return
	}
	pltHeaderEnc, err := hdr.MarshalJSON()
	if err != nil {
		log.Errorf("marshal header failed, err: %s", err)
		return
	}
	log.Infof("get palette block header with current height %d, header %s", curr, hexutil.Encode(pltHeaderEnc))

	// 3. sync palette header to poly
	{
		logsplit()
		chainID := uint64(config.Conf.Environment.NetworkID)
		if err := polyCli.SyncGenesisBlock(chainID, pltHeaderEnc); err != nil {
			log.Errorf("SyncEthGenesisHeader failed: %v", err)
			return
		}
		log.Infof("successful to sync eth genesis header: txhash %s, block number %d",
			hdr.Hash().Hex(), hdr.Number.Uint64())
	}

	// 4. get poly block and assemble book keepers to header
	{
		logsplit()

		// `epoch` related with the poly validators changing,
		// we can set it as 0 if poly validators never changed on develop environment.
		var hasValidatorsBlockNumber uint32 = 0
		gB, err := polyCli.GetBlockByHeight(hasValidatorsBlockNumber)
		if err != nil {
			log.Errorf("failed to get block, err: %s", err)
			return
		}
		bookeepers, err := poly.GetBookeeper(gB)
		if err != nil {
			log.Errorf("failed to get bookeepers, err: %s", err)
			return
		}
		bookeepersEnc := poly.AssembleNoCompressBookeeper(bookeepers)
		headerEnc := gB.Header.ToArray()

		_, eccmAddr, _, err := config.ReadContracts()
		if err != nil {
			log.Errorf("failed to read eccm contract address, err: %s", err)
			return
		}
		txhash, err := cli.InitGenesisBlock(eccmAddr, headerEnc, bookeepersEnc)
		if err != nil {
			log.Errorf("failed to initGenesisBlock, err: %s", err)
			return
		} else {
			log.Infof("sync genesis header success, txhash %s", txhash.Hex())
		}
		_ = cli.DumpEventLog(txhash)
	}

	return true
}

// 获取并打印跨链事件
func GetProof() (succeed bool) {
	return true
}

func RegisterSideChain() (succeed bool) {
	polyRPC := config.Conf.Poly.RPCAddress
	polyValidators := config.Conf.Poly.LoadPolyAccountList()
	polyCli, err := poly.NewPolyClient(polyRPC, polyValidators)
	if err != nil {
		log.Errorf("failed to generate poly client, err: %s", err)
		return
	} else {
		log.Infof("generate poly client success!")
	}

	eccd, _, _, _ := config.ReadContracts()
	chainID := uint64(config.Conf.Environment.NetworkID)
	if err := polyCli.RegisterSideChain(chainID, eccd); err != nil {
		log.Errorf("failed to register side chain, err: %s", err)
		return
	}

	log.Infof("register side chain %d eccd %s success", chainID, eccd.Hex())
	return true
}

func UpdateSideChain() (succeed bool) {
	polyRPC := config.Conf.Poly.RPCAddress
	polyValidators := config.Conf.Poly.LoadPolyAccountList()
	polyCli, err := poly.NewPolyClient(polyRPC, polyValidators)
	if err != nil {
		log.Errorf("failed to generate poly client, err: %s", err)
		return
	} else {
		log.Infof("generate poly client success!")
	}

	eccd, _, _, _ := config.ReadContracts()
	chainID := uint64(config.Conf.Environment.NetworkID)
	if err := polyCli.UpdateSideChain(chainID, eccd); err != nil {
		log.Errorf("failed to update side chain, err: %s", err)
		return
	}

	log.Infof("update side chain %d eccd %s success", chainID, eccd.Hex())
	return true
}

func QuitSideChain() (succeed bool) {
	polyRPC := config.Conf.Poly.RPCAddress
	polyValidators := config.Conf.Poly.LoadPolyAccountList()
	polyCli, err := poly.NewPolyClient(polyRPC, polyValidators)
	if err != nil {
		log.Errorf("failed to generate poly client, err: %s", err)
		return
	} else {
		log.Infof("generate poly client success!")
	}

	chainID := uint64(config.Conf.Environment.NetworkID)
	if err := polyCli.QuitSideChain(chainID); err != nil {
		log.Errorf("failed to quit side chain, err: %s", err)
		return
	}

	log.Infof("quit side chain %d success", chainID)
	return true
}

func ApproveRegisterSideChain() (succeed bool) {
	polyRPC := config.Conf.Poly.RPCAddress
	polyValidators := config.Conf.Poly.LoadPolyAccountList()
	polyCli, err := poly.NewPolyClient(polyRPC, polyValidators)
	if err != nil {
		log.Errorf("failed to generate poly client, err: %s", err)
		return
	} else {
		log.Infof("generate poly client success!")
	}

	chainID := uint64(config.Conf.Environment.NetworkID)
	if err := polyCli.ApproveRegisterSideChain(chainID); err != nil {
		log.Errorf("failed to approve register side chain, err: %s", err)
		return
	}

	log.Infof("approve register side chain %d success", chainID)
	return true
}

func ApproveUpdateSideChain() (succeed bool) {
	polyRPC := config.Conf.Poly.RPCAddress
	polyValidators := config.Conf.Poly.LoadPolyAccountList()
	polyCli, err := poly.NewPolyClient(polyRPC, polyValidators)
	if err != nil {
		log.Errorf("failed to generate poly client, err: %s", err)
		return
	} else {
		log.Infof("generate poly client success!")
	}

	chainID := uint64(config.Conf.Environment.NetworkID)
	if err := polyCli.ApproveUpdateSideChain(chainID); err != nil {
		log.Errorf("failed to approve update side chain, err: %s", err)
		return
	}

	log.Infof("approve update side chain %d success", chainID)
	return true
}

func ApproveQuitSideChain() (succeed bool) {
	polyRPC := config.Conf.Poly.RPCAddress
	polyValidators := config.Conf.Poly.LoadPolyAccountList()
	polyCli, err := poly.NewPolyClient(polyRPC, polyValidators)
	if err != nil {
		log.Errorf("failed to generate poly client, err: %s", err)
		return
	} else {
		log.Infof("generate poly client success!")
	}

	chainID := uint64(config.Conf.Environment.NetworkID)
	if err := polyCli.ApproveQuitSideChain(chainID); err != nil {
		log.Errorf("failed to approve quit side chain, err: %s", err)
		return
	}

	log.Infof("approve quit side chain %d success", chainID)
	return true
}
