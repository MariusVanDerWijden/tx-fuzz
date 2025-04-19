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

Run livefuzzer and pass a funded account secret key.

```
./livefuzzer spam --sk <SK>
```

Example:

```
./livefuzzer spam --sk 0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80
```

Tx-fuzz allows for an optional seed parameter to get reproducible fuzz transactions

## Advanced usage
You can optionally specify a seed parameter or a secret key to use as a faucet

```
./livefuzzer spam --seed <seed> --sk <SK>
```

You can set the RPC to use with `--rpc <RPC>`.
