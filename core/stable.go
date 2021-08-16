/*
 * Copyright (C) 2021 The Zion Authors
 * This file is part of The Zion library.
 *
 * The Zion is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The Zion is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The Zion.  If not, see <http://www.gnu.org/licenses/>.
 */

package core

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/contracts/native/plt"
	"github.com/ethereum/go-ethereum/contracts/native/utils"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/palettechain/onRobot/config"
	"github.com/palettechain/onRobot/pkg/log"
	"github.com/palettechain/onRobot/pkg/sdk"
)

func Stable() (succeed bool) {
	var params struct {
		Number int
		NodeIndex int
	}

	if err := config.LoadParams("Stable.json", &params); err != nil {
		log.Error(err)
		return
	}
	type Fan struct {
		Address common.Address
		Cli     *sdk.Client
	}

	num := params.Number
	fans := make([]*Fan, num)
	url := config.Conf.Nodes[0].RPCAddr()
	for i := 0; i < num; i++ {
		key, err := crypto.GenerateKey()
		if err != nil {
			continue
		}
		cli := sdk.NewSender(url, key)
		addr := crypto.PubkeyToAddress(key.PublicKey)
		fans[i] = &Fan{Address: addr, Cli: cli}
	}

	amount := plt.MultiPLT(10)
	admcli := getPaletteCli(pltCTypeAdmin)
	for i := 0; i < num; i++ {
		if _, err := admcli.PLTTransferWithoutWaiting(fans[i].Address, amount); err != nil {
			log.Error("admin transfer to fans %s failed, err: %v", fans[i].Address.Hex(), err)
		}
	}
	wait(10)

	// fans delegate
	nodeIndex := params.NodeIndex
	node := config.Conf.Nodes[nodeIndex]
	validator := node.NodeAddr()
	stakeAccount := node.StakeAddr()
	{
		logsplit()
		log.Infof("fans delegate......")
		for _, fan := range fans {
			if _, err := fan.Cli.StakeWithoutWaiting(validator, stakeAccount, amount, false); err != nil {
				log.Errorf("%s stake to node%d failed, err: %v", fan.Address, node.Index, err)
			}
		}
	}

	wait(20)

	for _, fan := range fans {
		amount, err := fan.Cli.Withdrawable(fan.Address, "latest")
		if err != nil {
			continue
		} else {
			log.Infof("fans %s withdrawable %.5f", fan.Address.Hex(), plt.PrintFPLT(utils.DecimalFromBigInt(amount)))
		}
	}

	wait(20)

	for _, fan := range fans {
		fan.Cli.WithdrawForWithoutWaiting(fan.Address)
	}

	wait(30)

	//for _, fan := range fans {
	//	if _, err := fan.Cli.StakeWithoutWaiting(validator, stakeAccount, amount, true); err != nil {
	//		log.Errorf("%s stake to node%d failed, err: %v", fan.Address, node.Index, err)
	//	}
	//}

	for _, fan := range fans {
		fan.Cli.WithdrawForWithoutWaiting(fan.Address)
	}

	wait(10)
	return true
}
