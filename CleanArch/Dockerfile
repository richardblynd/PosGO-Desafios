# Build stage
FROM golang:1.24-alpine AS builder

# Set necessary environmet variables needed for our image
ENV GO111MODULE=on \
    CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=amd64

# Install dependencies
RUN apk add --no-cache gcc musl-dev

# Move to working directory /build
WORKDIR /build

# Copy and download dependency using go mod
COPY go.mod go.sum ./
RUN go mod download

# Copy the code into the container
COPY . .

# Build the application
RUN go build -o ordersystem cmd/ordersystem/main.go cmd/ordersystem/wire_gen.go

# Move to /dist directory as the place for resulting binary folder
WORKDIR /dist

# Copy binary from build to main folder
RUN cp /build/ordersystem .

# Build a small image
FROM alpine:latest

# Install ca-certificates
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /dist/ordersystem .

# Copy migrations folder
COPY migrations ./migrations/

# Copy config files
COPY configs ./configs/
COPY cmd/ordersystem/.env ./.env

# Expose port
EXPOSE 8000 8080 50051

# Command to run
ENTRYPOINT ["./ordersystem"]