ARG GO_VERSION=1.18.0
ARG GID=1000
ARG UID=1000

FROM golang:${GO_VERSION}-alpine3.15 AS builder
ARG GID
ARG UID

# Create user and group
RUN addgroup -g ${GID} -S appgroup && \
    adduser -u ${UID} -S appuser -G appgroup

RUN apk add --update --no-cache ca-certificates~=20220614 make~=4.3 git~=2.34 curl~=7.80

ARG PACKAGE=/build

RUN mkdir -p /${PACKAGE}
WORKDIR /${PACKAGE}

COPY Makefile /${PACKAGE}/

COPY . /${PACKAGE}
RUN BUILD_DIR='' BINARY_NAME=app make build-release


# hadolint ignore=DL3007
FROM gcr.io/distroless/static:latest
ARG GID
ARG UID

COPY --from=builder /app /app
ENV GIN_MODE=release

COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
USER ${UID}:${GID}

CMD ["/app"]
