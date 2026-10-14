CREATE TABLE warehouses (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(20) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    address TEXT,
    pic_name VARCHAR(100),
    phone VARCHAR(20),
    archived_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT warehouses_code_uppercase
        CHECK (code = UPPER(code)),

    CONSTRAINT warehouses_name_length
        CHECK (CHAR_LENGTH(name) BETWEEN 2 AND 100),

    CONSTRAINT warehouses_phone_format
        CHECK (
            phone IS NULL
            OR phone ~ '^[0-9+ -]+$'
        )
);