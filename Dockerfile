# Use the official Golang image as the base image
FROM golang:1.24-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go modules manifests
COPY go.mod go.sum ./

# Download Go module dependencies
RUN go mod download

# Copy the application source code
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 go build -o rag cmd/rag/main.go

FROM alpine:latest

RUN apk add --no-cache ca-certificates
# Set the working directory inside the container
WORKDIR /app

# Copy the built Go binary from the builder stage
COPY --from=builder /app/rag .

# Expose the application port
EXPOSE 8080

# Command to run the application
CMD ["./rag"]
