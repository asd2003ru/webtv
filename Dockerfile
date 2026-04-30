# syntax=docker/dockerfile:1

ARG GO_VERSION=1.26
ARG NODE_VERSION=22

FROM --platform=$BUILDPLATFORM node:${NODE_VERSION}-alpine AS web-build
WORKDIR /app/web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine AS go-build
ARG TARGETOS=linux
ARG TARGETARCH=amd64
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=web-build /app/internal/webui/dist ./internal/webui/dist
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /out/webtv ./cmd/webtv

FROM alpine:3.21
WORKDIR /app
RUN apk add --no-cache ffmpeg ca-certificates tzdata
COPY --from=go-build /out/webtv /usr/local/bin/webtv

ENV WEBTV_ADDR=:8080 \
    WEBTV_DB_PATH=/data/webtv.db \
    WEBTV_LOG_LEVEL=info \
    WEBTV_SINGLE_CONN_POLICY=reject \
    WEBTV_FFMPEG_BIN=ffmpeg \
    WEBTV_STREAM_IDLE_TIMEOUT_SEC=12 \
    WEBTV_STREAM_RETRY_MAX=5 \
    WEBTV_STREAM_MODE_CACHE_TTL_MIN=360 \
    WEBTV_STREAM_MANIFEST_BACKOFF_ENABLED=false

VOLUME ["/data"]
EXPOSE 8080
ENTRYPOINT ["webtv"]
