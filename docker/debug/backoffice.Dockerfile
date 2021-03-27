FROM golang:latest as builder

RUN mkdir -p /tmp/debug/backoffice

WORKDIR /source

COPY go.mod go.sum ./

RUN go mod download

COPY cmd/ ./cmd
COPY internal/ ./internal
COPY .env ./

EXPOSE 8889

ENTRYPOINT ["go", "run", "-race", "/source/cmd/backoffice/backoffice.go"]
