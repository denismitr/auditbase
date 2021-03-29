FROM golang:latest as builder

ENV CGO_ENABLED 0

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download
RUN go mod verify

COPY cmd/consumer ./cmd/consumer
COPY internal/ ./internal
COPY .env ./

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o consumer ./cmd/consumer

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /source

COPY --from=builder /app/consumer .
COPY --from=builder /app/.env .

RUN chmod +x ./consumer

CMD ["./consumer"]
