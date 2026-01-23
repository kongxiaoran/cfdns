# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o cfdns .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/cfdns .

# Create a directory for config
RUN mkdir -p /etc/cfdns

# Expose port
EXPOSE 8082

# Run the application
CMD ["./cfdns", "-config", "/etc/cfdns/data.json"]
