module github.com/MariusVanDerWijden/tx-fuzz

go 1.16

require (
	github.com/MariusVanDerWijden/FuzzyVM v0.0.0-20220901111237-4348e62e228d
	github.com/VictoriaMetrics/fastcache v1.12.0 // indirect
	github.com/ethereum/go-ethereum v1.10.26
	github.com/holiman/goevmlab v0.0.0-20220902091028-02faf03e18e4
	github.com/holiman/uint256 v1.2.2-0.20230321075855-87b91420868c
	github.com/mattn/go-runewidth v0.0.14 // indirect
	github.com/protolambda/ztyp v0.2.2
	github.com/rivo/uniseg v0.4.2 // indirect
	github.com/urfave/cli/v2 v2.24.1
)

replace github.com/ethereum/go-ethereum => github.com/mariusvanderwijden/go-ethereum v1.8.22-0.20230626175218-2d586a9714d9
