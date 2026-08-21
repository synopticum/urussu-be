CREATE TABLE users (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role       TEXT NOT NULL DEFAULT 'user',
    email      TEXT UNIQUE NOT NULL,
    password   TEXT NOT NULL, -- bcrypt hash
    first_name TEXT NOT NULL,
    last_name  TEXT NOT NULL
);

INSERT INTO users (id, role, email, password, first_name, last_name)
VALUES (
    '9f27aa45-efd3-40e7-bba1-72cc9004de5f',
    'user',
    'test@test.com',
    '$2a$10$LIWle3L2IaDE/IHwcaoUYuWpK1pqfHLQv7BLsY6bNrzhlsD9UNDL6', -- bcrypt of "12345678"
    'Сергей',
    'Новиков'
);
