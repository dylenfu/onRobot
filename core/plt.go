package core

import (
	"bytes"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/contracts/native"
	"github.com/ethereum/go-ethereum/contracts/native/plt"
	"github.com/ethereum/go-ethereum/contracts/native/utils"
	"github.com/palettechain/onRobot/config"
	"github.com/palettechain/onRobot/pkg/log"
	"github.com/palettechain/onRobot/pkg/poly"
	"github.com/palettechain/onRobot/pkg/sdk"
)

func TotalSupply() (succeed bool) {
	expect := 1e9

	totalSupply, err := admcli.PLTTotalSupply("latest")
	if err != nil {
		log.Error(err)
		return
	}

	actual := plt.PrintUPLT(totalSupply)
	if actual != uint64(expect) {
		log.Errorf("totalSupply expect %d actually %d", expect, actual)
		return
	}

	log.Infof("totalSupply %d", utils.UnsafeDiv(totalSupply, plt.OnePLT))

	return true
}

func Decimal() (succeed bool) {
	data, err := admcli.PLTDecimals()
	if err != nil {
		log.Error(err)
		return
	}

	log.Infof("decimal %d", data)

	return true
}

func Name() (succeed bool) {
	actual, err := admcli.PLTName()
	if err != nil {
		log.Error(err)
		return
	}

	expect := "Palette Token"
	if actual != expect {
		log.Errorf("contract name expect %s actual %s", expect, actual)
		return
	}

	log.Infof("contract name %s", actual)

	return true
}

func AdminBalance() (succeed bool) {
	balance, err := admcli.BalanceOf(config.AdminAddr, "latest")
	if err != nil {
		log.Error(err)
		return
	}

	actual := plt.PrintUPLT(balance)
	log.Infof("admin %s balance %d", config.AdminAddr.Hex(), actual)

	return true
}

func GovernanceBalance() (succeed bool) {
	owner := common.HexToAddress(native.GovernanceContractAddress)
	balance, err := admcli.BalanceOf(owner, "latest")
	if err != nil {
		log.Error(err)
		return
	}

	actual := plt.PrintUPLT(balance)
	log.Infof("governance %s balance %d", owner.Hex(), actual)

	return true
}

func BalanceOf() (succeed bool) {
	var params struct {
		Owner    string
		BlockNum string
	}

	if err := config.LoadParams("BalanceOf.json", &params); err != nil {
		log.Error(err)
		return
	}

	owner := common.HexToAddress(params.Owner)
	balance, err := admcli.BalanceOf(owner, params.BlockNum)
	if err != nil {
		log.Error(err)
		return
	}

	log.Infof("balance %d", utils.UnsafeDiv(balance, plt.OnePLT))

	return true
}

func Transfer() (succeed bool) {
	var params struct {
		From   string
		To     string
		Amount int64
	}

	if err := config.LoadParams("Transfer.json", &params); err != nil {
		log.Error(err)
		return
	}

	key := config.LoadAccount(params.From)
	to := common.HexToAddress(params.To)
	amount := utils.SafeMul(big.NewInt(params.Amount), plt.OnePLT)

	// balance before transfer
	fromBalanceBeforeTrans, err := admcli.BalanceOf(PrivKey2Addr(key), "latest")
	if err != nil {
		log.Error(err)
		return
	}
	toBalanceBeforeTrans, err := admcli.BalanceOf(to, "latest")
	if err != nil {
		log.Error(err)
		return
	}
	if fromBalanceBeforeTrans.Cmp(amount) < 0 {
		log.Errorf("%s balance not enough %d", params.From, utils.UnsafeDiv(fromBalanceBeforeTrans, plt.OnePLT))
		return
	}

	// transfer and waiting for commit
	hash, err := admcli.PLTTransfer(to, amount)
	if err != nil {
		log.Error(err)
		return
	}
	wait(2)
	if err := admcli.DumpEventLog(hash); err != nil {
		log.Error(err)
		return
	}

	// balance after transfer
	fromBalanceAfterTrans, err := admcli.BalanceOf(PrivKey2Addr(key), "latest")
	if err != nil {
		log.Error(err)
		return
	}
	toBalanceAfterTrans, err := admcli.BalanceOf(to, "latest")
	if err != nil {
		log.Error(err)
		return
	}

	// expect sum
	if utils.SafeAdd(toBalanceBeforeTrans, amount).Cmp(toBalanceAfterTrans) != 0 {
		log.Errorf("dst balance before transfer %d, balance after transfer %d, amount %d",
			utils.UnsafeDiv(toBalanceBeforeTrans, plt.OnePLT),
			utils.UnsafeDiv(toBalanceAfterTrans, plt.OnePLT),
			params.Amount,
		)
		return
	}
	if utils.SafeSub(fromBalanceBeforeTrans, amount).Cmp(fromBalanceAfterTrans) != 0 {
		log.Errorf("src balance before transfer %d, balance after transfer %d, amount %d",
			utils.UnsafeDiv(fromBalanceAfterTrans, plt.OnePLT),
			utils.UnsafeDiv(fromBalanceAfterTrans, plt.OnePLT),
			params.Amount,
		)
		return
	}

	return true
}

func Approve() (succeed bool) {
	var params struct {
		Owner   string
		Spender string
		Amount  int
	}

	if err := config.LoadParams("Approve.json", &params); err != nil {
		log.Error(err)
		return
	}

	baseUrl := config.Conf.Nodes[0].RPCAddr()
	key := config.LoadAccount(params.Owner)
	cli := sdk.NewSender(baseUrl, key)

	owner := PrivKey2Addr(key)
	spender := common.HexToAddress(params.Spender)
	amount := plt.MultiPLT(params.Amount)

	// allowance before approve
	allowanceBeforeApprove, err := cli.PLTAllowance(owner, spender, "latest")
	if err != nil {
		log.Error(err)
		return
	}

	hash, err := cli.PLTApprove(spender, amount)
	if err != nil {
		log.Error(err)
		return
	}
	wait(2)
	if err := cli.DumpEventLog(hash); err != nil {
		log.Error(err)
		return
	}

	// allowance after approve
	allowanceAfterApprove, err := cli.PLTAllowance(owner, spender, "latest")
	if err != nil {
		log.Error(err)
		return
	}

	if allowanceAfterApprove.Cmp(amount) != 0 {
		log.Errorf("owner %s, spender %s, allowance before approve %d, allowance after approve %d, amount %d",
			owner.Hex(), spender.Hex(),
			utils.UnsafeDiv(allowanceBeforeApprove, plt.OnePLT),
			utils.UnsafeDiv(allowanceAfterApprove, plt.OnePLT),
			utils.UnsafeDiv(amount, plt.OnePLT))
		return
	}

	return true
}

// 在palette native plt合约mint一定量的PLT token到某个已经存在的用户地址
func Mint() (succeed bool) {
	var p struct {
		To    common.Address
		Value int
	}
	if err := config.LoadParams("Mint.json", &p); err != nil {
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
func Burn() (succeed bool) {
	var p struct {
		Value int
	}
	if err := config.LoadParams("Burn.json", &p); err != nil {
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

// 在palette合约部署ccmp合约成功之后，需要在plt合约记录管理合约地址
func SetCCMP() (succeed bool) {
	_, _, ccmp, err := config.ReadContracts()
	if err != nil {
		log.Error("read ccmp contract err")
		return
	} else {
		log.Infof("ccmp contract addr %s", ccmp.Hex())
	}

	tx, err := admcli.PLTSetCCMP(ccmp)
	if err != nil {
		log.Error(err)
		return
	}
	log.Infof("tx hash: %s", tx.Hex())
	wait(2)
	proxy, err := admcli.PLTGetCCMP("latest")
	if err != nil {
		log.Error(err)
		return
	}
	if !bytes.Equal(proxy.Bytes(), ccmp.Bytes()) {
		log.Errorf("wrong ccmp: should be %s but get %s", ccmp.Hex(), proxy.Hex())
		return
	}
	return true
}

// 在palette native合约上记录以太坊localProxy地址,
// 这里我们将实现palette->poly->palette的循环，不走ethereum，那么proxy就直接是plt地址，
// asset的地址也是palette plt地址
func BindProxy() (succeed bool) {
	var params struct {
		ChainID uint64
		Proxy   common.Address
	}
	if err := config.LoadParams("BindProxy.json", &params); err != nil {
		log.Error(err)
		return
	}

	// bind proxy
	{
		log.Infof("bind proxy...")
		tx, err := admcli.BindProxy(params.ChainID, params.Proxy)
		if err != nil {
			log.Error(err)
			return
		}
		log.Infof("tx hash: %s", tx.Hex())
		wait(2)

		if err := admcli.DumpEventLog(tx); err != nil {
			log.Error(err)
			return
		}
	}

	// get and compare proxy
	{
		log.Infof("get bind proxy...")
		proxy, err := admcli.GetBindProxy(params.ChainID, "latest")
		if err != nil {
			log.Error(err)
			return
		}

		if !bytes.Equal(proxy.Bytes(), params.Proxy.Bytes()) {
			log.Errorf("wrong proxy: expect  %s but get %s", params.Proxy.Hex(), proxy.Hex())
			return
		} else {
			log.Infof("bind proxy success %s", proxy.Hex())
		}
	}

	return true
}

// 在palette native合约上记录以太坊erc20资产地址
func BindAsset() (succeed bool) {
	var params struct {
		ChainID uint64
		Asset   common.Address
	}

	if err := config.LoadParams("BindAsset.json", &params); err != nil {
		log.Error(err)
		return
	}

	// bind asset
	{
		log.Infof("bind asset...")
		tx, err := admcli.BindAsset(params.ChainID, params.Asset)
		if err != nil {
			log.Error(err)
			return
		}
		wait(2)
		if err := admcli.DumpEventLog(tx); err != nil {
			log.Error(err)
			return
		}
	}

	// get and compare asset
	{
		log.Infof("get bind asset...")
		asset, err := admcli.GetBindAsset(params.ChainID, "latest")
		if err != nil {
			log.Error(err)
			return
		}
		if asset != params.Asset {
			log.Errorf("asset err, expect %s, actual %s", params.Asset.Hex(), asset.Hex())
			return
		} else {
			log.Infof("get asset %s success", asset.Hex())
		}
	}

	return true
}

func Lock() (succeed bool) {
	var params struct {
		ChainID      uint64
		AccountIndex int
		Proxy        common.Address
		Asset        common.Address
		BindTo       common.Address
		Amount       int
	}

	if err := config.LoadParams("Lock.json", &params); err != nil {
		log.Error(err)
		return
	}

	if params.AccountIndex > len(config.Conf.Accounts)-1 {
		log.Errorf("account index out of range")
		return
	}

	user := config.Conf.Accounts[params.AccountIndex]
	privKey := config.LoadAccount(user)
	baseUrl := config.Conf.Nodes[0].RPCAddr()
	userAddr := common.HexToAddress(user)
	cli := sdk.NewSender(baseUrl, privKey)
	amount := plt.MultiPLT(params.Amount)

	// prepare balance
	{
		logsplit()
		log.Infof("prepare test account balance...")
		hash, err := admcli.PLTTransfer(userAddr, amount)
		if err != nil {
			log.Error(err)
			return
		}
		wait(2)
		if err := admcli.DumpEventLog(hash); err != nil {
			log.Error(err)
			return
		}
	}

	// lock plt
	{
		logsplit()
		log.Infof("lock PLT...")
		balanceBeforeLock, err := cli.BalanceOf(userAddr, "latest")
		if err != nil {
			log.Error(err)
			return
		}

		hash, err := cli.Lock(params.ChainID, params.BindTo, amount)
		if err != nil {
			log.Errorf("failed to call `lock` err: %v", err)
			return
		}
		wait(2)
		if err := cli.DumpEventLog(hash); err != nil {
			log.Error("failed to dump `lock` event hash %s, err: %v", hash.Hex(), err)
			return
		}

		balanceAfterLock, err := cli.BalanceOf(userAddr, "latest")
		if err != nil {
			log.Error(err)
			return
		}

		subAmount := utils.SafeSub(balanceBeforeLock, balanceAfterLock)
		if subAmount.Cmp(amount) != 0 {
			log.Errorf("balance before lock %d, after lock %d, the sub amount should be %d",
				plt.PrintUPLT(balanceBeforeLock), plt.PrintUPLT(balanceAfterLock), plt.PrintUPLT(amount))
			return
		}
	}

	// unlock
	//{
	//	logsplit()
	//	log.Infof("unlock PLT...")
	//	balanceBeforeUnLock, err := cli.BalanceOf(userAddr, "latest")
	//	if err != nil {
	//		log.Error(err)
	//		return
	//	}
	//
	//	args := &plt.TxArgs{
	//		ToAssetHash: []byte{},
	//		ToAddress:   params.BindTo.Bytes(),
	//		Amount:      amount,
	//	}
	//	hash, err := cli.UnLock(args, params.BindTo, params.ChainID)
	//	if err != nil {
	//		log.Error(err)
	//		return
	//	}
	//	wait(2)
	//	if err := cli.DumpEventLog(hash); err != nil {
	//		log.Error(err)
	//		return
	//	}
	//
	//	balanceAfterUnLock, err := cli.BalanceOf(userAddr, "latest")
	//	if err != nil {
	//		log.Error(err)
	//		return
	//	}
	//
	//	subAmount := utils.SafeSub(balanceAfterUnLock, balanceBeforeUnLock)
	//	if subAmount.Cmp(amount) != 0 {
	//		log.Errorf("balance before unlock %d, after unlock %d, the sub amount should be %d",
	//			plt.PrintUPLT(balanceBeforeUnLock), plt.PrintUPLT(balanceAfterUnLock), plt.PrintUPLT(amount))
	//		return
	//	}
	//}
	return true
}

// 同步palette区块头到poly链上
// 1. 环境准备，palette cli: 使用任意palette签名者对应的cli, poly cli: 必须是poly验证节点的validators作为多签地址
// 2. 获取palette当前块高的区块头, 并使用json序列化为bytes
// 3. 使用poly cli同步第二步的bytes以及palette network id到poly native管理合约,
//	  这笔交易发出后等待poly当前块高超过交易块高, 作为落账的判断条件
// 4. 获取poly当前块高作为写入palette管理合约的genesis块高，获取对应的block，将block header及block book keeper
//    序列化，提交到palette管理合约
func SyncGenesis() (succeed bool) {

	// 1. prepare
	polyRPC := config.Conf.Poly.RPCAddress
	polyValidators := config.Conf.Poly.LoadPolyAccountList()
	polyCli, err := poly.NewPolyClient(polyRPC, polyValidators)
	if err != nil {
		log.Errorf("failed to generate poly client, err: %s", err)
		return
	} else {
		log.Infof("generate poly client success!")
	}

	// 2. get palette current block header
	logsplit()
	cli := admcli
	curr, hdr, err := cli.GetCurrentBlockHeader()
	if err != nil {
		log.Errorf("failed to get block header, err: %s", err)
		return
	}
	pltHeaderEnc, err := hdr.MarshalJSON()
	if err != nil {
		log.Errorf("marshal header failed, err: %s", err)
		return
	}
	log.Infof("get palette block header with current height %d, header %s", curr, hexutil.Encode(pltHeaderEnc))

	// 3. sync palette header to poly
	{
		logsplit()
		chainID := uint64(config.Conf.Environment.NetworkID)
		txhash, err := polyCli.SyncGenesisBlock(chainID, pltHeaderEnc)
		if err != nil {
			log.Errorf("SyncEthGenesisHeader failed: %v", err)
			return
		}
		if err := polyCli.WaitPolyTx(txhash); err != nil {
			log.Errorf("waiting for poly tx err: %s", err)
			return
		}
		log.Infof("successful to sync eth genesis header: txhash %s", txhash.Hex())
	}

	// 4. get poly block and assemble book keepers to header
	{
		logsplit()

		// `epoch` related with the poly validators changing,
		// we can set it as 0 if poly validators never changed on develop environment.
		epoch, err := polyCli.GetCurrentBlockHeight()
		if err != nil {
			log.Errorf("failed to get poly height, err: %s", err)
			return
		}
		gB, err := polyCli.GetBlockByHeight(epoch)
		if err != nil {
			log.Errorf("failed to get block, err: %s", err)
			return
		}
		bookeepers, err := poly.GetBookeeper(gB)
		if err != nil {
			log.Errorf("failed to get bookeepers, err: %s", err)
			return
		}
		bookeepersEnc := poly.AssembleNoCompressBookeeper(bookeepers)
		headerEnc := gB.Header.ToArray()

		_, eccmAddr, _, err := config.ReadContracts()
		if err != nil {
			log.Errorf("failed to read eccm contract address, err: %s", err)
			return
		}
		txhash, err := cli.InitGenesisBlock(eccmAddr, headerEnc, bookeepersEnc)
		if err != nil {
			log.Errorf("failed to initGenesisBlock, err: %s", err)
			return
		} else {
			log.Infof("sync genesis header success, txhash %s", txhash.Hex())
		}
		_ = cli.DumpEventLog(txhash)
	}

	return true
}

// 获取并打印跨链事件
func GetProof() (succeed bool) {
	return true
}
