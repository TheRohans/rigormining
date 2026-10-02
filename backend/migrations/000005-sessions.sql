-- +migrate Up
-- Server-side sessions replace the old uuid:md5(email+authid+salt) cookie,
-- so a session can be revoked (logout) without touching the user's other
-- devices. Only the SHA-256 of the cookie value is stored.
CREATE TABLE IF NOT EXISTS session (
  id_hash TEXT PRIMARY KEY,
  user_uuid TEXT NOT NULL REFERENCES users(uuid) ON DELETE CASCADE,
  created_at TEXT NOT NULL,
  expires_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS session_user_uuid ON session (user_uuid);

-- The salt only existed to build the old cookie hash. Dropping it logs
-- everyone out once, which is intended.
ALTER TABLE users DROP COLUMN salt;

-- API tokens are now stored as SHA-256 hashes. Existing plaintext values
-- are hashed in Go at startup (SQLite has no sha256()), see
-- repository.HashLegacyTokens; this flag marks which rows are done.
ALTER TABLE token ADD COLUMN hashed INTEGER NOT NULL DEFAULT 0;

-- +migrate Down
ALTER TABLE token DROP COLUMN hashed;
ALTER TABLE users ADD COLUMN salt TEXT;
DROP INDEX IF EXISTS session_user_uuid;
DROP TABLE session;
