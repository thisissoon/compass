# Stage 1 - Binary Build
# BUILD_X args should be passed at build time as docker build args
FROM golang:1.9.1-alpine AS builder
ARG APP
ARG GIT_COMMIT
ARG GIT_SHA
ARG GIT_TAG
ARG GIT_DIRTY
RUN apk update && apk add build-base libressl-dev
WORKDIR /go/src/compass
COPY ./ /go/src/compass
RUN BINDIR=/usr/local/bin APP=${APP} make static

# Stage 2 - Final Image
# The application should be statically linked
FROM alpine:3.6
ARG APP
ENV ENTRYPOINT ${APP}
RUN apk update && apk add --no-cache ca-certificates && rm -rf /var/cache/apk/*
COPY --from=builder /usr/local/bin/${APP} /usr/local/bin/${APP}
ENTRYPOINT $ENTRYPOINT
