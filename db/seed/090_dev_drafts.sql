-- Chỉ cho môi trường dev: một khoá NHÁP có bài nhưng chưa có nội dung, để thử
-- những gì chỉ admin thấy (nhãn "Bản nháp", 404 với người học). make seed nạp
-- file này; make seed-prod thì không — production không cần một khoá rỗng.
--
-- Chạy SAU 001: dùng chung cách tra bài theo (khoá, slug).

INSERT INTO courses (id, slug, title, description, level, status, cover_image_url) VALUES
    (
        '01929f00-0000-7000-8000-000000000003',
        'luyen-thi-ielts-writing',
        'Luyện thi IELTS Writing',
        'Task 1 và Task 2: cấu trúc bài, từ vựng học thuật, lỗi thường gặp.',
        'B2',
        'draft',
        NULL
    )
ON CONFLICT (id) DO NOTHING;

INSERT INTO lessons (course_id, slug, title, position)
SELECT c.id, v.slug, v.title, v.position
FROM (VALUES
    ('task-1-bieu-do', 'Task 1: mô tả biểu đồ', 0),
    ('task-2-cau-truc', 'Task 2: cấu trúc bài', 1),
    ('tu-noi-hoc-thuat', 'Từ nối học thuật', 2),
    ('loi-thuong-gap', 'Lỗi thường gặp', 3)
) AS v(slug, title, position)
JOIN courses AS c ON c.slug = 'luyen-thi-ielts-writing' AND c.deleted_at IS NULL
ON CONFLICT (course_id, slug) WHERE deleted_at IS NULL DO NOTHING;
