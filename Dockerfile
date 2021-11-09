ARG GO_VERSION=1.17.2

FROM golang:${GO_VERSION}-alpine3.13 AS builder

RUN apk add --update --no-cache ca-certificates~=20191127 make~=4.3 git~=2.30 curl~=7.79

ARG PACKAGE=/build

RUN mkdir -p /${PACKAGE}
WORKDIR /${PACKAGE}

COPY Makefile /${PACKAGE}/

COPY . /${PACKAGE}
RUN BUILD_DIR='' BINARY_NAME=app make build-release


# hadolint ignore=DL3007
FROM gcr.io/distroless/static:latest
COPY --from=builder /app /app
ENV GIN_MODE=release
USER nobody:nobody
CMD ["/app"]
