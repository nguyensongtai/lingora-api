-- Dữ liệu mẫu cho môi trường dev. Chạy lại được nhiều lần.
-- Id cố định để mọi máy dev cùng tham chiếu được một bộ dữ liệu.

INSERT INTO courses (id, slug, title, description, level, status, cover_image_url) VALUES
    (
        '01929f00-0000-7000-8000-000000000001',
        'ngu-phap-co-ban',
        'Ngữ pháp cơ bản',
        'Khoá nhập môn cho người mới bắt đầu: thì hiện tại, danh từ, mạo từ.',
        'A1',
        'published',
        NULL
    ),
    (
        '01929f00-0000-7000-8000-000000000002',
        'giao-tiep-hang-ngay',
        'Giao tiếp hàng ngày',
        'Mẫu câu và từ vựng cho các tình huống thường gặp.',
        'A2',
        'published',
        NULL
    ),
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

INSERT INTO lessons (id, course_id, slug, title, position) VALUES
    ('01929f01-0000-7000-8000-000000000001', '01929f00-0000-7000-8000-000000000001', 'thi-hien-tai-don', 'Thì hiện tại đơn', 3),
    ('01929f01-0000-7000-8000-000000000002', '01929f00-0000-7000-8000-000000000001', 'danh-tu-dem-duoc', 'Danh từ đếm được và không đếm được', 2),
    ('01929f01-0000-7000-8000-000000000003', '01929f00-0000-7000-8000-000000000001', 'mao-tu-a-an-the', 'Mạo từ a, an, the', 1),
    ('01929f01-0000-7000-8000-000000000004', '01929f00-0000-7000-8000-000000000002', 'chao-hoi-lam-quen', 'Chào hỏi và làm quen', 0),
    ('01929f01-0000-7000-8000-000000000005', '01929f00-0000-7000-8000-000000000002', 'goi-mon-o-nha-hang', 'Gọi món ở nhà hàng', 1)
ON CONFLICT (id) DO NOTHING;
