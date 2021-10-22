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
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	pchannel "perun.network/go-perun/channel"
	pchtest "perun.network/go-perun/channel/test"
	pkgtest "perun.network/go-perun/pkg/test"
	"perun.network/go-perun/wallet"

	"github.com/perun-network/perun-polkadot-backend/channel"
	"github.com/perun-network/perun-polkadot-backend/channel/pallet"
	subtest "github.com/perun-network/perun-polkadot-backend/pkg/substrate/test"
	wallettest "github.com/perun-network/perun-polkadot-backend/wallet/sr25519/test"
)

type (
	// Setup is the test setup for the channel package tests.
	Setup struct {
		T   *testing.T
		Rng *rand.Rand

		Accs       []*wallettest.DevAccount
		Alice, Bob *wallettest.DevAccount

		*subtest.Setup
	}

	// FundingSetup is the setup for funder tests.
	FundingSetup struct {
		*Setup

		FReqs     []*pchannel.FundingReq
		Fids      []channel.FundingId
		FinalBals []pchannel.Bal
	}

	// DepositSetup is the setup for depositor tests.
	DepositSetup struct {
		*FundingSetup

		DReqs []*pallet.DepositReq
	}
)

// DefaultTestTimeout default timeout for a test in block-time.
var DefaultTestTimeout = 10

// NewSetup returns a new setup and assumes that the sr25519 wallet is used.
func NewSetup(t *testing.T) *Setup {
	s := subtest.NewSetup(t)
	accs := wallettest.LoadDevAccounts(t)

	return &Setup{t, pkgtest.Prng(t), accs, accs[0], accs[1], s}
}

// SignState returns the signatures for Alice and Bob on the state.
func (s *Setup) SignState(state *channel.State) []wallet.Sig {
	data, err := channel.ScaleEncode(state)
	require.NoError(s.T, err)

	sig1, err := s.Alice.Acc.SignData(data)
	require.NoError(s.T, err)
	sig2, err := s.Bob.Acc.SignData(data)
	require.NoError(s.T, err)

	return []wallet.Sig{sig1, sig2}
}

// NewRandomParamAndState generates compatible Params and State.
func (s *Setup) NewRandomParamAndState() (*pchannel.Params, *pchannel.State) {
	params, state := pchtest.NewRandomParamsAndState(s.Rng, DefaultRandomOpts())
	state.Allocation.Balances = state.Balances
	return params, state
}

// NewCtx returns a new context that will timeout after DefaultTestTimeout
// blocks and cancel on test cleanup.
func (s *Setup) NewCtx() context.Context {
	timeout := s.BlockTime * time.Duration(DefaultTestTimeout)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	s.T.Cleanup(cancel)
	return ctx
}

// NewFundingSetup returns a new FundingSetup.
func NewFundingSetup(params *pchannel.Params, state *pchannel.State) *FundingSetup {
	reqAlice := pchannel.NewFundingReq(params, state, 0, state.Balances)
	reqBob := pchannel.NewFundingReq(params, state, 1, state.Balances)
	fReqAlice, _ := channel.MakeFundingReq(reqAlice)
	fReqBob, _ := channel.MakeFundingReq(reqBob)
	fidAlice, _ := fReqAlice.ID()
	fidBob, _ := fReqBob.ID()
	balAlice := state.Balances[0][reqAlice.Idx]
	balBob := state.Balances[0][reqBob.Idx]

	return &FundingSetup{
		FReqs:     []*pchannel.FundingReq{reqAlice, reqBob},
		Fids:      []channel.FundingId{fidAlice, fidBob},
		FinalBals: []pchannel.Bal{balAlice, balBob},
	}
}

// NewDepositSetup returns a new DepositSetup.
func NewDepositSetup(s *FundingSetup, accs ...wallet.Account) *DepositSetup {
	reqs := make([]*pallet.DepositReq, len(accs))
	for i := range accs {
		reqs[i], _ = pallet.NewDepositReqFromPerun(s.FReqs[i], accs[i])
	}
	return &DepositSetup{DReqs: reqs}
}
