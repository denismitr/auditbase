REST_PORT ?= 5000
PERCONA_PORT ?= 3306
GO_VERSION ?= 1.13.3
AUDITBASE_VERSION ?= $(shell git rev-parse --short HEAD || echo "GitNotFound")

# Version is a git tag name for current SHA commit hash or this hash if tag is not presented
# APP_VERSION ?= $$(git describe --exact-match --tags $(git log -n1 --pretty='%h') 2> /dev/null || \
# 				  git log -n1 --pretty='commit_sha:%h')

vars:
	@echo APP_VERSION=${APP_VERSION}
	@echo REST_PORT=${REST_PORT}
	@echo PERCONA_PORT=${PERCONA_PORT}

.PHONY: test clean mock wrk debug recompile up

up: vars
	docker-compose -f docker-compose-dev.yml up -d --build --force-recreate

down:
	docker-compose -f docker-compose-dev.yml down --remove-orphans

recompile:
	docker-compose -f docker-compose-dev.yml up -d --build --force-recreate auditbase_receiver auditbase_backoffice auditbase_consumer auditbase_errors_consumer

clean:
	docker-compose -f docker-compose-dev.yml down --remove-orphans
	docker-compose -f docker-compose-dev.yml rm -svf
	docker-compose -f docker-compose-debug.yml down --remove-orphans
	docker-compose -f docker-compose-debug.yml rm -svf
	docker-compose -f docker-compose-test.yml down --remove-orphans
	docker-compose -f docker-compose-test.yml rm -svf

mock:
	mockgen -source flow/flow.go -destination ./test/mock_flow/flow.go
	mockgen -source flow/event.go -destination ./test/mock_flow/event.go
	mockgen -source cache/cache.go -destination ./test/mock_cache/cache.go
	mockgen -source queue/queue.go -destination ./test/mock_queue/queue.go
	mockgen -source queue/message.go -destination ./test/mock_queue/message.go
	mockgen -source model/event.go -destination ./test/mock_model/event.go
	mockgen -source model/microservice.go -destination ./test/mock_model/microservice.go
	mockgen -source model/entity.go -destination ./test/mock_model/entity.go
	mockgen -source model/change.go -destination ./test/mock_model/change.go
	mockgen -source model/property.go -destination ./test/mock_model/property.go
	mockgen -source model/factory.go -destination ./test/mock_model/factory.go
	mockgen -source db/persister.go -destination ./test/mock_db/persister.go
	mockgen -source utils/clock/clock.go -destination ./test/mock_utils/mock_clock/clock.go
	mockgen -source utils/uuid/uuid.go -destination ./test/mock_utils/mock_uuid/uuid.go

test/local:
	go test ./db/mysql ./model ./rest

test:
	docker-compose -f docker-compose-test.yml up --build --force-recreate

integration_test:
	go test ./test/integration/...

debug: vars
	docker-compose -f docker-compose-debug.yml up --build --force-recreate

wrk/local:
	wrk -c50 -t3 -d100s -s ./test/lua/events.lua http://127.0.0.1:8888

wrk/debug:
	wrk -c20 -t2 -d20s -s ./test/lua/events.lua http://127.0.0.1:8888

docker-remove:
	docker rm --force `docker ps -a -q` || true
	docker rmi --force `docker images -q` || true

docker-kill:
	docker kill `docker ps -q` || true

