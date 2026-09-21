# lingora-api

Backend của Lingora — web app học tiếng Anh. Frontend nằm ở repo
[lingora-web](https://github.com/nguyensongtai/lingora-web).

**Stack:** Go 1.26, chi, PostgreSQL 18 + sqlc, pgx/v5, golang-migrate.

## Chạy local

```bash
docker compose up -d                       # postgres:18 + redis:8
cp .env.example .env                       # rồi export các biến này
make migrate-up                            # áp dụng migration
make dev                                   # http://localhost:8080/healthz
```

## Layout

```
cmd/api/main.go            # entrypoint, DI thủ công
internal/
├── config/                # đọc env
├── course/                # domain: handler.go → service.go → repo.go
├── db/                    # code sqlc sinh ra (không sửa tay)
└── platform/postgres/     # pgx pool + helper uuid
db/migrations/             # golang-migrate
db/query/                  # input của sqlc
```

Mỗi domain là một package trong `internal/<domain>`, interface khai báo ở phía
consumer, sentinel error + wrap bằng `fmt.Errorf("%w")`, `context.Context`
truyền xuyên suốt.

## Công cụ

```bash
brew install go sqlc
go install -tags postgres github.com/golang-migrate/migrate/v4/cmd/migrate@latest
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
```

`make help` liệt kê toàn bộ target.
