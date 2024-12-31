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

FROM ubuntu:20.04

RUN mkdir /usr/local/corex
RUN mkdir /etc/ixexporter

ARG LIB_DIR
ARG EXEC

COPY $LIB_DIR /usr/local/corex/lib64
COPY $EXEC /usr/bin
COPY etc/metrics.yaml /etc/ixexporter

ENV LD_LIBRARY_PATH="/usr/local/corex/lib64"
ENV LIBRARY_PATH="/usr/local/corex/lib64"

LABEL io.k8s.display-name="Iluvatar Corex Exporter"
LABEL name="Iluvatar Corex Exporter"
LABEL vendor="Iluvatar Corex"
ARG PLUGIN_VERSION="N/A"
LABEL version=${PLUGIN_VERSION}
LABEL summary="Exports GPU Metrics to Prometheus"
LABEL description="See summary"
ARG GIT_COMMIT="N/A"
LABEL git-commit ${GIT_COMMIT}

ENTRYPOINT ["/usr/bin/ix-exporter"]