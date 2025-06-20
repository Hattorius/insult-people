# Stage 1: Build the Go binary
FROM golang:1.22-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go source code
COPY main.go .

# Download dependencies (if needed) and build the binary
RUN go mod init app && go build -o main .

# Stage 2: Run the Go binary in a minimal image
FROM alpine:latest

# Set the working directory inside the container
WORKDIR /app

# Copy the compiled Go binary from the builder stage
COPY --from=builder /app/main .

# Expose port (CapRover expects something to run on port 80)
EXPOSE 8080

# Command to run the binary
CMD ["./main"]
