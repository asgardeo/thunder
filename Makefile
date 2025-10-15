# ----------------------------------------------------------------------------
# Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
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

# Constants
VERSION_FILE=version.txt
VERSION=$(shell cat $(VERSION_FILE))
BINARY_NAME=thunder

# Tools
PROJECT_DIR := $(realpath $(dir $(abspath $(lastword $(MAKEFILE_LIST)))))/backend
PROJECT_BIN_DIR := $(PROJECT_DIR)/bin
TOOL_BIN ?= $(PROJECT_BIN_DIR)/tools
GOLANGCI_LINT ?= $(TOOL_BIN)/golangci-lint
GOLANGCI_LINT_VERSION ?= v1.64.8

$(TOOL_BIN):
	mkdir -p $(TOOL_BIN)

all: lint prepare clean build_with_coverage build

backend: lint prepare clean build_with_coverage build_backend

prepare:
	chmod +x build.sh

clean_all:
	./build.sh clean_all $(OS) $(ARCH)

clean:
	./build.sh clean $(OS) $(ARCH)

build: build_backend build_samples

build_backend:
	./build.sh build_backend $(OS) $(ARCH)

package_samples:
	./build.sh package_samples $(OS) $(ARCH)

build_samples:
	./build.sh build_samples

test:
	./build.sh test $(OS) $(ARCH)

test_unit:
	./build.sh test_unit $(OS) $(ARCH)

test_integration:
	./build.sh test_integration $(OS) $(ARCH)

build_with_coverage:
	@echo "================================================================"
	@echo "Building with coverage for unit and integration tests..."
	@echo "================================================================"
	./build.sh test_unit $(OS) $(ARCH)
	ENABLE_COVERAGE=true ./build.sh build_backend $(OS) $(ARCH)
	./build.sh test_integration $(OS) $(ARCH)
	./build.sh merge_coverage $(OS) $(ARCH)
	@echo "================================================================"

run:
	./build.sh run $(OS) $(ARCH)

docker-build:
	docker build -t thunder:$(VERSION) .

docker-build-latest:
	docker build -t thunder:latest .

docker-build-multiarch:
	docker buildx build --platform linux/amd64,linux/arm64 -t thunder:$(VERSION) .

docker-build-multiarch-latest:
	docker buildx build --platform linux/amd64,linux/arm64 -t thunder:latest .

docker-build-multiarch-push:
	docker buildx build --platform linux/amd64,linux/arm64 -t thunder:$(VERSION) -t thunder:latest --push .

lint: golangci-lint
	cd backend && $(GOLANGCI_LINT) run ./...

help:
	@echo "Makefile targets:"
	@echo "  all                           - Clean, build, and test the project."
	@echo "  backend                       - Clean, build, and test only the backend."
	@echo "  clean                         - Remove build artifacts."
	@echo "  clean_all                     - Remove all build artifacts including distribution files."
	@echo "  build                         - Build the Go project and frontend, then package."
	@echo "  build_backend                 - Build the backend Go application."
	@echo "  package_samples               - Package sample applications."
	@echo "  build_samples                 - Build sample applications."
	@echo "  test_unit                     - Run unit tests."
	@echo "  test_integration              - Run integration tests."
	@echo "  build_with_coverage  		   - Build with coverage flags, run unit and integration tests, and generate combined coverage report."
	@echo "  test                          - Run all tests (unit and integration)."
	@echo "  run                           - Build and run the application locally."
	@echo "  docker-build                  - Build single-arch Docker image with version tag."
	@echo "  docker-build-latest           - Build single-arch Docker image with latest tag."
	@echo "  docker-build-multiarch        - Build multi-arch Docker image with version tag."
	@echo "  docker-build-multiarch-latest - Build multi-arch Docker image with latest tag."
	@echo "  docker-build-multiarch-push   - Build and push multi-arch images to registry."
	@echo "  lint                          - Run golangci-lint on the project."
	@echo "  help                          - Show this help message."

.PHONY: all prepare clean clean_all build build_samples package_samples run
.PHONY: docker-build docker-build-latest docker-build-multiarch 
.PHONY: docker-build-multiarch-latest docker-build-multiarch-push
.PHONY: test_unit test_integration build_with_coverage test
.PHONY: lint help go_install_tool golangci-lint

define go_install_tool
	cd /tmp && \
	GOBIN=$(TOOL_BIN) go install $(2)@$(3)
endef

golangci-lint: $(GOLANGCI_LINT)

$(GOLANGCI_LINT): $(TOOL_BIN)
	$(call go_install_tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/cmd/golangci-lint,$(GOLANGCI_LINT_VERSION))
