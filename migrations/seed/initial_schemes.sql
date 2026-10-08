-- ============================================================================
-- Initial curated central schemes.
-- Data is manually curated for MVP; every scheme is marked verified/published.
-- Rules use the expr-lang expression language evaluated by the eligibility engine.
-- ============================================================================

-- Provenance sources
INSERT INTO sources (id, name, source_type, url, description) VALUES
    ('00000000-0000-0000-0000-000000000001', 'MyScheme (National Platform)', 'gov_portal', 'https://www.myscheme.gov.in', 'Government of India scheme discovery platform'),
    ('00000000-0000-0000-0000-000000000002', 'India.gov.in National Portal', 'gov_portal', 'https://www.india.gov.in', 'National Portal of India'),
    ('00000000-0000-0000-0000-000000000003', 'data.gov.in', 'official_api', 'https://data.gov.in', 'Open Government Data Platform India')
ON CONFLICT (id) DO NOTHING;

-- ---------------------------------------------------------------------------
-- Schemes
-- ---------------------------------------------------------------------------
INSERT INTO schemes (
    slug, status, name_en, name_hi, description_en, description_hi,
    benefit_type, benefit_amount, scope, ministry, official_url, application_url,
    source_id, is_active, completeness, verification, verified_at, published_at
) VALUES
    ('ignoaps', 'published', 'Indira Gandhi National Old Age Pension Scheme', 'इंदिरा गांधी राष्ट्रीय वृद्धावस्था पेंशन योजना',
     'Monthly pension for elderly persons belonging to below poverty line households.', 'गरीबी रेखा से नीचे के परिवारों के वृद्ध व्यक्तियों के लिए मासिक पेंशन।',
     'pension', '₹200-₹500/month (central share)', 'central', 'Ministry of Rural Development',
     'https://nsap.nic.in', 'https://nsap.nic.in',
     '00000000-0000-0000-0000-000000000001', TRUE,
     '{"basic_info":true,"eligibility_rules":true,"required_documents":true,"application_process":true,"application_url":true,"benefit_amount":true,"deadline":false}',
     '{"name_verified":true,"eligibility_rules_verified":true,"documents_verified":true,"application_url_verified":true}', NOW(), NOW()),

    ('ignwps', 'published', 'Indira Gandhi National Widow Pension Scheme', 'इंदिरा गांधी राष्ट्रीय विधवा पेंशन योजना',
     'Monthly pension for widows aged 40-79 from below poverty line households.', 'गरीबी रेखा से नीचे के परिवारों की 40-79 वर्ष की विधवाओं के लिए मासिक पेंशन।',
     'pension', '₹300/month (central share)', 'central', 'Ministry of Rural Development',
     'https://nsap.nic.in', 'https://nsap.nic.in',
     '00000000-0000-0000-0000-000000000001', TRUE,
     '{"basic_info":true,"eligibility_rules":true,"required_documents":true,"application_process":true,"application_url":true,"benefit_amount":true,"deadline":false}',
     '{"name_verified":true,"eligibility_rules_verified":true,"documents_verified":true,"application_url_verified":true}', NOW(), NOW()),

    ('igndps', 'published', 'Indira Gandhi National Disability Pension Scheme', 'इंदिरा गांधी राष्ट्रीय विकलांगता पेंशन योजना',
     'Monthly pension for persons with severe or multiple disabilities from BPL households.', 'गरीबी रेखा से नीचे के परिवारों के गंभीर या बहु-विकलांग व्यक्तियों के लिए मासिक पेंशन।',
     'pension', '₹300/month (central share)', 'central', 'Ministry of Rural Development',
     'https://nsap.nic.in', 'https://nsap.nic.in',
     '00000000-0000-0000-0000-000000000001', TRUE,
     '{"basic_info":true,"eligibility_rules":true,"required_documents":true,"application_process":true,"application_url":true,"benefit_amount":true,"deadline":false}',
     '{"name_verified":true,"eligibility_rules_verified":true,"documents_verified":true,"application_url_verified":true}', NOW(), NOW()),

    ('nfbs', 'published', 'National Family Benefit Scheme', 'राष्ट्रीय परिवार लाभ योजना',
     'One-time financial assistance to a BPL household on the death of its primary breadwinner.', 'बीपीएल परिवार के प्राथमिक कमाने वाले की मृत्यु पर एकमुश्त वित्तीय सहायता।',
     'financial', '₹20,000 one-time', 'central', 'Ministry of Rural Development',
     'https://nsap.nic.in', 'https://nsap.nic.in',
     '00000000-0000-0000-0000-000000000001', TRUE,
     '{"basic_info":true,"eligibility_rules":true,"required_documents":true,"application_process":true,"application_url":true,"benefit_amount":true,"deadline":false}',
     '{"name_verified":true,"eligibility_rules_verified":true,"documents_verified":true,"application_url_verified":true}', NOW(), NOW()),

    ('pm-kisan', 'published', 'Pradhan Mantri Kisan Samman Nidhi', 'प्रधानमंत्री किसान सम्मान निधि',
     'Income support of ₹6,000 per year to eligible farmer families in three instalments.', 'पात्र किसान परिवारों को तीन किश्तों में ₹6,000 प्रति वर्ष की आय सहायता।',
     'agriculture', '₹6,000/year', 'central', 'Ministry of Agriculture and Farmers Welfare',
     'https://pmkisan.gov.in', 'https://pmkisan.gov.in',
     '00000000-0000-0000-0000-000000000001', TRUE,
     '{"basic_info":true,"eligibility_rules":true,"required_documents":true,"application_process":true,"application_url":true,"benefit_amount":true,"deadline":false}',
     '{"name_verified":true,"eligibility_rules_verified":true,"documents_verified":true,"application_url_verified":true}', NOW(), NOW()),

    ('pmfby', 'published', 'Pradhan Mantri Fasal Bima Yojana', 'प्रधानमंत्री फसल बीमा योजना',
     'Crop insurance against yield losses due to natural calamities, pests and diseases.', 'प्राकृतिक आपदाओं, कीटों और रोगों से फसल नुकसान के विरुद्ध फसल बीमा।',
     'insurance', 'Crop loss compensation', 'central', 'Ministry of Agriculture and Farmers Welfare',
     'https://pmfby.gov.in', 'https://pmfby.gov.in',
     '00000000-0000-0000-0000-000000000001', TRUE,
     '{"basic_info":true,"eligibility_rules":true,"required_documents":true,"application_process":true,"application_url":true,"benefit_amount":false,"deadline":true}',
     '{"name_verified":true,"eligibility_rules_verified":true,"documents_verified":true,"application_url_verified":true}', NOW(), NOW()),

    ('pmjjby', 'published', 'Pradhan Mantri Jeevan Jyoti Bima Yojana', 'प्रधानमंत्री जीवन ज्योति बीमा योजना',
     'Life insurance cover of ₹2 lakh for bank account holders aged 18-50.', '18-50 वर्ष के बैंक खाताधारकों के लिए ₹2 लाख का जीवन बीमा कवर।',
     'insurance', '₹2,00,000 cover', 'central', 'Ministry of Finance',
     'https://www.jansuraksha.gov.in', 'https://www.jansuraksha.gov.in',
     '00000000-0000-0000-0000-000000000001', TRUE,
     '{"basic_info":true,"eligibility_rules":true,"required_documents":true,"application_process":true,"application_url":true,"benefit_amount":true,"deadline":false}',
     '{"name_verified":true,"eligibility_rules_verified":true,"documents_verified":true,"application_url_verified":true}', NOW(), NOW()),

    ('pmsby', 'published', 'Pradhan Mantri Suraksha Bima Yojana', 'प्रधानमंत्री सुरक्षा बीमा योजना',
     'Accident insurance cover of ₹2 lakh for bank account holders aged 18-70.', '18-70 वर्ष के बैंक खाताधारकों के लिए ₹2 लाख का दुर्घटना बीमा कवर।',
     'insurance', '₹2,00,000 cover', 'central', 'Ministry of Finance',
     'https://www.jansuraksha.gov.in', 'https://www.jansuraksha.gov.in',
     '00000000-0000-0000-0000-000000000001', TRUE,
     '{"basic_info":true,"eligibility_rules":true,"required_documents":true,"application_process":true,"application_url":true,"benefit_amount":true,"deadline":false}',
     '{"name_verified":true,"eligibility_rules_verified":true,"documents_verified":true,"application_url_verified":true}', NOW(), NOW()),

    ('apy', 'published', 'Atal Pension Yojana', 'अटल पेंशन योजना',
     'Guaranteed minimum monthly pension after 60 for subscribers aged 18-40.', '18-40 वर्ष के ग्राहकों के लिए 60 वर्ष के बाद गारंटीकृत न्यूनतम मासिक पेंशन।',
     'pension', '₹1,000-₹5,000/month after 60', 'central', 'Ministry of Finance',
     'https://www.npscra.nsdl.co.in', 'https://www.npscra.nsdl.co.in',
     '00000000-0000-0000-0000-000000000001', TRUE,
     '{"basic_info":true,"eligibility_rules":true,"required_documents":true,"application_process":true,"application_url":true,"benefit_amount":true,"deadline":false}',
     '{"name_verified":true,"eligibility_rules_verified":true,"documents_verified":true,"application_url_verified":true}', NOW(), NOW()),

    ('pmjay', 'published', 'Ayushman Bharat Pradhan Mantri Jan Arogya Yojana', 'आयुष्मान भारत प्रधानमंत्री जन आरोग्य योजना',
     'Health cover of ₹5 lakh per family per year for secondary and tertiary care.', 'द्वितीयक और तृतीयक देखभाल के लिए प्रति परिवार प्रति वर्ष ₹5 लाख का स्वास्थ्य कवर।',
     'health', '₹5,00,000/family/year', 'central', 'Ministry of Health and Family Welfare',
     'https://pmjay.gov.in', 'https://pmjay.gov.in',
     '00000000-0000-0000-0000-000000000001', TRUE,
     '{"basic_info":true,"eligibility_rules":true,"required_documents":true,"application_process":true,"application_url":true,"benefit_amount":true,"deadline":false}',
     '{"name_verified":true,"eligibility_rules_verified":true,"documents_verified":true,"application_url_verified":true}', NOW(), NOW()),

    ('pmay-g', 'published', 'Pradhan Mantri Awas Yojana - Gramin', 'प्रधानमंत्री आवास योजना - ग्रामीण',
     'Financial assistance for construction of a pucca house for rural homeless families.', 'ग्रामीण बेघर परिवारों के लिए पक्का घर बनाने हेतु वित्तीय सहायता।',
     'housing', 'Up to ₹1,20,000', 'central', 'Ministry of Rural Development',
     'https://pmayg.nic.in', 'https://pmayg.nic.in',
     '00000000-0000-0000-0000-000000000001', TRUE,
     '{"basic_info":true,"eligibility_rules":true,"required_documents":true,"application_process":true,"application_url":true,"benefit_amount":true,"deadline":false}',
     '{"name_verified":true,"eligibility_rules_verified":true,"documents_verified":true,"application_url_verified":true}', NOW(), NOW()),

    ('pmay-u', 'published', 'Pradhan Mantri Awas Yojana - Urban', 'प्रधानमंत्री आवास योजना - शहरी',
     'Housing for All mission providing central assistance for urban affordable housing.', 'शहरी किफायती आवास के लिए केंद्रीय सहायता प्रदान करने वाला सबके लिए आवास मिशन।',
     'housing', 'Interest subsidy / central assistance', 'central', 'Ministry of Housing and Urban Affairs',
     'https://pmay-urban.gov.in', 'https://pmay-urban.gov.in',
     '00000000-0000-0000-0000-000000000001', TRUE,
     '{"basic_info":true,"eligibility_rules":true,"required_documents":true,"application_process":true,"application_url":true,"benefit_amount":false,"deadline":false}',
     '{"name_verified":true,"eligibility_rules_verified":true,"documents_verified":true,"application_url_verified":true}', NOW(), NOW()),

    ('mgnrega', 'published', 'Mahatma Gandhi National Rural Employment Guarantee Act', 'महात्मा गांधी राष्ट्रीय ग्रामीण रोजगार गारंटी अधिनियम',
     'Guarantees at least 100 days of wage employment per year to rural households.', 'ग्रामीण परिवारों को प्रति वर्ष कम से कम 100 दिन के मजदूरी रोजगार की गारंटी।',
     'employment', '100 days wage employment', 'central', 'Ministry of Rural Development',
     'https://nrega.nic.in', 'https://nrega.nic.in',
     '00000000-0000-0000-0000-000000000001', TRUE,
     '{"basic_info":true,"eligibility_rules":true,"required_documents":true,"application_process":true,"application_url":true,"benefit_amount":false,"deadline":false}',
     '{"name_verified":true,"eligibility_rules_verified":true,"documents_verified":true,"application_url_verified":true}', NOW(), NOW()),

    ('post-matric-sc', 'published', 'Post-Matric Scholarship for SC Students', 'अनुसूचित जाति के विद्यार्थियों के लिए पोस्ट-मैट्रिक छात्रवृत्ति',
     'Scholarship for SC students studying after class 10.', 'कक्षा 10 के बाद पढ़ने वाले अनुसूचित जाति के विद्यार्थियों के लिए छात्रवृत्ति।',
     'scholarship', 'Tuition and maintenance allowance', 'central', 'Ministry of Social Justice and Empowerment',
     'https://scholarships.gov.in', 'https://scholarships.gov.in',
     '00000000-0000-0000-0000-000000000001', TRUE,
     '{"basic_info":true,"eligibility_rules":true,"required_documents":true,"application_process":true,"application_url":true,"benefit_amount":false,"deadline":true}',
     '{"name_verified":true,"eligibility_rules_verified":true,"documents_verified":true,"application_url_verified":true}', NOW(), NOW()),

    ('post-matric-st', 'published', 'Post-Matric Scholarship for ST Students', 'अनुसूचित जनजाति के विद्यार्थियों के लिए पोस्ट-मैट्रिक छात्रवृत्ति',
     'Scholarship for ST students studying after class 10.', 'कक्षा 10 के बाद पढ़ने वाले अनुसूचित जनजाति के विद्यार्थियों के लिए छात्रवृत्ति।',
     'scholarship', 'Tuition and maintenance allowance', 'central', 'Ministry of Tribal Affairs',
     'https://scholarships.gov.in', 'https://scholarships.gov.in',
     '00000000-0000-0000-0000-000000000001', TRUE,
     '{"basic_info":true,"eligibility_rules":true,"required_documents":true,"application_process":true,"application_url":true,"benefit_amount":false,"deadline":true}',
     '{"name_verified":true,"eligibility_rules_verified":true,"documents_verified":true,"application_url_verified":true}', NOW(), NOW()),

    ('post-matric-obc', 'published', 'Post-Matric Scholarship for OBC Students', 'अन्य पिछड़ा वर्ग के विद्यार्थियों के लिए पोस्ट-मैट्रिक छात्रवृत्ति',
     'Scholarship for OBC students studying after class 10.', 'कक्षा 10 के बाद पढ़ने वाले अन्य पिछड़ा वर्ग के विद्यार्थियों के लिए छात्रवृत्ति।',
     'scholarship', 'Tuition and maintenance allowance', 'central', 'Ministry of Social Justice and Empowerment',
     'https://scholarships.gov.in', 'https://scholarships.gov.in',
     '00000000-0000-0000-0000-000000000001', TRUE,
     '{"basic_info":true,"eligibility_rules":true,"required_documents":true,"application_process":true,"application_url":true,"benefit_amount":false,"deadline":true}',
     '{"name_verified":true,"eligibility_rules_verified":true,"documents_verified":true,"application_url_verified":true}', NOW(), NOW()),

    ('post-matric-minority', 'published', 'Post-Matric Scholarship for Minority Students', 'अल्पसंख्यक विद्यार्थियों के लिए पोस्ट-मैट्रिक छात्रवृत्ति',
     'Scholarship for students from notified minority communities studying after class 10.', 'कक्षा 10 के बाद पढ़ने वाले अधिसूचित अल्पसंख्यक समुदायों के विद्यार्थियों के लिए छात्रवृत्ति।',
     'scholarship', 'Tuition and maintenance allowance', 'central', 'Ministry of Minority Affairs',
     'https://scholarships.gov.in', 'https://scholarships.gov.in',
     '00000000-0000-0000-0000-000000000001', TRUE,
     '{"basic_info":true,"eligibility_rules":true,"required_documents":true,"application_process":true,"application_url":true,"benefit_amount":false,"deadline":true}',
     '{"name_verified":true,"eligibility_rules_verified":true,"documents_verified":true,"application_url_verified":true}', NOW(), NOW()),

    ('pre-matric-sc', 'published', 'Pre-Matric Scholarship for SC Students', 'अनुसूचित जाति के विद्यार्थियों के लिए प्री-मैट्रिक छात्रवृत्ति',
     'Scholarship for SC students in classes 9 and 10.', 'कक्षा 9 और 10 के अनुसूचित जाति के विद्यार्थियों के लिए छात्रवृत्ति।',
     'scholarship', 'Maintenance allowance', 'central', 'Ministry of Social Justice and Empowerment',
     'https://scholarships.gov.in', 'https://scholarships.gov.in',
     '00000000-0000-0000-0000-000000000001', TRUE,
     '{"basic_info":true,"eligibility_rules":true,"required_documents":true,"application_process":true,"application_url":true,"benefit_amount":false,"deadline":true}',
     '{"name_verified":true,"eligibility_rules_verified":true,"documents_verified":true,"application_url_verified":true}', NOW(), NOW()),

    ('pre-matric-st', 'published', 'Pre-Matric Scholarship for ST Students', 'अनुसूचित जनजाति के विद्यार्थियों के लिए प्री-मैट्रिक छात्रवृत्ति',
     'Scholarship for ST students in classes 9 and 10.', 'कक्षा 9 और 10 के अनुसूचित जनजाति के विद्यार्थियों के लिए छात्रवृत्ति।',
     'scholarship', 'Maintenance allowance', 'central', 'Ministry of Tribal Affairs',
     'https://scholarships.gov.in', 'https://scholarships.gov.in',
     '00000000-0000-0000-0000-000000000001', TRUE,
     '{"basic_info":true,"eligibility_rules":true,"required_documents":true,"application_process":true,"application_url":true,"benefit_amount":false,"deadline":true}',
     '{"name_verified":true,"eligibility_rules_verified":true,"documents_verified":true,"application_url_verified":true}', NOW(), NOW()),

    ('ssy', 'published', 'Sukanya Samriddhi Yojana', 'सुकन्या समृद्धि योजना',
     'Small savings scheme for a girl child below 10 years with attractive interest.', '10 वर्ष से कम आयु की बालिका के लिए आकर्षक ब्याज वाली लघु बचत योजना।',
     'women_child', 'Tax-free savings account', 'central', 'Ministry of Finance',
     'https://www.india.gov.in', 'https://www.india.gov.in',
     '00000000-0000-0000-0000-000000000002', TRUE,
     '{"basic_info":true,"eligibility_rules":true,"required_documents":true,"application_process":true,"application_url":true,"benefit_amount":false,"deadline":false}',
     '{"name_verified":true,"eligibility_rules_verified":true,"documents_verified":true,"application_url_verified":true}', NOW(), NOW()),

    ('pmmvy', 'published', 'Pradhan Mantri Matru Vandana Yojana', 'प्रधानमंत्री मातृ वंदना योजना',
     'Cash incentive for pregnant and lactating mothers for the first living child.', 'पहले जीवित बच्चे के लिए गर्भवती और स्तनपान कराने वाली माताओं को नकद प्रोत्साहन।',
     'women_child', '₹5,000 in instalments', 'central', 'Ministry of Women and Child Development',
     'https://pmmvy.wcd.gov.in', 'https://pmmvy.wcd.gov.in',
     '00000000-0000-0000-0000-000000000001', TRUE,
     '{"basic_info":true,"eligibility_rules":true,"required_documents":true,"application_process":true,"application_url":true,"benefit_amount":true,"deadline":false}',
     '{"name_verified":true,"eligibility_rules_verified":true,"documents_verified":true,"application_url_verified":true}', NOW(), NOW()),

    ('jsy', 'published', 'Janani Suraksha Yojana', 'जननी सुरक्षा योजना',
     'Cash assistance to promote institutional delivery among poor pregnant women.', 'गरीब गर्भवती महिलाओं में संस्थागत प्रसव को बढ़ावा देने हेतु नकद सहायता।',
     'health', 'Cash assistance on delivery', 'central', 'Ministry of Health and Family Welfare',
     'https://nhm.gov.in', 'https://nhm.gov.in',
     '00000000-0000-0000-0000-000000000001', TRUE,
     '{"basic_info":true,"eligibility_rules":true,"required_documents":true,"application_process":true,"application_url":true,"benefit_amount":false,"deadline":false}',
     '{"name_verified":true,"eligibility_rules_verified":true,"documents_verified":true,"application_url_verified":true}', NOW(), NOW()),

    ('pmjdy', 'published', 'Pradhan Mantri Jan Dhan Yojana', 'प्रधानमंत्री जन धन योजना',
     'Financial inclusion programme providing zero-balance bank accounts.', 'शून्य-शेष बैंक खाते प्रदान करने वाला वित्तीय समावेशन कार्यक्रम।',
     'financial', 'Zero-balance account with benefits', 'central', 'Ministry of Finance',
     'https://pmjdy.gov.in', 'https://pmjdy.gov.in',
     '00000000-0000-0000-0000-000000000001', TRUE,
     '{"basic_info":true,"eligibility_rules":true,"required_documents":true,"application_process":true,"application_url":true,"benefit_amount":false,"deadline":false}',
     '{"name_verified":true,"eligibility_rules_verified":true,"documents_verified":true,"application_url_verified":true}', NOW(), NOW()),

    ('pmmy', 'published', 'Pradhan Mantri Mudra Yojana', 'प्रधानमंत्री मुद्रा योजना',
     'Collateral-free loans up to ₹10 lakh for micro and small enterprises.', 'सूक्ष्म और लघु उद्यमों के लिए ₹10 लाख तक का बिना जमानत ऋण।',
     'financial', 'Loan up to ₹10,00,000', 'central', 'Ministry of Finance',
     'https://www.mudra.org.in', 'https://www.mudra.org.in',
     '00000000-0000-0000-0000-000000000001', TRUE,
     '{"basic_info":true,"eligibility_rules":true,"required_documents":true,"application_process":true,"application_url":true,"benefit_amount":true,"deadline":false}',
     '{"name_verified":true,"eligibility_rules_verified":true,"documents_verified":true,"application_url_verified":true}', NOW(), NOW()),

    ('standup-india', 'published', 'Stand Up India Scheme', 'स्टैंड अप इंडिया योजना',
     'Bank loans for SC/ST and women entrepreneurs to set up greenfield enterprises.', 'अनुसूचित जाति/जनजाति और महिला उद्यमियों के लिए ग्रीनफील्ड उद्यम स्थापित करने हेतु बैंक ऋण।',
     'financial', '₹10 lakh - ₹1 crore loan', 'central', 'Ministry of Finance',
     'https://www.standupmitra.in', 'https://www.standupmitra.in',
     '00000000-0000-0000-0000-000000000001', TRUE,
     '{"basic_info":true,"eligibility_rules":true,"required_documents":true,"application_process":true,"application_url":true,"benefit_amount":true,"deadline":false}',
     '{"name_verified":true,"eligibility_rules_verified":true,"documents_verified":true,"application_url_verified":true}', NOW(), NOW()),

    ('pmuy', 'published', 'Pradhan Mantri Ujjwala Yojana', 'प्रधानमंत्री उज्ज्वला योजना',
     'Free LPG connections to women from below poverty line households.', 'गरीबी रेखा से नीचे के परिवारों की महिलाओं को मुफ्त एलपीजी कनेक्शन।',
     'other', 'Free LPG connection', 'central', 'Ministry of Petroleum and Natural Gas',
     'https://www.pmuy.gov.in', 'https://www.pmuy.gov.in',
     '00000000-0000-0000-0000-000000000001', TRUE,
     '{"basic_info":true,"eligibility_rules":true,"required_documents":true,"application_process":true,"application_url":true,"benefit_amount":false,"deadline":false}',
     '{"name_verified":true,"eligibility_rules_verified":true,"documents_verified":true,"application_url_verified":true}', NOW(), NOW()),

    ('saubhagya', 'published', 'Pradhan Mantri Sahaj Bijli Har Ghar Yojana (Saubhagya)', 'प्रधानमंत्री सहज बिजली हर घर योजना (सौभाग्य)',
     'Last-mile electricity connections to all willing rural and urban households.', 'सभी इच्छुक ग्रामीण और शहरी परिवारों को अंतिम छोर तक बिजली कनेक्शन।',
     'other', 'Free/subsidised electricity connection', 'central', 'Ministry of Power',
     'https://saubhagya.gov.in', 'https://saubhagya.gov.in',
     '00000000-0000-0000-0000-000000000001', TRUE,
     '{"basic_info":true,"eligibility_rules":true,"required_documents":true,"application_process":true,"application_url":true,"benefit_amount":false,"deadline":false}',
     '{"name_verified":true,"eligibility_rules_verified":true,"documents_verified":true,"application_url_verified":true}', NOW(), NOW()),

    ('pmkvy', 'published', 'Pradhan Mantri Kaushal Vikas Yojana', 'प्रधानमंत्री कौशल विकास योजना',
     'Short-term skill training and certification for unemployed youth.', 'बेरोजगार युवाओं के लिए अल्पकालिक कौशल प्रशिक्षण और प्रमाणन।',
     'employment', 'Free skill training and certification', 'central', 'Ministry of Skill Development and Entrepreneurship',
     'https://www.pmkvyofficial.org', 'https://www.pmkvyofficial.org',
     '00000000-0000-0000-0000-000000000001', TRUE,
     '{"basic_info":true,"eligibility_rules":true,"required_documents":true,"application_process":true,"application_url":true,"benefit_amount":false,"deadline":false}',
     '{"name_verified":true,"eligibility_rules_verified":true,"documents_verified":true,"application_url_verified":true}', NOW(), NOW()),

    ('pm-svanidhi', 'published', 'PM Street Vendor''s AtmaNirbhar Nidhi', 'प्रधानमंत्री स्ट्रीट वेंडर्स आत्मनिर्भर निधि',
     'Working capital loans to street vendors for resuming livelihoods.', 'आजीविका फिर से शुरू करने के लिए रेहड़ी-पटरी विक्रेताओं को कार्यशील पूंजी ऋण।',
     'financial', 'Loan up to ₹50,000', 'central', 'Ministry of Housing and Urban Affairs',
     'https://pmsvanidhi.mohua.gov.in', 'https://pmsvanidhi.mohua.gov.in',
     '00000000-0000-0000-0000-000000000001', TRUE,
     '{"basic_info":true,"eligibility_rules":true,"required_documents":true,"application_process":true,"application_url":true,"benefit_amount":true,"deadline":false}',
     '{"name_verified":true,"eligibility_rules_verified":true,"documents_verified":true,"application_url_verified":true}', NOW(), NOW()),

    ('kcc', 'published', 'Kisan Credit Card', 'किसान क्रेडिट कार्ड',
     'Short-term credit for farmers for crop production and allied activities.', 'फसल उत्पादन और संबद्ध गतिविधियों के लिए किसानों को अल्पकालिक ऋण।',
     'agriculture', 'Credit up to ₹3,00,000 at concessional interest', 'central', 'Ministry of Agriculture and Farmers Welfare',
     'https://www.nabard.org', 'https://www.nabard.org',
     '00000000-0000-0000-0000-000000000001', TRUE,
     '{"basic_info":true,"eligibility_rules":true,"required_documents":true,"application_process":true,"application_url":true,"benefit_amount":true,"deadline":false}',
     '{"name_verified":true,"eligibility_rules_verified":true,"documents_verified":true,"application_url_verified":true}', NOW(), NOW()),

    ('pmmatsya', 'published', 'Pradhan Mantri Matsya Sampada Yojana', 'प्रधानमंत्री मत्स्य संपदा योजना',
     'Support for sustainable development of fisheries and fisherfolk welfare.', 'मत्स्य पालन के सतत विकास और मछुआरों के कल्याण हेतु सहायता।',
     'agriculture', 'Subsidy for fishing equipment and infrastructure', 'central', 'Ministry of Fisheries, Animal Husbandry and Dairying',
     'https://pmmsy.dof.gov.in', 'https://pmmsy.dof.gov.in',
     '00000000-0000-0000-0000-000000000001', TRUE,
     '{"basic_info":true,"eligibility_rules":true,"required_documents":true,"application_process":true,"application_url":true,"benefit_amount":false,"deadline":false}',
     '{"name_verified":true,"eligibility_rules_verified":true,"documents_verified":true,"application_url_verified":true}', NOW(), NOW()),

    ('scst-obc-post-matric-topclass', 'published', 'Top Class Education Scheme for SC Students', 'अनुसूचित जाति विद्यार्थियों के लिए टॉप क्लास शिक्षा योजना',
     'Full scholarship for SC students admitted to premier institutions.', 'प्रमुख संस्थानों में प्रवेशित अनुसूचित जाति विद्यार्थियों के लिए पूर्ण छात्रवृत्ति।',
     'scholarship', 'Full tuition and living expenses', 'central', 'Ministry of Social Justice and Empowerment',
     'https://scholarships.gov.in', 'https://scholarships.gov.in',
     '00000000-0000-0000-0000-000000000001', TRUE,
     '{"basic_info":true,"eligibility_rules":true,"required_documents":true,"application_process":true,"application_url":true,"benefit_amount":false,"deadline":true}',
     '{"name_verified":true,"eligibility_rules_verified":true,"documents_verified":true,"application_url_verified":true}', NOW(), NOW())
ON CONFLICT (slug) DO NOTHING;

-- ---------------------------------------------------------------------------
-- Eligibility rules
-- ---------------------------------------------------------------------------
INSERT INTO scheme_eligibility_rules (scheme_id, rule_name, expression, description_en, description_hi, is_mandatory, sort_order)
SELECT id, v.rule_name, v.expression, v.des_en, v.des_hi, v.mandatory, v.ord
FROM schemes s
JOIN (VALUES
    ('ignoaps', 'Minimum age 60', 'age >= 60', 'Applicant must be 60 years or older', 'आवेदक की आयु 60 वर्ष या अधिक होनी चाहिए', TRUE, 0),
    ('ignoaps', 'Below poverty line', 'ration_card_type in ["bpl", "aay"]', 'Household must hold a BPL or Antyodaya ration card', 'परिवार के पास बीपीएल या अंत्योदय राशन कार्ड होना चाहिए', TRUE, 1),

    ('ignwps', 'Widow', 'marital_status == "widowed"', 'Applicant must be a widow', 'आवेदक विधवा होनी चाहिए', TRUE, 0),
    ('ignwps', 'Age 40 to 79', 'age >= 40 && age <= 79', 'Applicant must be between 40 and 79 years', 'आवेदक की आयु 40 से 79 वर्ष के बीच होनी चाहिए', TRUE, 1),
    ('ignwps', 'Below poverty line', 'ration_card_type in ["bpl", "aay"]', 'Household must hold a BPL or Antyodaya ration card', 'परिवार के पास बीपीएल या अंत्योदय राशन कार्ड होना चाहिए', TRUE, 2),

    ('igndps', 'Severe disability', 'disability_status in ["severe", "multiple"]', 'Applicant must have severe or multiple disability', 'आवेदक को गंभीर या बहु-विकलांगता होनी चाहिए', TRUE, 0),
    ('igndps', 'Age 18 to 79', 'age >= 18 && age <= 79', 'Applicant must be between 18 and 79 years', 'आवेदक की आयु 18 से 79 वर्ष के बीच होनी चाहिए', TRUE, 1),
    ('igndps', 'Below poverty line', 'ration_card_type in ["bpl", "aay"]', 'Household must hold a BPL or Antyodaya ration card', 'परिवार के पास बीपीएल या अंत्योदय राशन कार्ड होना चाहिए', TRUE, 2),

    ('nfbs', 'BPL household', 'ration_card_type in ["bpl", "aay"]', 'Household must be below the poverty line', 'परिवार गरीबी रेखा से नीचे होना चाहिए', TRUE, 0),
    ('nfbs', 'Breadwinner aged 18-59', 'age >= 18 && age <= 59', 'Deceased breadwinner was between 18 and 59 years', 'मृत कमाने वाला 18 से 59 वर्ष के बीच था', TRUE, 1),

    ('pm-kisan', 'Farmer household', 'occupation == "farmer"', 'Applicant must be a farmer', 'आवेदक किसान होना चाहिए', TRUE, 0),
    ('pm-kisan', 'Land holding limit', 'land_acres <= 5.0', 'Land holding must not exceed 5 acres', 'भूमि जोत 5 एकड़ से अधिक नहीं होनी चाहिए', TRUE, 1),

    ('pmfby', 'Farmer', 'occupation == "farmer"', 'Applicant must be a farmer', 'आवेदक किसान होना चाहिए', TRUE, 0),

    ('pmjjby', 'Age 18 to 50', 'age >= 18 && age <= 50', 'Applicant must be between 18 and 50 years', 'आवेदक की आयु 18 से 50 वर्ष के बीच होनी चाहिए', TRUE, 0),
    ('pmjjby', 'Bank account', 'has_bank_account == true', 'Applicant must have a bank account', 'आवेदक का बैंक खाता होना चाहिए', TRUE, 1),

    ('pmsby', 'Age 18 to 70', 'age >= 18 && age <= 70', 'Applicant must be between 18 and 70 years', 'आवेदक की आयु 18 से 70 वर्ष के बीच होनी चाहिए', TRUE, 0),
    ('pmsby', 'Bank account', 'has_bank_account == true', 'Applicant must have a bank account', 'आवेदक का बैंक खाता होना चाहिए', TRUE, 1),

    ('apy', 'Age 18 to 40', 'age >= 18 && age <= 40', 'Applicant must be between 18 and 40 years', 'आवेदक की आयु 18 से 40 वर्ष के बीच होनी चाहिए', TRUE, 0),
    ('apy', 'Bank account', 'has_bank_account == true', 'Applicant must have a savings bank account', 'आवेदक का बचत बैंक खाता होना चाहिए', TRUE, 1),

    ('pmjay', 'BPL or Antyodaya', 'ration_card_type in ["bpl", "aay"]', 'Household must be identified as deprived/BPL', 'परिवार को वंचित/बीपीएल के रूप में चिह्नित होना चाहिए', TRUE, 0),

    ('pmay-g', 'Rural residence', 'residence_type == "rural"', 'Applicant must reside in a rural area', 'आवेदक ग्रामीण क्षेत्र में रहता हो', TRUE, 0),
    ('pmay-g', 'No pucca house', 'has_pucca_house == false', 'Household must not own a pucca house', 'परिवार के पास पक्का घर नहीं होना चाहिए', TRUE, 1),
    ('pmay-g', 'Income limit', 'income_annual <= 300000', 'Annual household income must not exceed ₹3,00,000', 'वार्षिक घरेलू आय ₹3,00,000 से अधिक नहीं होनी चाहिए', TRUE, 2),

    ('pmay-u', 'Urban residence', 'residence_type == "urban"', 'Applicant must reside in an urban area', 'आवेदक शहरी क्षेत्र में रहता हो', TRUE, 0),
    ('pmay-u', 'Income limit', 'income_annual <= 600000', 'Annual household income must not exceed ₹6,00,000', 'वार्षिक घरेलू आय ₹6,00,000 से अधिक नहीं होनी चाहिए', TRUE, 1),

    ('mgnrega', 'Rural residence', 'residence_type == "rural"', 'Applicant must reside in a rural area', 'आवेदक ग्रामीण क्षेत्र में रहता हो', TRUE, 0),
    ('mgnrega', 'Adult', 'age >= 18', 'Applicant must be at least 18 years old', 'आवेदक की आयु कम से कम 18 वर्ष होनी चाहिए', TRUE, 1),

    ('post-matric-sc', 'Scheduled Caste', 'caste_category == "sc"', 'Applicant must belong to a Scheduled Caste', 'आवेदक अनुसूचित जाति से होना चाहिए', TRUE, 0),
    ('post-matric-sc', 'Post-matric student', 'has_student_child == true && student_child_class >= 11', 'A student in the family is studying after class 10', 'परिवार में कोई विद्यार्थी कक्षा 10 के बाद पढ़ रहा हो', TRUE, 1),
    ('post-matric-sc', 'Income limit', 'income_annual <= 250000', 'Annual family income must not exceed ₹2,50,000', 'वार्षिक पारिवारिक आय ₹2,50,000 से अधिक नहीं होनी चाहिए', TRUE, 2),

    ('post-matric-st', 'Scheduled Tribe', 'caste_category == "st"', 'Applicant must belong to a Scheduled Tribe', 'आवेदक अनुसूचित जनजाति से होना चाहिए', TRUE, 0),
    ('post-matric-st', 'Post-matric student', 'has_student_child == true && student_child_class >= 11', 'A student in the family is studying after class 10', 'परिवार में कोई विद्यार्थी कक्षा 10 के बाद पढ़ रहा हो', TRUE, 1),
    ('post-matric-st', 'Income limit', 'income_annual <= 250000', 'Annual family income must not exceed ₹2,50,000', 'वार्षिक पारिवारिक आय ₹2,50,000 से अधिक नहीं होनी चाहिए', TRUE, 2),

    ('post-matric-obc', 'Other Backward Class', 'caste_category == "obc"', 'Applicant must belong to an Other Backward Class', 'आवेदक अन्य पिछड़ा वर्ग से होना चाहिए', TRUE, 0),
    ('post-matric-obc', 'Post-matric student', 'has_student_child == true && student_child_class >= 11', 'A student in the family is studying after class 10', 'परिवार में कोई विद्यार्थी कक्षा 10 के बाद पढ़ रहा हो', TRUE, 1),
    ('post-matric-obc', 'Income limit', 'income_annual <= 250000', 'Annual family income must not exceed ₹2,50,000', 'वार्षिक पारिवारिक आय ₹2,50,000 से अधिक नहीं होनी चाहिए', TRUE, 2),

    ('post-matric-minority', 'Minority community', 'minority_status == true', 'Applicant must belong to a notified minority community', 'आवेदक अधिसूचित अल्पसंख्यक समुदाय से होना चाहिए', TRUE, 0),
    ('post-matric-minority', 'Post-matric student', 'has_student_child == true && student_child_class >= 11', 'A student in the family is studying after class 10', 'परिवार में कोई विद्यार्थी कक्षा 10 के बाद पढ़ रहा हो', TRUE, 1),
    ('post-matric-minority', 'Income limit', 'income_annual <= 200000', 'Annual family income must not exceed ₹2,00,000', 'वार्षिक पारिवारिक आय ₹2,00,000 से अधिक नहीं होनी चाहिए', TRUE, 2),

    ('pre-matric-sc', 'Scheduled Caste', 'caste_category == "sc"', 'Applicant must belong to a Scheduled Caste', 'आवेदक अनुसूचित जाति से होना चाहिए', TRUE, 0),
    ('pre-matric-sc', 'Student in class 9-10', 'has_student_child == true && student_child_class >= 9 && student_child_class <= 10', 'A student in the family is in class 9 or 10', 'परिवार में कोई विद्यार्थी कक्षा 9 या 10 में हो', TRUE, 1),
    ('pre-matric-sc', 'Income limit', 'income_annual <= 250000', 'Annual family income must not exceed ₹2,50,000', 'वार्षिक पारिवारिक आय ₹2,50,000 से अधिक नहीं होनी चाहिए', TRUE, 2),

    ('pre-matric-st', 'Scheduled Tribe', 'caste_category == "st"', 'Applicant must belong to a Scheduled Tribe', 'आवेदक अनुसूचित जनजाति से होना चाहिए', TRUE, 0),
    ('pre-matric-st', 'Student in class 9-10', 'has_student_child == true && student_child_class >= 9 && student_child_class <= 10', 'A student in the family is in class 9 or 10', 'परिवार में कोई विद्यार्थी कक्षा 9 या 10 में हो', TRUE, 1),
    ('pre-matric-st', 'Income limit', 'income_annual <= 250000', 'Annual family income must not exceed ₹2,50,000', 'वार्षिक पारिवारिक आय ₹2,50,000 से अधिक नहीं होनी चाहिए', TRUE, 2),

    ('ssy', 'Has a girl child below 10', 'has_girl_child == true && girl_child_age < 10', 'Family has a girl child below 10 years of age', 'परिवार में 10 वर्ष से कम आयु की बालिका हो', TRUE, 0),

    ('pmmvy', 'Pregnant or lactating mother', 'is_pregnant_or_lactating == true', 'Applicant is a pregnant or lactating mother', 'आवेदक गर्भवती या स्तनपान कराने वाली माता हो', TRUE, 0),
    ('pmmvy', 'Age 19 or above', 'age >= 19', 'Applicant must be at least 19 years old', 'आवेदक की आयु कम से कम 19 वर्ष होनी चाहिए', TRUE, 1),

    ('jsy', 'Pregnant woman', 'is_pregnant_or_lactating == true', 'Applicant is a pregnant woman', 'आवेदक गर्भवती महिला हो', TRUE, 0),
    ('jsy', 'BPL household', 'ration_card_type in ["bpl", "aay"]', 'Household must be below the poverty line', 'परिवार गरीबी रेखा से नीचे होना चाहिए', TRUE, 1),

    ('pmjdy', 'No existing bank account', 'has_bank_account == false', 'Applicant does not have a bank account', 'आवेदक का बैंक खाता नहीं है', TRUE, 0),
    ('pmjdy', 'Adult', 'age >= 10', 'Applicant must be at least 10 years old', 'आवेदक की आयु कम से कम 10 वर्ष होनी चाहिए', TRUE, 1),

    ('pmmy', 'Non-farm enterprise', 'occupation != "farmer"', 'Applicant runs a non-farm micro or small enterprise', 'आवेदक गैर-कृषि सूक्ष्म या लघु उद्यम चलाता है', TRUE, 0),
    ('pmmy', 'Adult', 'age >= 18', 'Applicant must be at least 18 years old', 'आवेदक की आयु कम से कम 18 वर्ष होनी चाहिए', TRUE, 1),

    ('standup-india', 'SC/ST or woman', 'caste_category in ["sc", "st"] || gender == "female"', 'Applicant must be SC/ST or a woman entrepreneur', 'आवेदक अनुसूचित जाति/जनजाति या महिला उद्यमी हो', TRUE, 0),
    ('standup-india', 'Adult', 'age >= 18', 'Applicant must be at least 18 years old', 'आवेदक की आयु कम से कम 18 वर्ष होनी चाहिए', TRUE, 1),

    ('pmuy', 'Woman applicant', 'gender == "female"', 'Applicant must be a woman', 'आवेदक महिला होनी चाहिए', TRUE, 0),
    ('pmuy', 'BPL household', 'ration_card_type in ["bpl", "aay"]', 'Household must be below the poverty line', 'परिवार गरीबी रेखा से नीचे होना चाहिए', TRUE, 1),

    ('saubhagya', 'No electricity connection', 'has_electricity == false', 'Household does not have an electricity connection', 'परिवार के पास बिजली कनेक्शन नहीं है', TRUE, 0),

    ('pmkvy', 'Unemployed youth', 'employment_status == "unemployed"', 'Applicant is unemployed and seeking skills', 'आवेदक बेरोजगार है और कौशल चाहता है', TRUE, 0),
    ('pmkvy', 'Age 15 to 45', 'age >= 15 && age <= 45', 'Applicant must be between 15 and 45 years', 'आवेदक की आयु 15 से 45 वर्ष के बीच होनी चाहिए', TRUE, 1),

    ('pm-svanidhi', 'Street vendor', 'occupation == "street_vendor"', 'Applicant must be a street vendor', 'आवेदक रेहड़ी-पटरी विक्रेता होना चाहिए', TRUE, 0),
    ('pm-svanidhi', 'Adult', 'age >= 18', 'Applicant must be at least 18 years old', 'आवेदक की आयु कम से कम 18 वर्ष होनी चाहिए', TRUE, 1),

    ('kcc', 'Farmer', 'occupation == "farmer"', 'Applicant must be a farmer', 'आवेदक किसान होना चाहिए', TRUE, 0),
    ('kcc', 'Adult', 'age >= 18', 'Applicant must be at least 18 years old', 'आवेदक की आयु कम से कम 18 वर्ष होनी चाहिए', TRUE, 1),

    ('pmmatsya', 'Fisher or fish farmer', 'occupation in ["fisher", "fish_farmer"]', 'Applicant must be a fisher or fish farmer', 'आवेदक मछुआरा या मत्स्य किसान होना चाहिए', TRUE, 0),

    ('scst-obc-post-matric-topclass', 'Scheduled Caste', 'caste_category == "sc"', 'Applicant must belong to a Scheduled Caste', 'आवेदक अनुसूचित जाति से होना चाहिए', TRUE, 0),
    ('scst-obc-post-matric-topclass', 'Premier institution student', 'has_student_child == true', 'A student in the family is enrolled in a premier institution', 'परिवार में कोई विद्यार्थी प्रमुख संस्थान में नामांकित हो', TRUE, 1)
) AS v(slug, rule_name, expression, des_en, des_hi, mandatory, ord)
    ON s.slug = v.slug
ON CONFLICT (scheme_id, rule_name) DO UPDATE SET
    expression = EXCLUDED.expression,
    description_en = EXCLUDED.description_en,
    description_hi = EXCLUDED.description_hi,
    is_mandatory = EXCLUDED.is_mandatory,
    sort_order = EXCLUDED.sort_order;

-- ---------------------------------------------------------------------------
-- Required documents
-- ---------------------------------------------------------------------------
INSERT INTO scheme_required_documents (scheme_id, document_type, document_name_en, document_name_hi, is_mandatory, sort_order)
SELECT id, v.document_type, v.name_en, v.name_hi, v.mandatory, v.ord
FROM schemes s
JOIN (VALUES
    ('ignoaps', 'aadhaar', 'Aadhaar Card', 'आधार कार्ड', TRUE, 0),
    ('ignoaps', 'ration_card', 'Ration Card', 'राशन कार्ड', TRUE, 1),
    ('ignoaps', 'bank_account', 'Bank Account', 'बैंक खाता', TRUE, 2),

    ('ignwps', 'aadhaar', 'Aadhaar Card', 'आधार कार्ड', TRUE, 0),
    ('ignwps', 'ration_card', 'Ration Card', 'राशन कार्ड', TRUE, 1),
    ('ignwps', 'death_cert', 'Death Certificate of Spouse', 'पति का मृत्यु प्रमाण पत्र', TRUE, 2),
    ('ignwps', 'bank_account', 'Bank Account', 'बैंक खाता', TRUE, 3),

    ('igndps', 'aadhaar', 'Aadhaar Card', 'आधार कार्ड', TRUE, 0),
    ('igndps', 'disability_cert', 'Disability Certificate', 'विकलांगता प्रमाण पत्र', TRUE, 1),
    ('igndps', 'ration_card', 'Ration Card', 'राशन कार्ड', TRUE, 2),
    ('igndps', 'bank_account', 'Bank Account', 'बैंक खाता', TRUE, 3),

    ('nfbs', 'aadhaar', 'Aadhaar Card', 'आधार कार्ड', TRUE, 0),
    ('nfbs', 'death_cert', 'Death Certificate', 'मृत्यु प्रमाण पत्र', TRUE, 1),
    ('nfbs', 'ration_card', 'Ration Card', 'राशन कार्ड', TRUE, 2),
    ('nfbs', 'bank_account', 'Bank Account', 'बैंक खाता', TRUE, 3),

    ('pm-kisan', 'aadhaar', 'Aadhaar Card', 'आधार कार्ड', TRUE, 0),
    ('pm-kisan', 'land_record', 'Land Ownership Record', 'भूमि स्वामित्व रिकॉर्ड', TRUE, 1),
    ('pm-kisan', 'bank_account', 'Bank Account', 'बैंक खाता', TRUE, 2),

    ('pmfby', 'aadhaar', 'Aadhaar Card', 'आधार कार्ड', TRUE, 0),
    ('pmfby', 'land_record', 'Land Ownership Record', 'भूमि स्वामित्व रिकॉर्ड', TRUE, 1),
    ('pmfby', 'bank_account', 'Bank Account', 'बैंक खाता', TRUE, 2),

    ('pmjjby', 'aadhaar', 'Aadhaar Card', 'आधार कार्ड', TRUE, 0),
    ('pmjjby', 'bank_account', 'Bank Account', 'बैंक खाता', TRUE, 1),

    ('pmsby', 'aadhaar', 'Aadhaar Card', 'आधार कार्ड', TRUE, 0),
    ('pmsby', 'bank_account', 'Bank Account', 'बैंक खाता', TRUE, 1),

    ('apy', 'aadhaar', 'Aadhaar Card', 'आधार कार्ड', TRUE, 0),
    ('apy', 'bank_account', 'Bank Account', 'बैंक खाता', TRUE, 1),

    ('pmjay', 'aadhaar', 'Aadhaar Card', 'आधार कार्ड', TRUE, 0),
    ('pmjay', 'ration_card', 'Ration Card', 'राशन कार्ड', TRUE, 1),

    ('pmay-g', 'aadhaar', 'Aadhaar Card', 'आधार कार्ड', TRUE, 0),
    ('pmay-g', 'income_cert', 'Income Certificate', 'आय प्रमाण पत्र', TRUE, 1),
    ('pmay-g', 'bank_account', 'Bank Account', 'बैंक खाता', TRUE, 2),

    ('pmay-u', 'aadhaar', 'Aadhaar Card', 'आधार कार्ड', TRUE, 0),
    ('pmay-u', 'income_cert', 'Income Certificate', 'आय प्रमाण पत्र', TRUE, 1),
    ('pmay-u', 'bank_account', 'Bank Account', 'बैंक खाता', TRUE, 2),

    ('mgnrega', 'aadhaar', 'Aadhaar Card', 'आधार कार्ड', TRUE, 0),
    ('mgnrega', 'bank_account', 'Bank Account', 'बैंक खाता', TRUE, 1),

    ('post-matric-sc', 'aadhaar', 'Aadhaar Card', 'आधार कार्ड', TRUE, 0),
    ('post-matric-sc', 'caste_cert', 'Caste Certificate', 'जाति प्रमाण पत्र', TRUE, 1),
    ('post-matric-sc', 'income_cert', 'Income Certificate', 'आय प्रमाण पत्र', TRUE, 2),
    ('post-matric-sc', 'student_cert', 'Bonafide Student Certificate', 'बोनाफाइड छात्र प्रमाण पत्र', TRUE, 3),

    ('post-matric-st', 'aadhaar', 'Aadhaar Card', 'आधार कार्ड', TRUE, 0),
    ('post-matric-st', 'caste_cert', 'Caste Certificate', 'जाति प्रमाण पत्र', TRUE, 1),
    ('post-matric-st', 'income_cert', 'Income Certificate', 'आय प्रमाण पत्र', TRUE, 2),
    ('post-matric-st', 'student_cert', 'Bonafide Student Certificate', 'बोनाफाइड छात्र प्रमाण पत्र', TRUE, 3),

    ('post-matric-obc', 'aadhaar', 'Aadhaar Card', 'आधार कार्ड', TRUE, 0),
    ('post-matric-obc', 'caste_cert', 'Caste Certificate', 'जाति प्रमाण पत्र', TRUE, 1),
    ('post-matric-obc', 'income_cert', 'Income Certificate', 'आय प्रमाण पत्र', TRUE, 2),
    ('post-matric-obc', 'student_cert', 'Bonafide Student Certificate', 'बोनाफाइड छात्र प्रमाण पत्र', TRUE, 3),

    ('post-matric-minority', 'aadhaar', 'Aadhaar Card', 'आधार कार्ड', TRUE, 0),
    ('post-matric-minority', 'income_cert', 'Income Certificate', 'आय प्रमाण पत्र', TRUE, 1),
    ('post-matric-minority', 'student_cert', 'Bonafide Student Certificate', 'बोनाफाइड छात्र प्रमाण पत्र', TRUE, 2),

    ('pre-matric-sc', 'aadhaar', 'Aadhaar Card', 'आधार कार्ड', TRUE, 0),
    ('pre-matric-sc', 'caste_cert', 'Caste Certificate', 'जाति प्रमाण पत्र', TRUE, 1),
    ('pre-matric-sc', 'income_cert', 'Income Certificate', 'आय प्रमाण पत्र', TRUE, 2),

    ('pre-matric-st', 'aadhaar', 'Aadhaar Card', 'आधार कार्ड', TRUE, 0),
    ('pre-matric-st', 'caste_cert', 'Caste Certificate', 'जाति प्रमाण पत्र', TRUE, 1),
    ('pre-matric-st', 'income_cert', 'Income Certificate', 'आय प्रमाण पत्र', TRUE, 2),

    ('ssy', 'aadhaar', 'Aadhaar Card of Guardian', 'अभिभावक का आधार कार्ड', TRUE, 0),
    ('ssy', 'birth_cert', 'Birth Certificate of Girl Child', 'बालिका का जन्म प्रमाण पत्र', TRUE, 1),
    ('ssy', 'bank_account', 'Bank Account', 'बैंक खाता', TRUE, 2),

    ('pmmvy', 'aadhaar', 'Aadhaar Card', 'आधार कार्ड', TRUE, 0),
    ('pmmvy', 'bank_account', 'Bank Account', 'बैंक खाता', TRUE, 1),
    ('pmmvy', 'marriage_cert', 'Marriage Certificate', 'विवाह प्रमाण पत्र', FALSE, 2),

    ('jsy', 'aadhaar', 'Aadhaar Card', 'आधार कार्ड', TRUE, 0),
    ('jsy', 'ration_card', 'Ration Card', 'राशन कार्ड', TRUE, 1),
    ('jsy', 'bank_account', 'Bank Account', 'बैंक खाता', TRUE, 2),

    ('pmjdy', 'aadhaar', 'Aadhaar Card', 'आधार कार्ड', TRUE, 0),
    ('pmjdy', 'voter_id', 'Voter ID', 'मतदाता पहचान पत्र', FALSE, 1),

    ('pmmy', 'aadhaar', 'Aadhaar Card', 'आधार कार्ड', TRUE, 0),
    ('pmmy', 'pan_card', 'PAN Card', 'पैन कार्ड', TRUE, 1),
    ('pmmy', 'bank_account', 'Bank Account', 'बैंक खाता', TRUE, 2),

    ('standup-india', 'aadhaar', 'Aadhaar Card', 'आधार कार्ड', TRUE, 0),
    ('standup-india', 'pan_card', 'PAN Card', 'पैन कार्ड', TRUE, 1),
    ('standup-india', 'caste_cert', 'Caste Certificate', 'जाति प्रमाण पत्र', FALSE, 2),

    ('pmuy', 'aadhaar', 'Aadhaar Card', 'आधार कार्ड', TRUE, 0),
    ('pmuy', 'ration_card', 'Ration Card', 'राशन कार्ड', TRUE, 1),
    ('pmuy', 'bank_account', 'Bank Account', 'बैंक खाता', TRUE, 2),

    ('saubhagya', 'aadhaar', 'Aadhaar Card', 'आधार कार्ड', TRUE, 0),
    ('saubhagya', 'ration_card', 'Ration Card', 'राशन कार्ड', FALSE, 1),

    ('pmkvy', 'aadhaar', 'Aadhaar Card', 'आधार कार्ड', TRUE, 0),
    ('pmkvy', 'bank_account', 'Bank Account', 'बैंक खाता', TRUE, 1),

    ('pm-svanidhi', 'aadhaar', 'Aadhaar Card', 'आधार कार्ड', TRUE, 0),
    ('pm-svanidhi', 'voter_id', 'Voter ID', 'मतदाता पहचान पत्र', FALSE, 1),

    ('kcc', 'aadhaar', 'Aadhaar Card', 'आधार कार्ड', TRUE, 0),
    ('kcc', 'land_record', 'Land Ownership Record', 'भूमि स्वामित्व रिकॉर्ड', TRUE, 1),
    ('kcc', 'bank_account', 'Bank Account', 'बैंक खाता', TRUE, 2),

    ('pmmatsya', 'aadhaar', 'Aadhaar Card', 'आधार कार्ड', TRUE, 0),
    ('pmmatsya', 'bank_account', 'Bank Account', 'बैंक खाता', TRUE, 1),

    ('scst-obc-post-matric-topclass', 'aadhaar', 'Aadhaar Card', 'आधार कार्ड', TRUE, 0),
    ('scst-obc-post-matric-topclass', 'caste_cert', 'Caste Certificate', 'जाति प्रमाण पत्र', TRUE, 1),
    ('scst-obc-post-matric-topclass', 'student_cert', 'Bonafide Student Certificate', 'बोनाफाइड छात्र प्रमाण पत्र', TRUE, 2)
) AS v(slug, document_type, name_en, name_hi, mandatory, ord)
    ON s.slug = v.slug
ON CONFLICT (scheme_id, document_type) DO UPDATE SET
    document_name_en = EXCLUDED.document_name_en,
    document_name_hi = EXCLUDED.document_name_hi,
    is_mandatory = EXCLUDED.is_mandatory,
    sort_order = EXCLUDED.sort_order;

-- ---------------------------------------------------------------------------
-- Application fields
-- ---------------------------------------------------------------------------
INSERT INTO scheme_application_fields (scheme_id, field_name, field_label_en, field_label_hi, field_type, profile_mapping, is_required, sort_order)
SELECT id, v.field_name, v.label_en, v.label_hi, v.field_type, v.mapping, v.required, v.ord
FROM schemes s
JOIN (VALUES
    ('ignoaps', 'full_name', 'Full Name', 'पूरा नाम', 'text', 'full_name', TRUE, 0),
    ('ignoaps', 'age', 'Age', 'आयु', 'number', 'age', TRUE, 1),
    ('ignoaps', 'gender', 'Gender', 'लिंग', 'select', 'gender', TRUE, 2),
    ('ignoaps', 'address', 'Address', 'पता', 'text', 'address', TRUE, 3),
    ('ignoaps', 'bank_account_number', 'Bank Account Number', 'बैंक खाता संख्या', 'text', 'bank_account_number', TRUE, 4),

    ('ignwps', 'full_name', 'Full Name', 'पूरा नाम', 'text', 'full_name', TRUE, 0),
    ('ignwps', 'age', 'Age', 'आयु', 'number', 'age', TRUE, 1),
    ('ignwps', 'marital_status', 'Marital Status', 'वैवाहिक स्थिति', 'select', 'marital_status', TRUE, 2),
    ('ignwps', 'bank_account_number', 'Bank Account Number', 'बैंक खाता संख्या', 'text', 'bank_account_number', TRUE, 3),

    ('igndps', 'full_name', 'Full Name', 'पूरा नाम', 'text', 'full_name', TRUE, 0),
    ('igndps', 'age', 'Age', 'आयु', 'number', 'age', TRUE, 1),
    ('igndps', 'disability_percentage', 'Disability Percentage', 'विकलांगता प्रतिशत', 'number', 'disability_percentage', TRUE, 2),
    ('igndps', 'bank_account_number', 'Bank Account Number', 'बैंक खाता संख्या', 'text', 'bank_account_number', TRUE, 3),

    ('nfbs', 'applicant_name', 'Applicant Name', 'आवेदक का नाम', 'text', 'full_name', TRUE, 0),
    ('nfbs', 'deceased_name', 'Name of Deceased', 'मृतक का नाम', 'text', NULL, TRUE, 1),
    ('nfbs', 'relationship', 'Relationship with Deceased', 'मृतक से संबंध', 'text', NULL, TRUE, 2),

    ('pm-kisan', 'full_name', 'Full Name', 'पूरा नाम', 'text', 'full_name', TRUE, 0),
    ('pm-kisan', 'aadhaar_number', 'Aadhaar Number', 'आधार संख्या', 'text', 'aadhaar_number', TRUE, 1),
    ('pm-kisan', 'land_acres', 'Land Holding (acres)', 'भूमि जोत (एकड़)', 'number', 'land_acres', TRUE, 2),
    ('pm-kisan', 'bank_account_number', 'Bank Account Number', 'बैंक खाता संख्या', 'text', 'bank_account_number', TRUE, 3),

    ('pmjay', 'full_name', 'Full Name', 'पूरा नाम', 'text', 'full_name', TRUE, 0),
    ('pmjay', 'ration_card_number', 'Ration Card Number', 'राशन कार्ड संख्या', 'text', 'ration_card_number', TRUE, 1),
    ('pmjay', 'family_size', 'Family Size', 'परिवार का आकार', 'number', 'family_size', TRUE, 2),

    ('pmay-g', 'full_name', 'Full Name', 'पूरा नाम', 'text', 'full_name', TRUE, 0),
    ('pmay-g', 'village', 'Village', 'गांव', 'text', NULL, TRUE, 1),
    ('pmay-g', 'income_annual', 'Annual Income', 'वार्षिक आय', 'number', 'income_annual', TRUE, 2),
    ('pmay-g', 'bank_account_number', 'Bank Account Number', 'बैंक खाता संख्या', 'text', 'bank_account_number', TRUE, 3),

    ('pmay-u', 'full_name', 'Full Name', 'पूरा नाम', 'text', 'full_name', TRUE, 0),
    ('pmay-u', 'income_annual', 'Annual Income', 'वार्षिक आय', 'number', 'income_annual', TRUE, 1),
    ('pmay-u', 'city', 'City', 'शहर', 'text', NULL, TRUE, 2),

    ('mgnrega', 'full_name', 'Full Name', 'पूरा नाम', 'text', 'full_name', TRUE, 0),
    ('mgnrega', 'age', 'Age', 'आयु', 'number', 'age', TRUE, 1),
    ('mgnrega', 'village', 'Village', 'गांव', 'text', NULL, TRUE, 2),
    ('mgnrega', 'bank_account_number', 'Bank Account Number', 'बैंक खाता संख्या', 'text', 'bank_account_number', TRUE, 3),

    ('post-matric-sc', 'student_name', 'Student Name', 'विद्यार्थी का नाम', 'text', NULL, TRUE, 0),
    ('post-matric-sc', 'institution', 'Institution Name', 'संस्थान का नाम', 'text', NULL, TRUE, 1),
    ('post-matric-sc', 'course', 'Course', 'पाठ्यक्रम', 'text', NULL, TRUE, 2),
    ('post-matric-sc', 'income_annual', 'Annual Family Income', 'वार्षिक पारिवारिक आय', 'number', 'income_annual', TRUE, 3),
    ('post-matric-sc', 'bank_account_number', 'Bank Account Number', 'बैंक खाता संख्या', 'text', 'bank_account_number', TRUE, 4),

    ('post-matric-st', 'student_name', 'Student Name', 'विद्यार्थी का नाम', 'text', NULL, TRUE, 0),
    ('post-matric-st', 'institution', 'Institution Name', 'संस्थान का नाम', 'text', NULL, TRUE, 1),
    ('post-matric-st', 'course', 'Course', 'पाठ्यक्रम', 'text', NULL, TRUE, 2),
    ('post-matric-st', 'income_annual', 'Annual Family Income', 'वार्षिक पारिवारिक आय', 'number', 'income_annual', TRUE, 3),
    ('post-matric-st', 'bank_account_number', 'Bank Account Number', 'बैंक खाता संख्या', 'text', 'bank_account_number', TRUE, 4),

    ('post-matric-obc', 'student_name', 'Student Name', 'विद्यार्थी का नाम', 'text', NULL, TRUE, 0),
    ('post-matric-obc', 'institution', 'Institution Name', 'संस्थान का नाम', 'text', NULL, TRUE, 1),
    ('post-matric-obc', 'income_annual', 'Annual Family Income', 'वार्षिक पारिवारिक आय', 'number', 'income_annual', TRUE, 2),

    ('post-matric-minority', 'student_name', 'Student Name', 'विद्यार्थी का नाम', 'text', NULL, TRUE, 0),
    ('post-matric-minority', 'institution', 'Institution Name', 'संस्थान का नाम', 'text', NULL, TRUE, 1),
    ('post-matric-minority', 'income_annual', 'Annual Family Income', 'वार्षिक पारिवारिक आय', 'number', 'income_annual', TRUE, 2),

    ('ssy', 'guardian_name', 'Guardian Name', 'अभिभावक का नाम', 'text', 'full_name', TRUE, 0),
    ('ssy', 'girl_child_name', 'Girl Child Name', 'बालिका का नाम', 'text', NULL, TRUE, 1),
    ('ssy', 'girl_child_dob', 'Girl Child Date of Birth', 'बालिका की जन्म तिथि', 'date', NULL, TRUE, 2),

    ('pmmvy', 'mother_name', 'Mother Name', 'माता का नाम', 'text', 'full_name', TRUE, 0),
    ('pmmvy', 'age', 'Age', 'आयु', 'number', 'age', TRUE, 1),
    ('pmmvy', 'bank_account_number', 'Bank Account Number', 'बैंक खाता संख्या', 'text', 'bank_account_number', TRUE, 2),

    ('pmjdy', 'full_name', 'Full Name', 'पूरा नाम', 'text', 'full_name', TRUE, 0),
    ('pmjdy', 'aadhaar_number', 'Aadhaar Number', 'आधार संख्या', 'text', 'aadhaar_number', FALSE, 1),

    ('pmmy', 'applicant_name', 'Applicant Name', 'आवेदक का नाम', 'text', 'full_name', TRUE, 0),
    ('pmmy', 'business_type', 'Business Type', 'व्यवसाय का प्रकार', 'text', 'occupation', TRUE, 1),
    ('pmmy', 'loan_amount', 'Loan Amount Required', 'आवश्यक ऋण राशि', 'number', NULL, TRUE, 2),

    ('standup-india', 'applicant_name', 'Applicant Name', 'आवेदक का नाम', 'text', 'full_name', TRUE, 0),
    ('standup-india', 'business_type', 'Business Type', 'व्यवसाय का प्रकार', 'text', NULL, TRUE, 1),
    ('standup-india', 'loan_amount', 'Loan Amount Required', 'आवश्यक ऋण राशि', 'number', NULL, TRUE, 2),

    ('pmuy', 'applicant_name', 'Applicant Name', 'आवेदक का नाम', 'text', 'full_name', TRUE, 0),
    ('pmuy', 'ration_card_number', 'Ration Card Number', 'राशन कार्ड संख्या', 'text', 'ration_card_number', TRUE, 1),
    ('pmuy', 'bank_account_number', 'Bank Account Number', 'बैंक खाता संख्या', 'text', 'bank_account_number', TRUE, 2),

    ('saubhagya', 'applicant_name', 'Applicant Name', 'आवेदक का नाम', 'text', 'full_name', TRUE, 0),
    ('saubhagya', 'address', 'Address', 'पता', 'text', 'address', TRUE, 1),

    ('pmkvy', 'applicant_name', 'Applicant Name', 'आवेदक का नाम', 'text', 'full_name', TRUE, 0),
    ('pmkvy', 'age', 'Age', 'आयु', 'number', 'age', TRUE, 1),
    ('pmkvy', 'preferred_trade', 'Preferred Trade', 'पसंदीदा व्यवसाय', 'text', NULL, FALSE, 2),

    ('pm-svanidhi', 'applicant_name', 'Applicant Name', 'आवेदक का नाम', 'text', 'full_name', TRUE, 0),
    ('pm-svanidhi', 'vending_site', 'Vending Site', 'विक्रय स्थल', 'text', NULL, TRUE, 1),
    ('pm-svanidhi', 'loan_amount', 'Loan Amount Required', 'आवश्यक ऋण राशि', 'number', NULL, TRUE, 2),

    ('kcc', 'farmer_name', 'Farmer Name', 'किसान का नाम', 'text', 'full_name', TRUE, 0),
    ('kcc', 'land_acres', 'Land Holding (acres)', 'भूमि जोत (एकड़)', 'number', 'land_acres', TRUE, 1),
    ('kcc', 'bank_account_number', 'Bank Account Number', 'बैंक खाता संख्या', 'text', 'bank_account_number', TRUE, 2),

    ('pmmatsya', 'applicant_name', 'Applicant Name', 'आवेदक का नाम', 'text', 'full_name', TRUE, 0),
    ('pmmatsya', 'activity_type', 'Fishery Activity Type', 'मत्स्य गतिविधि प्रकार', 'text', 'occupation', TRUE, 1),

    ('scst-obc-post-matric-topclass', 'student_name', 'Student Name', 'विद्यार्थी का नाम', 'text', NULL, TRUE, 0),
    ('scst-obc-post-matric-topclass', 'institution', 'Institution Name', 'संस्थान का नाम', 'text', NULL, TRUE, 1),
    ('scst-obc-post-matric-topclass', 'income_annual', 'Annual Family Income', 'वार्षिक पारिवारिक आय', 'number', 'income_annual', TRUE, 2)
) AS v(slug, field_name, label_en, label_hi, field_type, mapping, required, ord)
    ON s.slug = v.slug
ON CONFLICT (scheme_id, field_name) DO UPDATE SET
    field_label_en = EXCLUDED.field_label_en,
    field_label_hi = EXCLUDED.field_label_hi,
    field_type = EXCLUDED.field_type,
    profile_mapping = EXCLUDED.profile_mapping,
    is_required = EXCLUDED.is_required,
    sort_order = EXCLUDED.sort_order;
