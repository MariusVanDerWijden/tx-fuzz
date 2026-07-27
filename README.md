# TX-Fuzz

TX-Fuzz is a package containing helpful functions to create random transactions. 
It can be used to easily access fuzzed transactions from within other programs.

## Usage

```
cd cmd/livefuzzer
go build
```

Run an execution layer client such as [Geth][1] locally in a standalone bash window.
Tx-fuzz sends transactions to port `8545` by default.

```
geth --http --http.port 8545
```

Run livefuzzer.

```
./livefuzzer spam
```

## Transaction types

`spam` rotates over every transaction type that shares the plain transaction
pool: legacy (`0x00`), access list (`0x01`), dynamic fee (`0x02`) and set code
(`0x04`), as contract creations and as calls, with and without a node-built
access list.

Blob transactions (`0x03`) have a command of their own, `blobs`, because a node
keeps them in a separate pool and refuses to hold both a blob and a non-blob
transaction for the same account. Sidecars carry EIP-7594 cell proofs by
default; use `--blob-sidecar-version 0` to target a node from before Osaka.

`pectra` spams set code transactions specifically, with authorization lists that
mix universal (chain id 0) authorizations, wrong chains and nonces,
self-delegation and delegation clearing.

## Advanced usage
You can optionally specify a seed parameter or a secret key to use as a faucet

```
./livefuzzer spam --seed <seed> --sk <SK>
```

You can set the RPC to use with `--rpc <RPC>`.

`spam` and `blobs` reproduce their transactions from the seed, so a run can be
replayed. Two things are outside the seed's control: what the node contributes
(nonces, gas price and the access lists it builds), and, in `pectra` or with
`--corpus`, the choice of authorizing account and of corpus element, which still
come from global randomness (see `FIXES.md`).

[1]: https://github.com/ethereum/go-ethereum
