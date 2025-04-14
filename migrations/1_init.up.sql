BEGIN TRANSACTION;

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    login TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    balance NUMERIC(20, 2) NOT NULL CHECK (balance >= 0) DEFAULT 0.00,
    withdrawn NUMERIC(20, 2) NOT NULL DEFAULT 0.00
);

CREATE INDEX IF NOT EXISTS idx_login ON users (login);

CREATE TABLE IF NOT EXISTS withdrawals (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    number TEXT NOT NULL,
    sum NUMERIC(20, 2) NOT NULL,
    processed_at TIMESTAMP,
    CONSTRAINT fk_withdrawals_users FOREIGN KEY (user_id) REFERENCES users (id)
);

CREATE TABLE IF NOT EXISTS order_statuses (
    id SMALLINT PRIMARY KEY,
    status VARCHAR(15) UNIQUE
);

INSERT INTO
    order_statuses (id, status)
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
    CONSTRAINT fk_orders_users FOREIGN KEY (user_id) REFERENCES users (id),
    CONSTRAINT fk_orders_order_statuses FOREIGN KEY (user_id) REFERENCES order_statuses (id)
);

COMMIT TRANSACTION;
