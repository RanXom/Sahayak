-- LGD state/UT reference data (28 states + 8 union territories).
INSERT INTO states (id, code, name_en, name_hi, name_local, is_ut) VALUES
    (28, 'AP', 'Andhra Pradesh',        'आंध्र प्रदेश',           'ఆంధ్రప్రదేశ్',       FALSE),
    (12, 'AR', 'Arunachal Pradesh',     'अरुणाचल प्रदेश',          'अरुणाचल प्रदेश',      FALSE),
    (18, 'AS', 'Assam',                 'असम',                    'অসম',                FALSE),
    (10, 'BR', 'Bihar',                 'बिहार',                   'बिहार',              FALSE),
    (22, 'CG', 'Chhattisgarh',          'छत्तीसगढ़',               'छत्तीसगढ़',          FALSE),
    (30, 'GA', 'Goa',                   'गोवा',                    'गोंय',               FALSE),
    (24, 'GJ', 'Gujarat',               'गुजरात',                  'ગુજરાત',             FALSE),
    ( 6, 'HR', 'Haryana',               'हरियाणा',                 'हरियाणा',            FALSE),
    ( 2, 'HP', 'Himachal Pradesh',      'हिमाचल प्रदेश',           'हिमाचल प्रदेश',       FALSE),
    (20, 'JH', 'Jharkhand',             'झारखंड',                  'झारखंड',             FALSE),
    (29, 'KA', 'Karnataka',             'कर्नाटक',                 'ಕರ್ನಾಟಕ',            FALSE),
    (32, 'KL', 'Kerala',                'केरल',                    'കേരളം',              FALSE),
    (23, 'MP', 'Madhya Pradesh',        'मध्य प्रदेश',             'मध्य प्रदेश',         FALSE),
    (27, 'MH', 'Maharashtra',           'महाराष्ट्र',               'महाराष्ट्र',          FALSE),
    (14, 'MN', 'Manipur',               'मणिपुर',                  'মণিপুর',              FALSE),
    (17, 'ML', 'Meghalaya',             'मेघालय',                  'মেঘালয়',             FALSE),
    (15, 'MZ', 'Mizoram',               'मिज़ोरम',                  'মিজোরাম',             FALSE),
    (13, 'NL', 'Nagaland',              'नागालैंड',                 'নাগাল্যান্ড',         FALSE),
    (21, 'OD', 'Odisha',                'ओडिशा',                   'ଓଡ଼ିଶା',              FALSE),
    ( 3, 'PB', 'Punjab',                'पंजाब',                   'ਪੰਜਾਬ',              FALSE),
    ( 8, 'RJ', 'Rajasthan',             'राजस्थान',                'राजस्थान',            FALSE),
    (11, 'SK', 'Sikkim',                'सिक्किम',                  'सिक्किम',             FALSE),
    (33, 'TN', 'Tamil Nadu',            'तमिल नाडु',               'தமிழ்நாடு',           FALSE),
    (36, 'TS', 'Telangana',             'तेलंगाना',                 'తెలంగాణ',            FALSE),
    (16, 'TR', 'Tripura',               'त्रिपुरा',                 'ত্রিপুরা',            FALSE),
    ( 9, 'UP', 'Uttar Pradesh',         'उत्तर प्रदेश',            'उत्तर प्रदेश',        FALSE),
    ( 5, 'UK', 'Uttarakhand',           'उत्तराखंड',               'उत्तराखंड',           FALSE),
    (19, 'WB', 'West Bengal',           'पश्चिम बंगाल',            'পশ্চিমবঙ্গ',          FALSE),
    (35, 'AN', 'Andaman and Nicobar Islands', 'अंडमान और निकोबार द्वीपसमूह', 'अंडमान और निकोबार द्वीपसमूह', TRUE),
    ( 4, 'CH', 'Chandigarh',            'चंडीगढ़',                 'ਚੰਡੀਗੜ੍ਹ',            TRUE),
    (38, 'DN', 'Dadra and Nagar Haveli and Daman and Diu', 'दादरा और नगर हवेली एवं दमन और दीव', NULL, TRUE),
    ( 7, 'DL', 'Delhi',                 'दिल्ली',                   'दिल्ली',              TRUE),
    ( 1, 'JK', 'Jammu and Kashmir',     'जम्मू और कश्मीर',         'جموں و کشمیر',        TRUE),
    (37, 'LA', 'Ladakh',                'लद्दाख',                   'लद्दाख',              TRUE),
    (31, 'LD', 'Lakshadweep',           'लक्षद्वीप',                'ലക്ഷദ്വീപ്',          TRUE),
    (34, 'PY', 'Puducherry',            'पुदुचेरी',                 'புதுச்சேரி',           TRUE)
ON CONFLICT (id) DO UPDATE SET
    code = EXCLUDED.code,
    name_en = EXCLUDED.name_en,
    name_hi = EXCLUDED.name_hi,
    name_local = EXCLUDED.name_local,
    is_ut = EXCLUDED.is_ut;

-- Default coverage: no state-specific scheme coverage until curated.
INSERT INTO coverage_status (state_id, level, scheme_count, notes)
SELECT id, 'none', 0, 'State scheme coverage is being developed. Central schemes are still available.'
FROM states
ON CONFLICT (state_id) DO NOTHING;
