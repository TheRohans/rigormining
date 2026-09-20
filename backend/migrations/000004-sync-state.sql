-- +migrate Up
-- Replaces the old delivered_at timestamp with a state machine that also
-- tracks pending requests, not just "has this ever been delivered":
-- NULL (not on the sync list), request_sync (queued to copy),
-- synced (confirmed on the device), request_remove (queued to delete).
ALTER TABLE library_item ADD COLUMN sync_state TEXT;
UPDATE library_item SET sync_state = 'synced' WHERE delivered_at IS NOT NULL;
ALTER TABLE library_item DROP COLUMN delivered_at;

-- +migrate Down
ALTER TABLE library_item ADD COLUMN delivered_at TEXT;
ALTER TABLE library_item DROP COLUMN sync_state;
