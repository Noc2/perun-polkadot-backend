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

package substrate

import "github.com/centrifuge/go-substrate-rpc-client/v3/types"

// Pallet binds to a pallet that is deployed on a substrate chain.
type Pallet struct {
	name string
	api  *Api
	ext  *ExtFactory
}

// NewPallet returns a new pallet.
func NewPallet(api *Api, name string) *Pallet {
	return &Pallet{name, api, NewExtFactory(api)}
}

// Transact sends an Extrinsic and returns a status sub for it.
func (p *Pallet) Transact(ext *types.Extrinsic) (*ExtStatusSub, error) {
	return p.api.Transact(ext)
}

// Subscribe subscribes on all events of the pallet.
func (p *Pallet) Subscribe(pastBlocks types.BlockNumber) (*EventSource, error) {
	// Use the SystemEventsKey here since there is no specific key for a pallet,
	// they need to be filtered later on.
	return NewEventSource(p.api, pastBlocks, SystemEventsKey())
}

func (p *Pallet) BuildQuery(variable string, args ...[]byte) (types.StorageKey, error) {
	return p.api.BuildKey(p.name, variable, args...)
}

// BuildExt builds and signs an extrinsic.
func (p *Pallet) BuildExt(call *ExtName, args []interface{}, addr types.AccountID, signer ExtSigner) (*types.Extrinsic, error) {
	// Build the call.
	ext, err := p.ext.BuildExt(call, args)
	if err != nil {
		return nil, err
	}
	// Get signature options.
	opts, err := p.ext.SigOptions(addr)
	if err != nil {
		return nil, err
	}
	// Sign the ext.
	return ext, signer.SignExt(ext, *opts, p.api.Network())
}
