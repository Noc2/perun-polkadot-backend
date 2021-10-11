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

package pallet_test

import (
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/perun-network/perun-polkadot-backend/channel"
	"github.com/perun-network/perun-polkadot-backend/channel/pallet"
	"github.com/perun-network/perun-polkadot-backend/channel/pallet/test"
	chtest "github.com/perun-network/perun-polkadot-backend/channel/test"
)

func TestPalletEventSub_Deposit(t *testing.T) {
	n := 3
	s := test.NewSetup(t)
	params, state := s.NewRandomParamAndState()
	fSetup := chtest.NewFundingSetup(params, state)
	dSetup := chtest.NewDepositSetup(fSetup, s.Alice.Acc)

	// Subscribe to 'deposited' events for the correct funding ID.
	sub, err := s.Pallet.Subscribe(func(_event channel.PerunEvent) bool {
		event, ok := _event.(*channel.DepositedEvent)
		if !ok {
			return false
		}
		return event.Fid == fSetup.Fids[0]
	}, 0)
	require.NoError(t, err)

	// Fund `n` times.
	for i := 0; i < n; i++ {
		require.NoError(t, s.Deps[0].Deposit(s.NewCtx(), dSetup.DReqs[0]))
	}

	aliceBal := state.Balances[0][0]
	// Wait for `n` events and check that the values match.
	events := AssertNEvents(t, 2*s.BlockTime, sub, n)
	for i, _event := range events {
		event := _event.(*channel.DepositedEvent)

		absolute := new(big.Int).Mul(aliceBal, big.NewInt(int64(i+1)))
		require.Equal(t, channel.MakePerunBalance(event.Balance), absolute)
	}
	sub.Close()
	assert.NoError(t, <-sub.Err())
}

// AssertNEvents reads `n` events and waits that there arrives no new event
// within the passed timeout. Returns the events.
func AssertNEvents(t *testing.T, timeout time.Duration, sub *pallet.EventSub, n int) []channel.PerunEvent {
	events := make([]channel.PerunEvent, n)
	// Read `n` events.
	for i := 0; i < n; i++ {
		select {
		case events[i] = <-sub.Events():
		case err := <-sub.Err():
			t.Errorf("Error channel received value: %v", err)
			t.FailNow()
		}
	}
	// Check that no more events arrive.
	select {
	case <-time.After(timeout):
	case event := <-sub.Events():
		t.Errorf("Got event but expected none but got: %#v", event)
		t.FailNow()
	case <-sub.Err():
		t.Error("Got error but expected none")
		t.FailNow()
	}

	return events
}
