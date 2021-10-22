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
	"context"

	"github.com/centrifuge/go-substrate-rpc-client/v3/types"
	"github.com/pkg/errors"
	pchannel "perun.network/go-perun/channel"
	"perun.network/go-perun/log"
	pwallet "perun.network/go-perun/wallet"

	"github.com/perun-network/perun-polkadot-backend/channel"
	"github.com/perun-network/perun-polkadot-backend/pkg/substrate"
)

// Adjudicator implements the Perun Adjudicator interface.
type Adjudicator struct {
	log.Embedding

	pallet     *Pallet
	storage    substrate.StorageQueryer
	onChain    pwallet.Account
	pastBlocks types.BlockNumber
}

var (
	// ErrConcludedDifferentVersion a channel was concluded with a different version.
	ErrConcludedDifferentVersion = errors.New("channel was concluded with a different version")
	//ErrAdjudicatorReqIncompatible the adjudicator request was not compatible
	ErrAdjudicatorReqIncompatible = errors.New("adjudicator request was not compatible")
)

// NewAdjudicator returns a new Adjudicator.
func NewAdjudicator(onChain pwallet.Account, pallet *Pallet, storage substrate.StorageQueryer, pastBlocks types.BlockNumber) *Adjudicator {
	return &Adjudicator{log.MakeEmbedding(log.Get()), pallet, storage, onChain, pastBlocks}
}

// Register registers and disputes a channel.
func (a *Adjudicator) Register(ctx context.Context, req pchannel.AdjudicatorReq, states []pchannel.SignedState) error {
	defer a.Log().Trace("register done")
	// Input validation.
	if err := a.checkRegister(req, states); err != nil {
		return err
	}
	// Execute dispute.
	return a.dispute(ctx, req)
}

// dispute sends a dispute Ext and waits for the event.
func (a *Adjudicator) dispute(ctx context.Context, req pchannel.AdjudicatorReq) error {
	defer a.Log().Trace("Dispute done")

	// Setup the subscription for Disputed events.
	sub, err := a.pallet.Subscribe(channel.EventIsDisputed(req.Params.ID()), a.pastBlocks)
	if err != nil {
		return err
	}
	defer sub.Close()
	// Build Dispute Tx.
	ext, err := a.pallet.BuildDispute(a.onChain, req.Params, req.Tx.State, req.Tx.Sigs)
	if err != nil {
		return err
	}
	a.Log().WithField("cid", req.Tx.ID).WithField("version", req.Tx.Version).Debug("Dispute")
	// Send and wait for TX finalization.
	if err := a.call(ctx, ext); err != nil {
		return err
	}
	// Wait for disputed event.
	return a.waitForDispute(ctx, sub, req.Tx.Version)
}

// waitForDispute blocks until a dispute event with version greater or equal to
// the specified version is received.
func (a *Adjudicator) waitForDispute(ctx context.Context, sub *EventSub, version channel.Version) error {
	a.Log().Tracef("Waiting for dispute event with version >= %d", version)
	defer a.Log().Trace("waitForDispute returned")

loop:
	for {
		select {
		case _event := <-sub.Events(): // never returns nil
			event := _event.(*channel.DisputedEvent)
			if !event.Phase.IsApplyExtrinsic {
				continue loop
			}
			if event.State.Version < version {
				a.Log().Tracef("Discarded dispute event. Version: %d", event.State.Version)
				continue loop
			}

			a.Log().Debugf("Accepted dispute event. Version: %d", event.State.Version)
			return nil
		case err := <-sub.Err():
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// Withdraw concludes a channel and withdraws all funds.
func (a *Adjudicator) Withdraw(ctx context.Context, req pchannel.AdjudicatorReq, states pchannel.StateMap) error {
	if len(states) != 0 {
		return errors.New("sub-channels unsupported")
	}
	a.Log().WithField("version", req.Tx.Version).Tracef("Withdrawing cid: 0x%x", req.Tx.ID)

	if err := a.ensureConcluded(ctx, req); err != nil {
		return err
	}
	return a.withdraw(ctx, req)
}

// withdraw sends and waits for a withdrawal extrinsic.
func (a *Adjudicator) withdraw(ctx context.Context, req pchannel.AdjudicatorReq) error {
	ext, err := a.pallet.BuildWithdraw(a.onChain, req.Acc, req.Tx.ID)
	if err != nil {
		return err
	}
	return a.call(ctx, ext)
}

// Progress returns an error because app channels are currently not supported.
func (a *Adjudicator) Progress(ctx context.Context, req pchannel.ProgressReq) error {
	return errors.New("progression not supported")
}

// Subscribe subscribes to adjudicator events.
func (a *Adjudicator) Subscribe(ctx context.Context, cid pchannel.ID) (pchannel.AdjudicatorSubscription, error) {
	return NewAdjudicatorSub(cid, a.pallet, a.storage, a.pastBlocks)
}

// ensureConcluded ensures that a channel was concluded.
// Returns an `ErrConcludedDifferentVersion` if the version of the state
// that the channel was concluded with does not match the expected version.
func (a *Adjudicator) ensureConcluded(ctx context.Context, req pchannel.AdjudicatorReq) error {
	// Fetch on-chain dispute.
	dis, err := a.pallet.QueryStateRegister(req.Params.ID(), a.storage, a.pastBlocks)
	if err != nil {
		return err
	}
	// Non-final states need to respect the dispute timeout, if there is a dispute.
	if dis != nil && !req.Tx.IsFinal {
		timeout := channel.MakeTimeout(dis.Timeout, a.storage)
		if err := timeout.Wait(ctx); err != nil {
			return err
		}
	}

	// Setup the subscription for Concluded events.
	sub, err := a.pallet.Subscribe(channel.EventIsConcluded(req.Params.ID()), a.pastBlocks)
	if err != nil {
		return err
	}
	defer sub.Close()
	// Build our conclude Extrinsic.
	ext, err := a.pallet.BuildConclude(a.onChain, req.Params, req.Tx.State, req.Tx.Sigs)
	if err != nil {
		return err
	}
	// Send the Extrinsic.
	if err := a.call(ctx, ext); err != nil {
		return err
	}
	// Wait for a concluded event that either we or some other party caused.
	if err := a.waitForConcluded(ctx, sub, req.Tx.ID); err != nil {
		return err
	}
	// Fetch on-chain dispute again since the `Conduded` event
	// does not contain the version.
	dis, err = a.pallet.QueryStateRegister(req.Params.ID(), a.storage, a.pastBlocks)
	if err != nil {
		return err
	}
	// Check that our version was concluded.
	if req.Tx.Version != dis.State.Version {
		return ErrConcludedDifferentVersion
	}
	return nil
}

// waitForConcluded waits for a concluded event of the specified channel.
func (a *Adjudicator) waitForConcluded(ctx context.Context, sub *EventSub, ID channel.ChannelID) error {
	a.Log().Tracef("Waiting for conclude event")

loop:
	for {
		select {
		case _event := <-sub.Events(): // never returns nil
			event := _event.(*channel.ConcludedEvent)
			if !event.Phase.IsApplyExtrinsic {
				continue loop
			}

			a.Log().WithField("cid", ID).Debugf("Accepted Concluded event")
			return nil
		case err := <-sub.Err():
			return err
		}
	}
}

// call sends an Extrinsic and waits for it to be finalized.
// Does not indicate whether or not the call succeeded.
// Use EventSubs to be sure that the call did not error.
func (a *Adjudicator) call(ctx context.Context, ext *types.Extrinsic) error {
	sub, err := a.pallet.Transact(ext)
	if err != nil {
		return err
	}
	defer sub.Close()
	a.Log().Trace("waiting for TX confirmation")
	return sub.WaitUntil(ctx, substrate.ExtIsFinal)
}

// checkRegister returns an `ErrAdjudicatorReqIncompatible` error if
// the passed request cannot be handled by the Adjudicator.
func (*Adjudicator) checkRegister(req pchannel.AdjudicatorReq, states []pchannel.SignedState) error {
	switch {
	case req.Secondary:
		return errors.WithMessage(ErrAdjudicatorReqIncompatible, "secondary is not supported")
	case req.Tx.IsFinal:
		return errors.WithMessage(ErrAdjudicatorReqIncompatible, "cannot dispute a final state")
	case len(states) != 0:
		return errors.WithMessage(ErrAdjudicatorReqIncompatible, "sub-channels unsupported")
	default:
		return nil
	}
}
