package core

import (
	"bytes"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/contracts/native/plt"
	"github.com/ethereum/go-ethereum/contracts/native/utils"
	"github.com/palettechain/onRobot/config"
	"github.com/palettechain/onRobot/pkg/log"
	"github.com/palettechain/onRobot/pkg/poly"
	"github.com/palettechain/onRobot/pkg/sdk"
)

func PolyHeight() (succeed bool) {
	rpc := config.Conf.CrossChain.PolyRPCAddress
	polyValidators := config.Conf.CrossChain.LoadPolyAccountList()
	polyCli, err := poly.NewPolyClient(rpc, polyValidators)
	if err != nil {
		log.Errorf("failed to generate poly client, err: %s", err)
		return
	} else {
		log.Infof("generate poly client success!")
	}

	height, err := polyCli.GetCurrentBlockHeight()
	if err != nil {
		log.Error(err)
		return
	}

	log.Infof("%s current height %d", rpc, height)
	return true
}

// 注意: bindProxy&bindAsset&Lock三个测试都是基于palette-poly-palette的回路测试
type DeployContractParams struct {
	Abi    string `json:"Abi"`
	Object string `json:"Object"`
}

func UpgradeECCM() (succeed bool) {
	params := new(DeployContractParams)

	eccdAddr, _, ccmpAddr := config.Conf.CrossChain.CrossContractAddressList()

	chainID := uint64(config.Conf.Environment.NetworkID)
	if err := config.LoadParams("UpdateEccm.json", params); err != nil {
		log.Error(err)
		return
	}

	eccmAddr, _, err := deployContract(params.Abi, params.Object, eccdAddr, chainID)
	if err != nil {
		log.Errorf("failed to deploy test contract, err: %v", err)
		return
	}
	log.Infof("new eccm contract %s", eccmAddr.Hex())

	// eccm contract transfer ownership
	node := config.Conf.ValidatorNodes()[0]
	cli := sdk.NewSender(node.RPCAddr(), node.PrivateKey())

	// eccd contract transfer ownership
	{
		logsplit()
		log.Info("eccd transferOwnership")
		hash, err := cli.ECCDTransferOwnerShip(eccdAddr, eccmAddr)
		if err != nil {
			log.Error(err)
			return
		}
		wait(3)
		if err := admcli.DumpEventLog(hash); err != nil {
			log.Error(err)
			return
		}
	}

	// eccm contract transfer ownership
	{
		logsplit()
		log.Info("eccm transferOwnership")
		hash, err := cli.ECCMTransferOwnerShip(eccmAddr, ccmpAddr)
		if err != nil {
			log.Error(err)
			return
		}
		wait(3)
		if err := admcli.DumpEventLog(hash); err != nil {
			log.Error(err)
			return
		}
	}

	// pause eccmp
	{
		logsplit()
		hash, err := cli.PauseCCMP(ccmpAddr)
		if err != nil {
			log.Error(err)
			return
		}
		log.Infof("pause tx %s", hash)
		wait(3)
		if err := admcli.DumpEventLog(hash); err != nil {
			log.Error(err)
			return
		}
		log.Infof("pause success!")
	}

	// upgrade eccm
	{
		logsplit()
		hash, err := cli.UpgradeECCM(eccmAddr, ccmpAddr)
		if err != nil {
			log.Error(err)
			return
		}
		log.Infof("upgrade tx %s", hash.Hex())
		wait(3)
		if err := admcli.DumpEventLog(hash); err != nil {
			log.Error(err)
			return
		}
		log.Infof("upgrade success!")
	}

	// unpause eccmp
	{
		logsplit()
		hash, err := cli.UnPauseCCMP(ccmpAddr)
		if err != nil {
			log.Error(err)
			return
		}
		log.Infof("unpause tx %s", hash.Hex())
		wait(3)
		if err := admcli.DumpEventLog(hash); err != nil {
			log.Error(err)
			return
		}
		log.Infof("unpause success!")
	}

	// record contracts address
	{
		log.Infof(" {\n\teccd: %s\n\teccm: %s\n\tccmp: %s\n}", eccdAddr.Hex(), eccmAddr.Hex(), ccmpAddr.Hex())
		log.Info("record these address in config.json NOW!")
	}

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
	eccdAddr, _, err := deployContract(eccdParams.Abi, eccdParams.Object)
	if err != nil {
		log.Errorf("failed to deploy eccd contract, err: %v", err)
		return
	}
	wait(3)

	if err := config.LoadContract(eccmFileName, eccmParams); err != nil {
		log.Errorf("failed to load contract %s, err: %v", eccmFileName, err)
		return
	}
	crossChainID := uint64(config.Conf.CrossChain.CrossChainID)
	eccmAddr, _, err := deployContract(eccmParams.Abi, eccmParams.Object, eccdAddr, crossChainID)
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

	// eccd contract transfer ownership
	{
		logsplit()
		log.Info("eccd transferOwnership")
		hash, err := cli.ECCDTransferOwnerShip(eccdAddr, eccmAddr)
		if err != nil {
			log.Error(err)
			return
		}
		wait(3)
		if err := admcli.DumpEventLog(hash); err != nil {
			log.Error(err)
			return
		}
	}

	// eccm contract transfer ownership
	{
		logsplit()
		log.Info("eccm transferOwnership")
		hash, err := cli.ECCMTransferOwnerShip(eccmAddr, ccmpAddr)
		if err != nil {
			log.Error(err)
			return
		}
		wait(3)
		if err := admcli.DumpEventLog(hash); err != nil {
			log.Error(err)
			return
		}
	}

	log.Infof(" {\n\teccd: %s\n\teccm: %s\n\tccmp: %s\n}", eccdAddr.Hex(), eccmAddr.Hex(), ccmpAddr.Hex())
	log.Info("record these contract address in config.json NOW!")

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
	_, _, ccmp := config.Conf.CrossChain.CrossContractAddressList()
	log.Infof("ccmp contract addr %s", ccmp.Hex())

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
	proxy := config.Conf.CrossChain.PLTCrossChainProxy()
	sideChainID := uint64(config.Conf.CrossChain.SideChainID)

	// bind proxy
	{
		log.Infof("bind proxy...")
		tx, err := admcli.BindProxy(sideChainID, proxy)
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
		actual, err := admcli.GetBindProxy(sideChainID, "latest")
		if err != nil {
			log.Error(err)
			return
		}

		if !bytes.Equal(actual.Bytes(), proxy.Bytes()) {
			log.Errorf("wrong proxy: expect  %s but get %s", proxy.Hex(), actual.Hex())
			return
		} else {
			log.Infof("bind proxy success %s", proxy.Hex())
		}
	}

	return true
}

// 在palette native合约上记录以太坊erc20资产地址
func BindAsset() (succeed bool) {
	asset := config.Conf.CrossChain.PLTCrossChainAsset()
	sideChainID := uint64(config.Conf.CrossChain.SideChainID)

	// bind asset
	{
		log.Infof("bind asset...")
		tx, err := admcli.BindAsset(sideChainID, asset)
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
		actual, err := admcli.GetBindAsset(sideChainID, "latest")
		if err != nil {
			log.Error(err)
			return
		}
		if asset != actual {
			log.Errorf("asset err, expect %s, actual %s", asset.Hex(), actual.Hex())
			return
		} else {
			log.Infof("get asset %s success", actual.Hex())
		}
	}

	return true
}

func Lock() (succeed bool) {
	var params struct {
		AccountIndex int
		Amount       int
	}
	sideChainID := uint64(config.Conf.CrossChain.SideChainID)

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
	bindTo := common.HexToAddress(user) // lock to self
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

		hash, err := cli.Lock(sideChainID, bindTo, amount)
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
			log.Infof("balance before lock %d, after lock %d, the sub amount is %d",
				plt.PrintUPLT(balanceBeforeLock), plt.PrintUPLT(balanceAfterLock), plt.PrintUPLT(subAmount))
		}
	}

	// waiting for unlock
	{
		logsplit()
		log.Infof("unlock PLT...")

		var (
			balanceBeforeUnlock,
			balanceAfterUnlock *big.Int
		)
		for i := 0; i < 100; i++ {
			balance, err := cli.BalanceOf(bindTo, "latest")
			if err != nil {
				log.Error(err)
				return
			}
			if i == 0 {
				balanceBeforeUnlock = balance
				log.Infof("waiting for unlock")
			} else if balance.Cmp(balanceBeforeUnlock) > 0 {
				balanceAfterUnlock = balance
				subAmount := utils.SafeSub(balanceAfterUnlock, balanceBeforeUnlock)
				log.Infof("balance before unlock %d, after unlock %d, the sub amount is %d",
					plt.PrintUPLT(balanceBeforeUnlock), plt.PrintUPLT(balanceAfterUnlock), plt.PrintUPLT(subAmount))
				break
			}
			time.Sleep(3 * time.Second)
		}
	}

	// return plt
	{
		hash, err := cli.PLTTransfer(common.HexToAddress(config.Conf.AdminAccount), amount)
		if err != nil {
			log.Infof("transfer back PLT to admin err: %s", err)
			return true
		}
		_ = cli.DumpEventLog(hash)
		balance, err := cli.BalanceOf(userAddr, "latest")
		if err != nil {
			log.Infof("check balance after unlock err: %s", err)
		} else {
			log.Infof("balance after unlock %d", plt.PrintUPLT(balance))
		}
	}

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
	polyRPC := config.Conf.CrossChain.PolyRPCAddress
	polyValidators := config.Conf.CrossChain.LoadPolyAccountList()
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
		crossChainID := uint64(config.Conf.CrossChain.CrossChainID)
		if err := polyCli.SyncGenesisBlock(crossChainID, pltHeaderEnc); err != nil {
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

		_, eccmAddr, _ := config.Conf.CrossChain.CrossContractAddressList()
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
	polyRPC := config.Conf.CrossChain.PolyRPCAddress
	polyValidators := config.Conf.CrossChain.LoadPolyAccountList()
	polyCli, err := poly.NewPolyClient(polyRPC, polyValidators)
	if err != nil {
		log.Errorf("failed to generate poly client, err: %s", err)
		return
	} else {
		log.Infof("generate poly client success!")
	}

	eccd, _, _ := config.Conf.CrossChain.CrossContractAddressList()
	crossChainID := uint64(config.Conf.CrossChain.CrossChainID)
	if err := polyCli.RegisterSideChain(crossChainID, eccd); err != nil {
		log.Errorf("failed to register side chain, err: %s", err)
		return
	}

	log.Infof("register side chain %d eccd %s success", crossChainID, eccd.Hex())
	return true
}

func UpdateSideChain() (succeed bool) {
	polyRPC := config.Conf.CrossChain.PolyRPCAddress
	polyValidators := config.Conf.CrossChain.LoadPolyAccountList()
	polyCli, err := poly.NewPolyClient(polyRPC, polyValidators)
	if err != nil {
		log.Errorf("failed to generate poly client, err: %s", err)
		return
	} else {
		log.Infof("generate poly client success!")
	}

	eccd, _, _ := config.Conf.CrossChain.CrossContractAddressList()
	crossChainID := uint64(config.Conf.CrossChain.CrossChainID)
	if err := polyCli.UpdateSideChain(crossChainID, eccd); err != nil {
		log.Errorf("failed to update side chain, err: %s", err)
		return
	}

	log.Infof("update side chain %d eccd %s success", crossChainID, eccd.Hex())
	return true
}

func QuitSideChain() (succeed bool) {
	polyRPC := config.Conf.CrossChain.PolyRPCAddress
	polyValidators := config.Conf.CrossChain.LoadPolyAccountList()
	polyCli, err := poly.NewPolyClient(polyRPC, polyValidators)
	if err != nil {
		log.Errorf("failed to generate poly client, err: %s", err)
		return
	} else {
		log.Infof("generate poly client success!")
	}

	crossChainID := uint64(config.Conf.CrossChain.CrossChainID)
	if err := polyCli.QuitSideChain(crossChainID); err != nil {
		log.Errorf("failed to quit side chain, err: %s", err)
		return
	}

	log.Infof("quit side chain %d success", crossChainID)
	return true
}

func ApproveRegisterSideChain() (succeed bool) {
	polyRPC := config.Conf.CrossChain.PolyRPCAddress
	polyValidators := config.Conf.CrossChain.LoadPolyAccountList()
	polyCli, err := poly.NewPolyClient(polyRPC, polyValidators)
	if err != nil {
		log.Errorf("failed to generate poly client, err: %s", err)
		return
	} else {
		log.Infof("generate poly client success!")
	}

	crossChainID := uint64(config.Conf.CrossChain.CrossChainID)
	if err := polyCli.ApproveRegisterSideChain(crossChainID); err != nil {
		log.Errorf("failed to approve register side chain, err: %s", err)
		return
	}

	log.Infof("approve register side chain %d success", crossChainID)
	return true
}

func ApproveUpdateSideChain() (succeed bool) {
	polyRPC := config.Conf.CrossChain.PolyRPCAddress
	polyValidators := config.Conf.CrossChain.LoadPolyAccountList()
	polyCli, err := poly.NewPolyClient(polyRPC, polyValidators)
	if err != nil {
		log.Errorf("failed to generate poly client, err: %s", err)
		return
	} else {
		log.Infof("generate poly client success!")
	}

	crossChainID := uint64(config.Conf.CrossChain.CrossChainID)
	if err := polyCli.ApproveUpdateSideChain(crossChainID); err != nil {
		log.Errorf("failed to approve update side chain, err: %s", err)
		return
	}

	log.Infof("approve update side chain %d success", crossChainID)
	return true
}

func ApproveQuitSideChain() (succeed bool) {
	polyRPC := config.Conf.CrossChain.PolyRPCAddress
	polyValidators := config.Conf.CrossChain.LoadPolyAccountList()
	polyCli, err := poly.NewPolyClient(polyRPC, polyValidators)
	if err != nil {
		log.Errorf("failed to generate poly client, err: %s", err)
		return
	} else {
		log.Infof("generate poly client success!")
	}

	crossChainID := uint64(config.Conf.CrossChain.CrossChainID)
	if err := polyCli.ApproveQuitSideChain(crossChainID); err != nil {
		log.Errorf("failed to approve quit side chain, err: %s", err)
		return
	}

	log.Infof("approve quit side chain %d success", crossChainID)
	return true
}

func ChangePaletteBookKeepers() (succeed bool) {
	var params struct {
		InitAmount int
		NodeNumber int
	}

	if err := config.LoadParams("DelValidator.json", &params); err != nil {
		log.Error(err)
		return
	}

	// spare node
	if params.NodeNumber > len(config.Conf.SpareNodes()) {
		log.Errorf("node number out of range")
		return
	}
	nodes := config.Conf.SpareNodes()[0:params.NodeNumber]

	// init nodes
	{
		for _, node := range nodes {
			execInitNode(node)
		}
		time.Sleep(2 * time.Second)
	}

	// start node and sync blocks
	{
		for _, node := range nodes {
			execStartNode(node)
		}
		wait(5)
	}

	// check balance before stake
	checkBalance := func(node *config.Node, mark string) int {
		data, err := admcli.BalanceOf(node.StakeAddr(), "latest")
		if err != nil {
			log.Error("failed to check %s balance", node.NodeAddr().Hex())
			return 0
		}
		balance := plt.PrintUPLT(data)
		log.Infof("%s balance %s %d", node.NodeAddr().Hex(), mark, balance)
		return int(balance)
	}

	stakeAndDumpEvent := func(revoke bool) {
		stakeHashList := make([]common.Hash, 0)
		for _, node := range nodes {
			cli := sdk.NewSender(node.RPCAddr(), node.StakePrivateKey())
			stkAmt := plt.MultiPLT(params.InitAmount)
			if revoke {
				stkAmt = cli.GetStakeAmount(node.NodeAddr(), node.StakeAddr(), "latest")
			}

			hash, err := cli.Stake(node.NodeAddr(), node.StakeAddr(), stkAmt, revoke)
			if err != nil {
				log.Error("failed to stake for validator %s stake account %s amount %d", node.NodeAddr().Hex(), node.StakeAddr().Hex(), stkAmt)
				return
			}
			log.Infof("stake for validator, hash %s", hash.Hex())
			stakeHashList = append(stakeHashList, hash)
		}
		wait(2)
		if err := DumpHashList(stakeHashList, "stake"); err != nil {
			return
		}
	}

	checkStakeAmt := func(mark string) {
		for _, node := range nodes {
			data := admcli.GetStakeAmount(node.NodeAddr(), node.StakeAddr(), "latest")
			value := plt.PrintFPLT(utils.DecimalFromBigInt(data))
			log.Infof("check stake amount %f %s", value, mark)
		}
	}

	adminAddValidator := func(revoke bool) {
		hs := make([]common.Hash, 0)
		for _, node := range nodes {
			hash, err := admcli.AddValidator(node.NodeAddr(), node.StakeAddr(), revoke)
			if err != nil {
				log.Errorf("failed to add validator %s, err: %s", node.NodeAddr().Hex(), err)
				return
			}
			log.Infof("admin add validator %s success, tx hash %s", node.NodeAddr().Hex(), hash.Hex())
			hs = append(hs, hash)
		}
		wait(2)
		if err := DumpHashList(hs, "admin add validators"); err != nil {
			return
		}
	}

	// 1.deposit and dump event log
	{
		log.Infof("admin deposit to validator")
		hs := make([]common.Hash, 0)
		for _, node := range nodes {
			balance := checkBalance(node, "before deposit")
			if balance >= params.InitAmount {
				continue
			}
			addAmount := params.InitAmount - balance
			hash, err := admcli.PLTTransfer(node.StakeAddr(), plt.MultiPLT(addAmount))
			if err != nil {
				log.Errorf("failed to deposit to node %s, err: %s", node.NodeAddr().Hex(), err)
				return
			} else {
				log.Infof("admin deposit to %s %d PLT, hash %s", node.NodeAddr().Hex(), addAmount, hash.Hex())
			}
			hs = append(hs, hash)
		}
		wait(2)
		if err := DumpHashList(hs, "deposit for validator"); err != nil {
			log.Error(err)
			return
		}
	}

	// 2.stake and dump event log
	{
		logsplit()
		log.Infof("validators stake at block %d", admcli.GetBlockNumber())
		stakeAndDumpEvent(false)
		wait(2 * config.Conf.RewardEffectivePeriod)
		checkStakeAmt("after stake")
	}

	// 3.admin add validator
	{
		log.Infof("admin add validator at block %d", admcli.GetBlockNumber())
		adminAddValidator(false)
		wait(config.Conf.RewardEffectivePeriod + 2)
	}

	// 4. lock
	Lock()

	// 5.admin del validator
	{
		logsplit()
		log.Infof("admin del validator at block %d", admcli.GetBlockNumber())
		adminAddValidator(true)
		wait(config.Conf.RewardEffectivePeriod + 2)
	}

	// 6.revoke stake
	{
		log.Infof("revoking stake......")
		stakeAndDumpEvent(true)
		wait(config.Conf.RewardEffectivePeriod)
		checkStakeAmt("after revoke stake")
	}

	// 7.check balance after revoke stake
	{
		for _, node := range nodes {
			checkBalance(node, "after revoke stake")
		}
	}

	// 8. stop and clear nodes
	{
		for _, node := range nodes {
			execStopNode(node)
		}
		for _, node := range nodes {
			execClearNode(node)
		}
	}

	// 9. lock
	Lock()

	return true
}

func ChangePolyBookKeepers() (succeed bool) {
	node, err := config.Conf.CrossChain.LoadPolyTestCaseAccount("newpolynode.dat")
	if err != nil {
		log.Errorf("load new node account err: %s", err)
		return
	}

	// 1. get poly client
	polyRPC := config.Conf.CrossChain.PolyRPCAddress
	polyValidators := config.Conf.CrossChain.LoadPolyAccountList()
	cli, err := poly.NewPolyClient(polyRPC, polyValidators)
	if err != nil {
		log.Errorf("failed to generate poly client, err: %s", err)
		return
	} else {
		log.Infof("generate poly client success!")
	}

	// 2. register node
	if err := cli.RegNode(node); err != nil {
		log.Error(err)
		return
	} else {
		log.Infof("register node %s success", node.Address.ToBase58())
	}
	wait(5)

	// 3. lock
	Lock()

	// 4. quit node
	if err := cli.QuitNode(node); err != nil {
		log.Error(err)
		return
	} else {
		log.Infof("quit node %s success", node.Address.ToBase58())
	}
	wait(5)

	// 5. lock
	Lock()

	return true
}

func QuitNode() (succeed bool) {
	node, err := config.Conf.CrossChain.LoadPolyTestCaseAccount("newpolynode.dat")
	if err != nil {
		log.Errorf("load new node account err: %s", err)
		return
	}

	// 1. get poly client
	polyRPC := config.Conf.CrossChain.PolyRPCAddress
	polyValidators := config.Conf.CrossChain.LoadPolyAccountList()
	cli, err := poly.NewPolyClient(polyRPC, polyValidators)
	if err != nil {
		log.Errorf("failed to generate poly client, err: %s", err)
		return
	} else {
		log.Infof("generate poly client success!")
	}

	// 4. quit node
	if err := cli.QuitNode(node); err != nil {
		log.Error(err)
		return
	} else {
		log.Infof("quit node %s success", node.Address.ToBase58())
	}

	return true
}
