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
	"sync"

	pwallet "perun.network/go-perun/wallet"

	"github.com/ChainSafe/go-schnorrkel"
	"github.com/centrifuge/go-substrate-rpc-client/v3/types"
	"github.com/pkg/errors"

	"github.com/perun-network/perun-polkadot-backend/pkg/substrate"
)

// Wallet implements the Perun Wallet interface. It uses sr25519 cryptography.
type Wallet struct {
	mtx     *sync.RWMutex               // protects all
	keyRing map[pwallet.AddrKey]keyPair // address -> keyPair
}

var (
	// ErrWrongAddrType is returned when the type of a Perun Address was not
	// of type Address.
	ErrWrongAddrType = errors.New("got wrong address type")
	// ErrAccountNotPresent is returned when no Account could be found for
	// a specific Address.
	ErrAccountNotPresent = errors.New("account is not present in the wallet")
)

// NewWallet returns a new wallet.
func NewWallet() *Wallet {
	return &Wallet{
		mtx:     new(sync.RWMutex),
		keyRing: make(map[pwallet.AddrKey]keyPair),
	}
}

// ImportSK can be used to import a secret key into the wallet.
func (w *Wallet) ImportSK(sk *schnorrkel.MiniSecretKey) *Account {
	w.mtx.Lock()
	defer w.mtx.Unlock()

	keyPair := makeKeyPair(sk)
	w.keyRing[keyPair.pk.key()] = keyPair

	return newAccount(w, keyPair.pk)
}

// Unlock unlocks an Account. Needed by the Perun Wallet interface.
// Returns ErrWrongAddrType or ErrAccountNotPresent in case of an error.
func (w *Wallet) Unlock(pAddr pwallet.Address) (pwallet.Account, error) {
	w.mtx.RLock()
	defer w.mtx.RUnlock()

	if !IsAddr(pAddr) {
		return nil, ErrWrongAddrType
	}
	addr := AsAddr(pAddr)
	// Check if the account exists.
	if _, found := w.keyRing[addr.key()]; !found {
		return nil, ErrAccountNotPresent
	}
	return newAccount(w, addr), nil
}

// SignTx signs an Extrinsic for an address that is present in the wallet;
// otherwise returns ErrAccountNotPresent.
func (w *Wallet) signExt(addr *Address, tx *types.Extrinsic, opts types.SignatureOptions, net substrate.NetworkID) error {
	w.mtx.RLock()
	defer w.mtx.RUnlock()

	kp, found := w.keyRing[addr.key()]
	if !found {
		return ErrAccountNotPresent
	}
	// Temporarily re-derive a KeyringPair for signing. GSRPC requires it.
	gsrpcKp, err := kp.keyRing(net)
	if err != nil {
		return errors.WithMessage(err, "deriving keypair")
	}
	return tx.Sign(gsrpcKp, opts)
}

// signData signs arbitrary data for an address that is present in the wallet;
// otherwise returns ErrAccountNotPresent.
func (w *Wallet) signData(addr *Address, data []byte) ([]byte, error) {
	w.mtx.RLock()
	defer w.mtx.RUnlock()

	kp, found := w.keyRing[addr.key()]
	if !found {
		return nil, ErrAccountNotPresent
	}
	// Prepend SignaturePrefix and sign.
	env := schnorrkel.NewSigningContext(substrate.SignaturePrefix, data)
	_sig, err := kp.sk.Sign(env)
	if err != nil {
		return nil, err
	}
	return encodeSig(_sig), nil
}

// LockAll does nothing. Needed by the Perun Wallet interface.
func (*Wallet) LockAll() {}

// IncrementUsage does nothing. Needed by the Perun Wallet interface.
func (*Wallet) IncrementUsage(pwallet.Address) {}

// DecrementUsage does nothing. Needed by the Perun Wallet interface.
func (*Wallet) DecrementUsage(pwallet.Address) {}
