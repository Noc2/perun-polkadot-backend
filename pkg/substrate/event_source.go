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

import (
	"github.com/centrifuge/go-substrate-rpc-client/v3/rpc/state"
	"github.com/centrifuge/go-substrate-rpc-client/v3/types"
	"perun.network/go-perun/log"
	pkgsync "perun.network/go-perun/pkg/sync"
)

type (
	// EventSource collects all events from the chain.
	// Can then be used to filter out events, eg. for a specific pallet.
	EventSource struct {
		*pkgsync.Closer
		log.Embedding

		future *state.StorageSubscription
		api    *API

		events chan types.EventRecordsRaw
		err    chan error
	}

	// EventKey identifies an event type that an EventSource can listen on.
	EventKey struct {
		firstName, lastName string
		args                [][]byte
	}
)

// ChanBuffSize is the buffer size of an EventSource.
// GSRPC buffers up to 20k events.
const ChanBuffSize = 1024

// SystemEventsKey is the key of all system events.
func SystemEventsKey() *EventKey {
	return &EventKey{"System", "Events", nil}
}

// SystemAccountKey is the key to query an account.
func SystemAccountKey(accountID types.AccountID) *EventKey {
	return &EventKey{"System", "Account", [][]byte{accountID[:]}}
}

// Key returns the EventKey as GSRPC key.
func (k *EventKey) Key(meta *types.Metadata) (types.StorageKey, error) {
	return types.CreateStorageKey(meta, k.firstName, k.lastName, k.args...)
}

// NewEventSource returns a new EventSource.
// It queries pastBlocks into the past to retrieve old events and starts listening
// for new events with the passed EventKeys.
func NewEventSource(api *API, pastBlocks types.BlockNumber, keys ...*EventKey) (*EventSource, error) {
	eks, err := mergeEventKeys(api.Metadata(), keys...)
	if err != nil {
		return nil, err
	}
	// Subscribe to future events.
	future, err := api.Subscribe(eks...)
	if err != nil {
		return nil, err
	}

	events := make(chan types.EventRecordsRaw, ChanBuffSize)
	source := &EventSource{
		Closer:    new(pkgsync.Closer),
		Embedding: log.MakeEmbedding(log.Get()),
		future:    future,
		api:       api,
		events:    events,
		err:       make(chan error, 1),
	}
	source.OnClose(func() {
		future.Unsubscribe()
	})
	return source, source.init(eks, pastBlocks)
}

// init reads all past events and starts to forward future events.
func (s *EventSource) init(keys []types.StorageKey, pastBlocks types.BlockNumber) error {
	startBlock, err := s.api.PastBlock(pastBlocks)
	if err != nil {
		return err
	}
	// Query and parse past events.
	past, err := s.api.QueryAll(keys, startBlock)
	if err != nil {
		return err
	}
	s.Log().WithField("pastEvents", len(past)).Debug("PalletEventSource started")
	// Process all past events.
	// This could dead-lock if there are more events than the channel can hold.
	for _, set := range past {
		s.parseEvent(set)
	}
	// Listen to all future events.
	go func() {
		defer s.Log().Debug("PalletEventSource stopped")
		defer close(s.err)
		defer s.Close()

		for {
			select {
			case set := <-s.future.Chan():
				s.parseEvent(set)
			case err := <-s.future.Err():
				s.err <- err
				return
			case <-s.Closed():
				return
			}
		}
	}()
	return nil
}

// parseEvent parses an Event from the passed change set and puts it into the
// events channel.
func (s *EventSource) parseEvent(set types.StorageChangeSet) {
	for _, change := range set.Changes {
		if !change.HasStorageData {
			continue
		}

		s.events <- types.EventRecordsRaw(change.StorageData)
	}
}

// Events returns a channel that contains all events that the EventSource found.
// This channel will never be closed.
func (s *EventSource) Events() <-chan types.EventRecordsRaw {
	return s.events
}

// Err returns the error channel. Will be closed when the EventSource is closed.
func (s *EventSource) Err() <-chan error {
	return s.err
}

func mergeEventKeys(meta *types.Metadata, eks ...*EventKey) ([]types.StorageKey, error) {
	keys := make([]types.StorageKey, len(eks))

	for i, _key := range eks {
		key, err := _key.Key(meta)
		if err != nil {
			return nil, err
		}
		keys[i] = key
	}

	return keys, nil
}
