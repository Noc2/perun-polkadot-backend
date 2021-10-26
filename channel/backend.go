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

package channel

import (
	"fmt"
	"io"
	"log"

	eth "github.com/ethereum/go-ethereum/crypto"
	pchannel "perun.network/go-perun/channel"
	pwallet "perun.network/go-perun/wallet"
)

// backend implements the backend interface.
// The type is private since it only needs to be exposed as singleton by the
// `Backend` variable.
type backend struct{}

// Backend is the channel backend. Is a singleton since there is only one backend.
var Backend backend

// CalcID calculates the channelID.
func (*backend) CalcID(params *pchannel.Params) (id pchannel.ID) {
	return CalcID(params)
}

// Sign signs a state with the passed account.
func (*backend) Sign(acc pwallet.Account, state *pchannel.State) (pwallet.Sig, error) {
	_state, err := NewState(state)
	if err != nil {
		return nil, err
	}
	data, err := ScaleEncode(_state)
	if err != nil {
		return nil, err
	}
	return acc.SignData(data)
}

// Verify verifies a signature on a state.
func (*backend) Verify(addr pwallet.Address, state *pchannel.State, sig pwallet.Sig) (bool, error) {
	_state, err := NewState(state)
	if err != nil {
		return false, err
	}
	data, err := ScaleEncode(_state)
	if err != nil {
		return false, err
	}
	return pwallet.VerifySignature(data, sig, addr)
}

// DecodeAsset returns the unique asset of the backend. Cannot error.
func (*backend) DecodeAsset(r io.Reader) (pchannel.Asset, error) {
	return Asset, nil
}

// CalcID calculates the channelID by encoding and hashing the params.
func CalcID(params *pchannel.Params) (id pchannel.ID) {
	_params, err := NewParams(params)
	if err != nil {
		panic(fmt.Sprintf("cannot calculate channel ID: %v", err))
	}
	bytes, err := ScaleEncode(_params)
	if err != nil {
		log.Panicf("could not encode parameters: %v", err)
	}
	return eth.Keccak256Hash(bytes)
}
