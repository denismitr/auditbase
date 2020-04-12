FROM golang:latest

RUN mkdir -p /tmp/debug/consumer

WORKDIR /source

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/consumer ./cmd/consumer
COPY cmd/healthcheck ./cmd/healthcheck
COPY consumer/ ./consumer
COPY queue/ ./queue
COPY model/ ./model
COPY flow/ ./flow
COPY utils/ ./utils
COPY db/ ./db
COPY .env ./

ENTRYPOINT ["go", "run", "/source/cmd/consumer/consumer.go"]
