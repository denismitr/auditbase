FROM golang:latest

RUN mkdir -p /tmp/debug/consumer

WORKDIR /source

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ ./cmd
COPY internal/ ./internal
COPY .env ./

ENTRYPOINT ["go", "run", "-race", "/source/cmd/consumer/consumer.go"]
