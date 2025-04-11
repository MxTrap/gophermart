CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    login TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_login ON users (login);

CREATE TABLE IF NOT EXISTS tokens (
    id SERIAL PRIMARY KEY,
    refresh_token TEXT UNIQUE,
    expiration_time BIGINT,
    user_id INT,
    CONSTRAINT fk_tokens_users FOREIGN KEY (user_id) REFERENCES users (id)
);
