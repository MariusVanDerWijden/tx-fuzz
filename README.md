# TX-Fuzz

TX-Fuzz is a package containing helpful functions to create random transactions. 
It can be used to easily access fuzzed transactions from within other programs.

## Usage

```
cd cmd/livefuzzer
go build
./livefuzzer spam
```

Tx-fuzz allows for an optional seed parameter to get reproducible fuzz transactions