// Copyright 2021 PolyCrypt GmbH
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

package pallet_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	pchannel "perun.network/go-perun/channel"
	pchtest "perun.network/go-perun/channel/test"
	ctxtest "perun.network/go-perun/pkg/context/test"
	pwallet "perun.network/go-perun/wallet"

	"github.com/perun-network/perun-polkadot-backend/channel"
	"github.com/perun-network/perun-polkadot-backend/channel/pallet"
	"github.com/perun-network/perun-polkadot-backend/channel/pallet/test"
	chtest "github.com/perun-network/perun-polkadot-backend/channel/test"
)

func TestAdjudicatorSub_Register(t *testing.T) {
	s := test.NewSetup(t)
	adj := pallet.NewAdjudicator(s.Alice.Acc, s.Pallet, s.Api, test.PastBlocks)
	req, params, _ := newAdjReq(s, false)

	sub, err := adj.Subscribe(s.NewCtx(), params.ID())
	require.NoError(t, err)
	// Register the channel twice.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	assert.NoError(t, adj.Register(ctx, req, nil))
	assert.NoError(t, adj.Register(ctx, req, nil))
	// Wait for one Registered event
	event := sub.Next().(*pchannel.RegisteredEvent)
	assert.Equal(t, params.ID(), event.IDV)
	// No second event should be emitted
	ctxtest.AssertNotTerminates(t, 2*s.BlockTime, func() { sub.Next() })
	// Sub returns nil after close
	assert.NoError(t, sub.Close())
	var noEvent channel.PerunEvent
	ctxtest.AssertTerminatesQuickly(t, func() { noEvent = sub.Next() })
	assert.Nil(t, noEvent)
	assert.Nil(t, sub.Err())
}

func TestAdjudicatorSub_ConcludeFinal(t *testing.T) {
	s := test.NewSetup(t)
	adj := pallet.NewAdjudicator(s.Alice.Acc, s.Pallet, s.Api, test.PastBlocks)
	req, params, state := newAdjReq(s, true)

	// Deposit funds for Alice and bob.
	depositBoth(s, params, state)

	sub, err := adj.Subscribe(s.NewCtx(), params.ID())
	require.NoError(t, err)
	// Register the channel twice.
	assert.NoError(t, adj.Withdraw(s.NewCtx(), req, nil))
	assert.NoError(t, adj.Withdraw(s.NewCtx(), req, nil))
	// Wait for one Concluded event
	event := sub.Next().(*pchannel.ConcludedEvent)
	assert.Equal(t, params.ID(), event.IDV)
	// No second event should be emitted
	ctxtest.AssertNotTerminates(t, 2*s.BlockTime, func() { sub.Next() })
	// Sub returns nil after close
	assert.NoError(t, sub.Close())
	var noEvent channel.PerunEvent
	ctxtest.AssertTerminatesQuickly(t, func() { noEvent = sub.Next() })
	assert.Nil(t, noEvent)
	assert.Nil(t, sub.Err())
}

func depositBoth(s *test.Setup, params *pchannel.Params, state *pchannel.State) {
	depositor := pallet.NewDepositor(s.Pallet)
	for i := 0; i < 2; i++ {
		myBal := state.Balances[0][i]
		fReq := pchannel.NewFundingReq(params, state, pchannel.Index(i), state.Balances)
		_fReq, err := channel.MakeFundingReqFromPerun(fReq)
		require.NoError(s.T, err)
		fid, _ := _fReq.ID()
		req := pallet.NewDepositReq(myBal, s.Alice.Acc, fid)
		assert.NoError(s.T, depositor.Deposit(s.NewCtx(), req))
	}
}

func newAdjReq(s *test.Setup, final bool) (pchannel.AdjudicatorReq, *pchannel.Params, *pchannel.State) {
	state := pchtest.NewRandomState(s.Rng, chtest.DefaultRandomOpts())
	state.IsFinal = final
	var data [20]byte
	s.Rng.Read(data[:])
	nonce := pchannel.NonceFromBytes(data[:])
	params, err := pchannel.NewParams(60, []pwallet.Address{s.Alice.Acc.Address(), s.Bob.Acc.Address()}, pchannel.NoApp(), nonce, true, false)
	require.NoError(s.T, err)
	state.ID = params.ID()
	wState, err := channel.NewState(state)
	require.NoError(s.T, err)
	sigs := s.SignState(wState)
	req := pchannel.AdjudicatorReq{
		Params:    params,
		Acc:       s.Alice.Acc,
		Tx:        pchannel.Transaction{State: state, Sigs: sigs},
		Idx:       0,
		Secondary: false,
	}
	return req, params, state
}
