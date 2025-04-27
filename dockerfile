# Stage 1: Build the application
FROM golang:1.24 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files to download dependencies first
# This leverages Docker cache
COPY go.mod ./
# Assuming go.sum exists after a 'go mod tidy'
COPY go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the application for Linux
# Based on your taskfiles, the output is in the bin directory
RUN CGO_ENABLED=0 GOOS=linux go build -v -ldflags="-s -w" -o bin/ePrometna_Server .

# Stage 2: Create the final runtime image
FROM alpine:latest

# Set the working directory
WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/bin/ePrometna_Server .

# Expose the port the application runs on (from ePrometna.json)
EXPOSE 8090

# Command to run the executable
ENTRYPOINT ["./ePrometna_Server"]
