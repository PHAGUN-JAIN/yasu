# Makefile for building the Docker image

# Variables
DOCKER_COMPOSE_BUILD_FILE=build/images/build.yaml
DOCKER_COMPOSE_FILE=build/compose/services.yaml

SERVICE_NAME=yasu

# Default target
.PHONY: build
build:
	docker-compose -f $(DOCKER_COMPOSE_BUILD_FILE) build $(SERVICE_NAME)

# Bring up all services
.PHONY: up
up:
	docker-compose -f $(DOCKER_COMPOSE_FILE) up -d --build 						# --build ensures that the image is rebuilt if there are changes

# Clean up all containers and networks
.PHONY: clean
down:
	docker-compose -f $(DOCKER_COMPOSE_FILE) down --volumes --remove-orphans