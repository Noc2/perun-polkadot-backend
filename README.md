<h1 align="center"><br>
    <a href="https://perun.network/"><img src=".assets/go-perun.png" alt="Perun" width="196"></a>
<br></h1>

<h2 align="center">Perun Polkadot: Backend</h2>

<p align="center">
  <a href="https://www.apache.org/licenses/LICENSE-2.0.txt"><img src="https://img.shields.io/badge/license-Apache%202-blue" alt="License: Apache 2.0"></a>
  <a href="https://github.com/perun-network/perun-polkadot-backend/actions/workflows/rust.yml"><img src="https://github.com/perun-network/perun-polkadot-backend/actions/workflows/go.yml/badge.svg?branch=main" alt="CI status"></a>
</p>

This project implements a [Substrate] backend for our [go-perun] framework. 
It enables *go-perun* applications to create state-channels on [Polkadot] by connecting to the [Perun Pallet].

## Project structure
* `channel/` channel interface implementations
  * `pallet/` [Perun Pallet] specific code
* `wallet/` wallet interface implementations
  * `sr25519/` *Schnorrkel-Ristretto Ed25519* wallet
* `pkg/` 3rd-party helpers
* `client/` helper functions for setting up a *go-perun* client
* `wallet/sr25519/test/accounts.json` config for testing

## Development Setup
If you want to locally develop with this project:

1. Clone the repo:
```sh
git clone https://github.com/perun-network/perun-polkadot-backend
cd perun-polkadot-backend
```

2. Start a local substrate chain by following the instructions in our [polkadot node] repo.

3. Run all tests:  
```sh
go test -p 1 ./...
```
This can take while but should eventually finish successfully. The long testing time results from the block-time of the node, which is set to one second.  
The `-p 1` flag is important, since the tests otherwise are started in parallel and mess up
the account nonce.

## Demo

We prepared a [demo CLI node] to play around with payment channels on substrate chains.  
Highly recommended you to check it out 👍

## Funding

The development of this project is supported by the [Web3 Foundation] through the [Open Grants Program].  
The development of the go-perun library is supported by the German Ministry of Education and Science (BMBF) through a Startup Secure grant.

## Security Disclaimer

This software is still under development.
The authors take no responsibility for any loss of digital assets or other damage caused by the use of it.

## Copyright

Copyright 2021 PolyCrypt GmbH.  
Use of the source code is governed by the Apache 2.0 license that can be found in the [LICENSE file](LICENSE).

<!--- Links -->

[Polkadot]: https://polkadot.network/
[Substrate]: https://substrate.dev/
[go-perun]: https://github.com/hyperledger-labs/go-perun
[Perun Pallet]: https://github.com/perun-network/perun-polkadot-pallet
[frontend template]: https://github.com/substrate-developer-hub/substrate-front-end-template
[demo CLI node]: https://github.com/perun-network/perun-polkadot-demo
[polkadot node]: https://github.com/perun-network/perun-polkadot-node#usage

[Open Grant]: https://github.com/perun-network/Open-Grants-Program/blob/master/applications/perun_channels.md#w3f-open-grant-proposal
[Web3 Foundation]: https://web3.foundation/about/
[Open Grants Program]: https://github.com/w3f/Open-Grants-Program#open-grants-program-
