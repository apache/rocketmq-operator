# Build the manager binary
FROM --platform=$BUILDPLATFORM golang:1.16 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY pkg/ pkg/

# Build
ARG TARGETOS TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -a -o manager main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
#FROM gcr.io/distroless/static:nonroot
FROM --platform=$TARGETPLATFORM openjdk:8-alpine

# Install rocketmq release into image
RUN apk add --no-cache bash gettext nmap-ncat openssl busybox-extras
ENV ROCKETMQ_HOME  /home/rocketmq
ENV ROCKETMQ_VERSION 4.9.4
WORKDIR  ${ROCKETMQ_HOME}
RUN set -eux; \
    apk add --virtual .build-deps curl gnupg unzip; \
    curl https://archive.apache.org/dist/rocketmq/${ROCKETMQ_VERSION}/rocketmq-all-${ROCKETMQ_VERSION}-bin-release.zip -o rocketmq.zip; \
    curl https://archive.apache.org/dist/rocketmq/${ROCKETMQ_VERSION}/rocketmq-all-${ROCKETMQ_VERSION}-bin-release.zip.asc -o rocketmq.zip.asc; \
    curl -L https://www.apache.org/dist/rocketmq/KEYS -o KEYS; \
    gpg --import KEYS; \
    gpg --batch --verify rocketmq.zip.asc rocketmq.zip; \
    unzip rocketmq.zip; \
	mv rocketmq-*/* . ; \
    chmod a+x * ; \
	rmdir rocketmq-* ; \
	rm rocketmq.zip; \
	apk del .build-deps ; \
    rm -rf /var/cache/apk/* ; \
    rm -rf /tmp/*
RUN chown -R 65532:0 ${ROCKETMQ_HOME}

# Install controller binary
WORKDIR /
COPY --from=builder /workspace/manager .
USER 65532:65532

ENTRYPOINT ["/manager"]
