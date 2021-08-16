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
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/palettechain/onRobot/config"
	"github.com/palettechain/onRobot/pkg/log"
	"github.com/palettechain/onRobot/pkg/sdk"
)

// 节点代理用户质押:
// 1. admin给用户充值
// 2. fans delegate
// 3. 质押后不等待生效，立即取消质押
// 4. 充值返还给admin
func Stable() (succeed bool) {
	const num = 50000

	type Fan struct {
		Address common.Address
		Cli     *sdk.Client
	}

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
	{
		logsplit()
		nodeIndex := 5
		node := config.Conf.Nodes[nodeIndex]
		validator := node.NodeAddr()
		stakeAccount := node.StakeAddr()
		log.Infof("fans delegate......")
		for _, fan := range fans {
			if _, err := fan.Cli.StakeWithoutWaiting(validator, stakeAccount, amount, false); err != nil {
				log.Errorf("%s stake to node%d failed, err: %v", fan.Address, node.Index, err)
			}
		}
	}

	wait(10)
	return
}
