package core

import (
	"github.com/palettechain/onRobot/config"
	"github.com/palettechain/onRobot/pkg/log"
	"github.com/palettechain/onRobot/pkg/poly"
)

func PolyHeight() (succeed bool) {
	rpc := config.Conf.CrossChain.PolyRPCAddress
	polyValidators := config.Conf.CrossChain.LoadPolyAccountList()
	polyCli, err := poly.NewPolyClient(rpc, polyValidators)
	if err != nil {
		log.Errorf("failed to generate poly client, err: %s", err)
		return
	} else {
		log.Infof("generate poly client success!")
	}

	height, err := polyCli.GetCurrentBlockHeight()
	if err != nil {
		log.Error(err)
		return
	}

	log.Infof("%s current height %d", rpc, height)
	return true
}
//
//func PolyTx() (succeed bool) {
//	var params struct {
//		Tx string
//	}
//
//	if err := config.LoadParams("PolyTx.json", &params); err != nil {
//		log.Error(err)
//		return
//	}
//
//	rpc := config.Conf.CrossChain.PolyRPCAddress
//	polyValidators := config.Conf.CrossChain.LoadPolyAccountList()
//	polyCli, err := poly.NewPolyClient(rpc, polyValidators)
//	if err != nil {
//		log.Errorf("failed to generate poly client, err: %s", err)
//		return
//	} else {
//		log.Infof("generate poly client success!")
//	}
//
//	polyCli.CommitPolyDpos()
//}
