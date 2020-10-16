package sdk

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/contracts/native/governance"
	"github.com/ethereum/go-ethereum/contracts/native/utils"
)

// params are validator and isRevoke
func (c *Client) AddValidator(validator common.Address, revoke bool) (common.Hash, error) {
	payload, err := utils.PackMethod(GovernanceABI, governance.MethodAddValidator, validator, revoke)
	if err != nil {
		return common.Hash{}, err
	}
	return c.SendGovernanceTx(payload)
}

// todo
func (c *Client) GetValidators() []common.Address {
	//payload, err := utils.PackMethod(GovernanceABI, )
	return nil
}

func (c *Client) GetRewardRecordBlock(blockNum string) (*big.Int, error) {
	payload, err := utils.PackMethod(GovernanceABI, governance.MethodGetRewardRecordBlockHeight)
	if err != nil {
		return nil, err
	}

	enc, err := c.CallGovernance(payload, blockNum)
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

func (c *Client) GetLatestRewardProposer(blockNum string) (common.Address, error) {
	payload, err := utils.PackMethod(GovernanceABI, governance.MethodGetLastRewardProposer)
	if err != nil {
		return utils.EmptyAddress, err
	}

	enc, err := c.CallGovernance(payload, blockNum)
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

func (c *Client) packGovernance(method string, args ...interface{}) ([]byte, error) {
	return utils.PackMethod(GovernanceABI, method, args...)
}
func (c *Client) SendGovernanceTx(payload []byte) (common.Hash, error) {
	return c.SendTransaction(GovernanceAddress, payload)
}
func (c *Client) CallGovernance(payload []byte, blockNum string) ([]byte, error) {
	return c.CallContract(c.Address(), GovernanceAddress, payload, blockNum)
}
