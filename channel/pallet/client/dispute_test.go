// Copyright 2019 - See NOTICE file for copyright holders.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client_test

import (
	"context"
	"math/big"
	"testing"

	pclient "perun.network/go-perun/client"
	clienttest "perun.network/go-perun/client/test"
	"perun.network/go-perun/log"
	"perun.network/go-perun/wire"

	"github.com/perun-network/perun-polkadot-backend/channel"
	"github.com/perun-network/perun-polkadot-backend/channel/pallet/test"
	"github.com/perun-network/perun-polkadot-backend/wallet"
	"github.com/stretchr/testify/assert"
)

const TestTimeoutBlocks = 100

func TestDisputeMalloryCarol(t *testing.T) {
	s := test.NewSetup(t)
	ctx, cancel := context.WithTimeout(context.Background(), TestTimeoutBlocks*s.BlockTime)
	defer cancel()

	const A, B = 0, 1 // Indices of Mallory and Carol
	var (
		name = [2]string{"Mallory", "Carol"}
		role [2]clienttest.Executer
	)

	setup := makeRoleSetups(s, name)
	role[A] = clienttest.NewMallory(setup[A], t)
	role[B] = clienttest.NewCarol(setup[B], t)

	execConfig := &clienttest.MalloryCarolExecConfig{
		BaseExecConfig: clienttest.MakeBaseExecConfig(
			[2]wire.Address{setup[A].Identity.Address(), setup[B].Identity.Address()},
			channel.NewAsset(),
			[2]*big.Int{big.NewInt(100000000000000), big.NewInt(100000000000000)},
			pclient.WithoutApp(),
		),
		NumPayments: [2]int{5, 0},
		TxAmounts:   [2]*big.Int{big.NewInt(2000000000000), big.NewInt(0)},
	}

	// Compensate the fees of the extrinsics.
	epsilon := new(big.Int).SetUint64(test.DefaultExtFee * 6)
	aliceToBob := big.NewInt(int64(execConfig.NumPayments[A])*execConfig.TxAmounts[A].Int64() - int64(execConfig.NumPayments[B])*execConfig.TxAmounts[B].Int64())

	// aliceToBob is transferred from alice to bob.
	s.AssertBalanceChange(wallet.AsAddr(s.Alice.Acc.Address()).AccountId(), aliceToBob, epsilon,
		func() {
			bobToAlice := new(big.Int).Neg(aliceToBob)
			// bobToAlice is transferred from bob to alice.
			s.AssertBalanceChange(wallet.AsAddr(s.Bob.Acc.Address()).AccountId(), bobToAlice, epsilon,
				func() {
					err := clienttest.ExecuteTwoPartyTest(ctx, role, execConfig)
					assert.NoError(t, err)
				})
		})

	log.Info("Dispute test done")
}
