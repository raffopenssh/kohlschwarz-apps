package jobs

import (
	"context"
	"database/sql"
	"regexp"
	"sort"
	"strings"
)

// Dedupe collapses the same job that appears under several URLs / titles /
// languages into one row. Two postings are the same job when
//
//  1. their canonical URL is identical (LinkedIn job id across country
//     subdomains, tracking params), or
//  2. their normalised org|title key is identical, or
//  3. they are in the same org bucket (org, or source when org is blank) and
//     their synonym-normalised title token sets overlap strongly (Jaccard ≥ 0.6),
//     e.g. "Directeur.trice du Parc National de Conkouati-Douli (PNCD)" vs
//     "Conkouati-Douli Park Manager", or
//  4. they share a non-empty location, their titles overlap strongly and their
//     orgs are related (one is a word-prefix of the other: "WWF" / "WWF Cities"),
//     e.g. LinkedIn copies posted from a sub-brand account, or
//  5. their normalised titles are identical, their orgs are related ("UNEP"
//     on one board, "UNEP - United Nations Environment Programme" on another;
//     orgs filled in by the brief step make this fire more often) and their
//     locations do not conflict (equal, or one of them blank).
//
// The input order (best score first, as List orders) decides which row
// represents the group; the group inherits earliest first_seen, latest
// last_seen, any deadline / org / location / reported_at its copies have, and
// Dupes = number of collapsed copies.
func Dedupe(rows []Row) []Row { return DedupeWith(rows, nil) }

// DedupeWith is Dedupe plus LLM-confirmed pairs (dedupe_pairs, written by
// BriefPending): the two ids of each pair are merged if both are present.
func DedupeWith(rows []Row, pairs [][2]int64) []Row {
	n := len(rows)
	if n == 0 {
		return rows
	}
	uf := newUF(n)
	if len(pairs) > 0 {
		byID := make(map[int64]int, n)
		for i, r := range rows {
			byID[r.ID] = i
		}
		for _, p := range pairs {
			if i, ok := byID[p[0]]; ok {
				if j, ok := byID[p[1]]; ok {
					uf.union(i, j)
				}
			}
		}
	}
	byURL := map[string]int{}
	byKey := map[string]int{}
	buckets := map[string][]int{}
	locs := map[string][]int{}
	titles := map[string][]int{}
	toks := make([]map[string]bool, n)
	orgs := make([]string, n)
	for i, r := range rows {
		if u := CanonicalURL(r.URL); u != "" {
			if j, ok := byURL[u]; ok {
				uf.union(i, j)
			} else {
				byURL[u] = i
			}
		}
		k := DedupeKey(r)
		if j, ok := byKey[k]; ok {
			uf.union(i, j)
		} else {
			byKey[k] = i
		}
		toks[i] = titleTokens(r.Title)
		b := orgBucket(r)
		buckets[b] = append(buckets[b], i)
		orgs[i] = normOrg(r.Org)
		if l := normLoc(r.Location); l != "" && orgs[i] != "" {
			locs[l] = append(locs[l], i)
		}
		if orgs[i] != "" {
			if t := k[strings.IndexByte(k, '|')+1:]; len(t) >= 12 {
				titles[t] = append(titles[t], i)
			}
		}
	}
	for _, idx := range titles {
		for a := 0; a < len(idx); a++ {
			for b := a + 1; b < len(idx); b++ {
				i, j := idx[a], idx[b]
				if uf.find(i) == uf.find(j) || !orgRelated(orgs[i], orgs[j]) {
					continue
				}
				li, lj := normLoc(rows[i].Location), normLoc(rows[j].Location)
				if li == "" || lj == "" || li == lj {
					uf.union(i, j)
				}
			}
		}
	}
	for _, idx := range buckets {
		for a := 0; a < len(idx); a++ {
			for b := a + 1; b < len(idx); b++ {
				i, j := idx[a], idx[b]
				if uf.find(i) == uf.find(j) {
					continue
				}
				if similar(toks[i], toks[j]) {
					uf.union(i, j)
				}
			}
		}
	}
	for _, idx := range locs {
		for a := 0; a < len(idx); a++ {
			for b := a + 1; b < len(idx); b++ {
				i, j := idx[a], idx[b]
				if uf.find(i) == uf.find(j) || !orgRelated(orgs[i], orgs[j]) {
					continue
				}
				if similar(toks[i], toks[j]) {
					uf.union(i, j)
				}
			}
		}
	}
	// Group representative = lowest index (rows come best-first).
	rep := map[int]int{}
	var out []Row
	for i := range rows {
		root := uf.find(i)
		if p, ok := rep[root]; ok {
			merge(&out[p], rows[i])
			continue
		}
		rep[root] = len(out)
		out = append(out, rows[i])
	}
	return out
}

func merge(dst *Row, src Row) {
	dst.Dupes++
	if dst.latestFirstSeen == "" {
		dst.latestFirstSeen = dst.FirstSeen
	}
	if src.FirstSeen > dst.latestFirstSeen {
		dst.latestFirstSeen = src.FirstSeen
	}
	if src.FirstSeen != "" && (dst.FirstSeen == "" || src.FirstSeen < dst.FirstSeen) {
		dst.FirstSeen = src.FirstSeen
	}
	// Signals: the group is live if any copy is live; keep the richest data.
	if src.ClosedAt == nil {
		dst.ClosedAt = nil
	}
	if src.Applicants != nil && (dst.Applicants == nil || *src.Applicants > *dst.Applicants) {
		dst.Applicants = src.Applicants
	}
	if dst.Salary == "" {
		dst.Salary = src.Salary
	}
	dst.Reposted = dst.Reposted || src.Reposted
	dst.Pinned = dst.Pinned || src.Pinned
	dst.Events = append(dst.Events, src.Events...)
	if src.LastSeen > dst.LastSeen {
		dst.LastSeen = src.LastSeen
	}
	if dst.Deadline == "" {
		dst.Deadline = src.Deadline
	}
	if dst.Org == "" {
		dst.Org = src.Org
	}
	if dst.Location == "" {
		dst.Location = src.Location
	}
	if dst.ReportedAt == nil {
		dst.ReportedAt = src.ReportedAt
	}
	if dst.Why == "" {
		dst.Why = src.Why
	}
	if dst.Score == nil {
		dst.Score = src.Score
	}
}

// DedupeKey collapses the same job posted under several URLs (LinkedIn per-country
// copies, EN/FR duplicates on the same careers page) to one key.
func DedupeKey(r Row) string {
	t := strings.ToLower(r.Title)
	t = nonAlnum.ReplaceAllString(t, " ")
	t = strings.Join(strings.Fields(t), " ")
	o := strings.ToLower(strings.Join(strings.Fields(nonAlnum.ReplaceAllString(r.Org, " ")), " "))
	return o + "|" + t
}

var (
	nonAlnum = regexp.MustCompile(`[^\p{L}\p{N}]+`)
	// gender markers & short acronyms in brackets: (m/w/d) (h/f) (gn) (PNCD) (f/m/x) (REMOTE)
	parenNoise  = regexp.MustCompile(`(?i)\((?:[mwdfhxgn/\s*]{1,9}|[A-Z0-9]{2,8}|remote|hybrid)\)`)
	linkedinID  = regexp.MustCompile(`linkedin\.com/jobs/view/(?:[^/?#]*?-)?(\d{6,})`)
	reliefwebID = regexp.MustCompile(`reliefweb\.int/job/(\d+)`)
)

// CanonicalURL strips host country prefixes and tracking so identical jobs match.
func CanonicalURL(u string) string {
	u = strings.TrimSpace(strings.ToLower(u))
	if u == "" {
		return ""
	}
	if m := linkedinID.FindStringSubmatch(u); m != nil {
		return "linkedin:" + m[1]
	}
	if m := reliefwebID.FindStringSubmatch(u); m != nil {
		return "reliefweb:" + m[1]
	}
	u = strings.TrimPrefix(u, "http://")
	u = strings.TrimPrefix(u, "https://")
	u = strings.TrimPrefix(u, "www.")
	if i := strings.IndexAny(u, "?#"); i >= 0 {
		u = u[:i]
	}
	return strings.TrimRight(u, "/")
}

func normOrg(o string) string {
	return strings.ToLower(strings.Join(strings.Fields(nonAlnum.ReplaceAllString(o, " ")), " "))
}

func normLoc(l string) string {
	return strings.ToLower(strings.Join(strings.Fields(nonAlnum.ReplaceAllString(l, " ")), " "))
}

// orgRelated: equal, or one org is a word-prefix of the other ("wwf" / "wwf cities",
// "unep" / "unep wcmc"). Both must be non-empty.
func orgRelated(a, b string) bool {
	if a == "" || b == "" {
		return false
	}
	if a == b {
		return true
	}
	if len(a) > len(b) {
		a, b = b, a
	}
	return strings.HasPrefix(b, a+" ")
}

func orgBucket(r Row) string {
	o := normOrg(r.Org)
	if o != "" {
		return "o:" + o
	}
	// No org: only compare within the same source family (e.g. "Noé · nous rejoindre").
	s := r.Source
	if i := strings.Index(s, " · "); i >= 0 {
		s = s[:i]
	}
	return "s:" + strings.ToLower(s)
}

// synonyms maps EN/DE/FR role & object words to one token so multilingual
// copies of a posting compare equal.
var synonyms = map[string]string{
	"director": "director", "directeur": "director", "directrice": "director", "direktor": "director", "direktorin": "director",
	"manager": "director", "geschäftsführer": "director", "geschaftsfuhrer": "director", "geschäftsführerin": "director", "ceo": "director",
	"conservateur": "director", "conservatrice": "director", "leiter": "director", "leiterin": "director", "leitung": "director", "head": "director",
	"park": "park", "parc": "park", "parcs": "park", "parks": "park", "nationalpark": "park", "nationalparks": "park",
	"national": "national", "nationale": "national", "nationaler": "national",
	"protected": "pa", "protégée": "pa", "protegee": "pa", "protégées": "pa", "schutzgebiet": "pa", "schutzgebiete": "pa",
	"area": "area", "areas": "area", "aire": "area", "aires": "area", "gebiet": "area",
	"senior": "senior", "principal": "senior",
	"consultant": "consultant", "consultancy": "consultant", "consultance": "consultant", "berater": "consultant", "beratung": "consultant", "expert": "consultant", "experte": "consultant",
	"advisor": "consultant", "adviser": "consultant", "conseiller": "consultant",
	"programme": "program", "program": "program", "programm": "program",
	"project": "project", "projet": "project", "projekt": "project",
	"republic": "", "republique": "", "république": "", "the": "", "of": "", "and": "", "for": "", "in": "", "at": "", "to": "", "a": "", "an": "",
	"du": "", "de": "", "des": "", "la": "", "le": "", "les": "", "et": "", "der": "", "die": "", "das": "", "und": "", "für": "", "fur": "", "im": "", "am": "", "en": "", "au": "", "aux": "",
	"m": "", "w": "", "d": "", "f": "", "h": "", "x": "", "gn": "", "remote": "", "hybrid": "",
}

func titleTokens(title string) map[string]bool {
	t := parenNoise.ReplaceAllString(title, " ")
	t = strings.ToLower(t)
	// directeur.trice → directeur ; direktor*in → direktor
	t = strings.NewReplacer(".trice", "", "-trice", "", "/trice", "", "(trice)", "", "*in", "", "/in", "", ":in", "", "_in", "", "(in)", "", "·e", "", ".e ", " ").Replace(t)
	out := map[string]bool{}
	for _, w := range strings.Fields(nonAlnum.ReplaceAllString(t, " ")) {
		if s, ok := synonyms[w]; ok {
			w = s
		}
		if w == "" || len(w) < 2 {
			continue
		}
		out[w] = true
	}
	return out
}

func similar(a, b map[string]bool) bool {
	if len(a) < 2 || len(b) < 2 {
		return false
	}
	inter := 0
	for w := range a {
		if b[w] {
			inter++
		}
	}
	union := len(a) + len(b) - inter
	if union == 0 {
		return false
	}
	j := float64(inter) / float64(union)
	// Small titles: require near-total overlap; larger titles: 0.6.
	if len(a) <= 3 || len(b) <= 3 {
		return j >= 0.75
	}
	return j >= 0.6
}

// --- tiny union-find --------------------------------------------------------

type uf struct{ p []int }

func newUF(n int) *uf {
	p := make([]int, n)
	for i := range p {
		p[i] = i
	}
	return &uf{p}
}
func (u *uf) find(i int) int {
	for u.p[i] != i {
		u.p[i] = u.p[u.p[i]]
		i = u.p[i]
	}
	return i
}
func (u *uf) union(a, b int) {
	ra, rb := u.find(a), u.find(b)
	if ra == rb {
		return
	}
	if ra > rb { // keep lowest index as root (best-ranked row wins)
		ra, rb = rb, ra
	}
	u.p[rb] = ra
}

// sortedTokens is for tests/debugging.
func sortedTokens(m map[string]bool) []string {
	var s []string
	for k := range m {
		s = append(s, k)
	}
	sort.Strings(s)
	return s
}

// MarkDuplicate records an LLM verdict that postings a and b are the same vacancy.
func MarkDuplicate(ctx context.Context, db *sql.DB, a, b int64) error {
	if a == b {
		return nil
	}
	if a > b {
		a, b = b, a
	}
	_, err := db.ExecContext(ctx, `INSERT OR REPLACE INTO dedupe_pairs (a_id, b_id, same) VALUES (?,?,1)`, a, b)
	return err
}

// DuplicatePairs loads the confirmed pairs for DedupeWith.
func DuplicatePairs(ctx context.Context, db *sql.DB) [][2]int64 {
	rs, err := db.QueryContext(ctx, `SELECT a_id, b_id FROM dedupe_pairs WHERE same = 1`)
	if err != nil {
		return nil
	}
	defer rs.Close()
	var out [][2]int64
	for rs.Next() {
		var p [2]int64
		if rs.Scan(&p[0], &p[1]) == nil {
			out = append(out, p)
		}
	}
	return out
}
