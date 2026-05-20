# Copyright (c) 2024, Shanghai Iluvatar CoreX Semiconductor Co., Ltd.
# All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

REGISTRY ?= registry.iluvatar.com.cn/k8s
TARGET := ix-exporter
VERSION ?= latest
ARCH ?= x86_64
KUBE_VERSION ?= v1.32.0

MODULE := gitee.com/deep-spark/ixexporter
DOCKER ?= docker

# Use $(if ...) so the name is resolved when expanded (not only at parse time).
# Prefix assignment (REGISTRY=... make image), CLI overrides (make image REGISTRY=...),
# and exported env all work; avoid sudo without -E or it drops those vars.
IMAGE_NAME = $(if $(strip $(REGISTRY)),$(strip $(REGISTRY))/ix-exporter,ix-exporter):$(VERSION)-$(ARCH)

GOOS := linux

BUILD_DIR := build
COREX_PATH := /usr/local/corex
COREX_LIBS := libixml.so \
        	  libcuda.so \
        	  libcuda.so.1 \
        	  libixthunk.so

IXDCGM_PATH := /usr/local/ixdcgm
IXDCGM_LIBS := libixdcgm.so		   

.PHONY: all
all: build image

# Debug: REGISTRY=... VERSION=... make print-image
.PHONY: print-image
print-image:
	@echo "$(IMAGE_NAME)"

.PHONY: build
build:
	CGO_CFLAGS=-I${COREX_PATH}/include GOOS=$(GOOS) CGO_ENABLED=1 \
	go build -mod=mod -ldflags "-s -w" \
	    -o $(BUILD_DIR)/$(TARGET) $(MODULE)/cmd/$(TARGET)

.PHONY: image
image: build
	mkdir -p $(BUILD_DIR)/lib64
	@$(foreach lib, $(COREX_LIBS), if [ ! -f $(COREX_PATH)/lib64/$(lib) ]; then echo "$(lib) not found"; exit 1; fi;)
	$(foreach lib, $(COREX_LIBS), cp -P $(COREX_PATH)/lib64/$(lib) $(BUILD_DIR)/lib64;)
	@$(foreach lib, $(IXDCGM_LIBS), if [ ! -f $(IXDCGM_PATH)/lib64/$(lib) ]; then echo "$(lib) not found"; exit 1; fi;)
	$(foreach lib, $(IXDCGM_LIBS), cp -P $(IXDCGM_PATH)/lib64/$(lib)* $(BUILD_DIR)/lib64;)

	DOCKER_BUILDKIT=1 $(DOCKER) build \
	        -t $(IMAGE_NAME) \
	        --build-arg EXEC=$(BUILD_DIR)/$(TARGET) \
			--build-arg LIB_DIR=$(BUILD_DIR)/lib64 \
			--build-arg KUBE_VERSION=$(KUBE_VERSION) \
			--build-arg PLUGIN_VERSION=$(VERSION) \
	        -f Dockerfile \
			.

CHART_NAME := $(shell grep 'name:' deployment/helm/ix-exporter/Chart.yaml | awk '{print $$2}')
CHART_VERSION := $(shell grep 'version:' deployment/helm/ix-exporter/Chart.yaml | awk '{print $$2}')
CHART_PACKAGE := $(CHART_NAME)-$(CHART_VERSION).tgz

.PHONY: chart-push
chart-push:
	@echo "Packaging chart $(CHART_NAME) version $(CHART_VERSION)..."
	helm package deployment/helm/ix-exporter
	
	@echo "Pushing $(CHART_PACKAGE) to OCI registry..."
	helm push $(CHART_PACKAGE) oci://$(REGISTRY)
	
	rm -f $(CHART_PACKAGE)

clean:
	rm -rf $(BUILD_DIR)
	rm -f $(CHART_PACKAGE)

# Apply go tools to the codebase
GO_TARGETS := fmt vet vendor
.PHONY: $(GO_TARGETS)

vendor:
	go mod tidy
	go mod vendor
	go mod verify

fmt: vendor
	go fmt $(MODULE)/...

vet: vendor
	go vet $(MODULE)/...