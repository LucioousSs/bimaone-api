-- Data contoh brand untuk pengembangan lokal. Aman dijalankan berulang: nama yang sudah ada dilewati.
-- Jalankan setelah migrate up: psql "<DSN>" -f db/seeds/brands.sql

INSERT INTO brands (name, description, archived_at)
SELECT v.name, v.description, v.archived_at
FROM (VALUES
    ('Grundfos',           'Pompa sentrifugal, booster, dan sistem dosing', NULL::timestamptz),
    ('Ebara',              'Pompa air industri dan domestik',              NULL::timestamptz),
    ('Wilo',               'Pompa air dan sistem manajemen tekanan',       NULL::timestamptz),
    ('Schneider Electric', 'Panel listrik, kontaktor, dan otomasi',        NULL::timestamptz),
    ('Chint',              'Panel listrik dan komponen instalasi',         NOW()) -- satu brand arsip agar filter status bisa dicoba
) AS v(name, description, archived_at)
WHERE NOT EXISTS (
    SELECT 1 FROM brands b WHERE lower(trim(b.name)) = lower(trim(v.name))
);
