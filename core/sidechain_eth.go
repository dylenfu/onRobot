package core

import (
	"bytes"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/contracts/native"
	"github.com/ethereum/go-ethereum/contracts/native/plt"
	"github.com/palettechain/onRobot/config"
	"github.com/palettechain/onRobot/pkg/log"
	"github.com/palettechain/onRobot/pkg/poly"
	polyutils "github.com/polynetwork/poly/native/service/utils"
)

///////////////////////////////////////////////////////
//
// deploy eccd, eccm, ccmp and transfer ownership
//
///////////////////////////////////////////////////////
func ETHDeployECCD() (succeed bool) {
	ethOwner := getEthereumCli(ethCTypeOwner)
	eccd, err := ethOwner.DeployECCDContract()
	if err != nil {
		log.Errorf("deploy eccd on ethereum failed, err: %s", err.Error())
		return
	} else {
		log.Infof("deploy eccd %s on ethereum success", eccd.Hex())
	}

	if err := config.Conf.CrossChain.StoreEthereumECCD(eccd); err != nil {
		log.Error("store ethereum eccd failed")
		return
	}

	return true
}

func ETHDeployECCM() (succeed bool) {
	eccd := config.Conf.CrossChain.EthereumECCD
	sideChainID := config.Conf.CrossChain.EthereumSideChainID
	// todo: make sure that there are no more contracts
	whiteList := []common.Address{
		config.Conf.CrossChain.EthereumPLTProxy,
		config.Conf.CrossChain.EthereumNFTProxy,
	}
	curPkBytes := config.Conf.CrossChain.LoadCurrentBookKeeperBytes()
	ethOwner := getEthereumCli(ethCTypeOwner)
	eccm, err := ethOwner.DeployECCMContract(eccd, sideChainID, whiteList, curPkBytes)
	if err != nil {
		log.Errorf("deploy eccm on ethereum failed, err: %s", err.Error())
		return
	} else {
		log.Infof("deploy eccm %s on ethereum success, eecd %s", eccm.Hex(), eccd.Hex())
	}

	if err := config.Conf.CrossChain.StoreEthereumECCM(eccm); err != nil {
		log.Error("store ethereum eccm failed")
		return
	}

	return true
}

func ETHDeployCCMP() (succeed bool) {
	eccm := config.Conf.CrossChain.EthereumECCM
	ethOwner := getEthereumCli(ethCTypeOwner)
	ccmp, err := ethOwner.DeployCCMPContract(eccm)
	if err != nil {
		log.Errorf("deploy ccmp on ethereum failed, err: %s", err.Error())
		return
	} else {
		log.Infof("deploy ccmp %s on ethereum success, eccm %s", ccmp.Hex(), eccm.Hex())
	}

	if err := config.Conf.CrossChain.StoreEthereumCCMP(ccmp); err != nil {
		log.Error("store ethereum ccmp failed")
		return
	}

	return true
}

func ETHTransferECCDOwnership() (succeed bool) {
	eccd := config.Conf.CrossChain.EthereumECCD
	eccm := config.Conf.CrossChain.EthereumECCM
	ethOwner := getEthereumCli(ethCTypeOwner)

	curOwner, _ := ethOwner.ECCDOwnership(eccd)
	if bytes.Equal(eccm.Bytes(), curOwner.Bytes()) {
		log.Infof("eccd %s owner is %s already", eccd.Hex(), eccm.Hex())
		return true
	}

	hash, err := ethOwner.TransferECCDOwnership(eccd, eccm)
	if err != nil {
		log.Errorf("transfer eccd ownership to eccm on ethereum failed, err: %s", err.Error())
		return
	}

	actual, err := ethOwner.ECCDOwnership(eccd)
	if err != nil {
		log.Errorf("get eccd new owner failed, err :%v", err)
		return
	}
	if !bytes.Equal(eccm.Bytes(), actual.Bytes()) {
		log.Errorf("transfer eccd ownership failed. expect %s != actual %s", eccm.Hex(), actual.Hex())
		return
	}

	log.Infof("transfer eccd %s ownership to eccm %s on ethereum success, tx %s", eccd.Hex(), eccm.Hex(), hash.Hex())
	return true
}

func ETHTransferECCMOwnership() (succeed bool) {
	eccm := config.Conf.CrossChain.EthereumECCM
	ccmp := config.Conf.CrossChain.EthereumCCMP
	ethOwner := getEthereumCli(ethCTypeOwner)

	curOwner, _ := ethOwner.ECCMOwnership(eccm)
	if bytes.Equal(ccmp.Bytes(), curOwner.Bytes()) {
		log.Infof("eccm %s ownership is %s already", eccm.Hex(), ccmp.Hex())
		return true
	}

	hash, err := ethOwner.TransferECCMOwnership(eccm, ccmp)
	if err != nil {
		log.Errorf("transfer eccm ownership to ccmp on ethereum failed, err: %s", err.Error())
		return
	}

	actual, err := ethOwner.ECCMOwnership(eccm)
	if err != nil {
		log.Errorf("get eccm new owner failed, err :%v", err)
		return
	}
	if !bytes.Equal(ccmp.Bytes(), actual.Bytes()) {
		log.Errorf("transfer eccm ownership failed. expect %s != actual %s", ccmp.Hex(), actual.Hex())
		return
	}

	log.Infof("transfer eccm %s ownership to ccmp %s on ethereum success, tx %s", eccm.Hex(), ccmp.Hex(), hash.Hex())
	return true
}

func ETHTransferCCMPOwnership() (succeed bool) {
	ccmp := config.Conf.CrossChain.EthereumCCMP
	newOwner := config.Conf.FinalOwner.EthereumFinalOwner
	ethOwner := getEthereumCli(ethCTypeOwner)

	curOwner, _ := ethOwner.CCMPOwnership(ccmp)
	if bytes.Equal(newOwner.Bytes(), curOwner.Bytes()) {
		log.Infof("ccmp %s ownership is %s already", ccmp.Hex(), newOwner.Hex())
		return true
	}

	hash, err := ethOwner.TransferECCMOwnership(ccmp, newOwner)
	if err != nil {
		log.Errorf("transfer ccmp ownership to new owner on ethereum failed, err: %s", err.Error())
		return
	}

	actual, err := ethOwner.CCMPOwnership(ccmp)
	if err != nil {
		log.Errorf("get ccmp new owner failed, err :%v", err)
		return
	}
	if !bytes.Equal(newOwner.Bytes(), actual.Bytes()) {
		log.Errorf("transfer ccmp ownership failed. expect %s != actual %s", ccmp.Hex(), actual.Hex())
		return
	}

	log.Infof("transfer ccmp %s ownership to new owner %s on ethereum success, tx %s", ccmp.Hex(), newOwner.Hex(), hash.Hex())
	return true
}

///////////////////////////////////////////////////////
//
// register side chain and approve
//
///////////////////////////////////////////////////////

func ETHRegisterSideChain() (succeed bool) {
	polyRPC := config.Conf.CrossChain.PolyRPCAddress
	polyValidators := config.Conf.CrossChain.LoadPolyAccountList()
	polyCli, err := poly.NewPolyClient(polyRPC, polyValidators)
	if err != nil {
		log.Errorf("failed to generate poly client, err: %s", err)
		return
	} else {
		log.Infof("generate poly client success!")
	}

	eccd := config.Conf.CrossChain.EthereumECCD
	router := polyutils.ETH_ROUTER
	name := config.Conf.CrossChain.EthereumSideChainName
	crossChainID := config.Conf.CrossChain.EthereumSideChainID
	if err := polyCli.RegisterSideChain(crossChainID, eccd, router, name); err != nil {
		log.Errorf("failed to register side chain, err: %s", err)
		return
	}

	log.Infof("register side chain %d eccd %s success", crossChainID, eccd.Hex())
	return true
}

func ETHApproveRegisterSideChain() (succeed bool) {
	polyRPC := config.Conf.CrossChain.PolyRPCAddress
	polyValidators := config.Conf.CrossChain.LoadPolyAccountList()
	polyCli, err := poly.NewPolyClient(polyRPC, polyValidators)
	if err != nil {
		log.Errorf("failed to generate poly client, err: %s", err)
		return
	} else {
		log.Infof("generate poly client success!")
	}

	crossChainID := config.Conf.CrossChain.EthereumSideChainID
	if err := polyCli.ApproveRegisterSideChain(crossChainID); err != nil {
		log.Errorf("failed to approve register side chain, err: %s", err)
		return
	}

	log.Infof("approve register side chain %d success", crossChainID)
	return true
}

///////////////////////////////////////////////////////
//
// 1. deploy plt asset
// 2. deploy plt proxy
// 3. bind plt asset with proxy
// 4. set plt ccmp
///////////////////////////////////////////////////////

func ETHDeployPLTAsset() (succeed bool) {
	ethOwner := getEthereumCli(ethCTypeOwner)
	pltAsset, err := ethOwner.DeployPLTAsset()
	if err != nil {
		log.Errorf("deploy PLT asset on ethereum failed, err: %s", err)
		return
	}

	log.Infof("deploy PLT asset %s on ethereum success!", pltAsset.Hex())

	if err := config.Conf.CrossChain.StoreEthereumPLTAsset(pltAsset); err != nil {
		log.Error("store ethereum plt asset failed")
		return
	}
	return true
}

func ETHDeployPLTProxy() (succeed bool) {
	ethOwner := getEthereumCli(ethCTypeOwner)
	proxy, err := ethOwner.DeployPLTLockProxy()
	if err != nil {
		log.Errorf("deploy PLT proxy on ethereum failed, err: %s", err)
		return
	} else {
		log.Infof("deploy PLT proxy %s on ethereum success!", proxy.Hex())
	}

	if err := config.Conf.CrossChain.StoreEthereumPLTProxy(proxy); err != nil {
		log.Error("store ethereum plt proxy failed")
		return
	}

	return true
}

func ETHBindPLTProxy() (succeed bool) {
	localLockProxy := config.Conf.CrossChain.EthereumPLTProxy
	targetLockProxy := common.HexToAddress(native.PLTContractAddress)
	targetSideChainID := config.Conf.CrossChain.PaletteSideChainID
	ethOwner := getEthereumCli(ethCTypeOwner)

	cur, _ := ethOwner.GetBoundPLTProxy(localLockProxy, targetSideChainID)
	if bytes.Equal(cur.Bytes(), targetLockProxy.Bytes()) {
		log.Infof("PLT proxy %s already bound to %s", localLockProxy.Hex(), targetLockProxy.Hex())
		return true
	}

	hash, err := ethOwner.BindPLTProxy(localLockProxy, targetLockProxy, targetSideChainID)
	if err != nil {
		log.Errorf("bind PLT proxy on ethereum failed, err: %s", err.Error())
		return
	}

	actual, err := ethOwner.GetBoundPLTProxy(localLockProxy, targetSideChainID)
	if err != nil {
		log.Error(err)
		return
	}
	if !bytes.Equal(targetLockProxy.Bytes(), actual.Bytes()) {
		log.Errorf("proxy bind failed, expect %s != actual %s", targetLockProxy.Hex(), actual.Hex())
		return
	}

	log.Infof("bind PLT proxy %s to %s on ethereum success, hash %s", localLockProxy.Hex(), targetLockProxy.Hex(), hash.Hex())
	return true
}

func ETHBindPLTAsset() (succeed bool) {
	localLockProxy := config.Conf.CrossChain.EthereumPLTProxy
	fromAsset := config.Conf.CrossChain.EthereumPLTAsset
	toAsset := common.HexToAddress(native.PLTContractAddress)
	toChainId := config.Conf.CrossChain.PaletteSideChainID
	ethOwner := getEthereumCli(ethCTypeOwner)

	cur, _ := ethOwner.GetBoundPLTAsset(localLockProxy, fromAsset, toChainId)
	if bytes.Equal(cur.Bytes(), toAsset.Bytes()) {
		log.Infof("PLT asset %s already bound to %s", fromAsset.Hex(), toAsset.Hex())
		return true
	}

	hash, err := ethOwner.BindPLTAsset(localLockProxy, fromAsset, toAsset, toChainId)
	if err != nil {
		log.Errorf("bind PLT asset on ethereum failed, err: %s", err.Error())
		return
	}

	actual, err := ethOwner.GetBoundPLTAsset(localLockProxy, fromAsset, toChainId)
	if err != nil {
		log.Error(err)
		return
	}
	if !bytes.Equal(toAsset.Bytes(), actual.Bytes()) {
		log.Errorf("bind plt asset on ethereum failed, expect %s != actual %s", toAsset.Hex(), actual.Hex())
		return
	}

	log.Infof("bind PLT asset %s to %s on ethereum success, hash %s", fromAsset.Hex(), toAsset.Hex(), hash.Hex())
	return true
}

func ETHSetPLTCCMP() (succeed bool) {
	proxy := config.Conf.CrossChain.EthereumPLTProxy
	ccmp := config.Conf.CrossChain.EthereumCCMP
	ethOwner := getEthereumCli(ethCTypeOwner)

	cur, _ := ethOwner.GetPLTCCMP(proxy)
	if bytes.Equal(cur.Bytes(), ccmp.Bytes()) {
		log.Infof("PLT proxy %s already managed to %s", proxy.Hex(), ccmp.Hex())
		return true
	}

	hash, err := ethOwner.SetPLTCCMP(proxy, ccmp)
	if err != nil {
		log.Errorf("register PLT proxy to ccmp on ethereum failed, err: %s", err.Error())
		return
	}

	actual, err := ethOwner.GetPLTCCMP(proxy)
	if err != nil {
		log.Error(err)
		return
	}
	if !bytes.Equal(ccmp.Bytes(), actual.Bytes()) {
		log.Errorf("set plt ccmp failed, expect %s != actual %s", ccmp.Hex(), actual.Hex())
		return
	}

	log.Infof("register PLT proxy %s to ccmp %s on ethereum success, tx %s", proxy.Hex(), ccmp.Hex(), hash.Hex())
	return true
}

///////////////////////////////////////////////////////
//
// 1. deploy new nft on ethereum
// 2. deploy nft proxy
// 3. set nft ccmp
// 4. bind nft asset
///////////////////////////////////////////////////////

func ETHDeployNFTAsset() (succeed bool) {
	var params struct {
		Name   string
		Symbol string
	}
	if err := config.LoadParams("NFT-Deploy.json", &params); err != nil {
		log.Error(err)
		return
	}
	proxy := config.Conf.CrossChain.EthereumNFTProxy
	ethOwner := getEthereumCli(ethCTypeOwner)
	contract, err := ethOwner.DeployNFT(proxy, params.Name, params.Symbol)
	if err != nil {
		log.Errorf("deploy new NFT contract on ethereum failed, err: %s", err.Error())
		return
	} else {
		log.Infof("deploy new NFT contract %s on ethereum success!", contract.Hex())
	}

	return true
}

func ETHDeployNFTProxy() (succeed bool) {
	ethOwner := getEthereumCli(ethCTypeOwner)
	proxy, err := ethOwner.DeployNFTLockProxy()
	if err != nil {
		log.Errorf("deploy nft lock proxy on ethereum failed, err: %s", err.Error())
		return
	} else {
		log.Infof("deploy NFT lock proxy %s on ethereum success!", proxy.Hex())
	}

	if err := config.Conf.CrossChain.StoreEthereumNFTProxy(proxy); err != nil {
		log.Error("save ethereum nft proxy failed")
		return
	}
	return true
}

func ETHSetNFTCCMP() (succeed bool) {
	proxy := config.Conf.CrossChain.EthereumNFTProxy
	ccmp := config.Conf.CrossChain.EthereumCCMP
	ethOwner := getEthereumCli(ethCTypeOwner)

	cur, _ := ethOwner.GetNFTCCMP(proxy)
	if bytes.Equal(cur.Bytes(), ccmp.Bytes()) {
		log.Infof("NFT proxy %s already managed to %s", proxy.Hex(), ccmp.Hex())
		return true
	}

	hash, err := ethOwner.SetNFTCCMP(proxy, ccmp)
	if err != nil {
		log.Errorf("register NFT proxy to ccmp on ethereum failed, err: %s", err.Error())
		return
	}

	actual, err := ethOwner.GetNFTCCMP(proxy)
	if err != nil {
		log.Error(err)
		return
	}
	if !bytes.Equal(ccmp.Bytes(), actual.Bytes()) {
		log.Errorf("register NFT proxy to ccmp failed, expect %s != actual %s", ccmp.Hex(), actual.Hex())
		return
	}

	log.Infof("register NFT proxy %s to ccmp %s on ethereum success, tx %s", proxy.Hex(), ccmp.Hex(), hash.Hex())
	return true
}

func ETHBindNFTProxy() (succeed bool) {
	localLockProxy := config.Conf.CrossChain.EthereumNFTProxy
	targetLockProxy := config.Conf.CrossChain.PaletteNFTProxy
	targetSideChainID := config.Conf.CrossChain.PaletteSideChainID
	ethOwner := getEthereumCli(ethCTypeOwner)

	cur, _ := ethOwner.GetBoundNFTProxy(localLockProxy, targetSideChainID)
	if bytes.Equal(cur.Bytes(), targetLockProxy.Bytes()) {
		log.Infof("NFT proxy %s already bound to %s", localLockProxy.Hex(), targetLockProxy.Hex())
		return true
	}

	hash, err := ethOwner.BindNFTProxy(localLockProxy, targetLockProxy, targetSideChainID)
	if err != nil {
		log.Errorf("bind NFT proxy on ethereum failed, err: %s", err.Error())
		return
	}

	actual, err := ethOwner.GetBoundNFTProxy(localLockProxy, targetSideChainID)
	if err != nil {
		log.Error(err)
		return
	}
	if !bytes.Equal(targetLockProxy.Bytes(), actual.Bytes()) {
		log.Errorf("bind NFT proxy to ccmp failed, expect %s != actual %s", targetLockProxy.Hex(), actual.Hex())
		return
	}

	log.Infof("bind NFT proxy %s to %s on ethereum success, tx %s", localLockProxy.Hex(), targetLockProxy.Hex(), hash.Hex())
	return true
}

func ETHBindNFTAsset() (succeed bool) {
	var params = struct {
		EthereumNFTAsset common.Address
		PaletteNFTAsset  common.Address
	}{}
	if err := config.LoadParams("BindNFTAsset.json", &params); err != nil {
		log.Error(err)
		return
	}

	proxy := config.Conf.CrossChain.EthereumNFTProxy
	fromAsset := params.EthereumNFTAsset
	toAsset := params.PaletteNFTAsset
	chainID := config.Conf.CrossChain.PaletteSideChainID
	ethOwner := getEthereumCli(ethCTypeOwner)

	cur, _ := ethOwner.GetBoundNFTAsset(proxy, fromAsset, chainID)
	if bytes.Equal(cur.Bytes(), toAsset.Bytes()) {
		log.Infof("NFT asset %s already bound to %s", fromAsset.Hex(), toAsset.Hex())
		return true
	}

	hash, err := ethOwner.BindNFTAsset(
		proxy,
		fromAsset,
		toAsset,
		chainID,
	)
	if err != nil {
		log.Errorf("bind NFT asset on ethereum failed, err: %s", err.Error())
		return
	}

	actual, err := ethOwner.GetBoundNFTAsset(proxy, fromAsset, chainID)
	if err != nil {
		log.Error(err)
		return
	}
	if !bytes.Equal(toAsset.Bytes(), actual.Bytes()) {
		log.Errorf("bind NFT asset failed, expect %s != actual %s", toAsset.Hex(), actual.Hex())
		return
	}

	log.Infof("bind NFT asset %s to %s on ethereum success, hash %s", fromAsset.Hex(), toAsset.Hex(), hash.Hex())
	return true
}

func ETHTransferPLTAssetOwnership() (succeed bool) {
	asset := config.Conf.CrossChain.EthereumPLTAsset
	newOwner := config.Conf.FinalOwner.EthereumFinalOwner
	ethOwner := getEthereumCli(ethCTypeOwner)

	cur, _ := ethOwner.PLTAssetOwnership(asset)
	if bytes.Equal(newOwner.Bytes(), cur.Bytes()) {
		log.Infof("plt asset %s owner is %s already", asset.Hex(), newOwner.Hex())
		return true
	}

	hash1, err := ethOwner.TransferPLTAssetOwnership(asset, newOwner)
	if err != nil {
		log.Errorf("transfer plt asset ownership to eccm on ethereum failed, err: %s", err.Error())
		return
	}

	newEthOwner := customEthereumCli(newOwner)
	hash2, err := newEthOwner.AcceptOwnership(asset)
	if err != nil {
		log.Errorf("accept plt asset ownership to eccm on ethereum failed, err: %s", err.Error())
		return
	}

	actual, err := ethOwner.PLTAssetOwnership(asset)
	if err != nil {
		log.Errorf("get plt asset new owner failed, err :%v", err)
		return
	}
	if !bytes.Equal(newOwner.Bytes(), actual.Bytes()) {
		log.Errorf("transfer plt asset ownership failed. expect %s != actual %s", newOwner.Hex(), actual.Hex())
		return
	}

	log.Infof("transfer plt asset ownership to eccm on ethereum success, transfer ownership tx %s, accept ownership tx %s", hash1.Hex(), hash2.Hex())
	return true
}

func ETHTransferPLTProxyOwnership() (succeed bool) {
	proxy := config.Conf.CrossChain.EthereumPLTProxy
	newOwner := config.Conf.FinalOwner.EthereumFinalOwner
	ethOwner := getEthereumCli(ethCTypeOwner)

	cur, _ := ethOwner.PLTProxyOwnership(proxy)
	if bytes.Equal(newOwner.Bytes(), cur.Bytes()) {
		log.Infof("plt proxy %s owner is %s already", proxy.Hex(), newOwner.Hex())
		return true
	}

	hash, err := ethOwner.TransferPLTProxyOwnership(proxy, newOwner)
	if err != nil {
		log.Errorf("transfer plt proxy ownership to eccm on ethereum failed, err: %s", err.Error())
		return
	}

	actual, err := ethOwner.PLTProxyOwnership(proxy)
	if err != nil {
		log.Errorf("get plt proxy new owner failed, err :%v", err)
		return
	}
	if !bytes.Equal(newOwner.Bytes(), actual.Bytes()) {
		log.Errorf("transfer plt proxy ownership failed. expect %s != actual %s", newOwner.Hex(), actual.Hex())
		return
	}

	log.Infof("transfer plt proxy ownership to eccm on ethereum success, tx %s", hash.Hex())
	return true
}

func ETHTransferNFTProxyOwnership() (succeed bool) {
	proxy := config.Conf.CrossChain.EthereumNFTProxy
	newOwner := config.Conf.FinalOwner.EthereumFinalOwner
	ethOwner := getEthereumCli(ethCTypeOwner)

	cur, _ := ethOwner.NFTProxyOwnership(proxy)
	if bytes.Equal(newOwner.Bytes(), cur.Bytes()) {
		log.Infof("nft proxy %s owner is %s already", proxy.Hex(), newOwner.Hex())
		return true
	}

	hash, err := ethOwner.TransferNFTProxyOwnership(proxy, newOwner)
	if err != nil {
		log.Errorf("transfer nft proxy ownership to eccm on ethereum failed, err: %s", err.Error())
		return
	}

	actual, err := ethOwner.NFTProxyOwnership(proxy)
	if err != nil {
		log.Errorf("get nft proxy new owner failed, err :%v", err)
		return
	}
	if !bytes.Equal(newOwner.Bytes(), actual.Bytes()) {
		log.Errorf("transfer nft proxy ownership failed. expect %s != actual %s", newOwner.Hex(), actual.Hex())
		return
	}

	log.Infof("transfer nft proxy ownership to eccm on ethereum success, tx %s", hash.Hex())
	return true
}

// 同步以太坊区块头到poly
func ETHSyncEthGenesis() (succeed bool) {
	polyRPC := config.Conf.CrossChain.PolyRPCAddress
	polyValidators := config.Conf.CrossChain.LoadPolyAccountList()
	crossChainID := config.Conf.CrossChain.EthereumSideChainID
	ethOwner := getEthereumCli(ethCTypeOwner)

	polyCli, err := poly.NewPolyClient(polyRPC, polyValidators)
	if err != nil {
		log.Errorf("failed to generate poly client, err: %s", err)
		return
	} else {
		log.Infof("generate poly client success!")
	}

	curr, err := ethOwner.GetCurrentHeight()
	if err != nil {
		log.Error(err)
		return
	}
	hdr, err := ethOwner.GetHeader(curr)
	if err != nil {
		log.Error(err)
		return
	}
	hdrEnc, err := hdr.MarshalJSON()
	if err != nil {
		log.Error(err)
		return
	}

	if err := polyCli.SyncGenesisBlock(crossChainID, hdrEnc); err != nil {
		log.Errorf("SyncEthGenesisHeader, cross chainID %d, failed: %v", crossChainID, err)
		return
	}
	log.Infof("successful to sync eth genesis header: txhash %s, block number %d",
		hdr.Hash().Hex(), hdr.Number.Uint64())

	return true
}

// 同步poly区块头到以太坊
func ETHSyncPolyGenesis() (succeed bool) {
	polyRPC := config.Conf.CrossChain.PolyRPCAddress
	polyCli, err := poly.NewPolyClient(polyRPC, nil)
	if err != nil {
		log.Errorf("failed to generate poly client, err: %s", err)
		return
	} else {
		log.Infof("generate poly client success!")
	}

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

	eccm := config.Conf.CrossChain.EthereumECCM
	ethOwner := getEthereumCli(ethCTypeOwner)
	txhash, err := ethOwner.InitGenesisBlock(eccm, headerEnc, bookeepersEnc)
	if err != nil {
		log.Errorf("failed to initGenesisBlock, err: %s", err)
		return
	} else {
		log.Infof("sync genesis header success, txhash %s", txhash.Hex())
	}

	return true
}

// 以太坊上从PLT owner 跨链转移资产到palette governance合约，作为palette的reward奖励池
func ETHPLTMintGovernance() (succeed bool) {
	var params struct {
		Amount int
	}

	if err := config.LoadParams("ETH-PLT-Mint-Gov.json", &params); err != nil {
		log.Error(err)
		return
	}

	invoker := getEthereumCli(ethCTypeOwner)
	from := invoker.Address()
	to := common.HexToAddress(native.GovernanceContractAddress)
	proxy := config.Conf.CrossChain.EthereumPLTProxy
	targetSideChainID := config.Conf.CrossChain.PaletteSideChainID
	asset := config.Conf.CrossChain.EthereumPLTAsset
	amount := plt.MultiPLT(params.Amount)
	cli := getPaletteCli(pltCTypeCustomer)

	// please make sure that eth account's balance is enough for gas fee.

	// prepare allowance
	logsplit()
	log.Infof("prepare from account allowance for proxy......")
	if err := prepareAllowance(invoker, from, proxy, amount); err != nil {
		log.Error(err)
		return
	}

	// unlock
	logsplit()
	log.Infof("lock plt on ethereum......")
	totalSupplyBeforeLockOnEthereum, err := invoker.PLTTotalSupply(asset)
	if err != nil {
		log.Error(err)
		return
	}
	totalSupplyBeforeLockOnPalette, err := cli.PLTTotalSupply("latest")
	if err != nil {
		log.Error(err)
		return
	}
	hash, err := invoker.PLTLock(proxy, asset, targetSideChainID, to, amount)
	if err != nil {
		log.Error(err)
		return
	} else {
		log.Infof("lock plt on ethereum, tx hash %s", hash.Hex())
	}

	time.Sleep(25 * time.Second)
	logsplit()
	log.Info("check balance on both of palette chain and ethereum chain...")
	for i := 0; i < 100; i++ {
		totalSupplyAfterLockOnEthereum, err := invoker.PLTTotalSupply(asset)
		if err != nil {
			log.Error(err)
			return
		}
		totalSupplyAfterLockOnPalette, err := cli.PLTTotalSupply("latest")
		if err != nil {
			log.Error(err)
			return
		}

		log.Infof("ethereum %s: totalSupply before lock [%d], totalSupply after lock [%d]",
			from.Hex(),
			plt.PrintUPLT(totalSupplyBeforeLockOnEthereum),
			plt.PrintUPLT(totalSupplyAfterLockOnEthereum),
		)
		log.Infof("palette %s: totalSupply before lock [%d], totalSupply after lock [%d]",
			to.Hex(),
			plt.PrintUPLT(totalSupplyBeforeLockOnPalette),
			plt.PrintUPLT(totalSupplyAfterLockOnPalette),
		)

		sub := new(big.Int).Sub(totalSupplyAfterLockOnPalette, totalSupplyBeforeLockOnPalette)
		if sub.Cmp(amount) == 0 {
			log.Infof("lock tx hash %s success!", hash.Hex())
			break
		}
		logsplit()
		wait(1)
	}

	return true
}

func ETHPLTMintAdmin() (succeed bool) {
	var params struct {
		Amount int
	}

	if err := config.LoadParams("ETH-PLT-Mint-Admin.json", &params); err != nil {
		log.Error(err)
		return
	}

	invoker := getEthereumCli(ethCTypeOwner)
	from := invoker.Address()
	to := config.Conf.AdminAccount
	proxy := config.Conf.CrossChain.EthereumPLTProxy
	targetSideChainID := config.Conf.CrossChain.PaletteSideChainID
	asset := config.Conf.CrossChain.EthereumPLTAsset
	amount := plt.MultiPLT(params.Amount)
	cli := getPaletteCli(pltCTypeCustomer)

	// please make sure that eth account's balance is enough for gas fee.
	// prepare ETH for gas fee
	//{
	//	logsplit()
	//	log.Infof("prepare eth gas fee......")
	//	gasLimit := 210000
	//	gasFee, err := calculateGasFee(invoker, uint64(gasLimit))
	//	if err != nil {
	//		log.Errorf("calculate gas fee err %s", err.Error())
	//	}
	//	amount := utils.SafeMul(gasFee, big.NewInt(2))
	//	if err := prepareEth(from, amount); err != nil {
	//		log.Errorf("prepare eth as gas failed, err: %s", err.Error())
	//		return
	//	}
	//}

	// prepare allowance
	logsplit()
	log.Infof("prepare from account allowance for proxy......")
	if err := prepareAllowance(invoker, from, proxy, amount); err != nil {
		log.Error(err)
		return
	}

	// unlock
	logsplit()
	log.Infof("lock plt on ethereum......")
	fromBalanceBeforeLockOnEthereum, err := invoker.PLTBalanceOf(asset, from)
	if err != nil {
		log.Error(err)
		return
	}
	toBalanceBeforeLockOnPalette, err := cli.BalanceOf(to, "latest")
	if err != nil {
		log.Error(err)
		return
	}
	hash, err := invoker.PLTLock(proxy, asset, targetSideChainID, to, amount)
	if err != nil {
		log.Error(err)
		return
	} else {
		log.Infof("lock plt on ethereum, tx hash %s", hash.Hex())
	}

	time.Sleep(25 * time.Second)
	logsplit()
	log.Info("check balance on both of palette chain and ethereum chain...")
	for i := 0; i < 100; i++ {
		fromBalanceAfterLockOnEthereum, err := invoker.PLTBalanceOf(asset, from)
		if err != nil {
			log.Error(err)
			return
		}
		toBalanceAfterLockOnPalette, err := cli.BalanceOf(to, "latest")
		if err != nil {
			log.Error(err)
			return
		}

		log.Infof("ethereum %s: balance before lock [%d], balance after lock [%d]",
			from.Hex(),
			plt.PrintUPLT(fromBalanceBeforeLockOnEthereum),
			plt.PrintUPLT(fromBalanceAfterLockOnEthereum),
		)
		log.Infof("palette %s: balance before lock [%d], balance after lock [%d]",
			to.Hex(),
			plt.PrintUPLT(toBalanceBeforeLockOnPalette),
			plt.PrintUPLT(toBalanceAfterLockOnPalette),
		)
		sub := new(big.Int).Sub(toBalanceAfterLockOnPalette, toBalanceBeforeLockOnPalette)
		if sub.Cmp(amount) == 0 {
			log.Infof("lock tx hash %s success!", hash.Hex())
			break
		}
		logsplit()
		wait(1)
	}

	return true
}

func ETHPLTBalance() (succeed bool) {
	var params struct {
		Owner common.Address
	}
	if err := config.LoadParams("ETH-PLT-Balance.json", &params); err != nil {
		log.Error(err)
		return
	}
	ethInvoker := getEthereumCli(ethCTypeInvoker)
	data, err := ethInvoker.PLTBalanceOf(config.Conf.CrossChain.EthereumPLTAsset, params.Owner)
	if err != nil {
		log.Error(err)
		return
	}

	log.Infof("PLT on ethereum %s balance %d", params.Owner.Hex(), plt.PrintUPLT(data))
	return true
}

func ETHPLTTotalSupply() (succeed bool) {
	ethInvoker := getEthereumCli(ethCTypeInvoker)
	data, err := ethInvoker.PLTTotalSupply(config.Conf.CrossChain.EthereumPLTAsset)
	if err != nil {
		log.Error(err)
		return
	}

	log.Infof("ethereum PLT asset total supply %d", plt.PrintUPLT(data))
	return true
}

func ETHPLTTransfer() (succeed bool) {
	var params struct {
		From   common.Address
		To     common.Address
		Amount int
	}
	if err := config.LoadParams("ETH-PLT-Transfer.json", &params); err != nil {
		log.Error(err)
		return
	}
	amount := plt.MultiPLT(params.Amount)
	asset := config.Conf.CrossChain.EthereumPLTAsset
	ethInvoker := getEthereumCli(ethCTypeInvoker)

	fromBalanceBeforeTransfer, err := ethInvoker.PLTBalanceOf(asset, params.From)
	if err != nil {
		log.Error(err)
		return
	}
	toBalanceBeforeTransfer, err := ethInvoker.PLTBalanceOf(asset, params.To)
	if err != nil {
		log.Error(err)
		return
	}

	hash, err := ethInvoker.PLTTransfer(asset, params.From, params.To, amount)
	if err != nil {
		log.Error(err)
		return
	}
	fromBalanceAfterTransfer, err := ethInvoker.PLTBalanceOf(asset, params.From)
	if err != nil {
		log.Error(err)
		return
	}
	toBalanceAfterTransfer, err := ethInvoker.PLTBalanceOf(asset, params.To)
	if err != nil {
		log.Error(err)
		return
	}

	log.Infof("tx hash %s \r\n"+
		"from %s, balance before transfer %d, balance after transfer %d \r\n"+
		"to %s, balance before transfer %d, balance after transfer %d",
		hash.Hex(),
		params.From.Hex(),
		plt.PrintUPLT(fromBalanceBeforeTransfer),
		plt.PrintUPLT(fromBalanceAfterTransfer),
		params.To.Hex(),
		plt.PrintUPLT(toBalanceBeforeTransfer),
		plt.PrintUPLT(toBalanceAfterTransfer),
	)

	return true
}

// transfer from eth owner
func ETHETHTransfer() (succeed bool) {
	var params struct {
		List   []common.Address
		Amount int
	}
	if err := config.LoadParams("ETH-ETH-Transfer.json", &params); err != nil {
		log.Error(err)
		return
	}

	invoker := getEthereumCli(ethCTypeOwner)
	from := invoker.Address()
	amount := plt.MultiPLT(params.Amount)
	fromBalanceBeforeTransfer, err := invoker.ETHBalance(from)
	if err != nil {
		log.Error(err)
		return
	}
	for _, to := range params.List {
		toBalanceBeforeTransfer, err := invoker.ETHBalance(to)
		if err != nil {
			log.Error(err)
			return
		}

		hash, err := invoker.TransferETH(to, amount)
		if err != nil {
			log.Error(err)
			return
		}
		fromBalanceAfterTransfer, err := invoker.ETHBalance(to)
		if err != nil {
			log.Error(err)
			return
		}
		toBalanceAfterTransfer, err := invoker.ETHBalance(to)
		if err != nil {
			log.Error(err)
			return
		}

		log.Infof("tx hash %s \r\n"+
			"from %s, balance before transfer %d, balance after transfer %d \r\n"+
			"to %s, balance before transfer %d, balance after transfer %d",
			hash.Hex(),
			from.Hex(),
			plt.PrintUPLT(fromBalanceBeforeTransfer),
			plt.PrintUPLT(fromBalanceAfterTransfer),
			to.Hex(),
			plt.PrintUPLT(toBalanceBeforeTransfer),
			plt.PrintUPLT(toBalanceAfterTransfer),
		)
	}

	return true
}
