-- +migrate Up
-- Item deletes used to leave the item's tag/collection links behind:
-- SQLite doesn't enforce foreign keys by default, so nothing stopped it.
-- DeleteItem now removes them, and this clears the ones already left over,
-- which would otherwise block copying the data into Postgres (which does
-- enforce them).
DELETE FROM item_tag WHERE item_uuid NOT IN (SELECT uuid FROM library_item);
DELETE FROM item_collection WHERE item_uuid NOT IN (SELECT uuid FROM library_item);

-- +migrate Down
-- Nothing to restore: the rows pointed at items that no longer exist.
