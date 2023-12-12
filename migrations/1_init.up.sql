CREATE TABLE IF NOT EXISTS users
(
    user_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    email VARCHAR NOT NULL UNIQUE,
    password_hash BYTEA NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_email ON users(email);