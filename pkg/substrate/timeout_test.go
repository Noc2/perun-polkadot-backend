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

package substrate_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ctxtest "perun.network/go-perun/pkg/context/test"

	"github.com/perun-network/perun-polkadot-backend/pkg/substrate"
	"github.com/perun-network/perun-polkadot-backend/pkg/substrate/test"
)

func TestTimeout(t *testing.T) {
	s := test.NewSetup(t)

	waitTime := 5 * s.BlockTime
	deadline := time.Now().Add(waitTime)
	timeout := substrate.NewTimeout(s.API, deadline)

	var err error
	assert.False(t, timeout.IsElapsed(context.Background()))
	ctxtest.AssertNotTerminates(t, waitTime/2, func() {
		_ = timeout.Wait(context.Background())
	})
	assert.False(t, timeout.IsElapsed(context.Background()))
	ctxtest.AssertTerminates(t, waitTime, func() { err = timeout.Wait(context.Background()) })
	require.NoError(t, err)
	assert.True(t, timeout.IsElapsed(context.Background()))
}
