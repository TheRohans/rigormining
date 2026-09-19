-- +migrate Up
ALTER TABLE library_item ADD COLUMN item_type TEXT;
ALTER TABLE library_item ADD COLUMN venue TEXT;
ALTER TABLE library_item ADD COLUMN volume TEXT;
ALTER TABLE library_item ADD COLUMN number TEXT;
ALTER TABLE library_item ADD COLUMN pages TEXT;
ALTER TABLE library_item ADD COLUMN publisher TEXT;

-- +migrate Down
ALTER TABLE library_item DROP COLUMN item_type;
ALTER TABLE library_item DROP COLUMN venue;
ALTER TABLE library_item DROP COLUMN volume;
ALTER TABLE library_item DROP COLUMN number;
ALTER TABLE library_item DROP COLUMN pages;
ALTER TABLE library_item DROP COLUMN publisher;
