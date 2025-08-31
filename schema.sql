CREATE TABLE IF NOT EXISTS orders (
    order_uid VARCHAR(255) PRIMARY KEY,
    track_number VARCHAR(255),
    entry VARCHAR(255),
    delivery JSONB,
    payment JSONB,
    items JSONB,
    locale VARCHAR(10),
    customer_id VARCHAR(255),
    date_created TIMESTAMP WITH TIME ZONE
);