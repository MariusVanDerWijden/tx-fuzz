module github.com/MariusVanDerWijden/tx-fuzz

go 1.16

require (
	github.com/MariusVanDerWijden/FuzzyVM v0.0.0-20220901111237-4348e62e228d
	github.com/VictoriaMetrics/fastcache v1.12.0 // indirect
	github.com/ethereum/go-ethereum v1.11.6-0.20230404163452-2adce0b06640
	github.com/holiman/goevmlab v0.0.0-20230602194133-da4e4913b799
	github.com/holiman/uint256 v1.2.2
	github.com/urfave/cli/v2 v2.25.1
)

replace github.com/ethereum/go-ethereum => github.com/mariusvanderwijden/go-ethereum v1.8.22-0.20230626175218-2d586a9714d9
