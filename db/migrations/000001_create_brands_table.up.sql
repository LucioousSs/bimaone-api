CREATE TABLE brands (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    archived_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT brands_name_length
        CHECK (CHAR_LENGTH(name) BETWEEN 2 AND 100),

    CONSTRAINT brands_description_length
        CHECK (description IS NULL OR CHAR_LENGTH(description) <= 500)
);

CREATE UNIQUE INDEX brands_name_unique_idx
    ON brands (LOWER(TRIM(name)));
