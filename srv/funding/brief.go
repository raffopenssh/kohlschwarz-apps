package funding

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"srv.exe.dev/srv/jobs"
)

// BriefMinScore: only entries that matter get a fetched brief.
const BriefMinScore = 35

// BriefLabels are the five lines of a funding brief, in order.
var BriefLabels = []string{"What", "Money", "Eligible", "Timeline", "Next"}

const briefPrompt = `You brief ONE person on ONE funding opportunity (grant, prize, accelerator, incubator, investor or credits programme). He is a solo founder, a former national park director (Chinko, CAR co-founder; 12 years in African protected areas) with an Oxford postgraduate diploma; no company yet (natural person; would incorporate when a programme requires it – Austria by default, but a UK, Swedish or Central African Republic company OR a registered NGO/association is possible if the opportunity needs it), no revenue, no co-founder yet; works between Vienna and the Central African Republic.
He runs TWO ventures; the record tells you which one (PROJECT line):
- PALANTIR = "Veridical Earth" ("Palantir for land use", YC application): land-use governance / earth-observation data platform (LIDAR, Copernicus, VIIRS fusion, flood/fire/drought, protected areas), plus free open-source apps on kohlschwarz.at and a free protected-area management app for African parks. Vienna-based; Austrian/EU/UK programmes matter.
- NGI = "Landscape Governance Initiative": Central-African, company-like vehicle (not a conventional NGO, but will register one if needed) that trains and equips a two-person Landscape Oversight Unit inside the CAR Ministry of Environment (remote sensing, control room, oversight of NGO-managed parks) and drafts a National Landscape Plan; budget ~USD 585k year 1, ~650k year 2; later all Africa Keystone Partnership countries. Vienna is irrelevant for NGI; what matters is CAR eligibility, government-capacity/PA-governance themes and whether a company or a new NGO can receive the money.
- BOTH = judge for whichever venture fits the opportunity better and say which.

You get the curated record (name, amount, deadline, eligibility notes, scored 0-100 by hand) and the fetched official page text (may contain navigation noise; ignore it). Answer in English with EXACTLY these five lines, each "Label: text", telegraphic style (drop articles and filler), no markdown, no preamble, do not repeat the name, only what the page or record supports. Prefer the PAGE for facts and say so when it contradicts the record (e.g. "page says …"):
What: funder and instrument; what is funded (activities, costs); form (grant / equity / voucher / prize / programme)
Money: amount or range, funding rate, own-contribution, in-kind extras (coaching, office, credits)
Eligible: who may apply – natural persons? company age limit? location/seat requirement (and whether an AT, UK, SE or CAR entity – company or NGO – would satisfy it)? team size? sector; the ONE blocker for him if any
Timeline: call status per page (open / closed / rolling / next round), concrete dates, decision time, steps (registration, pitch, interview)
Next: the single most useful concrete action for him this month, <=20 words
If the page text is missing or unreadable, set What to 'Page not readable; from record:' and keep other lines short.`

// Unbriefed returns entries worth briefing that have no brief yet. Briefs are
// never refreshed automatically (owner uses "re-brief all"), so the daily hook
// costs nothing unless new entries were seeded.
func Unbriefed(ctx context.Context, db *sql.DB, min, limit int) ([]Entry, error) {
	rows, err := db.QueryContext(ctx, `SELECT `+cols+` FROM funding WHERE score >= ? AND status NOT IN ('skip','rejected','won') AND briefed_at IS NULL ORDER BY
		CASE WHEN deadline <> '' AND deadline >= date('now') THEN 0 ELSE 1 END, score DESC LIMIT ?`, min, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scan(rows)
}

// ClearBriefs drops all briefs so the next pass regenerates them (owner action).
func ClearBriefs(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `UPDATE funding SET brief = '', briefed_at = NULL`)
	return err
}

// BriefPending fetches the official page of each unbriefed entry (score >=
// BriefMinScore) and asks the LLM for a five-line structured brief. Books into
// job_runs (kind "fbrief") and respects the shared monthly budget.
func BriefPending(ctx context.Context, db *sql.DB, maxItems int) jobs.Run {
	run := jobs.Run{Started: time.Now().UTC().Format("2006-01-02 15:04:05"), Kind: "fbrief", Model: jobs.Model}
	var logb strings.Builder
	cost := jobs.GetCost(ctx, db)
	budget := jobs.MaxMonthUSD()
	rows, err := Unbriefed(ctx, db, BriefMinScore, maxItems)
	if err != nil {
		run.Log = "unbriefed: " + err.Error()
		jobs.InsertRun(ctx, db, run)
		return run
	}
	if len(rows) == 0 {
		run.Log = "nothing to brief"
		jobs.InsertRun(ctx, db, run)
		return run
	}
	spent := 0.0
	for _, e := range rows {
		if ctx.Err() != nil {
			break
		}
		if cost.MonthUSD+spent >= budget {
			fmt.Fprintf(&logb, "budget reached (%s >= %s), %d left unbriefed\n", jobs.USD(cost.MonthUSD+spent), jobs.USD(budget), len(rows)-int(run.Ranked))
			break
		}
		fctx, cancel := context.WithTimeout(ctx, 45*time.Second)
		text, src := jobs.FetchPageText(fctx, e.URL, 6000)
		cancel()
		var sb strings.Builder
		fmt.Fprintf(&sb, "NAME: %s\nPROJECT: %s\nKIND: %s · TRACK: %s\nAMOUNT (record): %s\nDEADLINE (record): %s\nELIGIBILITY (record): %s\nCURATOR NOTE: %s\nHAND SCORE: %d — %s\nURL: %s\nTODAY: %s\n\nPAGE TEXT (%s):\n%s\n",
			e.Name, strings.ToUpper(e.ProjKey()), e.Kind, e.Track, e.Amount, e.DL(), e.Eligibility, e.Note, e.Score, e.Why, e.URL, time.Now().UTC().Format("2006-01-02"), src, text)
		out, in, nOut, err := jobs.Chat(ctx, briefPrompt, sb.String(), 1600)
		c := jobs.CostUSD(in, nOut)
		run.InTokens += in
		run.OutTokens += nOut
		run.CostUSD += c
		spent += c
		if err != nil {
			slog.Warn("funding brief", "id", e.ID, "error", err)
			fmt.Fprintf(&logb, "✗ #%d %.60s: %v\n", e.ID, e.Name, err)
			continue
		}
		brief := jobs.NormalizeLabelled(out, BriefLabels)
		if brief == "" {
			fmt.Fprintf(&logb, "✗ #%d %.60s: unparseable reply %.80q\n", e.ID, e.Name, out)
			continue
		}
		if len(brief) > 1200 {
			brief = brief[:1200]
		}
		if _, err := db.ExecContext(ctx, `UPDATE funding SET brief = ?, briefed_at = datetime('now') WHERE id = ?`, brief, e.ID); err != nil {
			fmt.Fprintf(&logb, "✗ #%d store: %v\n", e.ID, err)
			continue
		}
		run.Ranked++
		fmt.Fprintf(&logb, "✓ #%d %.60s (%s, %d in / %d out, %s)\n", e.ID, e.Name, src, in, nOut, jobs.USD(c))
	}
	run.Log = logb.String()
	if err := jobs.InsertRun(ctx, db, run); err != nil {
		slog.Warn("funding insert brief run", "error", err)
	}
	return run
}

// BriefItem is one labelled line of a brief.
type BriefItem struct{ Label, Text string }

// BriefItems splits the stored brief into labelled items (nil when missing).
func (e Entry) BriefItems() []BriefItem {
	var out []BriefItem
	for _, ln := range strings.Split(e.Brief, "\n") {
		if i := strings.IndexByte(ln, ':'); i > 0 && i < 12 {
			out = append(out, BriefItem{strings.TrimSpace(ln[:i]), strings.TrimSpace(ln[i+1:])})
		}
	}
	if len(out) < 2 {
		return nil
	}
	return out
}

// BriefTeaser is the collapsed one-liner: the Next line if present, else What.
func (e Entry) BriefTeaser() string {
	items := e.BriefItems()
	for _, it := range items {
		if it.Label == "Next" {
			return it.Text
		}
	}
	if len(items) > 0 {
		return items[0].Text
	}
	return ""
}

// BriefedDate is the YYYY-MM-DD the brief was generated ("" when none).
func (e Entry) BriefedDate() string {
	if e.BriefedAt == nil || len(*e.BriefedAt) < 10 {
		return ""
	}
	return (*e.BriefedAt)[:10]
}
