INSERT INTO document_types (
    id, name_en, name_hi, description_en, description_hi, category,
    issuing_authority_en, issuing_authority_hi, acquisition_url,
    acquisition_steps_en, acquisition_steps_hi, typical_time, typical_fee, source_url
) VALUES
    ('aadhaar', 'Aadhaar Card', 'आधार कार्ड',
     '12-digit unique identity number issued by UIDAI.', 'UIDAI द्वारा जारी 12 अंकों की विशिष्ट पहचान संख्या।',
     'identity', 'UIDAI', 'यूआईडीएआई', 'https://uidai.gov.in/',
     '1. Visit an Aadhaar Enrolment Centre or use the online update portal.' || chr(10) ||
     '2. Submit proof of identity and address.' || chr(10) ||
     '3. Biometrics are captured.' || chr(10) ||
     '4. Aadhaar is delivered by post.', 
     '1. आधार नामांकन केंद्र पर जाएं या ऑनलाइन पोर्टल का उपयोग करें।' || chr(10) ||
     '2. पहचान और पते का प्रमाण जमा करें।' || chr(10) ||
     '3. बायोमेट्रिक्स दर्ज किए जाते हैं।' || chr(10) ||
     '4. आधार डाक से भेजा जाता है।',
     'Up to 90 days', 'Free', 'https://uidai.gov.in/'),

    ('income_cert', 'Income Certificate', 'आय प्रमाण पत्र',
     'Certificate stating the annual income of the family.', 'परिवार की वार्षिक आय दर्शाने वाला प्रमाण पत्र।',
     'financial', 'Revenue Department / Tehsil', 'राजस्व विभाग / तहसील', NULL,
     '1. Apply at the Tehsil or online state e-District portal.' || chr(10) ||
     '2. Submit identity, address and income proof.' || chr(10) ||
     '3. Verification by the revenue officer.' || chr(10) ||
     '4. Download or collect the certificate.',
     '1. तहसील कार्यालय या राज्य के ई-डिस्ट्रिक्ट पोर्टल पर आवेदन करें।' || chr(10) ||
     '2. पहचान, पता और आय प्रमाण जमा करें।' || chr(10) ||
     '3. राजस्व अधिकारी द्वारा सत्यापन।' || chr(10) ||
     '4. प्रमाण पत्र डाउनलोड या प्राप्त करें।',
     '7-15 working days', '₹50', NULL),

    ('domicile_cert', 'Domicile Certificate', 'मूल निवास प्रमाण पत्र',
     'Proof of residence in a particular state.', 'किसी विशेष राज्य में निवास का प्रमाण।',
     'residence', 'Revenue Department / Tehsil', 'राजस्व विभाग / तहसील', NULL,
     '1. Apply at the Tehsil or state e-District portal.' || chr(10) ||
     '2. Submit residence and identity proof.' || chr(10) ||
     '3. Field verification may be conducted.' || chr(10) ||
     '4. Collect the certificate.',
     '1. तहसील या राज्य के ई-डिस्ट्रिक्ट पोर्टल पर आवेदन करें।' || chr(10) ||
     '2. निवास और पहचान प्रमाण जमा करें।' || chr(10) ||
     '3. क्षेत्र सत्यापन किया जा सकता है।' || chr(10) ||
     '4. प्रमाण पत्र प्राप्त करें।',
     '7-15 working days', '₹50', NULL),

    ('caste_cert', 'Caste Certificate', 'जाति प्रमाण पत्र',
     'Certificate of belonging to a reserved category.', 'आरक्षित श्रेणी से संबंधित होने का प्रमाण पत्र।',
     'identity', 'Revenue Department / Tehsil', 'राजस्व विभाग / तहसील', NULL,
     '1. Apply at Tehsil or state portal.' || chr(10) ||
     '2. Submit identity, address and family records.' || chr(10) ||
     '3. Verification against revenue records.' || chr(10) ||
     '4. Collect the certificate.',
     '1. तहसील या राज्य पोर्टल पर आवेदन करें।' || chr(10) ||
     '2. पहचान, पता और पारिवारिक रिकॉर्ड जमा करें।' || chr(10) ||
     '3. राजस्व रिकॉर्ड के विरुद्ध सत्यापन।' || chr(10) ||
     '4. प्रमाण पत्र प्राप्त करें।',
     '7-15 working days', '₹50', NULL),

    ('bank_account', 'Bank Account', 'बैंक खाता',
     'A savings account in a bank or post office.', 'बैंक या डाकघर में बचत खाता।',
     'financial', 'Any nationalized or scheduled bank', 'कोई भी राष्ट्रीयकृत या अनुसूचित बैंक', NULL,
     '1. Visit a bank branch with identity and address proof.' || chr(10) ||
     '2. Complete KYC and fill the account opening form.' || chr(10) ||
     '3. Receive account number and passbook.',
     '1. पहचान और पता प्रमाण के साथ बैंक शाखा जाएं।' || chr(10) ||
     '2. केवाईसी पूर्ण करें और खाता खोलने का फॉर्म भरें।' || chr(10) ||
     '3. खाता संख्या और पासबुक प्राप्त करें।',
     'Same day', 'Free', NULL),

    ('ration_card', 'Ration Card', 'राशन कार्ड',
     'Card entitling the family to subsidised food grains.', 'परिवार को रियायती खाद्यान्न का हक देने वाला कार्ड।',
     'financial', 'Food and Civil Supplies Department', 'खाद्य एवं नागरिक आपूर्ति विभाग', NULL,
     '1. Apply at the local Food and Civil Supplies office or portal.' || chr(10) ||
     '2. Submit family details and residence proof.' || chr(10) ||
     '3. Verification and card issuance.',
     '1. स्थानीय खाद्य एवं नागरिक आपूर्ति कार्यालय या पोर्टल पर आवेदन करें।' || chr(10) ||
     '2. परिवार विवरण और निवास प्रमाण जमा करें।' || chr(10) ||
     '3. सत्यापन और कार्ड जारी करना।',
     '15-30 working days', 'Free', NULL),

    ('voter_id', 'Voter ID (EPIC)', 'मतदाता पहचान पत्र',
     'Electoral Photo Identity Card.', 'मतदाता फोटो पहचान पत्र।',
     'identity', 'Election Commission of India', 'भारत निर्वाचन आयोग', 'https://voters.eci.gov.in/',
     '1. Register online on the ECI voter portal or via Form 6.' || chr(10) ||
     '2. Submit identity and address proof.' || chr(10) ||
     '3. Verification and EPIC dispatch.',
     '1. ईसीआई मतदाता पोर्टल पर या फॉर्म 6 के माध्यम से ऑनलाइन पंजीकरण करें।' || chr(10) ||
     '2. पहचान और पता प्रमाण जमा करें।' || chr(10) ||
     '3. सत्यापन और ईपीआईसी प्रेषण।',
     '15-30 working days', 'Free', 'https://voters.eci.gov.in/'),

    ('birth_cert', 'Birth Certificate', 'जन्म प्रमाण पत्र',
     'Official record of birth.', 'जन्म का आधिकारिक रिकॉर्ड।',
     'identity', 'Municipal Corporation / Gram Panchayat', 'नगर निगम / ग्राम पंचायत', NULL,
     '1. Apply at the local registrar (municipal or panchayat).' || chr(10) ||
     '2. Submit hospital record and parents details.' || chr(10) ||
     '3. Certificate issued after verification.',
     '1. स्थानीय रजिस्ट्रार (नगरपालिका या पंचायत) में आवेदन करें।' || chr(10) ||
     '2. अस्पताल रिकॉर्ड और माता-पिता का विवरण जमा करें।' || chr(10) ||
     '3. सत्यापन के बाद प्रमाण पत्र जारी।',
     '7-15 working days', '₹20-₹50', NULL),

    ('pan_card', 'PAN Card', 'पैन कार्ड',
     'Permanent Account Number issued by the Income Tax Department.', 'आयकर विभाग द्वारा जारी स्थायी खाता संख्या।',
     'financial', 'Income Tax Department', 'आयकर विभाग', 'https://www.onlineservices.nsdl.com/',
     '1. Apply online via NSDL or UTIITSL.' || chr(10) ||
     '2. Submit identity and address proof.' || chr(10) ||
     '3. e-PAN or physical card is issued.',
     '1. एनएसडीएल या यूटीआईआईटीएसएल के माध्यम से ऑनलाइन आवेदन करें।' || chr(10) ||
     '2. पहचान और पता प्रमाण जमा करें।' || chr(10) ||
     '3. ई-पैन या भौतिक कार्ड जारी किया जाता है।',
     '7-15 working days', '₹107', 'https://www.onlineservices.nsdl.com/'),

    ('disability_cert', 'Disability Certificate', 'विकलांगता प्रमाण पत्र',
     'Certificate of disability issued by a medical board.', 'चिकित्सा बोर्ड द्वारा जारी विकलांगता प्रमाण पत्र।',
     'other', 'District Medical Board', 'जिला चिकित्सा बोर्ड', NULL,
     '1. Apply at the district hospital or online state portal.' || chr(10) ||
     '2. Appear before the medical board for assessment.' || chr(10) ||
     '3. Certificate issued with disability percentage.',
     '1. जिला अस्पताल या ऑनलाइन राज्य पोर्टल पर आवेदन करें।' || chr(10) ||
     '2. मूल्यांकन के लिए चिकित्सा बोर्ड के समक्ष उपस्थित हों।' || chr(10) ||
     '3. विकलांगता प्रतिशत के साथ प्रमाण पत्र जारी।',
     '15-30 working days', 'Free', NULL),

    ('land_record', 'Land Ownership Record', 'भूमि स्वामित्व रिकॉर्ड',
     'Record of land ownership (RoR / Jamabandi / Patta).', 'भूमि स्वामित्व का रिकॉर्ड (रिकॉर्ड ऑफ राइट्स / जमाबंदी / पट्टा)।',
     'residence', 'Revenue Department', 'राजस्व विभाग', NULL,
     '1. Obtain from the Tehsil or state land records portal.' || chr(10) ||
     '2. Provide survey or khasra number.' || chr(10) ||
     '3. Download or collect the record.',
     '1. तहसील या राज्य भूमि रिकॉर्ड पोर्टल से प्राप्त करें।' || chr(10) ||
     '2. सर्वेक्षण या खसरा संख्या प्रदान करें।' || chr(10) ||
     '3. रिकॉर्ड डाउनलोड या प्राप्त करें।',
     '1-7 working days', '₹10-₹30', NULL),

    ('student_cert', 'Bonafide Student Certificate', 'बोनाफाइड छात्र प्रमाण पत्र',
     'Certificate from an educational institution confirming enrolment.', 'नामांकन की पुष्टि करने वाला शैक्षणिक संस्थान का प्रमाण पत्र।',
     'education', 'Educational Institution', 'शैक्षणिक संस्थान', NULL,
     '1. Request from the school or college office.' || chr(10) ||
     '2. Provide admission number and details.' || chr(10) ||
     '3. Certificate issued on institution letterhead.',
     '1. स्कूल या कॉलेज कार्यालय से अनुरोध करें।' || chr(10) ||
     '2. प्रवेश संख्या और विवरण प्रदान करें।' || chr(10) ||
     '3. संस्थान के लेटरहेड पर प्रमाण पत्र जारी।',
     '1-3 working days', 'Free', NULL),

    ('death_cert', 'Death Certificate', 'मृत्यु प्रमाण पत्र',
     'Official record of death (required for widow schemes).', 'मृत्यु का आधिकारिक रिकॉर्ड (विधवा योजनाओं के लिए आवश्यक)।',
     'identity', 'Municipal Corporation / Gram Panchayat', 'नगर निगम / ग्राम पंचायत', NULL,
     '1. Apply at the local registrar with proof of death.' || chr(10) ||
     '2. Submit hospital or cremation record.' || chr(10) ||
     '3. Certificate issued after verification.',
     '1. मृत्यु के प्रमाण के साथ स्थानीय रजिस्ट्रार में आवेदन करें।' || chr(10) ||
     '2. अस्पताल या अंत्येष्टि रिकॉर्ड जमा करें।' || chr(10) ||
     '3. सत्यापन के बाद प्रमाण पत्र जारी।',
     '7-15 working days', 'Free', NULL),

    ('marriage_cert', 'Marriage Certificate', 'विवाह प्रमाण पत्र',
     'Official registration of marriage.', 'विवाह का आधिकारिक पंजीकरण।',
     'identity', 'Marriage Registrar', 'विवाह रजिस्ट्रार', NULL,
     '1. Apply at the marriage registrar office.' || chr(10) ||
     '2. Both parties submit identity and photographs.' || chr(10) ||
     '3. Certificate issued after registration.',
     '1. विवाह रजिस्ट्रार कार्यालय में आवेदन करें।' || chr(10) ||
     '2. दोनों पक्ष पहचान और तस्वीरें जमा करें।' || chr(10) ||
     '3. पंजीकरण के बाद प्रमाण पत्र जारी।',
     '7-15 working days', '₹100', NULL)
ON CONFLICT (id) DO UPDATE SET
    name_en = EXCLUDED.name_en,
    name_hi = EXCLUDED.name_hi,
    description_en = EXCLUDED.description_en,
    description_hi = EXCLUDED.description_hi,
    category = EXCLUDED.category,
    issuing_authority_en = EXCLUDED.issuing_authority_en,
    issuing_authority_hi = EXCLUDED.issuing_authority_hi,
    acquisition_url = EXCLUDED.acquisition_url,
    acquisition_steps_en = EXCLUDED.acquisition_steps_en,
    acquisition_steps_hi = EXCLUDED.acquisition_steps_hi,
    typical_time = EXCLUDED.typical_time,
    typical_fee = EXCLUDED.typical_fee,
    source_url = EXCLUDED.source_url;
