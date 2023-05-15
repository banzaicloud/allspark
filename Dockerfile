ARG GO_VERSION=1.18.0

FROM golang:${GO_VERSION}-alpine3.15 AS builder

# hadolint ignore=DL3018
RUN apk add --update --no-cache ca-certificates make git curl

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
