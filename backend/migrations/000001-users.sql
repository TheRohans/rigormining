-- +migrate Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE IF NOT EXISTS users (
  uuid TEXT PRIMARY KEY,
  email TEXT,
  username TEXT,
  picture TEXT,
  authid TEXT,
  salt TEXT,
  UNIQUE(email)
);

CREATE TABLE IF NOT EXISTS token (
  uuid TEXT PRIMARY KEY,
  user_uuid TEXT REFERENCES users(uuid),
  name TEXT,
  value TEXT UNIQUE,
  created_at TEXT
);

-- +migrate Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE token;
DROP TABLE users;
