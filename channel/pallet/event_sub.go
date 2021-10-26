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
	"github.com/centrifuge/go-substrate-rpc-client/v3/types"
	"perun.network/go-perun/log"
	pkgsync "perun.network/go-perun/pkg/sync"

	"github.com/perun-network/perun-polkadot-backend/channel"
	"github.com/perun-network/perun-polkadot-backend/pkg/substrate"
)

type (
	// EventSub listens on events and can filter them with an EventPredicate.
	EventSub struct {
		*pkgsync.Closer
		log.Embedding

		source  *substrate.EventSource
		p       EventPredicate
		sink    chan channel.PerunEvent
		errChan chan error
	}

	// EventPredicate can be used to filter events.
	EventPredicate func(channel.PerunEvent) bool
)

// NewEventSub creates a new EventSub.
// Takes ownership of `source` and closes it when done.
func NewEventSub(source *substrate.EventSource, meta *types.Metadata, p EventPredicate) *EventSub {
	sub := &EventSub{Closer: new(pkgsync.Closer), Embedding: log.MakeEmbedding(log.Get()), source: source, sink: make(chan channel.PerunEvent, substrate.ChanBuffSize), p: p, errChan: make(chan error, 1)}
	sub.OnClose(func() {
		if err := source.Close(); err != nil {
			sub.Log().WithError(err).Error("Could not close Closer.")
		}
	})

	go func() {
		sub.Log().Debug("EventSub started")
		defer sub.Log().Debug("EventSub stopped")
		defer sub.Close()

		var err error
	loop:
		for {
			select {
			case set := <-source.Events():
				if err = sub.decodeEventRecords(set, meta); err != nil {
					break loop
				}
			case err = <-source.Err():
				break loop
			case <-sub.Closed():
				break loop
			}
		}
		sub.errChan <- err
		close(sub.errChan)
	}()
	return sub
}

// decodeEventRecords decodes all Perun events from the passed event records.
func (p *EventSub) decodeEventRecords(rawRecord types.EventRecordsRaw, meta *types.Metadata) error {
	record := channel.EventRecords{}
	if err := rawRecord.DecodeEventRecords(meta, &record); err != nil {
		return err
	}
	for _, e := range record.Events() {
		if p.p(e) {
			p.sink <- e
		}
	}
	return nil
}

// Events returns the channel that contains all Perun events.
// Will never be closed.
func (p *EventSub) Events() <-chan channel.PerunEvent {
	return p.sink
}

// Err returns the error channel. Will be closed when the subscription is closed.
func (p *EventSub) Err() <-chan error {
	return p.errChan
}
