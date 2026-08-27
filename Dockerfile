# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates && \
    adduser -D -H -u 10001 app
WORKDIR /app
COPY --from=build --chown=app /out/server ./server
COPY --chown=app templates ./templates
COPY --chown=app static ./static
USER app
ENV HTTP_ADDR=:8080
EXPOSE 8080
CMD ["./server"]
