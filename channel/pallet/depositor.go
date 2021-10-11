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

	"github.com/perun-network/perun-polkadot-backend/channel"
	"github.com/perun-network/perun-polkadot-backend/pkg/substrate"
	"github.com/pkg/errors"
	pchannel "perun.network/go-perun/channel"
	"perun.network/go-perun/log"
	pwallet "perun.network/go-perun/wallet"
)

type (
	// Depositor deposits funds into a channel for a participant.
	Depositor struct {
		log.Embedding

		pallet *Pallet
	}

	// DepositReq contains values to specify a Deposit.
	DepositReq struct {
		Balance   pchannel.Bal
		Account   pwallet.Account
		FundingId channel.FundingId
	}
)

// ErrFundingReqIncompatible the funding request was incompatible.
var ErrFundingReqIncompatible = errors.New("incompatible funding request")

// NewDepositReq returns a new DepositReq.
func NewDepositReq(bal pchannel.Bal, acc pwallet.Account, fid channel.FundingId) *DepositReq {
	return &DepositReq{bal, acc, fid}
}

// NewDepositReqFromPerun returns a new deposit request from a Perun
// funding request.
func NewDepositReqFromPerun(req *pchannel.FundingReq, acc pwallet.Account) (*DepositReq, error) {
	if !req.Agreement.Equal(req.State.Balances) && (len(req.Agreement) == 1) {
		return nil, ErrFundingReqIncompatible
	}
	bal := req.Agreement[0][req.Idx]
	fReq, err := channel.MakeFundingReqFromPerun(req)
	if err != nil {
		// There is no WithError, so we use the error string here…
		return nil, errors.WithMessage(ErrFundingReqIncompatible, err.Error())
	}
	fid, err := fReq.ID()
	return NewDepositReq(bal, acc, fid), err
}

// NewDepositor returns a new Depositor.
func NewDepositor(pallet *Pallet) *Depositor {
	return &Depositor{log.MakeEmbedding(log.Get()), pallet}
}

// Deposit deposits funds into a channel as specified by the request.
// Returns as soon as the transaction was finalized; this does not guarantee success.
func (d *Depositor) Deposit(ctx context.Context, req *DepositReq) error {
	ext, err := d.pallet.BuildDeposit(req.Account, req.Balance, req.FundingId)
	if err != nil {
		return err
	}
	d.Log().WithField("fid", req.FundingId).Debugf("Depositing %v", req.Balance)
	tx, err := d.pallet.Transact(ext)
	if err != nil {
		return err
	}
	defer tx.Close()
	return tx.WaitUntil(ctx, substrate.ExtIsFinal)
}
