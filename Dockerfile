# ----------------------------------------------------------------------------
# Copyright (c) 2025, WSO2 LLC. (http://www.wso2.com).
#
# WSO2 LLC. licenses this file to you under the Apache License,
# Version 2.0 (the "License"); you may not use this file except
# in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied. See the License for the
# specific language governing permissions and limitations
# under the License.
# ----------------------------------------------------------------------------

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

# Start the application
CMD ["./start.sh"]
