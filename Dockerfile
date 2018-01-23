# Stage 1 - Binary Build
# BUILD_X args should be passed at build time as docker build args
FROM golang:1.9.1-alpine AS builder
ARG APP
ARG BUILD_TIME
ARG BUILD_VERSION
ARG BUILD_COMMIT
RUN apk update && apk add build-base libressl-dev
WORKDIR /go/src/compass
COPY ./ /go/src/compass
RUN COMPRESS_BINARY=1 GOBUILD_VERBOSE=1 BIN_NAME=bin make ${APP}

# Stage 2 - Final Image
# The application should be statically linked
FROM alpine:3.6
ARG APP
ENV ENTRYPOINT ${APP}
RUN apk update && apk add --no-cache ca-certificates && rm -rf /var/cache/apk/*
COPY --from=builder /go/src/compass/bin /usr/bin/${APP}
ENTRYPOINT $ENTRYPOINT
