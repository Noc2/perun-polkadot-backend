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
	"github.com/perun-network/perun-polkadot-backend/channel"
	pchannel "perun.network/go-perun/channel"
	"perun.network/go-perun/log"
	pwallet "perun.network/go-perun/wallet"
)

// Funder implements the Funder interface to fund channels.
type Funder struct {
	log.Embedding

	pallet *Pallet
	acc    pwallet.Account
	// pastBlocks is the number of blocks to query into the past.
	pastBlocks types.BlockNumber
}

// NewFunder returns a new Funder.
func NewFunder(pallet *Pallet, acc pwallet.Account, pastBlocks types.BlockNumber) *Funder {
	return &Funder{log.MakeEmbedding(log.Get()), pallet, acc, pastBlocks}
}

// Fund funds a channel. Needed by the Funder interface.
func (f *Funder) Fund(ctx context.Context, req pchannel.FundingReq) error {
	// Listen for Deposited events.
	sub, err := f.pallet.Subscribe(channel.EventIsDeposited, f.pastBlocks)
	if err != nil {
		return err
	}
	defer sub.Close()

	// Deposit our funds.
	wReq, err := NewDepositReqFromPerun(&req, f.acc)
	if err != nil {
		return err
	}
	if err := NewDepositor(f.pallet).Deposit(ctx, wReq); err != nil {
		return err
	}

	// Wait for all Deposited events.
	return f.waitForFundings(ctx, sub, req)
}

// waitForFundings blocks until either; all fundings of the request were received or
// the context was cancelled.
func (f *Funder) waitForFundings(ctx context.Context, sub *EventSub, req pchannel.FundingReq) error {
	// Calculate the funding IDs.
	fids, err := calcFids(req)
	if err != nil {
		return err
	}
	f.Log().Tracef("Waiting for funding from %d peers", len(fids))
	defer f.Log().Debug("All peers funded")

	for len(fids) != 0 {
		select {
		case _event := <-sub.Events(): // never returns nil
			event := _event.(*channel.DepositedEvent)
			// Only consider final events.
			if !event.Phase.IsApplyExtrinsic {
				continue
			}
			// Find the peer index of the event.
			idx, found := fids[event.Fid]
			if !found {
				f.Log().WithField("fid", event.Fid).Trace("Ignored funding")
				continue
			}
			// Remove the entry from the map if the peer funded enough.
			if need := req.Agreement[0][idx]; event.Balance.Cmp(need) >= 0 {
				delete(fids, event.Fid)
				f.Log().WithField("fid", event.Fid).Tracef("Peer funded successfully, %d remain", len(fids))
			}
		case err := <-sub.Err():
			return err
		case <-ctx.Done():
			return makeTimeoutErr(fids)
		}
	}
	return nil
}

// makeTimeoutErr returns a FundingTimeoutError.
func makeTimeoutErr(remains map[channel.FundingID]pchannel.Index) error {
	indices := make([]pchannel.Index, 0, len(remains))
	for _, idx := range remains {
		indices = append(indices, idx)
	}
	return pchannel.NewFundingTimeoutError(
		[]*pchannel.AssetFundingError{{
			Asset:         channel.Asset.Index(),
			TimedOutPeers: indices,
		}},
	)
}

// calcFids calculates all funding ids of the funding request.
func calcFids(req pchannel.FundingReq) (map[channel.FundingID]pchannel.Index, error) {
	ids := make(map[channel.FundingID]pchannel.Index)

	for i, part := range req.Params.Parts {
		_part, err := channel.MakeOffIdent(part)
		if err != nil {
			return nil, err
		}
		fid, err := channel.NewFunding(req.State.ID, _part).ID()
		if err != nil {
			return nil, err
		}
		ids[fid] = pchannel.Index(i)
	}

	return ids, nil
}
