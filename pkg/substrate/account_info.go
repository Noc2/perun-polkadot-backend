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

import "github.com/centrifuge/go-substrate-rpc-client/v3/types"

// AccountInfo replaces substrate.AccountInfo since it is outdated.
// This is advised by the GSRPC team.
type AccountInfo struct {
	Nonce       types.U32
	Consumers   types.U32
	Providers   types.U32
	Sufficients types.U32

	Free       types.U128
	Reserved   types.U128
	MiscFrozen types.U128
	FreeFrozen types.U128
}
