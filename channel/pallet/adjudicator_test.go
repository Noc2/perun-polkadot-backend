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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	pchannel "perun.network/go-perun/channel"

	"github.com/perun-network/perun-polkadot-backend/channel"
	"github.com/perun-network/perun-polkadot-backend/channel/pallet"
	"github.com/perun-network/perun-polkadot-backend/channel/pallet/test"
	chtest "github.com/perun-network/perun-polkadot-backend/channel/test"
)

func TestAdjudicator_Register(t *testing.T) {
	s := test.NewSetup(t)
	adj := pallet.NewAdjudicator(s.Alice.Acc, s.Pallet, s.Api, 50)
	req, _, state := newAdjReq(s, false)

	// Channel is not yet registered
	s.AssertNoRegistered(state.ID)
	// Register the channel twice. Register should be idempotent.
	assert.NoError(t, adj.Register(s.NewCtx(), req, nil))
	assert.NoError(t, adj.Register(s.NewCtx(), req, nil))
	// Check on-chain state for the register.
	_state, err := channel.NewState(state)
	require.NoError(t, err)
	s.AssertRegistered(_state, false)
}

func TestAdjudicator_ConcludeFinal(t *testing.T) {
	s := test.NewSetup(t)
	req, params, state := newAdjReq(s, true)
	fSetup := chtest.NewFundingSetup(params, state)

	// Fund
	err := test.FundAll(s.NewCtx(), s.Funders, fSetup.FReqs)
	assert.NoError(t, err)
	// Withdraw
	{
		// Alice
		adj := pallet.NewAdjudicator(s.Alice.Acc, s.Pallet, s.Api, test.PastBlocks)
		assert.NoError(t, adj.Withdraw(s.NewCtx(), req, nil))
		req.Idx = 1
		req.Acc = s.Bob.Acc
		adj = pallet.NewAdjudicator(s.Bob.Acc, s.Pallet, s.Api, test.PastBlocks)
		assert.NoError(t, adj.Withdraw(s.NewCtx(), req, nil))
	}
}

func TestAdjudicator_Walkthrough(t *testing.T) {
	s := test.NewSetup(t)
	req, params, state := newAdjReq(s, false)
	fSetup := chtest.NewFundingSetup(params, state)
	adjAlice := pallet.NewAdjudicator(s.Alice.Acc, s.Pallet, s.Api, test.PastBlocks)
	adjBob := pallet.NewAdjudicator(s.Bob.Acc, s.Pallet, s.Api, test.PastBlocks)
	ctx, cancel := context.WithTimeout(context.Background(), 100*s.BlockTime)
	defer cancel()

	// Fund
	err := test.FundAll(s.NewCtx(), s.Funders, fSetup.FReqs)
	assert.NoError(t, err)
	// Dispute
	{
		// Register non-final state
		require.NoError(t, adjAlice.Register(ctx, req, nil))

		// Register non-final state with higher version
		next := req.Tx.State.Clone()
		next.Version++                        // increase version to allow progression
		test.MixBals(s.Rng, next.Balances[0]) // mix up the balances
		next.IsFinal = false
		_next, err := channel.NewState(next)
		require.NoError(t, err)
		sigs := s.SignState(_next)
		req.Acc = s.Bob.Acc
		req.Tx = pchannel.Transaction{State: next, Sigs: sigs}
		req.Idx = 1
		require.NoError(t, adjBob.Register(ctx, req, nil))

		// Register final state with higher version
		next = next.Clone()
		next.Version++ // increase version to allow progression
		next.IsFinal = true
		_next, err = channel.NewState(next)
		require.NoError(t, err)
		sigs = s.SignState(_next)
		req.Tx = pchannel.Transaction{State: next, Sigs: sigs}
	}
	// Withdraw
	{
		// Bob
		require.NoError(t, adjBob.Withdraw(ctx, req, nil))

		// Alice
		req.Idx = 0
		req.Acc = s.Alice.Acc
		require.NoError(t, adjAlice.Withdraw(ctx, req, nil))
	}
}
