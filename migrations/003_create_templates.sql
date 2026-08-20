-- +goose Up
CREATE TABLE templates (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_templates_name ON templates(name);


CREATE TABLE fields (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_active BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE INDEX idx_fields_name ON fields(name);

CREATE TABLE template_fields (
    template_id   BIGINT NOT NULL REFERENCES templates(id) ON DELETE CASCADE,
    field_id BIGINT NOT NULL REFERENCES fields(id) ON DELETE CASCADE,
    sort_order INT DEFAULT 0,
    PRIMARY KEY (template_id, field_id)
);

CREATE TABLE field_options (
    id BIGSERIAL PRIMARY KEY,
    field_id BIGINT NOT NULL REFERENCES fields(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    sort_order INT DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_active BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE INDEX idx_field_options_name ON field_options(name);

-- +goose Down
DROP INDEX idx_field_options_name;
DROP TABLE field_options;
DROP INDEX idx_fields_name;
DROP TABLE template_fields;
DROP TABLE fields;
DROP INDEX idx_templates_name;
DROP TABLE templates;