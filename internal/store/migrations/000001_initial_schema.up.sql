CREATE TABLE IF NOT EXISTS users (
    id            TEXT    PRIMARY KEY,
    email         TEXT    NOT NULL UNIQUE,
    name          TEXT    NOT NULL DEFAULT '',
    password_hash TEXT    NOT NULL,
    searchable    INTEGER NOT NULL DEFAULT 1,
    created_at    DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS sessions (
    token      TEXT     PRIMARY KEY,
    user_id    TEXT     NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS password_resets (
    token      TEXT     PRIMARY KEY,
    user_id    TEXT     NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at DATETIME NOT NULL,
    used_at    DATETIME
);

CREATE TABLE IF NOT EXISTS projects (
    id          TEXT     PRIMARY KEY,
    title       TEXT     NOT NULL,
    description TEXT     NOT NULL DEFAULT '',
    color       TEXT     NOT NULL DEFAULT '#4f46e5',
    move_done   INTEGER  NOT NULL DEFAULT 1,
    created_by  TEXT     NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at  DATETIME NOT NULL,
    updated_at  DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS project_members (
    project_id TEXT     NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id    TEXT     NOT NULL REFERENCES users(id)    ON DELETE CASCADE,
    role       TEXT     NOT NULL DEFAULT 'member',
    joined_at  DATETIME NOT NULL,
    PRIMARY KEY (project_id, user_id)
);

CREATE TABLE IF NOT EXISTS project_favorites (
    user_id    TEXT NOT NULL REFERENCES users(id)    ON DELETE CASCADE,
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, project_id)
);

CREATE TABLE IF NOT EXISTS project_columns (
    id         TEXT     PRIMARY KEY,
    project_id TEXT     NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    title      TEXT     NOT NULL,
    position   INTEGER  NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_project_columns_project ON project_columns(project_id, position);

CREATE TABLE IF NOT EXISTS todos (
    id          TEXT     PRIMARY KEY,
    project_id  TEXT     NOT NULL REFERENCES projects(id)       ON DELETE CASCADE,
    column_id   TEXT     REFERENCES project_columns(id)         ON DELETE SET NULL,
    title       TEXT     NOT NULL,
    description TEXT     NOT NULL DEFAULT '',
    due_date    DATETIME,
    priority    TEXT     NOT NULL DEFAULT 'normal',
    completed   INTEGER  NOT NULL DEFAULT 0,
    position    INTEGER  NOT NULL,
    created_by  TEXT     NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at  DATETIME NOT NULL,
    updated_at  DATETIME NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_todos_project_position ON todos(project_id, position);

-- Backfill project_members for any projects that already exist.
INSERT OR IGNORE INTO project_members (project_id, user_id, role, joined_at)
SELECT id, created_by, 'owner', created_at FROM projects;
