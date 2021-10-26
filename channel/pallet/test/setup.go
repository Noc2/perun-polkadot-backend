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

package test

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/centrifuge/go-substrate-rpc-client/v3/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/perun-network/perun-polkadot-backend/channel"
	"github.com/perun-network/perun-polkadot-backend/channel/pallet"
	chtest "github.com/perun-network/perun-polkadot-backend/channel/test"
	"github.com/perun-network/perun-polkadot-backend/pkg/substrate"
)

// Setup is the test setup.
type Setup struct {
	*chtest.Setup

	Pallet *pallet.Pallet

	Deps    []*pallet.Depositor
	Funders []*pallet.Funder
	Adjs    []*pallet.Adjudicator
}

const (
	// DefaultExtFee is a rough cost estimation for one extrinsic.
	// The value is oriented at the base-fee of a default substrate
	// node: https://crates.parity.io/frame_support/weights/constants/struct.ExtrinsicBaseWeight.html
	// The exact value can be configured in the substrate node itself.
	// More info on weight calculation: https://docs.substrate.io/v3/runtime/weights-and-fees/
	DefaultExtFee = uint64(125100000)
	// PastBlocks defines how many past blocks should be queried.
	// Must be large enough to ensure that event subs can query all past events
	// of the current test.
	PastBlocks = 100
)

// NewSetup returns a new Setup.
func NewSetup(t *testing.T) *Setup {
	s := chtest.NewSetup(t)
	p := pallet.NewPallet(pallet.NewPerunPallet(s.API), s.API.Metadata())
	ret := &Setup{Setup: s, Pallet: p}

	for i := 0; i < len(s.Accs); i++ {
		dep := pallet.NewDepositor(p)
		ret.Deps = append(ret.Deps, dep)
		ret.Funders = append(ret.Funders, pallet.NewFunder(p, s.Accs[i].Acc, PastBlocks))
		ret.Adjs = append(ret.Adjs, pallet.NewAdjudicator(s.Accs[i].Acc, p, s.API, PastBlocks))
	}

	return ret
}

// AssertNoRegistered checks that the channel is not registered.
func (s *Setup) AssertNoRegistered(cid channel.ChannelID) {
	_, err := s.Pallet.QueryStateRegister(cid, s.API, PastBlocks)
	assert.ErrorIs(s.T, err, pallet.ErrNoRegisteredState)
}

// AssertRegistered checks that the channel is registered with the passed state
// and concluded value.
func (s *Setup) AssertRegistered(state *channel.State, concluded bool) {
	reg, err := s.Pallet.QueryStateRegister(state.Channel, s.API, PastBlocks)
	require.NoError(s.T, err)
	assert.Equal(s.T, concluded, reg.Concluded)
	assert.Equal(s.T, *state, reg.State)
}

// AssertNoDeposit checks that the funding ID holds no amount.
func (s *Setup) AssertNoDeposit(fid channel.FundingID) {
	_, err := s.Pallet.QueryDeposit(fid, s.API, PastBlocks)
	require.ErrorIs(s.T, err, pallet.ErrNoDeposit)
}

// AssertDeposit checks that the funding ID holds the specified amount.
func (s *Setup) AssertDeposit(fid channel.FundingID, bal *big.Int) {
	dep, err := s.Pallet.QueryDeposit(fid, s.API, PastBlocks)
	require.NoError(s.T, err)
	assert.Equal(s.T, dep.Int, bal)
}

// AssertDeposit checks that the funding ids holds the specified amounts.
func (s *Setup) AssertDeposits(fids []channel.FundingID, bals []*big.Int) {
	for i, fid := range fids {
		s.AssertDeposit(fid, bals[i])
	}
}

// AssertBalanceChange checks that an on-chain account gains delta with
// absolute error of epsilon by executing f.
func (s *Setup) AssertBalanceChanges(deltas map[types.AccountID]*big.Int, epsilon *big.Int, f func()) {
	// Get the old balances.
	before := make(map[types.AccountID]*big.Int)
	for addr := range deltas {
		accInfo, err := s.API.AccountInfo(addr)
		require.NoError(s.T, err)
		before[addr] = accInfo.Free.Int
	}
	// Change the balances.
	f()
	// Get the new balances.
	after := make(map[types.AccountID]*big.Int)
	for addr := range deltas {
		accInfo, err := s.API.AccountInfo(addr)
		require.NoError(s.T, err)
		after[addr] = accInfo.Free.Int
	}
	// Check the change.
	for addr, delta := range deltas {
		gotDelta := new(big.Int).Sub(before[addr], after[addr])
		gotEpsilon := new(big.Int).Sub(delta, gotDelta)
		msg := fmt.Sprintf("Addr: 0x%x, gotDelta: %v, wantDelta: %v, gotEps: %v, wantEps: %v", addr, substrate.NewDotFromPlank(gotDelta), substrate.NewDotFromPlank(delta), substrate.NewDotFromPlank(gotEpsilon), substrate.NewDotFromPlank(epsilon))
		require.True(s.T, gotEpsilon.CmpAbs(epsilon) <= 0, msg)
	}
}
