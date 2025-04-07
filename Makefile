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

TARGET := ix-exporter
VERSION ?= 4.2.0

MODULE := gitee.com/deep-spark/ixexporter
DOCKER ?= docker

ifeq ($(REGISTRY),)
IMAGE_NAME = ix-exporter:$(VERSION)-x86_64
else 
IMAGE_NAME = $(REGISTRY)/ix-exporter:$(VERSION)-x86_64
endif

GOOS := linux

BUILD_DIR := build
COREX_PATH := /usr/local/corex
COREX_LIBS := libixml.so \
        	  libcuda.so \
        	  libcuda.so.1 \
        	  libcudart.so \
        	  libcudart.so.10.2 \
        	  libcudart.so.10.2.89 \
        	  libixthunk.so

IXDCGM_PATH := /usr/local/ixdcgm
IXDCGM_LIBS := libixdcgm.so		   
                                                                                                 
.PHONY: all
all: build image

.PHONY: build
build:
	CGO_CFLAGS=-I${COREX_PATH}/include GOOS=$(GOOS) CGO_ENABLED=1 \
	go build -ldflags "-s -w" \
	    -o $(BUILD_DIR)/$(TARGET) $(MODULE)/cmd/$(TARGET)

.PHONY: image
image: clean build
	mkdir -p $(BUILD_DIR)/lib64
	@$(foreach lib, $(COREX_LIBS), if [ ! -f $(COREX_PATH)/lib64/$(lib) ]; then echo "$(lib) not found"; exit 1; fi;)
	$(foreach lib, $(COREX_LIBS), cp -P $(COREX_PATH)/lib64/$(lib) $(BUILD_DIR)/lib64;)
	@$(foreach lib, $(IXDCGM_LIBS), if [ ! -f $(IXDCGM_PATH)/lib64/$(lib) ]; then echo "$(lib) not found"; exit 1; fi;)
	$(foreach lib, $(IXDCGM_LIBS), cp -P $(IXDCGM_PATH)/lib64/$(lib)* $(BUILD_DIR)/lib64;)

	$(DOCKER) build \
	        -t $(IMAGE_NAME) \
	        --build-arg EXEC=$(BUILD_DIR)/$(TARGET) \
			--build-arg LIB_DIR=$(BUILD_DIR)/lib64 \
	        -f Dockerfile \
			.

clean:
	rm -rf $(BUILD_DIR)

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
		