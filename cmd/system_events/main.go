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
	"github.com/centrifuge/go-substrate-rpc-client/v3/config"
	"github.com/centrifuge/go-substrate-rpc-client/v3/types"
	"github.com/sirupsen/logrus"
	"perun.network/go-perun/log"
	plogrus "perun.network/go-perun/log/logrus"

	"github.com/perun-network/perun-polkadot-backend/channel"
	"github.com/perun-network/perun-polkadot-backend/pkg/substrate"
)

func main() {
	plogrus.Set(logrus.InfoLevel, &logrus.TextFormatter{ForceColors: true})
	api, err := substrate.NewApi(config.Default().RPCURL, 42, 100)
	noErr(err)

	logSystemEvents(api, log.Get())
}

func logSystemEvents(api *substrate.Api, log logrus.StdLogger) {
	meta := api.Metadata()
	// Subscribe to system events via storage
	key, err := types.CreateStorageKey(meta, "System", "Events", nil, nil)
	noErr(err)

	sub, err := api.Subscribe(key)
	noErr(err)
	defer sub.Unsubscribe()

	for event := range sub.Chan() {
		for _, chng := range event.Changes {
			if !types.Eq(chng.StorageKey, key) || !chng.HasStorageData {
				// skip, we are only interested in events with content
				continue
			}

			events := channel.EventRecords{}
			err = types.EventRecordsRaw(chng.StorageData).DecodeEventRecords(meta, &events)
			noErr(err)

			for _, e := range events.Balances_Endowed {
				log.Printf("Balances:Endowed:: %#x…, %v\n", e.Who[:4], e.Balance)
			}
			for _, e := range events.Balances_DustLost {
				log.Printf("Balances:DustLost:: %#x…, %v\n", e.Who[:4], e.Balance)
			}
			for _, e := range events.Balances_Transfer {
				log.Printf("Balances:Transfer:: %#x… => %#x…: %d\n", e.From[:4], e.To[:4], e.Value)
			}
			for _, e := range events.Balances_BalanceSet {
				log.Printf("Balances:BalanceSet:: %v, %v, %v\n", e.Who, e.Free, e.Reserved)
			}
			for _, e := range events.Balances_Deposit {
				log.Printf("Balances:Deposit:: %v, %v\n", e.Who, e.Balance)
			}
			for _, e := range events.System_ExtrinsicSuccess {
				if e.Phase.AsApplyExtrinsic != 0 {
					log.Printf("System:ExtrinsicSuccess::\n")
				}
			}
			for _, e := range events.System_ExtrinsicFailed {
				log.Printf("System:ExtrinsicFailed:: %s\n", substrate.DecodeError(meta, e.DispatchError))
			}
			for _, e := range events.System_NewAccount {
				log.Printf("System:NewAccount:: %#x…\n", e.Who[:4])
			}
			for _, e := range events.System_KilledAccount {
				log.Printf("System:KilledAccount:: %#x…\n", e.Who[:4])
			}
		}
	}
}

func noErr(err error) {
	if err != nil {
		panic(err)
	}
}
