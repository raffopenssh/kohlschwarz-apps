package jobs

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

// MODELAB=1 go test ./srv/jobs -run TestModelAB -v -count=1
// Re-ranks 12 rows scored by the old model with each candidate and prints
// side-by-side scores + cost. Read-only on db.sqlite3; skipped by default.
func TestModelAB(t *testing.T) {
	if os.Getenv("MODELAB") == "" {
		t.Skip("set MODELAB=1")
	}
	db, err := sql.Open("sqlite", "../../db.sqlite3?mode=ro")
	if err != nil {
		t.Fatal(err)
	}
	rs, err := db.Query(`SELECT id, title, org, location, posted, deadline, source, snippet, score, kind, region, why FROM job_postings
		WHERE score IS NOT NULL AND scored_at < '2026-09-26' AND id IN (235,952,966,939,969,264,495,937,232,286,907,753)`)
	if err != nil {
		t.Fatal(err)
	}
	var rows []Row
	base := map[int64]string{}
	for rs.Next() {
		var r Row
		var sc int
		var k, reg, why string
		var deadline, posted sql.NullString
		if err := rs.Scan(&r.ID, &r.Title, &r.Org, &r.Location, &posted, &deadline, &r.Source, &r.Snippet, &sc, &k, &reg, &why); err != nil {
			t.Fatal(err)
		}
		r.Posted, r.Deadline = posted.String, deadline.String
		rows = append(rows, r)
		base[r.ID] = fmt.Sprintf("%3d %-11s %-7s %s", sc, k, reg, why)
	}
	models := []struct {
		m       string
		in, out float64
	}{
		{"fireworks/gpt-oss-120b", 0.15, 0.60},
		{"fireworks/nemotron-lightning-3p5-30b-a3b", 0.05, 0.20},
		{"fireworks/glm-5p3-flash", 0.15, 0.50},
	}
	for _, m := range models {
		Model = m.m
		t0 := time.Now()
		res, in, out, err := rankBatch(context.Background(), rows, "")
		fmt.Printf("\n===== %s  in=%d out=%d cost=$%.5f  %s  err=%v  n=%d\n", m.m, in, out, float64(in)*m.in/1e6+float64(out)*m.out/1e6, time.Since(t0).Round(time.Second), err, len(res))
		for _, x := range res {
			fmt.Printf("id=%d  glimmer: %s\n        new:  %3d %-11s %-7s %s\n", x.ID, base[x.ID], x.Score, x.Kind, x.Region, x.Why)
		}
		_ = json.Marshal
	}
}
