-- Giáo trình mẫu cho môi trường dev: đủ khoá ở cả sáu bậc CEFR và đủ bài để
-- các màn Lộ trình, Từ vựng, Luyện tập và Tiến độ có gì để hiển thị. Từ vựng
-- nằm ở 004, nội dung bài ở 005 và 006.
--
-- Chạy lại được nhiều lần. Không dùng id cố định: khoá tra theo slug, bài tra
-- theo (khoá, slug) — đúng các unique index đang có.
--
-- 001_sample_courses.sql tạo ba khoá đầu; file này bổ sung phần còn lại và
-- thêm bài cho chính ba khoá đó.

/* ---------- khoá học ---------- */

INSERT INTO courses (slug, title, description, level, status) VALUES
    ('tu-vung-doi-song', 'Từ vựng đời sống',
     'Những từ gặp mỗi ngày: gia đình, màu sắc, đồ ăn, thời gian.', 'A1', 'published'),
    ('qua-khu-va-tuong-lai', 'Quá khứ và tương lai',
     'Kể lại chuyện đã qua và nói về dự định sắp tới.', 'A2', 'published'),
    ('ke-chuyen-va-mo-ta', 'Kể chuyện và mô tả',
     'Tả người, tả nơi chốn, kể một sự việc và gọi tên cảm xúc.', 'B1', 'published'),
    ('email-va-tin-nhan', 'Email và tin nhắn',
     'Viết thư trang trọng và thân mật, hẹn lịch, xin lỗi, cảm ơn.', 'B1', 'published'),
    ('tranh-luan-va-y-kien', 'Tranh luận và nêu ý kiến',
     'Nêu quan điểm, phản biện, dẫn chứng và chốt lại vấn đề.', 'B2', 'published'),
    ('tieng-anh-cong-so', 'Tiếng Anh công sở',
     'Họp hành, báo cáo tiến độ, thương lượng và phản hồi đồng nghiệp.', 'B2', 'published'),
    ('hoc-thuat-va-nghien-cuu', 'Học thuật và nghiên cứu',
     'Trích dẫn nguồn, mô tả dữ liệu và nói về giới hạn của nghiên cứu.', 'C1', 'published'),
    ('sac-thai-va-thanh-ngu', 'Sắc thái và mức độ trang trọng',
     'Tính từ mạnh, động từ trang trọng thay cụm động từ, nói giảm nói tránh.', 'C1', 'published'),
    ('van-phong-va-tu-tu', 'Văn phong và tu từ',
     'Biện pháp tu từ, nhịp điệu câu văn và vốn từ tinh tế.', 'C2', 'published'),
    ('dien-thuyet-va-thuyet-phuc', 'Diễn thuyết và thuyết phục',
     'Mở đầu ấn tượng, dựng lập luận và giữ người nghe tới cuối.', 'C2', 'published')
ON CONFLICT (slug) WHERE deleted_at IS NULL DO NOTHING;

/* ---------- bài học ---------- */

INSERT INTO lessons (course_id, slug, title, position)
SELECT c.id, v.slug, v.title, v.position
FROM (VALUES
    -- A1 · Ngữ pháp cơ bản (ba bài còn lại đã có ở 001)
    ('ngu-phap-co-ban', 'dong-tu-to-be', 'Động từ to be', 0),

    -- A1 · Từ vựng đời sống
    ('tu-vung-doi-song', 'gia-dinh-va-ban-be', 'Gia đình và bạn bè', 0),
    ('tu-vung-doi-song', 'mau-sac-va-hinh-dang', 'Màu sắc và hình dáng', 1),
    ('tu-vung-doi-song', 'do-an-va-thuc-uong', 'Đồ ăn và thức uống', 2),
    ('tu-vung-doi-song', 'thoi-gian-va-ngay-thang', 'Thời gian và ngày tháng', 3),

    -- A2 · Giao tiếp hàng ngày (hai bài đầu đã có ở 001)
    ('giao-tiep-hang-ngay', 'hoi-duong-di', 'Hỏi đường', 2),
    ('giao-tiep-hang-ngay', 'mua-sam', 'Mua sắm', 3),

    -- A2 · Quá khứ và tương lai
    ('qua-khu-va-tuong-lai', 'thi-qua-khu-don', 'Thì quá khứ đơn', 0),
    ('qua-khu-va-tuong-lai', 'dong-tu-bat-quy-tac', 'Động từ bất quy tắc', 1),
    ('qua-khu-va-tuong-lai', 'noi-ve-du-dinh', 'Nói về dự định', 2),
    ('qua-khu-va-tuong-lai', 'ke-ve-ky-nghi', 'Kể về kỳ nghỉ', 3),

    -- B1 · Kể chuyện và mô tả
    ('ke-chuyen-va-mo-ta', 'mo-ta-nguoi', 'Mô tả một người', 0),
    ('ke-chuyen-va-mo-ta', 'mo-ta-noi-chon', 'Mô tả nơi chốn', 1),
    ('ke-chuyen-va-mo-ta', 'ke-mot-su-viec', 'Kể lại một sự việc', 2),
    ('ke-chuyen-va-mo-ta', 'cam-xuc-va-thai-do', 'Cảm xúc và thái độ', 3),

    -- B1 · Email và tin nhắn
    ('email-va-tin-nhan', 'email-trang-trong', 'Email trang trọng', 0),
    ('email-va-tin-nhan', 'email-than-mat', 'Email thân mật', 1),
    ('email-va-tin-nhan', 'hen-lich-hop', 'Hẹn lịch họp', 2),
    ('email-va-tin-nhan', 'xin-loi-va-cam-on', 'Xin lỗi và cảm ơn', 3),

    -- B2 · Tranh luận và nêu ý kiến
    ('tranh-luan-va-y-kien', 'neu-quan-diem', 'Nêu quan điểm', 0),
    ('tranh-luan-va-y-kien', 'phan-bien', 'Phản biện', 1),
    ('tranh-luan-va-y-kien', 'dan-chung', 'Đưa dẫn chứng', 2),
    ('tranh-luan-va-y-kien', 'chot-van-de', 'Chốt lại vấn đề', 3),

    -- B2 · Tiếng Anh công sở
    ('tieng-anh-cong-so', 'hop-hanh', 'Trong cuộc họp', 0),
    ('tieng-anh-cong-so', 'bao-cao-tien-do', 'Báo cáo tiến độ', 1),
    ('tieng-anh-cong-so', 'thuong-luong', 'Thương lượng', 2),
    ('tieng-anh-cong-so', 'phan-hoi-dong-nghiep', 'Phản hồi đồng nghiệp', 3),

    -- C1 · Học thuật và nghiên cứu
    ('hoc-thuat-va-nghien-cuu', 'trich-dan-nguon', 'Trích dẫn nguồn', 0),
    ('hoc-thuat-va-nghien-cuu', 'mo-ta-du-lieu', 'Mô tả dữ liệu', 1),
    ('hoc-thuat-va-nghien-cuu', 'gioi-han-nghien-cuu', 'Giới hạn của nghiên cứu', 2),
    ('hoc-thuat-va-nghien-cuu', 'thuat-ngu-hoc-thuat', 'Thuật ngữ học thuật', 3),

    -- C1 · Sắc thái và mức độ trang trọng
    ('sac-thai-va-thanh-ngu', 'thanh-ngu-thong-dung', 'Tính từ mạnh và mức độ', 0),
    ('sac-thai-va-thanh-ngu', 'cum-dong-tu', 'Động từ trang trọng thay cụm động từ', 1),
    ('sac-thai-va-thanh-ngu', 'noi-giam-noi-tranh', 'Nói giảm nói tránh', 2),
    ('sac-thai-va-thanh-ngu', 'sac-thai-trang-trong', 'Sắc thái trang trọng', 3),

    -- C2 · Văn phong và tu từ
    ('van-phong-va-tu-tu', 'bien-phap-tu-tu', 'Biện pháp tu từ', 0),
    ('van-phong-va-tu-tu', 'nhip-dieu-cau-van', 'Nhịp điệu câu văn', 1),
    ('van-phong-va-tu-tu', 'tu-vung-tinh-te', 'Vốn từ tinh tế', 2),
    ('van-phong-va-tu-tu', 'van-phong-bao-chi', 'Văn phong báo chí', 3),

    -- C2 · Diễn thuyết và thuyết phục
    ('dien-thuyet-va-thuyet-phuc', 'mo-dau-an-tuong', 'Mở đầu ấn tượng', 0),
    ('dien-thuyet-va-thuyet-phuc', 'dung-lap-luan', 'Dựng lập luận', 1),
    ('dien-thuyet-va-thuyet-phuc', 'keo-huong-nguoi-nghe', 'Kéo hướng người nghe', 2),
    ('dien-thuyet-va-thuyet-phuc', 'ket-bai-dong-lai', 'Kết bài đọng lại', 3)
) AS v(course_slug, slug, title, position)
JOIN courses AS c ON c.slug = v.course_slug AND c.deleted_at IS NULL
ON CONFLICT (course_id, slug) WHERE deleted_at IS NULL DO NOTHING;

/* ---------- chỉnh lại những gì đã nạp ---------- */
-- Hai câu INSERT ở trên bỏ qua dòng đã có (DO NOTHING), nên sửa tên trong đó chỉ
-- có tác dụng với database mới. Những sửa đổi dưới đây áp cả lên database đã
-- seed từ trước. Slug giữ nguyên: nó là khoá để các file seed khác tra tới.

-- Khoá C1 này từng dạy thành ngữ bằng cách lưu chúng dưới dạng một từ (ropes,
-- backburner, iron…) — dạy sai. Từ vựng đã được thay bằng tính từ mạnh và
-- động từ trang trọng, nhưng tên khoá và tên bài vẫn hứa dạy thành ngữ và cụm
-- động từ.
UPDATE courses SET title = v.title, description = v.description, updated_at = now()
FROM (VALUES
    ('sac-thai-va-thanh-ngu', 'Sắc thái và mức độ trang trọng',
     'Tính từ mạnh, động từ trang trọng thay cụm động từ, nói giảm nói tránh.')
) AS v(slug, title, description)
WHERE courses.slug = v.slug AND courses.deleted_at IS NULL
  AND (courses.title, courses.description) IS DISTINCT FROM (v.title, v.description);

UPDATE lessons SET title = v.title, updated_at = now()
FROM (VALUES
    ('thanh-ngu-thong-dung', 'Tính từ mạnh và mức độ'),
    ('cum-dong-tu', 'Động từ trang trọng thay cụm động từ')
) AS v(slug, title)
WHERE lessons.slug = v.slug AND lessons.deleted_at IS NULL AND lessons.title <> v.title;

-- to be đứng đầu khoá ngữ pháp A1 chứ không đứng cuối: nó là động từ dùng
-- sớm nhất, và bài hiện tại đơn đã dùng nó ("They are often late", trạng từ
-- đứng sau to be). Thứ tự theo giáo trình vỡ lòng thông thường: to be, mạo từ,
-- danh từ đếm được, rồi mới tới thì hiện tại đơn.
UPDATE lessons SET position = v.position, updated_at = now()
FROM (VALUES
    ('dong-tu-to-be', 0),
    ('mao-tu-a-an-the', 1),
    ('danh-tu-dem-duoc', 2),
    ('thi-hien-tai-don', 3)
) AS v(slug, position)
JOIN courses AS c ON c.slug = 'ngu-phap-co-ban' AND c.deleted_at IS NULL
WHERE lessons.slug = v.slug AND lessons.course_id = c.id
  AND lessons.deleted_at IS NULL AND lessons.position <> v.position;

/* ---------- từ vựng ---------- */
-- Không nằm ở đây. 004_vocabulary.sql nạp toàn bộ từ vựng của mọi bài, với
-- bậc CEFR đối chiếu được bằng máy. Danh sách từng có ở file này bị 004 ghi đè
-- trọn vẹn, và là đúng bộ từ cũ: 42% đặt sai bậc, có cả thành ngữ lưu dưới
-- dạng một từ (ropes, backburner) — thứ dạy sai.
