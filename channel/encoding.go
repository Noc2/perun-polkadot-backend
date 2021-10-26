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
	"math/big"
	"time"

	"github.com/centrifuge/go-substrate-rpc-client/v3/types"
	"github.com/perun-network/perun-polkadot-backend/pkg/substrate"
	"github.com/pkg/errors"
	pchannel "perun.network/go-perun/channel"
	pwallet "perun.network/go-perun/wallet"
)

// ScaleEncode encodes any struct according to the SCALE codec.
func ScaleEncode(obj interface{}) ([]byte, error) {
	return types.EncodeToBytes(obj)
}

// ScaleDecode decodes any struct according to the SCALE codec.
func ScaleDecode(obj interface{}, data []byte) error {
	return types.DecodeFromBytes(data, obj)
}

// MakeBalance creates a new Balance.
func MakeBalance(bal *big.Int) (Balance, error) {
	if bal.Sign() < 0 || bal.Cmp(MaxBalance.Int) > 0 {
		return types.NewU128(*big.NewInt(0)), errors.New("invalid balance")
	}
	return types.NewU128(*bal), nil
}

// MakePerunBalance creates a Perun balance.
func MakePerunBalance(bal Balance) *big.Int {
	return new(big.Int).Set(bal.Int)
}

// MakePerunAlloc creates a new Perun allocation and fills the Locked balances
// with a default value. It uses the fixed backend asset.
func MakePerunAlloc(bals []Balance) pchannel.Allocation {
	return pchannel.Allocation{
		Assets:   []pchannel.Asset{Asset},
		Balances: MakePerunBalances(bals),
		Locked:   nil,
	}
}

// MakePerunBalances creates Perun balances with the fixed backend asset.
func MakePerunBalances(bals []Balance) pchannel.Balances {
	ret := make(pchannel.Balances, 1)
	ret[0] = make([]pchannel.Bal, len(bals))
	for i, bal := range bals {
		ret[0][i] = new(big.Int).Set(bal.Int)
	}
	return ret
}

// MakeChallengeDuration creates a new ChallengeDuration from the argument.
func MakeChallengeDuration(challengeDuration uint64) ChallengeDuration {
	return challengeDuration
}

// MakePerunChallengeDuration creates a new Perun ChallengeDuration.
func MakePerunChallengeDuration(sec ChallengeDuration) uint64 {
	return sec
}

// MakeDuration creates a new duration from the argument.
func MakeDuration(sec ChallengeDuration) time.Duration {
	return time.Second * time.Duration(sec)
}

// MakeTime creates a new time from the argument.
func MakeTime(sec ChallengeDuration) time.Time {
	return time.Unix(int64(sec), 0)
}

// MakeTimeout creates a new timeout.
func MakeTimeout(sec ChallengeDuration, storage substrate.StorageQueryer) pchannel.Timeout {
	return substrate.NewTimeout(storage, MakeTime(sec), substrate.DefaultTimeoutPollInterval)
}

// MakeNonce creates a new Nonce or an error if the argument was out of range.
func MakeNonce(nonce *big.Int) (Nonce, error) {
	var ret Nonce

	if nonce.Sign() < 0 { // negative?
		return ret, ErrNonceOutOfRange
	}
	if nonce.BitLen() > (8*NonceLen) || nonce.BitLen() == 0 { // too long/short?
		return ret, ErrNonceOutOfRange
	}
	copy(ret[:], nonce.Bytes())

	return ret, nil
}

// NewState creates a new State. It discards app data because app-channels are
// currently not supported.
func NewState(s *pchannel.State) (*State, error) {
	if err := s.Valid(); err != nil {
		return nil, ErrStateIncompatible
	}
	bals, err := MakeAlloc(&s.Allocation)

	return &State{
		Channel:  s.ID,
		Version:  s.Version,
		Balances: bals,
		Final:    s.IsFinal,
	}, err
}

// NewPerunState creates a new Perun state and fills the App-Data with a
// default value.
func NewPerunState(s *State) *pchannel.State {
	return &pchannel.State{
		ID:         s.Channel,
		Version:    s.Version,
		Allocation: MakePerunAlloc(s.Balances),
		Data:       pchannel.NoData(),
		IsFinal:    s.Final,
	}
}

// MakeAlloc converts an Allocation to Balances. Currently, it supports only a
// single asset.
func MakeAlloc(a *pchannel.Allocation) ([]Balance, error) {
	var err error
	ret := make([]Balance, len(a.Balances[0]))

	if len(a.Assets) != 1 || len(a.Balances) != 1 || len(a.Locked) != 0 {
		return ret, ErrAllocIncompatible
	}
	for i, bal := range a.Balances[0] {
		if ret[i], err = MakeBalance(bal); err != nil {
			break
		}
	}

	return ret, err
}

// NewWithdrawal creates a new Withdrawal.
func NewWithdrawal(cid pchannel.ID, part, receiver pwallet.Address) (*Withdrawal, error) {
	off, err := MakeOffIdent(part)
	if err != nil {
		return nil, err
	}
	on, err := MakeOnIdent(receiver)

	return &Withdrawal{
		cid, off, on,
	}, err
}

// NewParams creates backend-specific parameters from generic Perun parameters.
func NewParams(p *pchannel.Params) (*Params, error) {
	nonce, err := MakeNonce(p.Nonce)
	if err != nil {
		return nil, err
	}
	parts, err := MakeOffIdents(p.Parts)

	return &Params{
		Nonce:             nonce,
		Participants:      parts,
		ChallengeDuration: MakeChallengeDuration(p.ChallengeDuration),
	}, err
}

// MakeSig creates a new Sig.
func MakeSig(sig pwallet.Sig) (Sig, error) {
	var ret Sig

	if len(sig) != SigLen {
		return ret, nil
	}
	copy(ret[:], sig)

	return ret, nil
}

// MakeSigs creates Sigs.
func MakeSigs(sigs []pwallet.Sig) ([]Sig, error) {
	var err error
	ret := make([]Sig, len(sigs))

	for i, sig := range sigs {
		if ret[i], err = MakeSig(sig); err != nil {
			break
		}
	}

	return ret, err
}

// MakeFundingReq creates a new Funding.
func MakeFundingReq(req *pchannel.FundingReq) (Funding, error) {
	ident, err := MakeOffIdent(req.Params.Parts[req.Idx])

	return Funding{
		req.State.ID,
		ident,
	}, err
}

// MakeOnIdent creates a new OnIdentity.
func MakeOnIdent(addr pwallet.Address) (OnIdentity, error) {
	var ret OnIdentity
	data := addr.Bytes()

	if len(data) != OnIdentityLen {
		return ret, ErrIdentLenMismatch
	}
	copy(ret[:], data)

	return ret, nil
}

// MakeOffIdent creates a new OffIdentity.
func MakeOffIdent(part pwallet.Address) (OffIdentity, error) {
	var ret OffIdentity
	data := part.Bytes()

	if len(data) != OffIdentityLen {
		return ret, ErrIdentLenMismatch
	}
	copy(ret[:], data)

	return ret, nil
}

// MakeOffIdents creates a new []OffIdentity.
func MakeOffIdents(parts []pwallet.Address) ([]OffIdentity, error) {
	var err error
	ret := make([]OffIdentity, len(parts))

	for i, part := range parts {
		if ret[i], err = MakeOffIdent(part); err != nil {
			break
		}
	}

	return ret, err
}
