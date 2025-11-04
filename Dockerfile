# Dockerfile

# --- Build Stage ---
# Use a Go version that matches your go.mod file
FROM golang:1.25-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy Go modules manifests and download dependencies first to leverage Docker layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application. CGO_ENABLED=0 creates a static binary.
# The output binary is named 'proxy-server' and placed in the root of the builder.
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /proxy-server ./cmd/server

# --- Final Stage ---
# Use a minimal 'distroless' base image. It contains only our application and its runtime dependencies.
# It doesn't contain package managers, shells, or other programs, making it highly secure.
FROM gcr.io/distroless/static-debian11

# Set the working directory
WORKDIR /app

# Copy the static binary from the builder stage
COPY --from=builder /proxy-server .

# Copy the configuration file that the application will use
COPY configs/config.yaml .

# Expose the port the app runs on
EXPOSE 8080

# The command to run the application. We point to the copied config file.
CMD ["/app/proxy-server", "-config", "/app/config.yaml"]