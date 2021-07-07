ARG GO_VERSION=1.16

FROM golang:${GO_VERSION}-alpine3.13 AS builder

RUN apk add --update --no-cache ca-certificates~=20191127 make~=4.3 git~=2.30 curl~=7.77

ARG PACKAGE=/build

RUN mkdir -p /${PACKAGE}
WORKDIR /${PACKAGE}

COPY Makefile /${PACKAGE}/

COPY . /${PACKAGE}
RUN BUILD_DIR='' BINARY_NAME=app make build-release


FROM alpine:3.7
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app /app
USER nobody:nobody
CMD ["/app"]
