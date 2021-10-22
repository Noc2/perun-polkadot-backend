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
	"github.com/centrifuge/go-substrate-rpc-client/v3/types"
)

type (
	// EventRecords contains all events that can be emitted by a substrate chain
	// that has the the Perun pallet deployed.
	EventRecords struct {
		types.EventRecords

		PerunModule_Deposited []DepositedEvent
		PerunModule_Disputed  []DisputedEvent
		PerunModule_Concluded []ConcludedEvent
		PerunModule_Withdrawn []WithdrawnEvent
	}

	// PerunEvent is a Perun event.
	PerunEvent interface{}

	// DepositedEvent is emitted when a deposit is received.
	DepositedEvent struct {
		Phase   types.Phase // required
		Fid     FundingId
		Balance Balance      // total deposit of the Fid
		Topics  []types.Hash // required
	}

	// DisputedEvent is emitted when a dispute was opened or updated.
	DisputedEvent struct {
		Phase  types.Phase // required
		Cid    ChannelID
		State  State
		Topics []types.Hash // required
	}

	// ConcludedEvent is emitted when a channel is concluded.
	ConcludedEvent struct {
		Phase  types.Phase // required
		Cid    ChannelID
		Topics []types.Hash // required
	}

	// WithdrawnEvent is emitted when all funds are withdrawn from a funding ID.
	WithdrawnEvent struct {
		Phase  types.Phase // required
		Fid    FundingId
		Topics []types.Hash // required
	}
)

// EventIsDeposited returns whether an event is a DepositedEvent.
func EventIsDeposited(e PerunEvent) bool {
	_, ok := e.(*DepositedEvent)
	return ok
}

// EventIsDisputed checks whether an event is a DisputedEvent for a
// specific channel.
func EventIsDisputed(cid ChannelID) func(PerunEvent) bool {
	return func(e PerunEvent) bool {
		event, ok := e.(*DisputedEvent)
		return ok && event.Cid == cid
	}
}

// EventIsConcluded checks whether an event is a ConcludedEvent for a
// specific channel.
func EventIsConcluded(cid ChannelID) func(PerunEvent) bool {
	return func(e PerunEvent) bool {
		event, ok := e.(*ConcludedEvent)
		return ok && event.Cid == cid
	}
}

// EventIsWithdrawn returns whether an event is a WithdrawnEvent.
func EventIsWithdrawn(e PerunEvent) bool {
	_, ok := e.(*WithdrawnEvent)
	return ok
}

// Merge merges all Perun events into one slice and returns it.
func (r *EventRecords) Merge() []PerunEvent {
	var ret []PerunEvent
	// The order here does not matter here since just one slice is non empty.
	for _, e := range r.PerunModule_Deposited {
		e := e
		ret = append(ret, &e)
	}
	for _, e := range r.PerunModule_Disputed {
		e := e
		ret = append(ret, &e)
	}
	for _, e := range r.PerunModule_Concluded {
		e := e
		ret = append(ret, &e)
	}
	for _, e := range r.PerunModule_Withdrawn {
		e := e
		ret = append(ret, &e)
	}
	return ret
}
