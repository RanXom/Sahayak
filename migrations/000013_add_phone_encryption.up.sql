-- Reversible encryption for phone numbers, required to dispatch SMS/WhatsApp.
-- The plaintext is never stored; this column holds AES-256-GCM ciphertext.
ALTER TABLE users ADD COLUMN phone_encrypted TEXT;
ALTER TABLE otp_challenges ADD COLUMN phone_encrypted TEXT;
