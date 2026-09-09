-- Per-user "seen" state for radar cards (Feedly-style unread). kind = job | grant,
-- user = exe.dev email (lowercased) or 'admin' for the basic-auth fallback.
CREATE TABLE IF NOT EXISTS seen (
  kind    TEXT NOT NULL,
  item_id INTEGER NOT NULL,
  user    TEXT NOT NULL,
  at      TEXT NOT NULL DEFAULT (datetime('now')),
  PRIMARY KEY (kind, item_id, user)
);
CREATE INDEX IF NOT EXISTS seen_user ON seen(user, kind);

INSERT OR IGNORE INTO migrations (migration_number, migration_name)
VALUES (015, '015-seen');
