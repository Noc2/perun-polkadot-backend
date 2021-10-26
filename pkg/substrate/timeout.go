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
	"context"
	"time"

	"github.com/centrifuge/go-substrate-rpc-client/v3/types"
	"perun.network/go-perun/log"
)

type (
	// ExpiredTimeout is always expired.
	// Implements the Perun Timeout interface.
	ExpiredTimeout struct{}

	// Timeout can be used to wait until a specific timepoint is reached by
	// the blockchain. Implements the Perun Timeout interface.
	Timeout struct {
		log.Embedding

		when         time.Time
		pollInterval time.Duration
		storage      StorageQueryer
	}

	// TimePoint as defined by pallet Timestamp.
	TimePoint uint64
)

// DefaultTimeoutPollInterval default value for the PollInterval of a Timeout.
const DefaultTimeoutPollInterval = time.Second

// NewExpiredTimeout returns a new ExpiredTimeout.
func NewExpiredTimeout() *ExpiredTimeout {
	return &ExpiredTimeout{}
}

// IsElapsed returns true.
func (*ExpiredTimeout) IsElapsed(context.Context) bool {
	return true
}

// Wait returns nil.
func (*ExpiredTimeout) Wait(context.Context) error {
	return nil
}

// NewTimeout returns a new Timeout which expires at the given time.
func NewTimeout(storage StorageQueryer, when time.Time, pollInterval time.Duration) *Timeout {
	return &Timeout{log.MakeEmbedding(log.Get()), when, pollInterval, storage}
}

// IsElapsed returns whether the timeout is elapsed.
func (t *Timeout) IsElapsed(context.Context) bool {
	// Get the current time.
	now, err := t.pollTime()
	if err != nil {
		log.WithError(err).Error("Polling time failed")
		return false
	}
	// Check for elapsed. There is no t.Cmp, so use an or here.
	elapsed := t.when.Before(now) || t.when == now
	// Fancy logging.
	if delta := now.Sub(t.when); elapsed {
		t.Log().Tracef("Timeout elapsed since %v", delta)
	} else {
		t.Log().Tracef("Timeout target in %v", delta)
	}

	return elapsed
}

// Wait waits for the timeout or until the context is cancelled.
func (t *Timeout) Wait(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(t.pollInterval):
			if t.IsElapsed(ctx) {
				return nil
			}
		}
	}
}

// pollTime returns the current time of the blockchain.
func (t *Timeout) pollTime() (time.Time, error) {
	key, err := t.storage.BuildKey("Timestamp", "Now")
	if err != nil {
		return time.Unix(0, 0), err
	}
	_now, err := t.storage.QueryOne(0, key)
	if err != nil {
		return time.Unix(0, 0), err
	}

	var now TimePoint
	if err := types.DecodeFromBytes(_now.StorageData, &now); err != nil {
		return time.Unix(0, 0), err
	}
	unixNow := time.Unix(int64(now/1000), 0)
	t.Log().Tracef("Polled time: %v", unixNow.UTC())
	return unixNow, nil
}
