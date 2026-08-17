CREATE TABLE users (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role       TEXT NOT NULL DEFAULT 'user',
    email      TEXT UNIQUE NOT NULL,
    password   TEXT NOT NULL, -- bcrypt hash
    first_name TEXT NOT NULL,
    last_name  TEXT NOT NULL
);

INSERT INTO users (role, email, password, first_name, last_name)
VALUES (
    'nswsergey@yandex.ru',
    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
    'user',
    'Сергей',
    'Новиков'
);
