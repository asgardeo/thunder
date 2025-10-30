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
MOCKERY ?= $(TOOL_BIN)/mockery

# Tools versions
GOLANGCI_LINT_VERSION ?= v1.64.8
MOCKERY_VERSION ?= v3.5.5

$(TOOL_BIN):
	mkdir -p $(TOOL_BIN)

all: prepare clean build_with_coverage build

backend: prepare clean build_with_coverage build_backend

prepare:
	chmod +x build.sh

clean_all:
	./build.sh clean_all $(OS) $(ARCH)

clean:
	./build.sh clean $(OS) $(ARCH)

build: build_backend build_frontend build_samples

build_backend:
	./build.sh build_backend $(OS) $(ARCH)

build_frontend:
	./build.sh build_frontend

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
	./build.sh build_frontend
	./build.sh test_integration $(OS) $(ARCH)
	./build.sh merge_coverage $(OS) $(ARCH)
	@echo "================================================================"

build_with_coverage_only:
	@echo "================================================================"
	@echo "Building with coverage instrumentation (unit tests only)..."
	@echo "================================================================"
	./build.sh test_unit $(OS) $(ARCH)
	ENABLE_COVERAGE=true ./build.sh build_backend $(OS) $(ARCH)
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

lint: lint_backend lint_frontend

lint_backend: golangci-lint
	cd backend && $(GOLANGCI_LINT) run ./...

lint_frontend:
	cd frontend && pnpm install && pnpm build && pnpm lint

mockery: install-mockery
	cd backend && $(MOCKERY) --config .mockery.public.yml
	cd backend && $(MOCKERY) --config .mockery.private.yml

help:
	@echo "Makefile targets:"
	@echo "  all                           - Clean, build, and test the project."
	@echo "  backend                       - Clean, build, and test only the backend."
	@echo "  clean                         - Remove build artifacts."
	@echo "  clean_all                     - Remove all build artifacts including distribution files."
	@echo "  build                         - Build Thunder (backend + frontend + samples)."
	@echo "  build_backend                 - Build the backend Go application."
	@echo "  build_frontend                - Build the frontend applications."
	@echo "  package_samples               - Package sample applications."
	@echo "  build_samples                 - Build sample applications."
	@echo "  test_unit                     - Run unit tests."
	@echo "  test_integration              - Run integration tests."
	@echo "  build_with_coverage  		   - Build with coverage flags, run unit and integration tests, and generate combined coverage report."
	@echo "  build_with_coverage_only      - Build with coverage instrumentation (unit tests only, no integration tests)."
	@echo "  test                          - Run all tests (unit and integration)."
	@echo "  run                           - Build and run the application locally."
	@echo "  docker-build                  - Build single-arch Docker image with version tag."
	@echo "  docker-build-latest           - Build single-arch Docker image with latest tag."
	@echo "  docker-build-multiarch        - Build multi-arch Docker image with version tag."
	@echo "  docker-build-multiarch-latest - Build multi-arch Docker image with latest tag."
	@echo "  docker-build-multiarch-push   - Build and push multi-arch images to registry."
	@echo "  lint                          - Run linting on both backend and frontend code."
	@echo "  lint_backend                  - Run golangci-lint on the backend code."
	@echo "  lint_frontend                 - Run ESLint on the frontend code."
	@echo "  mockery                       - Generate mocks for unit tests using mockery."
	@echo "  help                          - Show this help message."

.PHONY: all prepare clean clean_all build build_backend build_frontend build_samples package_samples run
.PHONY: docker-build docker-build-latest docker-build-multiarch 
.PHONY: docker-build-multiarch-latest docker-build-multiarch-push
.PHONY: test_unit test_integration build_with_coverage build_with_coverage_only test
.PHONY: help go_install_tool
.PHONY: lint lint_backend lint_frontend golangci-lint mockery install-mockery

define go_install_tool
	cd /tmp && \
	GOBIN=$(TOOL_BIN) go install $(2)@$(3)
endef

golangci-lint: $(GOLANGCI_LINT)

$(GOLANGCI_LINT): $(TOOL_BIN)
	$(call go_install_tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/cmd/golangci-lint,$(GOLANGCI_LINT_VERSION))

install-mockery: $(MOCKERY)

$(MOCKERY): $(TOOL_BIN)
	$(call go_install_tool,$(MOCKERY),github.com/vektra/mockery/v3,$(MOCKERY_VERSION))
