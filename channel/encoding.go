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

// MakeBalance returns a new Balance or an error.
func MakeBalance(bal *big.Int) (Balance, error) {
	if bal.Sign() < 0 || bal.Cmp(MaxBalance.Int) > 0 {
		return Balance(types.NewU128(*big.NewInt(0))), errors.New("invalid balance")
	}
	return Balance(types.NewU128(*bal)), nil
}

// MakePerunBalance returns a Perun balance.
func MakePerunBalance(bal Balance) *big.Int {
	return new(big.Int).Set(bal.Int)
}

// MakePerunAlloc returns a new Perun allocation and fills unknown fields with
// default values.
func MakePerunAlloc(bals []Balance) pchannel.Allocation {
	return pchannel.Allocation{
		Assets:   []pchannel.Asset{NewAsset()},
		Balances: MakePerunBalances(bals),
		Locked:   nil,
	}
}

// MakePerunBalances returns Perun balances by using only one asset.
func MakePerunBalances(bals []Balance) pchannel.Balances {
	ret := make(pchannel.Balances, 1)
	ret[0] = make([]pchannel.Bal, len(bals))
	for i, bal := range bals {
		ret[0][i] = new(big.Int).Set(bal.Int)
	}
	return ret

}

// MakeChallengeDuration returns a new ChallengeDuration from the argument.
func MakeChallengeDuration(challengeDuration uint64) ChallengeDuration {
	return ChallengeDuration(challengeDuration)
}

// MakePerunChallengeDuration returns a new Perun ChallengeDuration.
func MakePerunChallengeDuration(sec ChallengeDuration) uint64 {
	return uint64(sec)
}

// MakeDuration returns a new duration from the argument.
func MakeDuration(sec ChallengeDuration) time.Duration {
	return time.Second * time.Duration(sec)
}

// MakeTime returns a new time from the argument.
func MakeTime(sec ChallengeDuration) time.Time {
	return time.Unix(int64(sec), 0)
}

// MakeTimeout returns a new timeout.
func MakeTimeout(sec ChallengeDuration, storage substrate.StorageQueryer) pchannel.Timeout {
	return substrate.NewTimeout(storage, MakeTime(sec))
}

// MakeNonce returns a new Nonce or an error iff the argument was out of range.
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

// NewState returns a new State and discards unnecessary fields.
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

// NewPerunState returns a new Perun state and fills unknown fields with
// default values.
func NewPerunState(s *State) *pchannel.State {
	return &pchannel.State{
		ID:         s.Channel,
		Version:    s.Version,
		Allocation: MakePerunAlloc(s.Balances),
		Data:       pchannel.NoData(),
		IsFinal:    s.Final,
	}
}

// MakeAlloc converts an Allocation to Balances and only supports one asset.
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

// NewWithdrawal returns a new Withdrawal or an error.
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

// NewParams returns new Params or an error.
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

// MakeSig returns a new Sig or an error.
func MakeSig(sig pwallet.Sig) (Sig, error) {
	var ret Sig

	if len(sig) != SigLen {
		return ret, nil
	}
	copy(ret[:], sig)

	return ret, nil
}

// MakeSigsFromPerun returns Sigs or an error.
func MakeSigsFromPerun(sigs []pwallet.Sig) ([]Sig, error) {
	var err error
	ret := make([]Sig, len(sigs))

	for i, sig := range sigs {
		if ret[i], err = MakeSig(sig); err != nil {
			break
		}
	}

	return ret, err
}

// MakeFundingReqFromPerun returns a new Funding or an error.
func MakeFundingReqFromPerun(req *pchannel.FundingReq) (Funding, error) {
	ident, err := MakeOffIdent(req.Params.Parts[req.Idx])

	return Funding{
		req.State.ID,
		ident,
	}, err
}

// MakeOnIdent returns a new OnIdentity or an error.
func MakeOnIdent(addr pwallet.Address) (OnIdentity, error) {
	var ret OnIdentity
	data := addr.Bytes()

	if len(data) != OnIdentityLen {
		return ret, ErrIdentLenMismatch
	}
	copy(ret[:], data)

	return ret, nil
}

// MakeOffIdent returns a new OffIdentity or an error.
func MakeOffIdent(part pwallet.Address) (OffIdentity, error) {
	var ret OffIdentity
	data := part.Bytes()

	if len(data) != OffIdentityLen {
		return ret, ErrIdentLenMismatch
	}
	copy(ret[:], data)

	return ret, nil
}

// MakeOffIdents returns a new []OffIdentity or an error.
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
