ARG BASE_IMAGE=docker.io/library/ubuntu:23.04
####################################################################################################
# Builder image
# Initial stage which pulls prepares build dependencies and CLI tooling we need for our final image
# Also used as the image in CI jobs so needs all dependencies
####################################################################################################
FROM docker.io/library/golang:1.19.5 AS builder

RUN echo 'deb http://deb.debian.org/debian buster-backports main' >> /etc/apt/sources.list

RUN apt-get update && apt-get install --no-install-recommends -y \
    openssh-server \
    nginx \
    unzip \
    fcgiwrap \
    git \
    make \
    wget \
    gcc \
    sudo \
    zip && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

WORKDIR /tmp

COPY hack/install.sh hack/tool-versions.sh ./
COPY hack/installers installers

####################################################################################################
# Build helm
####################################################################################################
FROM golang:1.19.5 as helm-builder
WORKDIR /
RUN git clone -b v3.10.3 https://github.com/helm/helm && \
    cd helm && \
    make install

####################################################################################################
# Build kustomize
####################################################################################################
FROM golang:1.19.5 as kustomize-builder
WORKDIR /
RUN GOBIN=$(pwd)/ GO111MODULE=on go install sigs.k8s.io/kustomize/kustomize/v4@latest

####################################################################################################
# Argo CD Base - used as the base for both the release and dev argocd images
####################################################################################################
FROM $BASE_IMAGE AS argocd-base

USER root

ENV DEBIAN_FRONTEND=noninteractive

RUN groupadd -g 999 argocd && \
    useradd -r -u 999 -g argocd argocd && \
    mkdir -p /home/argocd && \
    chown argocd:0 /home/argocd && \
    chmod g=u /home/argocd && \
    apt-get update && \
    apt-get upgrade -y && \
    apt-get dist-upgrade -y && \
    apt-get install -y git tini gpg tzdata wget && \
    # START - Install git-lfs
    wget https://github.com/git-lfs/git-lfs/releases/download/v3.3.0/git-lfs-linux-amd64-v3.3.0.tar.gz && \
    tar -xvf git-lfs-linux-amd64-v3.3.0.tar.gz && \
    cp ./git-lfs-3.3.0/git-lfs /usr/bin/git-lfs && \
    # END - Install git-lfs
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

COPY hack/gpg-wrapper.sh /usr/local/bin/gpg-wrapper.sh
COPY hack/git-verify-wrapper.sh /usr/local/bin/git-verify-wrapper.sh
COPY --from=helm-builder /helm/bin/helm /usr/local/bin/helm
COPY --from=kustomize-builder /kustomize /usr/local/bin/kustomize
COPY entrypoint.sh /usr/local/bin/entrypoint.sh
# keep uid_entrypoint.sh for backward compatibility
RUN ln -s /usr/local/bin/entrypoint.sh /usr/local/bin/uid_entrypoint.sh

# support for mounting configuration from a configmap
WORKDIR /app/config/ssh
RUN touch ssh_known_hosts && \
    ln -s /app/config/ssh/ssh_known_hosts /etc/ssh/ssh_known_hosts

WORKDIR /app/config
RUN mkdir -p tls && \
    mkdir -p gpg/source && \
    mkdir -p gpg/keys && \
    chown argocd gpg/keys && \
    chmod 0700 gpg/keys

ENV USER=argocd

USER 999
WORKDIR /home/argocd

####################################################################################################
# Argo CD UI stage
####################################################################################################
FROM --platform=$BUILDPLATFORM docker.io/library/node:19.4.0 AS argocd-ui

WORKDIR /src
COPY ["ui/package.json", "ui/yarn.lock", "./"]

RUN yarn --update-checksums && \
    yarn install --network-timeout 200000 && \
    yarn cache clean

COPY ["ui/", "."]

ARG ARGO_VERSION=latest
ENV ARGO_VERSION=$ARGO_VERSION
ARG TARGETARCH
RUN HOST_ARCH=$TARGETARCH NODE_ENV='production' NODE_ONLINE_ENV='online' NODE_OPTIONS=--max_old_space_size=8192 yarn build

####################################################################################################
# Argo CD Build stage which performs the actual build of Argo CD binaries
####################################################################################################
FROM --platform=$BUILDPLATFORM docker.io/library/golang:1.19.5 AS argocd-build

WORKDIR /go/src/github.com/argoproj/argo-cd

COPY go.* ./
RUN go mod download

# Perform the build
COPY . .
COPY --from=argocd-ui /src/dist/app /go/src/github.com/argoproj/argo-cd/ui/dist/app
ARG TARGETOS
ARG TARGETARCH
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH make argocd-all

####################################################################################################
# Final image
####################################################################################################
FROM argocd-base
COPY --from=argocd-build /go/src/github.com/argoproj/argo-cd/dist/argocd* /usr/local/bin/

USER root
RUN ln -s /usr/local/bin/argocd /usr/local/bin/argocd-server && \
    ln -s /usr/local/bin/argocd /usr/local/bin/argocd-repo-server && \
    ln -s /usr/local/bin/argocd /usr/local/bin/argocd-cmp-server && \
    ln -s /usr/local/bin/argocd /usr/local/bin/argocd-application-controller && \
    ln -s /usr/local/bin/argocd /usr/local/bin/argocd-dex && \
    ln -s /usr/local/bin/argocd /usr/local/bin/argocd-notifications && \
    ln -s /usr/local/bin/argocd /usr/local/bin/argocd-applicationset-controller && \
    ln -s /usr/local/bin/argocd /usr/local/bin/argocd-k8s-auth

USER 999