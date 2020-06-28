FROM golang:latest as builder

RUN mkdir -p /tmp/debug/receiver

WORKDIR /source

COPY go.mod go.sum ./

RUN go mod download

COPY cmd/ ./cmd
COPY rest/ ./rest
COPY queue/ ./queue
COPY model/ ./model
COPY utils/ ./utils
COPY persister/ ./persister
COPY flow/ ./flow
COPY db/ ./db
COPY cache/ ./cache
COPY .env ./

EXPOSE 8888

ENTRYPOINT ["go", "run", "-race", "/source/cmd/receiver/receiver.go"]
