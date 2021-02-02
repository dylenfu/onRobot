package core

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/contracts/native/plt"
	"github.com/ethereum/go-ethereum/contracts/native/utils"
	"github.com/palettechain/onRobot/config"
	"github.com/palettechain/onRobot/pkg/eth"
	"github.com/palettechain/onRobot/pkg/log"
	"github.com/palettechain/onRobot/pkg/sdk"
)

// 在palette native plt合约mint一定量的PLT token到某个已经存在的用户地址
func PLTMint() (succeed bool) {
	var p struct {
		To    common.Address
		Value int
	}
	if err := config.LoadParams("PLT-Mint.json", &p); err != nil {
		log.Error(err)
		return
	}
	amount := plt.MultiPLT(p.Value)

	toBalanceBeforeTrans, err := admcli.BalanceOf(p.To, "latest")
	if err != nil {
		log.Error(err)
		return
	}
	tx, err := admcli.PLTMint(p.To, amount)
	if err != nil {
		log.Error(err)
		return
	}
	wait(2)
	if err := admcli.DumpEventLog(tx); err != nil {
		log.Error(err)
		return
	}

	toBalanceAfterTrans, err := admcli.BalanceOf(p.To, "latest")
	if err != nil {
		log.Error(err)
		return
	}

	if utils.SafeSub(toBalanceAfterTrans, toBalanceBeforeTrans).Cmp(amount) != 0 {
		log.Errorf("wrong mint value %s and should be %s",
			utils.SafeSub(toBalanceAfterTrans, toBalanceBeforeTrans), p.Value)
	}
	return true
}

// 在palette native plt合约烧毁合约PLT总供应量对应的amount
func PLTBurn() (succeed bool) {
	var p struct {
		Value int
	}
	if err := config.LoadParams("PLT-Burn.json", &p); err != nil {
		log.Error(err)
		return
	}
	amount := plt.MultiPLT(p.Value)

	toBalanceBeforeTrans, err := admcli.BalanceOf(admcli.Address(), "latest")
	if err != nil {
		log.Error(err)
		return
	}
	if toBalanceBeforeTrans.Cmp(amount) == -1 {
		log.Error(err)
		return
	}

	tx, err := admcli.PLTBurn(amount)
	if err != nil {
		log.Error(err)
		return
	}
	wait(2)
	if err := admcli.DumpEventLog(tx); err != nil {
		log.Error(err)
		return
	}

	toBalanceAfterTrans, err := admcli.BalanceOf(admcli.Address(), "latest")
	if err != nil {
		log.Error(err)
		return
	}

	if utils.SafeSub(toBalanceBeforeTrans, toBalanceAfterTrans).Cmp(amount) != 0 {
		log.Errorf("wrong mint value %s and should be %s",
			utils.SafeSub(toBalanceAfterTrans, toBalanceBeforeTrans), p.Value)
	}

	return true
}

func PLTLock() (succeed bool) {
	var params struct {
		From   common.Address
		To     common.Address
		Amount int
	}
	if err := config.LoadParams("PLT-Lock.json", &params); err != nil {
		log.Error(err)
		return
	}

	baseUrl := config.Conf.Nodes[0].RPCAddr()
	privKey := config.LoadAccount(params.From.Hex())
	userAddr := params.From
	bindTo := params.To
	cli := sdk.NewSender(baseUrl, privKey)
	amount := plt.MultiPLT(params.Amount)
	targetSideChainID := config.Conf.CrossChain.EthereumSideChainID
	ethAsset := config.Conf.CrossChain.EthereumPLTAsset

	fromBalanceBeforeLockOnPalette, err := cli.BalanceOf(userAddr, "latest")
	if err != nil {
		log.Error(err)
		return
	}
	{
		if fromBalanceBeforeLockOnPalette.Cmp(amount) < 0 {
			logsplit()
			log.Infof("prepare test account balance...")
			if _, err := admcli.PLTTransfer(userAddr, amount); err != nil {
				log.Error(err)
				return
			}
			fromBalanceBeforeLockOnPalette, _ = cli.BalanceOf(userAddr, "latest")
		}
	}

	toBalanceBeforeLockOnPalette, err := cli.BalanceOf(bindTo, "latest")
	if err != nil {
		log.Error(err)
		return
	}
	fromBalanceBeforeLockOnEthereum, err := ethInvoker.PLTBalanceOf(ethAsset, userAddr)
	if err != nil {
		log.Error(err)
		return
	}
	toBalanceBeforeLockOnEthereum, err := ethInvoker.PLTBalanceOf(ethAsset, bindTo)
	if err != nil {
		log.Error(err)
		return
	}

	logsplit()
	hash, err := cli.LockPLT(targetSideChainID, bindTo, amount)
	if err != nil {
		log.Errorf("failed to call `lock` err: %v", err)
		return
	}

	logsplit()
	fromBalanceAfterLockOnPalette, err := cli.BalanceOf(userAddr, "latest")
	if err != nil {
		log.Error(err)
		return
	}
	fromBalanceAfterLockOnEthereum, err := ethInvoker.PLTBalanceOf(ethAsset, userAddr)
	if err != nil {
		log.Error(err)
		return
	}
	toBalanceAfterLockOnPalette, err := cli.BalanceOf(bindTo, "latest")
	if err != nil {
		log.Error(err)
		return
	}
	toBalanceAfterLockOnEthereum, err := ethInvoker.PLTBalanceOf(ethAsset, bindTo)
	if err != nil {
		log.Error(err)
		return
	}

	log.Infof("tx hash %s, \r\n"+
		"from %s: balance %d before lock on palette, balance %d after lock on palette, balance %d before lock on palette, balance %d after lock on ethereum\r\n"+
		"to %s: balance %d before lock on palette, balance %d after lock on palette, balance %d before lock on palette, balance %d after lock on ethereum\r\n",
		hash.Hex(),
		userAddr.Hex(),
		plt.PrintUPLT(fromBalanceBeforeLockOnPalette),
		plt.PrintUPLT(fromBalanceAfterLockOnPalette),
		plt.PrintUPLT(fromBalanceBeforeLockOnEthereum),
		plt.PrintUPLT(fromBalanceAfterLockOnEthereum),
		bindTo.Hex(),
		plt.PrintUPLT(toBalanceBeforeLockOnPalette),
		plt.PrintUPLT(toBalanceAfterLockOnPalette),
		plt.PrintUPLT(toBalanceBeforeLockOnEthereum),
		plt.PrintUPLT(toBalanceAfterLockOnEthereum),
	)

	return true
}

func PLTUnlock() (succeed bool) {
	var params struct {
		From   common.Address
		To     common.Address
		Amount int
	}
	if err := config.LoadParams("PLT-UnLock.json", &params); err != nil {
		log.Error(err)
		return
	}

	privKey := config.LoadAccount(params.From.Hex())
	toAddr := params.To
	proxy := config.Conf.CrossChain.EthereumPLTProxy
	targetSideChainID := config.Conf.CrossChain.PaletteSideChainID
	fromAsset := config.Conf.CrossChain.EthereumPLTAsset
	amount := plt.MultiPLT(params.Amount)

	invoker := eth.NewEInvoker(
		config.Conf.CrossChain.EthereumSideChainID,
		config.Conf.CrossChain.EthereumRPCUrl,
		privKey,
	)

	hash, err := invoker.PLTLock(proxy, fromAsset, targetSideChainID, toAddr, amount)
	if err != nil {
		log.Error(err)
		return
	}

	for i := 0; i < 100; i++ {
		wait(1)
		fromBalance, err := admcli.BalanceOf(params.From, "latest")
		if err == nil {
			log.Infof("%s balance %d", params.From.Hex(), plt.PrintUPLT(fromBalance))
		}
		toBalance, err := admcli.BalanceOf(params.To, "latest")
		if err == nil {
			log.Infof("%s balance %d", params.To.Hex(), plt.PrintUPLT(toBalance))
		}

		if amount.Cmp(utils.SafeSub(toBalance, fromBalance)) == 0 {
			log.Infof("%s unlock %d to %s success! hash %s",
				params.From.Hex(), params.Amount, params.To.Hex(), hash.Hex())
		}
	}

	return true
}

//func PLTUnlock() (succeed bool) {
//	var params = struct {
//		Proof        string
//		RawHeader    string
//		HeaderProof  string
//		CurRawHeader string
//		HeaderSig    string
//		UnlockTo common.Address
//	}{}
//
//	if err := config.LoadParams("PLT-Unlock.json", &params); err != nil {
//		log.Error(err)
//		return
//	}
//
//	balanceBeforeUnlock, _ := admcli.BalanceOf(params.UnlockTo, "latest")
//
//	proof, _ := hexutil.Decode(params.Proof)
//	rawHeader, _ := hexutil.Decode(params.RawHeader)
//	headerProof, _ := hexutil.Decode(params.HeaderProof)
//	curRawHeader, _ := hexutil.Decode(params.CurRawHeader)
//	headerSig, _ := hexutil.Decode(params.HeaderSig)
//
//	eccm := config.Conf.CrossChain.EthereumECCM
//	hash, err := ethInvoker.VerifyAndExecuteTx(
//		eccm,
//		proof,
//		rawHeader,
//		headerProof,
//		curRawHeader,
//		headerSig,
//	)
//	if err != nil {
//		log.Error(err)
//		return
//	}
//
//	for i := 0; i < 10000; i++ {
//		balance, err := admcli.BalanceOf(params.UnlockTo, "latest")
//		if err != nil {
//			log.Error(err)
//			return
//		}
//		if balance.Cmp(balanceBeforeUnlock) > 0 {
//			subAmount := utils.SafeSub(balance, balanceBeforeUnlock)
//			log.Infof("balance before unlock %d, after unlock %d, the sub amount is %d, eth hash %s",
//				plt.PrintUPLT(balanceBeforeUnlock), plt.PrintUPLT(balance), plt.PrintUPLT(subAmount), hash.Hex())
//			break
//		}
//		time.Sleep(3 * time.Second)
//	}
//
//	return true
//}
