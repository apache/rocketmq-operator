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
