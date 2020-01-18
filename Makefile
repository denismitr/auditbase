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

up: vars
	docker-compose -f docker-compose-dev.yml up --build --force-recreate -d

down:
	docker-compose -f docker-compose-dev.yml down --remove-orphans

clean:
	docker-compose -f docker-compose-dev.yml down --remove-orphans
	docker-compose -f docker-compose-dev.yml rm -svf
	docker-compose -f docker-compose-test.yml down --remove-orphans
	docker-compose -f docker-compose-test.yml rm -svf

test:
	docker-compose -f docker-compose-test.yml up --build --force-recreate

docker-remove:
	docker rm --force `docker ps -a -q` || true
	docker rmi --force `docker images -q` || true

docker-kill:
	docker kill `docker ps -q` || true

