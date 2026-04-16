ALTER TABLE users
    ADD COLUMN password_hash TEXT NOT NULL DEFAULT '',
    ADD COLUMN last_login_at TIMESTAMPTZ;

CREATE INDEX idx_users_email ON users(email);
