# suppress output, run `make XXX V=` to be verbose
V := @

# Common
NAME = sport-news
VCS = gitlab.com
VERSION := $(shell git describe --always --tags)
CURRENT_TIME := $(shell TZ="Europe/Moscow" date +"%d-%m-%y %T")

# Build
OUT_DIR = ./bin
MAIN_PKG = ./cmd/${NAME}
ACTION ?= build
GC_FLAGS = -gcflags 'all=-N -l'
LD_FLAGS = -ldflags "-s -v -w -X 'main.version=${VERSION}' -X 'main.buildTime=${CURRENT_TIME}'"
BUILD_CMD = CGO_ENABLED=1 go build -o ${OUT_DIR}/${NAME} ${LD_FLAGS} ${MAIN_PKG}
DEBUG_CMD = CGO_ENABLED=1 go build -o ${OUT_DIR}/${NAME} ${GC_FLAGS} ${MAIN_PKG}

# Docker
DOCKERFILE = infra/Dockerfile

# Other
.DEFAULT_GOAL = build

.PHONY: build
build:
	@echo BUILDING PRODUCTION $(NAME)
	$(V)${BUILD_CMD}
	@echo DONE

.PHONY: build-debug
build-debug:
	@echo BUILDING DEBUG $(NAME)
	$(V)${DEBUG_CMD}
	@echo DONE

.PHONY: docker-build
docker-build:
	$(call run_in_docker,make ${ACTION})

.PHONY: clean
clean:
	@echo "Removing $(OUT_DIR)"
	$(V)rm -rf $(OUT_DIR)

.PHONY: vendor
vendor:
	$(V)GOPRIVATE=${VCS}/* go mod tidy -compat=1.21
	$(V)GOPRIVATE=${VCS}/* go mod vendor
	$(V)git add vendor go.mod go.sum

CURR_REPO := /$(notdir $(PWD))
define run_in_docker
	$(V)docker run --rm \
		-v $(PWD):$(CURR_REPO) \
		-w $(CURR_REPO) \
		${DOCKER_GOLANG_IMAGE} $1
endef
