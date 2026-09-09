package seen

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func TestMarkSet(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if _, err := db.ExecContext(ctx, `CREATE TABLE seen (kind TEXT NOT NULL, item_id INTEGER NOT NULL, user TEXT NOT NULL, at TEXT NOT NULL DEFAULT (datetime('now')), PRIMARY KEY (kind, item_id, user))`); err != nil {
		t.Fatal(err)
	}
	n, err := Mark(ctx, db, "job", "a@x", []int64{1, 2, 0, -5, 2})
	if err != nil || n != 2 {
		t.Fatalf("Mark: n=%d err=%v (want 2, nil)", n, err)
	}
	n, _ = Mark(ctx, db, "job", "a@x", []int64{2, 3}) // 2 already seen → idempotent
	if n != 1 {
		t.Fatalf("second Mark n=%d, want 1", n)
	}
	Mark(ctx, db, "grant", "a@x", []int64{1})
	Mark(ctx, db, "job", "b@x", []int64{9})

	got, err := Set(ctx, db, "job", "a@x")
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []int64{1, 2, 3} {
		if !got[id] {
			t.Errorf("job %d not seen for a@x: %v", id, got)
		}
	}
	if len(got) != 3 || got[9] {
		t.Errorf("per-user/per-kind isolation broken: %v", got)
	}
	if g, _ := Set(ctx, db, "grant", "a@x"); len(g) != 1 || !g[1] {
		t.Errorf("grant set: %v", g)
	}
	if g, _ := Set(ctx, db, "job", "nobody"); len(g) != 0 {
		t.Errorf("unknown user should have empty set: %v", g)
	}
	if n, err := Mark(ctx, db, "job", "a@x", nil); n != 0 || err != nil {
		t.Errorf("empty Mark: %d %v", n, err)
	}
}

func TestUser(t *testing.T) {
	if User("") != "admin" || User("  A@B.c ") != "a@b.c" {
		t.Fatal("User normalisation")
	}
}
