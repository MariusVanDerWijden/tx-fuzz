FROM golang:latest AS builder

WORKDIR /build
# Copy and download dependencies using go mod
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy the code into the container
COPY . .

RUN go test ./...
# Build the application
RUN cd cmd/livefuzzer && GOOS=linux go build -o tx-fuzz.bin .

ENTRYPOINT ["/build/cmd/livefuzzer/tx-fuzz.bin"]
CMD ["/build/cmd/livefuzzer/tx-fuzz.bin"]