package sdk

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/palettechain/onRobot/pkg/log"
	common2 "github.com/polynetwork/poly/common"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/contracts/native/nft"
	"github.com/ethereum/go-ethereum/contracts/native/nftmanager"
	"github.com/ethereum/go-ethereum/contracts/native/utils"
)

func (c *Client) NFTDeploy(name string, symbol string) (common.Hash, common.Address, error) {
	payload, err := c.packNFTManager(nftmanager.MethodDeploy, name, symbol, c.Address())
	if err != nil {
		return utils.EmptyHash, utils.EmptyAddress, err
	}

	hash, err := c.sendNFTManager(payload)
	if err != nil {
		return utils.EmptyHash, utils.EmptyAddress, err
	}

	c.WaitTransaction(hash)
	receipts, err := c.GetReceipt(hash)
	if err != nil {
		return utils.EmptyHash, utils.EmptyAddress, fmt.Errorf("nft depoly - get receipt err: %s", err)
	}
	if len(receipts.Logs) == 0 {
		return utils.EmptyHash, utils.EmptyAddress, fmt.Errorf("invalid tx, no receipts events")
	}

	for _, event := range receipts.Logs {
		if event.Topics[0] == NFTABI.Events[nft.EventDeploy].ID() {
			return hash, event.Address, nil
		}
	}

	return utils.EmptyHash, utils.EmptyAddress, fmt.Errorf("no valid nft address")
}


func (c *Client) NFTName(asset common.Address, blockNum string) (string, error) {
	payload, err := c.packNFT(nft.MethodName)
	if err != nil {
		return "", err
	}
	data, err := c.callNFT(asset, payload, blockNum)
	if err != nil {
		return "", err
	}
	result := &nft.NameResult{}
	err = c.unpackNFT(nft.MethodName, result, data)
	if err != nil{
		return "", err
	}

	return result.Name, nil
}

func (c *Client) NFTSymbol(asset common.Address, blockNum  string) (string, error) {
	payload, err := c.packNFT(nft.MethodSymbol)
	if err != nil {
		return "", err
	}
	data, err := c.callNFT(asset, payload, blockNum)
	if err != nil {
		return "", err
	}
	result := &nft.SymbolResult{}
	err = c.unpackNFT(nft.MethodSymbol, result, data)
	if err != nil{
		return "", err
	}

	return result.Symbol, nil
}

func (c *Client) NFTOwner(asset common.Address, blockNum string) (common.Address, error) {
	payload, err := c.packNFT(nft.MethodOwner)
	if err != nil {
		return utils.EmptyAddress, err
	}
	data, err := c.callNFT(asset, payload, blockNum)
	if err != nil {
		return utils.EmptyAddress, err
	}
	result := &nft.OwnerResult{}
	err = c.unpackNFT(nft.MethodOwner, result, data)
	if err != nil{
		return utils.EmptyAddress, err
	}

	return result.Owner, nil
}

func (c *Client) NFTTotalSupply(asset common.Address, blockNum string) (*big.Int, error) {
	payload, err := c.packNFT(nft.MethodTotalSupply)
	if err != nil {
		return utils.EmptyBig, err
	}
	data, err := c.callNFT(asset, payload, blockNum)
	if err != nil {
		return utils.EmptyBig, err
	}
	result := &nft.TotalSupplyResult{}
	err = c.unpackNFT(nft.MethodTotalSupply, result, data)
	if err != nil{
		return utils.EmptyBig, err
	}

	return result.Supply, nil
}

// NFTMint validator mint asset to someone owner
func (c *Client) NFTMint(asset common.Address, owner common.Address, tokenID *big.Int, uri string) (common.Hash, error) {
	payload, err := c.packNFT(nft.MethodMint, owner, tokenID, uri)
	if err != nil {
		return utils.EmptyHash, err
	}

	hash, err := c.sendNFT(asset, payload)
	if err != nil {
		return utils.EmptyHash, err
	}
	return hash, nil
}

func (c *Client) NFTBurn(asset common.Address, tokenID *big.Int) (common.Hash, error) {
	payload, err := c.packNFT(nft.MethodBurn, tokenID)
	if err != nil {
		return common.BytesToHash([]byte{}), err
	}

	hash, err := c.sendNFT(asset, payload)
	if err != nil {
		return utils.EmptyHash, err
	}
	return hash, nil
}

func (c *Client) NFTBalance(asset, user common.Address, blockNum string) (*big.Int, error) {
	payload, err := c.packNFT(nft.MethodBalanceOf, user)
	if err != nil {
		return big.NewInt(0), err
	}

	data, err := c.callNFT(asset, payload, blockNum)
	if err != nil {
		return big.NewInt(0), err
	}

	result := &nft.BalanceOfResult{}
	err = c.unpackNFT(nft.MethodBalanceOf, result, data)
	if err != nil{
		return big.NewInt(0), err
	}

	return result.Balance, nil
}

func (c *Client) NFTTransferFrom(
	asset common.Address,
	from common.Address,
	to common.Address,
	tokenID *big.Int,
	) (common.Hash, error) {
	payload, err := c.packNFT(nft.MethodTransferFrom, from, to, tokenID)
	if err != nil {
		return utils.EmptyHash, err
	}

	return c.sendNFT(asset, payload)
}

// safe transfer to contract
func (c *Client) NFTSafeTransferFrom(
	asset common.Address,
	from common.Address,
	proxy common.Address,
	tokenID *big.Int,
	to common.Address,
	toChainID uint64,
	) (common.Hash, error) {

	data := assembleSafeTransferCallData(to, toChainID)
	log.Infof("asset %s, from %s, proxy %s, tokenID %d, data %s",
		asset.Hex(), from.Hex(), proxy.Hex(), tokenID.Uint64(), hexutil.Encode(data))

	payload, err := c.packNFT(nft.MethodSafeTransferFrom, from, proxy, tokenID, data)
	if err != nil {
		return utils.EmptyHash, err
	}

	return c.sendNFT(asset, payload)
}

func (c *Client) NFTApprove(
	asset common.Address,
	to common.Address,
	tokenID *big.Int,
	) (common.Hash, error) {
	payload, err := c.packNFT(nft.MethodApprove, to, tokenID)
	if err != nil {
		return utils.EmptyHash, err
	}

	return c.sendNFT(asset, payload)
}

func (c *Client) NFTGetApproved(asset common.Address, amount *big.Int, blockNum string) (common.Address, error) {
	payload, err := c.packNFT(nft.MethodGetApproved, amount)
	if err != nil {
		return utils.EmptyAddress, err
	}

	data, err := c.callNFT(asset, payload, blockNum)
	if err != nil {
		return utils.EmptyAddress, err
	}

	result := &nft.GetApprovedResult{}
	err = c.unpackNFT(nft.MethodGetApproved, result, data)
	if err != nil{
		return utils.EmptyAddress, err
	}

	return result.Spender, nil
}

func assembleSafeTransferCallData(toAddress common.Address, chainID uint64) []byte {
	sink := common2.NewZeroCopySink(nil)
	sink.WriteVarBytes(toAddress.Bytes())
	sink.WriteUint64(chainID)
	return sink.Bytes()
}

// NFT
func (c *Client) packNFT(method string, args ...interface{}) ([]byte, error) {
	return utils.PackMethod(NFTABI, method, args...)
}
func (c *Client) unpackNFT(method string, output interface{}, enc []byte) error {
	return utils.UnpackOutputs(NFTABI, method, output, enc)
}
func (c *Client) sendNFT(nftAddr common.Address, payload []byte) (common.Hash, error) {
	return c.SendTransaction(nftAddr, payload)
}
func (c *Client) callNFT(nftAddr common.Address, payload []byte, blockNum string) ([]byte, error) {
	return c.CallContract(c.Address(), nftAddr, payload, blockNum)
}

// nft manager
func (c *Client) packNFTManager(method string, args ...interface{}) ([]byte, error) {
	return utils.PackMethod(NFTManagerABI, method, args...)
}
func (c *Client) unpackNFTManager(method string, output interface{}, enc []byte) error {
	return utils.UnpackOutputs(NFTManagerABI, method, output, enc)
}
func (c *Client) sendNFTManager(payload []byte) (common.Hash, error) {
	return c.SendTransaction(NFTMangerAddress, payload)
}
func (c *Client) callNFTManager(payload []byte, blockNum string) ([]byte, error) {
	return c.CallContract(c.Address(), NFTMangerAddress, payload, blockNum)
}
