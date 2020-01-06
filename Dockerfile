ARG GO_VERSION=1.12

FROM golang:${GO_VERSION}-alpine AS builder

RUN apk add --update --no-cache ca-certificates=20191127-r0 make=4.2.1-r2 git=2.24.1-r0 curl=7.67.0-r0

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
