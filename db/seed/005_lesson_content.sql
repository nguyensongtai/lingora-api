-- Nội dung bài học: phần dẫn và các khối giải thích, câu mẫu, hội thoại.
--
-- Do Claude soạn. Bậc CEFR của từ vựng thì đối chiếu được bằng máy
-- (`make verify-vocab`); nội dung dạy thì KHÔNG có nguồn nào để đối chiếu —
-- giải thích ngữ pháp và dựng hội thoại là việc sư phạm. Cần người có chuyên
-- môn đọc lại trước khi có người học thật.
--
-- Thay toàn bộ nội dung của những bài nó phụ trách, nên chạy lại là hội tụ.

DELETE FROM lesson_blocks b
USING lessons l, courses c
WHERE b.lesson_id = l.id AND l.course_id = c.id
  AND c.slug IN (
      'ngu-phap-co-ban', 'tu-vung-doi-song', 'giao-tiep-hang-ngay',
      'qua-khu-va-tuong-lai', 'ke-chuyen-va-mo-ta', 'email-va-tin-nhan',
      'tranh-luan-va-y-kien', 'tieng-anh-cong-so', 'hoc-thuat-va-nghien-cuu',
      'sac-thai-va-thanh-ngu', 'van-phong-va-tu-tu', 'dien-thuyet-va-thuyet-phuc'
  );

UPDATE lessons SET summary = v.summary, updated_at = now()
FROM (VALUES
    ('thi-hien-tai-don',        'Thì dùng cho thói quen và sự thật hiển nhiên — thì bạn sẽ gặp nhiều nhất.'),
    ('danh-tu-dem-duoc',        'Vì sao nói "two bottles" được mà "two waters" thì không.'),
    ('mao-tu-a-an-the',         'a, an hay the: chọn theo âm đầu và theo việc người nghe đã biết chưa.'),
    ('dong-tu-to-be',           'Động từ đổi hình nhiều nhất trong tiếng Anh, và cũng là động từ dùng sớm nhất.'),
    ('gia-dinh-va-ban-be',      'Từ để giới thiệu người thân và người quen quanh bạn.'),
    ('mau-sac-va-hinh-dang',    'Tính từ mô tả cơ bản, và thứ tự đặt chúng trước danh từ.'),
    ('do-an-va-thuc-uong',      'Từ cho bữa ăn hằng ngày và cách nói mình thích hay không thích.'),
    ('thoi-gian-va-ngay-thang', 'Nói giờ, nói ngày, và những giới từ đi kèm.'),
    ('chao-hoi-lam-quen',       'Mở đầu một cuộc gặp và giữ cho nó không tắt sau ba câu.'),
    ('goi-mon-o-nha-hang',      'Từ lúc được dẫn vào bàn tới lúc xin hoá đơn.'),
    ('hoi-duong-di',            'Hỏi đường và hiểu được câu trả lời — phần khó hơn nằm ở vế sau.'),
    ('mua-sam',                 'Hỏi giá, hỏi cỡ, đổi trả, và trả tiền.'),
    ('thi-qua-khu-don',         'Kể lại việc đã xong: quy tắc thêm -ed và ba cách đọc đuôi đó.'),
    ('dong-tu-bat-quy-tac',     'Những động từ không theo quy tắc -ed, nhóm theo kiểu biến đổi cho dễ nhớ.'),
    ('noi-ve-du-dinh',          'will, be going to và hiện tại tiếp diễn: ba cách nói về tương lai.'),
    ('ke-ve-ky-nghi',           'Ghép quá khứ đơn với từ vựng du lịch thành một câu chuyện ngắn.')
) AS v(slug, summary)
WHERE lessons.slug = v.slug AND lessons.deleted_at IS NULL;

INSERT INTO lesson_blocks (lesson_id, kind, body, text_en, text_vi, speaker, position)
SELECT l.id, v.kind::lesson_block_kind, v.body, v.text_en, v.text_vi, v.speaker, v.position
FROM (VALUES
    -- A1 · Thì hiện tại đơn
    ('thi-hien-tai-don','note','Hiện tại đơn nói về việc lặp đi lặp lại hoặc luôn đúng, không phải việc đang xảy ra lúc này. Điểm duy nhất phải nhớ: chủ ngữ là he, she, it thì động từ thêm -s.','','','',0),
    ('thi-hien-tai-don','note','Trạng từ tần suất (always, usually, often, never) đứng TRƯỚC động từ thường, nhưng đứng SAU động từ to be. Đây là lỗi hay gặp nhất ở bài này.','','','',1),
    ('thi-hien-tai-don','example','','I usually walk to school.','Tôi thường đi bộ đến trường.','',2),
    ('thi-hien-tai-don','example','','She always drinks tea after lunch.','Cô ấy luôn uống trà sau bữa trưa.','',3),
    ('thi-hien-tai-don','example','','He never forgets her birthday.','Anh ấy không bao giờ quên sinh nhật cô ấy.','',4),
    ('thi-hien-tai-don','example','','They are often late.','Họ thường đến muộn.','',5),

    -- A1 · Danh từ đếm được và không đếm được
    ('danh-tu-dem-duoc','note','Danh từ đếm được có số nhiều: one bottle, two bottles. Danh từ không đếm được thì không: sugar, bread, rice, water. Chúng luôn đi với động từ số ít.','','','',0),
    ('danh-tu-dem-duoc','note','Muốn đếm một danh từ không đếm được, phải mượn một đơn vị: a glass of water, a piece of bread, a bottle of juice.','','','',1),
    ('danh-tu-dem-duoc','example','','There is one bottle of water on the table.','Có một chai nước trên bàn.','',2),
    ('danh-tu-dem-duoc','example','','He ate one piece of cake.','Anh ấy ăn một miếng bánh.','',3),
    ('danh-tu-dem-duoc','example','','I do not take sugar in my coffee.','Tôi không cho đường vào cà phê.','',4),

    -- A1 · Mạo từ
    ('mao-tu-a-an-the','note','a hay an không chọn theo chữ cái mà theo ÂM đầu. an hour vì h câm nên nghe ra nguyên âm; a university vì u ở đây đọc là "yu", một phụ âm.','','','',0),
    ('mao-tu-a-an-the','note','Dùng a/an khi lần đầu nhắc tới, dùng the khi cả hai người đều đã biết đang nói vật nào.','','','',1),
    ('mao-tu-a-an-the','example','','We waited for an hour.','Chúng tôi đã chờ một tiếng.','',2),
    ('mao-tu-a-an-the','example','','Take an umbrella with you.','Mang theo một cái ô nhé.','',3),
    ('mao-tu-a-an-the','example','','I bought a ticket for the show.','Tôi đã mua một vé xem buổi diễn.','',4),

    -- A1 · Động từ to be
    ('dong-tu-to-be','note','to be có ba hình ở hiện tại: I am, he/she/it is, you/we/they are. Nó nối chủ ngữ với một tính từ hoặc một danh từ, chứ không mô tả hành động.','','','',0),
    ('dong-tu-to-be','note','Tiếng Việt nói "Tôi đói" không cần động từ, nên người học hay bỏ quên am/is/are. Trong tiếng Anh nó bắt buộc.','','','',1),
    ('dong-tu-to-be','example','','I am tired after the trip.','Tôi mệt sau chuyến đi.','',2),
    ('dong-tu-to-be','example','','The children are hungry.','Bọn trẻ đang đói.','',3),
    ('dong-tu-to-be','example','','She is busy this afternoon.','Chiều nay cô ấy bận.','',4),

    -- A1 · Gia đình và bạn bè
    ('gia-dinh-va-ban-be','note','Tiếng Anh gộp nhiều vai mà tiếng Việt tách riêng: uncle là cả chú, bác lẫn cậu; cousin là mọi anh chị em họ, không phân biệt tuổi hay bên nội ngoại.','','','',0),
    ('gia-dinh-va-ban-be','example','','My cousin is the same age as me.','Anh họ tôi bằng tuổi tôi.','',1),
    ('gia-dinh-va-ban-be','example','','My uncle lives in Da Nang.','Chú tôi sống ở Đà Nẵng.','',2),
    ('gia-dinh-va-ban-be','dialogue','','Do you have any brothers or sisters?','Bạn có anh chị em không?','Mai',3),
    ('gia-dinh-va-ban-be','dialogue','','One sister. She is a teacher.','Một chị gái. Chị ấy là giáo viên.','Tom',4),

    -- A1 · Màu sắc và hình dáng
    ('mau-sac-va-hinh-dang','note','Tính từ trong tiếng Anh đứng TRƯỚC danh từ và không đổi theo số nhiều: a bright dress, two bright dresses.','','','',0),
    ('mau-sac-va-hinh-dang','note','Khi có nhiều tính từ, thứ tự quen thuộc là: ý kiến → kích thước → màu sắc. "a nice long dark coat" nghe tự nhiên, đảo lại thì không.','','','',1),
    ('mau-sac-va-hinh-dang','example','','She wore a bright yellow dress.','Cô ấy mặc một chiếc váy vàng rực rỡ.','',2),
    ('mau-sac-va-hinh-dang','example','','He likes dark blue shirts.','Anh ấy thích áo sơ mi màu xanh sẫm.','',3),
    ('mau-sac-va-hinh-dang','example','','She has long hair.','Cô ấy có mái tóc dài.','',4),

    -- A1 · Đồ ăn và thức uống
    ('do-an-va-thuc-uong','note','Nói thích hay không thích: I like, I do not like, I love. Sau chúng dùng danh từ số nhiều hoặc danh từ không đếm được: I like vegetables, I like juice.','','','',0),
    ('do-an-va-thuc-uong','example','','This soup is delicious.','Món canh này rất ngon.','',1),
    ('do-an-va-thuc-uong','example','','He does not eat meat.','Anh ấy không ăn thịt.','',2),
    ('do-an-va-thuc-uong','dialogue','','What do you want for dinner?','Bữa tối bạn muốn ăn gì?','Mai',3),
    ('do-an-va-thuc-uong','dialogue','','Something light. Maybe fruit and bread.','Gì đó nhẹ thôi. Có thể là trái cây với bánh mì.','Tom',4),

    -- A1 · Thời gian và ngày tháng
    ('thoi-gian-va-ngay-thang','note','Ba giới từ hay nhầm: at cho giờ (at seven), on cho ngày (on Monday), in cho tháng và năm (in March, in 2026).','','','',0),
    ('thoi-gian-va-ngay-thang','note','in the morning, in the evening — nhưng at night. Đây là ngoại lệ phải nhớ riêng.','','','',1),
    ('thoi-gian-va-ngay-thang','example','','I study in the evening.','Tôi học vào buổi tối.','',2),
    ('thoi-gian-va-ngay-thang','example','','We moved here last month.','Chúng tôi chuyển đến đây tháng trước.','',3),
    ('thoi-gian-va-ngay-thang','example','','The test is tomorrow.','Bài kiểm tra vào ngày mai.','',4)
) AS v(lesson_slug, kind, body, text_en, text_vi, speaker, position)
JOIN lessons AS l ON l.slug = v.lesson_slug AND l.deleted_at IS NULL;

INSERT INTO lesson_blocks (lesson_id, kind, body, text_en, text_vi, speaker, position)
SELECT l.id, v.kind::lesson_block_kind, v.body, v.text_en, v.text_vi, v.speaker, v.position
FROM (VALUES
    -- A2 · Chào hỏi và làm quen
    ('chao-hoi-lam-quen','note','Câu chào trong tiếng Anh thường đi kèm một câu hỏi mở, để cuộc trò chuyện có chỗ đi tiếp. Chỉ nói "Hello" rồi dừng là một câu chào chưa xong.','','','',0),
    ('chao-hoi-lam-quen','note','"How are you?" phần lớn là câu xã giao, không phải câu hỏi thật về sức khoẻ. Trả lời ngắn rồi hỏi lại là đủ.','','','',1),
    ('chao-hoi-lam-quen','dialogue','','Hi, I do not think we have met. I am Mai.','Chào bạn, hình như mình chưa gặp nhau. Mình là Mai.','Mai',2),
    ('chao-hoi-lam-quen','dialogue','','Nice to meet you, Mai. I am Tom. What do you do?','Rất vui được gặp Mai. Mình là Tom. Bạn làm nghề gì?','Tom',3),
    ('chao-hoi-lam-quen','dialogue','','I work at a school near here. And you?','Mình làm ở một trường gần đây. Còn bạn?','Mai',4),
    ('chao-hoi-lam-quen','example','','Could you spell your name for me?','Bạn đánh vần tên giúp mình được không?','',5),

    -- A2 · Gọi món ở nhà hàng
    ('goi-mon-o-nha-hang','note','Gọi món lịch sự thường bắt đầu bằng "Could I have..." hoặc "I will have...". Nói thẳng "I want" nghe cộc, dù không sai ngữ pháp.','','','',0),
    ('goi-mon-o-nha-hang','dialogue','','Are you ready to order?','Anh chị gọi món chưa ạ?','Phục vụ',1),
    ('goi-mon-o-nha-hang','dialogue','','Almost. What would you recommend?','Gần rồi. Bạn gợi ý món gì?','Khách',2),
    ('goi-mon-o-nha-hang','dialogue','','The chef recommends the fish today.','Hôm nay đầu bếp gợi ý món cá ạ.','Phục vụ',3),
    ('goi-mon-o-nha-hang','example','','Could we have the bill, please?','Cho chúng tôi xin hoá đơn.','',4),
    ('goi-mon-o-nha-hang','example','','I am allergic to peanuts.','Tôi bị dị ứng đậu phộng.','',5),

    -- A2 · Hỏi đường
    ('hoi-duong-di','note','Hỏi đường thì dễ, hiểu câu trả lời mới khó. Nghe kỹ ba thứ: hướng rẽ, số lần rẽ, và mốc để nhận ra chỗ cần dừng.','','','',0),
    ('hoi-duong-di','note','Nếu không kịp hiểu, câu cứu nguy là "Sorry, could you say that again more slowly?" — dùng nó còn hơn gật đầu rồi đi lạc.','','','',1),
    ('hoi-duong-di','dialogue','','Excuse me, how do I get to the station?','Xin lỗi, đi tới nhà ga thế nào ạ?','Khách',2),
    ('hoi-duong-di','dialogue','','Go straight, then turn left. It is opposite the market.','Đi thẳng rồi rẽ trái. Nó đối diện chợ.','Người dân',3),
    ('hoi-duong-di','example','','Walk across the bridge.','Đi băng qua cây cầu.','',4),
    ('hoi-duong-di','example','','I think we are going in the wrong direction.','Tôi nghĩ chúng ta đang đi sai hướng.','',5),

    -- A2 · Mua sắm
    ('mua-sam','note','Ba câu đủ dùng cho hầu hết tình huống mua sắm: hỏi giá, hỏi cỡ khác, và hỏi chính sách đổi trả.','','','',0),
    ('mua-sam','dialogue','','Do you have this in a larger size?','Cái này có cỡ lớn hơn không ạ?','Khách',1),
    ('mua-sam','dialogue','','Let me check. Would you like to pay in cash or by card?','Để tôi xem. Anh chị trả tiền mặt hay thẻ ạ?','Nhân viên',2),
    ('mua-sam','example','','Keep the receipt for the warranty.','Giữ biên lai để bảo hành.','',3),
    ('mua-sam','example','','That bag is outside my budget.','Cái túi đó vượt ngân sách của tôi.','',4),

    -- A2 · Thì quá khứ đơn
    ('thi-qua-khu-don','note','Động từ có quy tắc thêm -ed. Đuôi đó có ba cách đọc: /t/ sau âm vô thanh (worked), /d/ sau âm hữu thanh (played), và /ɪd/ sau t hoặc d (decided).','','','',0),
    ('thi-qua-khu-don','note','Ở câu phủ định và câu hỏi, did đã mang nghĩa quá khứ nên động từ chính quay về nguyên thể: "He did not admit it", không phải "did not admitted".','','','',1),
    ('thi-qua-khu-don','example','','They decided to stay one more day.','Họ quyết định ở thêm một ngày.','',2),
    ('thi-qua-khu-don','example','','She complained about the noise.','Cô ấy phàn nàn về tiếng ồn.','',3),
    ('thi-qua-khu-don','example','','He did not admit his mistake.','Anh ấy không chịu thừa nhận lỗi của mình.','',4),

    -- A2 · Động từ bất quy tắc
    ('dong-tu-bat-quy-tac','note','Học động từ bất quy tắc theo NHÓM biến đổi dễ hơn học rời: buy–bought, bring–brought, teach–taught đều đổi thành -ought/-aught.','','','',0),
    ('dong-tu-bat-quy-tac','note','Một nhóm khác giữ nguyên ở cả ba dạng: cut–cut–cut, put–put–put. Nhóm này không phải học gì thêm.','','','',1),
    ('dong-tu-bat-quy-tac','example','','The dog will not bite you.','Con chó sẽ không cắn bạn đâu.','',2),
    ('dong-tu-bat-quy-tac','example','','Do not burn the rice.','Đừng làm cháy cơm.','',3),
    ('dong-tu-bat-quy-tac','example','','Do not blame the weather.','Đừng đổ lỗi cho thời tiết.','',4),

    -- A2 · Nói về dự định
    ('noi-ve-du-dinh','note','Ba cách nói tương lai, khác nhau ở mức chắc chắn: will cho quyết định ngay lúc nói, be going to cho dự định đã có sẵn, hiện tại tiếp diễn cho lịch đã chốt.','','','',0),
    ('noi-ve-du-dinh','note','"We are moving in June" nghe chắc hơn "We will move in June", vì nó hàm ý đã sắp xếp xong.','','','',1),
    ('noi-ve-du-dinh','example','','She will probably call tonight.','Có lẽ tối nay cô ấy sẽ gọi.','',2),
    ('noi-ve-du-dinh','example','','They intend to open a shop.','Họ có ý định mở một cửa hàng.','',3),
    ('noi-ve-du-dinh','example','','We will continue after lunch.','Chúng ta sẽ tiếp tục sau bữa trưa.','',4),

    -- A2 · Kể về kỳ nghỉ
    ('ke-ve-ky-nghi','note','Kể chuyện thì đừng liệt kê. Chọn một chi tiết đáng nhớ rồi dựng câu quanh nó — người nghe nhớ chi tiết, không nhớ lịch trình.','','','',0),
    ('ke-ve-ky-nghi','dialogue','','How was your trip?','Chuyến đi của bạn thế nào?','Mai',1),
    ('ke-ve-ky-nghi','dialogue','','Long but good. We drove along the coast for two days.','Dài nhưng vui. Bọn mình lái xe dọc bờ biển hai ngày.','Tom',2),
    ('ke-ve-ky-nghi','dialogue','','Was it crowded?','Có đông không?','Mai',3),
    ('ke-ve-ky-nghi','dialogue','','Not at all. That was the best part.','Không hề. Đó mới là điểm hay nhất.','Tom',4),
    ('ke-ve-ky-nghi','example','','We went there to relax.','Chúng tôi đến đó để thư giãn.','',5)
) AS v(lesson_slug, kind, body, text_en, text_vi, speaker, position)
JOIN lessons AS l ON l.slug = v.lesson_slug AND l.deleted_at IS NULL;

UPDATE lessons SET summary = v.summary, updated_at = now()
FROM (VALUES
    ('mo-ta-nguoi',            'Tả tính cách chứ không chỉ tả ngoại hình, và cách nói giảm khi nhận xét tiêu cực.'),
    ('mo-ta-noi-chon',         'Dựng không khí của một nơi bằng vài tính từ đúng chỗ.'),
    ('ke-mot-su-viec',         'Trạng từ nối để câu chuyện có nhịp: suddenly, eventually, meanwhile.'),
    ('cam-xuc-va-thai-do',     'Gọi tên cảm xúc chính xác hơn là chỉ nói happy hay sad.'),
    ('email-trang-trong',      'Bố cục một email công việc và những cụm mở đầu, kết thúc chuẩn.'),
    ('email-than-mat',         'Viết cho bạn bè: ngắn, thoáng, và được phép lược bỏ.'),
    ('hen-lich-hop',           'Đề xuất giờ, xác nhận, và dời lịch mà không làm phiền ai.'),
    ('xin-loi-va-cam-on',      'Xin lỗi đúng mức: nhận lỗi, nói cách khắc phục, rồi dừng.'),
    ('neu-quan-diem',          'Nêu quan điểm kèm mức độ chắc chắn, thay vì khẳng định suông.'),
    ('phan-bien',              'Phản bác ý kiến mà không công kích người nói.'),
    ('dan-chung',              'Phân biệt bằng chứng với giai thoại, và cách dẫn số liệu.'),
    ('chot-van-de',            'Tóm lại và đóng cuộc thảo luận mà mọi người thấy được lắng nghe.'),
    ('hop-hanh',               'Ngôn ngữ trong cuộc họp: xin nói, làm rõ, và hoãn một mục.'),
    ('bao-cao-tien-do',        'Báo tin xấu về tiến độ sao cho vẫn giữ được lòng tin.'),
    ('thuong-luong',           'Nhượng bộ có điều kiện, và biết khi nào nên dừng.'),
    ('phan-hoi-dong-nghiep',   'Góp ý cụ thể và hành động được, thay vì khen chê chung chung.'),
    ('trich-dan-nguon',        'Dẫn nguồn, diễn đạt lại, và ranh giới của đạo văn.'),
    ('mo-ta-du-lieu',          'Mô tả xu hướng và con số mà không nói quá.'),
    ('gioi-han-nghien-cuu',    'Nói về điểm yếu của chính nghiên cứu mình — dấu hiệu của người làm nghiêm túc.'),
    ('thuat-ngu-hoc-thuat',    'Những thuật ngữ gặp ở mọi bài báo, bất kể ngành nào.'),
    ('thanh-ngu-thong-dung',   'Tính từ mạnh dùng trong công việc, và mức độ của chúng.'),
    ('cum-dong-tu',            'Động từ trang trọng thay cho cụm động từ thân mật.'),
    ('noi-giam-noi-tranh',     'Nói nhẹ đi để giữ cửa cho đối phương — một phản xạ của tiếng Anh trang trọng.'),
    ('sac-thai-trang-trong',   'Từ nối và cách diễn đạt của văn bản chính thức.'),
    ('bien-phap-tu-tu',        'Ẩn dụ, đối lập, uyển ngữ: nhận ra chúng trước khi dùng chúng.'),
    ('nhip-dieu-cau-van',      'Câu dài câu ngắn xen nhau, và vì sao câu ngắn có sức nặng.'),
    ('tu-vung-tinh-te',        'Những từ thay đổi hẳn sắc thái của một câu.'),
    ('van-phong-bao-chi',      'Ngôn ngữ của toà soạn: kiểm chứng, khách quan, và những chỗ nó trượt.'),
    ('mo-dau-an-tuong',        'Ba mươi giây đầu quyết định người nghe có ở lại hay không.'),
    ('dung-lap-luan',          'Dựng lập luận đứng vững và nhận ra nguỵ biện trong lập luận người khác.'),
    ('keo-huong-nguoi-nghe',   'Giữ sự chú ý bằng nhịp và khoảng lặng, không bằng âm lượng.'),
    ('ket-bai-dong-lai',       'Kết sao cho người nghe mang được một thứ về nhà.')
) AS v(slug, summary)
WHERE lessons.slug = v.slug AND lessons.deleted_at IS NULL;

INSERT INTO lesson_blocks (lesson_id, kind, body, text_en, text_vi, speaker, position)
SELECT l.id, v.kind::lesson_block_kind, v.body, v.text_en, v.text_vi, v.speaker, v.position
FROM (VALUES
    -- B1
    ('mo-ta-nguoi','note','Tả người thì tính cách đáng nói hơn ngoại hình. Khi phải nhận xét tiêu cực, tiếng Anh thường nói giảm: thay vì "He is stubborn", người ta nói "He can be stubborn" — thêm can be là đã nhẹ đi nhiều.','','','',0),
    ('mo-ta-nguoi','example','','He is generous with his time.','Anh ấy rất hào phóng về thời gian.','',1),
    ('mo-ta-nguoi','example','','My brother can be stubborn.','Em trai tôi đôi khi rất bướng.','',2),
    ('mo-ta-nguoi','example','','She is ambitious about her career.','Cô ấy đầy tham vọng với sự nghiệp.','',3),

    ('mo-ta-noi-chon','note','Một nơi chốn hiện lên rõ nhất qua cảm giác nó gây ra, không qua kích thước. atmosphere, peaceful, bare nói được nhiều hơn big hay small.','','','',0),
    ('mo-ta-noi-chon','example','','The café has a warm atmosphere.','Quán cà phê có bầu không khí ấm cúng.','',1),
    ('mo-ta-noi-chon','example','','The walls were completely bare.','Các bức tường trống trơn.','',2),
    ('mo-ta-noi-chon','example','','A broad river runs through the city.','Một dòng sông rộng chảy qua thành phố.','',3),

    ('ke-mot-su-viec','note','Trạng từ nối cho câu chuyện một nhịp. suddenly bẻ ngoặt, meanwhile chuyển cảnh, eventually khép lại. Dùng đúng ba từ này là câu chuyện đã có hình.','','','',0),
    ('ke-mot-su-viec','example','','Suddenly the lights went out.','Đột nhiên đèn tắt hết.','',1),
    ('ke-mot-su-viec','example','','Meanwhile the others waited outside.','Trong khi đó những người khác chờ bên ngoài.','',2),
    ('ke-mot-su-viec','example','','Eventually we found the key.','Cuối cùng chúng tôi cũng tìm thấy chìa khoá.','',3),

    ('cam-xuc-va-thai-do','note','happy và sad quá rộng để nói được gì. annoyed khác angry ở cường độ, embarrassed khác ashamed ở chỗ một bên là ngượng nhất thời, một bên là xấu hổ về chính mình.','','','',0),
    ('cam-xuc-va-thai-do','example','','He looked annoyed by the noise.','Anh ấy có vẻ khó chịu vì tiếng ồn.','',1),
    ('cam-xuc-va-thai-do','example','','She felt embarrassed about the mistake.','Cô ấy thấy ngượng vì lỗi đó.','',2),
    ('cam-xuc-va-thai-do','example','','We were amazed by the view.','Chúng tôi kinh ngạc trước khung cảnh.','',3),

    ('email-trang-trong','note','Email công việc có bố cục cố định: lý do viết, nội dung, việc cần người nhận làm, lời kết. Đặt việc cần làm ở câu cuối thì dễ bị bỏ sót — nên đưa lên sớm.','','','',0),
    ('email-trang-trong','example','','I am writing regarding your invoice.','Tôi viết thư về việc hoá đơn của quý vị.','',1),
    ('email-trang-trong','example','','Please attach the signed form.','Vui lòng đính kèm biểu mẫu đã ký.','',2),
    ('email-trang-trong','example','','We will inform you by Friday.','Chúng tôi sẽ thông báo cho quý vị trước thứ Sáu.','',3),

    ('email-than-mat','note','Thư thân mật được phép lược chủ ngữ và dùng câu cụt: "Hope you are well", "Talk soon". Viết trang trọng cho bạn bè nghe xa cách hơn là lịch sự.','','','',0),
    ('email-than-mat','example','','Have you been busy lately?','Dạo này bạn có bận không?','',1),
    ('email-than-mat','example','','Send me an update when you can.','Khi nào rảnh gửi tôi tin cập nhật nhé.','',2),
    ('email-than-mat','example','','I suppose we could meet on Sunday.','Tôi đoán chừng chủ nhật mình gặp được.','',3),

    ('hen-lich-hop','note','Đề xuất hai ba khung giờ thay vì hỏi "Khi nào bạn rảnh?" — câu hỏi mở đẩy việc tra lịch sang cho người kia.','','','',0),
    ('hen-lich-hop','dialogue','','Are you available on Thursday morning?','Sáng thứ Năm bạn có rảnh không?','An',1),
    ('hen-lich-hop','dialogue','','Thursday is full. Could we arrange Friday instead?','Thứ Năm kín rồi. Mình sắp xếp thứ Sáu được không?','Linh',2),
    ('hen-lich-hop','example','','Please confirm the time.','Vui lòng xác nhận giờ hẹn.','',3),

    ('xin-loi-va-cam-on','note','Một lời xin lỗi đủ có ba phần: nhận việc gì đã sai, nói cách khắc phục, rồi dừng. Giải thích dài thành ra biện minh.','','','',0),
    ('xin-loi-va-cam-on','example','','Please forgive the short notice.','Mong bạn thứ lỗi vì báo gấp.','',1),
    ('xin-loi-va-cam-on','example','','I am writing on behalf of the team.','Tôi viết thư thay mặt cả nhóm.','',2),
    ('xin-loi-va-cam-on','example','','We want to show our appreciation.','Chúng tôi muốn bày tỏ sự trân trọng.','',3),

    -- B2
    ('neu-quan-diem','note','Khẳng định trần trụi mời người ta phản bác. Gắn mức độ chắc chắn vào câu — arguably, it seems, in my view — thì người nghe thấy chỗ để đối thoại thay vì phải chọn phe.','','','',0),
    ('neu-quan-diem','example','','That argument rests on one assumption.','Lập luận đó dựa trên một giả định.','',1),
    ('neu-quan-diem','example','','Timing is crucial here.','Thời điểm ở đây là then chốt.','',2),
    ('neu-quan-diem','example','','The wording is ambiguous.','Cách diễn đạt còn mơ hồ.','',3),

    ('phan-bien','note','Phản biện nhắm vào lập luận, không nhắm vào người. Công thức an toàn: thừa nhận phần đúng trước, rồi mới nêu chỗ hổng.','','','',0),
    ('phan-bien','example','','I concede that the timing was poor.','Tôi thừa nhận là thời điểm không tốt.','',1),
    ('phan-bien','example','','These figures contradict the report.','Những con số này mâu thuẫn với báo cáo.','',2),
    ('phan-bien','example','','The study seems to overlook cost.','Nghiên cứu có vẻ bỏ sót yếu tố chi phí.','',3),

    ('dan-chung','note','Một câu chuyện không phải là bằng chứng. Khi dẫn số liệu, nói luôn nguồn và cỡ mẫu — thiếu hai thứ đó thì con số nào nghe cũng thuyết phục.','','','',0),
    ('dan-chung','example','','These statistics come from the census.','Số liệu thống kê này lấy từ cuộc tổng điều tra.','',1),
    ('dan-chung','example','','The results are consistent across groups.','Kết quả nhất quán giữa các nhóm.','',2),
    ('dan-chung','example','','They attribute the rise to policy.','Họ quy mức tăng đó cho chính sách.','',3),

    ('chot-van-de','note','Chốt lại không phải là nhắc lại. Nêu điều đã thống nhất, điều còn mở, và việc tiếp theo của từng người.','','','',0),
    ('chot-van-de','example','','Ultimately the decision is yours.','Rốt cuộc quyết định là của bạn.','',1),
    ('chot-van-de','example','','Overall the project went well.','Nhìn chung dự án diễn ra tốt.','',2),
    ('chot-van-de','example','','We proceeded with their consent.','Chúng tôi tiến hành với sự đồng ý của họ.','',3),

    ('hop-hanh','note','Xin nói trong cuộc họp bằng một câu ngắn báo trước ý mình: "Can I add something on that?" Nhảy thẳng vào nội dung dễ bị coi là cắt lời.','','','',0),
    ('hop-hanh','dialogue','','Could you clarify the deadline?','Bạn làm rõ giúp hạn chót được không?','Linh',1),
    ('hop-hanh','dialogue','','End of the month. Can we defer the budget item?','Cuối tháng. Mục ngân sách mình hoãn lại được không?','An',2),
    ('hop-hanh','example','','We should consult the legal team.','Chúng ta nên tham vấn bộ phận pháp lý.','',3),

    ('bao-cao-tien-do','note','Tin xấu nói sớm và nói kèm phương án thì vẫn giữ được lòng tin. Nói muộn thì bản thân sự chậm trễ trở thành vấn đề lớn hơn cái chậm tiến độ.','','','',0),
    ('bao-cao-tien-do','example','','We did not anticipate this delay.','Chúng tôi không dự liệu được sự chậm trễ này.','',1),
    ('bao-cao-tien-do','example','','The complexity grew month by month.','Độ phức tạp tăng lên theo từng tháng.','',2),
    ('bao-cao-tien-do','example','','Attainment fell short of the target.','Mức đạt được thấp hơn mục tiêu.','',3),

    ('thuong-luong','note','Nhượng bộ nên có điều kiện đi kèm: "We could do that if..." Nhượng bộ trắng chỉ dạy đối phương rằng cứ đòi là được.','','','',0),
    ('thuong-luong','example','','We will compensate you for the delay.','Chúng tôi sẽ bù đắp cho bạn vì sự chậm trễ.','',1),
    ('thuong-luong','example','','Alternatively, we could split the cost.','Hoặc là chúng ta chia đôi chi phí.','',2),
    ('thuong-luong','example','','The price dropped considerably.','Giá đã giảm đáng kể.','',3),

    ('phan-hoi-dong-nghiep','note','Góp ý dùng được phải nói rõ ba điều: việc gì, ảnh hưởng ra sao, và lần sau làm khác thế nào. Thiếu vế cuối thì đó chỉ là lời phàn nàn.','','','',0),
    ('phan-hoi-dong-nghiep','example','','Read the draft critically.','Hãy đọc bản nháp một cách phê phán.','',1),
    ('phan-hoi-dong-nghiep','example','','We received contradictory feedback.','Chúng tôi nhận phản hồi mâu thuẫn nhau.','',2),
    ('phan-hoi-dong-nghiep','example','','I am not convinced by that reason.','Tôi chưa tin chắc vào lý do đó.','',3)
) AS v(lesson_slug, kind, body, text_en, text_vi, speaker, position)
JOIN lessons AS l ON l.slug = v.lesson_slug AND l.deleted_at IS NULL;

INSERT INTO lesson_blocks (lesson_id, kind, body, text_en, text_vi, speaker, position)
SELECT l.id, v.kind::lesson_block_kind, v.body, v.text_en, v.text_vi, v.speaker, v.position
FROM (VALUES
    -- C1
    ('trich-dan-nguon','note','Diễn đạt lại không phải là đổi vài từ đồng nghĩa. Phải đổi cả cấu trúc câu, và vẫn phải dẫn nguồn — giữ nguyên ý người khác mà không ghi nguồn vẫn là đạo văn.','','','',0),
    ('trich-dan-nguon','example','','Every claim needs a citation.','Mỗi khẳng định đều cần trích dẫn nguồn.','',1),
    ('trich-dan-nguon','example','','A single error damages credibility.','Một lỗi duy nhất cũng làm hỏng độ tin cậy.','',2),
    ('trich-dan-nguon','example','','That assertion needs support.','Luận điểm đó cần được chống đỡ.','',3),

    ('mo-ta-du-lieu','note','Mô tả dữ liệu cần đúng mức. plummet là rơi thẳng đứng, không dùng cho mức giảm 3%. negligible nghĩa là nhỏ tới mức bỏ qua được, không phải chỉ là nhỏ.','','','',0),
    ('mo-ta-du-lieu','example','','The difference was negligible.','Khác biệt là không đáng kể.','',1),
    ('mo-ta-du-lieu','example','','Only a fraction of users replied.','Chỉ một phần nhỏ người dùng trả lời.','',2),
    ('mo-ta-du-lieu','example','','The instrument lacks precision.','Thiết bị thiếu độ chính xác.','',3),

    ('gioi-han-nghien-cuu','note','Tự nêu giới hạn của nghiên cứu mình không làm nó yếu đi — nó cho thấy tác giả biết mình đang nói gì. Bài báo nào cũng có mục này, và người phản biện đọc nó đầu tiên.','','','',0),
    ('gioi-han-nghien-cuu','example','','Time was the main constraint.','Thời gian là ràng buộc chính.','',1),
    ('gioi-han-nghien-cuu','example','','That conclusion is speculative.','Kết luận đó mang tính suy đoán.','',2),
    ('gioi-han-nghien-cuu','example','','The method has one clear drawback.','Phương pháp có một nhược điểm rõ.','',3),

    ('thuat-ngu-hoc-thuat','note','empirical nghĩa là dựa trên quan sát hoặc thí nghiệm, không phải chỉ là "có số liệu". Dùng sai từ này là dấu hiệu rõ nhất của một bài viết chưa chắc tay.','','','',0),
    ('thuat-ngu-hoc-thuat','example','','The criteria are listed in section two.','Các tiêu chí được nêu ở mục hai.','',1),
    ('thuat-ngu-hoc-thuat','example','','There is an inherent bias here.','Ở đây có một thiên lệch vốn có.','',2),
    ('thuat-ngu-hoc-thuat','example','','Reviewers scrutinise every figure.','Người phản biện soi xét kỹ từng con số.','',3),

    ('thanh-ngu-thong-dung','note','Những tính từ này khác nhau ở cường độ, không phải ở nghĩa. tedious là chán, daunting là nản trước khi bắt đầu, gruelling là vắt kiệt sức khi đã làm.','','','',0),
    ('thanh-ngu-thong-dung','example','','The process is tedious but necessary.','Quy trình tẻ nhạt nhưng cần thiết.','',1),
    ('thanh-ngu-thong-dung','example','','The workload looked daunting.','Khối lượng việc trông thật nản lòng.','',2),
    ('thanh-ngu-thong-dung','example','','Her talk was riveting.','Bài nói của cô ấy cuốn không rời.','',3),

    ('cum-dong-tu','note','Văn bản trang trọng thường thay cụm động từ bằng một động từ gốc Latin: speed up thành expedite, hold back thành impede, give up thành relinquish. Nghĩa gần nhau, sắc thái khác hẳn.','','','',0),
    ('cum-dong-tu','example','','We can expedite the paperwork.','Chúng tôi có thể đẩy nhanh thủ tục.','',1),
    ('cum-dong-tu','example','','Poor data will impede the analysis.','Dữ liệu kém sẽ cản trở việc phân tích.','',2),
    ('cum-dong-tu','example','','We envisage a longer timeline.','Chúng tôi hình dung một tiến độ dài hơn.','',3),

    ('noi-giam-noi-tranh','note','Tiếng Anh trang trọng nói nhẹ đi theo phản xạ. "The improvement was marginal" thường có nghĩa là gần như không cải thiện gì. Người học cần nghe ra lớp nghĩa đó, không chỉ dịch mặt chữ.','','','',0),
    ('noi-giam-noi-tranh','example','','The improvement was marginal.','Mức cải thiện không đáng bao nhiêu.','',1),
    ('noi-giam-noi-tranh','example','','The review felt superficial.','Bản rà soát có vẻ hời hợt.','',2),
    ('noi-giam-noi-tranh','example','','His objection was rather timid.','Lời phản đối của anh ấy khá rụt rè.','',3),

    ('sac-thai-trang-trong','note','thereby, hence, whereby thuộc văn viết chính thức. Đưa chúng vào hội thoại thường ngày nghe gượng, giống như nói tiếng Việt kiểu công văn.','','','',0),
    ('sac-thai-trang-trong','example','','The data is thin; hence the caution.','Dữ liệu mỏng, do đó phải dè dặt.','',1),
    ('sac-thai-trang-trong','example','','It is customary to reply in writing.','Theo thông lệ thì trả lời bằng văn bản.','',2),
    ('sac-thai-trang-trong','example','','We are obliged to report this.','Chúng tôi buộc phải báo cáo việc này.','',3),

    -- C2
    ('bien-phap-tu-tu','note','Nhận ra biện pháp tu từ quan trọng hơn dùng nó. Uyển ngữ trong bản tin, ẩn dụ trong diễn văn chính trị — chúng định hướng người đọc trước khi người đọc kịp nghĩ.','','','',0),
    ('bien-phap-tu-tu','example','','The title is an allusion to a poem.','Nhan đề là một sự ám chỉ tới một bài thơ.','',1),
    ('bien-phap-tu-tu','example','','The word carries a negative connotation.','Từ đó mang hàm ý tiêu cực.','',2),
    ('bien-phap-tu-tu','example','','He meant it figuratively.','Anh ấy nói theo nghĩa bóng.','',3),

    ('nhip-dieu-cau-van','note','Một câu ngắn sau vài câu dài có sức nặng mà bản thân nó không có. Đó là hiệu ứng của tương phản, không phải của độ ngắn.','','','',0),
    ('nhip-dieu-cau-van','example','','Brevity is the strength of this piece.','Sự ngắn gọn là thế mạnh của bài này.','',1),
    ('nhip-dieu-cau-van','example','','The opening chapter is ponderous.','Chương mở đầu khá nặng nề.','',2),
    ('nhip-dieu-cau-van','example','','The dialogue is deliberately colloquial.','Đoạn thoại cố ý dùng khẩu ngữ.','',3),

    ('tu-vung-tinh-te','note','ostensibly báo hiệu người viết không tin vào lý do vừa nêu. Một từ đổi hẳn thái độ của cả câu — đó là lý do phải đọc kỹ những từ như thế.','','','',0),
    ('tu-vung-tinh-te','example','','The trip was ostensibly about research.','Chuyến đi bề ngoài là để nghiên cứu.','',1),
    ('tu-vung-tinh-te','example','','She made an oblique reference to it.','Cô ấy nhắc tới nó một cách gián tiếp.','',2),
    ('tu-vung-tinh-te','example','','There was a tacit agreement.','Đã có một thoả thuận ngầm.','',3),

    ('van-phong-bao-chi','note','Báo chí nghiêm túc kiểm chứng bằng nhiều nguồn độc lập. Khi một bài chỉ dẫn "sources say" mà không nói gì thêm, đó là chỗ cần dè chừng.','','','',0),
    ('van-phong-bao-chi','example','','Two papers will denounce the ruling.','Hai tờ báo sẽ lên án phán quyết.','',1),
    ('van-phong-bao-chi','example','','The quote turned out to be spurious.','Câu trích hoá ra là nguỵ tạo.','',2),
    ('van-phong-bao-chi','example','','Her sourcing is impeccable.','Cách dẫn nguồn của cô ấy không chê vào đâu được.','',3),

    ('mo-dau-an-tuong','note','Mở đầu không cần gây sốc, cần gây tò mò. Một câu hỏi mà người nghe chưa có câu trả lời sẽ giữ họ lâu hơn một con số giật gân.','','','',0),
    ('mo-dau-an-tuong','example','','State your proposition in one sentence.','Hãy nêu luận điểm trong một câu.','',1),
    ('mo-dau-an-tuong','example','','Clarity is paramount in the first minute.','Sự rõ ràng là tối quan trọng trong phút đầu.','',2),
    ('mo-dau-an-tuong','example','','Let us posit two possible causes.','Hãy đặt ra hai nguyên nhân có thể.','',3),

    ('dung-lap-luan','note','Nguỵ biện phổ biến nhất trong tranh luận đời thường là dựng người rơm: phản bác một phiên bản méo mó của ý đối phương. Nhận ra nó là bước đầu để không mắc phải.','','','',0),
    ('dung-lap-luan','example','','That is a common logical fallacy.','Đó là một nguỵ biện logic thường gặp.','',1),
    ('dung-lap-luan','example','','She made a cogent case for delay.','Cô ấy lập luận chặt chẽ cho việc hoãn lại.','',2),
    ('dung-lap-luan','example','','You must substantiate that figure.','Bạn phải chứng minh con số đó.','',3),

    ('keo-huong-nguoi-nghe','note','Khoảng lặng làm được điều mà âm lượng không làm được: nó buộc người nghe tự lấp vào chỗ trống, và thứ họ tự nghĩ ra thì họ nhớ lâu hơn.','','','',0),
    ('keo-huong-nguoi-nghe','example','','A pause can induce real attention.','Một khoảng lặng có thể làm nảy sinh sự chú ý thật.','',1),
    ('keo-huong-nguoi-nghe','example','','He ended on a rhetorical question.','Anh ấy kết bằng một câu hỏi tu từ.','',2),
    ('keo-huong-nguoi-nghe','example','','She argued compellingly.','Cô ấy lập luận một cách cuốn hút.','',3),

    ('ket-bai-dong-lai','note','Kết bài nên để lại một thứ, không phải tóm tắt mọi thứ. Người nghe mang về được một câu hoặc một hình ảnh là đủ.','','','',0),
    ('ket-bai-dong-lai','example','','The ending tries to reconcile both views.','Đoạn kết cố dung hoà cả hai quan điểm.','',1),
    ('ket-bai-dong-lai','example','','The final line is wistful.','Câu cuối mang vẻ bâng khuâng.','',2),
    ('ket-bai-dong-lai','example','','The last impression is paramount.','Ấn tượng cuối cùng là tối quan trọng.','',3)
) AS v(lesson_slug, kind, body, text_en, text_vi, speaker, position)
JOIN lessons AS l ON l.slug = v.lesson_slug AND l.deleted_at IS NULL;
