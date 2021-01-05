NAME := shure-monitoring-service
OWNER := byuoitav
PKG := github.com/${OWNER}/${NAME}
BUILD_PKG1 := ${PKG}/cmd/central
BUILD_PKG2 := ${PKG}/cmd/local
NAME_PKG1 := ${NAME}-central
NAME_PKG2 := ${NAME}-local
DOCKER_URL := docker.pkg.github.com
DOCKER_PKG := ${DOCKER_URL}/${OWNER}/${NAME}

# version:
# use the git tag, if this commit
# doesn't have a tag, use the git hash
COMMIT_HASH := $(shell git rev-parse --short HEAD)
TAG := $(shell git rev-parse --short HEAD)
ifneq ($(shell git describe --exact-match --tags HEAD 2> /dev/null),)
	TAG = $(shell git describe --exact-match --tags HEAD)
endif

PRD_TAG_REGEX := "v[0-9]+\.[0-9]+\.[0-9]+"
DEV_TAG_REGEX := "v[0-9]+\.[0-9]+\.[0-9]+-.+"

# go stuff
PKG_LIST := $(shell go list ${PKG}/...)

.PHONY: all deps build test test-cov clean

all: clean build

test:
	@go test -v ${PKG_LIST}

test-cov:
	@go test -coverprofile=coverage.txt -covermode=atomic ${PKG_LIST}

lint:
	@golangci-lint run --tests=false

deps:
	@echo Downloading backend dependencies...
	@go mod download

build: deps
	@mkdir -p dist

	@echo
	@echo Building central backend for linux-amd64...
	@env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o ./dist/${NAME_PKG1}-linux-amd64 ${BUILD_PKG1}

	@echo
	@echo Building local backend for linux-amd64...
	@env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o ./dist/${NAME_PKG2}-linux-amd64 ${BUILD_PKG2}
	@echo
	@echo Building local backend for linux-arm...
	@env CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -v -o ./dist/${NAME_PKG2}-linux-arm ${BUILD_PKG2}

	@echo
	@echo Build output is located in ./dist/.

docker: clean build
ifeq (${COMMIT_HASH}, ${TAG})
	@echo Building central dev container with tag ${COMMIT_HASH}

	@echo Building container ${DOCKER_PKG}/${NAME_PKG1}-dev:${COMMIT_HASH}
	@docker build -f dockerfile --build-arg NAME=${NAME_PKG1}-linux-amd64 -t ${DOCKER_PKG}/${NAME_PKG1}-dev:${COMMIT_HASH} dist

	@echo Building local arm64 dev container with tag ${COMMIT_HASH}

	@echo Building container ${DOCKER_PKG}/${NAME_PKG2}-amd64-dev:${COMMIT_HASH}
	@docker build -f dockerfile --build-arg NAME=${NAME_PKG2}-linux-amd64 -t ${DOCKER_PKG}/${NAME_PKG2}-amd64-dev:${COMMIT_HASH} dist

	@echo Building local arm dev container with tag ${COMMIT_HASH}

	@echo Building container ${DOCKER_PKG}/${NAME_PKG2}-arm-dev:${COMMIT_HASH}
	@docker build -f dockerfile --build-arg NAME=${NAME_PKG2}-linux-arm -t ${DOCKER_PKG}/${NAME_PKG2}-arm-dev:${COMMIT_HASH} dist

else ifneq ($(shell echo ${TAG} | grep -x -E ${DEV_TAG_REGEX}),)
	@echo Building dev central container with tag ${TAG}

	@echo Building container ${DOCKER_PKG}/${NAME_PKG1}-dev:${TAG}
	@docker build -f dockerfile --build-arg NAME=${NAME_PKG1}-linux-amd64 -t ${DOCKER_PKG}/${NAME_PKG1}-dev:${TAG} dist

	@echo Building dev local amd64 container with tag ${TAG}

	@echo Building container ${DOCKER_PKG}/${NAME_PKG2}-amd64-dev:${TAG}
	@docker build -f dockerfile --build-arg NAME=${NAME_PKG2}-linux-amd64 -t ${DOCKER_PKG}/${NAME_PKG2}-amd64-dev:${TAG} dist

	@echo Building dev local arm container with tag ${TAG}

	@echo Building container ${DOCKER_PKG}/${NAME_PKG2}-arm-dev:${TAG}
	@docker build -f dockerfile --build-arg NAME=${NAME_PKG2}-linux-arm -t ${DOCKER_PKG}/${NAME_PKG2}-arm-dev:${TAG} dist

else ifneq ($(shell echo ${TAG} | grep -x -E ${PRD_TAG_REGEX}),)
	@echo Building prd central container with tag ${TAG}

	@echo Building container ${DOCKER_PKG}/${NAME_PKG1}:${TAG}
	@docker build -f dockerfile --build-arg NAME=${NAME_PKG1}-linux-amd64 -t ${DOCKER_PKG}/${NAME_PKG1}:${TAG} dist

	@echo Building prd local amd64 container with tag ${TAG}

	@echo Building container ${DOCKER_PKG}/${NAME_PKG2}-amd64:${TAG}
	@docker build -f dockerfile --build-arg NAME=${NAME_PKG2}-linux-amd64 -t ${DOCKER_PKG}/${NAME_PKG2}-amd64:${TAG} dist

	@echo Building prd local arm container with tag ${TAG}

	@echo Building container ${DOCKER_PKG}/${NAME_PKG2}-arm:${TAG}
	@docker build -f dockerfile --build-arg NAME=${NAME_PKG2}-linux-arm -t ${DOCKER_PKG}/${NAME_PKG2}-arm:${TAG} dist

endif

deploy: docker
	@echo Logging into Github Package Registry
	@docker login ${DOCKER_URL} -u ${DOCKER_USERNAME} -p ${DOCKER_PASSWORD}

ifeq (${COMMIT_HASH}, ${TAG})
	@echo Pushing dev central container with tag ${COMMIT_HASH}

	@echo Pushing container ${DOCKER_PKG}/${NAME_PKG1}-dev:${COMMIT_HASH}
	@docker push ${DOCKER_PKG}/${NAME_PKG1}-dev:${COMMIT_HASH}

	@echo Pushing dev local amd64 container with tag ${COMMIT_HASH}

	@echo Pushing container ${DOCKER_PKG}/${NAME_PKG2}-amd64-dev:${COMMIT_HASH}
	@docker push ${DOCKER_PKG}/${NAME_PKG2}-amd64-dev:${COMMIT_HASH}

	@echo Pushing dev local arm container with tag ${COMMIT_HASH}

	@echo Pushing container ${DOCKER_PKG}/${NAME_PKG2}-arm-dev:${COMMIT_HASH}
	@docker push ${DOCKER_PKG}/${NAME_PKG2}-arm-dev:${COMMIT_HASH}

else ifneq ($(shell echo ${TAG} | grep -x -E ${DEV_TAG_REGEX}),)
	@echo Pushing dev central container with tag ${TAG}

	@echo Pushing container ${DOCKER_PKG}/${NAME_PKG1}-dev:${TAG}
	@docker push ${DOCKER_PKG}/${NAME_PKG1}-dev:${TAG}

	@echo Pushing dev local amd64 container with tag ${TAG}

	@echo Pushing container ${DOCKER_PKG}/${NAME_PKG2}-amd64-dev:${TAG}
	@docker push ${DOCKER_PKG}/${NAME_PKG2}-amd64-dev:${TAG}

	@echo Pushing dev local arm container with tag ${TAG}

	@echo Pushing container ${DOCKER_PKG}/${NAME_PKG2}-arm-dev:${TAG}
	@docker push ${DOCKER_PKG}/${NAME_PKG2}-arm-dev:${TAG}

else ifneq ($(shell echo ${TAG} | grep -x -E ${PRD_TAG_REGEX}),)
	@echo Pushing prd central container with tag ${TAG}

	@echo Pushing container ${DOCKER_PKG}/${NAME_PKG1}:${TAG}
	@docker push ${DOCKER_PKG}/${NAME_PKG1}:${TAG}

	@echo Pushing prd local amd64 container with tag ${TAG}

	@echo Pushing container ${DOCKER_PKG}/${NAME_PKG2}-amd64:${TAG}
	@docker push ${DOCKER_PKG}/${NAME_PKG2}-amd64:${TAG}

	@echo Pushing prd local arm container with tag ${TAG}

	@echo Pushing container ${DOCKER_PKG}/${NAME_PKG2}-arm:${TAG}
	@docker push ${DOCKER_PKG}/${NAME_PKG2}-arm:${TAG}

endif

clean:
	@go clean
	@rm -rf dist/
