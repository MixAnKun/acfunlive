FROM node:alpine AS node_build
ARG IGNORE_CHINA_MIRROR=0
ARG NODE_OPTIONS

LABEL stage=buildnode

ADD acfunlive-ui /acfunlive-ui-src
WORKDIR /acfunlive-ui-src

RUN \
    echo "NODE_OPTIONS=${NODE_OPTIONS}" && \
    if [ ! "$IGNORE_CHINA_MIRROR" = 1 ]; then \
    sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories; \
    fi; \
    apk update && \
    apk add yarn && \
    if [ ! "$IGNORE_CHINA_MIRROR" = 1 ]; then \
    yarn config set registry "https://registry.npm.taobao.org/"; \
    fi; \
    yarn install && \
    yarn generate

FROM golang:1-alpine AS go_build
ARG IGNORE_CHINA_MIRROR=0

LABEL stage=buildgo

ADD . /acfunlive-src
WORKDIR /acfunlive-src

ENV GO111MODULE=on \
    GOPROXY="https://goproxy.cn" \
    CGO_ENABLED=0

RUN if [ "$IGNORE_CHINA_MIRROR" = 1 ]; then \
    unset GOPROXY; \
    else \
    sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories; \
    fi; \ 
    apk add git && \
    go build

FROM alpine
ARG IGNORE_CHINA_MIRROR=0

ENV BINFILE="/acfunlive/acfunlive" \
    WEBUIDIR="/acfunlive/webui" \
    CONFIGDIR="/acfunlive/config" \
    RECORDDIR="/acfunlive/record"

EXPOSE 51880
EXPOSE 51890

RUN mkdir -p $WEBUIDIR && \
    mkdir -p $CONFIGDIR && \
    mkdir -p $RECORDDIR && \
    if [ ! "$IGNORE_CHINA_MIRROR" = 1 ]; then \
    sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories; \
    fi; \ 
    apk update && \
    apk upgrade && \
    apk --no-cache add ffmpeg libc6-compat tzdata && \
    ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

COPY --from=node_build /acfunlive-ui-src/dist $WEBUIDIR
COPY --from=go_build /acfunlive-src/acfunlive $BINFILE

VOLUME $CONFIGDIR $RECORDDIR

ENTRYPOINT ["/acfunlive/acfunlive", "-config", "/acfunlive/config", "-record", "/acfunlive/record"]
