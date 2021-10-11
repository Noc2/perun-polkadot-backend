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
	"fmt"
	"strings"

	"github.com/centrifuge/go-substrate-rpc-client/v3/types"
	"github.com/pkg/errors"
)

var (
	// ErrUnknownError the error could not be decoded.
	ErrUnknownError = errors.New("unknown dispatch error type")
	// ErrCallFailed an extrinsic returned an error.
	ErrCallFailed = errors.New("call failed")
)

// DecodeError decodes an error into a human-readable form.
// Returns either ErrUnknownError or ErrCallFailed.
func DecodeError(meta *types.Metadata, err types.DispatchError) error {
	metaV, ok := Meta(meta)
	if !ok {
		return errors.Wrap(ErrUnknownError, "wrong meta data version")
	}
	if !err.HasModule {
		return errors.Wrap(ErrUnknownError, "no module in error")
	}
	if int(err.Module) >= len(metaV.Modules) {
		return errors.Wrap(ErrUnknownError, "module index out of range")
	}
	module := metaV.Modules[err.Module]
	if int(err.Error) >= len(module.Errors) {
		return errors.Wrap(ErrUnknownError, "error index out of range")
	}
	e := module.Errors[err.Error]
	return errors.Wrap(ErrCallFailed, formatErrorMeta(e))
}

// formatErrorMeta formats an ErrorMetaDataV8 into a human-readable form.
func formatErrorMeta(err types.ErrorMetadataV8) string {
	docs := make([]string, len(err.Documentation))
	for i, doc := range err.Documentation {
		docs[i] = string(doc)
	}

	return fmt.Sprintf("%s:%s", err.Name, strings.Join(docs, ", "))
}
