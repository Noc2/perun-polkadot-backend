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

package test

import (
	"context"

	pchannel "perun.network/go-perun/channel"
	pkgerrors "perun.network/go-perun/pkg/errors"

	"github.com/perun-network/perun-polkadot-backend/channel/pallet"
)

// FundAll executes all requests with the given funders in parallel.
func FundAll(ctx context.Context, funders []*pallet.Funder, reqs []*pchannel.FundingReq) error {
	g := pkgerrors.NewGatherer()
	for i := range funders {
		i := i
		g.Go(func() error {
			return funders[i].Fund(ctx, *reqs[i])
		})
	}

	if g.WaitDoneOrFailedCtx(ctx) {
		return ctx.Err()
	}
	return g.Err()
}
