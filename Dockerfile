FROM oven/bun:1-alpine AS asset-builder
WORKDIR /app

COPY package.json bun.lock ./
ENV NODE_ENV=production
RUN bun install --frozen-lockfile

COPY . .

RUN bun run build

FROM golang:1.25-alpine AS binary-builder
WORKDIR /app

RUN apk add --no-cache git make

RUN go install github.com/a-h/templ/cmd/templ@v0.3.960

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN templ generate

ENV CGO_ENABLED=0 GOOS=linux GOARCH=arm64
RUN go build -ldflags="-s -w" -o server ./cmd/server

FROM alpine:latest
WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

RUN mkdir -p uploads static && \
    chown -R appuser:appgroup /app

COPY --from=binary-builder --chown=appuser:appgroup /app/server .

COPY --from=asset-builder --chown=appuser:appgroup /app/static ./static
USER appuser

EXPOSE 7899

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:7899/health || exit 1

ENTRYPOINT ["./server"]
