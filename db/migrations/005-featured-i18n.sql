-- Featured flag (higher = pinned to top) and German descriptions
ALTER TABLE apps ADD COLUMN featured INTEGER NOT NULL DEFAULT 0;
ALTER TABLE apps ADD COLUMN description_de TEXT;

-- Record execution of this migration
INSERT OR IGNORE INTO migrations (migration_number, migration_name)
VALUES (005, '005-featured-i18n');
