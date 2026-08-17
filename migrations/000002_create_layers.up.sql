CREATE TABLE layers (
    id         SMALLINT PRIMARY KEY,
    name       TEXT NOT NULL UNIQUE
);

INSERT INTO layers (id, name)
VALUES
    (1, '1940'),
    (2, '1950'),
    (3, '1960'),
    (4, '1970'),
    (5, '1980'),
    (6, '1990'),
    (7, '2000'),
    (8, '2010'),
    (9, '2020');
