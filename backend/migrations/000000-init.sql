-- +migrate Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE IF NOT EXISTS researcher (
  uuid UUID,
  name TEXT,
  email TEXT,
  UNIQUE(uuid)
);

CREATE TABLE IF NOT EXISTS highlight (
  uuid TEXT,
  device TEXT,
  start_offset INTEGER,
  end_offset INTEGER,
  page INTEGER,
  highlight_text TEXT,
  highlight_annotation TEXT,
  extra_annotation_data TEXT,
  date_created TEXT,
  content_id TEXT,
  title TEXT,
  ISBN TEXT,
  -- author
  attribution TEXT,
  -- note / highlight
  highlight_type TEXT,
  researcher_uuid UUID,
  FOREIGN KEY(researcher_uuid) REFERENCES researcher(uuid),
  UNIQUE(uuid)
);

-- CREATE UNIQUE INDEX idx_high_suuid ON highlight(suuid);
CREATE INDEX idx_high_text ON highlight(highlight_text);
CREATE INDEX idx_high_create ON highlight(date_created);

-- +migrate Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE highlight;
DROP TABLE researcher;
