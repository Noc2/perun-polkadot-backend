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
	"crypto/rand"
	"io"
	"math/big"
	mrand "math/rand"

	pchannel "perun.network/go-perun/channel"
)

// MixBals randomly modifies the values of bals while the sum stays the same.
// All values should be > 0.
func MixBals(rng io.Reader, bals []pchannel.Bal) {
	// Transfer a random amount between two entries `len(bals)` times.
	for i := 0; i < len(bals); i++ {
		from, to := mrand.Intn(len(bals)), mrand.Intn(len(bals))
		diff, err := rand.Int(rng, bals[from])
		if err != nil {
			panic(err)
		}
		bals[from].Sub(bals[from], diff)
		bals[to].Add(bals[to], diff)
	}
}

// Multiply multiplies the passed `bals` with `f` and returns them.
func Multiply(f int64, bals ...pchannel.Bal) []pchannel.Bal {
	res := make([]pchannel.Bal, len(bals))
	factor := big.NewInt(f)
	for i := range bals {
		res[i] = new(big.Int).Mul(factor, bals[i])
	}
	return res
}
