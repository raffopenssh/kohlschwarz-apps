package jobs

import (
	"context"
	"database/sql"
	"strings"
)

// Shared LLM plumbing for the other radars (funding). Everything books into
// job_runs so the single monthly budget (MaxMonthUSD) covers all LLM use.

// Chat is one completion against the gateway; returns text, prompt tokens, completion tokens.
func Chat(ctx context.Context, system, user string, maxTokens int) (string, int64, int64, error) {
	return chat(ctx, system, user, maxTokens)
}

// CostUSD prices a completion.
func CostUSD(in, out int64) float64 { return costUSD(in, out) }

// USD formats a dollar amount the way the run log does.
func USD(v float64) string { return usd(v) }

// InsertRun books a finished run (any Kind) into job_runs.
func InsertRun(ctx context.Context, db *sql.DB, r Run) error { return insertRun(ctx, db, r) }

// FetchPageText returns the readable text of a web page (reader proxy first,
// stripped HTML as fallback) truncated to max runes, plus a source tag.
func FetchPageText(ctx context.Context, u string, max int) (text, src string) {
	if u == "" {
		return "(none)", "missing"
	}
	if b, err := get(ctx, "https://r.jina.ai/"+u); err == nil && len(b) > 200 {
		return truncate(compactText(string(b)), max), "fetched via reader"
	}
	if b, err := get(ctx, u); err == nil {
		if t := clean(string(b)); len(t) > 200 {
			return truncate(t, max), "fetched html"
		}
	}
	return "(page not readable)", "unreadable"
}

// NormalizeLabelled keeps only "Label: text" lines for the given labels, in order.
func NormalizeLabelled(out string, labels []string) string {
	got := map[string]string{}
	for _, ln := range strings.Split(out, "\n") {
		ln = strings.TrimSpace(strings.TrimLeft(ln, "-*• "))
		for _, l := range labels {
			if len(ln) > len(l)+1 && strings.EqualFold(ln[:len(l)], l) && ln[len(l)] == ':' {
				got[l] = strings.TrimSpace(strings.Trim(strings.TrimSpace(ln[len(l)+1:]), "*"))
			}
		}
	}
	var lines []string
	for _, l := range labels {
		if v := got[l]; v != "" {
			lines = append(lines, l+": "+v)
		}
	}
	return strings.Join(lines, "\n")
}
