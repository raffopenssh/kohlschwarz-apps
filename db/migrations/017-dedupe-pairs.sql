-- LLM-confirmed duplicate verdicts between two postings (a_id < b_id).
CREATE TABLE IF NOT EXISTS dedupe_pairs (
  a_id INTEGER NOT NULL,
  b_id INTEGER NOT NULL,
  same INTEGER NOT NULL,          -- 1 = same job, 0 = different
  at   TEXT NOT NULL DEFAULT (datetime('now')),
  PRIMARY KEY (a_id, b_id)
);
INSERT OR IGNORE INTO migrations (migration_number, migration_name)
VALUES (017, '017-dedupe-pairs');
