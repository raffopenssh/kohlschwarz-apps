package jobs

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
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
// geoBlockedHosts drop connections from cloud IP ranges; their pages are read
// through the free-proxy pool (proxy.go) instead of directly.
var geoBlockedHosts = []string{"ktn.gv.at"}

func isGeoBlocked(u string) bool {
	h := u
	if i := strings.Index(h, "://"); i >= 0 {
		h = h[i+3:]
	}
	if i := strings.IndexAny(h, "/?#"); i >= 0 {
		h = h[:i]
	}
	for _, g := range geoBlockedHosts {
		if h == g || strings.HasSuffix(h, "."+g) {
			return true
		}
	}
	return false
}

// readPage tries, in order: the proxy pool (geo-blocked hosts only), the
// r.jina.ai reader, the raw HTML, and finally the Wayback Machine's latest
// snapshot (covers 404/410 "gone" pages). Returns "" when nothing is readable.
// readerErrRe: r.jina.ai answers 200 with a stub when the target itself failed.
var readerErrRe = regexp.MustCompile(`(?m)^Warning: Target URL returned error \d{3}`)

var mainRe = regexp.MustCompile(`(?is)<(main|article)\b[^>]*>(.*?)</(main|article)>`)

// focusText drops site chrome ahead of the posting itself: everything before
// the first occurrence of the title (or its first 40 chars) is navigation.
// Saves ~⅓ of the prompt tokens on portal pages without a <main> element.
func focusText(text, title string) string {
	// Key = leading words of the title up to ~24 chars: long enough to be
	// unique on the page, short enough to survive small title variants.
	key := ""
	for _, w := range strings.Fields(strings.ToLower(title)) {
		if key != "" {
			key += " "
		}
		key += w
		if len(key) >= 24 {
			break
		}
	}
	if len(key) < 8 {
		return text
	}
	i := strings.Index(strings.ToLower(text), key)
	if i <= 0 || len(text)-i < 300 {
		return text
	}
	return text[i:]
}

// looksLikeHTML rejects what free proxies like to answer instead of the
// target: echo pages, login walls, bare error strings.
func looksLikeHTML(b []byte) bool {
	l := strings.ToLower(string(b))
	return strings.Contains(l, "<html") && strings.Contains(l, "</html>") && strings.Contains(l, "<title") &&
		!strings.Contains(l, "remote_addr") && !strings.Contains(l, "http_user-agent")
}

// mainText strips scripts, styles and site chrome from raw HTML and prefers
// the <main>/<article> block when there is one with real content.
func mainText(h string) string {
	h = stripBlocks(h, []string{"script", "style", "noscript", "nav", "header", "footer", "svg"})
	if m := mainRe.FindStringSubmatch(h); m != nil {
		if t := clean(m[2]); len(t) > 200 {
			return t
		}
	}
	return clean(h)
}

// stripBlocks removes <tag …>…</tag> blocks (non-nested, case-insensitive).
func stripBlocks(h string, tags []string) string {
	for _, t := range tags {
		re := regexp.MustCompile(`(?is)<` + t + `\b[^>]*>.*?</` + t + `\s*>`)
		h = re.ReplaceAllString(h, " ")
	}
	return h
}

func readPage(ctx context.Context, u string, max int) (text, src string) {
	readable := func(b []byte) bool { return looksLikeHTML(b) && len(mainText(string(b))) > 200 }
	try := func(fetch func() ([]byte, error), tag string, fmtText func([]byte) string) bool {
		b, err := fetch()
		if err != nil {
			return false
		}
		if t := fmtText(b); len(t) > 200 {
			text, src = truncate(t, max), tag
			return true
		}
		return false
	}
	htmlText := func(b []byte) string { return mainText(string(b)) }
	if isGeoBlocked(u) {
		if try(func() ([]byte, error) { return getViaProxy(ctx, u, readable) }, "fetched via proxy", htmlText) {
			return
		}
	} else {
		if try(func() ([]byte, error) {
			b, err := get(ctx, "https://r.jina.ai/"+u)
			if err == nil && readerErrRe.Match(b) {
				return nil, errors.New("reader: target error")
			}
			return b, err
		}, "fetched via reader", func(b []byte) string { return compactText(string(b)) }) {
			return
		}
		if try(func() ([]byte, error) { return get(ctx, u) }, "fetched html", htmlText) {
			return
		}
	}
	if try(func() ([]byte, error) { return get(ctx, "https://web.archive.org/web/2id_/"+u) }, "fetched from wayback", htmlText) {
		return
	}
	return "", ""
}

func FetchPageText(ctx context.Context, u string, max int) (text, src string) {
	if u == "" {
		return "(none)", "missing"
	}
	if t, src := readPage(ctx, u, max); t != "" {
		return t, src
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
