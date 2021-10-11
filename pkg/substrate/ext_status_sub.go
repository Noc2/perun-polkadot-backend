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

	"github.com/centrifuge/go-substrate-rpc-client/v3/rpc/author"
	"github.com/centrifuge/go-substrate-rpc-client/v3/types"
	"github.com/pkg/errors"
	pkgsync "perun.network/go-perun/pkg/sync"
)

type (
	// ExtStatusSub can be used to subscribe to the status of an Extrinsic.
	ExtStatusSub struct {
		pkgsync.Closer

		sub *author.ExtrinsicStatusSubscription
	}

	// ExtStatusPred can be used to filter the status of an Extrinsic.
	ExtStatusPred func(*types.ExtrinsicStatus) bool
)

// NewExtStatusSub returns a new ExtStatusSub and takes ownership of the passed sub.
func NewExtStatusSub(sub *author.ExtrinsicStatusSubscription) *ExtStatusSub {
	ret := &ExtStatusSub{sub: sub}
	ret.Closer.OnCloseAlways(sub.Unsubscribe)
	return ret
}

// WaitUntil waits until the predicate returns true or the context is cancelled.
// Can be used for example to wait until an Extrinsic is final with `ExtIsFinal`.
func (e *ExtStatusSub) WaitUntil(ctx context.Context, until ExtStatusPred) error {
	for {
		select {
		case status := <-e.sub.Chan():
			if until(&status) {
				return nil
			}
		case err := <-e.sub.Err():
			return errors.WithMessage(err, "underlying sub closed")
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// Close closes the subscription.
func (e *ExtStatusSub) Close() {
	e.Closer.Close()
}

// ExtIsFinal returns whether an Extrinsic is final.
func ExtIsFinal(status *types.ExtrinsicStatus) bool {
	return status.IsFinalized
}
