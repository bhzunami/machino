CREATE TABLE IF NOT EXISTS app_settings (
    id                   INTEGER PRIMARY KEY CHECK (id = 1),
    app_domain           TEXT    NOT NULL DEFAULT '',
    registration_enabled INTEGER NOT NULL DEFAULT 1,
    smtp_host            TEXT    NOT NULL DEFAULT '',
    smtp_port            TEXT    NOT NULL DEFAULT '587',
    smtp_username        TEXT    NOT NULL DEFAULT '',
    smtp_password        TEXT    NOT NULL DEFAULT '',
    smtp_from            TEXT    NOT NULL DEFAULT '',
    created_at           DATETIME NOT NULL,
    updated_at           DATETIME NOT NULL
);
