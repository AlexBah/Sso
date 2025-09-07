CREATE TABLE IF NOT EXISTS users
(
    id        INTEGER PRIMARY KEY,
    name      TEXT NOT NULL,
    email     TEXT NOT NULL,
    phone     TEXT NOT NULL,
    pass_hash BLOB NOT NULL,
    is_admin  BOOLEAN NOT NULL DEFAULT FALSE,
    CONSTRAINT uq_users_phone UNIQUE (phone)
);

CREATE INDEX IF NOT EXISTS idx_phone ON users (phone);

CREATE TABLE IF NOT EXISTS apps
(
    id     INTEGER PRIMARY KEY,
    name   TEXT NOT NULL,
    secret TEXT NOT NULL,
    CONSTRAINT uq_apps_name UNIQUE (name),
    CONSTRAINT uq_apps_secret UNIQUE (secret)
);