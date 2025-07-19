# WSO2 Thunder Docker Image
# Build stage - compile the Go binary for the target architecture
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make bash sqlite openssl zip

# Set the working directory
WORKDIR /app

# Copy the entire source code
COPY . .

# Build the binary for the target architecture
ARG TARGETARCH
RUN if [ "$TARGETARCH" = "amd64" ]; then \
        ./build.sh build_backend linux amd64; \
    else \
        ./build.sh build_backend linux arm64; \
    fi

# Runtime stage
FROM alpine:3.19

# Install required packages
RUN apk add --no-cache \
    ca-certificates \
    lsof \
    sqlite \
    bash \
    curl \
    unzip

# Create thunder user and group
RUN addgroup -S thunder && adduser -S thunder -G thunder

# Create application directory
WORKDIR /opt/thunder

# Copy and extract the thunder package from builder stage
# TARGETARCH is automatically set by Docker during multi-arch builds
ARG TARGETARCH
COPY --from=builder /app/target/dist/ /tmp/dist/
RUN cd /tmp/dist && \
    if [ "$TARGETARCH" = "amd64" ]; then \
        find . -name "thunder-[0-9]*-linux-x64.zip" -exec cp {} /tmp/ \; ; \
    else \
        find . -name "thunder-[0-9]*-linux-arm64.zip" -exec cp {} /tmp/ \; ; \
    fi && \
    cd /tmp && \
    unzip thunder-*.zip && \
    cp -r thunder-*/* /opt/thunder/ && \
    rm -rf /tmp/thunder-* /tmp/dist

# Set ownership and permissions
RUN chown -R thunder:thunder /opt/thunder && \
    chmod +x thunder start.sh scripts/init_script.sh

# Expose the default port
EXPOSE 8090

# Switch to thunder user
USER thunder

# Set environment variables
ENV BACKEND_PORT=8090

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
    CMD curl -k -f https://localhost:8090/health || exit 1

# Start the application
CMD ["./start.sh"]
