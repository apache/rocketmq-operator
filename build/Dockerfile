#
# Licensed to the Apache Software Foundation (ASF) under one or more
# contributor license agreements.  See the NOTICE file distributed with
# this work for additional information regarding copyright ownership.
# The ASF licenses this file to You under the Apache License, Version 2.0
# (the "License"); you may not use this file except in compliance with
# the License.  You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

#FROM registry.access.redhat.com/ubi7-dev-preview/ubi-minimal:7.6
FROM openjdk:8-alpine

RUN apk add --no-cache bash gettext nmap-ncat openssl busybox-extras

ENV OPERATOR=/usr/local/bin/rocketmq-operator \
    USER_UID=1001 \
    USER_NAME=rocketmq-operator

# install operator binary
COPY build/_output/bin/rocketmq-operator ${OPERATOR}

COPY build/bin /usr/local/bin
RUN  /usr/local/bin/user_setup

ENV ROCKETMQ_VERSION 4.5.0

# Rocketmq home
ENV ROCKETMQ_HOME  /home/rocketmq/rocketmq-${ROCKETMQ_VERSION}

WORKDIR  ${ROCKETMQ_HOME}

COPY build/rocketmq.zip ${ROCKETMQ_HOME}/rocketmq.zip

# Install
RUN set -eux; \
    apk add --virtual .build-deps unzip; \
    unzip rocketmq.zip; \
	mv rocketmq-all*/* . ; \
	rmdir rocketmq-all* ; \
	rm rocketmq.zip; \
	apk del .build-deps ; \
    rm -rf /var/cache/apk/* ; \
    rm -rf /tmp/*

RUN chown -R ${USER_UID}:0 ${ROCKETMQ_HOME}

WORKDIR  ${ROCKETMQ_HOME}/bin

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}
