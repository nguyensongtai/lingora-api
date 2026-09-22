-- Từ vựng của giáo trình mẫu, bậc CEFR lấy từ dữ liệu công bố chứ không phải
-- phán đoán.
--
-- NGUỒN
--   A1–B2: The CEFR-J Wordlist Version 1.5, Yukio Tono, Tokyo University of
--          Foreign Studies. http://www.cefr-j.org/download.html
--          Dùng được cho cả nghiên cứu lẫn thương mại, kèm trích dẫn.
--   C1–C2: Octanove Vocabulary Profile C1/C2 version 1.0, Octanove Labs,
--          giấy phép Creative Commons Attribution-ShareAlike 4.0.
--   Cả hai phân phối qua https://github.com/openlanguageprofiles/olp-en-cefrj
--
-- Mỗi từ ở đây đã được đối chiếu bằng máy với hai danh sách trên: từ nằm trong
-- bài của khoá bậc X thì danh sách phải nói nó đúng là bậc X.
--
-- PHẦN CHƯA CÓ NGUỒN: phiên âm, nghĩa tiếng Việt và câu ví dụ do Claude soạn.
-- Chúng chưa được người có chuyên môn rà soát.
--
-- File này THAY THẾ toàn bộ từ vựng của những bài nó phụ trách, nên chạy lại
-- là hội tụ chứ không cộng dồn. Kèm theo đó, lịch ôn của người học với những
-- từ bị thay cũng mất (khoá ngoại ON DELETE CASCADE) — chấp nhận được với dữ
-- liệu dev, nhưng đừng chạy file này lên database thật.

DELETE FROM vocabulary_entries v
USING lessons l, courses c
WHERE v.lesson_id = l.id
  AND l.course_id = c.id
  AND c.slug IN (
      'ngu-phap-co-ban', 'tu-vung-doi-song', 'giao-tiep-hang-ngay',
      'qua-khu-va-tuong-lai', 'ke-chuyen-va-mo-ta', 'email-va-tin-nhan',
      'tranh-luan-va-y-kien', 'tieng-anh-cong-so', 'hoc-thuat-va-nghien-cuu',
      'sac-thai-va-thanh-ngu', 'van-phong-va-tu-tu', 'dien-thuyet-va-thuyet-phuc'
  );

INSERT INTO vocabulary_entries (lesson_id, word, ipa, meaning, example, example_vi, position)
SELECT l.id, v.word, v.ipa, v.meaning, v.example, v.example_vi, v.position
FROM (VALUES
    -- ===== A1 =====
    ('thi-hien-tai-don','always','/ˈɔːlweɪz/','luôn luôn','She always drinks tea after lunch.','Cô ấy luôn uống trà sau bữa trưa.',0),
    ('thi-hien-tai-don','usually','/ˈjuːʒuəli/','thường thường','We usually walk to school.','Chúng tôi thường đi bộ đến trường.',1),
    ('thi-hien-tai-don','never','/ˈnevər/','không bao giờ','He never forgets her birthday.','Anh ấy không bao giờ quên sinh nhật cô ấy.',2),
    ('thi-hien-tai-don','often','/ˈɔːfn/','thường xuyên','They often eat out on Friday.','Họ thường xuyên ăn ngoài vào thứ Sáu.',3),
    ('thi-hien-tai-don','weekend','/ˈwiːkend/','cuối tuần','I visit my parents every weekend.','Tôi về thăm bố mẹ mỗi cuối tuần.',4),
    ('thi-hien-tai-don','together','/təˈɡeðər/','cùng nhau','My family eats dinner together.','Gia đình tôi ăn tối cùng nhau.',5),

    ('danh-tu-dem-duoc','bottle','/ˈbɑːtl/','cái chai','There is one bottle of water on the table.','Có một chai nước trên bàn.',0),
    ('danh-tu-dem-duoc','sugar','/ˈʃʊɡər/','đường','I do not take sugar in my coffee.','Tôi không cho đường vào cà phê.',1),
    ('danh-tu-dem-duoc','bread','/bred/','bánh mì','We buy bread every morning.','Chúng tôi mua bánh mì mỗi sáng.',2),
    ('danh-tu-dem-duoc','rice','/raɪs/','cơm, gạo','She cooks rice for dinner.','Cô ấy nấu cơm cho bữa tối.',3),
    ('danh-tu-dem-duoc','glass','/ɡlæs/','cái ly','Pour me a glass of milk.','Rót cho tôi một ly sữa.',4),
    ('danh-tu-dem-duoc','piece','/piːs/','miếng, mẩu','He ate one piece of cake.','Anh ấy ăn một miếng bánh.',5),

    ('mao-tu-a-an-the','hour','/ˈaʊər/','giờ, tiếng đồng hồ','We waited for an hour.','Chúng tôi đã chờ một tiếng.',0),
    ('mao-tu-a-an-the','umbrella','/ʌmˈbrelə/','cái ô, cái dù','Take an umbrella with you.','Mang theo một cái ô nhé.',1),
    ('mao-tu-a-an-the','answer','/ˈænsər/','câu trả lời','She gave an answer right away.','Cô ấy đưa ra câu trả lời ngay.',2),
    ('mao-tu-a-an-the','ticket','/ˈtɪkɪt/','vé','I bought a ticket for the show.','Tôi đã mua một vé xem buổi diễn.',3),
    ('mao-tu-a-an-the','idea','/aɪˈdiːə/','ý tưởng','That is a good idea.','Đó là một ý tưởng hay.',4),
    ('mao-tu-a-an-the','office','/ˈɔːfɪs/','văn phòng','She works in an office downtown.','Cô ấy làm ở một văn phòng trong trung tâm.',5),

    ('dong-tu-to-be','tired','/ˈtaɪərd/','mệt','I am tired after the trip.','Tôi mệt sau chuyến đi.',0),
    ('dong-tu-to-be','hungry','/ˈhʌŋɡri/','đói','The children are hungry.','Bọn trẻ đang đói.',1),
    ('dong-tu-to-be','ready','/ˈredi/','sẵn sàng','We are ready to start.','Chúng tôi đã sẵn sàng bắt đầu.',2),
    ('dong-tu-to-be','busy','/ˈbɪzi/','bận','She is busy this afternoon.','Chiều nay cô ấy bận.',3),
    ('dong-tu-to-be','afraid','/əˈfreɪd/','sợ','They are afraid of dogs.','Họ sợ chó.',4),
    ('dong-tu-to-be','angry','/ˈæŋɡri/','tức giận','He was angry about the delay.','Anh ấy tức giận vì bị chậm trễ.',5),

    ('gia-dinh-va-ban-be','parent','/ˈperənt/','bố hoặc mẹ','Every parent worries about this.','Bố mẹ nào cũng lo chuyện này.',0),
    ('gia-dinh-va-ban-be','cousin','/ˈkʌzn/','anh chị em họ','My cousin is the same age as me.','Anh họ tôi bằng tuổi tôi.',1),
    ('gia-dinh-va-ban-be','neighbour','/ˈneɪbər/','hàng xóm','Our neighbour has a big garden.','Hàng xóm của chúng tôi có một khu vườn lớn.',2),
    ('gia-dinh-va-ban-be','grandmother','/ˈɡrænmʌðər/','bà','My grandmother tells wonderful stories.','Bà tôi kể những câu chuyện rất hay.',3),
    ('gia-dinh-va-ban-be','uncle','/ˈʌŋkl/','chú, bác, cậu','My uncle lives in Da Nang.','Chú tôi sống ở Đà Nẵng.',4),
    ('gia-dinh-va-ban-be','husband','/ˈhʌzbənd/','chồng','Her husband is a teacher.','Chồng cô ấy là giáo viên.',5),

    ('mau-sac-va-hinh-dang','bright','/braɪt/','sáng, rực rỡ','She wore a bright yellow dress.','Cô ấy mặc một chiếc váy vàng rực rỡ.',0),
    ('mau-sac-va-hinh-dang','dark','/dɑːrk/','tối, sẫm màu','He likes dark blue shirts.','Anh ấy thích áo sơ mi màu xanh sẫm.',1),
    ('mau-sac-va-hinh-dang','thick','/θɪk/','dày','The book has a thick cover.','Cuốn sách có bìa dày.',2),
    ('mau-sac-va-hinh-dang','long','/lɔːŋ/','dài','She has long hair.','Cô ấy có mái tóc dài.',3),
    ('mau-sac-va-hinh-dang','short','/ʃɔːrt/','ngắn','He drew a short line.','Anh ấy vẽ một đường ngắn.',4),
    ('mau-sac-va-hinh-dang','colour','/ˈkʌlər/','màu sắc','What colour is your bag?','Túi của bạn màu gì?',5),

    ('do-an-va-thuc-uong','vegetable','/ˈvedʒtəbl/','rau củ','Eat one vegetable with every meal.','Hãy ăn một loại rau củ trong mỗi bữa.',0),
    ('do-an-va-thuc-uong','delicious','/dɪˈlɪʃəs/','ngon','This soup is delicious.','Món canh này rất ngon.',1),
    ('do-an-va-thuc-uong','dinner','/ˈdɪnər/','bữa tối','We have dinner at seven.','Chúng tôi ăn tối lúc bảy giờ.',2),
    ('do-an-va-thuc-uong','fruit','/fruːt/','trái cây','She buys fruit at the market.','Cô ấy mua trái cây ở chợ.',3),
    ('do-an-va-thuc-uong','meat','/miːt/','thịt','He does not eat meat.','Anh ấy không ăn thịt.',4),
    ('do-an-va-thuc-uong','juice','/dʒuːs/','nước ép','I drink orange juice every morning.','Tôi uống nước ép cam mỗi sáng.',5),

    ('thoi-gian-va-ngay-thang','tomorrow','/təˈmɑːroʊ/','ngày mai','The test is tomorrow.','Bài kiểm tra vào ngày mai.',0),
    ('thoi-gian-va-ngay-thang','yesterday','/ˈjestərdeɪ/','hôm qua','It rained all day yesterday.','Hôm qua trời mưa cả ngày.',1),
    ('thoi-gian-va-ngay-thang','month','/mʌnθ/','tháng','We moved here last month.','Chúng tôi chuyển đến đây tháng trước.',2),
    ('thoi-gian-va-ngay-thang','early','/ˈɜːrli/','sớm','She arrived early.','Cô ấy đến sớm.',3),
    ('thoi-gian-va-ngay-thang','evening','/ˈiːvnɪŋ/','buổi tối','I study in the evening.','Tôi học vào buổi tối.',4),
    ('thoi-gian-va-ngay-thang','minute','/ˈmɪnɪt/','phút','Wait one minute, please.','Xin chờ một phút.',5),

    -- ===== A2 =====
    ('chao-hoi-lam-quen','polite','/pəˈlaɪt/','lịch sự','It is polite to greet the host first.','Chào chủ nhà trước là phép lịch sự.',0),
    ('chao-hoi-lam-quen','appearance','/əˈpɪrəns/','vẻ ngoài','Do not judge people by appearance.','Đừng đánh giá người khác qua vẻ ngoài.',1),
    ('chao-hoi-lam-quen','anyway','/ˈeniweɪ/','dù sao thì','Anyway, how have you been?','Dù sao thì dạo này bạn thế nào?',2),
    ('chao-hoi-lam-quen','certainly','/ˈsɜːrtnli/','chắc chắn rồi','I will certainly pass on your message.','Tôi chắc chắn sẽ chuyển lời của bạn.',3),
    ('chao-hoi-lam-quen','apologise','/əˈpɑːlədʒaɪz/','xin lỗi','I apologise for being late.','Tôi xin lỗi vì đến muộn.',4),
    ('chao-hoi-lam-quen','attitude','/ˈætɪtuːd/','thái độ','She has a friendly attitude.','Cô ấy có thái độ thân thiện.',5),

    ('goi-mon-o-nha-hang','menu','/ˈmenjuː/','thực đơn','Could we see the menu, please?','Cho chúng tôi xem thực đơn được không?',0),
    ('goi-mon-o-nha-hang','bill','/bɪl/','hoá đơn','Could we have the bill, please?','Cho chúng tôi xin hoá đơn.',1),
    ('goi-mon-o-nha-hang','dessert','/dɪˈzɜːrt/','món tráng miệng','We ordered one dessert to share.','Chúng tôi gọi một món tráng miệng ăn chung.',2),
    ('goi-mon-o-nha-hang','serve','/sɜːrv/','phục vụ','They serve breakfast until ten.','Họ phục vụ bữa sáng tới mười giờ.',3),
    ('goi-mon-o-nha-hang','chef','/ʃef/','đầu bếp','The chef recommends the fish.','Đầu bếp gợi ý món cá.',4),
    ('goi-mon-o-nha-hang','bargain','/ˈbɑːrɡən/','món hời','That set lunch is a real bargain.','Suất trưa đó đúng là món hời.',5),

    ('hoi-duong-di','opposite','/ˈɑːpəzɪt/','đối diện','The pharmacy is opposite the market.','Hiệu thuốc đối diện chợ.',0),
    ('hoi-duong-di','across','/əˈkrɔːs/','băng qua','Walk across the bridge.','Đi băng qua cây cầu.',1),
    ('hoi-duong-di','ahead','/əˈhed/','phía trước','The station is just ahead.','Nhà ga ngay phía trước.',2),
    ('hoi-duong-di','beyond','/bɪˈjɑːnd/','quá, bên kia','The school is beyond the park.','Trường học ở bên kia công viên.',3),
    ('hoi-duong-di','centre','/ˈsentər/','trung tâm','The hotel is in the centre of town.','Khách sạn nằm ở trung tâm thị trấn.',4),
    ('hoi-duong-di','direction','/dəˈrekʃn/','hướng','You are going in the wrong direction.','Bạn đang đi sai hướng.',5),

    ('mua-sam','receipt','/rɪˈsiːt/','biên lai','Keep the receipt for the warranty.','Giữ biên lai để bảo hành.',0),
    ('mua-sam','cheap','/tʃiːp/','rẻ','This shop is cheap but the service is good.','Cửa hàng này rẻ mà phục vụ tốt.',1),
    ('mua-sam','cash','/kæʃ/','tiền mặt','I would like to pay in cash.','Tôi muốn trả bằng tiền mặt.',2),
    ('mua-sam','customer','/ˈkʌstəmər/','khách hàng','Every customer gets a free sample.','Mỗi khách hàng được một mẫu thử miễn phí.',3),
    ('mua-sam','budget','/ˈbʌdʒɪt/','ngân sách','That bag is outside my budget.','Cái túi đó vượt ngân sách của tôi.',4),
    ('mua-sam','brand','/brænd/','thương hiệu','She only buys one brand of tea.','Cô ấy chỉ mua trà của một thương hiệu.',5),

    ('thi-qua-khu-don','decide','/dɪˈsaɪd/','quyết định','They could not decide where to stay.','Họ không quyết định được sẽ ở đâu.',0),
    ('thi-qua-khu-don','invite','/ɪnˈvaɪt/','mời','He wants to invite us to dinner.','Anh ấy muốn mời chúng tôi ăn tối.',1),
    ('thi-qua-khu-don','return','/rɪˈtɜːrn/','trở về','We did not return home until midnight.','Mãi tới nửa đêm chúng tôi mới trở về nhà.',2),
    ('thi-qua-khu-don','complain','/kəmˈpleɪn/','phàn nàn','She did not complain about the noise.','Cô ấy không phàn nàn về tiếng ồn.',3),
    ('thi-qua-khu-don','admit','/ədˈmɪt/','thừa nhận','He would not admit his mistake.','Anh ấy không chịu thừa nhận lỗi của mình.',4),
    ('thi-qua-khu-don','accept','/əkˈsept/','chấp nhận','They will accept our offer.','Họ sẽ chấp nhận lời đề nghị của chúng tôi.',5),

    ('dong-tu-bat-quy-tac','bite','/baɪt/','cắn','The dog will not bite you.','Con chó sẽ không cắn bạn đâu.',0),
    ('dong-tu-bat-quy-tac','bend','/bend/','cúi, uốn cong','Bend your knees slowly.','Từ từ khuỵu đầu gối xuống.',1),
    ('dong-tu-bat-quy-tac','burn','/bɜːrn/','cháy, làm cháy','Do not burn the rice.','Đừng làm cháy cơm.',2),
    ('dong-tu-bat-quy-tac','beg','/beɡ/','van xin','He had to beg for more time.','Anh ấy phải van xin thêm thời gian.',3),
    ('dong-tu-bat-quy-tac','blame','/bleɪm/','đổ lỗi','Do not blame the weather.','Đừng đổ lỗi cho thời tiết.',4),
    ('dong-tu-bat-quy-tac','bury','/ˈberi/','chôn, vùi','They bury the seeds in spring.','Họ vùi hạt giống vào mùa xuân.',5),

    ('noi-ve-du-dinh','probably','/ˈprɑːbəbli/','có lẽ','She will probably call tonight.','Có lẽ tối nay cô ấy sẽ gọi.',0),
    ('noi-ve-du-dinh','prepare','/prɪˈper/','chuẩn bị','We prepare the report every Friday.','Chúng tôi chuẩn bị báo cáo vào mỗi thứ Sáu.',1),
    ('noi-ve-du-dinh','ambition','/æmˈbɪʃn/','hoài bão','Her ambition is to study abroad.','Hoài bão của cô ấy là đi du học.',2),
    ('noi-ve-du-dinh','chance','/tʃæns/','cơ hội','There is a good chance of rain.','Khả năng mưa là khá cao.',3),
    ('noi-ve-du-dinh','choice','/tʃɔɪs/','lựa chọn','You have to make a choice today.','Hôm nay bạn phải đưa ra lựa chọn.',4),
    ('noi-ve-du-dinh','continue','/kənˈtɪnjuː/','tiếp tục','We will continue after lunch.','Chúng ta sẽ tiếp tục sau bữa trưa.',5),

    ('ke-ve-ky-nghi','abroad','/əˈbrɔːd/','ở nước ngoài','They travelled abroad last summer.','Họ đi nước ngoài hè năm ngoái.',0),
    ('ke-ve-ky-nghi','crowded','/ˈkraʊdɪd/','đông đúc','The beach was crowded.','Bãi biển rất đông.',1),
    ('ke-ve-ky-nghi','relax','/rɪˈlæks/','thư giãn','We went there to relax.','Chúng tôi đến đó để thư giãn.',2),
    ('ke-ve-ky-nghi','adventure','/ədˈventʃər/','cuộc phiêu lưu','The trip turned into an adventure.','Chuyến đi thành một cuộc phiêu lưu.',3),
    ('ke-ve-ky-nghi','castle','/ˈkæsl/','lâu đài','We visited an old castle.','Chúng tôi thăm một lâu đài cổ.',4),
    ('ke-ve-ky-nghi','coast','/koʊst/','bờ biển','We drove along the coast.','Chúng tôi lái xe dọc bờ biển.',5)
) AS v(lesson_slug, word, ipa, meaning, example, example_vi, position)
JOIN lessons AS l ON l.slug = v.lesson_slug AND l.deleted_at IS NULL;

INSERT INTO vocabulary_entries (lesson_id, word, ipa, meaning, example, example_vi, position)
SELECT l.id, v.word, v.ipa, v.meaning, v.example, v.example_vi, v.position
FROM (VALUES
    -- ===== B1 =====
    ('mo-ta-nguoi','generous','/ˈdʒenərəs/','hào phóng','He is generous with his time.','Anh ấy rất hào phóng về thời gian.',0),
    ('mo-ta-nguoi','stubborn','/ˈstʌbərn/','bướng bỉnh','My brother can be stubborn.','Em trai tôi đôi khi rất bướng.',1),
    ('mo-ta-nguoi','reliable','/rɪˈlaɪəbl/','đáng tin cậy','He is the most reliable person here.','Anh ấy là người đáng tin cậy nhất ở đây.',2),
    ('mo-ta-nguoi','cheerful','/ˈtʃɪrfl/','vui vẻ','She has a cheerful voice.','Cô ấy có giọng nói vui vẻ.',3),
    ('mo-ta-nguoi','honest','/ˈɑːnɪst/','trung thực','Give me an honest answer.','Cho tôi một câu trả lời trung thực.',4),
    ('mo-ta-nguoi','ambitious','/æmˈbɪʃəs/','đầy tham vọng','She is ambitious about her career.','Cô ấy đầy tham vọng với sự nghiệp.',5),

    ('mo-ta-noi-chon','atmosphere','/ˈætməsfɪr/','bầu không khí','The café has a warm atmosphere.','Quán cà phê có bầu không khí ấm cúng.',0),
    ('mo-ta-noi-chon','broad','/brɔːd/','rộng','A broad river runs through the city.','Một dòng sông rộng chảy qua thành phố.',1),
    ('mo-ta-noi-chon','basement','/ˈbeɪsmənt/','tầng hầm','The workshop is in the basement.','Xưởng nằm ở tầng hầm.',2),
    ('mo-ta-noi-chon','cabin','/ˈkæbɪn/','nhà gỗ nhỏ','They rented a cabin in the woods.','Họ thuê một căn nhà gỗ trong rừng.',3),
    ('mo-ta-noi-chon','bare','/ber/','trống trơn','The walls were completely bare.','Các bức tường trống trơn.',4),
    ('mo-ta-noi-chon','border','/ˈbɔːrdər/','biên giới','The village sits near the border.','Ngôi làng nằm gần biên giới.',5),

    ('ke-mot-su-viec','suddenly','/ˈsʌdənli/','đột nhiên','Suddenly the lights went out.','Đột nhiên đèn tắt hết.',0),
    ('ke-mot-su-viec','eventually','/ɪˈventʃuəli/','cuối cùng thì','Eventually we found the key.','Cuối cùng chúng tôi cũng tìm thấy chìa khoá.',1),
    ('ke-mot-su-viec','meanwhile','/ˈmiːnwaɪl/','trong khi đó','Meanwhile the others waited outside.','Trong khi đó những người khác chờ bên ngoài.',2),
    ('ke-mot-su-viec','incident','/ˈɪnsɪdənt/','sự việc','Nobody mentioned the incident again.','Không ai nhắc lại sự việc đó nữa.',3),
    ('ke-mot-su-viec','immediately','/ɪˈmiːdiətli/','ngay lập tức','She left immediately after the call.','Cô ấy đi ngay lập tức sau cuộc gọi.',4),
    ('ke-mot-su-viec','afterwards','/ˈæftərwərdz/','sau đó','We talked about it afterwards.','Sau đó chúng tôi có nói về chuyện ấy.',5),

    ('cam-xuc-va-thai-do','annoyed','/əˈnɔɪd/','khó chịu','He looked annoyed by the noise.','Anh ấy có vẻ khó chịu vì tiếng ồn.',0),
    ('cam-xuc-va-thai-do','embarrassed','/ɪmˈberəst/','ngượng','She felt embarrassed about the mistake.','Cô ấy thấy ngượng vì lỗi đó.',1),
    ('cam-xuc-va-thai-do','guilty','/ˈɡɪlti/','áy náy','I feel guilty about forgetting.','Tôi thấy áy náy vì đã quên.',2),
    ('cam-xuc-va-thai-do','anxiety','/æŋˈzaɪəti/','sự lo âu','Exams cause a lot of anxiety.','Các kỳ thi gây ra nhiều lo âu.',3),
    ('cam-xuc-va-thai-do','ashamed','/əˈʃeɪmd/','xấu hổ','He was ashamed of his reaction.','Anh ấy xấu hổ về phản ứng của mình.',4),
    ('cam-xuc-va-thai-do','amazed','/əˈmeɪzd/','kinh ngạc','We were amazed by the view.','Chúng tôi kinh ngạc trước khung cảnh.',5),

    ('email-trang-trong','regarding','/rɪˈɡɑːrdɪŋ/','về việc','I am writing regarding your invoice.','Tôi viết thư về việc hoá đơn của quý vị.',0),
    ('email-trang-trong','attach','/əˈtætʃ/','đính kèm','Please attach the signed form.','Vui lòng đính kèm biểu mẫu đã ký.',1),
    ('email-trang-trong','enquiry','/ɪnˈkwaɪəri/','thư hỏi','Thank you for your enquiry.','Cảm ơn quý vị đã gửi thư hỏi.',2),
    ('email-trang-trong','inform','/ɪnˈfɔːrm/','thông báo','We will inform you by Friday.','Chúng tôi sẽ thông báo cho quý vị trước thứ Sáu.',3),
    ('email-trang-trong','approval','/əˈpruːvl/','sự phê duyệt','The plan needs approval from the board.','Kế hoạch cần sự phê duyệt của hội đồng.',4),
    ('email-trang-trong','application','/ˌæplɪˈkeɪʃn/','đơn đăng ký','Your application has been received.','Đơn đăng ký của bạn đã được tiếp nhận.',5),

    ('email-than-mat','lately','/ˈleɪtli/','dạo này','Have you been busy lately?','Dạo này bạn có bận không?',0),
    ('email-than-mat','update','/ˈʌpdeɪt/','tin cập nhật','Send me an update when you can.','Khi nào rảnh gửi tôi tin cập nhật nhé.',1),
    ('email-than-mat','suppose','/səˈpoʊz/','đoán chừng','I suppose we could meet on Sunday.','Tôi đoán chừng chủ nhật mình gặp được.',2),
    ('email-than-mat','definitely','/ˈdefɪnətli/','chắc chắn','I will definitely be there.','Tôi chắc chắn sẽ có mặt.',3),
    ('email-than-mat','apology','/əˈpɑːlədʒi/','lời xin lỗi','You owe her an apology.','Bạn nợ cô ấy một lời xin lỗi.',4),
    ('email-than-mat','buddy','/ˈbʌdi/','bạn thân','He is an old buddy from school.','Anh ấy là bạn thân hồi đi học.',5),

    ('hen-lich-hop','available','/əˈveɪləbl/','rảnh, có thể','Are you available on Thursday?','Thứ Năm bạn có rảnh không?',0),
    ('hen-lich-hop','agenda','/əˈdʒendə/','chương trình họp','I will send the agenda tonight.','Tối nay tôi sẽ gửi chương trình họp.',1),
    ('hen-lich-hop','confirm','/kənˈfɜːrm/','xác nhận','Please confirm the time.','Vui lòng xác nhận giờ hẹn.',2),
    ('hen-lich-hop','arrange','/əˈreɪndʒ/','sắp xếp','Could you arrange a room for us?','Bạn sắp xếp giúp một phòng được không?',3),
    ('hen-lich-hop','attend','/əˈtend/','tham dự','Six people will attend.','Sáu người sẽ tham dự.',4),
    ('hen-lich-hop','arrangement','/əˈreɪndʒmənt/','sự thu xếp','The arrangement suits everyone.','Cách thu xếp đó hợp với mọi người.',5),

    ('xin-loi-va-cam-on','forgive','/fərˈɡɪv/','tha thứ','Please forgive the short notice.','Mong bạn thứ lỗi vì báo gấp.',0),
    ('xin-loi-va-cam-on','appreciation','/əˌpriːʃiˈeɪʃn/','sự trân trọng','We want to show our appreciation.','Chúng tôi muốn bày tỏ sự trân trọng.',1),
    ('xin-loi-va-cam-on','acknowledge','/əkˈnɑːlɪdʒ/','ghi nhận','Please acknowledge this email.','Vui lòng xác nhận đã nhận email này.',2),
    ('xin-loi-va-cam-on','blessing','/ˈblesɪŋ/','điều may mắn','Your help was a real blessing.','Sự giúp đỡ của bạn thật là may mắn.',3),
    ('xin-loi-va-cam-on','burden','/ˈbɜːrdn/','gánh nặng','I do not want to be a burden.','Tôi không muốn thành gánh nặng.',4),
    ('xin-loi-va-cam-on','behalf','/bɪˈhæf/','thay mặt','I am writing on behalf of the team.','Tôi viết thư thay mặt cả nhóm.',5),

    -- ===== B2 =====
    ('neu-quan-diem','assert','/əˈsɜːrt/','khẳng định','She will assert that the data is sound.','Cô ấy sẽ khẳng định dữ liệu là đáng tin.',0),
    ('neu-quan-diem','assumption','/əˈsʌmpʃn/','giả định','That argument rests on one assumption.','Lập luận đó dựa trên một giả định.',1),
    ('neu-quan-diem','ambiguous','/æmˈbɪɡjuəs/','mơ hồ','The wording is ambiguous.','Cách diễn đạt còn mơ hồ.',2),
    ('neu-quan-diem','controversy','/ˈkɑːntrəvɜːrsi/','tranh cãi','The policy caused a lot of controversy.','Chính sách đó gây nhiều tranh cãi.',3),
    ('neu-quan-diem','crucial','/ˈkruːʃl/','mang tính then chốt','Timing is crucial here.','Thời điểm ở đây là then chốt.',4),
    ('neu-quan-diem','conventional','/kənˈvenʃənl/','theo lối thông thường','This breaks with conventional thinking.','Điều này đi ngược lối nghĩ thông thường.',5),

    ('phan-bien','concede','/kənˈsiːd/','thừa nhận','I concede that the timing was poor.','Tôi thừa nhận là thời điểm không tốt.',0),
    ('phan-bien','contradict','/ˌkɑːntrəˈdɪkt/','mâu thuẫn với','These figures contradict the report.','Những con số này mâu thuẫn với báo cáo.',1),
    ('phan-bien','confront','/kənˈfrʌnt/','đối diện với','We must confront the real cause.','Chúng ta phải đối diện nguyên nhân thật.',2),
    ('phan-bien','criticism','/ˈkrɪtɪsɪzəm/','sự phê bình','His criticism was fair.','Lời phê bình của anh ấy là công bằng.',3),
    ('phan-bien','ambiguity','/ˌæmbɪˈɡjuːəti/','sự mơ hồ','The ambiguity weakens the argument.','Sự mơ hồ làm lập luận yếu đi.',4),
    ('phan-bien','overlook','/ˌoʊvərˈlʊk/','bỏ sót','The study seems to overlook cost.','Nghiên cứu có vẻ bỏ sót yếu tố chi phí.',5),

    ('dan-chung','assess','/əˈses/','đánh giá','We need to assess the risk first.','Trước hết cần đánh giá rủi ro.',0),
    ('dan-chung','attribute','/əˈtrɪbjuːt/','quy cho','They attribute the rise to policy.','Họ quy mức tăng đó cho chính sách.',1),
    ('dan-chung','comprehensive','/ˌkɑːmprɪˈhensɪv/','toàn diện','The report is comprehensive.','Báo cáo mang tính toàn diện.',2),
    ('dan-chung','consistent','/kənˈsɪstənt/','nhất quán','The results are consistent across groups.','Kết quả nhất quán giữa các nhóm.',3),
    ('dan-chung','statistics','/stəˈtɪstɪks/','số liệu thống kê','These statistics come from the census.','Số liệu thống kê này lấy từ cuộc tổng điều tra.',4),
    ('dan-chung','correspond','/ˌkɔːrəˈspɑːnd/','tương ứng','The numbers correspond to last year.','Các con số tương ứng với năm ngoái.',5),

    ('chot-van-de','ultimately','/ˈʌltɪmətli/','rốt cuộc','Ultimately the decision is yours.','Rốt cuộc quyết định là của bạn.',0),
    ('chot-van-de','overall','/ˌoʊvərˈɔːl/','nhìn chung','Overall the project went well.','Nhìn chung dự án diễn ra tốt.',1),
    ('chot-van-de','alter','/ˈɔːltər/','thay đổi','Nothing will alter that conclusion.','Không gì thay đổi được kết luận đó.',2),
    ('chot-van-de','competence','/ˈkɑːmpɪtəns/','năng lực','Nobody doubts her competence.','Không ai nghi ngờ năng lực của cô ấy.',3),
    ('chot-van-de','consent','/kənˈsent/','sự đồng ý','We proceeded with their consent.','Chúng tôi tiến hành với sự đồng ý của họ.',4),
    ('chot-van-de','adequate','/ˈædɪkwət/','đủ mức cần thiết','The budget is adequate for now.','Ngân sách hiện đủ dùng.',5),

    ('hop-hanh','clarify','/ˈklærɪfaɪ/','làm rõ','Could you clarify the deadline?','Bạn làm rõ giúp hạn chót được không?',0),
    ('hop-hanh','consult','/kənˈsʌlt/','tham vấn','We should consult the legal team.','Chúng ta nên tham vấn bộ phận pháp lý.',1),
    ('hop-hanh','consultant','/kənˈsʌltənt/','chuyên gia tư vấn','An outside consultant joined the meeting.','Một chuyên gia tư vấn bên ngoài dự họp.',2),
    ('hop-hanh','coherent','/koʊˈhɪrənt/','mạch lạc','We need one coherent plan.','Chúng ta cần một kế hoạch mạch lạc.',3),
    ('hop-hanh','conference','/ˈkɑːnfərəns/','hội nghị','The conference runs for two days.','Hội nghị kéo dài hai ngày.',4),
    ('hop-hanh','authorise','/ˈɔːθəraɪz/','cho phép chính thức','Only the director can authorise this.','Chỉ giám đốc mới cho phép việc này.',5),

    ('bao-cao-tien-do','assessment','/əˈsesmənt/','bản đánh giá','The assessment is due on Monday.','Bản đánh giá phải nộp vào thứ Hai.',0),
    ('bao-cao-tien-do','anticipate','/ænˈtɪsɪpeɪt/','dự liệu','We did not anticipate this delay.','Chúng tôi không dự liệu được sự chậm trễ này.',1),
    ('bao-cao-tien-do','complexity','/kəmˈpleksəti/','độ phức tạp','The complexity grew month by month.','Độ phức tạp tăng lên theo từng tháng.',2),
    ('bao-cao-tien-do','competent','/ˈkɑːmpɪtənt/','có năng lực','The team is competent and steady.','Nhóm có năng lực và ổn định.',3),
    ('bao-cao-tien-do','analyst','/ˈænəlɪst/','chuyên viên phân tích','Our analyst reviewed the numbers.','Chuyên viên phân tích của chúng tôi đã rà lại các con số.',4),
    ('bao-cao-tien-do','attainment','/əˈteɪnmənt/','mức đạt được','Attainment fell short of the target.','Mức đạt được thấp hơn mục tiêu.',5),

    ('thuong-luong','compensate','/ˈkɑːmpenseɪt/','bù đắp','We will compensate you for the delay.','Chúng tôi sẽ bù đắp cho bạn vì sự chậm trễ.',0),
    ('thuong-luong','compensation','/ˌkɑːmpenˈseɪʃn/','khoản bồi thường','They asked for compensation.','Họ yêu cầu một khoản bồi thường.',1),
    ('thuong-luong','ally','/ˈælaɪ/','đồng minh','He turned out to be a useful ally.','Anh ấy hoá ra là một đồng minh hữu ích.',2),
    ('thuong-luong','alternatively','/ɔːlˈtɜːrnətɪvli/','hoặc là','Alternatively, we could split the cost.','Hoặc là chúng ta chia đôi chi phí.',3),
    ('thuong-luong','considerably','/kənˈsɪdərəbli/','đáng kể','The price dropped considerably.','Giá đã giảm đáng kể.',4),
    ('thuong-luong','adaptable','/əˈdæptəbl/','dễ thích ứng','Their team is adaptable.','Nhóm của họ rất dễ thích ứng.',5),

    ('phan-hoi-dong-nghiep','critically','/ˈkrɪtɪkli/','một cách phê phán','Read the draft critically.','Hãy đọc bản nháp một cách phê phán.',0),
    ('phan-hoi-dong-nghiep','competent','/ˈkɑːmpɪtənt/','có năng lực','She is more than competent.','Cô ấy thừa năng lực.',1),
    ('phan-hoi-dong-nghiep','assert','/əˈsɜːrt/','nêu dứt khoát','Assert your point without attacking.','Hãy nêu ý dứt khoát mà không công kích.',2),
    ('phan-hoi-dong-nghiep','correspondence','/ˌkɔːrəˈspɑːndəns/','thư từ trao đổi','Keep the correspondence on file.','Hãy lưu lại thư từ trao đổi.',3),
    ('phan-hoi-dong-nghiep','contradictory','/ˌkɑːntrəˈdɪktəri/','mâu thuẫn nhau','We received contradictory feedback.','Chúng tôi nhận phản hồi mâu thuẫn nhau.',4),
    ('phan-hoi-dong-nghiep','convinced','/kənˈvɪnst/','tin chắc','I am not convinced by that reason.','Tôi chưa tin chắc vào lý do đó.',5)
) AS v(lesson_slug, word, ipa, meaning, example, example_vi, position)
JOIN lessons AS l ON l.slug = v.lesson_slug AND l.deleted_at IS NULL;

INSERT INTO vocabulary_entries (lesson_id, word, ipa, meaning, example, example_vi, position)
SELECT l.id, v.word, v.ipa, v.meaning, v.example, v.example_vi, v.position
FROM (VALUES
    -- ===== C1 =====
    ('trich-dan-nguon','citation','/saɪˈteɪʃn/','trích dẫn nguồn','Every claim needs a citation.','Mỗi khẳng định đều cần trích dẫn nguồn.',0),
    ('trich-dan-nguon','credible','/ˈkredəbl/','đáng tin','Use credible sources only.','Chỉ dùng những nguồn đáng tin.',1),
    ('trich-dan-nguon','credibility','/ˌkredəˈbɪləti/','độ tin cậy','A single error damages credibility.','Một lỗi duy nhất cũng làm hỏng độ tin cậy.',2),
    ('trich-dan-nguon','discourse','/ˈdɪskɔːrs/','diễn ngôn','The paper enters a long discourse.','Bài báo bước vào một dòng diễn ngôn dài.',3),
    ('trich-dan-nguon','assertion','/əˈsɜːrʃn/','luận điểm khẳng định','That assertion needs support.','Luận điểm đó cần được chống đỡ.',4),
    ('trich-dan-nguon','depiction','/dɪˈpɪkʃn/','sự khắc hoạ','The depiction of the era is accurate.','Sự khắc hoạ thời kỳ đó là chính xác.',5),

    ('mo-ta-du-lieu','negligible','/ˈneɡlɪdʒəbl/','không đáng kể','The difference was negligible.','Khác biệt là không đáng kể.',0),
    ('mo-ta-du-lieu','fraction','/ˈfrækʃn/','một phần nhỏ','Only a fraction of users replied.','Chỉ một phần nhỏ người dùng trả lời.',1),
    ('mo-ta-du-lieu','estimation','/ˌestɪˈmeɪʃn/','sự ước lượng','By our estimation the cost doubles.','Theo ước lượng của chúng tôi, chi phí tăng gấp đôi.',2),
    ('mo-ta-du-lieu','marginal','/ˈmɑːrdʒɪnl/','ở mức biên, rất nhỏ','The gain was marginal.','Mức tăng rất nhỏ.',3),
    ('mo-ta-du-lieu','plummet','/ˈplʌmɪt/','rơi thẳng đứng','Sales plummet every January.','Doanh số rơi thẳng đứng vào mỗi tháng Giêng.',4),
    ('mo-ta-du-lieu','precision','/prɪˈsɪʒn/','độ chính xác','The instrument lacks precision.','Thiết bị thiếu độ chính xác.',5),

    ('gioi-han-nghien-cuu','constraint','/kənˈstreɪnt/','ràng buộc','Time was the main constraint.','Thời gian là ràng buộc chính.',0),
    ('gioi-han-nghien-cuu','drawback','/ˈdrɔːbæk/','nhược điểm','The method has one clear drawback.','Phương pháp có một nhược điểm rõ.',1),
    ('gioi-han-nghien-cuu','speculative','/ˈspekjələtɪv/','mang tính suy đoán','That conclusion is speculative.','Kết luận đó mang tính suy đoán.',2),
    ('gioi-han-nghien-cuu','plausible','/ˈplɔːzəbl/','nghe hợp lý','It is a plausible explanation.','Đó là một lời giải thích nghe hợp lý.',3),
    ('gioi-han-nghien-cuu','obscure','/əbˈskjʊr/','mờ mịt, khó hiểu','The cause remains obscure.','Nguyên nhân vẫn còn mờ mịt.',4),
    ('gioi-han-nghien-cuu','rudimentary','/ˌruːdɪˈmentri/','sơ khai','Our tools were rudimentary.','Công cụ của chúng tôi còn sơ khai.',5),

    ('thuat-ngu-hoc-thuat','criteria','/kraɪˈtɪriə/','các tiêu chí','The criteria are listed in section two.','Các tiêu chí được nêu ở mục hai.',0),
    ('thuat-ngu-hoc-thuat','specimen','/ˈspesɪmən/','mẫu vật','Each specimen was labelled.','Mỗi mẫu vật đều được dán nhãn.',1),
    ('thuat-ngu-hoc-thuat','scrutinise','/ˈskruːtənaɪz/','soi xét kỹ','Reviewers scrutinise every figure.','Người phản biện soi xét kỹ từng con số.',2),
    ('thuat-ngu-hoc-thuat','formulate','/ˈfɔːrmjuleɪt/','xây dựng nên','We formulate the question first.','Trước hết chúng tôi xây dựng câu hỏi.',3),
    ('thuat-ngu-hoc-thuat','speculate','/ˈspekjuleɪt/','phỏng đoán','It is early to speculate.','Còn sớm để phỏng đoán.',4),
    ('thuat-ngu-hoc-thuat','inherent','/ɪnˈhɪrənt/','vốn có','There is an inherent bias here.','Ở đây có một thiên lệch vốn có.',5),

    ('thanh-ngu-thong-dung','tedious','/ˈtiːdiəs/','tẻ nhạt, lê thê','The process is tedious but necessary.','Quy trình tẻ nhạt nhưng cần thiết.',0),
    ('thanh-ngu-thong-dung','daunting','/ˈdɔːntɪŋ/','làm nản lòng','The workload looked daunting.','Khối lượng việc trông thật nản lòng.',1),
    ('thanh-ngu-thong-dung','gruelling','/ˈɡruːəlɪŋ/','vắt kiệt sức','It was a gruelling week.','Đó là một tuần vắt kiệt sức.',2),
    ('thanh-ngu-thong-dung','lacklustre','/ˈlæklʌstər/','nhạt nhoà','The response was lacklustre.','Phản hồi khá nhạt nhoà.',3),
    ('thanh-ngu-thong-dung','staggering','/ˈstæɡərɪŋ/','đến sững sờ','The cost was staggering.','Chi phí lớn đến sững sờ.',4),
    ('thanh-ngu-thong-dung','riveting','/ˈrɪvɪtɪŋ/','cuốn không rời','Her talk was riveting.','Bài nói của cô ấy cuốn không rời.',5),

    ('cum-dong-tu','expedite','/ˈekspədaɪt/','đẩy nhanh','We can expedite the paperwork.','Chúng tôi có thể đẩy nhanh thủ tục.',0),
    ('cum-dong-tu','impede','/ɪmˈpiːd/','cản trở','Poor data will impede the analysis.','Dữ liệu kém sẽ cản trở việc phân tích.',1),
    ('cum-dong-tu','heighten','/ˈhaɪtn/','làm tăng thêm','A long delay will heighten the tension.','Chậm trễ lâu sẽ làm tăng thêm căng thẳng.',2),
    ('cum-dong-tu','salvage','/ˈsælvɪdʒ/','vớt vát','We tried to salvage the project.','Chúng tôi cố vớt vát dự án.',3),
    ('cum-dong-tu','relinquish','/rɪˈlɪŋkwɪʃ/','từ bỏ, nhường lại','He had to relinquish control.','Anh ấy phải nhường lại quyền kiểm soát.',4),
    ('cum-dong-tu','envisage','/ɪnˈvɪzɪdʒ/','hình dung trước','We envisage a longer timeline.','Chúng tôi hình dung một tiến độ dài hơn.',5),

    ('noi-giam-noi-tranh','marginal','/ˈmɑːrdʒɪnl/','không đáng bao nhiêu','The improvement was marginal.','Mức cải thiện không đáng bao nhiêu.',0),
    ('noi-giam-noi-tranh','reluctantly','/rɪˈlʌktəntli/','một cách miễn cưỡng','She reluctantly agreed.','Cô ấy miễn cưỡng đồng ý.',1),
    ('noi-giam-noi-tranh','superficial','/ˌsuːpərˈfɪʃl/','hời hợt','The review felt superficial.','Bản rà soát có vẻ hời hợt.',2),
    ('noi-giam-noi-tranh','obscurely','/əbˈskjʊrli/','một cách mập mờ','The clause is obscurely worded.','Điều khoản được diễn đạt mập mờ.',3),
    ('noi-giam-noi-tranh','timid','/ˈtɪmɪd/','rụt rè','His objection was rather timid.','Lời phản đối của anh ấy khá rụt rè.',4),
    ('noi-giam-noi-tranh','sedate','/sɪˈdeɪt/','điềm đạm','She kept a sedate tone.','Cô ấy giữ giọng điềm đạm.',5),

    ('sac-thai-trang-trong','thereby','/ˌðerˈbaɪ/','qua đó','The clause was removed, thereby ending the dispute.','Điều khoản bị bỏ, qua đó chấm dứt tranh chấp.',0),
    ('sac-thai-trang-trong','hence','/hens/','do đó','The data is thin; hence the caution.','Dữ liệu mỏng, do đó phải dè dặt.',1),
    ('sac-thai-trang-trong','whilst','/waɪlst/','trong khi','Whilst we agree, the timing worries us.','Trong khi chúng tôi đồng ý, thời điểm vẫn đáng lo.',2),
    ('sac-thai-trang-trong','customary','/ˈkʌstəmeri/','theo thông lệ','It is customary to reply in writing.','Theo thông lệ thì trả lời bằng văn bản.',3),
    ('sac-thai-trang-trong','formality','/fɔːrˈmæləti/','thủ tục hình thức','The signature is a formality.','Chữ ký chỉ là thủ tục hình thức.',4),
    ('sac-thai-trang-trong','oblige','/əˈblaɪdʒ/','buộc phải','The rules oblige us to report this.','Quy định buộc chúng tôi phải báo cáo việc này.',5),

    -- ===== C2 =====
    ('bien-phap-tu-tu','allegory','/ˈæləɡɔːri/','truyện ngụ ý','The novel reads as an allegory.','Cuốn tiểu thuyết đọc như một truyện ngụ ý.',0),
    ('bien-phap-tu-tu','allusion','/əˈluːʒn/','sự ám chỉ','The title is an allusion to a poem.','Nhan đề là một sự ám chỉ tới một bài thơ.',1),
    ('bien-phap-tu-tu','antithesis','/ænˈtɪθəsɪs/','phép đối lập','The line turns on a sharp antithesis.','Câu văn xoay quanh một phép đối lập sắc bén.',2),
    ('bien-phap-tu-tu','euphemism','/ˈjuːfəmɪzəm/','uyển ngữ','Let go is a euphemism for dismissed.','Cho nghỉ là uyển ngữ của bị sa thải.',3),
    ('bien-phap-tu-tu','connotation','/ˌkɑːnəˈteɪʃn/','hàm ý','The word carries a negative connotation.','Từ đó mang hàm ý tiêu cực.',4),
    ('bien-phap-tu-tu','figuratively','/ˈfɪɡjərətɪvli/','theo nghĩa bóng','He meant it figuratively.','Anh ấy nói theo nghĩa bóng.',5),

    ('nhip-dieu-cau-van','brevity','/ˈbrevəti/','sự ngắn gọn','Brevity is the strength of this piece.','Sự ngắn gọn là thế mạnh của bài này.',0),
    ('nhip-dieu-cau-van','lyrical','/ˈlɪrɪkl/','giàu chất thơ','Her prose turns lyrical at the end.','Văn xuôi của bà trở nên giàu chất thơ ở đoạn cuối.',1),
    ('nhip-dieu-cau-van','colloquial','/kəˈloʊkwiəl/','thuộc khẩu ngữ','The dialogue is deliberately colloquial.','Đoạn thoại cố ý dùng khẩu ngữ.',2),
    ('nhip-dieu-cau-van','polysyllabic','/ˌpɑːlisɪˈlæbɪk/','nhiều âm tiết','He avoids polysyllabic words.','Ông ấy tránh những từ nhiều âm tiết.',3),
    ('nhip-dieu-cau-van','syllabic','/sɪˈlæbɪk/','thuộc âm tiết','The metre follows a syllabic pattern.','Nhịp thơ theo một khuôn âm tiết.',4),
    ('nhip-dieu-cau-van','ponderous','/ˈpɑːndərəs/','nặng nề, chậm chạp','The opening chapter is ponderous.','Chương mở đầu khá nặng nề.',5),

    ('tu-vung-tinh-te','ostensibly','/ɑːˈstensəbli/','bề ngoài là','The trip was ostensibly about research.','Chuyến đi bề ngoài là để nghiên cứu.',0),
    ('tu-vung-tinh-te','meticulous','/məˈtɪkjələs/','tỉ mỉ','She keeps meticulous notes.','Cô ấy ghi chép rất tỉ mỉ.',1),
    ('tu-vung-tinh-te','ephemeral','/ɪˈfemərəl/','phù du','Online fame is ephemeral.','Danh tiếng trên mạng là thứ phù du.',2),
    ('tu-vung-tinh-te','nonchalant','/ˌnɑːnʃəˈlɑːnt/','thản nhiên','He gave a nonchalant shrug.','Anh ấy nhún vai thản nhiên.',3),
    ('tu-vung-tinh-te','oblique','/əˈbliːk/','vòng vo, gián tiếp','She made an oblique reference to it.','Cô ấy nhắc tới nó một cách gián tiếp.',4),
    ('tu-vung-tinh-te','tacit','/ˈtæsɪt/','ngầm hiểu','There was a tacit agreement.','Đã có một thoả thuận ngầm.',5),

    ('van-phong-bao-chi','rhetoric','/ˈretərɪk/','lối hùng biện','The speech was pure rhetoric.','Bài phát biểu thuần là lối hùng biện.',0),
    ('van-phong-bao-chi','disseminate','/dɪˈsemɪneɪt/','phát tán rộng','Social media can disseminate a rumour overnight.','Mạng xã hội có thể phát tán một tin đồn chỉ sau một đêm.',1),
    ('van-phong-bao-chi','denounce','/dɪˈnaʊns/','lên án','Two papers will denounce the ruling.','Hai tờ báo sẽ lên án phán quyết.',2),
    ('van-phong-bao-chi','notoriety','/ˌnoʊtəˈraɪəti/','tiếng xấu','The case brought him notoriety.','Vụ việc mang lại tiếng xấu cho ông ta.',3),
    ('van-phong-bao-chi','spurious','/ˈspjʊriəs/','nguỵ tạo','The quote turned out to be spurious.','Câu trích hoá ra là nguỵ tạo.',4),
    ('van-phong-bao-chi','impeccable','/ɪmˈpekəbl/','không chê vào đâu được','Her sourcing is impeccable.','Cách dẫn nguồn của cô ấy không chê vào đâu được.',5),

    ('mo-dau-an-tuong','gambit','/ˈɡæmbɪt/','nước đi mở màn','Opening with a question is a common gambit.','Mở đầu bằng câu hỏi là nước đi quen thuộc.',0),
    ('mo-dau-an-tuong','proposition','/ˌprɑːpəˈzɪʃn/','mệnh đề, luận điểm','State your proposition in one sentence.','Hãy nêu luận điểm trong một câu.',1),
    ('mo-dau-an-tuong','posit','/ˈpɑːzɪt/','đặt ra giả định','Let us posit two possible causes.','Hãy đặt ra hai nguyên nhân có thể.',2),
    ('mo-dau-an-tuong','provocation','/ˌprɑːvəˈkeɪʃn/','sự khiêu khích','The opening line is a mild provocation.','Câu mở đầu là một sự khiêu khích nhẹ.',3),
    ('mo-dau-an-tuong','hitherto','/ˌhɪðərˈtuː/','cho tới nay','Hitherto nobody had asked that.','Cho tới nay chưa ai hỏi điều đó.',4),
    ('mo-dau-an-tuong','paramount','/ˈpærəmaʊnt/','tối quan trọng','Clarity is paramount in the first minute.','Sự rõ ràng là tối quan trọng trong phút đầu.',5),

    ('dung-lap-luan','fallacy','/ˈfæləsi/','nguỵ biện','That is a common logical fallacy.','Đó là một nguỵ biện logic thường gặp.',0),
    ('dung-lap-luan','cogent','/ˈkoʊdʒənt/','chặt chẽ, thuyết phục','She made a cogent case for delay.','Cô ấy lập luận chặt chẽ cho việc hoãn lại.',1),
    ('dung-lap-luan','substantiate','/səbˈstænʃieɪt/','chứng minh bằng chứng cứ','You must substantiate that figure.','Bạn phải chứng minh con số đó.',2),
    ('dung-lap-luan','hypothesis','/haɪˈpɑːθəsɪs/','giả thuyết','The hypothesis was not supported.','Giả thuyết không được ủng hộ.',3),
    ('dung-lap-luan','conjecture','/kənˈdʒektʃər/','sự phỏng đoán','Without data it is pure conjecture.','Không có dữ liệu thì chỉ là phỏng đoán.',4),
    ('dung-lap-luan','axiom','/ˈæksiəm/','tiên đề','The proof rests on one axiom.','Chứng minh dựa trên một tiên đề.',5),

    ('keo-huong-nguoi-nghe','rhetorical','/rɪˈtɔːrɪkl/','có tính tu từ','He ended on a rhetorical question.','Anh ấy kết bằng một câu hỏi tu từ.',0),
    ('keo-huong-nguoi-nghe','compellingly','/kəmˈpelɪŋli/','một cách cuốn hút','She argued compellingly.','Cô ấy lập luận một cách cuốn hút.',1),
    ('keo-huong-nguoi-nghe','mesmeric','/mezˈmerɪk/','như thôi miên','His delivery is mesmeric.','Cách trình bày của anh ấy như thôi miên.',2),
    ('keo-huong-nguoi-nghe','emulate','/ˈemjuleɪt/','noi theo','Younger speakers emulate her pacing.','Những người nói trẻ noi theo nhịp của cô ấy.',3),
    ('keo-huong-nguoi-nghe','induce','/ɪnˈduːs/','làm nảy sinh','A pause can induce real attention.','Một khoảng lặng có thể làm nảy sinh sự chú ý thật.',4),
    ('keo-huong-nguoi-nghe','tenacious','/təˈneɪʃəs/','kiên trì bám chặt','He is tenacious in debate.','Anh ấy kiên trì bám chặt khi tranh luận.',5),

    ('ket-bai-dong-lai','reconcile','/ˈrekənsaɪl/','dung hoà','The ending tries to reconcile both views.','Đoạn kết cố dung hoà cả hai quan điểm.',0),
    ('ket-bai-dong-lai','enduringly','/ɪnˈdʊrɪŋli/','một cách bền bỉ','The image stays enduringly in mind.','Hình ảnh đó đọng lại bền bỉ trong tâm trí.',1),
    ('ket-bai-dong-lai','wistful','/ˈwɪstfl/','bâng khuâng','The final line is wistful.','Câu cuối mang vẻ bâng khuâng.',2),
    ('ket-bai-dong-lai','rapture','/ˈræptʃər/','niềm hân hoan','The crowd listened in rapture.','Đám đông lắng nghe trong niềm hân hoan.',3),
    ('ket-bai-dong-lai','fruition','/fruˈɪʃn/','sự thành tựu','The plan came to fruition this year.','Kế hoạch thành tựu trong năm nay.',4),
    ('ket-bai-dong-lai','paramount','/ˈpærəmaʊnt/','tối quan trọng','The last impression is paramount.','Ấn tượng cuối cùng là tối quan trọng.',5)
) AS v(lesson_slug, word, ipa, meaning, example, example_vi, position)
JOIN lessons AS l ON l.slug = v.lesson_slug AND l.deleted_at IS NULL;
