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

package sr25519_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	pkgtest "perun.network/go-perun/pkg/test"

	"github.com/perun-network/perun-polkadot-backend/pkg/substrate"
	wallet "github.com/perun-network/perun-polkadot-backend/wallet/sr25519"
	"github.com/perun-network/perun-polkadot-backend/wallet/sr25519/test"
)

// TestAddress_SS58Addr tests that the SS58 address calculation returns
// fixed values.
func TestAddress_SS58Addr(t *testing.T) {
	for _, setup := range test.LoadDevAccounts(t) {
		// Check AccountID
		id := wallet.AsAddr(setup.Acc.Address()).AccountID()
		assert.Equal(t, []byte(setup.Id), id[:])

		// Check SS58 addresses
		for _, addr := range setup.Addr {
			gotAddr, err := substrate.SS58Address(id, addr.Network)
			require.NoError(t, err)
			assert.Equal(t, addr.Value, gotAddr)
		}
	}
}

// TestAddress_ZeroCmp tests that the zero address is strictly
// smaller than all other addresses.
func TestAddress_ZeroCmp(t *testing.T) {
	rng := pkgtest.Prng(t)
	r := test.NewRandomizer()
	addr := r.NewRandomAddress(rng)
	zero := test.NewAddressZero()

	for i := 0; i < 1000; i++ {
		require.False(t, addr.Cmp(zero) < 0)
		require.False(t, zero.Cmp(addr) > 0)
	}
}
