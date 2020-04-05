FROM golang:latest

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY consumer/ ./consumer
COPY queue/ ./queue
COPY model/ ./model
COPY flow/ ./flow
COPY utils/ ./utils
COPY rest/ ./rest
COPY db/ ./db
COPY test/ ./test
COPY .env ./

CMD go test ./flow ./rest ./model

