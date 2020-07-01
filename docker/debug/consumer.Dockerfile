FROM golang:latest

RUN mkdir -p /tmp/debug/consumer

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

ENTRYPOINT ["go", "run", "-race", "/source/cmd/consumer/consumer.go"]
