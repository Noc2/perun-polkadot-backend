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

package sr25519

import (
	"github.com/centrifuge/go-substrate-rpc-client/v3/types"
	"github.com/perun-network/perun-polkadot-backend/pkg/substrate"
	pwallet "perun.network/go-perun/wallet"
)

// Account implements the Account interface.
// Can be used to sign arbitrary data and extrinsics.
type Account struct {
	wallet *Wallet // back pointer to the wallet
	addr   *Address
}

// newAccount returns a new account.
func newAccount(wallet *Wallet, pk *Address) *Account {
	return &Account{
		wallet,
		pk,
	}
}

// Address returns the Address. Needed by the Account interface.
func (acc *Account) Address() pwallet.Address {
	return acc.addr
}

// SignExt signs an extrinsic by modifying it. This behaviour is required by GSRPC.
func (acc *Account) SignExt(ext *types.Extrinsic, opts types.SignatureOptions, net substrate.NetworkID) error {
	return acc.wallet.signExt(acc.addr, ext, opts, net)
}

// SignData signs a byte-slice. Needed by the Account interface.
func (acc *Account) SignData(data []byte) ([]byte, error) {
	return acc.wallet.signData(acc.addr, data)
}

// IsAcc returns whether a Perun Account has the expected Account type.
func IsAcc(acc pwallet.Account) bool {
	_, ok := acc.(*Account)
	return ok
}

// AsAcc returns a Perun Account as Account. Panics if the conversion failed.
func AsAcc(acc pwallet.Account) *Account {
	return acc.(*Account)
}
