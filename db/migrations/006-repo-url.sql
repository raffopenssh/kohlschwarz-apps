-- Optional link to the app's public source repository
ALTER TABLE apps ADD COLUMN repo_url TEXT;

-- Record execution of this migration
INSERT OR IGNORE INTO migrations (migration_number, migration_name)
VALUES (006, '006-repo-url');
