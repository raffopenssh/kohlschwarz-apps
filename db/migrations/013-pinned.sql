-- Owner can pin/star items; pinned rows always sort to the top of their radar.
ALTER TABLE job_postings ADD COLUMN pinned INTEGER NOT NULL DEFAULT 0;
ALTER TABLE funding ADD COLUMN pinned INTEGER NOT NULL DEFAULT 0;

INSERT OR IGNORE INTO migrations (migration_number, migration_name)
VALUES (013, '013-pinned');
