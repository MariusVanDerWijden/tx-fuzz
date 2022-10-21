module github.com/MariusVanDerWijden/tx-fuzz

go 1.16

require (
	github.com/MariusVanDerWijden/FuzzyVM v0.0.0-20221202121132-bd37e8fb1d0d
	github.com/ethereum/go-ethereum v1.10.26
	github.com/holiman/goevmlab v0.0.0-20221207202144-89074274e1b7
	github.com/urfave/cli/v2 v2.23.7 // indirect
)

replace github.com/ethereum/go-ethereum => github.com/mdehoog/go-ethereum v1.10.19-0.20221008022208-0aa8f1ddceb2
