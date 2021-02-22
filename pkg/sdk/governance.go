package sdk

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/contracts/native/governance"
	"github.com/ethereum/go-ethereum/contracts/native/utils"
)

// params are validator and isRevoke
func (c *Client) AddValidator(validator, stakeAccount common.Address, revoke bool) (common.Hash, error) {
	payload, err := utils.PackMethod(GovernanceABI, governance.MethodAddValidator, validator, stakeAccount, revoke)
	if err != nil {
		return common.Hash{}, err
	}
	return c.SendGovernanceTx(payload)
}

func (c *Client) Stake(validator, stakeAccount common.Address, amount *big.Int, revoke bool) (common.Hash, error) {
	payload, err := utils.PackMethod(GovernanceABI, governance.MethodStake, validator, stakeAccount, amount, revoke)
	if err != nil {
		return common.Hash{}, err
	}
	return c.SendGovernanceTx(payload)
}

func (c *Client) GetStakeAmount(validator, stakeAccount common.Address, blockNum string) *big.Int {
	payload, err := utils.PackMethod(GovernanceABI, governance.MethodGetStakeAmount, stakeAccount, validator)
	if err != nil {
		return nil
	}

	enc, err := c.CallGovernance(payload, blockNum)
	if err != nil {
		return nil
	}

	output := new(governance.MethodGetStakeAmountOutput)
	err = utils.UnpackOutputs(GovernanceABI, governance.MethodGetStakeAmount, output, enc)
	if err != nil {
		return nil
	}

	return output.Amount
}

func (c *Client) GetValidatorTotalStakeAmount(validator common.Address, blockNum string) *big.Int {
	payload, err := utils.PackMethod(GovernanceABI, governance.MethodGetValidatorTotalStakeAmount, validator)
	if err != nil {
		return nil
	}

	enc, err := c.CallGovernance(payload, blockNum)
	if err != nil {
		return nil
	}

	output := new(governance.MethodGetValidatorTotalStakeAmountOutput)
	err = utils.UnpackOutputs(GovernanceABI, governance.MethodGetValidatorTotalStakeAmount, output, enc)
	if err != nil {
		return nil
	}

	return output.Amount
}

func (c *Client) CheckValidator(validator common.Address, blockNum string) bool {
	payload, err := utils.PackMethod(GovernanceABI, governance.MethodCheckValidator, validator)
	if err != nil {
		return false
	}

	enc, err := c.CallGovernance(payload, blockNum)
	if err != nil {
		return false
	}

	output := new(governance.MethodCheckValidatorOutput)
	err = utils.UnpackOutputs(GovernanceABI, governance.MethodCheckValidator, output, enc)
	if err != nil {
		return false
	}

	return output.Succeed
}

func (c *Client) GetEffectiveValidators(blockNum string) []common.Address {
	payload, err := utils.PackMethod(GovernanceABI, governance.MethodGetEffectiveValidators)
	if err != nil {
		return nil
	}

	enc, err := c.CallGovernance(payload, blockNum)
	if err != nil {
		return nil
	}

	output := new(governance.MethodGetEffectiveValidatorsOutput)
	err = utils.UnpackOutputs(GovernanceABI, governance.MethodGetEffectiveValidators, output, enc)
	if err != nil {
		return nil
	}

	return output.List
}

func (c *Client) GetAllValidators(blockNum string) []common.Address {
	payload, err := utils.PackMethod(GovernanceABI, governance.MethodGetAllValidators)
	if err != nil {
		return nil
	}

	enc, err := c.CallGovernance(payload, blockNum)
	if err != nil {
		return nil
	}

	output := new(governance.MethodGetAllValidatorsOutput)
	err = utils.UnpackOutputs(GovernanceABI, governance.MethodGetAllValidators, output, enc)
	if err != nil {
		return nil
	}

	return output.List
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

func (c *Client) GetLastRewardBlock(blockNum string) (*big.Int, error) {
	payload, err := utils.PackMethod(GovernanceABI, governance.MethodGetLastRewardBlockHeight)
	if err != nil {
		return utils.EmptyBig, err
	}

	enc, err := c.CallGovernance(payload, blockNum)
	if err != nil {
		return utils.EmptyBig, fmt.Errorf("failed to get latest reward block number: [%v]", err)
	}

	output := new(governance.MethodGetLastRewardBlockHeightOutput)
	err = utils.UnpackOutputs(GovernanceABI, governance.MethodGetLastRewardBlockHeight, output, enc)
	if err != nil {
		return utils.EmptyBig, fmt.Errorf("failed to unpack encode bytes [%v]: [%v]", common.Bytes2Hex(enc), err)
	}

	return output.Value, nil
}

func (c *Client) Propose(proposalType uint8, value *big.Int) (common.Hash, error) {
	payload, err := utils.PackMethod(GovernanceABI, governance.MethodPropose, proposalType, value)
	if err != nil {
		return utils.EmptyHash, err
	}

	return c.SendGovernanceTx(payload)
}

func (c *Client) GetProposalFromReceipt(hash common.Hash) (common.Address, *governance.MethodGetProposalOutput, error) {
	raw, err := c.GetReceipt(hash)
	if err != nil {
		return utils.EmptyAddress, nil, err
	}
	if len(raw.Logs) < 1 {
		return utils.EmptyAddress, nil, fmt.Errorf("receipt %s has no logs", hash.Hex())
	}
	event := raw.Logs[0]

	output := new(governance.MethodGetProposalOutput)
	output.Proposer = utils.Hash2Address(event.Topics[1])
	proposalID := utils.Hash2Address(event.Topics[2])
	output.ProposalType = utils.Hash2Uint8(event.Topics[3])
	output.EndBlock = utils.Hash2Big(event.Topics[4])

	return proposalID, output, nil
}

func (c *Client) GetProposal(proposalID common.Address, blockNum string) (*governance.MethodGetProposalOutput, error) {
	payload, err := utils.PackMethod(GovernanceABI, governance.MethodGetProposal, proposalID)
	if err != nil {
		return nil, err
	}

	enc, err := c.CallGovernance(payload, blockNum)
	if err != nil {
		return nil, err
	}

	output := new(governance.MethodGetProposalOutput)
	err = utils.UnpackOutputs(GovernanceABI, governance.MethodGetProposal, output, enc)
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (c *Client) Vote(proposalID common.Address) (common.Hash, error) {
	payload, err := utils.PackMethod(GovernanceABI, governance.MethodVote, proposalID)
	if err != nil {
		return utils.EmptyHash, err
	}

	return c.SendGovernanceTx(payload)
}

func (c *Client) GetGlobalParams(proposalType uint8, blockNum string) (*big.Int, error) {
	payload, err := utils.PackMethod(GovernanceABI, governance.MethodGetGlobalParams, proposalType)
	if err != nil {
		return nil, err
	}

	enc, err := c.CallGovernance(payload, blockNum)
	if err != nil {
		return nil, err
	}

	output := new(governance.MethodGetGlobalParamsOutput)
	err = utils.UnpackOutputs(GovernanceABI, governance.MethodGetGlobalParams, output, enc)
	if err != nil {
		return nil, err
	}

	return output.Value, nil
}

func (c *Client) Reward(validators []common.Address, blockNum *big.Int) (common.Hash, error) {
	payload, err := utils.PackMethod(GovernanceABI, governance.MethodReward, validators, blockNum)
	if err != nil {
		return utils.EmptyHash, err
	}

	return c.SendGovernanceTx(payload)
}

func (c *Client) packGovernance(method string, args ...interface{}) ([]byte, error) {
	return utils.PackMethod(GovernanceABI, method, args...)
}
func (c *Client) unpackGovernance(method string, output interface{}, enc []byte) error {
	return utils.UnpackOutputs(GovernanceABI, method, output, enc)
}
func (c *Client) SendGovernanceTx(payload []byte) (common.Hash, error) {
	hash, err := c.SendTransaction(GovernanceAddress, payload)
	if err != nil {
		return utils.EmptyHash, err
	}
	if err := c.WaitTransaction(hash); err != nil {
		return utils.EmptyHash, err
	}
	return hash, nil
}
func (c *Client) CallGovernance(payload []byte, blockNum string) ([]byte, error) {
	return c.CallContract(c.Address(), GovernanceAddress, payload, blockNum)
}
