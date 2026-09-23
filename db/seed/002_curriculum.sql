-- Giáo trình mẫu cho môi trường dev: đủ khoá ở cả sáu bậc CEFR, đủ bài và đủ
-- từ vựng để các màn Lộ trình, Từ vựng, Luyện tập và Tiến độ có gì để hiển thị.
--
-- Chạy lại được nhiều lần. Không dùng id cố định: khoá tra theo slug, bài tra
-- theo (khoá, slug), từ tra theo (bài, từ) — đúng các unique index đang có.
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

    -- B2 · Luyện thi IELTS Writing (khoá nháp — vẫn có bài để admin xem)
    ('luyen-thi-ielts-writing', 'task-1-bieu-do', 'Task 1: mô tả biểu đồ', 0),
    ('luyen-thi-ielts-writing', 'task-2-cau-truc', 'Task 2: cấu trúc bài', 1),
    ('luyen-thi-ielts-writing', 'tu-noi-hoc-thuat', 'Từ nối học thuật', 2),
    ('luyen-thi-ielts-writing', 'loi-thuong-gap', 'Lỗi thường gặp', 3),

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
-- Sáu từ mỗi bài. Câu ví dụ luôn chứa nguyên từ đó: màn Luyện tập khoét chính
-- từ này ra để làm câu điền chỗ trống, nên ví dụ chia ở dạng khác sẽ khiến bài
-- đó rơi về dạng trắc nghiệm.
-- Khoá nháp (IELTS) cố ý không có từ vựng.

INSERT INTO vocabulary_entries (lesson_id, word, ipa, meaning, example, example_vi, position)
SELECT l.id, v.word, v.ipa, v.meaning, v.example, v.example_vi, v.position
FROM (VALUES
    -- A1 · Ngữ pháp cơ bản · Thì hiện tại đơn
    ('thi-hien-tai-don', 'wake', '/weɪk/', 'thức dậy', 'I wake up at six every morning.', 'Tôi thức dậy lúc sáu giờ mỗi sáng.', 0),
    ('thi-hien-tai-don', 'always', '/ˈɔːlweɪz/', 'luôn luôn', 'She always drinks tea after lunch.', 'Cô ấy luôn uống trà sau bữa trưa.', 1),
    ('thi-hien-tai-don', 'usually', '/ˈjuːʒuəli/', 'thường thường', 'We usually walk to school.', 'Chúng tôi thường đi bộ đến trường.', 2),
    ('thi-hien-tai-don', 'never', '/ˈnevər/', 'không bao giờ', 'He never forgets her birthday.', 'Anh ấy không bao giờ quên sinh nhật cô ấy.', 3),
    ('thi-hien-tai-don', 'weekend', '/ˈwiːkend/', 'cuối tuần', 'They visit their parents every weekend.', 'Họ về thăm bố mẹ mỗi cuối tuần.', 4),
    ('thi-hien-tai-don', 'together', '/təˈɡeðər/', 'cùng nhau', 'My family eats dinner together.', 'Gia đình tôi ăn tối cùng nhau.', 5),

    -- A1 · Ngữ pháp cơ bản · Danh từ đếm được và không đếm được
    ('danh-tu-dem-duoc', 'bottle', '/ˈbɑːtl/', 'cái chai', 'There is one bottle of water on the table.', 'Có một chai nước trên bàn.', 0),
    ('danh-tu-dem-duoc', 'sugar', '/ˈʃʊɡər/', 'đường', 'I do not take sugar in my coffee.', 'Tôi không cho đường vào cà phê.', 1),
    ('danh-tu-dem-duoc', 'slice', '/slaɪs/', 'lát, miếng', 'She ate one slice of bread.', 'Cô ấy ăn một lát bánh mì.', 2),
    ('danh-tu-dem-duoc', 'advice', '/ədˈvaɪs/', 'lời khuyên', 'He gave me some useful advice.', 'Anh ấy cho tôi vài lời khuyên hữu ích.', 3),
    ('danh-tu-dem-duoc', 'furniture', '/ˈfɜːrnɪtʃər/', 'đồ đạc, nội thất', 'The room has very little furniture.', 'Căn phòng có rất ít đồ đạc.', 4),
    ('danh-tu-dem-duoc', 'luggage', '/ˈlʌɡɪdʒ/', 'hành lý', 'Please leave your luggage at the door.', 'Xin để hành lý của bạn ở cửa.', 5),

    -- A1 · Ngữ pháp cơ bản · Mạo từ a, an, the
    ('mao-tu-a-an-the', 'hour', '/ˈaʊər/', 'giờ, tiếng đồng hồ', 'We waited for an hour.', 'Chúng tôi đã chờ một tiếng.', 0),
    ('mao-tu-a-an-the', 'umbrella', '/ʌmˈbrelə/', 'cái ô, cái dù', 'Take an umbrella with you.', 'Mang theo một cái ô nhé.', 1),
    ('mao-tu-a-an-the', 'university', '/ˌjuːnɪˈvɜːrsəti/', 'trường đại học', 'She studies at a university in Hanoi.', 'Cô ấy học ở một trường đại học tại Hà Nội.', 2),
    ('mao-tu-a-an-the', 'sun', '/sʌn/', 'mặt trời', 'The sun rises in the east.', 'Mặt trời mọc ở hướng đông.', 3),
    ('mao-tu-a-an-the', 'ticket', '/ˈtɪkɪt/', 'vé', 'I bought a ticket for the show.', 'Tôi đã mua một vé xem buổi diễn.', 4),
    ('mao-tu-a-an-the', 'answer', '/ˈænsər/', 'câu trả lời', 'She gave an answer right away.', 'Cô ấy đưa ra câu trả lời ngay.', 5),

    -- A1 · Ngữ pháp cơ bản · Động từ to be
    ('dong-tu-to-be', 'tired', '/ˈtaɪərd/', 'mệt', 'I am tired after the trip.', 'Tôi mệt sau chuyến đi.', 0),
    ('dong-tu-to-be', 'hungry', '/ˈhʌŋɡri/', 'đói', 'The children are hungry.', 'Bọn trẻ đang đói.', 1),
    ('dong-tu-to-be', 'ready', '/ˈredi/', 'sẵn sàng', 'We are ready to start.', 'Chúng tôi đã sẵn sàng bắt đầu.', 2),
    ('dong-tu-to-be', 'late', '/leɪt/', 'muộn, trễ', 'He is late again.', 'Anh ấy lại đến muộn.', 3),
    ('dong-tu-to-be', 'busy', '/ˈbɪzi/', 'bận', 'She is busy this afternoon.', 'Chiều nay cô ấy bận.', 4),
    ('dong-tu-to-be', 'afraid', '/əˈfreɪd/', 'sợ', 'They are afraid of dogs.', 'Họ sợ chó.', 5),

    -- A1 · Từ vựng đời sống · Gia đình và bạn bè
    ('gia-dinh-va-ban-be', 'parents', '/ˈperənts/', 'bố mẹ', 'My parents live in the countryside.', 'Bố mẹ tôi sống ở quê.', 0),
    ('gia-dinh-va-ban-be', 'cousin', '/ˈkʌzn/', 'anh chị em họ', 'My cousin is the same age as me.', 'Anh họ tôi bằng tuổi tôi.', 1),
    ('gia-dinh-va-ban-be', 'neighbour', '/ˈneɪbər/', 'hàng xóm', 'Our neighbour has a big garden.', 'Hàng xóm của chúng tôi có một khu vườn lớn.', 2),
    ('gia-dinh-va-ban-be', 'married', '/ˈmerid/', 'đã kết hôn', 'My sister got married last year.', 'Chị tôi đã kết hôn năm ngoái.', 3),
    ('gia-dinh-va-ban-be', 'grandmother', '/ˈɡrænmʌðər/', 'bà', 'My grandmother tells wonderful stories.', 'Bà tôi kể những câu chuyện rất hay.', 4),
    ('gia-dinh-va-ban-be', 'friendly', '/ˈfrendli/', 'thân thiện', 'Everyone here is very friendly.', 'Mọi người ở đây rất thân thiện.', 5),

    -- A1 · Từ vựng đời sống · Màu sắc và hình dáng
    ('mau-sac-va-hinh-dang', 'round', '/raʊnd/', 'tròn', 'The table is round.', 'Cái bàn có hình tròn.', 0),
    ('mau-sac-va-hinh-dang', 'square', '/skwer/', 'vuông', 'He drew a square on the board.', 'Anh ấy vẽ một hình vuông lên bảng.', 1),
    ('mau-sac-va-hinh-dang', 'narrow', '/ˈnæroʊ/', 'hẹp', 'The street is very narrow.', 'Con phố rất hẹp.', 2),
    ('mau-sac-va-hinh-dang', 'bright', '/braɪt/', 'sáng, rực rỡ', 'She wore a bright yellow dress.', 'Cô ấy mặc một chiếc váy vàng rực rỡ.', 3),
    ('mau-sac-va-hinh-dang', 'dark', '/dɑːrk/', 'tối, sẫm màu', 'He likes dark blue shirts.', 'Anh ấy thích áo sơ mi màu xanh sẫm.', 4),
    ('mau-sac-va-hinh-dang', 'thick', '/θɪk/', 'dày', 'The book has a thick cover.', 'Cuốn sách có bìa dày.', 5),

    -- A1 · Từ vựng đời sống · Đồ ăn và thức uống
    ('do-an-va-thuc-uong', 'breakfast', '/ˈbrekfəst/', 'bữa sáng', 'I never skip breakfast.', 'Tôi không bao giờ bỏ bữa sáng.', 0),
    ('do-an-va-thuc-uong', 'vegetable', '/ˈvedʒtəbl/', 'rau củ', 'Eat one vegetable with every meal.', 'Hãy ăn một loại rau củ trong mỗi bữa.', 1),
    ('do-an-va-thuc-uong', 'delicious', '/dɪˈlɪʃəs/', 'ngon', 'This soup is delicious.', 'Món canh này rất ngon.', 2),
    ('do-an-va-thuc-uong', 'spicy', '/ˈspaɪsi/', 'cay', 'The food here is quite spicy.', 'Đồ ăn ở đây khá cay.', 3),
    ('do-an-va-thuc-uong', 'fresh', '/freʃ/', 'tươi', 'We buy fresh fish every morning.', 'Chúng tôi mua cá tươi mỗi sáng.', 4),
    ('do-an-va-thuc-uong', 'thirsty', '/ˈθɜːrsti/', 'khát', 'I am thirsty after the run.', 'Tôi khát nước sau khi chạy.', 5),

    -- A1 · Từ vựng đời sống · Thời gian và ngày tháng
    ('thoi-gian-va-ngay-thang', 'tomorrow', '/təˈmɑːroʊ/', 'ngày mai', 'The test is tomorrow.', 'Bài kiểm tra vào ngày mai.', 0),
    ('thoi-gian-va-ngay-thang', 'yesterday', '/ˈjestərdeɪ/', 'hôm qua', 'It rained all day yesterday.', 'Hôm qua trời mưa cả ngày.', 1),
    ('thoi-gian-va-ngay-thang', 'month', '/mʌnθ/', 'tháng', 'We moved here last month.', 'Chúng tôi chuyển đến đây tháng trước.', 2),
    ('thoi-gian-va-ngay-thang', 'early', '/ˈɜːrli/', 'sớm', 'She arrived early.', 'Cô ấy đến sớm.', 3),
    ('thoi-gian-va-ngay-thang', 'evening', '/ˈiːvnɪŋ/', 'buổi tối', 'I study in the evening.', 'Tôi học vào buổi tối.', 4),
    ('thoi-gian-va-ngay-thang', 'minute', '/ˈmɪnɪt/', 'phút', 'Wait one minute, please.', 'Xin chờ một phút.', 5)
) AS v(lesson_slug, word, ipa, meaning, example, example_vi, position)
JOIN lessons AS l ON l.slug = v.lesson_slug AND l.deleted_at IS NULL
ON CONFLICT (lesson_id, lower(word)) WHERE deleted_at IS NULL DO NOTHING;

INSERT INTO vocabulary_entries (lesson_id, word, ipa, meaning, example, example_vi, position)
SELECT l.id, v.word, v.ipa, v.meaning, v.example, v.example_vi, v.position
FROM (VALUES
    -- A2 · Giao tiếp hàng ngày · Chào hỏi và làm quen
    ('chao-hoi-lam-quen', 'introduce', '/ˌɪntrəˈduːs/', 'giới thiệu', 'Let me introduce my colleague.', 'Để tôi giới thiệu đồng nghiệp của mình.', 0),
    ('chao-hoi-lam-quen', 'pleasure', '/ˈpleʒər/', 'niềm vui, hân hạnh', 'It is a pleasure to meet you.', 'Rất hân hạnh được gặp bạn.', 1),
    ('chao-hoi-lam-quen', 'hometown', '/ˈhoʊmtaʊn/', 'quê nhà', 'My hometown is near the sea.', 'Quê tôi ở gần biển.', 2),
    ('chao-hoi-lam-quen', 'occupation', '/ˌɑːkjuˈpeɪʃn/', 'nghề nghiệp', 'Please write your occupation here.', 'Vui lòng ghi nghề nghiệp của bạn vào đây.', 3),
    ('chao-hoi-lam-quen', 'spell', '/spel/', 'đánh vần', 'Could you spell your name for me?', 'Bạn đánh vần tên giúp tôi được không?', 4),
    ('chao-hoi-lam-quen', 'repeat', '/rɪˈpiːt/', 'nhắc lại', 'Could you repeat that, please?', 'Bạn nhắc lại được không?', 5),

    -- A2 · Giao tiếp hàng ngày · Gọi món ở nhà hàng
    ('goi-mon-o-nha-hang', 'order', '/ˈɔːrdər/', 'gọi món', 'Are you ready to order?', 'Anh chị đã sẵn sàng gọi món chưa?', 0),
    ('goi-mon-o-nha-hang', 'menu', '/ˈmenjuː/', 'thực đơn', 'Could we see the menu, please?', 'Cho chúng tôi xem thực đơn được không?', 1),
    ('goi-mon-o-nha-hang', 'recommend', '/ˌrekəˈmend/', 'gợi ý, giới thiệu', 'What would you recommend?', 'Bạn gợi ý món gì?', 2),
    ('goi-mon-o-nha-hang', 'bill', '/bɪl/', 'hoá đơn', 'Could we have the bill, please?', 'Cho chúng tôi xin hoá đơn.', 3),
    ('goi-mon-o-nha-hang', 'reservation', '/ˌrezərˈveɪʃn/', 'đặt bàn trước', 'I have a reservation for two.', 'Tôi có đặt bàn cho hai người.', 4),
    ('goi-mon-o-nha-hang', 'allergic', '/əˈlɜːrdʒɪk/', 'dị ứng', 'I am allergic to peanuts.', 'Tôi bị dị ứng đậu phộng.', 5),

    -- A2 · Giao tiếp hàng ngày · Hỏi đường
    ('hoi-duong-di', 'straight', '/streɪt/', 'thẳng', 'Go straight for two blocks.', 'Đi thẳng qua hai dãy nhà.', 0),
    ('hoi-duong-di', 'corner', '/ˈkɔːrnər/', 'góc đường', 'The bank is on the corner.', 'Ngân hàng nằm ở góc đường.', 1),
    ('hoi-duong-di', 'opposite', '/ˈɑːpəzɪt/', 'đối diện', 'The pharmacy is opposite the market.', 'Hiệu thuốc đối diện chợ.', 2),
    ('hoi-duong-di', 'crossroads', '/ˈkrɔːsroʊdz/', 'ngã tư', 'Turn left at the crossroads.', 'Rẽ trái ở ngã tư.', 3),
    ('hoi-duong-di', 'distance', '/ˈdɪstəns/', 'khoảng cách', 'The distance is about two kilometres.', 'Khoảng cách chừng hai cây số.', 4),
    ('hoi-duong-di', 'lost', '/lɔːst/', 'lạc đường', 'I think we are lost.', 'Tôi nghĩ chúng ta bị lạc rồi.', 5),

    -- A2 · Giao tiếp hàng ngày · Mua sắm
    ('mua-sam', 'discount', '/ˈdɪskaʊnt/', 'giảm giá', 'Is there a discount on this?', 'Cái này có giảm giá không?', 0),
    ('mua-sam', 'receipt', '/rɪˈsiːt/', 'biên lai', 'Keep the receipt for the warranty.', 'Giữ biên lai để bảo hành.', 1),
    ('mua-sam', 'refund', '/ˈriːfʌnd/', 'hoàn tiền', 'Can I get a refund?', 'Tôi được hoàn tiền không?', 2),
    ('mua-sam', 'size', '/saɪz/', 'cỡ, kích thước', 'Do you have this in a larger size?', 'Cái này có cỡ lớn hơn không?', 3),
    ('mua-sam', 'expensive', '/ɪkˈspensɪv/', 'đắt', 'That bag is too expensive.', 'Cái túi đó đắt quá.', 4),
    ('mua-sam', 'cash', '/kæʃ/', 'tiền mặt', 'I would like to pay in cash.', 'Tôi muốn trả bằng tiền mặt.', 5),

    -- A2 · Quá khứ và tương lai · Thì quá khứ đơn
    ('thi-qua-khu-don', 'arrived', '/əˈraɪvd/', 'đã đến', 'We arrived before noon.', 'Chúng tôi đã đến trước trưa.', 0),
    ('thi-qua-khu-don', 'finished', '/ˈfɪnɪʃt/', 'đã xong', 'She finished the report last night.', 'Cô ấy đã làm xong báo cáo tối qua.', 1),
    ('thi-qua-khu-don', 'decided', '/dɪˈsaɪdɪd/', 'đã quyết định', 'They decided to stay one more day.', 'Họ quyết định ở thêm một ngày.', 2),
    ('thi-qua-khu-don', 'invited', '/ɪnˈvaɪtɪd/', 'đã mời', 'He invited us to dinner.', 'Anh ấy đã mời chúng tôi ăn tối.', 3),
    ('thi-qua-khu-don', 'happened', '/ˈhæpənd/', 'đã xảy ra', 'Nothing happened after that.', 'Sau đó không có gì xảy ra.', 4),
    ('thi-qua-khu-don', 'ago', '/əˈɡoʊ/', 'cách đây', 'We met three years ago.', 'Chúng tôi gặp nhau cách đây ba năm.', 5),

    -- A2 · Quá khứ và tương lai · Động từ bất quy tắc
    ('dong-tu-bat-quy-tac', 'bought', '/bɔːt/', 'đã mua', 'I bought this yesterday.', 'Tôi mua cái này hôm qua.', 0),
    ('dong-tu-bat-quy-tac', 'caught', '/kɔːt/', 'đã bắt, đã kịp', 'She caught the last bus.', 'Cô ấy kịp chuyến xe buýt cuối.', 1),
    ('dong-tu-bat-quy-tac', 'brought', '/brɔːt/', 'đã mang', 'He brought some fruit.', 'Anh ấy mang theo ít hoa quả.', 2),
    ('dong-tu-bat-quy-tac', 'taught', '/tɔːt/', 'đã dạy', 'She taught here for ten years.', 'Cô ấy đã dạy ở đây mười năm.', 3),
    ('dong-tu-bat-quy-tac', 'chose', '/tʃoʊz/', 'đã chọn', 'They chose the cheaper option.', 'Họ chọn phương án rẻ hơn.', 4),
    ('dong-tu-bat-quy-tac', 'wrote', '/roʊt/', 'đã viết', 'He wrote three letters.', 'Anh ấy đã viết ba lá thư.', 5),

    -- A2 · Quá khứ và tương lai · Nói về dự định
    ('noi-ve-du-dinh', 'plan', '/plæn/', 'dự định', 'We plan to move in June.', 'Chúng tôi dự định chuyển nhà vào tháng Sáu.', 0),
    ('noi-ve-du-dinh', 'hope', '/hoʊp/', 'hy vọng', 'I hope to see you soon.', 'Tôi hy vọng sớm gặp lại bạn.', 1),
    ('noi-ve-du-dinh', 'probably', '/ˈprɑːbəbli/', 'có lẽ', 'She will probably call tonight.', 'Có lẽ tối nay cô ấy sẽ gọi.', 2),
    ('noi-ve-du-dinh', 'maybe', '/ˈmeɪbi/', 'có thể', 'Maybe we will go by train.', 'Có thể chúng tôi sẽ đi tàu.', 3),
    ('noi-ve-du-dinh', 'soon', '/suːn/', 'sớm', 'The results will come soon.', 'Kết quả sẽ có sớm.', 4),
    ('noi-ve-du-dinh', 'intend', '/ɪnˈtend/', 'có ý định', 'They intend to open a shop.', 'Họ có ý định mở một cửa hàng.', 5),

    -- A2 · Quá khứ và tương lai · Kể về kỳ nghỉ
    ('ke-ve-ky-nghi', 'journey', '/ˈdʒɜːrni/', 'chuyến đi', 'The journey took six hours.', 'Chuyến đi mất sáu tiếng.', 0),
    ('ke-ve-ky-nghi', 'souvenir', '/ˌsuːvəˈnɪr/', 'quà lưu niệm', 'I bought a small souvenir.', 'Tôi mua một món quà lưu niệm nhỏ.', 1),
    ('ke-ve-ky-nghi', 'scenery', '/ˈsiːnəri/', 'phong cảnh', 'The scenery was beautiful.', 'Phong cảnh rất đẹp.', 2),
    ('ke-ve-ky-nghi', 'crowded', '/ˈkraʊdɪd/', 'đông đúc', 'The beach was crowded.', 'Bãi biển rất đông.', 3),
    ('ke-ve-ky-nghi', 'relax', '/rɪˈlæks/', 'thư giãn', 'We went there to relax.', 'Chúng tôi đến đó để thư giãn.', 4),
    ('ke-ve-ky-nghi', 'sunburn', '/ˈsʌnbɜːrn/', 'cháy nắng', 'He got a bad sunburn.', 'Anh ấy bị cháy nắng nặng.', 5)
) AS v(lesson_slug, word, ipa, meaning, example, example_vi, position)
JOIN lessons AS l ON l.slug = v.lesson_slug AND l.deleted_at IS NULL
ON CONFLICT (lesson_id, lower(word)) WHERE deleted_at IS NULL DO NOTHING;

INSERT INTO vocabulary_entries (lesson_id, word, ipa, meaning, example, example_vi, position)
SELECT l.id, v.word, v.ipa, v.meaning, v.example, v.example_vi, v.position
FROM (VALUES
    -- B1 · Kể chuyện và mô tả · Mô tả một người
    ('mo-ta-nguoi', 'confident', '/ˈkɑːnfɪdənt/', 'tự tin', 'She is confident in front of a crowd.', 'Cô ấy tự tin khi đứng trước đám đông.', 0),
    ('mo-ta-nguoi', 'generous', '/ˈdʒenərəs/', 'hào phóng', 'He is generous with his time.', 'Anh ấy rất hào phóng về thời gian.', 1),
    ('mo-ta-nguoi', 'stubborn', '/ˈstʌbərn/', 'bướng bỉnh', 'My brother can be stubborn.', 'Em trai tôi đôi khi rất bướng.', 2),
    ('mo-ta-nguoi', 'reliable', '/rɪˈlaɪəbl/', 'đáng tin cậy', 'He is the most reliable person here.', 'Anh ấy là người đáng tin cậy nhất ở đây.', 3),
    ('mo-ta-nguoi', 'appearance', '/əˈpɪrəns/', 'vẻ ngoài', 'Do not judge by appearance alone.', 'Đừng chỉ đánh giá qua vẻ ngoài.', 4),
    ('mo-ta-nguoi', 'cheerful', '/ˈtʃɪrfl/', 'vui vẻ', 'She has a cheerful voice.', 'Cô ấy có giọng nói vui vẻ.', 5),

    -- B1 · Kể chuyện và mô tả · Mô tả nơi chốn
    ('mo-ta-noi-chon', 'peaceful', '/ˈpiːsfl/', 'yên bình', 'The village is peaceful at night.', 'Ngôi làng yên bình vào ban đêm.', 0),
    ('mo-ta-noi-chon', 'spacious', '/ˈspeɪʃəs/', 'rộng rãi', 'The kitchen is bright and spacious.', 'Căn bếp sáng và rộng rãi.', 1),
    ('mo-ta-noi-chon', 'surrounded', '/səˈraʊndɪd/', 'được bao quanh', 'The house is surrounded by trees.', 'Ngôi nhà được cây cối bao quanh.', 2),
    ('mo-ta-noi-chon', 'landmark', '/ˈlændmɑːrk/', 'địa danh nổi bật', 'The tower is a famous landmark.', 'Toà tháp là một địa danh nổi tiếng.', 3),
    ('mo-ta-noi-chon', 'atmosphere', '/ˈætməsfɪr/', 'không khí, bầu không khí', 'The café has a warm atmosphere.', 'Quán cà phê có không khí ấm cúng.', 4),
    ('mo-ta-noi-chon', 'remote', '/rɪˈmoʊt/', 'hẻo lánh, xa xôi', 'They live in a remote area.', 'Họ sống ở một vùng hẻo lánh.', 5),

    -- B1 · Kể chuyện và mô tả · Kể lại một sự việc
    ('ke-mot-su-viec', 'suddenly', '/ˈsʌdənli/', 'đột nhiên', 'Suddenly the lights went out.', 'Đột nhiên đèn tắt hết.', 0),
    ('ke-mot-su-viec', 'eventually', '/ɪˈventʃuəli/', 'cuối cùng thì', 'Eventually we found the key.', 'Cuối cùng chúng tôi cũng tìm thấy chìa khoá.', 1),
    ('ke-mot-su-viec', 'meanwhile', '/ˈmiːnwaɪl/', 'trong khi đó', 'Meanwhile the others waited outside.', 'Trong khi đó những người khác chờ bên ngoài.', 2),
    ('ke-mot-su-viec', 'realise', '/ˈriːəlaɪz/', 'nhận ra', 'I did not realise the time.', 'Tôi không nhận ra đã muộn thế.', 3),
    ('ke-mot-su-viec', 'incident', '/ˈɪnsɪdənt/', 'sự việc, vụ việc', 'Nobody mentioned the incident again.', 'Không ai nhắc lại vụ việc đó nữa.', 4),
    ('ke-mot-su-viec', 'fortunately', '/ˈfɔːrtʃənətli/', 'may mắn thay', 'Fortunately nobody was hurt.', 'May mắn thay không ai bị thương.', 5),

    -- B1 · Kể chuyện và mô tả · Cảm xúc và thái độ
    ('cam-xuc-va-thai-do', 'disappointed', '/ˌdɪsəˈpɔɪntɪd/', 'thất vọng', 'She was disappointed with the result.', 'Cô ấy thất vọng với kết quả.', 0),
    ('cam-xuc-va-thai-do', 'relieved', '/rɪˈliːvd/', 'nhẹ nhõm', 'I felt relieved after the call.', 'Tôi thấy nhẹ nhõm sau cuộc gọi.', 1),
    ('cam-xuc-va-thai-do', 'annoyed', '/əˈnɔɪd/', 'khó chịu', 'He looked annoyed by the noise.', 'Anh ấy có vẻ khó chịu vì tiếng ồn.', 2),
    ('cam-xuc-va-thai-do', 'grateful', '/ˈɡreɪtfl/', 'biết ơn', 'We are grateful for your help.', 'Chúng tôi biết ơn sự giúp đỡ của bạn.', 3),
    ('cam-xuc-va-thai-do', 'nervous', '/ˈnɜːrvəs/', 'lo lắng, hồi hộp', 'She gets nervous before exams.', 'Cô ấy hồi hộp trước mỗi kỳ thi.', 4),
    ('cam-xuc-va-thai-do', 'impressed', '/ɪmˈprest/', 'ấn tượng', 'They were impressed by her answer.', 'Họ ấn tượng với câu trả lời của cô ấy.', 5),

    -- B1 · Email và tin nhắn · Email trang trọng
    ('email-trang-trong', 'regarding', '/rɪˈɡɑːrdɪŋ/', 'về việc', 'I am writing regarding your invoice.', 'Tôi viết thư về việc hoá đơn của quý vị.', 0),
    ('email-trang-trong', 'attach', '/əˈtætʃ/', 'đính kèm', 'Please attach the signed form.', 'Vui lòng đính kèm biểu mẫu đã ký.', 1),
    ('email-trang-trong', 'sincerely', '/sɪnˈsɪrli/', 'trân trọng', 'The letter ended with Yours sincerely.', 'Lá thư kết thúc bằng lời chào trân trọng.', 2),
    ('email-trang-trong', 'enquiry', '/ɪnˈkwaɪəri/', 'thư hỏi, yêu cầu thông tin', 'Thank you for your enquiry.', 'Cảm ơn quý vị đã gửi thư hỏi.', 3),
    ('email-trang-trong', 'promptly', '/ˈprɑːmptli/', 'kịp thời, nhanh chóng', 'We will reply promptly.', 'Chúng tôi sẽ trả lời nhanh chóng.', 4),
    ('email-trang-trong', 'recipient', '/rɪˈsɪpiənt/', 'người nhận', 'Check the recipient before sending.', 'Kiểm tra người nhận trước khi gửi.', 5),

    -- B1 · Email và tin nhắn · Email thân mật
    ('email-than-mat', 'catch', '/kætʃ/', 'bắt kịp, gặp lại', 'Let us catch up next week.', 'Tuần sau gặp lại nhau nhé.', 0),
    ('email-than-mat', 'ages', '/ˈeɪdʒɪz/', 'lâu lắm rồi', 'I have not seen you for ages.', 'Lâu lắm rồi không gặp bạn.', 1),
    ('email-than-mat', 'anyway', '/ˈeniweɪ/', 'dù sao thì', 'Anyway, how is the new job?', 'Dù sao thì công việc mới thế nào?', 2),
    ('email-than-mat', 'cheers', '/tʃɪrz/', 'cảm ơn nhé (thân mật)', 'Cheers for the photos!', 'Cảm ơn nhé vì mấy tấm ảnh!', 3),
    ('email-than-mat', 'update', '/ˈʌpdeɪt/', 'tin mới, cập nhật', 'Send me an update when you can.', 'Khi nào rảnh gửi tôi tin mới nhé.', 4),
    ('email-than-mat', 'lately', '/ˈleɪtli/', 'dạo này', 'Have you been busy lately?', 'Dạo này bạn có bận không?', 5),

    -- B1 · Email và tin nhắn · Hẹn lịch họp
    ('hen-lich-hop', 'available', '/əˈveɪləbl/', 'có thể, rảnh', 'Are you available on Thursday?', 'Thứ Năm bạn có rảnh không?', 0),
    ('hen-lich-hop', 'reschedule', '/ˌriːˈskedʒuːl/', 'dời lịch', 'Could we reschedule to Friday?', 'Chúng ta dời sang thứ Sáu được không?', 1),
    ('hen-lich-hop', 'agenda', '/əˈdʒendə/', 'chương trình họp', 'I will send the agenda tonight.', 'Tối nay tôi sẽ gửi chương trình họp.', 2),
    ('hen-lich-hop', 'confirm', '/kənˈfɜːrm/', 'xác nhận', 'Please confirm the time.', 'Vui lòng xác nhận giờ hẹn.', 3),
    ('hen-lich-hop', 'clash', '/klæʃ/', 'trùng lịch', 'That time would clash with my class.', 'Giờ đó trùng với lớp của tôi.', 4),
    ('hen-lich-hop', 'venue', '/ˈvenjuː/', 'địa điểm', 'The venue has changed.', 'Địa điểm đã thay đổi.', 5),

    -- B1 · Email và tin nhắn · Xin lỗi và cảm ơn
    ('xin-loi-va-cam-on', 'apologise', '/əˈpɑːlədʒaɪz/', 'xin lỗi', 'I apologise for the delay.', 'Tôi xin lỗi vì sự chậm trễ.', 0),
    ('xin-loi-va-cam-on', 'inconvenience', '/ˌɪnkənˈviːniəns/', 'sự bất tiện', 'Sorry for any inconvenience.', 'Xin lỗi vì mọi bất tiện.', 1),
    ('xin-loi-va-cam-on', 'appreciate', '/əˈpriːʃieɪt/', 'trân trọng, cảm kích', 'I really appreciate your patience.', 'Tôi thật sự cảm kích sự kiên nhẫn của bạn.', 2),
    ('xin-loi-va-cam-on', 'oversight', '/ˈoʊvərsaɪt/', 'sơ suất', 'It was an oversight on our part.', 'Đó là sơ suất từ phía chúng tôi.', 3),
    ('xin-loi-va-cam-on', 'grateful', '/ˈɡreɪtfl/', 'biết ơn', 'I would be grateful for a reply.', 'Tôi sẽ rất biết ơn nếu nhận được hồi âm.', 4),
    ('xin-loi-va-cam-on', 'assure', '/əˈʃʊr/', 'cam đoan', 'I assure you it will not happen again.', 'Tôi cam đoan chuyện đó sẽ không lặp lại.', 5)
) AS v(lesson_slug, word, ipa, meaning, example, example_vi, position)
JOIN lessons AS l ON l.slug = v.lesson_slug AND l.deleted_at IS NULL
ON CONFLICT (lesson_id, lower(word)) WHERE deleted_at IS NULL DO NOTHING;

INSERT INTO vocabulary_entries (lesson_id, word, ipa, meaning, example, example_vi, position)
SELECT l.id, v.word, v.ipa, v.meaning, v.example, v.example_vi, v.position
FROM (VALUES
    -- B2 · Tranh luận và nêu ý kiến · Nêu quan điểm
    ('neu-quan-diem', 'arguably', '/ˈɑːrɡjuəbli/', 'có thể nói rằng', 'This is arguably the better option.', 'Có thể nói đây là phương án tốt hơn.', 0),
    ('neu-quan-diem', 'perspective', '/pərˈspektɪv/', 'góc nhìn', 'From my perspective, the cost matters more.', 'Theo góc nhìn của tôi, chi phí quan trọng hơn.', 1),
    ('neu-quan-diem', 'assumption', '/əˈsʌmpʃn/', 'giả định', 'That argument rests on one assumption.', 'Lập luận đó dựa trên một giả định.', 2),
    ('neu-quan-diem', 'advocate', '/ˈædvəkeɪt/', 'ủng hộ, cổ vũ', 'Many experts advocate a slower rollout.', 'Nhiều chuyên gia ủng hộ triển khai chậm hơn.', 3),
    ('neu-quan-diem', 'nuance', '/ˈnuːɑːns/', 'sắc thái tinh tế', 'The report misses an important nuance.', 'Báo cáo bỏ sót một sắc thái quan trọng.', 4),
    ('neu-quan-diem', 'stance', '/stæns/', 'lập trường', 'She made her stance very clear.', 'Cô ấy nêu rất rõ lập trường của mình.', 5),

    -- B2 · Tranh luận và nêu ý kiến · Phản biện
    ('phan-bien', 'counter', '/ˈkaʊntər/', 'phản bác', 'He could not counter that point.', 'Anh ấy không phản bác được điểm đó.', 0),
    ('phan-bien', 'flawed', '/flɔːd/', 'có lỗ hổng', 'The reasoning is flawed.', 'Lập luận có lỗ hổng.', 1),
    ('phan-bien', 'overlook', '/ˌoʊvərˈlʊk/', 'bỏ sót', 'The study seems to overlook cost.', 'Nghiên cứu có vẻ bỏ sót yếu tố chi phí.', 2),
    ('phan-bien', 'concede', '/kənˈsiːd/', 'thừa nhận', 'I concede that the timing was poor.', 'Tôi thừa nhận là thời điểm không tốt.', 3),
    ('phan-bien', 'misleading', '/ˌmɪsˈliːdɪŋ/', 'gây hiểu lầm', 'That chart is misleading.', 'Biểu đồ đó gây hiểu lầm.', 4),
    ('phan-bien', 'oversimplify', '/ˌoʊvərˈsɪmplɪfaɪ/', 'đơn giản hoá quá mức', 'We should not oversimplify the issue.', 'Chúng ta không nên đơn giản hoá quá mức vấn đề.', 5),

    -- B2 · Tranh luận và nêu ý kiến · Đưa dẫn chứng
    ('dan-chung', 'evidence', '/ˈevɪdəns/', 'bằng chứng', 'There is little evidence for that claim.', 'Có rất ít bằng chứng cho khẳng định đó.', 0),
    ('dan-chung', 'demonstrate', '/ˈdemənstreɪt/', 'chứng minh', 'The data demonstrate a clear trend.', 'Dữ liệu chứng minh một xu hướng rõ ràng.', 1),
    ('dan-chung', 'correlation', '/ˌkɔːrəˈleɪʃn/', 'mối tương quan', 'A correlation is not a cause.', 'Mối tương quan không phải là nguyên nhân.', 2),
    ('dan-chung', 'substantial', '/səbˈstænʃl/', 'đáng kể', 'The team made substantial progress.', 'Nhóm đã có tiến bộ đáng kể.', 3),
    ('dan-chung', 'anecdotal', '/ˌænɪkˈdoʊtl/', 'mang tính giai thoại', 'That is anecdotal rather than measured.', 'Cái đó mang tính giai thoại chứ chưa đo đạc.', 4),
    ('dan-chung', 'cite', '/saɪt/', 'trích dẫn', 'Please cite the original study.', 'Vui lòng trích dẫn nghiên cứu gốc.', 5),

    -- B2 · Tranh luận và nêu ý kiến · Chốt lại vấn đề
    ('chot-van-de', 'ultimately', '/ˈʌltɪmətli/', 'rốt cuộc', 'Ultimately the decision is yours.', 'Rốt cuộc quyết định là của bạn.', 0),
    ('chot-van-de', 'consensus', '/kənˈsensəs/', 'sự đồng thuận', 'We reached a consensus quickly.', 'Chúng tôi nhanh chóng đạt được đồng thuận.', 1),
    ('chot-van-de', 'trade', '/treɪd/', 'đánh đổi', 'Every plan involves a trade off.', 'Mọi kế hoạch đều có sự đánh đổi.', 2),
    ('chot-van-de', 'summarise', '/ˈsʌməraɪz/', 'tóm tắt', 'Let me summarise the main points.', 'Để tôi tóm tắt các ý chính.', 3),
    ('chot-van-de', 'outweigh', '/ˌaʊtˈweɪ/', 'lớn hơn, vượt trội', 'The benefits outweigh the risks.', 'Lợi ích vượt trội so với rủi ro.', 4),
    ('chot-van-de', 'pragmatic', '/præɡˈmætɪk/', 'thực dụng', 'We need a pragmatic solution.', 'Chúng ta cần một giải pháp thực dụng.', 5),

    -- B2 · Tiếng Anh công sở · Trong cuộc họp
    ('hop-hanh', 'minutes', '/ˈmɪnɪts/', 'biên bản họp', 'Who is taking the minutes?', 'Ai ghi biên bản họp?', 0),
    ('hop-hanh', 'circulate', '/ˈsɜːrkjəleɪt/', 'gửi luân chuyển', 'I will circulate the slides after this.', 'Tôi sẽ gửi slide cho mọi người sau buổi này.', 1),
    ('hop-hanh', 'defer', '/dɪˈfɜːr/', 'hoãn lại', 'Can we defer that item?', 'Chúng ta hoãn mục đó lại được không?', 2),
    ('hop-hanh', 'clarify', '/ˈklærɪfaɪ/', 'làm rõ', 'Could you clarify the deadline?', 'Bạn làm rõ giúp hạn chót được không?', 3),
    ('hop-hanh', 'recap', '/ˈriːkæp/', 'điểm lại', 'A quick recap before we finish.', 'Điểm lại nhanh trước khi kết thúc.', 4),
    ('hop-hanh', 'agenda', '/əˈdʒendə/', 'chương trình họp', 'That is not on the agenda today.', 'Hôm nay mục đó không có trong chương trình họp.', 5),

    -- B2 · Tiếng Anh công sở · Báo cáo tiến độ
    ('bao-cao-tien-do', 'milestone', '/ˈmaɪlstoʊn/', 'cột mốc', 'We hit the first milestone early.', 'Chúng tôi đạt cột mốc đầu tiên sớm.', 0),
    ('bao-cao-tien-do', 'blocker', '/ˈblɑːkər/', 'điểm tắc nghẽn', 'The main blocker is the missing data.', 'Điểm tắc nghẽn chính là thiếu dữ liệu.', 1),
    ('bao-cao-tien-do', 'estimate', '/ˈestɪmeɪt/', 'ước lượng', 'Our estimate was too optimistic.', 'Ước lượng của chúng tôi quá lạc quan.', 2),
    ('bao-cao-tien-do', 'behind', '/bɪˈhaɪnd/', 'chậm tiến độ', 'We are two weeks behind schedule.', 'Chúng tôi chậm tiến độ hai tuần.', 3),
    ('bao-cao-tien-do', 'scope', '/skoʊp/', 'phạm vi công việc', 'That request is outside the scope.', 'Yêu cầu đó nằm ngoài phạm vi công việc.', 4),
    ('bao-cao-tien-do', 'allocate', '/ˈæləkeɪt/', 'phân bổ', 'We need to allocate more time.', 'Chúng ta cần phân bổ thêm thời gian.', 5),

    -- B2 · Tiếng Anh công sở · Thương lượng
    ('thuong-luong', 'compromise', '/ˈkɑːmprəmaɪz/', 'thoả hiệp', 'We reached a fair compromise.', 'Chúng tôi đạt được một thoả hiệp công bằng.', 0),
    ('thuong-luong', 'leverage', '/ˈlevərɪdʒ/', 'lợi thế đàm phán', 'They have more leverage than we do.', 'Họ có lợi thế đàm phán hơn chúng ta.', 1),
    ('thuong-luong', 'concession', '/kənˈseʃn/', 'sự nhượng bộ', 'That is a significant concession.', 'Đó là một nhượng bộ đáng kể.', 2),
    ('thuong-luong', 'negotiate', '/nɪˈɡoʊʃieɪt/', 'thương lượng', 'We will negotiate the price next week.', 'Tuần sau chúng ta sẽ thương lượng giá.', 3),
    ('thuong-luong', 'terms', '/tɜːrmz/', 'điều khoản', 'The terms are acceptable to us.', 'Các điều khoản chúng tôi chấp nhận được.', 4),
    ('thuong-luong', 'withdraw', '/wɪðˈdrɔː/', 'rút lại', 'They may withdraw the offer.', 'Họ có thể rút lại lời đề nghị.', 5),

    -- B2 · Tiếng Anh công sở · Phản hồi đồng nghiệp
    ('phan-hoi-dong-nghiep', 'constructive', '/kənˈstrʌktɪv/', 'mang tính xây dựng', 'Her feedback was constructive.', 'Phản hồi của cô ấy mang tính xây dựng.', 0),
    ('phan-hoi-dong-nghiep', 'acknowledge', '/əkˈnɑːlɪdʒ/', 'ghi nhận', 'We should acknowledge her effort.', 'Chúng ta nên ghi nhận nỗ lực của cô ấy.', 1),
    ('phan-hoi-dong-nghiep', 'candid', '/ˈkændɪd/', 'thẳng thắn', 'Thank you for being candid.', 'Cảm ơn vì đã thẳng thắn.', 2),
    ('phan-hoi-dong-nghiep', 'tactful', '/ˈtæktfl/', 'khéo léo', 'He was tactful about the mistake.', 'Anh ấy khéo léo khi nói về lỗi đó.', 3),
    ('phan-hoi-dong-nghiep', 'actionable', '/ˈækʃənəbl/', 'có thể hành động được', 'Give me one actionable suggestion.', 'Cho tôi một gợi ý có thể hành động được.', 4),
    ('phan-hoi-dong-nghiep', 'defensive', '/dɪˈfensɪv/', 'phòng thủ, đề phòng', 'Try not to sound defensive.', 'Cố đừng nghe có vẻ phòng thủ.', 5)
) AS v(lesson_slug, word, ipa, meaning, example, example_vi, position)
JOIN lessons AS l ON l.slug = v.lesson_slug AND l.deleted_at IS NULL
ON CONFLICT (lesson_id, lower(word)) WHERE deleted_at IS NULL DO NOTHING;

INSERT INTO vocabulary_entries (lesson_id, word, ipa, meaning, example, example_vi, position)
SELECT l.id, v.word, v.ipa, v.meaning, v.example, v.example_vi, v.position
FROM (VALUES
    -- C1 · Học thuật và nghiên cứu · Trích dẫn nguồn
    ('trich-dan-nguon', 'attribute', '/əˈtrɪbjuːt/', 'quy cho, ghi nguồn', 'The authors attribute the rise to policy.', 'Các tác giả quy mức tăng đó cho chính sách.', 0),
    ('trich-dan-nguon', 'paraphrase', '/ˈpærəfreɪz/', 'diễn đạt lại', 'You must paraphrase rather than copy.', 'Bạn phải diễn đạt lại chứ không chép.', 1),
    ('trich-dan-nguon', 'plagiarism', '/ˈpleɪdʒərɪzəm/', 'đạo văn', 'The journal has strict rules on plagiarism.', 'Tạp chí có quy định nghiêm ngặt về đạo văn.', 2),
    ('trich-dan-nguon', 'seminal', '/ˈsemɪnl/', 'có tính nền tảng', 'This is a seminal paper in the field.', 'Đây là một bài báo nền tảng của ngành.', 3),
    ('trich-dan-nguon', 'bibliography', '/ˌbɪbliˈɑːɡrəfi/', 'thư mục tham khảo', 'The bibliography lists forty sources.', 'Thư mục tham khảo liệt kê bốn mươi nguồn.', 4),
    ('trich-dan-nguon', 'verbatim', '/vɜːrˈbeɪtɪm/', 'nguyên văn', 'Quote it verbatim inside quotation marks.', 'Hãy trích nguyên văn trong dấu ngoặc kép.', 5),

    -- C1 · Học thuật và nghiên cứu · Mô tả dữ liệu
    ('mo-ta-du-lieu', 'fluctuate', '/ˈflʌktʃueɪt/', 'dao động', 'Prices fluctuate throughout the year.', 'Giá dao động suốt cả năm.', 0),
    ('mo-ta-du-lieu', 'plateau', '/plæˈtoʊ/', 'chững lại', 'Growth began to plateau in March.', 'Tăng trưởng bắt đầu chững lại vào tháng Ba.', 1),
    ('mo-ta-du-lieu', 'outlier', '/ˈaʊtlaɪər/', 'giá trị ngoại lai', 'We removed one clear outlier.', 'Chúng tôi loại bỏ một giá trị ngoại lai rõ rệt.', 2),
    ('mo-ta-du-lieu', 'threshold', '/ˈθreʃhoʊld/', 'ngưỡng', 'The value never crossed the threshold.', 'Giá trị chưa bao giờ vượt ngưỡng.', 3),
    ('mo-ta-du-lieu', 'negligible', '/ˈneɡlɪdʒəbl/', 'không đáng kể', 'The difference was negligible.', 'Khác biệt là không đáng kể.', 4),
    ('mo-ta-du-lieu', 'aggregate', '/ˈæɡrɪɡət/', 'tổng hợp', 'We report the aggregate figure only.', 'Chúng tôi chỉ báo cáo con số tổng hợp.', 5),

    -- C1 · Học thuật và nghiên cứu · Giới hạn của nghiên cứu
    ('gioi-han-nghien-cuu', 'caveat', '/ˈkæviæt/', 'lưu ý dè dặt', 'One caveat applies to these results.', 'Có một lưu ý dè dặt với các kết quả này.', 0),
    ('gioi-han-nghien-cuu', 'generalise', '/ˈdʒenrəlaɪz/', 'khái quát hoá', 'We cannot generalise from one sample.', 'Không thể khái quát hoá từ một mẫu duy nhất.', 1),
    ('gioi-han-nghien-cuu', 'confounding', '/kənˈfaʊndɪŋ/', 'gây nhiễu', 'Income may be a confounding factor.', 'Thu nhập có thể là yếu tố gây nhiễu.', 2),
    ('gioi-han-nghien-cuu', 'tentative', '/ˈtentətɪv/', 'dè dặt, tạm thời', 'These conclusions remain tentative.', 'Những kết luận này vẫn còn dè dặt.', 3),
    ('gioi-han-nghien-cuu', 'replicate', '/ˈreplɪkeɪt/', 'lặp lại (thí nghiệm)', 'Other teams failed to replicate the finding.', 'Các nhóm khác không lặp lại được phát hiện đó.', 4),
    ('gioi-han-nghien-cuu', 'scope', '/skoʊp/', 'phạm vi', 'That question is beyond the scope of this paper.', 'Câu hỏi đó nằm ngoài phạm vi bài báo này.', 5),

    -- C1 · Học thuật và nghiên cứu · Thuật ngữ học thuật
    ('thuat-ngu-hoc-thuat', 'empirical', '/ɪmˈpɪrɪkl/', 'thực nghiệm', 'The claim needs empirical support.', 'Khẳng định đó cần chứng cứ thực nghiệm.', 0),
    ('thuat-ngu-hoc-thuat', 'hypothesis', '/haɪˈpɑːθəsɪs/', 'giả thuyết', 'The hypothesis was not supported.', 'Giả thuyết không được ủng hộ.', 1),
    ('thuat-ngu-hoc-thuat', 'methodology', '/ˌmeθəˈdɑːlədʒi/', 'phương pháp luận', 'The methodology is described in section two.', 'Phương pháp luận được mô tả ở mục hai.', 2),
    ('thuat-ngu-hoc-thuat', 'rigorous', '/ˈrɪɡərəs/', 'chặt chẽ', 'They applied a rigorous standard.', 'Họ áp dụng một chuẩn mực chặt chẽ.', 3),
    ('thuat-ngu-hoc-thuat', 'peer', '/pɪr/', 'đồng nghiệp cùng ngành', 'The article passed peer review.', 'Bài báo đã qua phản biện đồng nghiệp.', 4),
    ('thuat-ngu-hoc-thuat', 'inference', '/ˈɪnfərəns/', 'suy luận', 'That inference is not warranted.', 'Suy luận đó không có cơ sở.', 5),

    -- C1 · Sắc thái và thành ngữ · Thành ngữ thông dụng
    ('thanh-ngu-thong-dung', 'ballpark', '/ˈbɔːlpɑːrk/', 'ước chừng', 'Give me a ballpark figure.', 'Cho tôi một con số ước chừng.', 0),
    ('thanh-ngu-thong-dung', 'ropes', '/roʊps/', 'cách làm việc, quy trình', 'It takes a month to learn the ropes.', 'Phải mất một tháng để nắm được cách làm việc.', 1),
    ('thanh-ngu-thong-dung', 'backburner', '/ˌbækˈbɜːrnər/', 'việc tạm gác lại', 'That project is on the backburner.', 'Dự án đó đang tạm gác lại.', 2),
    ('thanh-ngu-thong-dung', 'longshot', '/ˈlɔːŋʃɑːt/', 'khả năng rất mong manh', 'Winning is a longshot at this point.', 'Đến lúc này thắng là khả năng rất mong manh.', 3),
    ('thanh-ngu-thong-dung', 'uphill', '/ˌʌpˈhɪl/', 'gian nan', 'It has been an uphill battle.', 'Đó là một cuộc chiến gian nan.', 4),
    ('thanh-ngu-thong-dung', 'bandwidth', '/ˈbændwɪdθ/', 'quỹ thời gian và sức lực', 'I do not have the bandwidth this week.', 'Tuần này tôi không còn quỹ thời gian và sức lực.', 5),

    -- C1 · Sắc thái và thành ngữ · Cụm động từ
    ('cum-dong-tu', 'iron', '/ˈaɪərn/', 'giải quyết (ổn thoả)', 'We need to iron out the details.', 'Chúng ta cần giải quyết cho ổn các chi tiết.', 0),
    ('cum-dong-tu', 'follow', '/ˈfɑːloʊ/', 'theo sát, đôn đốc', 'Please follow up with the supplier.', 'Vui lòng theo sát nhà cung cấp.', 1),
    ('cum-dong-tu', 'water', '/ˈwɔːtər/', 'làm nhẹ đi', 'They tried to water down the proposal.', 'Họ cố làm nhẹ đi đề xuất.', 2),
    ('cum-dong-tu', 'gloss', '/ɡlɔːs/', 'lướt qua, che nhẹ', 'The report seems to gloss over costs.', 'Báo cáo có vẻ lướt qua phần chi phí.', 3),
    ('cum-dong-tu', 'weigh', '/weɪ/', 'lên tiếng góp ý', 'Would you like to weigh in on this?', 'Bạn có muốn lên tiếng góp ý việc này không?', 4),
    ('cum-dong-tu', 'phase', '/feɪz/', 'loại bỏ dần', 'They will phase out the old system.', 'Họ sẽ loại bỏ dần hệ thống cũ.', 5),

    -- C1 · Sắc thái và thành ngữ · Nói giảm nói tránh
    ('noi-giam-noi-tranh', 'modest', '/ˈmɑːdɪst/', 'khiêm tốn, vừa phải', 'The results were modest.', 'Kết quả ở mức vừa phải.', 0),
    ('noi-giam-noi-tranh', 'somewhat', '/ˈsʌmwʌt/', 'hơi, phần nào', 'The answer was somewhat vague.', 'Câu trả lời hơi mơ hồ.', 1),
    ('noi-giam-noi-tranh', 'reservation', '/ˌrezərˈveɪʃn/', 'sự dè dặt', 'I have one reservation about the plan.', 'Tôi có một điểm dè dặt về kế hoạch.', 2),
    ('noi-giam-noi-tranh', 'arguably', '/ˈɑːrɡjuəbli/', 'có thể cho rằng', 'That was arguably premature.', 'Có thể cho rằng điều đó hơi vội.', 3),
    ('noi-giam-noi-tranh', 'unfortunate', '/ʌnˈfɔːrtʃənət/', 'đáng tiếc', 'The wording was unfortunate.', 'Cách diễn đạt thật đáng tiếc.', 4),
    ('noi-giam-noi-tranh', 'hesitant', '/ˈhezɪtənt/', 'ngần ngại', 'I am hesitant to commit today.', 'Tôi ngần ngại cam kết ngay hôm nay.', 5),

    -- C1 · Sắc thái và thành ngữ · Sắc thái trang trọng
    ('sac-thai-trang-trong', 'furthermore', '/ˈfɜːrðərmɔːr/', 'hơn nữa', 'Furthermore, the cost has doubled.', 'Hơn nữa, chi phí đã tăng gấp đôi.', 0),
    ('sac-thai-trang-trong', 'notwithstanding', '/ˌnɑːtwɪθˈstændɪŋ/', 'mặc dù vậy', 'Notwithstanding the delay, we shipped.', 'Mặc dù bị chậm, chúng tôi vẫn giao được.', 1),
    ('sac-thai-trang-trong', 'henceforth', '/ˌhensˈfɔːrθ/', 'từ nay trở đi', 'Henceforth all requests go through the portal.', 'Từ nay mọi yêu cầu đi qua cổng thông tin.', 2),
    ('sac-thai-trang-trong', 'pertaining', '/pərˈteɪnɪŋ/', 'liên quan tới', 'Documents pertaining to the case are sealed.', 'Tài liệu liên quan tới vụ việc được niêm phong.', 3),
    ('sac-thai-trang-trong', 'commence', '/kəˈmens/', 'bắt đầu', 'The session will commence at nine.', 'Phiên họp sẽ bắt đầu lúc chín giờ.', 4),
    ('sac-thai-trang-trong', 'deem', '/diːm/', 'coi là, xem là', 'The board deemed it necessary.', 'Hội đồng xem đó là cần thiết.', 5)
) AS v(lesson_slug, word, ipa, meaning, example, example_vi, position)
JOIN lessons AS l ON l.slug = v.lesson_slug AND l.deleted_at IS NULL
ON CONFLICT (lesson_id, lower(word)) WHERE deleted_at IS NULL DO NOTHING;

INSERT INTO vocabulary_entries (lesson_id, word, ipa, meaning, example, example_vi, position)
SELECT l.id, v.word, v.ipa, v.meaning, v.example, v.example_vi, v.position
FROM (VALUES
    -- C2 · Văn phong và tu từ · Biện pháp tu từ
    ('bien-phap-tu-tu', 'metaphor', '/ˈmetəfɔːr/', 'ẩn dụ', 'The whole speech turns on one metaphor.', 'Cả bài diễn văn xoay quanh một ẩn dụ.', 0),
    ('bien-phap-tu-tu', 'irony', '/ˈaɪrəni/', 'sự mỉa mai', 'There is a quiet irony in that line.', 'Có một sự mỉa mai lặng lẽ trong câu đó.', 1),
    ('bien-phap-tu-tu', 'understatement', '/ˈʌndərsteɪtmənt/', 'lối nói giảm', 'Calling it difficult is an understatement.', 'Gọi nó là khó thì đúng là lối nói giảm.', 2),
    ('bien-phap-tu-tu', 'juxtapose', '/ˈdʒʌkstəpoʊz/', 'đặt cạnh nhau', 'The author likes to juxtapose old and new.', 'Tác giả thích đặt cạnh nhau cái cũ và cái mới.', 3),
    ('bien-phap-tu-tu', 'allusion', '/əˈluːʒn/', 'sự ám chỉ', 'The title is an allusion to a poem.', 'Nhan đề là một sự ám chỉ tới một bài thơ.', 4),
    ('bien-phap-tu-tu', 'hyperbole', '/haɪˈpɜːrbəli/', 'lối nói phóng đại', 'The advert is pure hyperbole.', 'Mẩu quảng cáo thuần là lối nói phóng đại.', 5),

    -- C2 · Văn phong và tu từ · Nhịp điệu câu văn
    ('nhip-dieu-cau-van', 'cadence', '/ˈkeɪdəns/', 'nhịp điệu', 'Her prose has a steady cadence.', 'Văn xuôi của bà có nhịp điệu đều đặn.', 0),
    ('nhip-dieu-cau-van', 'terse', '/tɜːrs/', 'cô đọng, cộc', 'His reply was terse but clear.', 'Câu trả lời của anh ấy cô đọng nhưng rõ.', 1),
    ('nhip-dieu-cau-van', 'verbose', '/vɜːrˈboʊs/', 'dài dòng', 'The opening paragraph is verbose.', 'Đoạn mở đầu dài dòng.', 2),
    ('nhip-dieu-cau-van', 'parallelism', '/ˈpærəlelɪzəm/', 'phép song hành', 'The sentence gains force from parallelism.', 'Câu văn mạnh lên nhờ phép song hành.', 3),
    ('nhip-dieu-cau-van', 'clause', '/klɔːz/', 'mệnh đề', 'Cut the final clause and it lands harder.', 'Bỏ mệnh đề cuối đi thì câu mạnh hơn.', 4),
    ('nhip-dieu-cau-van', 'staccato', '/stəˈkɑːtoʊ/', 'ngắt quãng, dồn dập', 'Short staccato sentences build tension.', 'Những câu ngắn dồn dập tạo ra căng thẳng.', 5),

    -- C2 · Văn phong và tu từ · Vốn từ tinh tế
    ('tu-vung-tinh-te', 'ostensibly', '/ɑːˈstensəbli/', 'bề ngoài là', 'The trip was ostensibly about research.', 'Chuyến đi bề ngoài là để nghiên cứu.', 0),
    ('tu-vung-tinh-te', 'ubiquitous', '/juːˈbɪkwɪtəs/', 'có mặt khắp nơi', 'The logo is ubiquitous in the city.', 'Logo đó có mặt khắp nơi trong thành phố.', 1),
    ('tu-vung-tinh-te', 'innocuous', '/ɪˈnɑːkjuəs/', 'vô hại', 'It looked like an innocuous remark.', 'Nó trông như một nhận xét vô hại.', 2),
    ('tu-vung-tinh-te', 'meticulous', '/məˈtɪkjələs/', 'tỉ mỉ', 'She keeps meticulous notes.', 'Cô ấy ghi chép rất tỉ mỉ.', 3),
    ('tu-vung-tinh-te', 'ephemeral', '/ɪˈfemərəl/', 'phù du, chóng tàn', 'Online fame is ephemeral.', 'Danh tiếng trên mạng là thứ phù du.', 4),
    ('tu-vung-tinh-te', 'incisive', '/ɪnˈsaɪsɪv/', 'sắc sảo', 'He asked one incisive question.', 'Anh ấy đặt một câu hỏi sắc sảo.', 5),

    -- C2 · Văn phong và tu từ · Văn phong báo chí
    ('van-phong-bao-chi', 'byline', '/ˈbaɪlaɪn/', 'dòng ghi tên tác giả', 'The byline names two reporters.', 'Dòng ghi tên tác giả nêu hai phóng viên.', 0),
    ('van-phong-bao-chi', 'scrutiny', '/ˈskruːtəni/', 'sự soi xét', 'The claim did not survive scrutiny.', 'Khẳng định đó không trụ được trước sự soi xét.', 1),
    ('van-phong-bao-chi', 'corroborate', '/kəˈrɑːbəreɪt/', 'xác thực thêm', 'Two sources corroborate the account.', 'Hai nguồn xác thực thêm lời kể đó.', 2),
    ('van-phong-bao-chi', 'sensational', '/senˈseɪʃənl/', 'giật gân', 'The headline is deliberately sensational.', 'Tiêu đề cố ý giật gân.', 3),
    ('van-phong-bao-chi', 'impartial', '/ɪmˈpɑːrʃl/', 'khách quan', 'The coverage stayed impartial.', 'Bài đưa tin giữ được sự khách quan.', 4),
    ('van-phong-bao-chi', 'embargo', '/ɪmˈbɑːrɡoʊ/', 'lệnh hoãn đăng', 'The story is under embargo until Friday.', 'Bài viết bị hoãn đăng tới thứ Sáu.', 5),

    -- C2 · Diễn thuyết và thuyết phục · Mở đầu ấn tượng
    ('mo-dau-an-tuong', 'anecdote', '/ˈænɪkdoʊt/', 'mẩu chuyện kể', 'She opened with a short anecdote.', 'Cô ấy mở đầu bằng một mẩu chuyện ngắn.', 0),
    ('mo-dau-an-tuong', 'rapport', '/ræˈpɔːr/', 'sự gắn kết, thiện cảm', 'He built rapport in the first minute.', 'Anh ấy tạo được thiện cảm ngay phút đầu.', 1),
    ('mo-dau-an-tuong', 'premise', '/ˈpremɪs/', 'tiền đề', 'Start by stating your premise.', 'Hãy bắt đầu bằng việc nêu tiền đề của bạn.', 2),
    ('mo-dau-an-tuong', 'provocative', '/prəˈvɑːkətɪv/', 'khiêu khích, gây suy nghĩ', 'The opening question was provocative.', 'Câu hỏi mở đầu mang tính khiêu khích.', 3),
    ('mo-dau-an-tuong', 'concise', '/kənˈsaɪs/', 'súc tích', 'Keep the introduction concise.', 'Hãy giữ phần mở đầu súc tích.', 4),
    ('mo-dau-an-tuong', 'hook', '/hʊk/', 'chi tiết gây chú ý', 'Every talk needs a hook.', 'Bài nói nào cũng cần một chi tiết gây chú ý.', 5),

    -- C2 · Diễn thuyết và thuyết phục · Dựng lập luận
    ('dung-lap-luan', 'coherent', '/koʊˈhɪrənt/', 'mạch lạc', 'The argument is coherent throughout.', 'Lập luận mạch lạc từ đầu tới cuối.', 0),
    ('dung-lap-luan', 'fallacy', '/ˈfæləsi/', 'nguỵ biện', 'That is a common logical fallacy.', 'Đó là một nguỵ biện logic thường gặp.', 1),
    ('dung-lap-luan', 'refute', '/rɪˈfjuːt/', 'bác bỏ', 'He could not refute the figures.', 'Anh ấy không bác bỏ được các con số.', 2),
    ('dung-lap-luan', 'premise', '/ˈpremɪs/', 'tiền đề', 'If the premise fails, so does the rest.', 'Nếu tiền đề sai thì phần còn lại cũng sai.', 3),
    ('dung-lap-luan', 'cogent', '/ˈkoʊdʒənt/', 'thuyết phục, chặt chẽ', 'She made a cogent case for delay.', 'Cô ấy lập luận chặt chẽ cho việc hoãn lại.', 4),
    ('dung-lap-luan', 'rebuttal', '/rɪˈbʌtl/', 'lời phản bác', 'His rebuttal took two minutes.', 'Lời phản bác của anh ấy dài hai phút.', 5),

    -- C2 · Diễn thuyết và thuyết phục · Kéo hướng người nghe
    ('keo-huong-nguoi-nghe', 'resonate', '/ˈrezəneɪt/', 'gây đồng cảm', 'That story will resonate with parents.', 'Câu chuyện đó sẽ gây đồng cảm với các bậc phụ huynh.', 0),
    ('keo-huong-nguoi-nghe', 'rhetorical', '/rɪˈtɔːrɪkl/', 'tu từ', 'He ended on a rhetorical question.', 'Anh ấy kết bằng một câu hỏi tu từ.', 1),
    ('keo-huong-nguoi-nghe', 'emphasis', '/ˈemfəsɪs/', 'sự nhấn mạnh', 'Put the emphasis on the last word.', 'Hãy đặt sự nhấn mạnh vào từ cuối.', 2),
    ('keo-huong-nguoi-nghe', 'pause', '/pɔːz/', 'khoảng lặng', 'A short pause does more than volume.', 'Một khoảng lặng ngắn còn hiệu quả hơn nói to.', 3),
    ('keo-huong-nguoi-nghe', 'compelling', '/kəmˈpelɪŋ/', 'cuốn hút', 'The evidence was compelling.', 'Bằng chứng rất cuốn hút.', 4),
    ('keo-huong-nguoi-nghe', 'digress', '/daɪˈɡres/', 'lạc đề', 'Try not to digress in the middle.', 'Cố đừng lạc đề ở phần giữa.', 5),

    -- C2 · Diễn thuyết và thuyết phục · Kết bài đọng lại
    ('ket-bai-dong-lai', 'reiterate', '/riˈɪtəreɪt/', 'nhắc lại', 'Let me reiterate the main point.', 'Để tôi nhắc lại ý chính.', 0),
    ('ket-bai-dong-lai', 'imperative', '/ɪmˈperətɪv/', 'điều thiết yếu', 'Acting now is imperative.', 'Hành động ngay là điều thiết yếu.', 1),
    ('ket-bai-dong-lai', 'lasting', '/ˈlæstɪŋ/', 'lâu dài', 'We want a lasting change.', 'Chúng tôi muốn một thay đổi lâu dài.', 2),
    ('ket-bai-dong-lai', 'linger', '/ˈlɪŋɡər/', 'đọng lại', 'The final image should linger.', 'Hình ảnh cuối cùng nên đọng lại.', 3),
    ('ket-bai-dong-lai', 'succinct', '/səkˈsɪŋkt/', 'ngắn gọn', 'Keep the closing succinct.', 'Hãy giữ phần kết ngắn gọn.', 4),
    ('ket-bai-dong-lai', 'resolve', '/rɪˈzɑːlv/', 'quyết tâm', 'The speech strengthened their resolve.', 'Bài nói củng cố quyết tâm của họ.', 5)
) AS v(lesson_slug, word, ipa, meaning, example, example_vi, position)
JOIN lessons AS l ON l.slug = v.lesson_slug AND l.deleted_at IS NULL
ON CONFLICT (lesson_id, lower(word)) WHERE deleted_at IS NULL DO NOTHING;
