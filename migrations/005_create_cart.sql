-- +goose Up
CREATE TABLE cart_items (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    product_id BIGINT NOT NULL REFERENCES products(id),
    comment TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE cart_item_options (
    id BIGSERIAL PRIMARY KEY,
    cart_item_id BIGINT NOT NULL REFERENCES cart_items(id) ON DELETE CASCADE,
    field_id BIGINT NOT NULL REFERENCES fields(id),
    option_id BIGINT NOT NULL REFERENCES field_options(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(cart_item_id, field_id)
);

-- +goose Down
DROP TABLE cart_item_options;
DROP TABLE cart_items;