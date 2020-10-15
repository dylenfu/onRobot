package internal

import "github.com/palettechain/onRobot/pkg/log"

func Demo() bool {
	//// get block height
	//jsonrpcAddr := "http://172.168.3.158:20336"
	//height, err := sdk.GetBlockCurrentHeight(jsonrpcAddr)
	//if err != nil {
	//	log.Error(err)
	//	return false
	//}
	//log.Infof("jsonrpcAddr %s current block height %d", jsonrpcAddr, height)
	//
	//// recover kp
	//acc, err := sdk.RecoverAccount(conf.TransferWalletPath, conf.DefConfig.WalletPwd)
	//if err != nil {
	//	log.Error(err)
	//	return false
	//}
	//log.Infof("address %s", acc.Address.ToBase58())
	//
	//// get balance
	//resp, err := sdk.GetBalance(jsonrpcAddr, acc.Address)
	//if err != nil {
	//	log.Error(err)
	//	return false
	//}
	log.Infof("ont %s, ong %s, block height %s", "1", "1", "2")
	return true
}

func reset() {
}
