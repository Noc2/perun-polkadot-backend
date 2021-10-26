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

package pallet

import (
	"errors"

	"github.com/centrifuge/go-substrate-rpc-client/v3/types"
	pchannel "perun.network/go-perun/channel"
	"perun.network/go-perun/log"
	pwallet "perun.network/go-perun/wallet"

	"github.com/perun-network/perun-polkadot-backend/channel"
	"github.com/perun-network/perun-polkadot-backend/pkg/substrate"
	"github.com/perun-network/perun-polkadot-backend/wallet"
)

// Pallet exposes all functions of the Perun-pallet.
// https://github.com/perun-network/perun-polkadot-pallet
type Pallet struct {
	log.Embedding

	*substrate.Pallet
	meta *types.Metadata
}

// PerunPallet is the name of the Perun pallet as configured by the
// substrate chain.
const PerunPallet = "PerunModule"

var (
	// Deposit is the name of the deposit function of the pallet.
	Deposit = substrate.NewExtName(PerunPallet, "deposit")
	// Dispute is the name of the dispute function of the pallet.
	Dispute = substrate.NewExtName(PerunPallet, "dispute")
	// Conclude is the name of the conclude function of the pallet.
	Conclude = substrate.NewExtName(PerunPallet, "conclude")
	// Withdraw is the name of the withdraw function of the pallet.
	Withdraw = substrate.NewExtName(PerunPallet, "withdraw")

	// ErrNoDeposit no deposit could be found.
	ErrNoDeposit = errors.New("no deposit found")
	// ErrNoRegisteredState no registered state could be found.
	ErrNoRegisteredState = errors.New("no registered state")
)

// NewPallet returns a new Pallet.
func NewPallet(pallet *substrate.Pallet, meta *types.Metadata) *Pallet {
	return &Pallet{log.MakeEmbedding(log.Get()), pallet, meta}
}

// NewPerunPallet is a wrapper around NewPallet and returns a new Perun pallet.
func NewPerunPallet(api *substrate.API) *substrate.Pallet {
	return substrate.NewPallet(api, PerunPallet)
}

// Subscribe returns an EventSub that listens on all events of the pallet.
func (p *Pallet) Subscribe(f EventPredicate, pastBlocks types.BlockNumber) (*EventSub, error) {
	source, err := p.Pallet.Subscribe(pastBlocks)
	if err != nil {
		return nil, err
	}

	return NewEventSub(source, p.meta, f), nil
}

// QueryDeposit returns the deposit for a funding ID
// or ErrNoDeposit if no deposit was found.
func (p *Pallet) QueryDeposit(fid channel.FundingID, storage substrate.StorageQueryer, pastBlocks types.BlockNumber) (*channel.Balance, error) {
	// Create the query.
	key, err := p.BuildQuery("Deposits", fid[:])
	if err != nil {
		return nil, err
	}
	p.Log().WithField("fid", fid).Trace("Querying deposit")
	// Retrieve the last value, beginning at pastBlocks.
	res, err := storage.QueryOne(pastBlocks, key)
	if err != nil {
		return nil, err
	}
	if len(res.StorageData) == 0 {
		return nil, ErrNoDeposit
	}
	// Decode the result as a Balance.
	ret := new(channel.Balance)
	return ret, channel.ScaleDecode(ret, res.StorageData)
}

// QueryStateRegister returns the registered state of a channel
// or ErrNoRegisteredState if no state was found.
func (p *Pallet) QueryStateRegister(cid channel.ChannelID, storage substrate.StorageQueryer, pastBlocks types.BlockNumber) (*channel.RegisteredState, error) {
	// Create the query.
	key, err := p.BuildQuery("StateRegister", cid[:])
	if err != nil {
		return nil, err
	}
	p.Log().WithField("cid", cid).Trace("Querying stateRegister")
	// Retrieve the last value, beginning at pastBlocks.
	res, err := storage.QueryOne(pastBlocks, key)
	if err != nil {
		return nil, err
	}
	if len(res.StorageData) == 0 {
		return nil, ErrNoRegisteredState
	}
	// Decode the result as a RegisteredState.
	ret := new(channel.RegisteredState)
	return ret, channel.ScaleDecode(ret, res.StorageData)
}

// BuildDeposit returns an extrinsic that funds the specified funding ID.
func (p *Pallet) BuildDeposit(acc pwallet.Account, _amount pchannel.Bal, fid channel.FundingID) (*types.Extrinsic, error) {
	amount, err := channel.MakeBalance(_amount)
	if err != nil {
		return nil, err
	}
	return p.BuildExt(Deposit, []interface{}{
		fid,
		amount},
		wallet.AsAddr(acc.Address()).AccountID(),
		wallet.AsAcc(acc))
}

// BuildDispute returns an extrinsic that disputes a channel.
func (p *Pallet) BuildDispute(acc pwallet.Account, params *pchannel.Params, state *pchannel.State, sigs []pwallet.Sig) (*types.Extrinsic, error) {
	_params, err := channel.NewParams(params)
	if err != nil {
		return nil, err
	}
	_state, err := channel.NewState(state)
	if err != nil {
		return nil, err
	}
	_sigs, err := channel.MakeSigs(sigs)
	if err != nil {
		return nil, err
	}

	return p.BuildExt(Dispute,
		[]interface{}{
			_params,
			_state,
			_sigs},
		wallet.AsAddr(acc.Address()).AccountID(),
		wallet.AsAcc(acc))
}

// BuildConclude returns an extrinsic that concludes a channel.
func (p *Pallet) BuildConclude(acc pwallet.Account, params *pchannel.Params, state *pchannel.State, sigs []pwallet.Sig) (*types.Extrinsic, error) {
	_params, err := channel.NewParams(params)
	if err != nil {
		return nil, err
	}
	_state, err := channel.NewState(state)
	if err != nil {
		return nil, err
	}
	_sigs, err := channel.MakeSigs(sigs)
	if err != nil {
		return nil, err
	}

	return p.BuildExt(Conclude,
		[]interface{}{
			_params,
			_state,
			_sigs},
		wallet.AsAddr(acc.Address()).AccountID(),
		wallet.AsAcc(acc))
}

// BuildWithdraw returns an extrinsic that withdraws all funds from the channel.
func (p *Pallet) BuildWithdraw(onChain, offChain pwallet.Account, cid pchannel.ID) (*types.Extrinsic, error) {
	part := offChain.Address()
	receiver := onChain.Address()

	withdrawal, err := channel.NewWithdrawal(cid, part, receiver)
	if err != nil {
		return nil, err
	}
	data, err := channel.ScaleEncode(withdrawal)
	if err != nil {
		return nil, err
	}
	sig, err := offChain.SignData(data)
	if err != nil {
		return nil, err
	}
	_sig, err := channel.MakeSig(sig)
	if err != nil {
		return nil, err
	}

	return p.BuildExt(Withdraw,
		[]interface{}{
			withdrawal,
			_sig},
		wallet.AsAddr(onChain.Address()).AccountID(),
		wallet.AsAcc(onChain))
}
