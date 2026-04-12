# Makefile for building the Docker image

# Variables
DOCKER_COMPOSE_FILE=build/images/build.yaml
DEV_COMPOSE_FILE=build/dev/services.yaml
SERVICE_NAME=urlshort
NETWORK_NAME=urlshort-network

# Default target
.PHONY: build
build:
	docker-compose -f $(DOCKER_COMPOSE_FILE) build $(SERVICE_NAME)

# Bring up dev services
.PHONY: up-dev
up-dev:
	docker-compose -f $(DEV_COMPOSE_FILE) up -d

# Ensure the network exists
.PHONY: create-network
create-network:
	docker network inspect $(NETWORK_NAME) >/dev/null 2>&1 || docker network create $(NETWORK_NAME)

# Run the Docker container
.PHONY: run
run: up-dev build create-network
	docker run -p 8000:8000 --name $(SERVICE_NAME)-container --network=$(NETWORK_NAME) $(SERVICE_NAME):latest