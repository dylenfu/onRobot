/*
* Copyright (C) 2020 The poly network Authors
* This file is part of The poly network library.
*
* The poly network is free software: you can redistribute it and/or modify
* it under the terms of the GNU Lesser General Public License as published by
* the Free Software Foundation, either version 3 of the License, or
* (at your option) any later version.
*
* The poly network is distributed in the hope that it will be useful,
* but WITHOUT ANY WARRANTY; without even the implied warranty of
* MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
* GNU Lesser General Public License for more details.
* You should have received a copy of the GNU Lesser General Public License
* along with The poly network . If not, see <http://www.gnu.org/licenses/>.
 */
package eth

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/contracts/native/utils"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/palettechain/onRobot/pkg/encode"
	"github.com/palettechain/onRobot/pkg/log"
	pltabi "github.com/palettechain/palette_token/go_abi/plt"
	"github.com/polynetwork/eth-contracts/go_abi/eccd_abi"
	"github.com/polynetwork/eth-contracts/go_abi/eccm_abi"
	"github.com/polynetwork/eth-contracts/go_abi/eccmp_abi"
	"github.com/polynetwork/eth-contracts/go_abi/lock_proxy_abi"
	"github.com/polynetwork/eth-contracts/go_abi/nftlp"
	nftmapping "github.com/polynetwork/eth-contracts/go_abi/nftmapping_abi"
	polycm "github.com/polynetwork/poly/common"
)

// 部署在以太上的PLT token来自项目github.com/palettechain/palette-token.git
// 该项目的proxy和admin合约用于合约升级，并不是跨链使用的lockProxy。
// 而对应的proxy和NFT的proxy一样，来自项目github.com/polynetwork/eth-contracts.git

type EthInvoker struct {
	PrivateKey *ecdsa.PrivateKey
	ChainID    uint64
	Tools      *ETHTools
	NM         *NonceManager
	TestSigner *EthSigner
}

var (
	DefaultGasLimit = 7000000
)

func NewEInvoker(chainID uint64, url string, privateKey *ecdsa.PrivateKey) *EthInvoker {
	instance := &EthInvoker{}
	instance.ChainID = chainID
	instance.Tools = NewEthTools(url)
	if instance.Tools == nil {
		log.Errorf("dail eth failed")
	}
	instance.NM = NewNonceManager(instance.Tools.GetEthClient())
	instance.PrivateKey = privateKey
	address := crypto.PubkeyToAddress(privateKey.PublicKey)
	instance.TestSigner = &EthSigner{
		PrivateKey: privateKey,
		Address:    address,
	}
	return instance
}

func (i *EthInvoker) TransferETH(to common.Address, amount *big.Int) (common.Hash, error) {
	auth, err := i.makeAuth()
	if err != nil {
		return utils.EmptyHash, err
	}
	auth.Value = amount
	tx := types.NewTransaction(auth.Nonce.Uint64(), to, amount, auth.GasLimit, auth.GasPrice, []byte{})
	if tx, err = types.SignTx(tx, types.HomesteadSigner{}, i.PrivateKey); err != nil {
		return utils.EmptyHash, err
	}
	if err := i.Tools.ethclient.SendTransaction(context.Background(), tx, bind.PrivateTxArgs{}); err != nil {
		return utils.EmptyHash, err
	}
	if err := i.waitTxConfirm(tx.Hash()); err != nil {
		return utils.EmptyHash, err
	}
	return tx.Hash(), nil
}

func (i *EthInvoker) ETHBalance(owner common.Address) (*big.Int, error) {
	return i.Tools.ethclient.BalanceAt(context.Background(), owner, nil)
}

func (i *EthInvoker) DeployPLTLockProxy() (common.Address, error) {
	auth, err := i.makeAuth()
	if err != nil {
		return utils.EmptyAddress, err
	}
	contractAddr, tx, _, err := lock_proxy_abi.DeployLockProxy(auth, i.backend())
	if err != nil {
		return utils.EmptyAddress, err
	}
	if err := i.waitTxConfirm(tx.Hash()); err != nil {
		return utils.EmptyAddress, err
	}
	return contractAddr, nil
}

func (i *EthInvoker) DeployNFTLockProxy() (common.Address, error) {
	auth, err := i.makeAuth()
	if err != nil {
		return utils.EmptyAddress, err
	}
	contractAddr, tx, _, err := nftlp.DeployNFTLockProxy(auth, i.backend())
	if err != nil {
		return utils.EmptyAddress, err
	}
	if err := i.waitTxConfirm(tx.Hash()); err != nil {
		return utils.EmptyAddress, err
	}
	return contractAddr, nil
}

func (i *EthInvoker) SetPLTCCMP(proxyAddr, ccmpAddr common.Address) (common.Hash, error) {
	proxy, err := lock_proxy_abi.NewLockProxy(proxyAddr, i.backend())
	if err != nil {
		return utils.EmptyHash, err
	}
	auth, err := i.makeAuth()
	if err != nil {
		return utils.EmptyHash, err
	}
	tx, err := proxy.SetManagerProxy(auth, ccmpAddr)
	if err != nil {
		return utils.EmptyHash, err
	}
	if err := i.waitTxConfirm(tx.Hash()); err != nil {
		return utils.EmptyHash, err
	}
	return tx.Hash(), nil
}

func (i *EthInvoker) SetNFTCCMP(proxyAddr, ccmpAddr common.Address) (common.Hash, error) {
	proxy, err := nftlp.NewNFTLockProxy(proxyAddr, i.backend())
	if err != nil {
		return utils.EmptyHash, err
	}
	auth, err := i.makeAuth()
	if err != nil {
		return utils.EmptyHash, err
	}
	tx, err := proxy.SetManagerProxy(auth, ccmpAddr)
	if err != nil {
		return utils.EmptyHash, err
	}
	if err := i.waitTxConfirm(tx.Hash()); err != nil {
		return utils.EmptyHash, err
	}
	return tx.Hash(), nil
}

func (i *EthInvoker) DeployPLTAsset() (common.Address, error) {
	auth, err := i.makeAuth()
	if err != nil {
		return utils.EmptyAddress, err
	}
	contractAddr, tx, _, err := pltabi.DeployPaletteToken(auth, i.backend())
	if err != nil {
		return utils.EmptyAddress, err
	}
	if err := i.waitTxConfirm(tx.Hash()); err != nil {
		return utils.EmptyAddress, err
	}
	return contractAddr, nil
}

func (i *EthInvoker) DeployNFT(lockProxy common.Address, name, symbol string) (common.Address, error) {
	auth, err := i.makeAuth()
	if err != nil {
		return utils.EmptyAddress, err
	}
	address, tx, inst, err := nftmapping.DeployCrossChainNFTMapping(auth, i.backend(), lockProxy, name, symbol)
	if err != nil {
		return utils.EmptyAddress, err
	}
	if err := i.waitTxConfirm(tx.Hash()); err != nil {
		return utils.EmptyAddress, err
	}
	nameAfterDeploy, err := inst.Name(nil)
	if err != nil {
		return utils.EmptyAddress, err
	}
	if nameAfterDeploy != name {
		return utils.EmptyAddress, fmt.Errorf("mapping contract deployed name %s != %s", nameAfterDeploy, name)
	}
	return address, nil
}

func (i *EthInvoker) DeployECCDContract() (common.Address, error) {
	auth, err := i.makeAuth()
	if err != nil {
		return utils.EmptyAddress, fmt.Errorf("DeployECCDContract, err: %v", err)
	}
	contractAddress, tx, _, err := eccd_abi.DeployEthCrossChainData(auth, i.backend())
	if err != nil {
		return common.Address{}, fmt.Errorf("DeployECCDContract, err: %v", err)
	}
	if err := i.waitTxConfirm(tx.Hash()); err != nil {
		return utils.EmptyAddress, err
	}
	return contractAddress, nil
}

func (i *EthInvoker) DeployECCMContract(eccd common.Address) (common.Address, error) {
	auth, err := i.makeAuth()
	if err != nil {
		return utils.EmptyAddress, fmt.Errorf("DeployECCMContract, err: %v", err)
	}
	contractAddress, tx, _, err := eccm_abi.DeployEthCrossChainManager(auth, i.backend(), eccd, i.ChainID)
	if err != nil {
		return common.Address{}, fmt.Errorf("DeployECCMContract, err: %v", err)
	}
	if err := i.waitTxConfirm(tx.Hash()); err != nil {
		return utils.EmptyAddress, err
	}
	return contractAddress, nil
}

func (i *EthInvoker) DeployCCMPContract(eccmAddress common.Address) (common.Address, error) {
	auth, err := i.makeAuth()
	if err != nil {
		return utils.EmptyAddress, err
	}
	contractAddress, tx, _, err := eccmp_abi.DeployEthCrossChainManagerProxy(auth, i.backend(), eccmAddress)
	if err != nil {
		return utils.EmptyAddress, fmt.Errorf("DeployCCMPContract, err: %v", err)
	}
	if err := i.waitTxConfirm(tx.Hash()); err != nil {
		return utils.EmptyAddress, err
	}
	return contractAddress, nil
}

func (i *EthInvoker) BindPLTAsset(
	localLockProxyAddr,
	fromAssetHash,
	toAssetHash common.Address,
	toChainId uint64,
) (common.Hash, error) {

	proxy, err := lock_proxy_abi.NewLockProxy(localLockProxyAddr, i.backend())
	if err != nil {
		return utils.EmptyHash, err
	}

	auth, err := i.makeAuth()
	if err != nil {
		return utils.EmptyHash, err
	}
	tx, err := proxy.BindAssetHash(auth, fromAssetHash, toChainId, toAssetHash[:])
	if err != nil {
		return utils.EmptyHash, err
	}
	if err := i.waitTxConfirm(tx.Hash()); err != nil {
		return utils.EmptyHash, err
	}
	return tx.Hash(), nil
}

func (i *EthInvoker) BindPLTProxy(
	localLockProxy,
	targetLockProxy common.Address,
	targetSideChainID uint64,
) (common.Hash, error) {

	proxy, err := lock_proxy_abi.NewLockProxy(localLockProxy, i.backend())
	if err != nil {
		return utils.EmptyHash, err
	}

	auth, err := i.makeAuth()
	if err != nil {
		return utils.EmptyHash, err
	}
	tx, err := proxy.BindProxyHash(auth, targetSideChainID, targetLockProxy.Bytes())
	if err != nil {
		return utils.EmptyHash, err
	}
	if err := i.waitTxConfirm(tx.Hash()); err != nil {
		return utils.EmptyHash, err
	}
	return tx.Hash(), nil
}

func (i *EthInvoker) BindNFTAsset(
	lockProxyAddr,
	fromAssetHash,
	toAssetHash common.Address,
	targetSideChainId uint64) (common.Hash, error) {

	proxy, err := nftlp.NewNFTLockProxy(lockProxyAddr, i.backend())
	if err != nil {
		return utils.EmptyHash, err
	}

	auth, err := i.makeAuth()
	if err != nil {
		return utils.EmptyHash, err
	}
	tx, err := proxy.BindAssetHash(auth, fromAssetHash, targetSideChainId, toAssetHash[:])
	if err != nil {
		return utils.EmptyHash, err
	}
	if err := i.waitTxConfirm(tx.Hash()); err != nil {
		return utils.EmptyHash, err
	}
	return tx.Hash(), nil
}

func (i *EthInvoker) BindNFTProxy(
	localLockProxy,
	targetLockProxy common.Address,
	targetSideChainID uint64,
) (common.Hash, error) {

	proxy, err := nftlp.NewNFTLockProxy(localLockProxy, i.backend())
	if err != nil {
		return utils.EmptyHash, err
	}

	auth, err := i.makeAuth()
	if err != nil {
		return utils.EmptyHash, err
	}
	tx, err := proxy.BindProxyHash(auth, targetSideChainID, targetLockProxy.Bytes())
	if err != nil {
		return utils.EmptyHash, err
	}

	if err := i.waitTxConfirm(tx.Hash()); err != nil {
		return utils.EmptyHash, err
	}
	return tx.Hash(), nil
}

func (i *EthInvoker) TransferECCDOwnership(eccd, eccm common.Address) (common.Hash, error) {
	eccdContract, err := eccd_abi.NewEthCrossChainData(eccd, i.Tools.GetEthClient())
	if err != nil {
		return utils.EmptyHash, fmt.Errorf("TransferECCDOwnership, err: %v", err)
	}

	auth, err := i.makeAuth()
	if err != nil {
		return utils.EmptyHash, err
	}
	tx, err := eccdContract.TransferOwnership(auth, eccm)
	if err != nil {
		return utils.EmptyHash, fmt.Errorf("TransferECCDOwnership, err: %v", err)
	}
	if err := i.waitTxConfirm(tx.Hash()); err != nil {
		return utils.EmptyHash, err
	}
	return tx.Hash(), nil
}

func (i *EthInvoker) TransferECCMOwnership(eccm, ccmp common.Address) (common.Hash, error) {
	eccmContract, err := eccm_abi.NewEthCrossChainManager(eccm, i.backend())
	if err != nil {
		return utils.EmptyHash, fmt.Errorf("TransferECCMOwnership err: %v", err)
	}

	auth, err := i.makeAuth()
	if err != nil {
		return utils.EmptyHash, err
	}
	tx, err := eccmContract.TransferOwnership(auth, ccmp)
	if err != nil {
		return utils.EmptyHash, fmt.Errorf("TransferECCMOwnership err: %v", err)
	}
	if err := i.waitTxConfirm(tx.Hash()); err != nil {
		return utils.EmptyHash, err
	}
	return tx.Hash(), nil
}

func (i *EthInvoker) PLTBalanceOf(asset, user common.Address) (*big.Int, error) {
	instance, err := pltabi.NewPaletteToken(asset, i.backend())
	if err != nil {
		return nil, err
	}
	return instance.BalanceOf(nil, user)
}

func (i *EthInvoker) PLTAllowance(asset, owner, spender common.Address) (*big.Int, error) {
	instance, err := pltabi.NewPaletteToken(asset, i.backend())
	if err != nil {
		return nil, err
	}
	return instance.Allowance(nil, owner, spender)
}

func (i *EthInvoker) PLTApprove(asset, spender common.Address, amount *big.Int) (common.Hash, error) {
	instance, err := pltabi.NewPaletteToken(asset, i.backend())
	if err != nil {
		return utils.EmptyHash, err
	}
	auth, err := i.makeAuth()
	if err != nil {
		return utils.EmptyHash, err
	}
	tx, err := instance.Approve(auth, spender, amount)
	if err != nil {
		return utils.EmptyHash, err
	}
	if err := i.waitTxConfirm(tx.Hash()); err != nil {
		return utils.EmptyHash, err
	}
	return tx.Hash(), nil
}

func (i *EthInvoker) PLTTotalSupply(asset common.Address) (*big.Int, error) {
	instance, err := pltabi.NewPaletteToken(asset, i.backend())
	if err != nil {
		return nil, err
	}
	return instance.TotalSupply(nil)
}

func (i *EthInvoker) PLTTransfer(asset, from, to common.Address, amount *big.Int) (common.Hash, error) {
	instance, err := pltabi.NewPaletteToken(asset, i.backend())
	if err != nil {
		return utils.EmptyHash, err
	}
	auth, err := i.makeAuth()
	if err != nil {
		return utils.EmptyHash, err
	}
	tx, err := instance.Transfer(auth, to, amount)
	if err != nil {
		return utils.EmptyHash, err
	}

	if err := i.waitTxConfirm(tx.Hash()); err != nil {
		return utils.EmptyHash, err
	}
	return tx.Hash(), nil
}

func (i *EthInvoker) VerifyAndExecuteTx(
	eccmAddr common.Address,
	proof,
	rawHeader,
	headerProof,
	curRawHeader,
	headerSig []byte,
) (common.Hash, error) {

	eccm, err := eccm_abi.NewEthCrossChainManager(eccmAddr, i.backend())
	if err != nil {
		return utils.EmptyHash, err
	}

	auth, err := i.makeAuth()
	if err != nil {
		return utils.EmptyHash, err
	}
	tx, err := eccm.VerifyHeaderAndExecuteTx(
		auth,
		proof,
		rawHeader,
		headerProof,
		curRawHeader,
		headerSig,
	)
	if err != nil {
		return utils.EmptyHash, err
	}
	if err := i.waitTxConfirm(tx.Hash()); err != nil {
		return utils.EmptyHash, err
	}
	return tx.Hash(), nil
}

func (i *EthInvoker) PLTLock(
	proxyAddr common.Address,
	fromAsset common.Address,
	targetSideChainID uint64,
	toAddr common.Address,
	amount *big.Int,
) (common.Hash, error) {

	proxy, err := lock_proxy_abi.NewLockProxy(proxyAddr, i.backend())
	if err != nil {
		return utils.EmptyHash, err
	}

	auth, err := i.makeAuth()
	if err != nil {
		return utils.EmptyHash, err
	}
	tx, err := proxy.Lock(auth, fromAsset, targetSideChainID, toAddr.Bytes(), amount)
	if err != nil {
		return utils.EmptyHash, err
	}

	if err := i.waitTxConfirm(tx.Hash()); err != nil {
		return utils.EmptyHash, err
	}
	return tx.Hash(), nil
}

func (i *EthInvoker) PLTUnlock(
	fromChainID uint64,
	proxyAddr,
	fromContract,
	toAsset,
	toAddress common.Address,
	amount *big.Int,
) (common.Hash, error) {

	proxy, err := lock_proxy_abi.NewLockProxy(proxyAddr, i.backend())
	if err != nil {
		return utils.EmptyHash, err
	}

	auth, err := i.makeAuth()
	if err != nil {
		return utils.EmptyHash, err
	}
	args := encode.TxArgs{
		ToAssetHash: toAsset.Bytes(),
		ToAddress:   toAddress.Bytes(),
		Amount:      amount,
	}
	enc := args.Serialization()

	tx, err := proxy.Unlock(auth, enc, fromContract.Bytes(), fromChainID)
	if err != nil {
		return utils.EmptyHash, err
	}

	if err := i.waitTxConfirm(tx.Hash()); err != nil {
		return utils.EmptyHash, err
	}
	return tx.Hash(), nil
}

func (i *EthInvoker) NFTApprove(asset, to common.Address, token *big.Int) (common.Hash, error) {
	cm, err := nftmapping.NewCrossChainNFTMapping(asset, i.backend())
	if err != nil {
		return utils.EmptyHash, err
	}
	auth, err := i.makeAuth()
	if err != nil {
		return utils.EmptyHash, err
	}
	tx, err := cm.Approve(auth, to, token)
	if err != nil {
		return utils.EmptyHash, err
	}
	if err := i.waitTxConfirm(tx.Hash()); err != nil {
		return utils.EmptyHash, err
	}
	return tx.Hash(), nil
}

func (i *EthInvoker) NFTBalance(asset, owner common.Address) (*big.Int, error) {
	cm, err := nftmapping.NewCrossChainNFTMapping(asset, i.backend())
	if err != nil {
		return nil, err
	}
	return cm.BalanceOf(nil, owner)
}

func (i *EthInvoker) NFTGetApproved(asset common.Address, tokenID *big.Int) (common.Address, error) {
	cm, err := nftmapping.NewCrossChainNFTMapping(asset, i.backend())
	if err != nil {
		return utils.EmptyAddress, err
	}
	return cm.GetApproved(nil, tokenID)
}

func (i *EthInvoker) NFTOwner(asset common.Address, tokenID *big.Int) (common.Address, error) {
	cm, err := nftmapping.NewCrossChainNFTMapping(asset, i.backend())
	if err != nil {
		return utils.EmptyAddress, err
	}
	return cm.OwnerOf(nil, tokenID)
}

func (i *EthInvoker) NFTSafeTransferFrom(
	asset,
	from,
	proxy common.Address,
	tokenID *big.Int,
	to common.Address,
	toChainID uint64,
) (common.Hash, error) {

	cm, err := nftmapping.NewCrossChainNFTMapping(asset, i.backend())
	if err != nil {
		return utils.EmptyHash, err
	}

	auth, err := i.makeAuth()
	if err != nil {
		return utils.EmptyHash, err
	}
	data := assembleSafeTransferCallData(to, toChainID)
	tx, err := cm.SafeTransferFrom0(auth, from, proxy, tokenID, data)
	if err != nil {
		return utils.EmptyHash, err
	}

	if err := i.waitTxConfirm(tx.Hash()); err != nil {
		return utils.EmptyHash, err
	}
	return tx.Hash(), nil
}

func (i *EthInvoker) DumpTx(hash common.Hash) error {
	tx, err := i.GetReceipt(hash)
	if err != nil {
		return fmt.Errorf("faild to get receipt %s", hash.Hex())
	}

	if tx.Status == 0 {
		return fmt.Errorf("receipt failed %s", hash.Hex())
	}

	log.Infof("txhash %s, block height %d", hash.Hex(), tx.BlockNumber.Uint64())
	for _, event := range tx.Logs {
		log.Infof("eventlog address %s", event.Address.Hex())
		log.Infof("eventlog data %s", new(big.Int).SetBytes(event.Data).String())
		for i, topic := range event.Topics {
			log.Infof("eventlog topic[%d] %s", i, topic.String())
		}
	}
	return nil
}

func (i *EthInvoker) GetReceipt(hash common.Hash) (*types.Receipt, error) {
	tx, err := i.Tools.ethclient.TransactionReceipt(context.Background(), hash)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (i *EthInvoker) GetCurrentHeight() (uint64, error) {
	return i.Tools.GetNodeHeight()
}

func (i *EthInvoker) GetHeader(height uint64) (*types.Header, error) {
	return i.Tools.GetBlockHeader(height)
}

func (i *EthInvoker) InitGenesisBlock(eccmAddr common.Address, rawHdr, publickeys []byte) (common.Hash, error) {
	eccm, err := eccm_abi.NewEthCrossChainManager(eccmAddr, i.backend())
	if err != nil {
		return utils.EmptyHash, fmt.Errorf("new EthCrossChainManager err: %s", err)
	}

	auth, err := i.makeAuth()
	if err != nil {
		return utils.EmptyHash, err
	}
	tx, err := eccm.InitGenesisBlock(auth, rawHdr, publickeys)
	if err != nil {
		return utils.EmptyHash, fmt.Errorf("call eccm InitGenesisBlock err: %s", err)
	}

	if err := i.waitTxConfirm(tx.Hash()); err != nil {
		return utils.EmptyHash, err
	}
	return tx.Hash(), nil
}

func (i *EthInvoker) SuggestGasPrice() (*big.Int, error) {
	return i.backend().SuggestGasPrice(context.Background())
}

func (i *EthInvoker) makeAuth() (*bind.TransactOpts, error) {
	publicKey := i.PrivateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("makeAuth, cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := i.backend().PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return nil, fmt.Errorf("makeAuth, addr %s, err %v", fromAddress.Hex(), err)
	}

	gasPrice, err := i.backend().SuggestGasPrice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("makeAuth, %v", err)
	}

	auth := bind.NewKeyedTransactor(i.PrivateKey)
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(int64(0))       // in wei
	auth.GasLimit = uint64(DefaultGasLimit) // in units
	auth.GasPrice = gasPrice.Mul(gasPrice, big.NewInt(1))

	return auth, nil
}

func (i *EthInvoker) waitTxConfirm(hash common.Hash) error {
	i.Tools.WaitTransactionConfirm(hash)
	if err := i.DumpTx(hash); err != nil {
		return err
	}
	return nil
}

func (i *EthInvoker) backend() bind.ContractBackend {
	return i.Tools.GetEthClient()
}

func assembleSafeTransferCallData(toAddress common.Address, chainID uint64) []byte {
	sink := polycm.NewZeroCopySink(nil)
	sink.WriteVarBytes(toAddress.Bytes())
	sink.WriteUint64(chainID)
	return sink.Bytes()
}
