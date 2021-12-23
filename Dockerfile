FROM golang:1.17-alpine AS builder

# We disable CGO here due to:
# 1) https://github.com/golang/go/issues/28065 that prevents 'go test' from running inside an Alpine container
# 2) https://stackoverflow.com/questions/36279253/go-compiled-binary-wont-run-in-an-alpine-docker-container-on-ubuntu-host which
#       which prevents from just switching to the Buster build image
# Sadly, this is slower: https://stackoverflow.com/questions/47714278/why-is-compiling-with-cgo-enabled-0-slower
ENV CGO_ENABLED=0

WORKDIR /build
# Copy and download dependencies using go mod
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy the code into the container
COPY . .

RUN go test ./...

# Build the application
RUN GOOS=linux go build -o tx-fuzz.bin ./cmd/livefuzzer/main.go

# ============= Execution Stage ================
FROM alpine:3.12 AS execution

WORKDIR /run

# Copy the code into the container
COPY --from=builder /build/tx-fuzz.bin .

ENTRYPOINT ["./tx-fuzz.bin"]