REST_PORT ?= 5000
RECEIVER_PORT ?= 8888
PERCONA_PORT ?= 3306
GO_VERSION := 1.15.2
GO := go
GO_TEST := $(GO) test -race
GO_BUILD := $(GO) build
GO_COVER := $(GO) tool cover
COVER_OUT := ./.cover/c.out
AUDITBASE_VERSION := $(shell git rev-parse --short HEAD || echo "GitNotFound")

RECEIVER_ENDPOINT = http://localhost:$(RECEIVER_PORT)/api/v1/actions

BACK_OFFICE_BINARY := ./cmd/backoffice/backoffice
BACK_OFFICE_MAIN := ./cmd/backoffice/backoffice.go
CONSUMER_BINARY := ./cmd/consumer/consumer
CONSUMER_MAIN := ./cmd/consumer/consumer.go
RECEIVER_BINARY := ./cmd/receiver/receiver
RECEIVER_MAIN := ./cmd/receiver/receiver.go

# Version is a git tag name for current SHA commit hash or this hash if tag is not presented
# APP_VERSION ?= $$(git describe --exact-match --tags $(git log -n1 --pretty='%h') 2> /dev/null || \
# 				  git log -n1 --pretty='commit_sha:%h')

vars:
	@echo APP_VERSION=${APP_VERSION}
	@echo REST_PORT=${REST_PORT}
	@echo PERCONA_PORT=${PERCONA_PORT}
	@echo AUDITBASE_VERSION

.PHONY: test clean mock wrk debug recompile up build

up: vars
	docker-compose -f docker-compose-dev.yml up -d --build

recreate: vars
	docker-compose -f docker-compose-dev.yml up -d --build --renew-anon-volumes --force-recreate

down:
	docker-compose -f docker-compose-dev.yml down --remove-orphans

recompile:
	docker-compose -f docker-compose-dev.yml up -d --build --force-recreate auditbase_receiver auditbase_backoffice auditbase_consumer auditbase_errors_consumer

clean:
	docker-compose -f docker-compose-dev.yml rm --force --stop -v

build:
	@echo Building the backoffice
	$(GO_BUILD) -o $(BACK_OFFICE_BINARY) $(BACK_OFFICE_MAIN)
	@echo Building the receiver
	$(GO_BUILD) -o $(RECEIVER_BINARY) $(RECEIVER_MAIN)
	@echo Building the consumer
	$(GO_BUILD) -o $(CONSUMER_BINARY) $(CONSUMER_MAIN)

local/test:
	go test ./db/mysql ./model ./rest ./flow

docker/test:
	docker-compose -f docker-compose-test.yml up --build --force-recreate

seed:
	go run ./cmd/seed --endpoint=$(RECEIVER_ENDPOINT)

integration_test:
	go test ./test/integration/...

docker/debug: vars
	docker-compose -f docker-compose-debug.yml up -d --build --force-recreate

wrk/run:
	wrk -c50 -t3 -d100s -s ./test/lua/events.lua http://127.0.0.1:8888

wrk/debug:
	wrk -c20 -t2 -d20s --rate=30 -s ./test/lua/events.lua http://127.0.0.1:8888

docker-remove:
	docker rm --force `docker ps -a -q` || true
	docker rmi --force `docker images -q` || true

docker-kill:
	docker kill `docker ps -q` || true

