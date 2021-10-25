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

package channel_test

import (
	"math/big"
	"math/rand"
	"testing"

	ptest "perun.network/go-perun/channel/test"
	pkgtest "perun.network/go-perun/pkg/test"
	pwallet "perun.network/go-perun/wallet"
	pwallettest "perun.network/go-perun/wallet/test"

	_ "github.com/perun-network/perun-polkadot-backend/channel/test"        // init
	_ "github.com/perun-network/perun-polkadot-backend/wallet/sr25519/test" // init
)

func TestBackend_GenericBackend(t *testing.T) {
	setup := newSetup(pkgtest.Prng(t))
	ptest.GenericBackendTest(t, setup, ptest.IgnoreAssets)
}

func newSetup(rng *rand.Rand) *ptest.Setup {
	opts := ptest.WithNumLocked(0).Append(
		ptest.WithBalancesInRange(big.NewInt(0), big.NewInt(1<<60)),
		ptest.WithNumAssets(1),
		ptest.WithoutApp())
	params, state := ptest.NewRandomParamsAndState(rng, opts)

	opts2 := opts.Append(
		ptest.WithIsFinal(!state.IsFinal))
	params2, state2 := ptest.NewRandomParamsAndState(rng, opts2)

	createAddr := func() pwallet.Address {
		return pwallettest.NewRandomAddress(rng)
	}

	return &ptest.Setup{
		Params:        params,
		Params2:       params2,
		State:         state,
		State2:        state2,
		Account:       pwallettest.NewRandomAccount(rng),
		RandomAddress: createAddr,
	}
}
