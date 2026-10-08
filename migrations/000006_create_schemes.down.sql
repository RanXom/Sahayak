ALTER TABLE candidate_schemes DROP CONSTRAINT IF EXISTS fk_candidate_promoted_to;
DROP TABLE IF EXISTS scheme_application_fields;
DROP TABLE IF EXISTS scheme_required_documents;
DROP TABLE IF EXISTS scheme_eligibility_rules;
DROP TABLE IF EXISTS schemes;
