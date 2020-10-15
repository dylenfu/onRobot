package sdk

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/contracts/native/governance"
	"github.com/ethereum/go-ethereum/contracts/native/utils"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

// params are validator and isRevoke
func (c *PaletteClient) AddValidator(validator common.Address, revoke bool) (common.Hash, error) {
	payload, err := utils.PackMethod(GovernanceABI, governance.MethodAddValidator, validator, revoke)
	if err != nil {
		return common.Hash{}, err
	}

	return c.SendGovernanceTransaction(c.Admin.PrivateKey, payload)
}

func (c *PaletteClient) GetRewardRecordBlock(blockNum string) (*big.Int, error) {
	payload, err := utils.PackMethod(GovernanceABI, governance.MethodGetRewardRecordBlockHeight)
	if err != nil {
		return nil, err
	}

	enc, err := c.CallContract(c.AdminAddress(), GovernanceAddress, payload, blockNum)
	if err != nil {
		return nil, fmt.Errorf("failed to get reward record block: [%v]", err)
	}

	output := new(governance.MethodGetRewardRecordBlockHeightOutput)
	err = utils.UnpackOutputs(GovernanceABI, governance.MethodGetRewardRecordBlockHeight, output, enc)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack encode bytes [%v]: [%v]", common.Bytes2Hex(enc), err)
	}

	return output.Value, nil
}

func (c *PaletteClient) GetLatestRewardProposer(blockNum string) (common.Address, error) {
	payload, err := utils.PackMethod(GovernanceABI, governance.MethodGetLastRewardProposer)
	if err != nil {
		return utils.EmptyAddress, err
	}

	enc, err := c.CallContract(c.AdminAddress(), GovernanceAddress, payload, blockNum)
	if err != nil {
		return utils.EmptyAddress, fmt.Errorf("failed to get latest reward proposer: [%v]", err)
	}

	output := new(governance.MethodGetLastRewardProposerOutput)
	err = utils.UnpackOutputs(GovernanceABI, governance.MethodGetLastRewardProposer, output, enc)
	if err != nil {
		return utils.EmptyAddress, fmt.Errorf("failed to unpack encode bytes [%v]: [%v]", common.Bytes2Hex(enc), err)
	}

	return output.Proposer, nil
}

func (c *PaletteClient) SendGovernanceTransaction(key *ecdsa.PrivateKey, payload []byte) (common.Hash, error) {
	addr := crypto.PubkeyToAddress(key.PublicKey)

	nonce, err := c.GetNonce(addr.Hex())
	if err != nil {
		return common.Hash{}, err
	}
	tx := types.NewTransaction(
		nonce,
		GovernanceAddress,
		big.NewInt(0),
		Gas1GW,
		big.NewInt(GasPrice),
		payload,
	)

	signedTx, err := c.SignTransaction(key, tx)
	if err != nil {
		return common.Hash{}, err
	}
	return c.SendRawTransaction(signedTx)
}
