-- Nội dung bài học: phần dẫn và các khối giải thích, câu mẫu, hội thoại.
--
-- Do Claude soạn. Bậc CEFR của từ vựng thì đối chiếu được bằng máy
-- (`make verify-vocab`); nội dung dạy thì KHÔNG có nguồn nào để đối chiếu —
-- giải thích ngữ pháp và dựng hội thoại là việc sư phạm. Cần người có chuyên
-- môn đọc lại trước khi có người học thật.
--
-- Câu mẫu (example) minh hoạ điểm của ghi chú, KHÔNG lặp lại câu ví dụ của từ
-- vựng: từ vựng đã có bước riêng trên màn học, và câu mẫu trùng nguyên văn sẽ
-- bị màn học bỏ qua. Bản đầu tiên lấy thẳng câu ví dụ của từ (122/132 câu).
--
-- Thay toàn bộ nội dung của những bài nó phụ trách, nên chạy lại là hội tụ.
-- Hội thoại của phần lớn các bài nằm ở 006, chạy sau file này.

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
    ('thi-hien-tai-don','example','','My brother works in a bank.','Anh trai tôi làm ở ngân hàng.','',3),
    ('thi-hien-tai-don','example','','She is always happy in the morning.','Buổi sáng cô ấy lúc nào cũng vui.','',4),
    ('thi-hien-tai-don','example','','They are often late.','Họ thường đến muộn.','',5),

    -- A1 · Danh từ đếm được và không đếm được
    ('danh-tu-dem-duoc','note','Danh từ đếm được có số nhiều: one bottle, two bottles. Danh từ không đếm được thì không: sugar, bread, rice, water. Chúng luôn đi với động từ số ít.','','','',0),
    ('danh-tu-dem-duoc','note','Muốn đếm một danh từ không đếm được, phải mượn một đơn vị: a glass of water, a piece of bread, a bottle of juice.','','','',1),
    ('danh-tu-dem-duoc','example','','We bought three bottles of juice.','Chúng tôi đã mua ba chai nước ép.','',2),
    ('danh-tu-dem-duoc','example','','The rice is ready.','Cơm chín rồi.','',3),
    ('danh-tu-dem-duoc','example','','Can I have a glass of water, please?','Cho tôi xin một cốc nước nhé?','',4),

    -- A1 · Mạo từ
    ('mao-tu-a-an-the','note','a hay an không chọn theo chữ cái mà theo ÂM đầu. an hour vì h câm nên nghe ra nguyên âm; a university vì u ở đây đọc là "yu", một phụ âm.','','','',0),
    ('mao-tu-a-an-the','note','Dùng a/an khi lần đầu nhắc tới, dùng the khi cả hai người đều đã biết đang nói vật nào.','','','',1),
    ('mao-tu-a-an-the','example','','She is an honest person.','Cô ấy là một người trung thực.','',2),
    ('mao-tu-a-an-the','example','','He studies at a university in Hue.','Anh ấy học ở một trường đại học ở Huế.','',3),
    ('mao-tu-a-an-the','example','','I saw a cat. The cat was black.','Tôi thấy một con mèo. Con mèo đó màu đen.','',4),

    -- A1 · Động từ to be
    ('dong-tu-to-be','note','to be có ba hình ở hiện tại: I am, he/she/it is, you/we/they are. Nó nối chủ ngữ với một tính từ hoặc một danh từ, chứ không mô tả hành động.','','','',0),
    ('dong-tu-to-be','note','Tiếng Việt nói "Tôi đói" không cần động từ, nên người học hay bỏ quên am/is/are. Trong tiếng Anh nó bắt buộc.','','','',1),
    ('dong-tu-to-be','example','','I am a nurse.','Tôi là y tá.','',2),
    ('dong-tu-to-be','example','','My parents are at home.','Bố mẹ tôi đang ở nhà.','',3),
    ('dong-tu-to-be','example','','It is cold today.','Hôm nay trời lạnh.','',4),

    -- A1 · Gia đình và bạn bè
    ('gia-dinh-va-ban-be','note','Tiếng Anh gộp nhiều vai mà tiếng Việt tách riêng: uncle là cả chú, bác lẫn cậu; cousin là mọi anh chị em họ, không phân biệt tuổi hay bên nội ngoại.','','','',0),
    ('gia-dinh-va-ban-be','example','','My uncle is my mother''s younger brother.','Cậu tôi là em trai của mẹ tôi.','',1),
    ('gia-dinh-va-ban-be','example','','My aunt lives with my grandparents.','Cô tôi sống với ông bà tôi.','',2),
    ('gia-dinh-va-ban-be','example','','I have six cousins on my father''s side.','Tôi có sáu anh chị em họ bên nội.','',3),
    ('gia-dinh-va-ban-be','dialogue','','Do you have any brothers or sisters?','Bạn có anh chị em không?','Mai',4),
    ('gia-dinh-va-ban-be','dialogue','','One sister. She is a teacher.','Một chị gái. Chị ấy là giáo viên.','Tom',5),

    -- A1 · Màu sắc và hình dáng
    ('mau-sac-va-hinh-dang','note','Tính từ trong tiếng Anh đứng TRƯỚC danh từ và không đổi theo số nhiều: a bright dress, two bright dresses.','','','',0),
    ('mau-sac-va-hinh-dang','note','Khi có nhiều tính từ, thứ tự quen thuộc là: ý kiến → kích thước → màu sắc. "a nice long dark coat" nghe tự nhiên, đảo lại thì không.','','','',1),
    ('mau-sac-va-hinh-dang','example','','She bought two red dresses.','Cô ấy mua hai chiếc váy đỏ.','',2),
    ('mau-sac-va-hinh-dang','example','','He has a big black dog.','Anh ấy có một con chó đen to.','',3),
    ('mau-sac-va-hinh-dang','example','','What a lovely little house!','Ngôi nhà nhỏ xinh quá!','',4),

    -- A1 · Đồ ăn và thức uống
    ('do-an-va-thuc-uong','note','Nói thích hay không thích: I like, I do not like, I love. Sau chúng dùng danh từ số nhiều hoặc danh từ không đếm được: I like vegetables, I like juice.','','','',0),
    ('do-an-va-thuc-uong','example','','I like vegetables.','Tôi thích rau.','',1),
    ('do-an-va-thuc-uong','example','','She loves coffee.','Cô ấy rất thích cà phê.','',2),
    ('do-an-va-thuc-uong','example','','We do not like spicy food.','Chúng tôi không thích đồ cay.','',3),
    ('do-an-va-thuc-uong','dialogue','','What do you want for dinner?','Bữa tối bạn muốn ăn gì?','Mai',4),
    ('do-an-va-thuc-uong','dialogue','','Something light. Maybe fruit and bread.','Gì đó nhẹ thôi. Có thể là trái cây với bánh mì.','Tom',5),

    -- A1 · Thời gian và ngày tháng
    ('thoi-gian-va-ngay-thang','note','Ba giới từ hay nhầm: at cho giờ (at seven), on cho ngày (on Monday), in cho tháng và năm (in March, in 2026).','','','',0),
    ('thoi-gian-va-ngay-thang','note','in the morning, in the evening — nhưng at night. Đây là ngoại lệ phải nhớ riêng.','','','',1),
    ('thoi-gian-va-ngay-thang','example','','The shop opens at nine.','Cửa hàng mở cửa lúc chín giờ.','',2),
    ('thoi-gian-va-ngay-thang','example','','My birthday is on the fifth of May.','Sinh nhật tôi vào ngày 5 tháng 5.','',3),
    ('thoi-gian-va-ngay-thang','example','','She was born in 2001.','Cô ấy sinh năm 2001.','',4),
    ('thoi-gian-va-ngay-thang','example','','We never go out at night.','Chúng tôi không bao giờ ra ngoài vào ban đêm.','',5)
) AS v(lesson_slug, kind, body, text_en, text_vi, speaker, position)
JOIN lessons AS l ON l.slug = v.lesson_slug AND l.deleted_at IS NULL;

INSERT INTO lesson_blocks (lesson_id, kind, body, text_en, text_vi, speaker, position)
SELECT l.id, v.kind::lesson_block_kind, v.body, v.text_en, v.text_vi, v.speaker, v.position
FROM (VALUES
    -- A2 · Chào hỏi và làm quen
    ('chao-hoi-lam-quen','note','Câu chào trong tiếng Anh thường đi kèm một câu hỏi mở, để cuộc trò chuyện có chỗ đi tiếp. Chỉ nói "Hello" rồi dừng là một câu chào chưa xong.','','','',0),
    ('chao-hoi-lam-quen','note','"How are you?" phần lớn là câu xã giao, không phải câu hỏi thật về sức khoẻ. Trả lời ngắn rồi hỏi lại là đủ.','','','',1),
    ('chao-hoi-lam-quen','example','','Hi, I am Lan. Are you new here?','Chào, mình là Lan. Bạn mới đến à?','',2),
    ('chao-hoi-lam-quen','example','','Fine, thanks. And you?','Mình khoẻ, cảm ơn. Còn bạn?','',3),
    ('chao-hoi-lam-quen','example','','Could you spell your name for me?','Bạn đánh vần tên mình giúp được không?','',4),
    ('chao-hoi-lam-quen','dialogue','','Hi, I do not think we have met. I am Mai.','Chào bạn, hình như mình chưa gặp nhau. Mình là Mai.','Mai',5),
    ('chao-hoi-lam-quen','dialogue','','Nice to meet you, Mai. I am Tom. What do you do?','Rất vui được gặp Mai. Mình là Tom. Bạn làm nghề gì?','Tom',6),
    ('chao-hoi-lam-quen','dialogue','','I work at a school near here. And you?','Mình làm ở một trường gần đây. Còn bạn?','Mai',7),

    -- A2 · Gọi món ở nhà hàng
    ('goi-mon-o-nha-hang','note','Gọi món lịch sự thường bắt đầu bằng "Could I have..." hoặc "I will have...". Nói thẳng "I want" nghe cộc, dù không sai ngữ pháp.','','','',0),
    ('goi-mon-o-nha-hang','example','','Could I have the chicken soup, please?','Cho tôi món súp gà nhé.','',1),
    ('goi-mon-o-nha-hang','example','','I will have a glass of orange juice.','Tôi lấy một cốc nước cam.','',2),
    ('goi-mon-o-nha-hang','example','','I am allergic to peanuts.','Tôi bị dị ứng đậu phộng.','',3),
    ('goi-mon-o-nha-hang','dialogue','','Are you ready to order?','Anh chị gọi món chưa ạ?','Phục vụ',4),
    ('goi-mon-o-nha-hang','dialogue','','Almost. What would you recommend?','Gần rồi. Bạn gợi ý món gì?','Khách',5),
    ('goi-mon-o-nha-hang','dialogue','','The chef recommends the fish today.','Hôm nay đầu bếp gợi ý món cá ạ.','Phục vụ',6),

    -- A2 · Hỏi đường
    ('hoi-duong-di','note','Hỏi đường thì dễ, hiểu câu trả lời mới khó. Nghe kỹ ba thứ: hướng rẽ, số lần rẽ, và mốc để nhận ra chỗ cần dừng.','','','',0),
    ('hoi-duong-di','note','Nếu không kịp hiểu, câu cứu nguy là "Sorry, could you say that again more slowly?" — dùng nó còn hơn gật đầu rồi đi lạc.','','','',1),
    ('hoi-duong-di','example','','Take the second road on the right.','Rẽ vào con đường thứ hai bên phải.','',2),
    ('hoi-duong-di','example','','Sorry, could you say that again more slowly?','Xin lỗi, bạn nói lại chậm hơn được không?','',3),
    ('hoi-duong-di','example','','I think we are going in the wrong direction.','Tôi nghĩ chúng ta đang đi sai hướng.','',4),
    ('hoi-duong-di','dialogue','','Excuse me, how do I get to the station?','Xin lỗi, đi tới nhà ga thế nào ạ?','Khách',5),
    ('hoi-duong-di','dialogue','','Go straight, then turn left. It is opposite the market.','Đi thẳng rồi rẽ trái. Nó đối diện chợ.','Người dân',6),

    -- A2 · Mua sắm
    ('mua-sam','note','Ba câu đủ dùng cho hầu hết tình huống mua sắm: hỏi giá, hỏi cỡ khác, và hỏi chính sách đổi trả.','','','',0),
    ('mua-sam','example','','How much is this jacket?','Chiếc áo khoác này bao nhiêu tiền?','',1),
    ('mua-sam','example','','Is there a medium in blue?','Có cỡ M màu xanh không?','',2),
    ('mua-sam','example','','Can I return it if it does not fit?','Nếu không vừa thì tôi trả lại được không?','',3),
    ('mua-sam','dialogue','','Do you have this in a larger size?','Cái này có cỡ lớn hơn không ạ?','Khách',4),
    ('mua-sam','dialogue','','Let me check. Would you like to pay in cash or by card?','Để tôi xem. Anh chị trả tiền mặt hay thẻ ạ?','Nhân viên',5),

    -- A2 · Thì quá khứ đơn
    ('thi-qua-khu-don','note','Động từ có quy tắc thêm -ed. Đuôi đó có ba cách đọc: /t/ sau âm vô thanh (worked), /d/ sau âm hữu thanh (played), và /ɪd/ sau t hoặc d (decided).','','','',0),
    ('thi-qua-khu-don','note','Ở câu phủ định và câu hỏi, did đã mang nghĩa quá khứ nên động từ chính quay về nguyên thể: "He did not admit it", không phải "did not admitted".','','','',1),
    ('thi-qua-khu-don','example','','They decided to stay one more day.','Họ quyết định ở thêm một ngày.','',2),
    ('thi-qua-khu-don','example','','She complained about the noise.','Cô ấy phàn nàn về tiếng ồn.','',3),
    ('thi-qua-khu-don','example','','He did not admit his mistake.','Anh ấy không chịu thừa nhận lỗi của mình.','',4),

    -- A2 · Động từ bất quy tắc
    ('dong-tu-bat-quy-tac','note','Học động từ bất quy tắc theo NHÓM biến đổi dễ hơn học rời: buy–bought, bring–brought, teach–taught đều đổi thành -ought/-aught.','','','',0),
    ('dong-tu-bat-quy-tac','note','Một nhóm khác giữ nguyên ở cả ba dạng: cut–cut–cut, put–put–put. Nhóm này không phải học gì thêm.','','','',1),
    ('dong-tu-bat-quy-tac','example','','I bought some bread and brought it home.','Tôi mua ít bánh mì rồi mang về nhà.','',2),
    ('dong-tu-bat-quy-tac','example','','She taught English for ten years.','Cô ấy đã dạy tiếng Anh mười năm.','',3),
    ('dong-tu-bat-quy-tac','example','','He cut his finger yesterday.','Hôm qua anh ấy bị đứt tay.','',4),

    -- A2 · Nói về dự định
    ('noi-ve-du-dinh','note','Ba cách nói tương lai, khác nhau ở mức chắc chắn: will cho quyết định ngay lúc nói, be going to cho dự định đã có sẵn, hiện tại tiếp diễn cho lịch đã chốt.','','','',0),
    ('noi-ve-du-dinh','note','"We are moving in June" nghe chắc hơn "We will move in June", vì nó hàm ý đã sắp xếp xong.','','','',1),
    ('noi-ve-du-dinh','example','','It is raining. I will take a taxi.','Trời mưa rồi. Mình sẽ đi taxi.','',2),
    ('noi-ve-du-dinh','example','','We are going to paint the kitchen this summer.','Hè này chúng tôi định sơn lại bếp.','',3),
    ('noi-ve-du-dinh','example','','I am meeting the doctor at four.','Bốn giờ tôi có hẹn với bác sĩ.','',4),
    ('noi-ve-du-dinh','example','','They intend to open a shop.','Họ dự định mở một cửa hàng.','',5),

    -- A2 · Kể về kỳ nghỉ
    ('ke-ve-ky-nghi','note','Kể chuyện thì đừng liệt kê. Chọn một chi tiết đáng nhớ rồi dựng câu quanh nó — người nghe nhớ chi tiết, không nhớ lịch trình.','','','',0),
    ('ke-ve-ky-nghi','example','','The best part was a tiny café by the sea.','Điều hay nhất là một quán cà phê nhỏ xíu bên bờ biển.','',1),
    ('ke-ve-ky-nghi','example','','It rained every day, but we did not mind.','Ngày nào cũng mưa, nhưng chúng tôi chẳng bận tâm.','',2),
    ('ke-ve-ky-nghi','example','','An old man taught us a local song.','Một ông cụ dạy chúng tôi một bài hát địa phương.','',3),
    ('ke-ve-ky-nghi','dialogue','','How was your trip?','Chuyến đi của bạn thế nào?','Mai',4),
    ('ke-ve-ky-nghi','dialogue','','Long but good. We drove along the coast for two days.','Dài nhưng vui. Bọn mình lái xe dọc bờ biển hai ngày.','Tom',5),
    ('ke-ve-ky-nghi','dialogue','','Was it crowded?','Có đông không?','Mai',6),
    ('ke-ve-ky-nghi','dialogue','','Not at all. That was the best part.','Không hề. Đó mới là điểm hay nhất.','Tom',7)
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
    ('mo-ta-nguoi','example','','She is patient with new staff.','Chị ấy kiên nhẫn với nhân viên mới.','',1),
    ('mo-ta-nguoi','example','','He can be a little impatient in the morning.','Buổi sáng anh ấy có thể hơi thiếu kiên nhẫn.','',2),
    ('mo-ta-nguoi','example','','My grandmother is kind and very funny.','Bà tôi hiền và rất vui tính.','',3),

    ('mo-ta-noi-chon','note','Một nơi chốn hiện lên rõ nhất qua cảm giác nó gây ra, không qua kích thước. atmosphere, peaceful, bare nói được nhiều hơn big hay small.','','','',0),
    ('mo-ta-noi-chon','example','','The old library feels calm and a little mysterious.','Thư viện cũ mang cảm giác yên tĩnh và hơi bí ẩn.','',1),
    ('mo-ta-noi-chon','example','','The market is noisy, colourful and full of life.','Khu chợ ồn ào, đầy màu sắc và tràn sức sống.','',2),
    ('mo-ta-noi-chon','example','','The room was cold and smelled of rain.','Căn phòng lạnh lẽo và thoang thoảng mùi mưa.','',3),

    ('ke-mot-su-viec','note','Trạng từ nối cho câu chuyện một nhịp. suddenly bẻ ngoặt, meanwhile chuyển cảnh, eventually khép lại. Dùng đúng ba từ này là câu chuyện đã có hình.','','','',0),
    ('ke-mot-su-viec','example','','We were having dinner when suddenly the phone rang.','Chúng tôi đang ăn tối thì đột nhiên điện thoại reo.','',1),
    ('ke-mot-su-viec','example','','Meanwhile, back at the hotel, nobody knew where we were.','Trong khi đó, ở khách sạn, không ai biết chúng tôi ở đâu.','',2),
    ('ke-mot-su-viec','example','','Eventually, the bus arrived two hours late.','Cuối cùng xe buýt cũng đến, trễ hai tiếng.','',3),

    ('cam-xuc-va-thai-do','note','happy và sad quá rộng để nói được gì. annoyed khác angry ở cường độ, embarrassed khác ashamed ở chỗ một bên là ngượng nhất thời, một bên là xấu hổ về chính mình.','','','',0),
    ('cam-xuc-va-thai-do','example','','I was annoyed, not angry.','Tôi chỉ bực mình, chứ không giận.','',1),
    ('cam-xuc-va-thai-do','example','','He was embarrassed when he fell off his chair.','Anh ấy ngượng khi bị ngã khỏi ghế.','',2),
    ('cam-xuc-va-thai-do','example','','She felt ashamed of lying to her friend.','Cô ấy thấy xấu hổ vì đã nói dối bạn.','',3),

    ('email-trang-trong','note','Email công việc có bố cục cố định: lý do viết, nội dung, việc cần người nhận làm, lời kết. Đặt việc cần làm ở câu cuối thì dễ bị bỏ sót — nên đưa lên sớm.','','','',0),
    ('email-trang-trong','example','','I am writing to ask about your summer courses.','Tôi viết thư này để hỏi về các khoá học mùa hè của quý vị.','',1),
    ('email-trang-trong','example','','Could you send me the timetable by Monday?','Anh chị có thể gửi tôi thời khoá biểu trước thứ Hai được không?','',2),
    ('email-trang-trong','example','','I look forward to hearing from you.','Tôi mong sớm nhận được phản hồi.','',3),

    ('email-than-mat','note','Thư thân mật được phép lược chủ ngữ và dùng câu cụt: "Hope you are well", "Talk soon". Viết trang trọng cho bạn bè nghe xa cách hơn là lịch sự.','','','',0),
    ('email-than-mat','example','','Hope you are well.','Mong là cậu vẫn khoẻ.','',1),
    ('email-than-mat','example','','Thanks for the photos. Loved them!','Cảm ơn mấy tấm ảnh nhé. Thích lắm!','',2),
    ('email-than-mat','example','','Talk soon.','Nói chuyện sau nhé.','',3),

    ('hen-lich-hop','note','Đề xuất hai ba khung giờ thay vì hỏi "Khi nào bạn rảnh?" — câu hỏi mở đẩy việc tra lịch sang cho người kia.','','','',0),
    ('hen-lich-hop','example','','Would Tuesday at ten or Wednesday at two work for you?','Mười giờ thứ Ba hoặc hai giờ thứ Tư có tiện cho anh chị không?','',1),
    ('hen-lich-hop','example','','I am free any time after three on Friday.','Chiều thứ Sáu, sau ba giờ lúc nào tôi cũng rảnh.','',2),
    ('hen-lich-hop','example','','Let me know which time suits you best.','Anh chị cho tôi biết giờ nào tiện nhất nhé.','',3),
    ('hen-lich-hop','dialogue','','Are you available on Thursday morning?','Sáng thứ Năm bạn có rảnh không?','An',4),
    ('hen-lich-hop','dialogue','','Thursday is full. Could we arrange Friday instead?','Thứ Năm kín rồi. Mình sắp xếp thứ Sáu được không?','Linh',5),

    ('xin-loi-va-cam-on','note','Một lời xin lỗi đủ có ba phần: nhận việc gì đã sai, nói cách khắc phục, rồi dừng. Giải thích dài thành ra biện minh.','','','',0),
    ('xin-loi-va-cam-on','example','','I am sorry I missed your call. I will ring you back at five.','Xin lỗi vì mình lỡ cuộc gọi của bạn. Năm giờ mình gọi lại nhé.','',1),
    ('xin-loi-va-cam-on','example','','I sent the wrong file. The correct one is attached.','Tôi đã gửi nhầm tệp. Tệp đúng được đính kèm ở đây.','',2),
    ('xin-loi-va-cam-on','example','','Thank you for waiting so patiently.','Cảm ơn anh chị đã kiên nhẫn chờ đợi.','',3),

    -- B2
    ('neu-quan-diem','note','Khẳng định trần trụi mời người ta phản bác. Gắn mức độ chắc chắn vào câu — arguably, it seems, in my view — thì người nghe thấy chỗ để đối thoại thay vì phải chọn phe.','','','',0),
    ('neu-quan-diem','example','','Arguably, the plan is too ambitious.','Có thể nói kế hoạch này quá tham vọng.','',1),
    ('neu-quan-diem','example','','It seems the problem is timing, not money.','Có vẻ vấn đề nằm ở thời điểm, không phải tiền.','',2),
    ('neu-quan-diem','example','','In my view, we need more data first.','Theo tôi, trước hết chúng ta cần thêm dữ liệu.','',3),

    ('phan-bien','note','Phản biện nhắm vào lập luận, không nhắm vào người. Công thức an toàn: thừa nhận phần đúng trước, rồi mới nêu chỗ hổng.','','','',0),
    ('phan-bien','example','','You are right about the cost, but the timeline is unrealistic.','Anh nói đúng về chi phí, nhưng tiến độ thì không thực tế.','',1),
    ('phan-bien','example','','That is a fair point. However, it does not explain the drop in May.','Ý đó hợp lý. Tuy nhiên, nó không giải thích được mức giảm hồi tháng Năm.','',2),
    ('phan-bien','example','','I see why you think that, but the data suggests otherwise.','Tôi hiểu vì sao anh nghĩ vậy, nhưng số liệu lại cho thấy điều khác.','',3),

    ('dan-chung','note','Một câu chuyện không phải là bằng chứng. Khi dẫn số liệu, nói luôn nguồn và cỡ mẫu — thiếu hai thứ đó thì con số nào nghe cũng thuyết phục.','','','',0),
    ('dan-chung','example','','According to our survey of 400 customers, delivery time matters most.','Theo khảo sát của chúng tôi với 400 khách hàng, thời gian giao hàng là yếu tố quan trọng nhất.','',1),
    ('dan-chung','example','','One customer''s story is not enough to prove a trend.','Câu chuyện của một khách hàng không đủ để chứng minh một xu hướng.','',2),
    ('dan-chung','example','','The figure comes from the company''s annual report.','Con số này lấy từ báo cáo thường niên của công ty.','',3),

    ('chot-van-de','note','Chốt lại không phải là nhắc lại. Nêu điều đã thống nhất, điều còn mở, và việc tiếp theo của từng người.','','','',0),
    ('chot-van-de','example','','So, we have agreed on the price.','Vậy là chúng ta đã thống nhất về giá.','',1),
    ('chot-van-de','example','','The delivery date is still open.','Ngày giao hàng thì vẫn chưa chốt.','',2),
    ('chot-van-de','example','','Lan will send the contract by Thursday.','Lan sẽ gửi hợp đồng trước thứ Năm.','',3),

    ('hop-hanh','note','Xin nói trong cuộc họp bằng một câu ngắn báo trước ý mình: "Can I add something on that?" Nhảy thẳng vào nội dung dễ bị coi là cắt lời.','','','',0),
    ('hop-hanh','example','','Can I add something on that?','Tôi bổ sung một chút về việc đó được không?','',1),
    ('hop-hanh','example','','Sorry to interrupt, but could I ask a quick question?','Xin lỗi vì ngắt lời, nhưng tôi hỏi nhanh một câu được không?','',2),
    ('hop-hanh','example','','Could we go back to the budget for a moment?','Mình quay lại chuyện ngân sách một chút được không?','',3),
    ('hop-hanh','dialogue','','Could you clarify the deadline?','Bạn làm rõ giúp hạn chót được không?','Linh',4),
    ('hop-hanh','dialogue','','End of the month. Can we defer the budget item?','Cuối tháng. Mục ngân sách mình hoãn lại được không?','An',5),

    ('bao-cao-tien-do','note','Tin xấu nói sớm và nói kèm phương án thì vẫn giữ được lòng tin. Nói muộn thì bản thân sự chậm trễ trở thành vấn đề lớn hơn cái chậm tiến độ.','','','',0),
    ('bao-cao-tien-do','example','','I want to flag a risk early: testing may take an extra week.','Tôi muốn báo sớm một rủi ro: khâu kiểm thử có thể mất thêm một tuần.','',1),
    ('bao-cao-tien-do','example','','We are behind schedule, and here are two ways to catch up.','Chúng ta đang chậm tiến độ, và đây là hai cách để bắt kịp.','',2),
    ('bao-cao-tien-do','example','','The design phase is complete; development starts on Monday.','Giai đoạn thiết kế đã xong; phần phát triển bắt đầu từ thứ Hai.','',3),

    ('thuong-luong','note','Nhượng bộ nên có điều kiện đi kèm: "We could do that if..." Nhượng bộ trắng chỉ dạy đối phương rằng cứ đòi là được.','','','',0),
    ('thuong-luong','example','','We could extend the deadline if you cover the extra cost.','Chúng tôi có thể lùi hạn chót nếu bên anh chịu phần chi phí phát sinh.','',1),
    ('thuong-luong','example','','If you order a thousand units, we can offer free delivery.','Nếu anh đặt một nghìn sản phẩm, chúng tôi có thể miễn phí vận chuyển.','',2),
    ('thuong-luong','example','','That works for us, provided payment is made within thirty days.','Như vậy được, với điều kiện thanh toán trong vòng ba mươi ngày.','',3),

    ('phan-hoi-dong-nghiep','note','Góp ý dùng được phải nói rõ ba điều: việc gì, ảnh hưởng ra sao, và lần sau làm khác thế nào. Thiếu vế cuối thì đó chỉ là lời phàn nàn.','','','',0),
    ('phan-hoi-dong-nghiep','example','','The report arrived late, so the client meeting was delayed.','Báo cáo đến muộn, nên cuộc họp với khách bị lùi lại.','',1),
    ('phan-hoi-dong-nghiep','example','','Next time, could you let me know a day in advance?','Lần sau, bạn báo trước cho mình một ngày được không?','',2),
    ('phan-hoi-dong-nghiep','example','','Your summary was clear, and it saved us a lot of time.','Bản tóm tắt của bạn rất rõ ràng, và nó giúp cả nhóm tiết kiệm nhiều thời gian.','',3)
) AS v(lesson_slug, kind, body, text_en, text_vi, speaker, position)
JOIN lessons AS l ON l.slug = v.lesson_slug AND l.deleted_at IS NULL;

INSERT INTO lesson_blocks (lesson_id, kind, body, text_en, text_vi, speaker, position)
SELECT l.id, v.kind::lesson_block_kind, v.body, v.text_en, v.text_vi, v.speaker, v.position
FROM (VALUES
    -- C1
    ('trich-dan-nguon','note','Diễn đạt lại không phải là đổi vài từ đồng nghĩa. Phải đổi cả cấu trúc câu, và vẫn phải dẫn nguồn — giữ nguyên ý người khác mà không ghi nguồn vẫn là đạo văn.','','','',0),
    ('trich-dan-nguon','example','','As Nguyen (2021) argues, small schools adapt faster.','Như Nguyễn (2021) lập luận, các trường nhỏ thích nghi nhanh hơn.','',1),
    ('trich-dan-nguon','example','','Several studies suggest a similar pattern (Tran, 2019; Le, 2022).','Một số nghiên cứu cho thấy một xu hướng tương tự (Trần, 2019; Lê, 2022).','',2),
    ('trich-dan-nguon','example','','In her words, the reform was ''a quiet revolution''.','Theo lời bà ấy, cuộc cải cách là ''một cuộc cách mạng thầm lặng''.','',3),

    ('mo-ta-du-lieu','note','Mô tả dữ liệu cần đúng mức. plummet là rơi thẳng đứng, không dùng cho mức giảm 3%. negligible nghĩa là nhỏ tới mức bỏ qua được, không phải chỉ là nhỏ.','','','',0),
    ('mo-ta-du-lieu','example','','Sales fell slightly in the second quarter.','Doanh số giảm nhẹ trong quý hai.','',1),
    ('mo-ta-du-lieu','example','','The number of users rose sharply after the update.','Số người dùng tăng mạnh sau bản cập nhật.','',2),
    ('mo-ta-du-lieu','example','','Prices remained stable throughout the year.','Giá cả giữ ổn định suốt cả năm.','',3),

    ('gioi-han-nghien-cuu','note','Tự nêu giới hạn của nghiên cứu mình không làm nó yếu đi — nó cho thấy tác giả biết mình đang nói gì. Bài báo nào cũng có mục này, và người phản biện đọc nó đầu tiên.','','','',0),
    ('gioi-han-nghien-cuu','example','','The sample was small and drawn from one city.','Mẫu khảo sát nhỏ và chỉ lấy từ một thành phố.','',1),
    ('gioi-han-nghien-cuu','example','','These findings may not apply to older learners.','Những kết quả này có thể không áp dụng cho người học lớn tuổi.','',2),
    ('gioi-han-nghien-cuu','example','','Further research is needed to confirm the effect.','Cần nghiên cứu thêm để khẳng định tác động này.','',3),

    ('thuat-ngu-hoc-thuat','note','empirical nghĩa là dựa trên quan sát hoặc thí nghiệm, không phải chỉ là "có số liệu". Dùng sai từ này là dấu hiệu rõ nhất của một bài viết chưa chắc tay.','','','',0),
    ('thuat-ngu-hoc-thuat','example','','The study provides empirical evidence from three field trials.','Nghiên cứu đưa ra bằng chứng thực nghiệm từ ba đợt thử nghiệm thực địa.','',1),
    ('thuat-ngu-hoc-thuat','example','','Her argument is theoretical rather than empirical.','Lập luận của cô ấy mang tính lý thuyết hơn là thực nghiệm.','',2),
    ('thuat-ngu-hoc-thuat','example','','We tested the hypothesis empirically, using classroom data.','Chúng tôi kiểm chứng giả thuyết bằng thực nghiệm, dùng dữ liệu từ lớp học.','',3),

    ('thanh-ngu-thong-dung','note','Những tính từ này khác nhau ở cường độ, không phải ở nghĩa. tedious là chán, daunting là nản trước khi bắt đầu, gruelling là vắt kiệt sức khi đã làm.','','','',0),
    ('thanh-ngu-thong-dung','example','','Filling in the forms was tedious.','Điền đống giấy tờ đó thật chán.','',1),
    ('thanh-ngu-thong-dung','example','','The size of the task was daunting at first.','Lúc đầu khối lượng công việc khiến người ta nản.','',2),
    ('thanh-ngu-thong-dung','example','','After a gruelling week, she slept for twelve hours.','Sau một tuần vắt kiệt sức, cô ấy ngủ liền mười hai tiếng.','',3),

    ('cum-dong-tu','note','Văn bản trang trọng thường thay cụm động từ bằng một động từ gốc Latin: speed up thành expedite, hold back thành impede, give up thành relinquish. Nghĩa gần nhau, sắc thái khác hẳn.','','','',0),
    ('cum-dong-tu','example','','The council decided to postpone the vote.','Hội đồng quyết định hoãn cuộc bỏ phiếu.','',1),
    ('cum-dong-tu','example','','Please submit your report by Friday.','Vui lòng nộp báo cáo trước thứ Sáu.','',2),
    ('cum-dong-tu','example','','The company will investigate the complaint.','Công ty sẽ điều tra khiếu nại này.','',3),

    ('noi-giam-noi-tranh','note','Tiếng Anh trang trọng nói nhẹ đi theo phản xạ. "The improvement was marginal" thường có nghĩa là gần như không cải thiện gì. Người học cần nghe ra lớp nghĩa đó, không chỉ dịch mặt chữ.','','','',0),
    ('noi-giam-noi-tranh','example','','The results were not entirely encouraging.','Kết quả không hẳn là đáng khích lệ.','',1),
    ('noi-giam-noi-tranh','example','','There is some room for improvement.','Vẫn còn chỗ để cải thiện.','',2),
    ('noi-giam-noi-tranh','example','','I am not sure that is the best idea.','Tôi không chắc đó là ý hay nhất.','',3),

    ('sac-thai-trang-trong','note','thereby, hence, whereby thuộc văn viết chính thức. Đưa chúng vào hội thoại thường ngày nghe gượng, giống như nói tiếng Việt kiểu công văn.','','','',0),
    ('sac-thai-trang-trong','example','','The contract was signed, thereby ending the dispute.','Hợp đồng đã được ký, qua đó chấm dứt tranh chấp.','',1),
    ('sac-thai-trang-trong','example','','Members pay a fee, whereby they gain access to the archive.','Thành viên trả một khoản phí, nhờ đó được truy cập kho lưu trữ.','',2),
    ('sac-thai-trang-trong','example','','The road was closed; hence the delay.','Con đường bị đóng; do đó mới chậm trễ.','',3),
    ('sac-thai-trang-trong','example','','We are obliged to report this.','Chúng tôi có nghĩa vụ phải báo cáo việc này.','',4),

    -- C2
    ('bien-phap-tu-tu','note','Nhận ra biện pháp tu từ quan trọng hơn dùng nó. Uyển ngữ trong bản tin, ẩn dụ trong diễn văn chính trị — chúng định hướng người đọc trước khi người đọc kịp nghĩ.','','','',0),
    ('bien-phap-tu-tu','example','','The minister called the cuts ''a period of adjustment''.','Bộ trưởng gọi việc cắt giảm là ''một giai đoạn điều chỉnh''.','',1),
    ('bien-phap-tu-tu','example','','The city was a machine that never slept.','Thành phố là một cỗ máy không bao giờ ngủ.','',2),
    ('bien-phap-tu-tu','example','','It was the best of plans; it was the worst of plans.','Đó là kế hoạch hay nhất; đó là kế hoạch tệ nhất.','',3),

    ('nhip-dieu-cau-van','note','Một câu ngắn sau vài câu dài có sức nặng mà bản thân nó không có. Đó là hiệu ứng của tương phản, không phải của độ ngắn.','','','',0),
    ('nhip-dieu-cau-van','example','','We planned for months, tested every detail and hired the best people. It failed anyway.','Chúng tôi lên kế hoạch hàng tháng trời, thử từng chi tiết và thuê những người giỏi nhất. Vậy mà vẫn thất bại.','',1),
    ('nhip-dieu-cau-van','example','','The road wound through the hills, past empty farms and silent villages. Then, the sea.','Con đường uốn quanh những ngọn đồi, qua những trang trại bỏ hoang và những ngôi làng lặng im. Rồi, biển.','',2),
    ('nhip-dieu-cau-van','example','','She listened to every argument and weighed every risk. Then she said no.','Cô ấy nghe hết mọi lập luận và cân nhắc mọi rủi ro. Rồi cô ấy nói không.','',3),

    ('tu-vung-tinh-te','note','ostensibly báo hiệu người viết không tin vào lý do vừa nêu. Một từ đổi hẳn thái độ của cả câu — đó là lý do phải đọc kỹ những từ như thế.','','','',0),
    ('tu-vung-tinh-te','example','','The meeting was ostensibly friendly.','Cuộc gặp bề ngoài có vẻ thân thiện.','',1),
    ('tu-vung-tinh-te','example','','He was, apparently, too busy to reply.','Anh ta, nghe đâu, quá bận để trả lời.','',2),
    ('tu-vung-tinh-te','example','','The so-called reform changed very little.','Cái gọi là cải cách đã thay đổi rất ít.','',3),

    ('van-phong-bao-chi','note','Báo chí nghiêm túc kiểm chứng bằng nhiều nguồn độc lập. Khi một bài chỉ dẫn "sources say" mà không nói gì thêm, đó là chỗ cần dè chừng.','','','',0),
    ('van-phong-bao-chi','example','','Three independent sources confirmed the report.','Ba nguồn độc lập đã xác nhận bản tin này.','',1),
    ('van-phong-bao-chi','example','','The ministry declined to comment.','Bộ đã từ chối bình luận.','',2),
    ('van-phong-bao-chi','example','','The claim could not be independently verified.','Thông tin này chưa thể được kiểm chứng độc lập.','',3),

    ('mo-dau-an-tuong','note','Mở đầu không cần gây sốc, cần gây tò mò. Một câu hỏi mà người nghe chưa có câu trả lời sẽ giữ họ lâu hơn một con số giật gân.','','','',0),
    ('mo-dau-an-tuong','example','','What if the problem is not the price, but the wait?','Nếu vấn đề không nằm ở giá, mà ở thời gian chờ thì sao?','',1),
    ('mo-dau-an-tuong','example','','Ten years ago, this room did not exist.','Mười năm trước, căn phòng này chưa hề tồn tại.','',2),
    ('mo-dau-an-tuong','example','','Think of the last time you changed your mind.','Hãy nghĩ về lần gần nhất bạn thay đổi ý kiến.','',3),

    ('dung-lap-luan','note','Nguỵ biện phổ biến nhất trong tranh luận đời thường là dựng người rơm: phản bác một phiên bản méo mó của ý đối phương. Nhận ra nó là bước đầu để không mắc phải.','','','',0),
    ('dung-lap-luan','example','','Nobody said we should ban cars; the proposal is to close one street.','Không ai nói nên cấm ô tô; đề xuất chỉ là đóng một con phố.','',1),
    ('dung-lap-luan','example','','Allowing one exception does not mean the rule will collapse.','Cho phép một ngoại lệ không có nghĩa là quy định sẽ sụp đổ.','',2),
    ('dung-lap-luan','example','','Attack the argument, not the person making it.','Hãy phản bác lập luận, đừng công kích người đưa ra nó.','',3),

    ('keo-huong-nguoi-nghe','note','Khoảng lặng làm được điều mà âm lượng không làm được: nó buộc người nghe tự lấp vào chỗ trống, và thứ họ tự nghĩ ra thì họ nhớ lâu hơn.','','','',0),
    ('keo-huong-nguoi-nghe','example','','And then — nothing.','Và rồi — không có gì cả.','',1),
    ('keo-huong-nguoi-nghe','example','','Let that number sink in for a moment.','Hãy để con số đó ngấm một chút.','',2),
    ('keo-huong-nguoi-nghe','example','','What would you have done?','Nếu là bạn, bạn sẽ làm gì?','',3),

    ('ket-bai-dong-lai','note','Kết bài nên để lại một thứ, không phải tóm tắt mọi thứ. Người nghe mang về được một câu hoặc một hình ảnh là đủ.','','','',0),
    ('ket-bai-dong-lai','example','','If you remember one thing today, remember this: start small.','Nếu hôm nay bạn chỉ nhớ một điều, hãy nhớ điều này: bắt đầu từ việc nhỏ.','',1),
    ('ket-bai-dong-lai','example','','My grandmother never saw the garden finished. We still water it.','Bà tôi chưa bao giờ thấy khu vườn hoàn thành. Chúng tôi vẫn tưới nó.','',2),
    ('ket-bai-dong-lai','example','','That is where we started, and that is where I will stop.','Chúng ta bắt đầu từ đó, và tôi sẽ dừng lại ở đó.','',3)
) AS v(lesson_slug, kind, body, text_en, text_vi, speaker, position)
JOIN lessons AS l ON l.slug = v.lesson_slug AND l.deleted_at IS NULL;
