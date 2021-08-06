package sdk

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/contracts/native/plt"
	"github.com/ethereum/go-ethereum/contracts/native/utils"
	"github.com/polynetwork/eth-contracts/go_abi/eccd_abi"
	"github.com/polynetwork/eth-contracts/go_abi/eccm_abi"
	"github.com/polynetwork/eth-contracts/go_abi/eccmp_abi"
	"github.com/polynetwork/eth-contracts/go_abi/nftlp"
)

func (c *Client) DeployECCD() (common.Address, error) {
	auth := c.makeDeployAuth()
	addr, tx, _, err := eccd_abi.DeployEthCrossChainData(auth, c.backend)
	if err != nil {
		return utils.EmptyAddress, err
	}
	if err := c.WaitTransaction(tx.Hash()); err != nil {
		return utils.EmptyAddress, err
	}
	return addr, nil
}

func (c *Client) DeployECCM(eccd common.Address, sideChainID uint64) (common.Address, error) {
	auth := c.makeDeployAuth()
	addr, tx, _, err := eccm_abi.DeployEthCrossChainManager(auth, c.backend, eccd, sideChainID)
	if err != nil {
		return utils.EmptyAddress, err
	}
	if err := c.WaitTransaction(tx.Hash()); err != nil {
		return utils.EmptyAddress, err
	}
	return addr, nil
}

func (c *Client) DeployCCMP(eccm common.Address) (common.Address, error) {
	auth := c.makeDeployAuth()
	addr, tx, _, err := eccmp_abi.DeployEthCrossChainManagerProxy(auth, c.backend, eccm)
	if err != nil {
		return utils.EmptyAddress, err
	}
	if err := c.WaitTransaction(tx.Hash()); err != nil {
		return utils.EmptyAddress, err
	}
	return addr, nil
}

func (c *Client) PauseCCMP(ccmpAddr common.Address) (common.Hash, error) {
	ccmp, err := eccmp_abi.NewEthCrossChainManagerProxy(ccmpAddr, c.backend)
	if err != nil {
		return utils.EmptyHash, fmt.Errorf("new EthCrossChainManagerProxy err: %s", err)
	}

	auth := c.makeDeployAuth()
	tx, err := ccmp.PauseEthCrossChainManager(auth)
	if err != nil {
		return utils.EmptyHash, fmt.Errorf("call ccmp pause err: %s", err)
	}

	if err := c.WaitTransaction(tx.Hash()); err != nil {
		return utils.EmptyHash, err
	}
	return tx.Hash(), nil
}

func (c *Client) UnPauseCCMP(ccmpAddr common.Address) (common.Hash, error) {
	ccmp, err := eccmp_abi.NewEthCrossChainManagerProxy(ccmpAddr, c.backend)
	if err != nil {
		return utils.EmptyHash, fmt.Errorf("new EthCrossChainManagerProxy err: %s", err)
	}

	auth := c.makeDeployAuth()
	tx, err := ccmp.UnpauseEthCrossChainManager(auth)
	if err != nil {
		return utils.EmptyHash, fmt.Errorf("call ccmp unpause err: %s", err)
	}

	if err := c.WaitTransaction(tx.Hash()); err != nil {
		return utils.EmptyHash, err
	}
	return tx.Hash(), nil
}

func (c *Client) UpgradeECCM(newEccmAddr, ccmpAddr common.Address) (common.Hash, error) {
	ccmp, err := eccmp_abi.NewEthCrossChainManagerProxy(ccmpAddr, c.backend)
	if err != nil {
		return utils.EmptyHash, fmt.Errorf("new EthCrossChainManagerProxy err: %s", err)
	}

	auth := c.makeDeployAuth()
	tx, err := ccmp.UpgradeEthCrossChainManager(auth, newEccmAddr)
	if err != nil {
		return utils.EmptyHash, fmt.Errorf("call upgradeEthCrossChainManager err: %s", err)
	}

	if err := c.WaitTransaction(tx.Hash()); err != nil {
		return utils.EmptyHash, err
	}
	return tx.Hash(), nil
}

func (c *Client) ECCDTransferOwnerShip(eccdAddr, eccmAddr common.Address) (common.Hash, error) {
	eccd, err := eccd_abi.NewEthCrossChainData(eccdAddr, c.backend)
	if err != nil {
		return utils.EmptyHash, fmt.Errorf("new EthCrossChainData err: %s", err)
	}

	auth := c.makeAuth()
	tx, err := eccd.TransferOwnership(auth, eccmAddr)
	if err != nil {
		return utils.EmptyHash, fmt.Errorf("call transferOwnerShip err: %s", err)
	}
	if err := c.WaitTransaction(tx.Hash()); err != nil {
		return utils.EmptyHash, err
	}
	return tx.Hash(), nil
}

func (c *Client) ECCDOwnership(eccdAddr common.Address) (common.Address, error) {
	eccd, err := eccd_abi.NewEthCrossChainData(eccdAddr, c.backend)
	if err != nil {
		return utils.EmptyAddress, err
	}
	return eccd.Owner(nil)
}

func (c *Client) ECCMTransferOwnerShip(eccmAddr, ccmpAddr common.Address) (common.Hash, error) {
	eccm, err := eccm_abi.NewEthCrossChainManager(eccmAddr, c.backend)
	if err != nil {
		return utils.EmptyHash, fmt.Errorf("new EthCrossChainManager err: %s", err)
	}

	auth := c.makeAuth()
	tx, err := eccm.TransferOwnership(auth, ccmpAddr)
	if err != nil {
		return utils.EmptyHash, fmt.Errorf("call transferOwnerShip err: %s", err)
	}
	if err := c.WaitTransaction(tx.Hash()); err != nil {
		return utils.EmptyHash, err
	}
	return tx.Hash(), nil
}

func (c *Client) ECCMOwnership(eccmAddr common.Address) (common.Address, error) {
	eccm, err := eccm_abi.NewEthCrossChainManager(eccmAddr, c.backend)
	if err != nil {
		return utils.EmptyAddress, err
	}
	return eccm.Owner(nil)
}

func (c *Client) CCMPTransferOwnerShip(ccmpAddr, newOwner common.Address) (common.Hash, error) {
	ccmp, err := eccmp_abi.NewEthCrossChainManagerProxy(ccmpAddr, c.backend)
	if err != nil {
		return utils.EmptyHash, err
	}

	auth := c.makeAuth()
	tx, err := ccmp.TransferOwnership(auth, newOwner)
	if err != nil {
		return utils.EmptyHash, fmt.Errorf("call transferOwnerShip err: %s", err)
	}
	if err := c.WaitTransaction(tx.Hash()); err != nil {
		return utils.EmptyHash, err
	}
	return tx.Hash(), nil
}

func (c *Client) CCMPOwnership(ccmpAddr common.Address) (common.Address, error) {
	ccmp, err := eccmp_abi.NewEthCrossChainManagerProxy(ccmpAddr, c.backend)
	if err != nil {
		return utils.EmptyAddress, err
	}
	return ccmp.Owner(nil)
}

func (c *Client) TransferCrossChainAdminOwnership(newOwner common.Address) (common.Hash, error) {
	payload, err := c.packPLT(plt.MethodTransferOwnership, newOwner)
	if err != nil {
		return utils.EmptyHash, err
	}

	hash, err := c.sendPLTTx(payload)
	if err != nil {
		return utils.EmptyHash, err
	}

	if err := c.WaitTransaction(hash); err != nil {
		return utils.EmptyHash, err
	}

	return hash, nil
}

func (c *Client) CrossChainAdminOwnership(blockNum string) (common.Address, error) {
	payload, err := c.packPLT(plt.MethodOwnership)
	if err != nil {
		return utils.EmptyAddress, err
	}

	enc, err := c.callPLT(payload, blockNum)
	if err != nil {
		return utils.EmptyAddress, err
	}

	output := new(plt.MethodOwnershipOutput)
	if err := c.unpackPLT(plt.MethodOwnership, output, enc); err != nil {
		return utils.EmptyAddress, err
	}

	return output.Owner, nil
}

func (c *Client) SetPLTCCMP(ccmp common.Address) (common.Hash, error) {
	payload, err := c.packPLT(plt.MethodSetManagerProxy, ccmp)
	if err != nil {
		return utils.EmptyHash, err
	}
	return c.sendPLTTx(payload)
}

func (c *Client) GetPLTCCMP(blockNum string) (common.Address, error) {
	payload, err := c.packPLT(plt.MethodGetManagerProxy)
	if err != nil {
		return utils.EmptyAddress, err
	}

	enc, err := c.callPLT(payload, blockNum)
	if err != nil {
		return utils.EmptyAddress, err
	}

	var proxy common.Address
	if err := c.unpackPLT(plt.MethodGetManagerProxy, &proxy, enc); err != nil {
		return utils.EmptyAddress, err
	}

	return proxy, nil
}

func (c *Client) BindPLTProxy(targetChainID uint64, targetProxy common.Address) (common.Hash, error) {
	payload, err := c.packPLT(plt.MethodBindProxy, targetChainID, targetProxy.Bytes())
	if err != nil {
		return utils.EmptyHash, err
	}
	return c.sendPLTTx(payload)
}

func (c *Client) GetBindPLTProxy(targetChainID uint64, blockNum string) (common.Address, error) {
	payload, err := c.packPLT(plt.MethodGetBindedProxy, targetChainID)
	if err != nil {
		return utils.EmptyAddress, err
	}

	enc, err := c.callPLT(payload, blockNum)
	if err != nil {
		return utils.EmptyAddress, err
	}

	var proxy []byte
	if err := c.unpackPLT(plt.MethodGetBindedProxy, &proxy, enc); err != nil {
		return utils.EmptyAddress, err
	}

	return common.BytesToAddress(proxy), nil
}

func (c *Client) BindPLTAsset(targetChainID uint64, targetAsset common.Address) (common.Hash, error) {
	payload, err := c.packPLT(plt.MethodBindAsset, targetChainID, targetAsset.Bytes())
	if err != nil {
		return utils.EmptyHash, err
	}
	return c.sendPLTTx(payload)
}

func (c *Client) GetBindPLTAsset(targetChainID uint64, blockNum string) (common.Address, error) {
	payload, err := c.packPLT(plt.MethodGetBindedAsset, targetChainID)
	if err != nil {
		return utils.EmptyAddress, err
	}

	enc, err := c.callPLT(payload, blockNum)
	if err != nil {
		return utils.EmptyAddress, err
	}

	var asset []byte
	if err := c.unpackPLT(plt.MethodGetBindedAsset, &asset, enc); err != nil {
		return utils.EmptyAddress, err
	}

	return common.BytesToAddress(asset), nil
}

func (c *Client) PLTMint(to common.Address, val *big.Int) (common.Hash, error) {
	payload, err := c.packPLT(plt.MethodMint, to, val)
	if err != nil {
		return utils.EmptyHash, err
	}
	return c.sendPLTTx(payload)
}

func (c *Client) PLTBurn(val *big.Int) (common.Hash, error) {
	payload, err := c.packPLT(plt.MethodBurn, val)
	if err != nil {
		return utils.EmptyHash, err
	}
	return c.sendPLTTx(payload)
}

func (c *Client) LockPLT(targetChainID uint64, dstAddr common.Address, amount *big.Int) (common.Hash, error) {
	payload, err := c.packPLT(plt.MethodLock, targetChainID, dstAddr.Bytes(), amount)
	if err != nil {
		return utils.EmptyHash, err
	}
	return c.sendPLTTx(payload)
}

func (c *Client) SetNFTCCMP(proxyAddr, ccmp common.Address) (common.Hash, error) {
	proxy, err := nftlp.NewNFTLockProxy(proxyAddr, c.backend)
	if err != nil {
		return utils.EmptyHash, err
	}

	auth := c.makeAuth()
	tx, err := proxy.SetManagerProxy(auth, ccmp)
	if err != nil {
		return utils.EmptyHash, err
	}
	if err := c.WaitTransaction(tx.Hash()); err != nil {
		return utils.EmptyHash, err
	}
	return tx.Hash(), nil
}

func (c *Client) GetNFTCCMP(proxyAddr common.Address) (common.Address, error) {
	proxy, err := nftlp.NewNFTLockProxy(proxyAddr, c.backend)
	if err != nil {
		return utils.EmptyAddress, err
	}

	return proxy.ManagerProxyContract(nil)
}

func (c *Client) DeployNFTProxy() (common.Address, error) {
	auth := c.makeDeployAuth()
	addr, tx, _, err := nftlp.DeployNFTLockProxy(auth, c.backend)
	if err != nil {
		return utils.EmptyAddress, err
	}
	if err := c.WaitTransaction(tx.Hash()); err != nil {
		return utils.EmptyAddress, err
	}
	return addr, nil
}

func (c *Client) BindNFTProxy(
	localLockProxy common.Address,
	targetLockProxy common.Address,
	targetSideChainID uint64,
) (common.Hash, error) {

	proxy, err := nftlp.NewNFTLockProxy(localLockProxy, c.backend)
	if err != nil {
		return utils.EmptyHash, err
	}

	auth := c.makeAuth()
	tx, err := proxy.BindProxyHash(auth, targetSideChainID, targetLockProxy.Bytes())
	if err != nil {
		return utils.EmptyHash, err
	}
	if err := c.WaitTransaction(tx.Hash()); err != nil {
		return utils.EmptyHash, err
	}
	return tx.Hash(), nil
}

func (c *Client) GetBoundNFTProxy(
	localLockProxy common.Address,
	targetSideChainID uint64,
) (common.Address, error) {

	proxy, err := nftlp.NewNFTLockProxy(localLockProxy, c.backend)
	if err != nil {
		return utils.EmptyAddress, err
	}

	bz, err := proxy.ProxyHashMap(nil, targetSideChainID)
	if err != nil {
		return utils.EmptyAddress, err
	}

	return common.BytesToAddress(bz), nil
}

func (c *Client) TransferNFTProxyOwnership(proxyAddr, newOwner common.Address) (common.Hash, error) {
	proxy, err := nftlp.NewNFTLockProxy(proxyAddr, c.backend)
	if err != nil {
		return utils.EmptyHash, err
	}

	auth := c.makeAuth()
	tx, err := proxy.TransferOwnership(auth, newOwner)
	if err != nil {
		return utils.EmptyHash, err
	}
	if err := c.WaitTransaction(tx.Hash()); err != nil {
		return utils.EmptyHash, err
	}
	return tx.Hash(), nil
}

func (c *Client) NFTProxyOwnership(proxyAddr common.Address) (common.Address, error) {
	proxy, err := nftlp.NewNFTLockProxy(proxyAddr, c.backend)
	if err != nil {
		return utils.EmptyAddress, err
	}
	return proxy.Owner(nil)
}

func (c *Client) BindNFTAsset(
	localLockProxy,
	fromAsset,
	toAsset common.Address,
	targetSideChainID uint64,
) (common.Hash, error) {

	proxy, err := nftlp.NewNFTLockProxy(localLockProxy, c.backend)
	if err != nil {
		return utils.EmptyHash, err
	}

	auth := c.makeAuth()
	tx, err := proxy.BindAssetHash(auth, fromAsset, targetSideChainID, toAsset.Bytes())
	if err != nil {
		return utils.EmptyHash, err
	}
	if err := c.WaitTransaction(tx.Hash()); err != nil {
		return utils.EmptyHash, err
	}
	return tx.Hash(), nil
}

func (c *Client) GetBoundNFTAsset(lockProxy, fromAsset common.Address, toChainID uint64) (common.Address, error) {
	proxy, err := nftlp.NewNFTLockProxy(lockProxy, c.backend)
	if err != nil {
		return utils.EmptyAddress, err
	}

	bz, err := proxy.AssetHashMap(nil, fromAsset, toChainID)
	if err != nil {
		return utils.EmptyAddress, err
	}
	return common.BytesToAddress(bz), nil
}

func (c *Client) InitGenesisBlock(eccmAddr common.Address, rawHdr, publickeys []byte) (common.Hash, error) {
	eccm, err := eccm_abi.NewEthCrossChainManager(eccmAddr, c.backend)
	if err != nil {
		return utils.EmptyHash, fmt.Errorf("new EthCrossChainManager err: %s", err)
	}

	auth := c.makeDeployAuth()
	tx, err := eccm.InitGenesisBlock(auth, rawHdr, publickeys)
	if err != nil {
		return utils.EmptyHash, fmt.Errorf("call eccm InitGenesisBlock err: %s", err)
	}

	if err := c.WaitTransaction(tx.Hash()); err != nil {
		return utils.EmptyHash, err
	}
	return tx.Hash(), nil
}
