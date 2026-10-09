-- Restore the pre-000011 scheme URLs.
UPDATE schemes
SET official_url = 'https://www.npscra.nsdl.co.in',
    application_url = 'https://www.npscra.nsdl.co.in'
WHERE slug = 'apy';

UPDATE schemes
SET official_url = 'https://pmjay.gov.in',
    application_url = 'https://pmjay.gov.in'
WHERE slug = 'pmjay';

UPDATE schemes
SET official_url = 'https://nrega.nic.in',
    application_url = 'https://nrega.nic.in'
WHERE slug = 'mgnrega';

UPDATE schemes
SET official_url = 'https://www.pmkvyofficial.org',
    application_url = 'https://www.pmkvyofficial.org'
WHERE slug = 'pmkvy';

UPDATE schemes
SET official_url = 'https://www.nabard.org',
    application_url = 'https://www.nabard.org'
WHERE slug = 'kcc';

UPDATE schemes
SET official_url = 'https://www.india.gov.in',
    application_url = 'https://www.india.gov.in'
WHERE slug = 'ssy';
