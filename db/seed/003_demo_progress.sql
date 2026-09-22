-- Tiến độ mẫu cho tài khoản demo@lingora.vn: đủ lịch sử để màn Tiến độ có
-- biểu đồ thật và màn Luyện tập có đủ từ đã mở khoá.
--
-- Cố ý KHÔNG đụng tới tài khoản nào khác. Chưa tạo demo@lingora.vn thì file
-- này không làm gì cả, và cũng không báo lỗi:
--
--     make createuser email=demo@lingora.vn role=student
--     make seed-demo
--
-- Chạy lại được nhiều lần; bài đã đánh dấu thì giữ nguyên mốc hoàn thành cũ.

INSERT INTO lesson_progress (user_id, lesson_id, completed_at)
SELECT
    demo.id,
    ordered.lesson_id,
    -- Trải 24 bài đều trên 30 ngày, bài mới nhất rơi vào hôm nay. Số bài ít
    -- hơn số ngày nên tự nhiên có ngày trống — biểu đồ cần cả khoảng trống mới
    -- giống dữ liệu thật.
    -- AT TIME ZONE lần hai là bắt buộc: lần đầu cho ra một mốc KHÔNG có múi
    -- giờ, và cột timestamptz sẽ đọc nó theo giờ máy chủ. Thiếu nó thì trên
    -- máy chủ chạy giờ UTC, buổi tối giờ Việt Nam rơi sang ngày hôm sau.
    (
        date_trunc('day', now() AT TIME ZONE 'Asia/Ho_Chi_Minh')
            - make_interval(days => ((ordered.total - 1 - ordered.n) * 30 / greatest(ordered.total - 1, 1))::integer)
            + interval '20 hours'
    ) AT TIME ZONE 'Asia/Ho_Chi_Minh'
FROM (
    SELECT id FROM users WHERE lower(email) = 'demo@lingora.vn' AND deleted_at IS NULL
) AS demo
CROSS JOIN (
    SELECT
        l.id AS lesson_id,
        (row_number() OVER w - 1)::integer AS n,
        count(*) OVER ()::integer AS total
    FROM lessons AS l
    JOIN courses AS c ON c.id = l.course_id
    WHERE l.deleted_at IS NULL
      AND c.deleted_at IS NULL
      AND c.status = 'published'
      -- Chỉ ba bậc đầu: người học demo đang ở giữa lộ trình, không phải đã
      -- xong hết — như vậy màn Lộ trình còn chỗ để hiện phần chưa học.
      AND c.level IN ('A1', 'A2', 'B1')
    WINDOW w AS (ORDER BY c.level, c.position, l.position)
) AS ordered
ON CONFLICT (user_id, lesson_id) DO NOTHING;
