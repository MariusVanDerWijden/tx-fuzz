module github.com/MariusVanDerWijden/tx-fuzz

go 1.16

require (
	github.com/MariusVanDerWijden/FuzzyVM v0.0.0-20220901111237-4348e62e228d
	github.com/VictoriaMetrics/fastcache v1.12.0 // indirect
	github.com/ethereum/go-ethereum v1.10.26
	github.com/holiman/goevmlab v0.0.0-20220902091028-02faf03e18e4
	github.com/holiman/uint256 v1.2.1
	github.com/mattn/go-runewidth v0.0.14 // indirect
	github.com/protolambda/ztyp v0.2.2
	github.com/rivo/uniseg v0.4.2 // indirect
	github.com/urfave/cli/v2 v2.17.2-0.20221006022127-8f469abc00aa
)

replace github.com/ethereum/go-ethereum => github.com/mdehoog/go-ethereum v1.10.19-0.20230503182922-ac64c4400e35

replace github.com/protolambda/go-kzg => github.com/Inphi/go-kzg v0.0.0-20220819034031-381084440411

replace github.com/kilic/bls12-381 => github.com/Inphi/bls12-381 v0.0.0-20220819032644-3ae7bcd28efc
