-- LLM brief (muse-glimmer) per funding entry: page fetched, 5 labelled lines.
ALTER TABLE funding ADD COLUMN brief TEXT NOT NULL DEFAULT '';
ALTER TABLE funding ADD COLUMN briefed_at TEXT;

INSERT OR IGNORE INTO migrations (migration_number, migration_name)
VALUES (014, '014-funding-brief');
