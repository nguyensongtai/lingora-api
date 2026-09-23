-- Hội thoại cho những bài chưa có đoạn nào — bước "Hội thoại" của màn học.
--
-- Do Claude soạn, giống 005: KHÔNG có nguồn nào đối chiếu được bằng máy.
-- Mỗi đoạn dựng quanh đúng điểm chính của bài (ghi chú ở 005), không phải
-- quanh danh sách từ: hội thoại dạy cách dùng, còn từ đã có bước riêng.
-- Cần người có chuyên môn đọc lại trước khi có người học thật.
--
-- Chạy SAU 005: 005 xoá sạch khối của các khoá nó phụ trách. File này tự
-- xoá hội thoại của đúng những bài nó phụ trách trước khi chèn, nên chạy lại
-- bao nhiêu lần cũng hội tụ.

DELETE FROM lesson_blocks b
USING lessons l
WHERE b.lesson_id = l.id AND b.kind = 'dialogue' AND l.deleted_at IS NULL
  AND l.slug IN (
      'mao-tu-a-an-the', 'danh-tu-dem-duoc', 'mau-sac-va-hinh-dang',
      'dong-tu-to-be', 'thoi-gian-va-ngay-thang', 'thi-hien-tai-don',
      'thi-qua-khu-don', 'dong-tu-bat-quy-tac', 'noi-ve-du-dinh',
      'mo-ta-nguoi', 'email-trang-trong', 'email-than-mat',
      'mo-ta-noi-chon', 'ke-mot-su-viec', 'cam-xuc-va-thai-do',
      'xin-loi-va-cam-on', 'neu-quan-diem', 'bao-cao-tien-do',
      'phan-bien', 'dan-chung', 'thuong-luong',
      'chot-van-de', 'phan-hoi-dong-nghiep', 'trich-dan-nguon',
      'thanh-ngu-thong-dung', 'cum-dong-tu', 'mo-ta-du-lieu',
      'gioi-han-nghien-cuu', 'noi-giam-noi-tranh', 'thuat-ngu-hoc-thuat',
      'sac-thai-trang-trong', 'bien-phap-tu-tu', 'mo-dau-an-tuong',
      'dung-lap-luan', 'nhip-dieu-cau-van', 'tu-vung-tinh-te',
      'keo-huong-nguoi-nghe', 'van-phong-bao-chi', 'ket-bai-dong-lai'
  );

INSERT INTO lesson_blocks (lesson_id, kind, body, text_en, text_vi, speaker, position)
SELECT l.id, 'dialogue', '', v.text_en, v.text_vi, v.speaker, v.position
FROM (VALUES
    -- A1 · Mạo từ
    ('mao-tu-a-an-the','Mai','Is it going to rain? I do not have an umbrella.','Trời sắp mưa à? Mình không có ô.',100),
    ('mao-tu-a-an-the','Tom','Take the umbrella by the door. It is mine.','Cầm cái ô cạnh cửa đi. Của mình đấy.',101),
    ('mao-tu-a-an-the','Mai','Thanks. I have an interview at the office in an hour.','Cảm ơn nhé. Một tiếng nữa mình có buổi phỏng vấn ở văn phòng.',102),
    ('mao-tu-a-an-the','Tom','Then it is a good idea to leave now. The bus is often late.','Vậy đi luôn bây giờ là ý hay đấy. Xe buýt hay trễ lắm.',103),

    -- A1 · Danh từ đếm được và không đếm được
    ('danh-tu-dem-duoc','Mai','What do we need from the shop?','Mình cần mua gì ở cửa hàng?',100),
    ('danh-tu-dem-duoc','Tom','Some rice, some bread and two bottles of milk.','Ít gạo, ít bánh mì và hai chai sữa.',101),
    ('danh-tu-dem-duoc','Mai','Is there any sugar at home?','Ở nhà còn đường không?',102),
    ('danh-tu-dem-duoc','Tom','A little. And get a piece of cake for me, please.','Còn một ít. Mua giúp mình một miếng bánh ngọt nhé.',103),

    -- A1 · Màu sắc và hình dáng
    ('mau-sac-va-hinh-dang','Linh','Which coat do you like?','Bạn thích cái áo khoác nào?',100),
    ('mau-sac-va-hinh-dang','Sam','The long dark one. It looks warm.','Cái dài màu tối. Trông ấm.',101),
    ('mau-sac-va-hinh-dang','Linh','I prefer the short red coat. It is a bright colour.','Mình thích cái áo ngắn màu đỏ hơn. Màu tươi.',102),
    ('mau-sac-va-hinh-dang','Sam','Red suits you. But the dark one is thicker.','Màu đỏ hợp với bạn. Nhưng cái màu tối dày hơn.',103),

    -- A1 · Động từ to be
    ('dong-tu-to-be','Mai','Are you ready? The film starts at eight.','Bạn xong chưa? Phim bắt đầu lúc tám giờ.',100),
    ('dong-tu-to-be','Tom','Not yet. I am tired and I am hungry.','Chưa. Mình mệt mà lại đói nữa.',101),
    ('dong-tu-to-be','Mai','The restaurant next to the cinema is not busy tonight.','Tối nay quán ăn cạnh rạp không đông đâu.',102),
    ('dong-tu-to-be','Tom','Great. Are your friends there already?','Tuyệt. Bạn của bạn tới đó chưa?',103),
    ('dong-tu-to-be','Mai','Yes, they are. And they are hungry too.','Rồi. Mà họ cũng đang đói.',104),

    -- A1 · Thời gian và ngày tháng
    ('thoi-gian-va-ngay-thang','Linh','Are you free on Friday evening?','Tối thứ Sáu bạn rảnh không?',100),
    ('thoi-gian-va-ngay-thang','Sam','Friday is busy. What about tomorrow? I finish work at six.','Thứ Sáu bận rồi. Mai thì sao? Mình xong việc lúc sáu giờ.',101),
    ('thoi-gian-va-ngay-thang','Linh','Tomorrow is fine. Let us meet at seven, in front of the cinema.','Mai được. Gặp nhau lúc bảy giờ trước rạp chiếu phim nhé.',102),
    ('thoi-gian-va-ngay-thang','Sam','OK. I will be there a few minutes early.','Ừ. Mình sẽ tới sớm vài phút.',103),

    -- A1 · Thì hiện tại đơn
    ('thi-hien-tai-don','Mai','What do you usually do at the weekend?','Cuối tuần bạn thường làm gì?',100),
    ('thi-hien-tai-don','Tom','I often play football on Saturday morning.','Mình hay đá bóng vào sáng thứ Bảy.',101),
    ('thi-hien-tai-don','Mai','And on Sunday?','Còn Chủ nhật?',102),
    ('thi-hien-tai-don','Tom','My family always has lunch together. I never miss it.','Nhà mình lúc nào cũng ăn trưa cùng nhau. Mình không bao giờ bỏ bữa đó.',103),

    -- A2 · Thì quá khứ đơn
    ('thi-qua-khu-don','Linh','Did you go to Nam''s party last night?','Tối qua bạn có đi tiệc của Nam không?',100),
    ('thi-qua-khu-don','Sam','Yes. He invited me last week, so I decided to go.','Có. Tuần trước cậu ấy mời mình, nên mình quyết định đi.',101),
    ('thi-qua-khu-don','Linh','Did the neighbours complain about the noise?','Hàng xóm có phàn nàn về tiếng ồn không?',102),
    ('thi-qua-khu-don','Sam','One did. Nam said sorry, and she accepted his apology.','Có một người. Nam xin lỗi, và bà ấy nhận lời xin lỗi.',103),

    -- A2 · Động từ bất quy tắc
    ('dong-tu-bat-quy-tac','Mai','What happened to your hand?','Tay bạn bị sao thế?',100),
    ('dong-tu-bat-quy-tac','Tom','I burnt it on the stove when I made dinner.','Mình chạm vào bếp lúc nấu bữa tối nên bị bỏng.',101),
    ('dong-tu-bat-quy-tac','Mai','And your bike? The wheel looks bent.','Còn xe đạp? Bánh xe trông cong rồi.',102),
    ('dong-tu-bat-quy-tac','Tom','A dog ran at me and I fell. It nearly bit me!','Một con chó lao vào nên mình ngã. Suýt nữa nó cắn mình!',103),

    -- A2 · Nói về dự định
    ('noi-ve-du-dinh','Linh','What are you doing after graduation?','Tốt nghiệp xong bạn định làm gì?',100),
    ('noi-ve-du-dinh','Sam','I am starting a job in Hanoi in July. It is all arranged.','Tháng Bảy mình bắt đầu đi làm ở Hà Nội. Mọi thứ đã sắp xếp xong.',101),
    ('noi-ve-du-dinh','Linh','And your plan to study abroad?','Còn dự định đi du học thì sao?',102),
    ('noi-ve-du-dinh','Sam','I am going to prepare for it next year. I will probably apply in 2028.','Năm sau mình sẽ chuẩn bị. Chắc năm 2028 mình mới nộp hồ sơ.',103),
    ('noi-ve-du-dinh','Linh','Good choice. Take the chance while you can.','Lựa chọn hay đấy. Có cơ hội thì cứ nắm lấy.',104),

    -- B1 · Mô tả một người
    ('mo-ta-nguoi','Hà','What is your new manager like?','Sếp mới của bạn là người thế nào?',100),
    ('mo-ta-nguoi','Minh','She is very reliable and quite cheerful, even on Mondays.','Chị ấy rất đáng tin và khá vui vẻ, kể cả vào thứ Hai.',101),
    ('mo-ta-nguoi','Hà','Any weaknesses?','Có điểm gì chưa tốt không?',102),
    ('mo-ta-nguoi','Minh','She can be a bit stubborn about deadlines. But she is honest with us, and I respect that.','Chị ấy hơi cố chấp chuyện hạn chót. Nhưng chị ấy thẳng thắn với bọn mình, và mình quý điều đó.',103),

    -- B1 · Email trang trọng
    ('email-trang-trong','Linh','Did you reply to the enquiry regarding the training course?','Bạn đã trả lời thư hỏi về khoá đào tạo chưa?',100),
    ('email-trang-trong','An','Yes. I attached the price list and the application form.','Rồi. Mình đã đính kèm bảng giá và mẫu đơn đăng ký.',101),
    ('email-trang-trong','Linh','Did you inform them that we need approval first?','Bạn có báo họ là cần được duyệt trước không?',102),
    ('email-trang-trong','An','I did. I put it in the first paragraph, so they cannot miss it.','Có. Mình để ngay đoạn đầu để họ không bỏ sót.',103),

    -- B1 · Email thân mật
    ('email-than-mat','Tom','Hey buddy, long time no see. What have you been up to lately?','Này ông bạn, lâu quá không gặp. Dạo này làm gì thế?',100),
    ('email-than-mat','Nam','Sorry, I owe you an apology. I never replied to your email.','Xin lỗi nhé, mình nợ bạn một lời xin lỗi. Mình chưa trả lời email của bạn.',101),
    ('email-than-mat','Tom','No worries. So, any updates? New job?','Không sao. Có gì mới không? Việc mới à?',102),
    ('email-than-mat','Nam','I suppose you could say that. Let us definitely get coffee this week.','Cũng có thể nói vậy. Tuần này nhất định đi cà phê nhé.',103),

    -- B1 · Mô tả nơi chốn
    ('mo-ta-noi-chon','Mai','How was the cabin you rented?','Căn nhà gỗ bạn thuê thế nào?',100),
    ('mo-ta-noi-chon','Tom','Quite bare inside, just a bed and a stove. But the atmosphere was so peaceful.','Bên trong khá trống, chỉ có cái giường với cái bếp. Nhưng không khí thì yên bình lắm.',101),
    ('mo-ta-noi-chon','Mai','Where exactly was it?','Chính xác thì nó ở đâu?',102),
    ('mo-ta-noi-chon','Tom','In the hills near the border, by a broad, slow river.','Trên đồi gần biên giới, cạnh một con sông rộng chảy chậm.',103),

    -- B1 · Kể lại một sự việc
    ('ke-mot-su-viec','Hà','I heard there was an incident on your flight.','Nghe nói chuyến bay của bạn có sự cố.',100),
    ('ke-mot-su-viec','Minh','Yes. Suddenly a passenger fell ill, and the crew immediately asked for a doctor.','Ừ. Đột nhiên một hành khách bị ốm, và tổ bay lập tức hỏi có bác sĩ không.',101),
    ('ke-mot-su-viec','Hà','Was there one on board?','Trên máy bay có ai là bác sĩ không?',102),
    ('ke-mot-su-viec','Minh','Eventually a nurse came forward. Meanwhile, the pilot turned back to Da Nang.','Cuối cùng có một y tá đứng ra. Trong lúc đó, phi công quay đầu về Đà Nẵng.',103),
    ('ke-mot-su-viec','Hà','Is the passenger OK?','Hành khách đó có sao không?',104),
    ('ke-mot-su-viec','Minh','Yes. We heard afterwards that she was fine.','Không sao. Sau đó bọn mình nghe tin bà ấy đã ổn.',105),

    -- B1 · Cảm xúc và thái độ
    ('cam-xuc-va-thai-do','Linh','You look annoyed. What happened?','Trông bạn bực mình thế. Có chuyện gì vậy?',100),
    ('cam-xuc-va-thai-do','Sam','I called my boss by the wrong name in the meeting. I was so embarrassed.','Mình gọi sai tên sếp trong cuộc họp. Ngượng chết đi được.',101),
    ('cam-xuc-va-thai-do','Linh','That is not a big deal. You should not feel guilty about it.','Chuyện đó có gì to tát đâu. Bạn đừng thấy có lỗi.',102),
    ('cam-xuc-va-thai-do','Sam','I know. I am just amazed that nobody laughed.','Mình biết. Chỉ là ngạc nhiên vì chẳng ai cười.',103),

    -- B1 · Xin lỗi và cảm ơn
    ('xin-loi-va-cam-on','Lễ tân','On behalf of the hotel, I apologise. We gave your room to another guest.','Thay mặt khách sạn, tôi xin lỗi. Chúng tôi đã giao nhầm phòng của anh chị cho khách khác.',100),
    ('xin-loi-va-cam-on','Khách','That was a long wait after a long flight.','Phải chờ lâu như vậy sau một chuyến bay dài.',101),
    ('xin-loi-va-cam-on','Lễ tân','I acknowledge that. We have moved you to a suite at no extra cost.','Tôi hiểu. Chúng tôi đã chuyển anh chị sang phòng suite, không tính thêm phí.',102),
    ('xin-loi-va-cam-on','Khách','Thank you. I appreciate that.','Cảm ơn. Tôi rất cảm kích.',103),

    -- B2 · Nêu quan điểm
    ('neu-quan-diem','An','Remote work is better for everyone. That is a fact.','Làm việc từ xa tốt hơn cho tất cả mọi người. Đó là sự thật.',100),
    ('neu-quan-diem','Linh','That is quite an assumption. The research is ambiguous.','Đó là một giả định khá lớn. Các nghiên cứu còn chưa rõ ràng.',101),
    ('neu-quan-diem','An','Fair point. Let me put it another way: in my view, it suits our team.','Cũng đúng. Để mình nói lại: theo mình, nó hợp với nhóm mình.',102),
    ('neu-quan-diem','Linh','Now that I can discuss. It seems the crucial question is how we measure it.','Vậy thì bàn được. Có vẻ câu hỏi then chốt là mình đo hiệu quả bằng cách nào.',103),

    -- B2 · Báo cáo tiến độ
    ('bao-cao-tien-do','Hà','How is the data project going?','Dự án dữ liệu tiến triển thế nào rồi?',100),
    ('bao-cao-tien-do','Minh','I need to flag a delay early. The old system is more complex than we anticipated.','Tôi cần báo sớm một việc chậm. Hệ thống cũ phức tạp hơn chúng ta dự tính.',101),
    ('bao-cao-tien-do','Hà','How long, and what are our options?','Chậm bao lâu, và mình có những phương án nào?',102),
    ('bao-cao-tien-do','Minh','About two weeks. Or we bring in a second analyst and keep the date. My assessment is that the second option is safer.','Khoảng hai tuần. Hoặc thêm một chuyên viên phân tích để giữ đúng hạn. Theo đánh giá của tôi, phương án thứ hai an toàn hơn.',103),

    -- B2 · Phản biện
    ('phan-bien','Linh','The new app will save us time. Everyone agrees.','Ứng dụng mới sẽ tiết kiệm thời gian. Ai cũng đồng ý.',100),
    ('phan-bien','An','I concede that it is faster for bookings. But the survey contradicts the idea that everyone agrees.','Mình công nhận nó đặt lịch nhanh hơn. Nhưng khảo sát lại cho thấy không phải ai cũng đồng ý.',101),
    ('phan-bien','Linh','Which part of the survey?','Phần nào của khảo sát?',102),
    ('phan-bien','An','Older customers found it confusing. I think we overlooked them.','Khách lớn tuổi thấy nó khó dùng. Mình nghĩ chúng ta đã bỏ sót nhóm này.',103),

    -- B2 · Đưa dẫn chứng
    ('dan-chung','Hà','Sales went up after the ad campaign, so the campaign worked.','Doanh số tăng sau chiến dịch quảng cáo, vậy là chiến dịch hiệu quả.',100),
    ('dan-chung','Minh','Maybe. But can we attribute the rise to the ads alone? It was also the holiday season.','Có thể. Nhưng mình có thể quy mức tăng đó hoàn toàn cho quảng cáo không? Lúc đó cũng là mùa lễ.',101),
    ('dan-chung','Hà','What would convince you?','Cần gì thì anh mới tin?',102),
    ('dan-chung','Minh','Statistics from last year''s holidays, with the source. If last year shows a similar rise, the ads made little difference.','Số liệu mùa lễ năm ngoái, có ghi nguồn. Nếu năm ngoái cũng tăng tương tự, thì quảng cáo không tạo ra khác biệt mấy.',103),

    -- B2 · Thương lượng
    ('thuong-luong','Linh','Your price is considerably higher than last year.','Giá của anh cao hơn năm ngoái đáng kể.',100),
    ('thuong-luong','David','Our costs went up. But we could lower it by five percent if you sign for two years.','Chi phí bên tôi tăng. Nhưng chúng tôi có thể giảm năm phần trăm nếu chị ký hợp đồng hai năm.',101),
    ('thuong-luong','Linh','Two years is long. Alternatively, we could pay for one year in advance.','Hai năm thì dài. Hoặc chúng tôi có thể trả trước một năm.',102),
    ('thuong-luong','David','That could work, if you can be flexible on delivery dates.','Cách đó được, nếu chị linh hoạt được về ngày giao hàng.',103),

    -- B2 · Chốt lại vấn đề
    ('chot-van-de','Hà','Let us wrap up. We agreed to move the launch to May.','Mình chốt lại nhé. Chúng ta đã thống nhất dời ngày ra mắt sang tháng Năm.',100),
    ('chot-van-de','Minh','Yes. The budget is still open.','Đúng. Còn ngân sách thì vẫn chưa chốt.',101),
    ('chot-van-de','Hà','Right. Minh, you will send the new plan by Friday. I will ask the director for her consent.','Đúng vậy. Minh gửi kế hoạch mới trước thứ Sáu. Tôi sẽ xin giám đốc đồng ý.',102),
    ('chot-van-de','Minh','Ultimately it is her decision, so that makes sense.','Rốt cuộc thì chị ấy quyết định, nên như vậy là hợp lý.',103),

    -- B2 · Phản hồi đồng nghiệp
    ('phan-hoi-dong-nghiep','Hà','Do you have a minute? I want to give you some feedback on the client email.','Bạn có một phút không? Mình muốn góp ý về email gửi khách hàng.',100),
    ('phan-hoi-dong-nghiep','Minh','Sure.','Được chứ.',101),
    ('phan-hoi-dong-nghiep','Hà','The two prices in it were contradictory, so the client called to ask which one was right.','Trong email có hai mức giá mâu thuẫn nhau, nên khách phải gọi hỏi cái nào đúng.',102),
    ('phan-hoi-dong-nghiep','Minh','You are right. I copied them from old correspondence.','Đúng thật. Mình đã chép từ thư từ cũ.',103),
    ('phan-hoi-dong-nghiep','Hà','Next time, could you check the numbers against the price list before you send it?','Lần sau, bạn đối chiếu con số với bảng giá trước khi gửi được không?',104),

    -- C1 · Trích dẫn nguồn
    ('trich-dan-nguon','Giảng viên','This paragraph is very close to the original article.','Đoạn này rất sát với bài báo gốc.',100),
    ('trich-dan-nguon','Anna','But I changed several words.','Nhưng em đã đổi vài từ rồi mà.',101),
    ('trich-dan-nguon','Giảng viên','Changing words is not enough. Rewrite the structure, and add a citation even then.','Đổi từ thôi chưa đủ. Em phải viết lại cả cấu trúc câu, và kể cả vậy vẫn phải ghi nguồn.',102),
    ('trich-dan-nguon','Anna','I see. Is a blog a credible source for this assertion?','Em hiểu rồi. Một trang blog có phải là nguồn đáng tin cho nhận định này không ạ?',103),
    ('trich-dan-nguon','Giảng viên','Only if you cannot find a peer-reviewed one. Otherwise it weakens your credibility.','Chỉ khi em không tìm được nguồn đã qua bình duyệt. Nếu không, nó làm giảm độ tin cậy của em.',104),

    -- C1 · Thành ngữ thông dụng
    ('thanh-ngu-thong-dung','Linh','How is the marathon training going?','Tập chạy marathon thế nào rồi?',100),
    ('thanh-ngu-thong-dung','Sam','Gruelling. Twenty kilometres every Sunday.','Kiệt sức. Chủ nhật nào cũng hai mươi cây số.',101),
    ('thanh-ngu-thong-dung','Linh','It sounds daunting. I would not even start.','Nghe đã thấy nản. Mình còn chẳng dám bắt đầu.',102),
    ('thanh-ngu-thong-dung','Sam','The running is fine. The stretching afterwards is what I find tedious.','Chạy thì không sao. Giãn cơ sau đó mới là phần mình thấy chán.',103),

    -- C1 · Cụm động từ
    ('cum-dong-tu','An','In the email to the ministry, I wrote ''Can you speed up the permit?''','Trong email gửi bộ, mình viết ''Can you speed up the permit?''',100),
    ('cum-dong-tu','Linh','For a ministry, ''expedite'' sounds better: ''Could you expedite the permit?''','Gửi cơ quan nhà nước thì ''expedite'' nghe hợp hơn: ''Could you expedite the permit?''',101),
    ('cum-dong-tu','An','And ''the delay is holding back the project''?','Còn ''the delay is holding back the project''?',102),
    ('cum-dong-tu','Linh','''The delay is impeding the project.'' Save the phrasal verbs for the team chat.','''The delay is impeding the project.'' Cụm động từ thì để dành cho nhóm chat nội bộ.',103),

    -- C1 · Mô tả dữ liệu
    ('mo-ta-du-lieu','Hà','Can I say sales plummeted in March?','Tôi viết doanh số tháng Ba lao dốc được không?',100),
    ('mo-ta-du-lieu','Minh','They fell by three percent. That is not plummeting.','Chỉ giảm ba phần trăm. Như vậy không phải là lao dốc.',101),
    ('mo-ta-du-lieu','Hà','Negligible, then?','Vậy là không đáng kể?',102),
    ('mo-ta-du-lieu','Minh','Not quite. Three percent of our revenue is still a lot of money. ''A slight decline'' is more precise.','Chưa hẳn. Ba phần trăm doanh thu vẫn là một khoản tiền lớn. ''Giảm nhẹ'' thì chính xác hơn.',103),

    -- C1 · Giới hạn của nghiên cứu
    ('gioi-han-nghien-cuu','Anna','Should I mention that we only surveyed eighty students?','Em có nên nêu là mình chỉ khảo sát tám mươi sinh viên không ạ?',100),
    ('gioi-han-nghien-cuu','Giảng viên','Yes. Reviewers will notice it anyway. It is better that you name the constraint first.','Có. Người phản biện kiểu gì cũng thấy. Tốt hơn là em tự nêu giới hạn đó trước.',101),
    ('gioi-han-nghien-cuu','Anna','Will it not make the paper look weak?','Như vậy bài có bị yếu đi không ạ?',102),
    ('gioi-han-nghien-cuu','Giảng viên','No. It shows you know which conclusions are solid and which are speculative.','Không. Nó cho thấy em biết kết luận nào chắc chắn và kết luận nào còn là suy đoán.',103),

    -- C1 · Nói giảm nói tránh
    ('noi-giam-noi-tranh','Hà','The client said our proposal was ''interesting''. That is good, right?','Khách hàng nói đề xuất của mình ''thú vị''. Vậy là tốt, đúng không?',100),
    ('noi-giam-noi-tranh','David','Not necessarily. In a formal meeting, ''interesting'' often means they are not convinced.','Chưa chắc. Trong một cuộc họp trang trọng, ''thú vị'' thường có nghĩa là họ chưa bị thuyết phục.',101),
    ('noi-giam-noi-tranh','Hà','They also said the savings looked marginal.','Họ còn nói khoản tiết kiệm có vẻ không đáng kể.',102),
    ('noi-giam-noi-tranh','David','Then they are telling us, politely, that the numbers are too small.','Vậy là họ đang nói một cách lịch sự rằng con số quá nhỏ.',103),

    -- C1 · Thuật ngữ học thuật
    ('thuat-ngu-hoc-thuat','Anna','My essay says the claim is empirical because it has numbers.','Bài luận của em nói nhận định đó mang tính thực nghiệm vì nó có số liệu.',100),
    ('thuat-ngu-hoc-thuat','Giảng viên','Numbers are not enough. Empirical means it comes from observation or experiment.','Có số thôi chưa đủ. Thực nghiệm nghĩa là nó đến từ quan sát hoặc thí nghiệm.',101),
    ('thuat-ngu-hoc-thuat','Anna','The figures came from a model, not from measurements.','Các con số đó lấy từ một mô hình, không phải từ đo đạc ạ.',102),
    ('thuat-ngu-hoc-thuat','Giảng viên','Then call it a projection, and scrutinise the model''s assumptions before you rely on it.','Vậy hãy gọi đó là dự phóng, và soi kỹ các giả định của mô hình trước khi dựa vào nó.',103),

    -- C1 · Sắc thái trang trọng
    ('sac-thai-trang-trong','Tom','I cannot come tonight, hence I am informing you now.','Tối nay mình không đến được, do đó mình thông báo với bạn ngay bây giờ.',100),
    ('sac-thai-trang-trong','Mai','Why are you talking like a legal document?','Sao bạn nói chuyện như văn bản pháp luật vậy?',101),
    ('sac-thai-trang-trong','Tom','I have been writing contracts all day.','Cả ngày nay mình viết hợp đồng mà.',102),
    ('sac-thai-trang-trong','Mai','With friends, just say ''I can''t come tonight, so I''m letting you know.''','Với bạn bè thì cứ nói ''I can''t come tonight, so I''m letting you know'' là được.',103),

    -- C2 · Biện pháp tu từ
    ('bien-phap-tu-tu','Anna','The news said the company is ''letting some staff go''. Is that a euphemism?','Bản tin nói công ty đang ''cho một số nhân viên nghỉ''. Đó có phải uyển ngữ không ạ?',100),
    ('bien-phap-tu-tu','Giảng viên','Yes. It means they are firing people, but it sounds almost kind.','Đúng. Nghĩa là họ đang sa thải, nhưng nghe gần như tử tế.',101),
    ('bien-phap-tu-tu','Anna','So the phrase decides how we feel before we have thought about it.','Vậy cụm từ đó định sẵn cảm xúc của mình trước khi mình kịp nghĩ.',102),
    ('bien-phap-tu-tu','Giảng viên','Exactly. That is why noticing it matters more than using it.','Chính xác. Vì vậy nhận ra nó quan trọng hơn dùng nó.',103),

    -- C2 · Mở đầu ấn tượng
    ('mo-dau-an-tuong','Linh','I am opening my talk with a shocking statistic about plastic.','Mình định mở đầu bài nói bằng một con số gây sốc về nhựa.',100),
    ('mo-dau-an-tuong','David','Everyone has heard shocking statistics. Why not start with a question they cannot answer yet?','Ai cũng nghe con số gây sốc rồi. Sao không mở đầu bằng một câu hỏi mà họ chưa trả lời được?',101),
    ('mo-dau-an-tuong','Linh','Such as?','Ví dụ?',102),
    ('mo-dau-an-tuong','David','''Where did the last toothbrush you threw away go?'' Then they have to wait for your answer.','''Chiếc bàn chải bạn vứt đi gần nhất đã đi đâu?'' Khi đó họ phải chờ câu trả lời của bạn.',103),

    -- C2 · Dựng lập luận
    ('dung-lap-luan','An','I think we should cut meetings to two a week.','Mình nghĩ nên giảm họp xuống còn hai buổi một tuần.',100),
    ('dung-lap-luan','Linh','So you want nobody to talk to each other?','Vậy là bạn muốn chẳng ai nói chuyện với ai nữa à?',101),
    ('dung-lap-luan','An','That is not what I said. That is a straw man.','Mình đâu có nói vậy. Đó là nguỵ biện người rơm.',102),
    ('dung-lap-luan','Linh','Fair. Then substantiate it: what do we lose if we cut them?','Công bằng. Vậy bạn chứng minh đi: cắt họp thì mình mất gì?',103),

    -- C2 · Nhịp điệu câu văn
    ('nhip-dieu-cau-van','Anna','My editor says the ending of my essay is ponderous.','Biên tập viên nói đoạn kết bài luận của em nặng nề.',100),
    ('nhip-dieu-cau-van','Giảng viên','Read me the last sentence.','Đọc cho tôi nghe câu cuối.',101),
    ('nhip-dieu-cau-van','Anna','''In consideration of all the aforementioned factors, the conclusion must necessarily be that the plan failed.''','''Xét tất cả các yếu tố đã nêu trên, kết luận tất yếu phải là kế hoạch đã thất bại.''',102),
    ('nhip-dieu-cau-van','Giảng viên','After three long sentences, try this: ''The plan failed.'' The contrast does the work.','Sau ba câu dài, thử thế này: ''Kế hoạch đã thất bại.'' Chính sự tương phản tạo nên sức nặng.',103),

    -- C2 · Vốn từ tinh tế
    ('tu-vung-tinh-te','Hà','The article says the minister visited the factory ''ostensibly to inspect safety''.','Bài báo viết bộ trưởng tới thăm nhà máy ''bề ngoài là để kiểm tra an toàn''.',100),
    ('tu-vung-tinh-te','David','Notice ''ostensibly''. The writer does not believe that reason.','Để ý chữ ''ostensibly''. Người viết không tin vào lý do đó.',101),
    ('tu-vung-tinh-te','Hà','So what was the real reason?','Vậy lý do thật là gì?',102),
    ('tu-vung-tinh-te','David','The article never says. It only makes an oblique reference to next month''s election.','Bài báo không nói thẳng. Nó chỉ nhắc gián tiếp tới cuộc bầu cử tháng sau.',103),

    -- C2 · Kéo hướng người nghe
    ('keo-huong-nguoi-nghe','Linh','I speak faster when I get nervous on stage.','Mình nói nhanh hơn mỗi khi hồi hộp trên sân khấu.',100),
    ('keo-huong-nguoi-nghe','David','Try the opposite. After your key point, stop for three seconds.','Thử làm ngược lại. Sau ý chính, dừng ba giây.',101),
    ('keo-huong-nguoi-nghe','Linh','Three seconds feels like forever.','Ba giây mà cảm giác dài như vô tận.',102),
    ('keo-huong-nguoi-nghe','David','For you. For the audience, it is time to think. What they work out themselves, they remember.','Với bạn thôi. Với người nghe, đó là lúc để nghĩ. Điều họ tự nghĩ ra, họ sẽ nhớ.',103),

    -- C2 · Văn phong báo chí
    ('van-phong-bao-chi','Anna','This article says ''sources say'' the deal is off.','Bài này viết ''các nguồn tin cho biết'' thương vụ đã đổ vỡ.',100),
    ('van-phong-bao-chi','Giảng viên','Which sources? How many? Are they independent of each other?','Nguồn nào? Bao nhiêu nguồn? Có độc lập với nhau không?',101),
    ('van-phong-bao-chi','Anna','It does not say.','Bài không nói ạ.',102),
    ('van-phong-bao-chi','Giảng viên','Then wait for a second paper to confirm it before you share it.','Vậy hãy chờ một tờ báo thứ hai xác nhận rồi hẵng chia sẻ.',103),

    -- C2 · Kết bài đọng lại
    ('ket-bai-dong-lai','Linh','My conclusion repeats all five points of the talk.','Phần kết của mình nhắc lại cả năm ý của bài nói.',100),
    ('ket-bai-dong-lai','David','Nobody will remember five points. Which one do you want them to take home?','Chẳng ai nhớ nổi năm ý đâu. Bạn muốn họ mang về ý nào?',101),
    ('ket-bai-dong-lai','Linh','Probably the story about my grandmother''s garden.','Có lẽ là câu chuyện về khu vườn của bà mình.',102),
    ('ket-bai-dong-lai','David','Then end with that image, and stop.','Vậy kết bằng hình ảnh đó, rồi dừng.',103)
) AS v(lesson_slug, speaker, text_en, text_vi, position)
JOIN lessons AS l ON l.slug = v.lesson_slug AND l.deleted_at IS NULL;
