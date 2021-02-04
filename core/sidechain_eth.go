package core

import (
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
	eccm, err := ethOwner.DeployECCMContract(eccd)
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

func ETHTransferOwnership() (succeed bool) {
	eccd := config.Conf.CrossChain.EthereumECCD
	eccm := config.Conf.CrossChain.EthereumECCM
	ccmp := config.Conf.CrossChain.EthereumCCMP

	hash, err := ethOwner.TransferECCDOwnership(eccd, eccm)
	if err != nil {
		log.Errorf("transfer eccd ownership to eccm on ethereum failed, err: %s", err.Error())
		return
	} else {
		log.Infof("transfer eccd ownership to eccm on ethereum success, tx %s", hash.Hex())
	}

	hash, err = ethInvoker.TransferECCMOwnership(eccm, ccmp)
	if err != nil {
		log.Errorf("transfer eccm ownership to ccmp on ethereum failed, err: %s", err.Error())
		return
	} else {
		log.Infof("transfer eccm ownership to ccmp on ethereum success, tx %s", hash.Hex())
	}

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

	hash, err := ethOwner.BindPLTProxy(localLockProxy, targetLockProxy, targetSideChainID)
	if err != nil {
		log.Errorf("bind PLT proxy on ethereum failed, err: %s", err.Error())
		return
	} else {
		log.Infof("bind PLT proxy on ethereum success, hash %s", hash.Hex())
	}

	return true
}

func ETHBindPLTAsset() (succeed bool) {
	localLockProxy := config.Conf.CrossChain.EthereumPLTProxy
	fromAsset := config.Conf.CrossChain.EthereumPLTAsset
	toAsset := common.HexToAddress(native.PLTContractAddress)
	toChainId := config.Conf.CrossChain.PaletteSideChainID

	hash, err := ethOwner.BindPLTAsset(localLockProxy, fromAsset, toAsset, toChainId)
	if err != nil {
		log.Errorf("bind PLT proxy on ethereum failed, err: %s", err.Error())
		return
	} else {
		log.Infof("bind PLT proxy on ethereum success, hash %s", hash.Hex())
	}

	return true
}

func ETHSetPLTCCMP() (succeed bool) {
	proxy := config.Conf.CrossChain.EthereumPLTProxy
	ccmp := config.Conf.CrossChain.EthereumCCMP
	hash, err := ethOwner.SetPLTCCMP(proxy, ccmp)
	if err != nil {
		log.Errorf("register PLT proxy to ccmp on ethereum failed, err: %s", err.Error())
		return
	} else {
		log.Infof("register PLT proxy to ccmp on ethereum success, tx %s", hash.Hex())
	}

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
	if err := config.LoadParams("ETH-NFT-Deploy.json", &params); err != nil {
		log.Error(err)
		return
	}
	proxy := config.Conf.CrossChain.EthereumNFTProxy
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
	hash, err := ethOwner.SetNFTCCMP(proxy, ccmp)
	if err != nil {
		log.Errorf("register NFT proxy to ccmp on ethereum failed, err: %s", err.Error())
		return
	} else {
		log.Infof("register NFT proxy to ccmp on ethereum success, tx %s", hash.Hex())
	}

	return true
}

func ETHBindNFTProxy() (succeed bool) {
	localLockProxy := config.Conf.CrossChain.EthereumNFTProxy
	targetLockProxy := config.Conf.CrossChain.PaletteNFTProxy
	targetSideChainID := config.Conf.CrossChain.PaletteSideChainID

	tx, err := ethOwner.BindNFTProxy(localLockProxy, targetLockProxy, targetSideChainID)
	if err != nil {
		log.Errorf("bind NFT proxy on ethereum failed, err: %s", err.Error())
		return
	} else {
		log.Infof("bind NFT proxy on ethereum success, hash %s", tx.Hex())
	}

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
	hash, err := ethOwner.BindNFTAsset(
		proxy,
		fromAsset,
		toAsset,
		chainID,
	)
	if err != nil {
		log.Errorf("bind NFT proxy on ethereum failed, err: %s", err.Error())
		return
	}

	log.Infof("bind NFT proxy on ethereum success, hash %s", hash.Hex())

	return true
}

// sync eth genesis
func ETHSyncGenesis() (succeed bool) {
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

	// 2. get ethereum current height&header and sync ethereum header to poly
	{
		logsplit()
		crossChainID := config.Conf.CrossChain.EthereumSideChainID

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
	}

	// 3. get poly block and assemble book keepers to header
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

		eccm := config.Conf.CrossChain.EthereumECCM
		txhash, err := ethOwner.InitGenesisBlock(eccm, headerEnc, bookeepersEnc)
		if err != nil {
			log.Errorf("failed to initGenesisBlock, err: %s", err)
			return
		} else {
			log.Infof("sync genesis header success, txhash %s", txhash.Hex())
		}
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
	data, err := ethInvoker.PLTBalanceOf(config.Conf.CrossChain.EthereumPLTAsset, params.Owner)
	if err != nil {
		log.Error(err)
		return
	}

	log.Infof("PLT on ethereum %s balance %d", params.Owner.Hex(), plt.PrintUPLT(data))
	return true
}

func ETHPLTTotalSupply() (succeed bool) {
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
