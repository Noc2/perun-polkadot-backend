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

	"github.com/centrifuge/go-substrate-rpc-client/v3/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
	pchannel "perun.network/go-perun/channel"
)

const (
	// OffIdentityLen is the length of an OffIdentity in byte.
	OffIdentityLen = 32
	// OnIdentityLen is the length of an OnIdentity in byte.
	OnIdentityLen = 32
	// NonceLen is the length of a Nonce in byte.
	NonceLen = 32
	// SigLen is the length of a Sig in byte.
	SigLen = 64
	// FIDLen is the length of a FundingId in byte.
	FIDLen = 32
)

type (
	// Nonce makes a channels ID unique by providing randomness to the params.
	Nonce = [NonceLen]byte
	// ChannelID the ID of a channel as defined by go-perun.
	ChannelID = pchannel.ID
	// FundingID used to a the funding of a participant in a channel.
	FundingID = [FIDLen]byte
	// OffIdentity is an off-chain identity.
	OffIdentity = [OffIdentityLen]byte
	// OnIdentity is an on-chain identity.
	OnIdentity = [OnIdentityLen]byte
	// Version of a state as defined by go-perun.
	Version = uint64
	// ChallengeDuration the duration of a challenge as defined by go-perun.
	ChallengeDuration = uint64
	// Balance is the balance of an on- or off-chain Address.
	Balance = types.U128
	// Sig is an off-chain signature.
	Sig = [SigLen]byte
	// AppID is the identifier of a channel application.
	AppID = OffIdentity

	// Params holds the fixed parameters of a channel and uniquely identifies it.
	Params struct {
		// Nonce is the unique nonce of a channel.
		Nonce Nonce
		// Participants are the off-chain participants of a channel.
		Participants []OffIdentity
		// ChallengeDuration is the duration that disputes can be refuted in.
		ChallengeDuration ChallengeDuration
		// App is the identifier of the channel application.
		App AppID
	}

	// State is the state of a channel.
	State struct {
		// Channel is the unique ID of the channel that this state belongs to.
		Channel ChannelID
		// Version is the version of the state.
		Version Version
		// Balances are the balances of the participants.
		Balances []Balance
		// Final whether or not this state is the final one.
		Final bool
		// Data is the channel's application data.
		Data []byte
	}

	// Withdrawal is used by a participant to withdraw his on-chain funds.
	Withdrawal struct {
		// Channel is the channel from which to withdraw.
		Channel ChannelID
		// Part is the participant who wants to withdraw.
		Part OffIdentity
		// Receiver is the receiver of the withdrawal.
		Receiver OnIdentity
	}

	// Funding is used to calculate a FundingId.
	Funding struct {
		// Channel is the channel to fund.
		Channel ChannelID
		// Part is the participant who wants to fund.
		Part OffIdentity
	}

	// RegisteredState is a channel state that was registered on-chain.
	RegisteredState struct {
		// State is the state of the channel.
		State State
		// App is the app of the channel.
		App AppID
		// Timeout is the duration that the dispute can be refuted in.
		Timeout ChallengeDuration
		// Concluded whether the channel is concluded.
		Concluded bool
	}
)

var (
	// MaxBalance is the highest possible value of a Balance.
	// Substrate uses U128 for balance representation as opposed to go-perun
	// which uses big.Int so this restriction is necessary.
	MaxBalance Balance = types.NewU128(*new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1)))

	// ErrNonceOutOfRange a nonce was out of range of valid values.
	ErrNonceOutOfRange = errors.New("nonce values was out of range")
	// ErrAllocIncompatible an allocation was incompatible.
	ErrAllocIncompatible = errors.New("incompatible allocation")
	// ErrStateIncompatible a state was incompatible.
	ErrStateIncompatible = errors.New("incompatible state")
	// ErrIdentLenMismatch the length of an identity was wrong.
	ErrIdentLenMismatch = errors.New("length of an identity was wrong")
)

// NewFunding returns a new Funding.
func NewFunding(id ChannelID, part OffIdentity) *Funding {
	return &Funding{id, part}
}

// ID calculates the funding ID by encoding and hashing the Funding.
func (p Funding) ID() (FundingID, error) {
	var fid FundingID
	data, err := ScaleEncode(&p)
	if err != nil {
		return fid, errors.WithMessage(err, "calculating funding ID")
	}
	return FundingID(crypto.Keccak256Hash(data)), nil
}
