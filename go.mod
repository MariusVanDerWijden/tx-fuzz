module github.com/MariusVanDerWijden/tx-fuzz

go 1.20

require (
<<<<<<< HEAD
	github.com/MariusVanDerWijden/FuzzyVM v0.0.0-20230606141057-24202385e744
	github.com/ethereum/go-ethereum v1.12.0
	github.com/holiman/goevmlab v0.0.0-20230602194133-da4e4913b799
	github.com/urfave/cli/v2 v2.25.1
)

require (
	github.com/DataDog/zstd v1.5.5 // indirect
	github.com/VictoriaMetrics/fastcache v1.12.1 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/btcsuite/btcd/btcec/v2 v2.3.2 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/cockroachdb/errors v1.10.0 // indirect
	github.com/cockroachdb/logtags v0.0.0-20230118201751-21c54148d20b // indirect
	github.com/cockroachdb/pebble v0.0.0-20230614191204-17e0ab3c2e0e // indirect
	github.com/cockroachdb/redact v1.1.5 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/deckarep/golang-set/v2 v2.3.0 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.2.0 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/getsentry/sentry-go v0.21.0 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/gofrs/flock v0.8.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/golang/snappy v0.0.5-0.20220116011046-fa5810519dcb // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/holiman/bloomfilter/v2 v2.0.3 // indirect
	github.com/holiman/uint256 v1.2.2 // indirect
	github.com/huin/goupnp v1.2.0 // indirect
	github.com/jackpal/go-nat-pmp v1.0.2 // indirect
	github.com/klauspost/compress v1.16.6 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/mattn/go-runewidth v0.0.14 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/olekukonko/tablewriter v0.0.5 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_golang v1.15.1 // indirect
	github.com/prometheus/client_model v0.4.0 // indirect
	github.com/prometheus/common v0.44.0 // indirect
	github.com/prometheus/procfs v0.10.1 // indirect
	github.com/rivo/uniseg v0.4.4 // indirect
	github.com/rogpeppe/go-internal v1.10.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/shirou/gopsutil v3.21.11+incompatible // indirect
	github.com/syndtr/goleveldb v1.0.1-0.20220721030215-126854af5e6d // indirect
	github.com/tklauser/go-sysconf v0.3.11 // indirect
	github.com/tklauser/numcpus v0.6.1 // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	github.com/yusufpapurcu/wmi v1.2.3 // indirect
	golang.org/x/crypto v0.10.0 // indirect
	golang.org/x/exp v0.0.0-20230522175609-2e198f4a06a1 // indirect
	golang.org/x/sync v0.3.0 // indirect
	golang.org/x/sys v0.9.0 // indirect
	golang.org/x/text v0.10.0 // indirect
	google.golang.org/protobuf v1.30.0 // indirect
	gopkg.in/natefinch/npipe.v2 v2.0.0-20160621034901-c1b8fa8bdcce // indirect
=======
	github.com/MariusVanDerWijden/FuzzyVM v0.0.0-20220901111237-4348e62e228d
	github.com/VictoriaMetrics/fastcache v1.12.0 // indirect
<<<<<<< HEAD
	github.com/ethereum/go-ethereum v1.10.26
	github.com/holiman/goevmlab v0.0.0-20220902091028-02faf03e18e4
	github.com/holiman/uint256 v1.2.2-0.20230321075855-87b91420868c
	github.com/mattn/go-runewidth v0.0.14 // indirect
	github.com/protolambda/ztyp v0.2.2
	github.com/rivo/uniseg v0.4.2 // indirect
<<<<<<< HEAD
	github.com/urfave/cli/v2 v2.17.2-0.20221006022127-8f469abc00aa
>>>>>>> a7a359f (cmd/4844: added stuff)
=======
	github.com/urfave/cli/v2 v2.24.1
>>>>>>> 948ebcf (cmd/4844: updated to newest spec)
=======
	github.com/ethereum/go-ethereum v1.11.6-0.20230404163452-2adce0b06640
	github.com/holiman/goevmlab v0.0.0-20230602194133-da4e4913b799
	github.com/holiman/uint256 v1.2.2
	github.com/urfave/cli/v2 v2.25.1
>>>>>>> bd263ea (go.mod: updates, fix encodeBlobs)
)

replace github.com/ethereum/go-ethereum => github.com/mariusvanderwijden/go-ethereum v1.8.22-0.20230707144623-e03b5add7133
