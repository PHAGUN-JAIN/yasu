# Makefile for building the Docker image

# Variables
DOCKER_COMPOSE_FILE=build/compose/services.yaml
SERVICE_NAME=urlshort
NETWORK_NAME=compose_urlshort-network

# Default target
.PHONY: build
build:
	docker-compose -f $(DOCKER_COMPOSE_FILE) build $(SERVICE_NAME)

# Bring up all services
.PHONY: up
up:
	docker-compose -f $(DOCKER_COMPOSE_FILE) up -d

# Clean up all containers and networks
.PHONY: clean
down:
	docker-compose -f $(DOCKER_COMPOSE_FILE) down --volumes --remove-orphans
	docker network rm $(NETWORK_NAME) || true