module github.com/MariusVanDerWijden/tx-fuzz

go 1.16

require (
	github.com/MariusVanDerWijden/FuzzyVM v0.0.0-20210904205340-da82a0d3e27a
	github.com/btcsuite/btcd/btcec/v2 v2.3.2 // indirect
	github.com/ethereum/go-ethereum v1.11.6
	github.com/holiman/goevmlab v0.0.0-20210406174504-acc14986d1a1
	github.com/holiman/uint256 v1.2.0
	github.com/protolambda/ztyp v0.2.1
)

replace github.com/ethereum/go-ethereum => github.com/mdehoog/go-ethereum v1.10.19-0.20221008022208-0aa8f1ddceb2
