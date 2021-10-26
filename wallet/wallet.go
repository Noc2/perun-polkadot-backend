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

package wallet

import (
	pwallet "perun.network/go-perun/wallet"

	"github.com/centrifuge/go-substrate-rpc-client/v3/types"
	"github.com/perun-network/perun-polkadot-backend/pkg/substrate"
)

type (
	// Account defines a Perun Account with substrate specific functions.
	Account interface {
		pwallet.Account

		substrate.ExtSigner
	}

	// Address defines a Perun Address with substrate specific functions.
	Address interface {
		pwallet.Address

		AccountID() types.AccountID
	}
)

// IsAcc returns whether a Perun Account has the expected Account type.
func IsAcc(acc pwallet.Account) bool {
	_, ok := acc.(Account)
	return ok
}

// AsAcc returns a Perun Account as Account. Panics if the conversion failed.
func AsAcc(acc pwallet.Account) Account {
	return acc.(Account)
}

// IsAddr returns whether a Perun Address has the expected Address type.
func IsAddr(acc pwallet.Address) bool {
	_, ok := acc.(Address)
	return ok
}

// AsAddr returns a Perun Address as Address. Panics if the conversion failed.
func AsAddr(acc pwallet.Address) Address {
	return acc.(Address)
}
