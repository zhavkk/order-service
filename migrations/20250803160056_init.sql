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
    date_created TIMESTAMP WITH TIME ZONE,
    oof_shard VARCHAR
);

CREATE TABLE delivery (
    delivery_id SERIAL PRIMARY KEY,
    order_uid VARCHAR REFERENCES orders(order_uid) ON DELETE CASCADE,
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
    order_uid VARCHAR REFERENCES orders(order_uid) ON DELETE CASCADE,
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
    order_uid VARCHAR REFERENCES orders(order_uid) ON DELETE CASCADE,
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

CREATE INDEX idx_items_order_uid ON items(order_uid);
CREATE INDEX idx_payments_order_uid ON payments(order_uid);
CREATE INDEX idx_delivery_order_uid ON delivery(order_uid);
CREATE INDEX idx_orders_order_uid ON orders(order_uid);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS items;
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS deliveries;
DROP TABLE IF EXISTS orders;
-- +goose StatementEnd
