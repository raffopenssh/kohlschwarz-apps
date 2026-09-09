// Package seen stores which radar cards each user has already had on screen
// (≥50 % visible for ~1 s, decided client-side). Purely a reading aid — it never
// feeds ranking or LLM prompts.
package seen

import (
	"context"
	"database/sql"
	"strings"
)

// MaxBatch caps one Mark call; radar.js flushes far smaller batches.
const MaxBatch = 500

// User derives the seen-state owner from the exe.dev identity header; the
// basic-auth fallback (no header) is the owner and maps to "admin".
func User(email string) string {
	e := strings.ToLower(strings.TrimSpace(email))
	if e == "" {
		return "admin"
	}
	return e
}

// Mark records ids as seen by user for kind (job | grant). Idempotent; the
// first sighting wins so `at` stays the true first-seen time.
func Mark(ctx context.Context, db *sql.DB, kind, user string, ids []int64) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	if len(ids) > MaxBatch {
		ids = ids[:MaxBatch]
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	st, err := tx.PrepareContext(ctx, `INSERT OR IGNORE INTO seen (kind, item_id, user) VALUES (?,?,?)`)
	if err != nil {
		return 0, err
	}
	defer st.Close()
	n := 0
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		res, err := st.ExecContext(ctx, kind, id, user)
		if err != nil {
			return n, err
		}
		if k, _ := res.RowsAffected(); k > 0 {
			n++
		}
	}
	return n, tx.Commit()
}

// Set returns every item id of kind the user has seen.
func Set(ctx context.Context, db *sql.DB, kind, user string) (map[int64]bool, error) {
	rows, err := db.QueryContext(ctx, `SELECT item_id FROM seen WHERE kind = ? AND user = ?`, kind, user)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[int64]bool{}
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out[id] = true
	}
	return out, rows.Err()
}
