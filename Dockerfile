# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server && \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/ingest-ted ./cmd/ingest-ted && \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/jobs ./cmd/jobs

FROM alpine:3.20 AS supercronic
ARG TARGETARCH
ARG SUPERCRONIC_VERSION=v0.2.48
RUN apk add --no-cache curl && \
    case "$TARGETARCH" in \
      amd64) SUPERCRONIC_SHA1SUM=016b7c9aebfc8d9fd9526e8ba33b191fc524485f ;; \
      arm64) SUPERCRONIC_SHA1SUM=2ab9b3bdcf290f60b59700aad876b6e68f3a6b06 ;; \
      *) echo "Unsupported architecture: $TARGETARCH" >&2; exit 1 ;; \
    esac && \
    curl -fsSL "https://github.com/aptible/supercronic/releases/download/$SUPERCRONIC_VERSION/supercronic-linux-$TARGETARCH" -o /supercronic && \
    echo "$SUPERCRONIC_SHA1SUM  /supercronic" | sha1sum -c - && \
    chmod +x /supercronic

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata && \
    adduser -D -H -u 10001 app
WORKDIR /app
COPY --from=build --chown=app /out/server ./server
COPY --from=build --chown=app /out/ingest-ted ./ingest-ted
COPY --from=build --chown=app /out/jobs ./jobs
COPY --from=supercronic --chown=app /supercronic /usr/local/bin/supercronic
COPY --chown=app templates ./templates
COPY --chown=app static ./static
COPY --chown=app crontab ./crontab
USER app
ENV HTTP_ADDR=:8080
EXPOSE 8080
CMD ["./server"]
