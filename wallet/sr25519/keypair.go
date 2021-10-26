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
	"github.com/ChainSafe/go-schnorrkel"
	gsrpcsig "github.com/centrifuge/go-substrate-rpc-client/v3/signature"
	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/perun-network/perun-polkadot-backend/pkg/substrate"
)

// keyPair is used by the wallet to store secret keys.
type keyPair struct {
	// Mini SK is needed for creating types.KeyringPair
	msk *schnorrkel.MiniSecretKey
	// SK is needed for signing
	sk *schnorrkel.SecretKey
	// PK needed for accountId and address
	pk *Address
}

// makeKeyPair returns a new keyPair.
func makeKeyPair(msk *schnorrkel.MiniSecretKey) keyPair {
	return keyPair{msk, msk.ExpandEd25519(), NewAddressFromPK(msk.Public())}
}

// keyRing returns the receiver as gsrpc.KeyringPair.
func (kp *keyPair) keyRing(net substrate.NetworkID) (gsrpcsig.KeyringPair, error) {
	msk := kp.msk.Encode()
	pair, err := gsrpcsig.KeyringPairFromSecret(hexutil.Encode(msk[:]), uint8(net))
	return pair, err
}
