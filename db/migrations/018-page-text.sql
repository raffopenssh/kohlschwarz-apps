-- Page text captured at fetch/brief time so briefs survive geo-blocks and 404/410.
ALTER TABLE job_postings ADD COLUMN page_text TEXT NOT NULL DEFAULT '';
ALTER TABLE job_postings ADD COLUMN page_src TEXT NOT NULL DEFAULT '';
INSERT OR IGNORE INTO migrations (migration_number, migration_name) VALUES (18, '018-page-text');
