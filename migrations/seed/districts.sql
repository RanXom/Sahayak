-- District reference data.
--
-- NOTE: The complete LGD district directory (~780 districts) is intended to be
-- imported from the official CSV via `cmd/seed --districts-csv`. This file seeds
-- a representative set for pilot states so the location API is usable out of the
-- box. District IDs use a deterministic synthetic scheme (state_id * 1000 + n)
-- until the full LGD import is run, which upserts by (state_id, name_en).
INSERT INTO districts (id, state_id, name_en, name_hi) VALUES
    -- Rajasthan (8)
    (8001, 8, 'Jaipur', 'जयपुर'),
    (8002, 8, 'Jodhpur', 'जोधपुर'),
    (8003, 8, 'Kota', 'कोटा'),
    (8004, 8, 'Udaipur', 'उदयपुर'),
    (8005, 8, 'Ajmer', 'अजमेर'),
    (8006, 8, 'Bikaner', 'बीकानेर'),
    (8007, 8, 'Alwar', 'अलवर'),
    (8008, 8, 'Bharatpur', 'भरतपुर'),
    (8009, 8, 'Sikar', 'सीकर'),
    (8010, 8, 'Bhilwara', 'भीलवाड़ा'),
    (8011, 8, 'Pali', 'पाली'),
    (8012, 8, 'Jhunjhunu', 'झुंझुनू'),
    (8013, 8, 'Nagaur', 'नागौर'),
    (8014, 8, 'Barmer', 'बाड़मेर'),
    (8015, 8, 'Churu', 'चूरू'),
    (8016, 8, 'Dausa', 'दौसा'),
    (8017, 8, 'Tonk', 'टोंक'),
    (8018, 8, 'Banswara', 'बांसवाड़ा'),
    (8019, 8, 'Chittorgarh', 'चित्तौड़गढ़'),
    (8020, 8, 'Dungarpur', 'डूंगरपुर'),
    (8021, 8, 'Jaisalmer', 'जैसलमेर'),
    (8022, 8, 'Jalore', 'जालोर'),
    (8023, 8, 'Jhalawar', 'झालावाड़'),
    (8024, 8, 'Karauli', 'करौली'),
    (8025, 8, 'Rajsamand', 'राजसमंद'),
    (8026, 8, 'Sawai Madhopur', 'सवाई माधोपुर'),
    (8027, 8, 'Sirohi', 'सिरोही'),
    (8028, 8, 'Sri Ganganagar', 'श्रीगंगानगर'),
    (8029, 8, 'Hanumangarh', 'हनुमानगढ़'),
    (8030, 8, 'Dholpur', 'धौलपुर'),
    (8031, 8, 'Baran', 'बारां'),
    (8032, 8, 'Pratapgarh', 'प्रतापगढ़'),
    (8033, 8, 'Bundi', 'बूंदी'),

    -- Tamil Nadu (33)
    (33001, 33, 'Chennai', 'चेन्नई'),
    (33002, 33, 'Coimbatore', 'कोयंबटूर'),
    (33003, 33, 'Madurai', 'मदुरै'),
    (33004, 33, 'Tiruchirappalli', 'तिरुचिरापल्ली'),
    (33005, 33, 'Salem', 'सेलम'),
    (33006, 33, 'Tirunelveli', 'तिरुनेलवेली'),
    (33007, 33, 'Tiruppur', 'तिरुपुर'),
    (33008, 33, 'Vellore', 'वेल्लोर'),
    (33009, 33, 'Erode', 'ईरोड'),
    (33010, 33, 'Thoothukkudi', 'तूतीकोरिन'),
    (33011, 33, 'Dindigul', 'डिंडीगुल'),
    (33012, 33, 'Thanjavur', 'तंजावुर'),
    (33013, 33, 'Kanchipuram', 'कांचीपुरम'),
    (33014, 33, 'Cuddalore', 'कुड्डालोर'),
    (33015, 33, 'Karur', 'करूर'),
    (33016, 33, 'Namakkal', 'नामक्कल'),
    (33017, 33, 'Nilgiris', 'नीलगिरी'),
    (33018, 33, 'Dharmapuri', 'धर्मपुरी'),
    (33019, 33, 'Krishnagiri', 'कृष्णगिरी'),
    (33020, 33, 'Villupuram', 'विलुप्पुरम'),

    -- Uttar Pradesh (9)
    (9001, 9, 'Lucknow', 'लखनऊ'),
    (9002, 9, 'Kanpur Nagar', 'कानपुर नगर'),
    (9003, 9, 'Ghaziabad', 'गाजियाबाद'),
    (9004, 9, 'Agra', 'आगरा'),
    (9005, 9, 'Varanasi', 'वाराणसी'),
    (9006, 9, 'Prayagraj', 'प्रयागराज'),
    (9007, 9, 'Meerut', 'मेरठ'),
    (9008, 9, 'Bareilly', 'बरेली'),
    (9009, 9, 'Aligarh', 'अलीगढ़'),
    (9010, 9, 'Moradabad', 'मुरादाबाद'),
    (9011, 9, 'Gorakhpur', 'गोरखपुर'),
    (9012, 9, 'Noida (Gautam Buddha Nagar)', 'नोएडा (गौतम बुद्ध नगर)'),

    -- Maharashtra (27)
    (27001, 27, 'Mumbai', 'मुंबई'),
    (27002, 27, 'Pune', 'पुणे'),
    (27003, 27, 'Nagpur', 'नागपुर'),
    (27004, 27, 'Thane', 'ठाणे'),
    (27005, 27, 'Nashik', 'नासिक'),
    (27006, 27, 'Aurangabad', 'औरंगाबाद'),
    (27007, 27, 'Solapur', 'सोलापुर'),
    (27008, 27, 'Kolhapur', 'कोल्हापुर'),
    (27009, 27, 'Amravati', 'अमरावती'),
    (27010, 27, 'Nanded', 'नांदेड़'),

    -- Karnataka (29)
    (29001, 29, 'Bengaluru Urban', 'बेंगलुरु शहरी'),
    (29002, 29, 'Mysuru', 'मैसूर'),
    (29003, 29, 'Dharwad', 'धारवाड़'),
    (29004, 29, 'Belagavi', 'बेलगावी'),
    (29005, 29, 'Mangaluru (Dakshina Kannada)', 'मंगलुरु (दक्षिण कन्नड़)'),
    (29006, 29, 'Kalaburagi', 'कलबुरगी'),
    (29007, 29, 'Ballari', 'बल्लारी'),
    (29008, 29, 'Vijayapura', 'विजयपुरा'),

    -- Bihar (10)
    (10001, 10, 'Patna', 'पटना'),
    (10002, 10, 'Gaya', 'गया'),
    (10003, 10, 'Bhagalpur', 'भागलपुर'),
    (10004, 10, 'Muzaffarpur', 'मुजफ्फरपुर'),
    (10005, 10, 'Darbhanga', 'दरभंगा'),
    (10006, 10, 'Purnia', 'पूर्णिया'),

    -- Madhya Pradesh (23)
    (23001, 23, 'Bhopal', 'भोपाल'),
    (23002, 23, 'Indore', 'इंदौर'),
    (23003, 23, 'Jabalpur', 'जबलपुर'),
    (23004, 23, 'Gwalior', 'ग्वालियर'),
    (23005, 23, 'Ujjain', 'उज्जैन'),
    (23006, 23, 'Sagar', 'सागर'),

    -- West Bengal (19)
    (19001, 19, 'Kolkata', 'कोलकाता'),
    (19002, 19, 'Howrah', 'हावड़ा'),
    (19003, 19, 'Darjeeling', 'दार्जिलिंग'),
    (19004, 19, 'Murshidabad', 'मुर्शिदाबाद'),
    (19005, 19, 'Nadia', 'नदिया'),

    -- Gujarat (24)
    (24001, 24, 'Ahmedabad', 'अहमदाबाद'),
    (24002, 24, 'Surat', 'सूरत'),
    (24003, 24, 'Vadodara', 'वडोदरा'),
    (24004, 24, 'Rajkot', 'राजकोट'),
    (24005, 24, 'Bhavnagar', 'भावनगर'),

    -- Kerala (32)
    (32001, 32, 'Thiruvananthapuram', 'तिरुवनंतपुरम'),
    (32002, 32, 'Ernakulam', 'एर्नाकुलम'),
    (32003, 32, 'Kozhikode', 'कोझिकोड'),
    (32004, 32, 'Thrissur', 'त्रिशूर'),
    (32005, 32, 'Kollam', 'कोल्लम')
ON CONFLICT (id) DO UPDATE SET
    state_id = EXCLUDED.state_id,
    name_en = EXCLUDED.name_en,
    name_hi = EXCLUDED.name_hi;
