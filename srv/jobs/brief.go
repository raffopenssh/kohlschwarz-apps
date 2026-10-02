package jobs

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"html"
	"log/slog"
	"regexp"
	"strings"
	"time"
)

var liJSONDescRe = regexp.MustCompile(`"description":"((?:[^"\\]|\\.)*)"`)

// BriefMinScore: only postings that made the list get a fetched brief.
const BriefMinScore = 35

const briefPrompt = `You brief a former national park director (fluent EN/DE/FR) on ONE job posting or tender. He wants (A) to lead a national park / protected area, (B) senior consultancies on protected-area management, governance, finance or evaluation, or (C) a substantive post inside an Austrian Land/Bund authority that governs a national park, as a step towards directing it.

You get the title, metadata and the fetched page text (may contain navigation noise; ignore it). Answer in English with EXACTLY these lines, each "Label: text", telegraphic style (drop articles and filler), no markdown, no preamble, do not repeat the title, only what the text supports:
Org: employer exactly as named in the text, acronym first if it has one (e.g. "UNEP - United Nations Environment Programme"); "unknown" if not stated
Location: duty station(s) / country, or "remote"; "unknown" if not stated
What: type (permanent post / fixed-term / consultancy ToR / tender lot / news item), employer and unit/department as named in the text
Terms: grade or salary or contract value, duration, location, deadline if stated
Duties: core responsibilities, <=18 words
Fit: one clause on why it does or does not match A, B or C (for C name the park the authority governs, or 'no park link')
If a CANDIDATES block is present, add a fifth line "Duplicate: <id>" naming the one candidate that is the SAME vacancy (same employer, same role, same contract — just listed on another site or in another language); otherwise "Duplicate: none". Be strict: a similar role at another employer or a different grade/duty station is NOT a duplicate.
If the page text is missing or a login wall, set What to 'Page not readable; from metadata:' followed by what you can infer, and keep the other lines short.`

// BriefPending fetches the page of each ranked posting (score >= BriefMinScore)
// without a brief and asks the LLM for a short, structured summary. Respects
// the monthly budget like RankPending.
func BriefPending(ctx context.Context, db *sql.DB, maxItems int) Run {
	run := Run{Started: time.Now().UTC().Format("2006-01-02 15:04:05"), Kind: "brief", Model: Model}
	var logb strings.Builder
	cost := GetCost(ctx, db)
	budget := MaxMonthUSD()
	rows, err := Unbriefed(ctx, db, BriefMinScore, maxItems)
	if err != nil {
		run.Log = "unbriefed: " + err.Error()
		insertRun(ctx, db, run)
		return run
	}
	if len(rows) == 0 {
		run.Log = "nothing to brief"
		insertRun(ctx, db, run)
		return run
	}
	// Listed rows (score >= BriefMinScore, live, not hidden) are the pool of
	// possible duplicates; rows already merged by the deterministic rules in
	// Dedupe are not offered, so the model only sees the near-misses.
	listed, _ := List(ctx, db, false, 400)
	spent := 0.0
	for _, r := range rows {
		if ctx.Err() != nil {
			break
		}
		if cost.MonthUSD+spent >= budget {
			fmt.Fprintf(&logb, "budget reached (%s >= %s), %d left unbriefed\n", usd(cost.MonthUSD+spent), usd(budget), len(rows)-int(run.Ranked))
			break
		}
		text, src := pageText(ctx, r)
		cands := dupeCandidates(r, listed)
		var sb strings.Builder
		fmt.Fprintf(&sb, "TITLE: %s\nORG: %s\nLOCATION: %s\nPOSTED: %s\nDEADLINE: %s\nSOURCE: %s\nURL: %s\nRANKER VERDICT: score %d, %s\n",
			r.Title, r.Org, r.Location, r.Posted, r.Deadline, r.Source, r.URL, r.ScoreVal(), r.Why)
		if len(cands) > 0 {
			sb.WriteString("\nCANDIDATES (possibly the same vacancy listed elsewhere):\n")
			for _, c := range cands {
				fmt.Fprintf(&sb, "[%d] %s — %s — %s — via %s — first seen %.10s\n", c.ID, c.Title, orDash(c.Org), orDash(c.Location), c.Source, c.FirstSeen)
			}
		}
		fmt.Fprintf(&sb, "\nPAGE TEXT (%s):\n%s\n", src, text)
		out, in, nOut, err := chat(ctx, briefPrompt, sb.String(), 1400)
		c := costUSD(in, nOut)
		run.InTokens += in
		run.OutTokens += nOut
		run.CostUSD += c
		spent += c
		if err != nil {
			slog.Warn("jobs brief", "id", r.ID, "error", err)
			fmt.Fprintf(&logb, "✗ #%d %.60s: %v\n", r.ID, r.Title, err)
			continue
		}
		if dup := parseDuplicate(out); dup != 0 {
			for _, c := range cands {
				if c.ID == dup {
					if err := MarkDuplicate(ctx, db, r.ID, dup); err == nil {
						fmt.Fprintf(&logb, "· #%d is a copy of #%d (%s)\n", r.ID, dup, c.Source)
					}
				}
			}
		}
		if org, loc := parseMeta(out); org != "" || loc != "" {
			db.ExecContext(ctx, `UPDATE job_postings SET org = CASE WHEN org = '' THEN ? ELSE org END,
				location = CASE WHEN location = '' THEN ? ELSE location END WHERE id = ?`, truncate(org, 120), truncate(loc, 120), r.ID)
			if (r.Org == "" && org != "") || (r.Location == "" && loc != "") {
				fmt.Fprintf(&logb, "· #%d filled org=%q location=%q\n", r.ID, org, loc)
			}
		}
		brief := normalizeBrief(out)
		if len(brief) > 900 {
			brief = truncate(brief, 900)
		}
		if _, err := db.ExecContext(ctx, `UPDATE job_postings SET brief = ?, briefed_at = datetime('now') WHERE id = ?`, brief, r.ID); err != nil {
			fmt.Fprintf(&logb, "✗ #%d store: %v\n", r.ID, err)
			continue
		}
		run.Ranked++
		fmt.Fprintf(&logb, "✓ #%d %.60s (%s, %d in / %d out, %s)\n", r.ID, r.Title, src, in, nOut, usd(c))
	}
	run.Log = logb.String()
	if err := insertRun(ctx, db, run); err != nil {
		slog.Warn("jobs insert brief run", "error", err)
	}
	return run
}

// pageText fetches the posting page as readable text: r.jina.ai first (renders
// JS, strips chrome), raw HTML stripped as fallback, stored snippet last.
func pageText(ctx context.Context, r Row) (text, src string) {
	const max = 6000
	if r.URL != "" {
		fctx, cancel := context.WithTimeout(ctx, 45*time.Second)
		defer cancel()
		if strings.Contains(r.URL, "linkedin.com/jobs/view/") {
			// Public job pages embed the full description as JSON-LD; the
			// reader proxy only sees the logged-out shell.
			if b, err := get(fctx, r.URL); err == nil {
				if m := liJSONDescRe.FindSubmatch(b); m != nil {
					var d string
					if json.Unmarshal(append(append([]byte{'"'}, m[1]...), '"'), &d) == nil {
						if t := clean(html.UnescapeString(d)); len(t) > 100 {
							return truncate(t, max), "fetched linkedin"
						}
					}
				}
			}
			time.Sleep(6 * time.Second)
		}
		if b, err := get(fctx, "https://r.jina.ai/"+r.URL); err == nil && len(b) > 200 {
			return truncate(compactText(string(b)), max), "fetched via reader"
		}
		if b, err := get(fctx, r.URL); err == nil {
			if t := clean(string(b)); len(t) > 200 {
				return truncate(t, max), "fetched html"
			}
		}
	}
	if s := strings.TrimSpace(r.Snippet); s != "" {
		return truncate(s, max), "stored snippet only"
	}
	return "(none)", "missing"
}

// compactText collapses the markdown-ish reader output: drops link targets,
// images and blank runs so the token budget goes to actual content.
func compactText(s string) string {
	var out []string
	for _, ln := range strings.Split(s, "\n") {
		ln = strings.TrimSpace(ln)
		if ln == "" || strings.HasPrefix(ln, "![") || strings.HasPrefix(ln, "URL Source:") || strings.HasPrefix(ln, "Markdown Content:") {
			continue
		}
		ln = mdLinkRe.ReplaceAllString(ln, "$1")
		out = append(out, ln)
	}
	return strings.Join(out, "\n")
}

var dupLineRe = regexp.MustCompile(`(?im)^\W*duplicate\s*:\s*\[?#?(\d+)`)

// parseDuplicate returns the candidate id named on the "Duplicate:" line, 0 for none.
func parseDuplicate(out string) int64 {
	m := dupLineRe.FindStringSubmatch(out)
	if m == nil {
		return 0
	}
	var id int64
	fmt.Sscan(m[1], &id)
	return id
}

// dupeCandidates picks the listed rows that look like they might be the same
// vacancy as r but that Dedupe's deterministic rules do not already merge:
// score >= BriefMinScore, first seen within 60 days, title tokens overlapping
// weakly (Jaccard >= 0.3 or all tokens of the shorter title contained). Max 6,
// so the extra prompt cost stays at a few dozen tokens.
func dupeCandidates(r Row, listed []Row) []Row {
	rt := titleTokens(r.Title)
	if len(rt) < 2 {
		return nil
	}
	var pool []Row
	for _, o := range listed {
		if o.ID == r.ID || o.ScoreVal() < BriefMinScore || o.Hidden || o.ClosedAt != nil {
			continue
		}
		if r.FirstSeen != "" && o.FirstSeen != "" && daysApart(r.FirstSeen, o.FirstSeen) > 60 {
			continue
		}
		ot := titleTokens(o.Title)
		if len(ot) < 2 {
			continue
		}
		inter := 0
		for w := range rt {
			if ot[w] {
				inter++
			}
		}
		short := min(len(rt), len(ot))
		j := float64(inter) / float64(len(rt)+len(ot)-inter)
		if j >= 0.3 || (inter >= 2 && inter == short) {
			pool = append(pool, o)
		}
	}
	if len(pool) == 0 {
		return nil
	}
	// Drop pairs the deterministic rules already merge: they are hidden as "+N copies" anyway.
	merged := Dedupe(append([]Row{r}, pool...))
	keep := map[int64]bool{}
	for _, m := range merged {
		keep[m.ID] = true
	}
	var out []Row
	for _, o := range pool {
		if keep[o.ID] {
			out = append(out, o)
		}
		if len(out) == 6 {
			break
		}
	}
	return out
}

func daysApart(a, b string) float64 {
	ta, ea := time.Parse("2006-01-02 15:04:05", a)
	tb, eb := time.Parse("2006-01-02 15:04:05", b)
	if ea != nil || eb != nil {
		return 0
	}
	d := ta.Sub(tb).Hours() / 24
	if d < 0 {
		d = -d
	}
	return d
}

// parseMeta reads the Org / Location lines the brief prompt asks for; "unknown"
// and similar placeholders become "".
func parseMeta(out string) (org, loc string) {
	for _, ln := range strings.Split(out, "\n") {
		ln = strings.TrimSpace(strings.TrimLeft(ln, "-*• "))
		l := strings.ToLower(ln)
		switch {
		case strings.HasPrefix(l, "org:"):
			org = metaVal(ln[4:])
		case strings.HasPrefix(l, "location:"):
			loc = metaVal(ln[9:])
		}
	}
	return
}

func metaVal(s string) string {
	s = strings.TrimSpace(strings.Trim(strings.TrimSpace(s), `"'`))
	switch strings.ToLower(strings.TrimRight(s, ".")) {
	case "", "unknown", "n/a", "na", "none", "not stated", "unclear", "-", "–":
		return ""
	}
	return s
}

var briefLabels = []string{"What", "Terms", "Duties", "Fit"}

// normalizeBrief keeps only the four labelled lines, one per line, in order.
func normalizeBrief(out string) string {
	got := map[string]string{}
	for _, ln := range strings.Split(out, "\n") {
		ln = strings.TrimSpace(strings.TrimLeft(ln, "-*• "))
		for _, l := range briefLabels {
			if len(ln) > len(l)+1 && strings.EqualFold(ln[:len(l)], l) && ln[len(l)] == ':' {
				got[l] = strings.TrimSpace(ln[len(l)+1:])
			}
		}
	}
	if len(got) == 0 {
		return strings.TrimSpace(strings.Trim(strings.TrimSpace(out), `"`))
	}
	var lines []string
	for _, l := range briefLabels {
		if v := got[l]; v != "" {
			lines = append(lines, l+": "+v)
		}
	}
	return strings.Join(lines, "\n")
}

// BriefItem is one labelled line of a brief.
type BriefItem struct{ Label, Text string }

// BriefItems splits a stored brief into labelled items; empty when the brief
// is a plain paragraph (old format) or missing.
func (r Row) BriefItems() []BriefItem {
	var out []BriefItem
	for _, ln := range strings.Split(r.Brief, "\n") {
		if i := strings.IndexByte(ln, ':'); i > 0 && i < 12 {
			out = append(out, BriefItem{strings.TrimSpace(ln[:i]), strings.TrimSpace(ln[i+1:])})
		}
	}
	if len(out) < 2 {
		return nil
	}
	return out
}
