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
	"bytes"
	"io"

	"github.com/ChainSafe/go-schnorrkel"
	"github.com/centrifuge/go-substrate-rpc-client/v3/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
	pwallet "perun.network/go-perun/wallet"
)

// Address implements the Address interface.
type Address struct {
	pk *schnorrkel.PublicKey
}

// AddressLen is the length of an encoded Address in byte.
const AddressLen = 32

// NewAddressFromPK returns a new Address from a public key.
func NewAddressFromPK(pk *schnorrkel.PublicKey) *Address {
	return &Address{pk}
}

// Bytes returns the address encoded as byte slice.
func (a *Address) Bytes() []byte {
	// Convert to ensure that the length of schnorrkel PKs did not change.
	bytes := [AddressLen]byte(a.pk.Encode()) // nolint: unconvert
	return bytes[:]
}

// Encode encodes an Address. Needed by the Perun Address interface.
func (a *Address) Encode(w io.Writer) error {
	return encodeAddress(a, w)
}

// Decode decodes an Address. Needed by the Perun Address interface.
func (a *Address) Decode(r io.Reader) error {
	var pk [AddressLen]byte
	if _, err := io.ReadFull(r, pk[:]); err != nil {
		return err
	}
	a.pk = schnorrkel.NewPublicKey(pk)
	return nil
}

// AccountID returns the substrate account id of an address.
func (a *Address) AccountID() types.AccountID {
	return a.pk.Encode()
}

// String returns the AccountID as hex string with 0x prefix.
// Needed by the Perun Address interface.
func (a *Address) String() string {
	aid := a.AccountID()
	return hexutil.Encode(aid[:])
}

// key returns the address encoded as Perun wallet AddrKey.
func (a *Address) key() pwallet.AddrKey {
	return pwallet.AddrKey(a.Bytes())
}

// Equals returns whether the passed address is equal to the receiver.
// Panics if the passed address is not of type Address.
// Needed by the Perun Address interface.
func (a *Address) Equals(b pwallet.Address) bool {
	return bytes.Equal(a.Bytes(), AsAddr(b).Bytes())
}

// Cmp returns 0 if a == b, -1 if a < b, and +1 if a > b.
// Where ==, < and > are an arbitrary but fixed total order over Address.
// Panics if the passed address is not of type Address.
// Needed by the Perun Address interface.
func (a *Address) Cmp(b pwallet.Address) int {
	return bytes.Compare(a.Bytes(), AsAddr(b).Bytes())
}

// IsAddr returns whether a Perun Address has the expected Address type.
func IsAddr(addr pwallet.Address) bool {
	_, ok := addr.(*Address)
	return ok
}

// AsAddr returns a Perun Address as Address. Panics if the conversion failed.
func AsAddr(addr pwallet.Address) *Address {
	return addr.(*Address)
}
