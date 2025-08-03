-- +goose Up
-- +goose StatementBegin
CREATE TABLE orders (
    order_uid VARCHAR PRIMARY KEY,
    track_number VARCHAR NOT NULL,
    entry VARCHAR,
    locale VARCHAR,
    internal_signature VARCHAR,
    customer_id VARCHAR,
    delivery_service VARCHAR,
    shardkey VARCHAR,
    sm_id INTEGER,
    date_created TIMESTAMP,
    oof_shard VARCHAR
);

CREATE TABLE deliveries (
    delivery_id SERIAL PRIMARY KEY,
    order_uid VARCHAR REFERENCES orders(order_uid),
    name VARCHAR,
    phone VARCHAR,
    zip VARCHAR,
    city VARCHAR,
    address VARCHAR,
    region VARCHAR,
    email VARCHAR
);

CREATE TABLE payments (
    transaction VARCHAR PRIMARY KEY,
    order_uid VARCHAR REFERENCES orders(order_uid),
    request_id VARCHAR,
    currency VARCHAR,
    provider VARCHAR,
    amount INTEGER,
    payment_dt BIGINT,
    bank VARCHAR,
    delivery_cost INTEGER,
    goods_total INTEGER,
    custom_fee INTEGER
);


CREATE TABLE items (
    item_id SERIAL PRIMARY KEY,
    order_uid VARCHAR REFERENCES orders(order_uid),
    chrt_id BIGINT,
    track_number VARCHAR,
    price INTEGER,
    rid VARCHAR,
    name VARCHAR,
    sale INTEGER,
    size VARCHAR,
    total_price INTEGER,
    nm_id BIGINT,
    brand VARCHAR,
    status INTEGER
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
