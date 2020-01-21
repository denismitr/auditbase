FROM golang:latest as builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY cmd/consumer ./cmd/consumer
COPY cmd/healthcheck ./cmd/healthcheck
COPY consumer/ ./consumer
COPY queue/ ./queue
COPY model/ ./model
COPY flow/ ./flow
COPY utils/ ./utils
COPY sql/ ./sql
COPY .env ./

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o consumer ./cmd/consumer
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o healthcheck ./cmd/healthcheck

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /source

COPY --from=builder /app/consumer .
COPY --from=builder /app/healthcheck .
COPY --from=builder /app/.env .

RUN chmod +x ./consumer
RUN chmod +x ./healthcheck

HEALTHCHECK --interval=1s --timeout=1s --start-period=60s --retries=3 CMD [ "/healthcheck" ] || exit 1

ENTRYPOINT ["./consumer"]
