package core

import (
	"bytes"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/contracts/native"
	"github.com/ethereum/go-ethereum/contracts/native/plt"
	"github.com/palettechain/onRobot/config"
	"github.com/palettechain/onRobot/pkg/log"
	"math/big"
)

func PLTDeployWrap() (succeed bool) {
	feeToken := common.HexToAddress(native.PLTContractAddress)
	chainId := new(big.Int).SetUint64(config.Conf.CrossChain.PaletteSideChainID)
	ccAdmCli := getPaletteCli(pltCTypeInvoker)
	owner := ccAdmCli.Address()

	contractAddr, err := ccAdmCli.DeployPaletteWrapper(owner, feeToken, chainId)
	if err != nil {
		log.Errorf("deploy wrap on palette failed, err: %s", err.Error())
		return
	}

	if err := config.Conf.CrossChain.StorePaletteWrapper(contractAddr); err != nil {
		log.Error("store palette wrapper failed")
		return
	}

	log.Infof("deploy wrap %s on palette success!", contractAddr.Hex())
	return true
}

func PLTWrapperSetLockProxy() (succeed bool) {
	ccAdmCli := getPaletteCli(pltCTypeInvoker)
	wrapAddr := config.Conf.CrossChain.PaletteWrapper
	targetLockProxy := common.HexToAddress(native.PLTContractAddress)

	cur, _ := ccAdmCli.GetPaletteWrapLockProxy(wrapAddr)
	if bytes.Equal(cur.Bytes(), targetLockProxy.Bytes()) {
		log.Infof("wrapper proxy %s already settled", targetLockProxy.Hex())
		return true
	}

	if _, err := ccAdmCli.PaletteWrapSetLockProxy(wrapAddr, targetLockProxy); err != nil {
		log.Errorf("wrapper set lock proxy failed, err: %v", err)
		return false
	}

	got, _ := ccAdmCli.GetPaletteWrapLockProxy(wrapAddr)
	if bytes.Equal(cur.Bytes(), targetLockProxy.Bytes()) {
		log.Infof("wrapper proxy set failed, expect %s, got %s", targetLockProxy.Hex(), got.Hex())
		return true
	}

	log.Infof("wrap set lock proxy %s on palette success!", targetLockProxy.Hex())
	return true
}

func PLTWrapperLock() (succeed bool) {
	var params struct {
		FromAsset common.Address
		To        common.Address
		ToChainID uint64
		Amount    int
		Fee       int
		ID        uint64
	}

	if err := config.LoadParams("PLT-Wrap-Lock.json", &params); err != nil {
		log.Error(err)
		return
	}

	ccAdmCli := getPaletteCli(pltCTypeInvoker)
	wrapAddr := config.Conf.CrossChain.PaletteWrapper

	amount := plt.MultiPLT(params.Amount)
	fee := plt.MultiPLT(params.Fee)
	id := new(big.Int).SetUint64(params.ID)
	if _, err := ccAdmCli.PaletteWrapLock(wrapAddr, params.FromAsset, params.To, params.ToChainID, amount, fee, id); err != nil {
		log.Errorf("wrapper set lock proxy failed, err: %v", err)
		return false
	}

	log.Infof("wrap lock success!")
	return true
}
