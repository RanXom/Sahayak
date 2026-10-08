# syntax=docker/dockerfile:1
FROM golang:1.26-alpine AS builder
RUN apk add --no-cache git ca-certificates
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /sahayak ./cmd/server \
    && CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /migrate ./cmd/migrate \
    && CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /seed ./cmd/seed

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata ffmpeg
WORKDIR /app
COPY --from=builder /sahayak /app/sahayak
COPY --from=builder /migrate /app/migrate
COPY --from=builder /seed /app/seed
COPY --from=builder /app/migrations /app/migrations
COPY --from=builder /app/fonts /app/fonts
ENV DATA_DIR=/data \
    PDF_DIR=/data/pdfs \
    FONT_DIR=/app/fonts
RUN mkdir -p /data/pdfs
EXPOSE 8080
ENTRYPOINT ["/app/sahayak"]
