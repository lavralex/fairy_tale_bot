-- +goose Up
ALTER TABLE products ADD CONSTRAINT fk_products_template_id
    FOREIGN KEY (template_id) REFERENCES templates(id) ON DELETE SET NULL;

-- +goose Down
ALTER TABLE products DROP CONSTRAINT fk_products_template_id;