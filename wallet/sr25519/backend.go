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
	"io"

	"github.com/ChainSafe/go-schnorrkel"
	pkgio "perun.network/go-perun/pkg/io"
	pwallet "perun.network/go-perun/wallet"

	"github.com/perun-network/perun-polkadot-backend/pkg/substrate"
)

type (
	// Backend implements the Backend interface.
	Backend struct{}

	signature = [SignatureLen]byte
)

const (
	// SignatureLen is the constant length of a signature in byte.
	SignatureLen = 64
)

// encodeAddress encodes an address into the writer.
func encodeAddress(a *Address, w io.Writer) error {
	return pkgio.Encode(w, a.Bytes())
}

// DecodeAddress decodes an address from the reader.
func (*Backend) DecodeAddress(r io.Reader) (pwallet.Address, error) {
	addr := new(Address)
	return addr, addr.Decode(r)
}

// encodeSig encodes a signature into byte slice.
func encodeSig(s *schnorrkel.Signature) []byte {
	// Convert to ensure that the length of schnorrkel sigs did not change.
	bytes := signature(s.Encode()) // nolint: unconvert
	return bytes[:]
}

// DecodeSig decodes a signature from the reader.
func (*Backend) DecodeSig(r io.Reader) (pwallet.Sig, error) {
	var _sig signature
	sig := _sig[:]
	if err := pkgio.Decode(r, &sig); err != nil {
		return nil, err
	}
	// Decode the sig with schnorrkel for error checking.
	return sig, new(schnorrkel.Signature).Decode(_sig)
}

// VerifySignature verifies that the signature was created by the address
// on the passed data. Panics on wrong address type.
func (*Backend) VerifySignature(msg []byte, s pwallet.Sig, a pwallet.Address) (bool, error) {
	var _s signature
	copy(_s[:], s)
	sig := new(schnorrkel.Signature)
	if err := sig.Decode(_s); err != nil {
		return false, err
	}
	context := schnorrkel.NewSigningContext(substrate.SignaturePrefix, msg)
	return AsAddr(a).pk.Verify(sig, context), nil
}
