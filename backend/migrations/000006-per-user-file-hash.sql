-- +migrate Up
-- Files are stored per user (library/<user>/<item>.<ext>) and the upload
-- dedupe check (GetItemByHash) is per user, so the hash only needs to be
-- unique within one user's library. The old global index made a second
-- user uploading the same paper fail with a constraint error.
DROP INDEX IF EXISTS idx_item_hash;
CREATE UNIQUE INDEX idx_item_user_hash ON library_item(user_uuid, file_hash);

-- +migrate Down
-- Fails if two users have since uploaded the same file.
DROP INDEX IF EXISTS idx_item_user_hash;
CREATE UNIQUE INDEX idx_item_hash ON library_item(file_hash);
