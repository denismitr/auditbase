FROM golang:latest as builder

RUN mkdir -p /tmp/debug/receiver

WORKDIR /source

COPY go.mod go.sum ./

RUN go mod download

COPY cmd/ ./cmd
COPY internal/ ./internal
COPY .env ./

EXPOSE 8888

ENTRYPOINT ["go", "run", "-race", "/source/cmd/receiver/receiver.go"]
