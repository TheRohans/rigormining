-- +migrate Up
CREATE TABLE IF NOT EXISTS library_item (
  uuid TEXT PRIMARY KEY,
  user_uuid TEXT REFERENCES users(uuid),
  title TEXT,
  authors TEXT,
  doi TEXT,
  isbn TEXT,
  year INTEGER,
  source_url TEXT,
  file_path TEXT,
  file_type TEXT,
  file_hash TEXT,
  added_date TEXT,
  delivered_at TEXT,
  notes TEXT
);
CREATE INDEX idx_item_user ON library_item(user_uuid);
CREATE UNIQUE INDEX idx_item_hash ON library_item(file_hash);

CREATE TABLE IF NOT EXISTS tag (
  uuid TEXT PRIMARY KEY,
  name TEXT UNIQUE
);

CREATE TABLE IF NOT EXISTS item_tag (
  item_uuid TEXT REFERENCES library_item(uuid),
  tag_uuid TEXT REFERENCES tag(uuid),
  PRIMARY KEY(item_uuid, tag_uuid)
);

CREATE TABLE IF NOT EXISTS collection (
  uuid TEXT PRIMARY KEY,
  user_uuid TEXT REFERENCES users(uuid),
  name TEXT
);

CREATE TABLE IF NOT EXISTS item_collection (
  item_uuid TEXT REFERENCES library_item(uuid),
  collection_uuid TEXT REFERENCES collection(uuid),
  PRIMARY KEY(item_uuid, collection_uuid)
);

-- +migrate Down
DROP TABLE item_collection;
DROP TABLE collection;
DROP TABLE item_tag;
DROP TABLE tag;
DROP TABLE library_item;
