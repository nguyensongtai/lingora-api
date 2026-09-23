# lingora-api

Backend của Lingora — web app học tiếng Anh cho người Việt. Frontend nằm ở repo
[lingora-web](https://github.com/nguyensongtai/lingora-web).

**Stack:** Go 1.26, chi v5, PostgreSQL 18 + sqlc, pgx/v5, golang-migrate,
Redis (chỉ để đếm hạn mức), JWT HS256, bcrypt.

---

## Mục lục

- [Chạy local](#chạy-local)
- [Cấu hình](#cấu-hình)
- [Mô hình nghiệp vụ](#mô-hình-nghiệp-vụ)
- [API](#api)
- [Xác thực](#xác-thực)
- [Hạn mức đăng nhập](#hạn-mức-đăng-nhập)
- [Lỗi](#lỗi)
- [Kiến trúc](#kiến-trúc)
- [Database](#database)
- [Sinh code](#sinh-code)
- [Kiểm thử](#kiểm-thử)
- [Công cụ](#công-cụ)

---

## Chạy local

```bash
make up                  # postgres:18 + redis:8 qua docker compose
cp .env.example .env     # sửa JWT_SECRET
make migrate-up          # áp dụng migration
make seed                # giáo trình mẫu: 13 khoá, 52 bài, 288 từ, 391 khối nội dung
make dev                 # http://localhost:8080/healthz
```

Tạo tài khoản đầu tiên — mật khẩu đọc từ stdin để không nằm lại trong lịch sử
shell hay trong danh sách tiến trình:

```bash
make createuser email=admin@lingora.vn role=admin
```

`make seed` nạp giáo trình mẫu phủ đủ sáu bậc CEFR — đủ khoá, bài và từ vựng để
mọi màn hình có gì hiển thị. Chạy lại được nhiều lần, không nhân đôi dữ liệu.

**Bậc CEFR của từng từ lấy từ dữ liệu công bố, không phải phán đoán**, và kiểm
lại được bằng một lệnh:

```bash
make verify-vocab
```

Lệnh này tải hai danh sách nguồn rồi đối chiếu với database, báo mọi từ đặt sai
bậc. Không có nó thì khẳng định "từ này là B1" chỉ là lời nói suông — không ai
nhìn bằng mắt mà kiểm được 288 từ.

Muốn xem màn Tiến độ và Luyện tập ở trạng thái đã có lịch sử thì tạo thêm một
tài khoản demo rồi nạp tiến độ mẫu cho nó — 24 bài trải trên 30 ngày:

```bash
make createuser email=demo@lingora.vn role=student
make seed-demo
```

`seed-demo` chỉ đụng tới đúng tài khoản đó. Chưa tạo thì nó không làm gì và
cũng không báo lỗi.

`make help` liệt kê toàn bộ target.

## Cấu hình

Đọc một lần lúc khởi động, trong `internal/config`. Thiếu biến bắt buộc thì
tiến trình **không boot** chứ không chạy với giá trị nửa vời.

| Biến | Bắt buộc | Mặc định | Việc |
| --- | --- | --- | --- |
| `DATABASE_URL` | ✅ | — | chuỗi kết nối pgx |
| `JWT_SECRET` | ✅ | — | tối thiểu 32 byte; `openssl rand -base64 48` |
| `PORT` | | `8080` | cổng lắng nghe |
| `APP_ENV` | | `development` | chỉ dùng để ghi log |
| `LOG_LEVEL` | | `info` | `debug` / `info` / `warn` / `error` |
| `CORS_ALLOWED_ORIGINS` | | `http://localhost:3000` | danh sách origin, phân tách bằng dấu phẩy |
| `ACCESS_TOKEN_TTL` | | `15m` | tuổi thọ access token |
| `REFRESH_TOKEN_TTL` | | `720h` | tuổi thọ refresh token |
| `REDIS_URL` | | — | kho đếm hạn mức; bỏ trống → đếm trong bộ nhớ |
| `GOOGLE_CLIENT_ID` | | — | bỏ trống → `POST /auth/google` trả 503 |
| `GOOGLE_CLIENT_SECRET` | | — | như trên |
| `TEST_DATABASE_URL` | | — | bỏ trống → integration test tự bỏ qua |

**`REDIS_URL` để trống là một lựa chọn, không phải mặc định an toàn.** Khi trống,
mỗi tiến trình đếm riêng, nên chạy *n* instance thì hạn mức thật là *n* lần con
số cấu hình. Tiến trình ghi rõ cảnh báo đó vào log lúc boot. `REDIS_URL` sai cú
pháp thì hỏng ngay lúc boot — không có đường nào lặng lẽ tắt hạn mức.

## Mô hình nghiệp vụ

```
Bậc CEFR (cột level)  →  Course  →  Lesson  →  VocabularyEntry
                                       │
                                       └── LessonProgress (user × lesson)
```

- **Course chính là "Unit"** của bản thiết kế. Không có bảng `units` ở giữa.
- **Thứ tự do người soạn đặt**: `courses.position` trong một bậc, `lessons.position`
  trong một khoá. Đổi bậc của một khoá thì nó về cuối bậc mới — position cũ là
  thứ tự trong bậc cũ, mang sang bậc khác thì vô nghĩa.
- **"Hoàn thành bài" = người học tự bấm**, không gắn với quiz hay ngưỡng điểm.
  Một dòng `lesson_progress` nghĩa là "người này đã đánh dấu bài này xong".
- **XP và streak không có bảng riêng** — suy ra từ `lesson_progress`: 20 XP/bài,
  mục tiêu 50 XP/ngày, cắt ngày theo giờ Việt Nam cố định (`Asia/Ho_Chi_Minh`).
  Streak vẫn sống nếu hôm nay chưa học nhưng hôm qua có.
- **Hàng đợi ôn từ vựng được *tính*, không được ghi**: `ListVocabularyForUser`
  join sang `lesson_progress`, nên học xong bài nào thì từ của bài đó vào hàng
  đợi — không có bước ghi nào, và bỏ đánh dấu bài thì từ tự rời hàng đợi.
- **Lịch ôn theo SM-2**: khoảng cách 1 → 6 → `round(trước × ease)` ngày, ease sàn
  1.3. Quên một từ thì `repetitions` về 0 nhưng **ease giữ nguyên** — một lần
  lỡ không nên phạt vĩnh viễn độ khó của từ. "Thành thạo" = khoảng cách ≥ 21 ngày.
- **`draft` là chặn theo người gọi, không phải bộ lọc**: khoá chưa xuất bản chỉ
  admin đọc được, kể cả qua id, qua slug, qua danh sách bài, hay đọc thẳng một
  bài. Người khác nhận 404 chứ không phải 403 — 403 đã xác nhận bản ghi tồn tại.
- **Xoá là xoá mềm** (`deleted_at`). Unique index đều là partial
  `WHERE deleted_at IS NULL`, nên slug của bản ghi đã xoá được dùng lại.

## API

Mọi route đều dưới tiền tố `/v1`. Hợp đồng đầy đủ nằm ở [`openapi.yaml`](openapi.yaml).

**Cột Quyền:** `—` công khai · `user` cần access token · `admin` cần access
token có role `admin`.

### Xác thực

| Method | Route | Quyền | Ghi chú |
| --- | --- | --- | --- |
| `POST` | `/auth/register` | — | luôn tạo role `student` |
| `POST` | `/auth/login` | — | có hạn mức, xem bên dưới |
| `POST` | `/auth/google` | — | 503 nếu chưa cấu hình Google |
| `POST` | `/auth/refresh` | — | refresh token nằm trong body |
| `POST` | `/auth/logout` | — | thu hồi refresh token trong body |
| `GET` | `/auth/me` | user | |

### Khoá học và bài học

| Method | Route | Quyền |
| --- | --- | --- |
| `GET` | `/courses` | — ¹ |
| `GET` | `/courses/{courseId}` | — ¹ |
| `GET` | `/courses/by-slug/{slug}` | — ¹ |
| `GET` | `/courses/{courseId}/lessons` | — ¹ |
| `GET` | `/lessons/{lessonId}` | — ¹ |
| `POST` | `/courses` | admin |
| `PATCH` | `/courses/{courseId}` | admin |
| `DELETE` | `/courses/{courseId}` | admin |
| `PUT` | `/courses/order` | admin |
| `POST` | `/courses/{courseId}/lessons` | admin |
| `PUT` | `/courses/{courseId}/lessons/order` | admin |
| `PATCH` | `/lessons/{lessonId}` | admin |
| `PUT` | `/lessons/{lessonId}/blocks` | admin |
| `DELETE` | `/lessons/{lessonId}` | admin |

¹ Năm route đọc này công khai nhưng **nội dung phụ thuộc người gọi**: gửi kèm
token admin thì thấy cả khoá `draft`, không gửi hoặc gửi token thường thì chỉ
thấy `published`. Token hỏng hay hết hạn bị coi như không có chứ không thành
401, nên một tab bỏ quên không biến trang công khai thành lỗi.

`GET /courses` nhận `status`, `level`, `page_size` (mặc định 20, **trần 100** —
vượt trần thì bị cắt im lặng chứ không lỗi) và `cursor`. Tham số `status` chỉ có
tác dụng với admin; người khác xin `draft` sẽ nhận một trang rỗng.

Phân trang là **keyset** trên `(created_at DESC, id DESC)`, không phải offset:
`cursor` là chuỗi mờ base64url lấy từ `next_cursor` của trang trước, `next_cursor`
rỗng nghĩa là hết dữ liệu. Chọn keyset vì offset sẽ nhảy hoặc lặp bản ghi khi có
khoá mới chen vào giữa lúc người dùng đang lật trang.

`GET /lessons/{id}` trả bài kèm **nội dung** và **từ vựng của bài**. Nội dung là
một danh sách khối có thứ tự, mỗi khối là `note` (đoạn giải thích tiếng Việt),
`example` (câu mẫu Anh–Việt) hoặc `dialogue` (như `example`, kèm tên người nói).
Mỗi dạng chỉ dùng một phần các trường, và ràng buộc `lesson_blocks_shape` trong
database bắt đúng phần đó phải có còn phần thừa phải rỗng.

`vocabulary` là những từ bài này dạy, trả cho mọi người đọc được bài — khác với
`/me/vocabulary`, vốn chỉ gồm từ của bài đã học xong. Trước đây trường này không
có, nên người học chỉ gặp từ của một bài *sau khi* bấm "Đã xong", ở màn Từ vựng.
`course.Service` đọc từ vựng qua một interface khai báo phía nó, và chỉ đọc
**sau** bước kiểm quyền: bài của khoá nháp dừng ở 404 trước khi chạm tới từ.

Màn học chia bài thành bước (Từ vựng → Hội thoại → Cách dùng → Luyện tập) từ
chính dữ liệu này — API không biết gì về bước, và không có cột nào để giữ đồng bộ.

`PUT /lessons/{id}/blocks` thay **cả danh sách**, giống hai endpoint `order`:
soạn bài là việc viết lại và kéo thả, nên gửi trọn trạng thái mong muốn đơn
giản hơn một chuỗi thêm/sửa/xoá/đổi chỗ. Gửi mảng rỗng là xoá sạch nội dung bài.

Hai endpoint `order` nhận **toàn bộ** mảng id của phạm vi đang sắp — cả bậc với
`/courses/order`, cả khoá với `/lessons/order`. Thiếu một id, thừa một id, hay
lẫn id của phạm vi khác đều là 400. Gửi đủ tập như vậy khiến yêu cầu tự mô tả
được trạng thái nó mong đợi, nên hai admin sắp cùng lúc không ghi đè lẫn nhau âm
thầm. Server đánh lại position bằng một câu `UPDATE ... unnest WITH ORDINALITY`,
nên không có trạng thái trung gian trùng position.

### Tiến độ

| Method | Route | Quyền |
| --- | --- | --- |
| `GET` | `/me/progress` | user |
| `GET` | `/me/progress/history` | user |
| `PUT` | `/me/progress/lessons/{lessonId}` | user |
| `DELETE` | `/me/progress/lessons/{lessonId}` | user |

`GET /me/progress` trả XP hôm nay, mục tiêu, `xp_per_lesson`, streak, hoạt động
7 ngày, tiến độ từng khoá và khoá học gần nhất. `GET /me/progress/history` trả
dải dài ngày (mặc định 30, trần 365) kèm tổng của cả đời — `longest_streak` là
dài nhất **trong phạm vi được hỏi**, không phải kỷ lục mọi thời. `PUT` là idempotent
(`ON CONFLICT DO NOTHING`) — bấm hai lần không cộng XP hai lần.

### Luyện tập

| Method | Route | Quyền |
| --- | --- | --- |
| `GET` | `/me/practice/session` | user |
| `GET` | `/me/practice/session?lesson_id=…` | user |
| `POST` | `/me/practice/answers` | user |

Câu hỏi **sinh tại chỗ** từ `vocabulary_entries`, không có bảng câu hỏi nào để
soạn. `internal/practice` vì thế không có repo riêng: nó đi qua
`vocabulary.Service`, nên quy tắc "chỉ ôn từ của bài đã học xong" tự có hiệu
lực mà không phải chép lại, và từ chưa mở khoá đơn giản là không có trong danh
sách nên không cần một lần kiểm quyền thứ hai.

Chấm điểm **chỉ phạt khi sai**: câu sai đẩy từ về đầu hàng đợi ôn, câu đúng
không kéo dài khoảng cách. SM-2 dựa trên việc nhớ được sau một quãng nghỉ, nên
luyện dồn một buổi không được biến một từ vừa gặp thành "thành thạo".

**Luyện trong bài** (`lesson_id` ở cả hai endpoint) là bước cuối của màn học:
câu hỏi lấy từ đúng những từ của bài đó, **kể cả khi chưa học xong** — đó là lúc
cần luyện nhất. Nó hỏi hết từ của bài chứ không theo `size`, và **không đụng tới
lịch ôn**, kể cả khi sai và kể cả khi bài đã xong: đây là lần gặp đầu, không phải
ôn (người dùng đã chốt như vậy). Quyền đọc bài đi qua `course.Service`, nên bài
của khoá nháp là 404 với người học y như `GET /lessons/{id}`.

Thiếu `lesson_id` ở `POST /answers` thì câu trả lời bị chấm như ôn tập và câu
sai bị phạt — phía trước phải gửi đúng phạm vi của lượt đang luyện.

`listen_choose` phải gửi cả chữ của từ xuống để trình duyệt đọc lên, nghĩa là
đáp án nằm sẵn trong trang — hệ quả không tránh được khi dùng `speechSynthesis`
thay vì file audio.

### Từ vựng

| Method | Route | Quyền |
| --- | --- | --- |
| `GET` | `/vocabulary?lesson_id=…` | admin |
| `POST` | `/vocabulary` | admin |
| `PATCH` | `/vocabulary/{entryId}` | admin |
| `DELETE` | `/vocabulary/{entryId}` | admin |
| `GET` | `/me/vocabulary` | user |
| `GET` | `/me/vocabulary/stats` | user |
| `POST` | `/me/vocabulary/{entryId}/review` | user |

Cả nhánh `/vocabulary` là **công cụ soạn nội dung**, không phải chỗ người học
đọc. Người học đọc ở `/me/vocabulary`, nơi hàng đợi được lọc theo bài họ đã học
xong — mở `/vocabulary` cho người ngoài là phát không toàn bộ từ và nghĩa, và
vô hiệu hoá luôn quy tắc mở khoá đó.

`POST /me/vocabulary/{entryId}/review` nhận `grade` là `remembered` hoặc
`forgot`, chạy SM-2 và trả về thẻ đã lên lịch lại. Endpoint từ chối những từ
thuộc bài người dùng chưa học xong.

### Vận hành

| Method | Route | Việc |
| --- | --- | --- |
| `GET` | `/healthz` | tiến trình còn sống (không chạm database) |
| `GET` | `/readyz` | ping database; hỏng → 503 `degraded` |

Mỗi request hoàn tất để lại đúng một dòng log JSON kèm `request_id`, `status`
và `took_ms`. Mức log theo status — 5xx là `ERROR`, 4xx là `WARN` (một tràng
401 là dấu hiệu dò mật khẩu), còn lại `INFO`. `/healthz` và `/readyz` không
được ghi: chúng bị gọi vài giây một lần và sẽ nhấn chìm mọi thứ khác. Query
string cũng không được ghi — nó do client đặt, còn log thì thường được chuyển
sang nơi khác lưu.

## Xác thực

Access token là **JWT HS256**, sống 15 phút, gửi qua `Authorization: Bearer`.
Khi kiểm tra có bắt buộc `WithValidMethods` và `WithExpirationRequired` — thiếu
hai cái đó thì một token `alg: none` hoặc không hạn sẽ lọt.

Refresh token là **chuỗi ngẫu nhiên mờ**, không phải JWT: database chỉ lưu
SHA-256 của nó, nên rò database không tự động thành rò phiên. Mỗi lần refresh
token cũ bị thu hồi và cấp token mới (rotation).

Mật khẩu băm bằng **bcrypt cost 12**. Email không tồn tại vẫn bị so với một hash
giả để thời gian trả lời không tiết lộ email nào có trong hệ thống; tài khoản
đăng nhập bằng Google (không có mật khẩu) cũng đi qua đúng đường đó.

**Google** chạy luồng authorization code phía server. `id_token` **cố ý không
kiểm chữ ký** — nó tới thẳng từ endpoint của Google qua TLS, không đi qua trình
duyệt — nhưng `iss` và `aud` vẫn được kiểm. Nhờ vậy không phải thêm thư viện
OAuth nào. Quy tắc gắn tài khoản: `google_sub` đã gắn → vào thẳng; email trùng
**và Google đã xác minh email đó** → gắn vào tài khoản cũ; còn lại → tạo tài
khoản mới không mật khẩu. `users.password_hash` vì thế nullable, và CHECK
`users_has_credential` đảm bảo mỗi tài khoản có ít nhất một cách đăng nhập.

## Hạn mức đăng nhập

Hai lớp, chặn hai kiểu tấn công khác nhau:

| Lớp | Hạn mức | Chặn |
| --- | --- | --- |
| Theo **IP** (middleware) | 20 / 15 phút trên `/auth/register`, `/login`, `/google` | một máy dò nhiều tài khoản |
| Theo **email** (tầng nghiệp vụ) | 5 / 15 phút, chỉ trên `/auth/login` | nhiều máy cùng dò một tài khoản |

Lớp email phải nằm ở service vì chỉ ở đó mới biết email người dùng gõ; khoá đếm
được chuẩn hoá nên viết hoa hay thường đều vào chung một bộ đếm. Đăng nhập đúng
thì xoá bộ đếm.

`/refresh` và `/logout` **không** bị bọc: hai route đó đã cầm sẵn một token hợp
lệ nên không phải cửa để dò, còn chặn chúng thì người mở nhiều tab bị khoá oan.

Vượt hạn mức → `429` kèm header `Retry-After`. Nếu Redis chết, middleware **mở
cửa** (cho qua) và ghi lỗi — mất hạn mức còn hơn mất cả trang đăng nhập.

## Lỗi

Mọi lỗi do handler sinh ra dùng chung một body:

```json
{
  "code": "validation_error",
  "message": "Dữ liệu không hợp lệ.",
  "details": { "slug": "không được rỗng" }
}
```

`details` là bản đồ field → lý do, chỉ có ở lỗi validation, để form phía client
gắn thông báo vào đúng ô mà không phải đoán từ `message`.

| Code | HTTP |
| --- | --- |
| `validation_error` | 400 |
| `malformed_body` | 400 |
| `unauthorized` | 401 |
| `forbidden` | 403 |
| `not_found` | 404 |
| `method_not_allowed` | 405 |
| `conflict` | 409 |
| `rate_limited` | 429 |
| `internal_error` | 500 |

Đường dẫn sai và method sai cũng dùng đúng body này: handler mặc định của chi
trả text thuần, nên chúng được thay bằng `router.NotFound` và
`router.MethodNotAllowed` riêng.

`internal_error` không bao giờ lộ chi tiết lỗi ra ngoài; chi tiết chỉ nằm trong
log của server.

## Kiến trúc

```
cmd/
├── api/           # entrypoint, DI thủ công, wiring route
├── createuser/    # tạo tài khoản, mật khẩu đọc từ stdin
└── devtoken/      # in access token admin để gọi thử
internal/
├── api/           # model sinh từ openapi.yaml + contract_test.go
├── auth/          # đăng nhập, JWT, refresh token, Google
├── config/        # đọc env một lần lúc boot
├── course/        # khoá học và bài học
├── db/            # code sqlc sinh ra — không sửa tay
├── httpx/         # JSON, CORS, middleware hạn mức
├── platform/      # pgx pool, helper uuid
├── practice/      # phiên luyện tập, không có bảng riêng
├── progress/      # tiến độ, XP, streak
├── ratelimit/     # interface Limiter + bản memory và redis
├── testdb/        # helper integration test
├── user/          # model người dùng
└── vocabulary/    # từ vựng và SM-2
db/migrations/     # golang-migrate, đánh số tuần tự
db/query/          # input của sqlc
db/seed/           # giáo trình mẫu cho dev
```

Mỗi domain là một package trong `internal/<domain>`, xếp ba tầng
**`handler.go` → `service.go` → `repo.go`**:

- `handler.go` chỉ làm HTTP: parse, validate cú pháp, map lỗi sang status.
- `service.go` giữ nghiệp vụ, không biết gì về `net/http`.
- `repo.go` bọc code sqlc, dịch model database sang model nghiệp vụ.

**Interface khai báo ở phía consumer**, không phải phía implementation — nhờ vậy
`service` chỉ nêu đúng những method nó cần, và package bị phụ thuộc không phải
biết ai đang dùng mình. Lỗi là sentinel error, bọc bằng `fmt.Errorf("%w")`.
`context.Context` truyền xuyên suốt. Toàn bộ dây nối nằm tay ở `cmd/api/main.go`
— không có container DI.

Model nghiệp vụ dùng `string` cho id chứ không phải `pgtype.UUID`, để tầng trên
không phải nhập khẩu kiểu của pgx.

## Database

PostgreSQL 18. Khoá chính là `uuidv7()` — sinh trong database, gần như tăng dần
theo thời gian nên chèn không phân mảnh index như uuid v4.

Migration chạy bằng golang-migrate, mỗi bước một cặp `up`/`down`:

| # | Việc |
| --- | --- |
| 000001 | `courses` |
| 000002 | `lessons` |
| 000003 | `users` + `refresh_tokens` |
| 000004 | `lesson_progress` |
| 000005 | danh tính Google: `password_hash` nullable, `google_sub` |
| 000006 | `vocabulary_entries` + `vocabulary_reviews` |
| 000007 | `courses.position` |

```bash
make migrate-up
make migrate-down                    # lùi đúng một bước
make migrate-create name=add_foo
```

Migration đã lên production thì **không sửa tại chỗ** — viết migration mới.

## Sinh code

Hai nguồn sinh code, cả hai đều phải commit kèm:

```bash
make sqlc       # db/query/*.sql  → internal/db
make openapi    # openapi.yaml    → internal/api
```

CI chạy lại cả hai rồi `git diff --exit-code`, nên quên chạy là đỏ build. Không
sửa tay file trong `internal/db` và `internal/api`.

`internal/api/contract_test.go` kiểm hai chiều mà compiler không thấy được: mọi
hằng `httpx.Code*` phải có mặt trong enum `ErrorCode` của spec, và enum trong
migration phải khớp enum trong spec. Thiếu là test đỏ, không phải phát hiện
trên production.

## Kiểm thử

```bash
make test               # unit test, có -race
make test-integration   # thêm cả integration test tầng repo
make check              # gofmt -l, go vet, go build
```

Integration test chạy thật với PostgreSQL: mỗi test nhận một `pgx.Tx` riêng và
**rollback khi xong**, nên chúng không thấy nhau và không để lại gì trong
database. Thiếu `TEST_DATABASE_URL` thì tự `t.Skip`.

Một chi tiết dễ vấp: PostgreSQL huỷ cả transaction ngay khi một câu lệnh lỗi,
nên test nào cố tình vi phạm ràng buộc phải bọc thao tác đó trong savepoint —
đó là việc của `testdb.Attempt`.

CI dựng sẵn service `postgres:18-alpine` và `redis:8-alpine`, nên integration
test chạy thật trên CI chứ không âm thầm bị bỏ qua.

## Nguồn dữ liệu từ vựng

Bậc CEFR trong `db/seed/004_vocabulary.sql` lấy từ hai bộ dữ liệu mở. Cả hai
đều **bắt buộc trích dẫn**:

| Bậc | Nguồn | Giấy phép |
| --- | --- | --- |
| A1–B2 | The CEFR-J Wordlist Version 1.5, Yukio Tono, Tokyo University of Foreign Studies | dùng được cho nghiên cứu và thương mại, kèm trích dẫn |
| C1–C2 | Octanove Vocabulary Profile C1/C2 version 1.0, Octanove Labs | CC BY-SA 4.0 |

Cả hai phân phối qua [openlanguageprofiles/olp-en-cefrj](https://github.com/openlanguageprofiles/olp-en-cefrj).

**Octanove là CC BY-SA**, nên phần C1/C2 của danh sách từ ở đây là tác phẩm
phái sinh và phải giữ cùng giấy phép. Nếu sau này cần một giáo trình thương mại
đóng, phần C1/C2 phải thay bằng nguồn khác — phần A1–B2 thì không vướng.

**Chỉ BẬC là có nguồn.** Phiên âm, nghĩa tiếng Việt và câu ví dụ do Claude soạn
và **chưa được người có chuyên môn rà soát**. Trước khi có người học thật, phần
đó cần một vòng biên tập.

## Nội dung bài học

| File | Nội dung |
| --- | --- |
| `005_lesson_content.sql` | phần dẫn, ghi chú và câu mẫu của 48 bài; hội thoại của 9 bài |
| `006_lesson_dialogues.sql` | hội thoại cho 39 bài còn lại (162 lượt); chạy **sau** 005 |

Mỗi bài đã xuất bản đều có đủ bốn bước. Câu mẫu minh hoạ **điểm của ghi chú**,
không lặp lại câu ví dụ của từ vựng — bản đầu lấy thẳng câu ví dụ của từ (122/132
câu), nên khi bài chia bước thì bước Cách dùng thành bản sao của bước Từ vựng.
Màn học giờ tự ẩn câu mẫu trùng như vậy.

**Toàn bộ phần này không có nguồn đối chiếu bằng máy** — khác với bậc CEFR của
từ vựng. Giải thích ngữ pháp, dựng hội thoại và chọn câu mẫu là việc sư phạm, do
Claude soạn, và cần người dạy tiếng Anh đọc lại trước khi có người học thật.

## Công cụ

```bash
brew install go sqlc
go install -tags postgres github.com/golang-migrate/migrate/v4/cmd/migrate@latest
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
```
