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
	"github.com/palettechain/onRobot/pkg/log"
	pltabi "github.com/palettechain/palette_token/go_abi" // todo: this package redeclared "Address_ABI"
	"github.com/polynetwork/eth-contracts/go_abi/eccd_abi"
	"github.com/polynetwork/eth-contracts/go_abi/eccm_abi"
	"github.com/polynetwork/eth-contracts/go_abi/eccmp_abi"
	"github.com/polynetwork/eth-contracts/go_abi/erc20_abi"
	"github.com/polynetwork/eth-contracts/go_abi/lock_proxy_abi"
	"github.com/polynetwork/eth-contracts/go_abi/nftlp"
	nftmapping "github.com/polynetwork/eth-contracts/go_abi/nftmapping_abi"
	polycm "github.com/polynetwork/poly/common"
)

type EthInvoker struct {
	PrivateKey *ecdsa.PrivateKey
	ChainID    uint64
	Tools      *ETHTools
	NM         *NonceManager
	TestSigner *EthSigner
}

var (
	DefaultGasLimit = 5000000
)

func NewEInvoker(chainID uint64, url string, privateKey *ecdsa.PrivateKey) *EthInvoker {
	instance := &EthInvoker{}
	instance.ChainID = chainID
	instance.Tools = NewEthTools(url)
	instance.NM = NewNonceManager(instance.Tools.GetEthClient())
	instance.PrivateKey = privateKey
	address := crypto.PubkeyToAddress(privateKey.PublicKey)
	instance.TestSigner = &EthSigner{
		PrivateKey: privateKey,
		Address:    address,
	}
	return instance
}

func (i *EthInvoker) DeployPLTLockProxy() (common.Address, error) {
	auth, _ := i.makeAuth()
	contractAddr, tx, _, err := lock_proxy_abi.DeployLockProxy(auth, i.backend())
	if err != nil {
		return utils.EmptyAddress, err
	}
	i.waitTxConfirm(tx.Hash())
	return contractAddr, nil
}

func (i *EthInvoker) DeployNFTLockProxy() (common.Address, error) {
	auth, _ := i.makeAuth()
	contractAddr, tx, _, err := nftlp.DeployNFTLockProxy(auth, i.backend())
	if err != nil {
		return utils.EmptyAddress, err
	}
	i.waitTxConfirm(tx.Hash())
	return contractAddr, nil
}

func (i *EthInvoker) SetPLTCCMP(proxyAddr, ccmpAddr common.Address) (common.Hash, error) {
	proxy, err := lock_proxy_abi.NewLockProxy(proxyAddr, i.backend())
	if err != nil {
		return utils.EmptyHash, err
	}
	auth, _ := i.makeAuth()
	tx, err := proxy.SetManagerProxy(auth, ccmpAddr)
	if err != nil {
		return utils.EmptyHash, err
	}
	i.waitTxConfirm(tx.Hash())
	return tx.Hash(), nil
}

func (i *EthInvoker) SetNFTCCMP(proxyAddr, ccmpAddr common.Address) (common.Hash, error) {
	proxy, err := nftlp.NewNFTLockProxy(proxyAddr, i.backend())
	if err != nil {
		return utils.EmptyHash, err
	}
	auth, _ := i.makeAuth()
	tx, err := proxy.SetManagerProxy(auth, ccmpAddr)
	if err != nil {
		return utils.EmptyHash, err
	}
	i.waitTxConfirm(tx.Hash())
	return tx.Hash(), nil
}

func (i *EthInvoker) DeployPLTAsset() (common.Address, error) {
	auth, _ := i.makeAuth()
	contractAddr, tx, _, err := pltabi.DeployPaletteToken(auth, i.backend())
	if err != nil {
		return utils.EmptyAddress, err
	}
	i.waitTxConfirm(tx.Hash())
	return contractAddr, nil
}

func (i *EthInvoker) DeployNewNFT() (common.Address, error) {
	auth, _ := i.makeAuth()
	contractAddr, tx, _, err := nftmapping.DeployAddress(auth, i.backend())
	if err != nil {
		return utils.EmptyAddress, err
	}
	i.waitTxConfirm(tx.Hash())
	return contractAddr, nil
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
	i.waitTxConfirm(tx.Hash())
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
	i.waitTxConfirm(tx.Hash())
	return contractAddress, nil
}

func (i *EthInvoker) DeployCCMPContract(eccmAddress common.Address) (common.Address, error) {
	auth, _ := i.makeAuth()
	contractAddress, tx, _, err := eccmp_abi.DeployEthCrossChainManagerProxy(auth, i.backend(), eccmAddress)
	if err != nil {
		return utils.EmptyAddress, fmt.Errorf("DeployCCMPContract, err: %v", err)
	}
	i.waitTxConfirm(tx.Hash())
	return contractAddress, nil
}

func (i *EthInvoker) BindPLTAssetHash(
	lockProxyAddr,
	fromAssetHash,
	toAssetHash common.Address,
	toChainId uint64) (common.Hash, error) {

	proxy, err := lock_proxy_abi.NewLockProxy(lockProxyAddr, i.backend())
	if err != nil {
		return utils.EmptyHash, err
	}

	auth, _ := i.makeAuth()
	tx, err := proxy.BindAssetHash(auth, fromAssetHash, toChainId, toAssetHash[:])
	if err != nil {
		return utils.EmptyHash, err
	}
	i.waitTxConfirm(tx.Hash())
	return tx.Hash(), nil
}

func (i *EthInvoker) BindNFTAssetHash(
	lockProxyAddr,
	fromAssetHash,
	toAssetHash common.Address,
	toChainId uint64) (common.Hash, error) {

	proxy, err := nftlp.NewNFTLockProxy(lockProxyAddr, i.backend())
	if err != nil {
		return utils.EmptyHash, err
	}

	auth, _ := i.makeAuth()
	tx, err := proxy.BindAssetHash(auth, fromAssetHash, toChainId, toAssetHash[:])
	if err != nil {
		return utils.EmptyHash, err
	}
	i.waitTxConfirm(tx.Hash())
	return tx.Hash(), nil
}

func (i *EthInvoker) TransferECCDOwnership(eccd, eccm common.Address) (common.Hash, error) {
	eccdContract, err := eccd_abi.NewEthCrossChainData(eccd, i.Tools.GetEthClient())
	if err != nil {
		return utils.EmptyHash, fmt.Errorf("TransferECCDOwnership, err: %v", err)
	}

	auth, _ := i.makeAuth()
	tx, err := eccdContract.TransferOwnership(auth, eccm)
	if err != nil {
		return utils.EmptyHash, fmt.Errorf("TransferECCDOwnership, err: %v", err)
	}
	i.waitTxConfirm(tx.Hash())
	return tx.Hash(), nil
}

func (i *EthInvoker) TransferECCMOwnership(eccm, ccmp common.Address) (common.Hash, error) {
	eccmContract, err := eccm_abi.NewEthCrossChainManager(eccm, i.backend())
	if err != nil {
		return utils.EmptyHash, fmt.Errorf("TransferECCMOwnership err: %v", err)
	}

	auth, _ := i.makeAuth()
	tx, err := eccmContract.TransferOwnership(auth, ccmp)
	if err != nil {
		return utils.EmptyHash, fmt.Errorf("TransferECCMOwnership err: %v", err)
	}
	i.waitTxConfirm(tx.Hash())
	return tx.Hash(), nil
}

func (i *EthInvoker) PLTBalanceOf(asset, user common.Address) (*big.Int, error) {
	instance, err := erc20_abi.NewERC20(asset, i.backend())
	if err != nil {
		return nil, err
	}
	return instance.BalanceOf(nil, user)
}

func (i *EthInvoker) VerifyAndExecuteTx(
	eccmAddr common.Address,
	proof,
	rawHeader,
	headerProof,
	curRawHeader,
	headerSig []byte) (common.Hash, error) {
	eccm, err := eccm_abi.NewEthCrossChainManager(eccmAddr, i.backend())
	if err != nil {
		return utils.EmptyHash, err
	}

	auth, _ := i.makeAuth()
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

	i.waitTxConfirm(tx.Hash())
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

	auth, _ := i.makeAuth()
	enc := _serializeTxArgs(toAsset, toAddress, amount)
	tx, err := proxy.Unlock(auth, enc, fromContract.Bytes(), fromChainID)
	if err != nil {
		return utils.EmptyHash, err
	}

	i.waitTxConfirm(tx.Hash())

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

func (i *EthInvoker) makeAuth() (*bind.TransactOpts, error) {
	publicKey := i.PrivateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("makeAuth, cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := i.backend().PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return nil, fmt.Errorf("makeAuth, %v", err)
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

func (i *EthInvoker) waitTxConfirm(hash common.Hash) {
	i.Tools.WaitTransactionConfirm(hash)
	_ = i.DumpTx(hash)
}

func (i *EthInvoker) backend() bind.ContractBackend {
	return i.Tools.GetEthClient()
}

func _serializeTxArgs(toAsset, toAddress common.Address, amount *big.Int) []byte {
	sink := polycm.NewZeroCopySink(nil)

	sink.WriteVarBytes(toAsset.Bytes())
	sink.WriteVarBytes(toAddress.Bytes())
	sink.WriteVarBytes(amount.Bytes())

	//bytes memory buff;
	//buff = abi.encodePacked(
	//ZeroCopySink.WriteVarBytes(args.toAssetHash),
	//ZeroCopySink.WriteVarBytes(args.toAddress),
	//ZeroCopySink.WriteUint255(args.amount)
	//);
	//return buff;

	return sink.Bytes()
}

func _deserializeTxArgs(valueBs []byte) (toAsset, toAddress common.Address, amount *big.Int) {
	source := polycm.NewZeroCopySource(valueBs)
	toAssetBz, _ := source.NextVarBytes()
	toAddrBz, _ := source.NextVarBytes()
	amountBz, _ := source.NextVarBytes()

	toAsset = common.BytesToAddress(toAssetBz)
	toAddress = common.BytesToAddress(toAddrBz)
	amount = new(big.Int).SetBytes(amountBz)

	return
	//TxArgs memory args;
	//uint256 off = 0;
	//(args.toAssetHash, off) = ZeroCopySource.NextVarBytes(valueBs, off);
	//(args.toAddress, off) = ZeroCopySource.NextVarBytes(valueBs, off);
	//(args.amount, off) = ZeroCopySource.NextUint255(valueBs, off);
	//return args;
}
