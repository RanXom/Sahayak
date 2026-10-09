-- ============================================================================
-- Refresh official/application URLs that have moved since the initial seed.
-- Verified against the live web (October 2026):
--   APY: NSDL e-Gov CRA migrated to Protean (notice on the new site itself)
--   PMJAY: citizen portal is now the NHA beneficiary portal
--   MGNREGA: nrega.nic.in redirects to nrega.dord.gov.in
--   PMKVY: pmkvyofficial.org is dead; PMKVY 4.0 lives on Skill India Digital
--   KCC: JanSamarth is the national portal with online KCC application
--   SSY: India Post hosts the official Sukanya Samriddhi page
-- The seed INSERT uses ON CONFLICT DO NOTHING, so existing databases need
-- these explicit UPDATEs (document_types uses an upsert and needs none).
-- ============================================================================

UPDATE schemes
SET official_url = 'https://www.npscra.proteantech.in',
    application_url = 'https://www.npscra.proteantech.in'
WHERE slug = 'apy';

UPDATE schemes
SET official_url = 'https://beneficiary.nha.gov.in/',
    application_url = 'https://beneficiary.nha.gov.in/'
WHERE slug = 'pmjay';

UPDATE schemes
SET official_url = 'https://nrega.dord.gov.in/',
    application_url = 'https://nrega.dord.gov.in/'
WHERE slug = 'mgnrega';

UPDATE schemes
SET official_url = 'https://www.skillindiadigital.gov.in/pmkvy-landing',
    application_url = 'https://www.skillindiadigital.gov.in/pmkvy-landing'
WHERE slug = 'pmkvy';

UPDATE schemes
SET official_url = 'https://www.jansamarth.in/kisan-credit-card-scheme',
    application_url = 'https://www.jansamarth.in/kisan-credit-card-scheme'
WHERE slug = 'kcc';

UPDATE schemes
SET official_url = 'https://www.indiapost.gov.in/SukanyaSamriddhi.aspx',
    application_url = 'https://www.indiapost.gov.in/SukanyaSamriddhi.aspx'
WHERE slug = 'ssy';
