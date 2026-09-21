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

## API

| Method | Route | Quyền |
| --- | --- | --- |
| `GET` | `/v1/courses` | công khai |
| `GET` | `/v1/courses/{id}` | công khai |
| `GET` | `/v1/courses/by-slug/{slug}` | công khai |
| `GET` | `/v1/courses/{id}/lessons` | công khai |
| `POST` | `/v1/courses` | `Bearer` token role `admin` |
| `PATCH` | `/v1/courses/{id}` | `Bearer` token role `admin` |
| `DELETE` | `/v1/courses/{id}` | `Bearer` token role `admin` |

`GET /v1/courses` nhận `status`, `level`, `page_size` (mặc định 20, trần 100) và
`cursor`. Con trỏ là chuỗi mờ base64url, lấy từ `next_cursor` của trang trước.

Mọi lỗi dùng chung một body:

```json
{ "code": "validation_error", "message": "Dữ liệu không hợp lệ.", "details": { "slug": "không được rỗng" } }
```

| Code | HTTP |
| --- | --- |
| `validation_error` | 400 |
| `malformed_body` | 400 |
| `unauthorized` | 401 |
| `forbidden` | 403 |
| `not_found` | 404 |
| `conflict` | 409 |
| `internal_error` | 500 |

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
