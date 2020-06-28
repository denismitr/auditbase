FROM golang:latest as builder

WORKDIR /app

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

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o receiver ./cmd/receiver

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /source

COPY --from=builder /app/receiver .
COPY --from=builder /app/.env .

EXPOSE 8888

RUN chmod +x ./receiver

CMD ["./receiver"]
