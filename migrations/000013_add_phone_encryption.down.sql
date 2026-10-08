ALTER TABLE otp_challenges DROP COLUMN IF EXISTS phone_encrypted;
ALTER TABLE users DROP COLUMN IF EXISTS phone_encrypted;
