package main

//go:generate ./solc-static-linux-8-7 --overwrite --bin --abi -o ./ contract.sol
//go:generate abigen --pkg main --type Selfdestructer --abi Selfdestructer.abi --bin Selfdestructer.bin --out contract.go
