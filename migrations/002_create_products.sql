-- +goose Up
CREATE TABLE products (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price NUMERIC(10, 2) NOT NULL,
    template_id BIGINT NULL, -- будущий FK для template
    tags TEXT[],
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_available BOOLEAN NOT NULL DEFAULT FALSE,
    comment_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    template_enabled BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX idx_products_name ON products(name);
CREATE INDEX idx_products_tags ON products USING GIN(tags);

CREATE TABLE product_images (
    id BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    src VARCHAR(255) NOT NULL,
    alt TEXT,
    sort_order INT DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE product_images;
DROP INDEX idx_products_tags;
DROP INDEX idx_products_name;
DROP TABLE products;