CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    login TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_login ON users (login);

CREATE TABLE IF NOT EXISTS balance (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    balance NUMERIC(20, 2) NOT NULL,
    withdrawn NUMERIC(20, 2),
    CONSTRAINT fk_balance_users FOREIGN KEY (user_id) REFERENCES users (id)
);

CREATE TABLE IF NOT EXISTS order_status (
    id SMALLINT PRIMARY KEY,
    status VARCHAR(15) UNIQUE
);

INSERT INTO
    order_status (id, status)
VALUES
    (1, 'NEW'),
    (2, 'PROCESSING'),
    (3, 'INVALID'),
    (4, 'PROCESSED');

CREATE TABLE IF NOT EXISTS orders (
    id BIGSERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    number TEXT UNIQUE NOT NULL,
    status_id SMALLINT NOT NULL,
    accrual NUMERIC(20, 2),
    uploaded_at TIMESTAMP,
    CONSTRAINT fk_order_user FOREIGN KEY (user_id) REFERENCES users (id)
);
