# Stage 1: Build the Go binary
FROM golang:1.25 AS builder

# Update to the latest version of the base image
# RUN apt-get update && apt-get upgrade -y

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum to download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project
COPY . .

# Build the Go binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o urlshort ./cmd/urlshort/main.go

# --- 

# Stage 2: Create a minimal image
FROM scratch

# Set the working directory inside the container
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/urlshort /usr/local/bin/urlshort

# Expose the port the application will run on
EXPOSE 8000

# Command to run the application
ENTRYPOINT ["urlshort"]