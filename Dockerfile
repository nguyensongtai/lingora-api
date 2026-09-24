# syntax=docker/dockerfile:1

# golang-migrate — đúng công cụ Makefile đã dùng ở máy dev (make migrate-up).
# Ghim phiên bản để lần deploy nào cũng chạy migration bằng cùng một bản.
FROM migrate/migrate:v4.18.3 AS migrate

FROM golang:1.26-alpine AS builder
WORKDIR /src
ENV CGO_ENABLED=0 GOTOOLCHAIN=local
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

# Alpine thay cho distroless: lệnh pre-deploy của Railway chạy qua shell, và
# distroless không có /bin/sh. Vẫn chạy bằng user không phải root.
FROM alpine:3.22
RUN addgroup -S app && adduser -S -G app app
COPY --from=migrate /usr/local/bin/migrate /usr/local/bin/migrate
COPY db/migrations /migrations
COPY --chmod=0755 deploy/migrate-up.sh /usr/local/bin/migrate-up
COPY --from=builder /out/api /api
USER app
EXPOSE 8080
# CMD chứ không phải ENTRYPOINT: Railway chạy lệnh pre-deploy (migrate-up) thay
# cho lệnh khởi động, và một ENTRYPOINT sẽ biến nó thành "/api migrate-up".
CMD ["/api"]
