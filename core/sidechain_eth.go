package core

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/contracts/native"
	"github.com/palettechain/onRobot/config"
	"github.com/palettechain/onRobot/pkg/log"
	"github.com/palettechain/onRobot/pkg/poly"
)

///////////////////////////////////////////////////////
//
// deploy eccd, eccm, ccmp and transfer ownership
//
///////////////////////////////////////////////////////
func ETHDeployECCD() (succeed bool) {
	eccd, err := ethInvoker.DeployECCDContract()
	if err != nil {
		log.Errorf("deploy eccd on ethereum failed, err: %s", err.Error())
		return
	} else {
		log.Infof("deploy eccd %s on ethereum success", eccd.Hex())
	}

	return true
}

func ETHDeployECCM() (succeed bool) {
	eccd := config.Conf.CrossChain.EthereumECCD
	eccm, err := ethInvoker.DeployECCMContract(eccd)
	if err != nil {
		log.Errorf("deploy eccm on ethereum failed, err: %s", err.Error())
		return
	} else {
		log.Infof("deploy eccm %s on ethereum success", eccm.Hex())
	}

	return true
}

func ETHDeployCCMP() (succeed bool) {
	eccm := config.Conf.CrossChain.EthereumECCM
	ccmp, err := ethInvoker.DeployCCMPContract(eccm)
	if err != nil {
		log.Errorf("deploy ccmp on ethereum failed, err: %s", err.Error())
		return
	} else {
		log.Infof("deploy ccmp %s on ethereum success", ccmp.Hex())
	}

	return true
}

func ETHTransferOwnership() (succeed bool) {
	eccd := config.Conf.CrossChain.EthereumECCD
	eccm := config.Conf.CrossChain.EthereumECCM
	ccmp := config.Conf.CrossChain.EthereumCCMP

	hash, err := ethInvoker.TransferECCDOwnership(eccd, eccm)
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
	crossChainID := config.Conf.CrossChain.EthereumSideChainID
	if err := polyCli.RegisterSideChain(crossChainID, eccd); err != nil {
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
	pltAsset, err := ethInvoker.DeployPLTAsset()
	if err != nil {
		log.Errorf("deploy PLT asset on ethereum failed, err: %s", err)
		return
	}

	log.Infof("deploy PLT asset %s on ethereum success!", pltAsset.Hex())

	return true
}

func ETHDeployPLTProxy() (succeed bool) {
	proxy, err := ethInvoker.DeployPLTLockProxy()
	if err != nil {
		log.Errorf("deploy PLT proxy on ethereum failed, err: %s", err)
		return
	} else {
		log.Infof("deploy PLT proxy %s on ethereum success!", proxy.Hex())
	}

	return true
}

func ETHBindPLTProxy() (succeed bool) {
	proxy := config.Conf.CrossChain.EthereumPLTProxy
	fromAsset := config.Conf.CrossChain.EthereumPLTAsset
	toAsset := common.HexToAddress(native.PLTContractAddress)
	toChainId := config.Conf.CrossChain.PaletteSideChainID

	hash, err := ethInvoker.BindPLTAssetHash(proxy, fromAsset, toAsset, toChainId)
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
	hash, err := ethInvoker.SetPLTCCMP(proxy, ccmp)
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
	contract, err := ethInvoker.DeployNewNFT()
	if err != nil {
		log.Errorf("deploy new NFT contract on ethereum failed, err: %s", err.Error())
		return
	} else {
		log.Infof("deploy new NFT contract %s on ethereum success!", contract.Hex())
	}

	return true
}

func ETHDeployNFTProxy() (succeed bool) {
	contract, err := ethInvoker.DeployNFTLockProxy()
	if err != nil {
		log.Errorf("deploy nft lock proxy on ethereum failed, err: %s", err.Error())
		return
	} else {
		log.Infof("deploy NFT lock proxy %s on ethereum success!", contract.Hex())
	}

	return true
}

func ETHSetNFTCCMP() (succeed bool) {
	proxy := config.Conf.CrossChain.EthereumNFTProxy
	ccmp := config.Conf.CrossChain.EthereumCCMP
	hash, err := ethInvoker.SetNFTCCMP(proxy, ccmp)
	if err != nil {
		log.Errorf("register NFT proxy to ccmp on ethereum failed, err: %s", err.Error())
		return
	} else {
		log.Infof("register NFT proxy to ccmp on ethereum success, tx %s", hash.Hex())
	}

	return true
}

func ETHBindNFTProxy() (succeed bool) {
	var params = struct {
		EthereumNFTAsset common.Address
		PaletteNFTAsset  common.Address
	}{}
	if err := config.LoadParams("ETH-NFT-BindAsset.json", &params); err != nil {
		log.Error(err)
		return
	}

	proxy := config.Conf.CrossChain.EthereumNFTProxy
	fromAsset := params.EthereumNFTAsset
	toAsset := params.PaletteNFTAsset
	chainID := config.Conf.CrossChain.PaletteSideChainID
	hash, err := ethInvoker.BindNFTAssetHash(
		proxy,
		fromAsset,
		toAsset,
		chainID,
	)
	if err != nil {
		log.Errorf("bind NFT proxy on ethereum failed, err: %s", err.Error())
		return
	}

	if err := ethInvoker.DumpTx(hash); err != nil {
		log.Error(err)
		return
	}
	log.Infof("bind NFT proxy on ethereum success, hash %s", hash.Hex())

	return true
}
