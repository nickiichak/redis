################ Preparation ####################
REGISTRY := nick1chak
VERSION := $(shell cat VERSION)
DOCKER_IMAGE_SERV := $(REGISTRY)/redis:$(VERSION)
DOCKER_IMAGE_CLI := $(REGISTRY)/red_client:$(VERSION)

REPOSITORY_PATH := /usr/src/redis_mk2
SERVER_PATH := serv
CLIENT_PATH := client

DOCKER_BUILDER := golang:1.11
DOCKER_RUNNER =  docker run --rm -v $(CURDIR):$(REPOSITORY_PATH)
DOCKER_RUNNER += -w $(REPOSITORY_PATH)/
################ End Preparation ####################



################ Binary Target ####################
.PHONY: build
build: build_client build_serv

.PHONY: build_client
build_client:
	$(DOCKER_RUNNER)$(CLIENT_PATH) $(DOCKER_BUILDER) go build

.PHONY: build_serv
build_serv:
	$(DOCKER_RUNNER)$(SERVER_PATH) $(DOCKER_BUILDER) go build

############### Docker Target ####################
.PHONY: run
run: build_serv
	$(CURDIR)/$(SERVER_PATH)/serv

.PHONY: docker_build_serv
docker_build_serv:
	docker build -t $(DOCKER_IMAGE_SERV) ./

.PHONY: docker_build_cli
docker_build_cli:
	docker build -t $(DOCKER_IMAGE_CLI) ./
