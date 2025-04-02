# Build Stage: Use official Golang image to compile the binary
FROM golang:1.23 AS builder

# Set working directory
WORKDIR /app

# Copy Go modules manifests
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the application source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o pod-label-controller .

# Final Stage: Use a minimal base image
FROM gcr.io/distroless/static:nonroot

# Set the working directory
WORKDIR /

# Copy the compiled binary from builder stage
COPY --from=builder /app/pod-label-controller .

# Run the binary
CMD ["/pod-label-controller"]
