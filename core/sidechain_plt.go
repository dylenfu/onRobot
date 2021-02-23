package core

import (
	"bytes"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/contracts/native/plt"
	"github.com/ethereum/go-ethereum/contracts/native/utils"
	"github.com/palettechain/onRobot/config"
	"github.com/palettechain/onRobot/pkg/log"
	"github.com/palettechain/onRobot/pkg/poly"
	"github.com/palettechain/onRobot/pkg/sdk"
	polyutils "github.com/polynetwork/poly/native/service/utils"
)

// 在palette合约部署成功后由三本合约:
// eccd: 管理epoch
// eccm: 管理跨链转账
// ccmp: 记录eccm地址及升级等
// 加入跨链事件从poly回到palette，事件流如下:
// relayer:
// 1. 执行palette eccm合约的verifyProofAndExecuteTx，这个方法会进入到palette native PLT合约的unlock方法
// 2. palette native PLT unlock 取出ccmp地址，并进入该合约查询eccm地址，比较从relayer过来的eccm地址与该地址是否匹配
// 3. 进入unlock资金逻辑

func PLTDeployECCD() (succeed bool) {
	eccd, err := ccAdmCli.DeployECCD()
	if err != nil {
		log.Errorf("deploy eccd on palette failed, err: %s", err.Error())
		return
	}

	log.Infof("deploy eccd %s on palette success!", eccd.Hex())

	if err := config.Conf.CrossChain.StorePaletteECCD(eccd); err != nil {
		log.Error("store palette eccd failed")
		return
	}

	return true
}

func PLTDeployECCM() (succeed bool) {
	eccd := config.Conf.CrossChain.PaletteECCD
	sideChainID := config.Conf.CrossChain.PaletteSideChainID
	eccm, err := ccAdmCli.DeployECCM(eccd, sideChainID)
	if err != nil {
		log.Errorf("deploy eccm on palette failed, err: %s", err.Error())
		return
	}

	log.Infof("deploy eccm %s on palette success!", eccm.Hex())

	if err := config.Conf.CrossChain.StorePaletteECCM(eccm); err != nil {
		log.Error("store palette eccm failed")
		return
	}

	return true
}

func PLTDeployCCMP() (succeed bool) {
	eccm := config.Conf.CrossChain.PaletteECCM
	ccmp, err := ccAdmCli.DeployCCMP(eccm)
	if err != nil {
		log.Errorf("deploy ccmp on palette failed, err: %s", err.Error())
		return
	}

	log.Infof("deploy ccmp %s on palette success!", ccmp.Hex())

	if err := config.Conf.CrossChain.StorePaletteCCMP(ccmp); err != nil {
		log.Error("store palette ccmp failed")
		return
	}

	return true
}

func PLTTransferECCDOwnerShip() (succeed bool) {
	eccd := config.Conf.CrossChain.PaletteECCD
	eccm := config.Conf.CrossChain.PaletteECCM

	cur, _ := ccAdmCli.ECCDOwnership(eccd)
	if bytes.Equal(eccm.Bytes(), cur.Bytes()) {
		log.Infof("eccd %s owner is %s already", eccd.Hex(), eccm.Hex())
		return true
	}

	hash, err := ccAdmCli.ECCDTransferOwnerShip(eccd, eccm)
	if err != nil {
		log.Error(err)
		return
	}
	actual, err := ccAdmCli.ECCDOwnership(eccd)
	if err != nil {
		log.Error(err)
		return
	}
	if !bytes.Equal(eccm.Bytes(), actual.Bytes()) {
		log.Error("new owner %s != acutal %s", eccm.Hex(), actual.Hex())
		return
	}
	log.Infof("transfer eccd %s to eccm %s success! hash %s", eccd.Hex(), eccm.Hex(), hash.Hex())

	return true
}

func PLTTransferECCMOwnerShip() (succeed bool) {
	eccm := config.Conf.CrossChain.PaletteECCM
	ccmp := config.Conf.CrossChain.PaletteCCMP

	cur, _ := ccAdmCli.ECCMOwnership(eccm)
	if bytes.Equal(ccmp.Bytes(), cur.Bytes()) {
		log.Infof("eccm %s owner is %s already", eccm.Hex(), ccmp.Hex())
		return true
	}

	hash, err := ccAdmCli.ECCMTransferOwnerShip(eccm, ccmp)
	if err != nil {
		log.Error(err)
		return
	}
	actual, err := ccAdmCli.ECCMOwnership(eccm)
	if err != nil {
		log.Error(err)
		return
	}
	if !bytes.Equal(ccmp.Bytes(), actual.Bytes()) {
		log.Error("new owner %s != acutal %s", ccmp.Hex(), actual.Hex())
		return
	}
	log.Infof("transfer eccm %s to ccmp %s success! hash %s", eccm.Hex(), ccmp.Hex(), hash.Hex())

	return true
}

func PLTTransferCCMPOwnerShip() (succeed bool) {
	ccmp := config.Conf.CrossChain.PaletteCCMP
	newOwner := config.Conf.FinalOwner.PaletteFinalOwner

	cur, _ := ccAdmCli.CCMPOwnership(ccmp)
	if bytes.Equal(newOwner.Bytes(), cur.Bytes()) {
		log.Infof("ccmp %s owner is %s already", ccmp.Hex(), newOwner.Hex())
		return true
	}

	hash, err := ccAdmCli.CCMPTransferOwnerShip(ccmp, newOwner)
	if err != nil {
		log.Error(err)
		return
	}
	actual, err := ccAdmCli.CCMPOwnership(ccmp)
	if err != nil {
		log.Error(err)
		return
	}
	if !bytes.Equal(newOwner.Bytes(), actual.Bytes()) {
		log.Error("new owner %s != acutal %s", newOwner.Hex(), actual.Hex())
		return
	}
	log.Infof("transfer ccmp %s to new owner %s success! hash %s", ccmp.Hex(), actual.Hex(), hash.Hex())

	return true
}

func PLTTransferNFTProxyOwnership() (succeed bool) {
	proxy := config.Conf.CrossChain.PaletteNFTProxy
	newOwner := config.Conf.FinalOwner.PaletteFinalOwner

	cur, _ := ccAdmCli.NFTProxyOwnership(proxy)
	if bytes.Equal(proxy.Bytes(), cur.Bytes()) {
		log.Infof("nft proxy %s owner is %s already", proxy.Hex(), newOwner.Hex())
		return true
	}

	hash, err := ccAdmCli.TransferNFTProxyOwnership(proxy, newOwner)
	if err != nil {
		log.Error(err)
		return
	}
	actual, err := ccAdmCli.NFTProxyOwnership(proxy)
	if err != nil {
		log.Error(err)
		return
	}
	if !bytes.Equal(newOwner.Bytes(), actual.Bytes()) {
		log.Error("new owner %s != acutal %s", newOwner.Hex(), actual.Hex())
		return
	}
	log.Infof("transfer nft proxy %s to new owner %s success! hash %s", proxy.Hex(), newOwner.Hex(), hash.Hex())
	return true
}

func PLTTransferCrossChainAdminOwnership() (succeed bool) {
	oldOwner := config.Conf.CrossChain.PaletteCrossChainAdminAccount
	newOwner := config.Conf.FinalOwner.PaletteFinalOwner

	cur, _ := ccAdmCli.CrossChainAdminOwnership("latest")
	if bytes.Equal(newOwner.Bytes(), cur.Bytes()) {
		log.Infof("cross chain admin %s owner is %s already", oldOwner.Hex(), newOwner.Hex())
		return true
	}

	hash, err := ccAdmCli.TransferCrossChainAdminOwnership(newOwner)
	if err != nil {
		log.Error(err)
		return
	}

	actual, err := ccAdmCli.CrossChainAdminOwnership("latest")
	if err != nil {
		log.Error(err)
		return
	}
	if !bytes.Equal(newOwner.Bytes(), actual.Bytes()) {
		log.Error("new owner %s != acutal %s", newOwner.Hex(), actual.Hex())
		return
	}

	log.Infof("transfer cross chain admin %s to new owner %s success! hash %s", oldOwner.Hex(), newOwner.Hex(), hash.Hex())
	return true
}

func PLTSetCCMP() (succeed bool) {
	ccmp := config.Conf.CrossChain.PaletteCCMP

	cur, _ := ccAdmCli.GetPLTCCMP("latest")
	if bytes.Equal(ccmp.Bytes(), cur.Bytes()) {
		log.Infof("PLT proxy already managed by %s", ccmp.Hex())
		return true
	}

	hash, err := ccAdmCli.SetPLTCCMP(ccmp)
	if err != nil {
		log.Error(err)
		return
	}

	actual, err := ccAdmCli.GetPLTCCMP("latest")
	if err != nil {
		log.Error(err)
		return
	}
	if !bytes.Equal(actual.Bytes(), ccmp.Bytes()) {
		log.Errorf("set proxy manager failed, expect %s != actual %s", ccmp.Hex(), actual.Hex())
		return
	}

	log.Infof("set PLT ccmp success! hash %s", hash.Hex())
	return true
}

// 在palette native合约上记录以太坊localProxy地址,
// 这里我们将实现palette->poly->palette的循环，不走ethereum，那么proxy就直接是plt地址，
// asset的地址也是palette plt地址
func PLTBindPLTProxy() (succeed bool) {
	proxy := config.Conf.CrossChain.EthereumPLTProxy
	sideChainID := config.Conf.CrossChain.EthereumSideChainID

	cur, _ := ccAdmCli.GetBindPLTProxy(sideChainID, "latest")
	if bytes.Equal(proxy.Bytes(), cur.Bytes()) {
		log.Infof("PLT proxy already bound to by %s", proxy.Hex())
		return true
	}

	hash, err := ccAdmCli.BindPLTProxy(sideChainID, proxy)
	if err != nil {
		log.Error(err)
		return
	}

	actual, err := ccAdmCli.GetBindPLTProxy(sideChainID, "latest")
	if err != nil {
		log.Error(err)
		return
	}
	if !bytes.Equal(actual.Bytes(), proxy.Bytes()) {
		log.Errorf("bind PLT proxy failed, expect  %s != actual %s", proxy.Hex(), actual.Hex())
		return
	}

	log.Infof("bind PLT proxy to %s on palette success! hash %s", proxy.Hex(), hash.Hex())
	return true
}

// 在palette native合约上记录以太坊erc20资产地址
func PLTBindPLTAsset() (succeed bool) {
	asset := config.Conf.CrossChain.EthereumPLTAsset
	sideChainID := config.Conf.CrossChain.EthereumSideChainID

	cur, _ := ccAdmCli.GetBindPLTAsset(sideChainID, "latest")
	if bytes.Equal(asset.Bytes(), cur.Bytes()) {
		log.Infof("PLT asset already bound to by %s", asset.Hex())
		return true
	}

	hash, err := ccAdmCli.BindPLTAsset(sideChainID, asset)
	if err != nil {
		log.Error(err)
		return
	}

	actual, err := ccAdmCli.GetBindPLTAsset(sideChainID, "latest")
	if err != nil {
		log.Error(err)
		return
	}
	if !bytes.Equal(asset.Bytes(), actual.Bytes()) {
		log.Errorf("bind PLT asset err, expect %s != actual %s", asset.Hex(), actual.Hex())
		return
	}

	log.Infof("bind PLT asset to %s on palette success! hash %s", asset.Hex(), hash.Hex())
	return true
}

func PLTRegisterSideChain() (succeed bool) {
	polyRPC := config.Conf.CrossChain.PolyRPCAddress
	polyValidators := config.Conf.CrossChain.LoadPolyAccountList()
	polyCli, err := poly.NewPolyClient(polyRPC, polyValidators)
	if err != nil {
		log.Errorf("failed to generate poly client, err: %s", err)
		return
	} else {
		log.Infof("generate poly client success!")
	}

	crossChainID := config.Conf.CrossChain.PaletteSideChainID
	eccd := config.Conf.CrossChain.PaletteECCD
	router := polyutils.QUORUM_ROUTER
	name := config.Conf.CrossChain.PaletteSideChainName
	if err := polyCli.RegisterSideChain(crossChainID, eccd, router, name); err != nil {
		log.Errorf("failed to register side chain, err: %s", err)
		return
	}

	log.Infof("register side chain %d eccd %s success", crossChainID, eccd.Hex())
	return true
}

func PLTApproveRegisterSideChain() (succeed bool) {
	polyRPC := config.Conf.CrossChain.PolyRPCAddress
	polyValidators := config.Conf.CrossChain.LoadPolyAccountList()
	polyCli, err := poly.NewPolyClient(polyRPC, polyValidators)
	if err != nil {
		log.Errorf("failed to generate poly client, err: %s", err)
		return
	} else {
		log.Infof("generate poly client success!")
	}

	crossChainID := config.Conf.CrossChain.PaletteSideChainID
	if err := polyCli.ApproveRegisterSideChain(crossChainID); err != nil {
		log.Errorf("failed to approve register side chain, err: %s", err)
		return
	}

	log.Infof("approve register side chain %d success", crossChainID)
	return true
}

func PLTDeployNFTProxy() (succeed bool) {
	proxy, err := ccAdmCli.DeployNFTProxy()
	if err != nil {
		log.Errorf("deploy NFT proxy on palette failed, err: %s", err.Error())
		return
	}

	log.Infof("deploy NFT proxy %s on palette success!", proxy.Hex())

	if err := config.Conf.CrossChain.StorePaletteNFTProxy(proxy); err != nil {
		log.Error("store palette nft proxy failed")
		return
	}

	return true
}

func PLTBindNFTProxy() (succeed bool) {
	localLockproxy := config.Conf.CrossChain.PaletteNFTProxy
	targetLockProxy := config.Conf.CrossChain.EthereumNFTProxy
	targetSideChainID := config.Conf.CrossChain.EthereumSideChainID

	cur, _ := ccAdmCli.GetBoundNFTProxy(localLockproxy, targetSideChainID)
	if bytes.Equal(targetLockProxy.Bytes(), cur.Bytes()) {
		log.Infof("NFT proxy %s already bound to by %s", localLockproxy.Hex(), targetLockProxy.Hex())
		return true
	}

	hash, err := ccAdmCli.BindNFTProxy(localLockproxy, targetLockProxy, targetSideChainID)
	if err != nil {
		log.Errorf("bind NFT proxy on palette failed, err: %s", err.Error())
		return
	}

	actual, err := ccAdmCli.GetBoundNFTProxy(localLockproxy, targetSideChainID)
	if err != nil {
		log.Error(err)
		return
	}
	if !bytes.Equal(targetLockProxy.Bytes(), actual.Bytes()) {
		log.Errorf("asset err, expect %s != actual %s", targetLockProxy.Hex(), actual.Hex())
		return
	}

	log.Infof("bind NFT proxy %s to %s on palette success! hash %s", localLockproxy.Hex(), targetLockProxy.Hex(), hash.Hex())
	return true
}

func PLTSetNFTCCMP() (succeed bool) {
	proxy := config.Conf.CrossChain.PaletteNFTProxy
	ccmp := config.Conf.CrossChain.PaletteCCMP

	cur, _ := ccAdmCli.GetNFTCCMP(proxy)
	if bytes.Equal(ccmp.Bytes(), cur.Bytes()) {
		log.Infof("NFT proxy %s already managed by %s", proxy.Hex(), ccmp.Hex())
		return true
	}

	hash, err := ccAdmCli.SetNFTCCMP(proxy, ccmp)
	if err != nil {
		log.Errorf("set ccmp on palette failed, err: %s", err.Error())
		return
	}

	actual, err := ccAdmCli.GetNFTCCMP(proxy)
	if err != nil {
		log.Error(err)
		return
	}
	if !bytes.Equal(ccmp.Bytes(), actual.Bytes()) {
		log.Errorf("asset err, expect %s, actual %s", ccmp.Hex(), actual.Hex())
		return
	}
	log.Infof("set NFT proxy manager %s for nft proxy %s on palette success! hash %s", actual.Hex(), proxy.Hex(), hash.Hex())
	return true
}

// 同步palette区块头到poly链上
// 1. 环境准备，palette cli: 使用任意palette签名者对应的cli, poly cli: 必须是poly验证节点的validators作为多签地址
// 2. 获取palette当前块高的区块头, 并使用json序列化为bytes
// 3. 使用poly cli同步第二步的bytes以及palette network id到poly native管理合约,
//	  这笔交易发出后等待poly当前块高超过交易块高, 作为落账的判断条件
// 4. 获取poly当前块高作为写入palette管理合约的genesis块高，获取对应的block，将block header及block book keeper
//    序列化，提交到palette管理合约
func PLTSyncGenesis() (succeed bool) {

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
	curr, hdr, err := ccAdmCli.GetCurrentBlockHeader()
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
		crossChainID := config.Conf.CrossChain.PaletteSideChainID
		if err := polyCli.SyncGenesisBlock(crossChainID, pltHeaderEnc); err != nil {
			log.Errorf("SyncEthGenesisHeader failed: %v", err)
			return
		}
		log.Infof("sync palette genesis header to poly success, txhash %s, block number %d",
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

		eccm := config.Conf.CrossChain.PaletteECCM
		txhash, err := ccAdmCli.InitGenesisBlock(eccm, headerEnc, bookeepersEnc)
		if err != nil {
			log.Errorf("failed to initGenesisBlock, err: %s", err)
			return
		} else {
			log.Infof("sync poly genesis header to palette success, txhash %s, block number %d",
				txhash.Hex(), gB.Header.Height)
		}
	}

	return true
}

func PLTBindNFTAsset() (succeed bool) {
	var params struct {
		EthereumNFTAsset common.Address
		PaletteNFTAsset  common.Address
	}
	if err := config.LoadParams("BindNFTAsset.json", &params); err != nil {
		log.Error(err)
		return
	}

	proxy := config.Conf.CrossChain.PaletteNFTProxy
	fromAsset := params.PaletteNFTAsset
	toAsset := params.EthereumNFTAsset
	targetSideChainID := config.Conf.CrossChain.EthereumSideChainID

	curAddr, _ := ccAdmCli.GetBoundNFTAsset(proxy, fromAsset, targetSideChainID)
	if curAddr != utils.EmptyAddress {
		if curAddr == toAsset {
			log.Infof("ethereum NFT asset %s bound already", toAsset.Hex())
			return
		} else {
			log.Infof("ethereum NFT asset %s bound != asset %s", curAddr.Hex(), toAsset.Hex())
		}
	}

	hash, err := ccAdmCli.BindNFTAsset(
		proxy,
		fromAsset,
		toAsset,
		targetSideChainID,
	)
	if err != nil {
		log.Errorf("bind NFT proxy on palette failed, err: %s", err.Error())
		return
	}

	actual, err := ccAdmCli.GetBoundNFTAsset(proxy, fromAsset, targetSideChainID)
	if err != nil {
		log.Error(err)
		return
	}
	if !bytes.Equal(toAsset.Bytes(), actual.Bytes()) {
		log.Errorf("asset err, expect %s, actual %s", toAsset.Hex(), actual.Hex())
		return
	}

	log.Infof("bind NFT asset %s to %s on palette success, hash %s", fromAsset.Hex(), toAsset.Hex(), hash.Hex())
	return true
}

// 注意: bindProxy&bindAsset&Lock三个测试都是基于palette-poly-palette的回路测试
type DeployContractParams struct {
	Abi    string `json:"Abi"`
	Object string `json:"Object"`
}

func PLTUpgradeECCM() (succeed bool) {
	params := new(DeployContractParams)
	eccdAddr := config.Conf.CrossChain.PaletteECCD
	ccmpAddr := config.Conf.CrossChain.PaletteCCMP

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

	// eccd contract transfer ownership
	{
		logsplit()
		log.Info("eccd transferOwnership")
		hash, err := ccAdmCli.ECCDTransferOwnerShip(eccdAddr, eccmAddr)
		if err != nil {
			log.Error(err)
			return
		}
		log.Infof("transfer eccd %s ownership to eccm %s success! hash %s", eccdAddr.Hex(), eccmAddr.Hex(), hash.Hex())
	}

	// eccm contract transfer ownership
	{
		logsplit()
		log.Info("eccm transferOwnership")
		hash, err := ccAdmCli.ECCMTransferOwnerShip(eccmAddr, ccmpAddr)
		if err != nil {
			log.Error(err)
			return
		}
		log.Infof("transfer eccm %s ownership to ccmp %s success! hash %s", eccmAddr.Hex(), ccmpAddr.Hex(), hash.Hex())
	}

	// pause eccmp
	{
		logsplit()
		hash, err := ccAdmCli.PauseCCMP(ccmpAddr)
		if err != nil {
			log.Error(err)
			return
		}
		log.Infof("pause tx %s", hash)
	}

	// upgrade eccm
	{
		logsplit()
		hash, err := ccAdmCli.UpgradeECCM(eccmAddr, ccmpAddr)
		if err != nil {
			log.Error(err)
			return
		}
		log.Infof("upgrade tx %s", hash.Hex())
	}

	// unpause eccmp
	{
		logsplit()
		hash, err := ccAdmCli.UnPauseCCMP(ccmpAddr)
		if err != nil {
			log.Error(err)
			return
		}
		log.Infof("unpause tx %s", hash.Hex())
	}

	// record contracts address
	{
		log.Infof(" {\n\teccd: %s\n\teccm: %s\n\tccmp: %s\n}", eccdAddr.Hex(), eccmAddr.Hex(), ccmpAddr.Hex())
		log.Info("record these address in config.json NOW!")
	}

	return true
}

func PLTUpdateSideChain() (succeed bool) {
	polyRPC := config.Conf.CrossChain.PolyRPCAddress
	polyValidators := config.Conf.CrossChain.LoadPolyAccountList()
	polyCli, err := poly.NewPolyClient(polyRPC, polyValidators)
	if err != nil {
		log.Errorf("failed to generate poly client, err: %s", err)
		return
	} else {
		log.Infof("generate poly client success!")
	}

	eccd := config.Conf.CrossChain.PaletteECCD
	crossChainID := config.Conf.CrossChain.PaletteSideChainID
	router := polyutils.QUORUM_ROUTER
	name := config.Conf.CrossChain.PaletteSideChainName
	if err := polyCli.UpdateSideChain(crossChainID, eccd, router, name); err != nil {
		log.Errorf("failed to update side chain, err: %s", err)
		return
	}

	log.Infof("update side chain %d eccd %s success", crossChainID, eccd.Hex())
	return true
}

func PLTQuitSideChain() (succeed bool) {
	polyRPC := config.Conf.CrossChain.PolyRPCAddress
	polyValidators := config.Conf.CrossChain.LoadPolyAccountList()
	polyCli, err := poly.NewPolyClient(polyRPC, polyValidators)
	if err != nil {
		log.Errorf("failed to generate poly client, err: %s", err)
		return
	} else {
		log.Infof("generate poly client success!")
	}

	crossChainID := config.Conf.CrossChain.PaletteSideChainID
	if err := polyCli.QuitSideChain(crossChainID); err != nil {
		log.Errorf("failed to quit side chain, err: %s", err)
		return
	}

	log.Infof("quit side chain %d success", crossChainID)
	return true
}

func PLTApproveUpdateSideChain() (succeed bool) {
	polyRPC := config.Conf.CrossChain.PolyRPCAddress
	polyValidators := config.Conf.CrossChain.LoadPolyAccountList()
	polyCli, err := poly.NewPolyClient(polyRPC, polyValidators)
	if err != nil {
		log.Errorf("failed to generate poly client, err: %s", err)
		return
	} else {
		log.Infof("generate poly client success!")
	}

	crossChainID := config.Conf.CrossChain.PaletteSideChainID
	if err := polyCli.ApproveUpdateSideChain(crossChainID); err != nil {
		log.Errorf("failed to approve update side chain, err: %s", err)
		return
	}

	log.Infof("approve update side chain %d success", crossChainID)
	return true
}

func PLTApproveQuitSideChain() (succeed bool) {
	polyRPC := config.Conf.CrossChain.PolyRPCAddress
	polyValidators := config.Conf.CrossChain.LoadPolyAccountList()
	polyCli, err := poly.NewPolyClient(polyRPC, polyValidators)
	if err != nil {
		log.Errorf("failed to generate poly client, err: %s", err)
		return
	} else {
		log.Infof("generate poly client success!")
	}

	crossChainID := config.Conf.CrossChain.PaletteSideChainID
	if err := polyCli.ApproveQuitSideChain(crossChainID); err != nil {
		log.Errorf("failed to approve quit side chain, err: %s", err)
		return
	}

	log.Infof("approve quit side chain %d success", crossChainID)
	return true
}

func PLTChangeBookKeepers() (succeed bool) {
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
		for _, node := range nodes {
			hash, err := admcli.AddValidator(node.NodeAddr(), node.StakeAddr(), revoke)
			if err != nil {
				log.Errorf("failed to add validator %s, err: %s", node.NodeAddr().Hex(), err)
				return
			}
			log.Infof("admin add validator %s success, tx hash %s", node.NodeAddr().Hex(), hash.Hex())
		}
	}

	// 1.deposit and dump event log
	{
		log.Infof("admin deposit to validator")
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
	PLTLock()

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
	PLTLock()

	return true
}

func PolyChangeBookKeepers() (succeed bool) {
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
	PLTLock()

	// 4. quit node
	if err := cli.QuitNode(node); err != nil {
		log.Error(err)
		return
	} else {
		log.Infof("quit node %s success", node.Address.ToBase58())
	}
	wait(5)

	// 5. lock
	PLTLock()

	return true
}

func PLTQuitNode() (succeed bool) {
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

func PLTDumpContractCode() (succeed bool) {
	var param struct {
		List []common.Address
	}
	if err := config.LoadParams("PLT-DumpContract.json", &param); err != nil {
		log.Error(err)
		return
	}
	for _, addr := range param.List {
		if err := admcli.DumpContractCode(addr); err != nil {
			log.Errorf("dum contract %s code err: %s", addr.Hex(), err)
		}
	}
	return true
}
