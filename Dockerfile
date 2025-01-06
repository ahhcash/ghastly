# Stage 1: Download dependencies
FROM ubuntu:22.04 AS deps

# Install curl and other necessary tools
RUN apt-get update && apt-get install -y \
    curl \
    tar \
    && rm -rf /var/lib/apt/lists/*

# Download and extract tokenizers library
RUN mkdir -p /libs/static/libtokenizers
RUN curl -L -o tokenizers.tar.gz "https://github.com/daulet/tokenizers/releases/latest/download/libtokenizers.linux-x86_64.tar.gz" \
    && tar xzf tokenizers.tar.gz -C /libs/static/libtokenizers

# Stage 2: Build the application
FROM golang:1.23 AS builder

# Install build essentials and ONNX Runtime
RUN apt-get update && apt-get install -y \
    build-essential \
    pkg-config \
    python3-pip \
    && pip3 install onnxruntime \
    && rm -rf /var/lib/apt/lists/*

# Copy tokenizers library from deps stage
COPY --from=deps /libs/static/libtokenizers /libs/static/libtokenizers

# Set working directory
WORKDIR /app

# Copy go mod files first to leverage Docker cache
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application with CGO enabled
ENV CGO_ENABLED=1
ENV CGO_LDFLAGS="-L/libs/static/libtokenizers"
RUN make build

# Stage 3: Final runtime image
FROM ubuntu:22.04

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    python3-pip \
    && pip3 install onnxruntime \
    && rm -rf /var/lib/apt/lists/*

# Copy the built binary from builder stage
COPY --from=builder /app/bin/ghastly /app/ghastly

# Copy the tokenizers library for runtime
COPY --from=deps /libs/static/libtokenizers /libs/static/libtokenizers

# Set environment variables for runtime
ENV CGO_ENABLED=1
ENV CGO_LDFLAGS="-L/libs/static/libtokenizers"
ENV PORT=8080

# Set the working directory
WORKDIR /app

# Command to run the application
CMD ["/app/ghastly"]