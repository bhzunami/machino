-- SQLite does not support DROP COLUMN portably; recreate table without the role column.
CREATE TABLE users_new (
    id            TEXT     PRIMARY KEY,
    email         TEXT     NOT NULL UNIQUE,
    name          TEXT     NOT NULL DEFAULT '',
    password_hash TEXT     NOT NULL,
    searchable    INTEGER  NOT NULL DEFAULT 1,
    created_at    DATETIME NOT NULL
);
INSERT INTO users_new SELECT id, email, name, password_hash, searchable, created_at FROM users;
DROP TABLE users;
ALTER TABLE users_new RENAME TO users;
