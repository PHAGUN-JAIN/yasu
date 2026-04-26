# Makefile for building the Docker image

# Variables
DOCKER_COMPOSE_BUILD_FILE=build/images/build.yaml
DOCKER_COMPOSE_FILE=build/compose/services.yaml

DOCKER_DEV_COMPOSE_FILE=build/compose/services-dev.yaml

SERVICE_NAME=yasu

# Run linter
.PHONY: lint  # show all issues
lint:
	golangci-lint run ./...

# Run linter and auto-fix what it can
.PHONY: lint-fix # auto fix what's possible
lint-fix:
	golangci-lint run --fix ./...

# Check formatting only
.PHONY: fmt  # show unformatted files
fmt:
	gofmt -l .

# Auto format all files
.PHONY: fmt-fix  # format everything
fmt-fix:
	gofmt -w .


# Setup git hooks — run once after cloning
.PHONY: setup
setup:
	git config core.hooksPath .githooks
	@echo "Git hooks configured"

# Default target
.PHONY: build
build:
	docker-compose -f $(DOCKER_COMPOSE_BUILD_FILE) build $(SERVICE_NAME)

# Bring up all services
.PHONY: up
up:
	docker-compose -f $(DOCKER_COMPOSE_FILE) up -d --build 						# --build ensures that the image is rebuilt if there are changes

# Bring up development infrastructure services
.PHONY: dev-up
dev-up:
	docker-compose -f $(DOCKER_DEV_COMPOSE_FILE) up -d --build 		

# Clean up all containers and networks
.PHONY: clean
down:
	docker-compose -f $(DOCKER_COMPOSE_FILE) down --volumes --remove-orphans