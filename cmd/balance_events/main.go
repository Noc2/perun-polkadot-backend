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

package main

import (
	"math/big"

	"github.com/centrifuge/go-substrate-rpc-client/v3/config"
	"github.com/centrifuge/go-substrate-rpc-client/v3/types"
	"github.com/sirupsen/logrus"
	"perun.network/go-perun/log"
	plogrus "perun.network/go-perun/log/logrus"

	pkgsr25519 "github.com/perun-network/perun-polkadot-backend/pkg/sr25519"
	"github.com/perun-network/perun-polkadot-backend/pkg/substrate"
	"github.com/perun-network/perun-polkadot-backend/wallet/sr25519"
)

func main() {
	plogrus.Set(logrus.InfoLevel, &logrus.TextFormatter{ForceColors: true})
	api, err := substrate.NewApi(config.Default().RPCURL, 42, 100)
	noErr(err)
	log.Infof("PlankPerDot = %v", substrate.PlankPerDot)

	go logBalEvents("alice", "0xd43593c715fdd31c61141abd04a99fd6822c8558854ccde39a5684e7a56da27d", api)
	go logBalEvents("bob  ", "0x8eaf04151687736326c9fea17e25fc5287613693c912909cb226aa4794f26a48", api)

	select {}
}

func logBalEvents(name, hexAddr string, api *substrate.Api) {
	pk, err := pkgsr25519.NewPKFromHex(hexAddr)
	noErr(err)
	addr := sr25519.NewAddressFromPK(pk).AccountId()

	source, err := substrate.NewEventSource(api, 1, substrate.SystemAccountKey(addr))
	noErr(err)
	defer source.Close()

	info, err := api.AccountInfo(addr)
	noErr(err)
	oldBal := substrate.NewDotFromPlank(info.Free.Int)
	log.Printf("%s: %v\n", name, oldBal)

	for {
		var event types.EventRecordsRaw
		select {
		case event = <-source.Events():
		case err := <-source.Err():
			noErr(err)
		}
		err := types.DecodeFromBytes([]byte(event), &info)
		noErr(err)
		newBal := substrate.NewDotFromPlank(info.Free.Int)

		// Calculate the delta
		var change = new(big.Int).Sub(newBal.Plank(), oldBal.Plank())
		dot := substrate.NewDotFromPlank(change)
		if sign := change.Sign(); sign > 0 {
			log.Printf("%s got  %v and has now %v\n", name, dot, newBal)
		} else if sign < 0 {
			log.Printf("%s lost %v and has now %v\n", name, dot.Abs(), newBal)
		}
		oldBal = newBal
	}
}

func noErr(err error) {
	if err != nil {
		panic(err)
	}
}
