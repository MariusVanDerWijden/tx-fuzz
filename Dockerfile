FROM golang:1.23.8-alpine3.20 AS builder

WORKDIR /build
# Copy and download dependencies using go mod
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy the code into the container
COPY . .

# Build the application
RUN cd cmd/livefuzzer && CGO_ENABLED=0 GOOS=linux go build -o tx-fuzz.bin .

FROM alpine:latest

COPY --from=builder /build/cmd/livefuzzer/tx-fuzz.bin /tx-fuzz.bin

ENTRYPOINT ["/tx-fuzz.bin"]
