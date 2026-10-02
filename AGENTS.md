# AGENTS.md — kohlschwarz.at (Go, sqlite, no JS framework)

Public site (app showcase, DE/EN) + two private admin radars: **jobs** (`/admin/jobs`) and **funding** (`/admin/funding`).
Live: https://kohlschwarz.exe.xyz (exe.dev login) · https://kohlschwarz.at (Cloudflare → basic auth fallback).

## Build / run / deploy
```
make build && sudo systemctl restart srv      # binary = ./server (NOT ./srv — that's a source dir)
go test ./srv/...                              # dedupe tests etc.
journalctl -u srv -n 50 --no-pager
```
- `.env` holds `ADMIN_PASSWORD` (basic auth user `admin`); optional `ADMIN_EMAIL`, `ADMIN_URL`, `JOBS_LLM_BUDGET_USD`.
- Local check: `curl -u admin:$(grep ADMIN_PASSWORD .env|cut -d= -f2) localhost:8000/admin/jobs`. Headless browser can't send basic auth → save HTML to a tmp dir and serve with `busybox httpd` on a free port (8765 is often taken).
- If restart loops with "address already in use": `sudo ss -ltnp | grep :8000` and kill the orphan `server`.
- Templates are parsed per request (`renderTemplate`, FuncMap: `runAgo`) → template/CSS/JS edits need no rebuild, Go edits do. Bump `?v=` on `radar.css`/`radar.js` links after changes.
- Migrations: `db/migrations/NNN-name.sql`, applied at startup; end with `INSERT OR IGNORE INTO migrations …`. Latest: 017 (dedupe_pairs).
- Commit with `git add <files>` explicitly (blind `git add -A` is blocked).

## Layout
```
cmd/srv/main.go            entrypoint
srv/srv.go                 routes (mux.HandleFunc), auth (isAdmin/requireAuth), renderTemplate, scheduler start
srv/jobs_handlers.go       /admin/jobs* handlers; background runs via startBackground + jobs.Current
srv/funding_handlers.go    /admin/funding*
srv/jobs/                  sources.go (71 feeds) · fetch.go · match.go (keyword filter) · rank.go (LLM, budget)
                           dedupe.go (union-find: canonical URL + org|title + synonym Jaccard) · report.go (weekly email, Scheduler)
                           signals.go (hiring-difficulty tags, CheckPending page re-check, no LLM) · status.go · store.go (Row, Run, events)
srv/feedback/              feedback.go: Reasons (trash-reason chips), Log/Recent/ReasonCounts/Totals, PromptHints (owner verdicts → rank prompt)
srv/seen/                  seen.go: per-user read state (Mark/Set/User); seen_handlers.go POST /admin/{jobs,funding}/seen {ids:[…]} (viewers allowed)
srv/feedback_handlers.go   /admin/{jobs,funding}/{vote,note,pin}/{id}, jobs/hide, funding/status — JSON when Accept: application/json, else redirect
srv/templates/_react.html  shared {{define "react"}} partial (star · thumbs · note · status · trash · one-time "why?"); icon-only, lucide paths, data-tip tooltips; renderTemplate globs _*.html
srv/funding/               seed.go (hand-curated entries) · verified.go · store.go · verification-*/ (raw notes)
srv/templates/*.html       jobs.html + funding.html share static/radar.css + radar.js (chips, collapse, live status poll)
db/                        sqlite open + migrations; dbgen = sqlc output for public-site tables only (radars use raw sql)
```

## Jobs radar conventions
- Pipeline: `FetchAll` (upsert by url) → `RankPending` (glm-5p3-flash via exe.dev gateway at `Effort` medium, cap `MaxMonthUSD` default $0.30; model A/B: `MODELAB=1 go test ./srv/jobs -run TestModelAB`) → `BriefPending` (brief.go: fetches page via r.jina.ai / LinkedIn JSON-LD, 4-line What/Terms/Duties/Fit brief for score ≥ 35, shown as collapsed `<details>` and in email picks) → `CheckPending` (signals.go: every 3 d re-reads pages of score ≥ 35 rows — closed markers/404 → `closed_at`, LinkedIn applicants, salary, JSON-LD dates; LinkedIn `validThrough` is ignored, it's synthetic) → `WeeklyReport` (Mon 04:00 UTC; daily fetch). Only one of fetch/rank/brief/check/email runs at a time (`jobs.Current.Start/Finish`); UI polls `/admin/jobs/status.json`.
- List is deduped at render time (`jobs.DedupeWith(rows, jobs.DuplicatePairs)`), never in the DB; merged copies show as “+N copies merged”. Rules: canonical URL · org|title key · title Jaccard within org bucket · related org + same location · identical title + related org + non-conflicting location. Add multilingual role words to `synonyms` in dedupe.go, add a case to `dedupe_test.go`. The brief prompt additionally returns `Org:`/`Location:` (fill blank DB fields, so the algorithmic rules see them) and, given up to 6 near-miss `CANDIDATES`, a `Duplicate: <id>` line → `dedupe_pairs` (no extra LLM call). Impactpool has its own card parser (`fetchImpactpool`), the generic anchor scraper mangled title+org+location into one string.
- Hiring signals: `Upsert` writes `job_events` (deadline_extended/shortened, reposted, reappeared, reopened); `Row.Signals()` → tags, `Row.Verdict()` → `hard to fill | closed | gone` ribbon + `unfilled`/`live only` chips; email gets "STILL UNFILLED" / "CLOSED THIS WEEK" for rows with an event that week. Signals only for score ≥ 35 and kind ≠ other; LinkedIn datePosted shifts < 14 d are ignored. Add a case to `signals_test.go` when tuning thresholds.
- Every report/UI cost line must use `Cost.CostLine()`.
- Owner feedback (both radars): ★ pin (`pinned`; `ORDER BY pinned DESC` first in both `List`s, radar.js moves the card client-side), 👍/👎 (`vote`), inline autosaving note (`user_note`), trash = jobs `hidden` / funding `skip|rejected`. After the first trash of an item the UI asks once for a reason chip (`trash_reason`, `ask_reason` in JSON reply; keys in `feedback.Reasons`). Everything is appended to `feedback_log`; `RankPending` appends `feedback.PromptHints` (last 40 job verdicts) to the system prompt. "What you've taught the radar" panel above the list summarises it.
- Adding a source: append to `Sources` in sources.go; LinkedIn sleeps 6s between requests.

- Seen/unseen (both radars, no LLM): radar.js marks a card seen after ≥50 % visible for 1 s (IntersectionObserver), POSTs ids batched (sendBeacon on pagehide). Per user = `seen.User(X-ExeDev-Email)` (basic auth → `admin`). Cards carry `data-seen`, `.card.unseen::before` hairline fades via `.seen-now`; `unseen` filter chip + tappable “N unseen” header stat (`#unseen-stat`/`#unseen-n`). Cards seen in the current visit stay in the `unseen` view until the filter is re-applied.

## Funding radar conventions
- Data lives in code (`seed.go`); `reseed` replaces DB rows but keeps `status`/`user_note`/`brief`. Scores/deadlines are manually verified — record notes under `srv/funding/verification-<date>/`; bump `VerifiedDate` in verified.go (the `[verified …]` note prefix is stripped by `CleanNote`/`VerifiedDate` at render).
- LLM briefs (`brief.go`): `BriefPending` reads the official page (`jobs.FetchPageText`) and asks muse-glimmer for 5 lines What/Money/Eligible/Timeline/Next for every non-skipped entry **without** a brief (`BriefMinScore` = 0: the list is hand-curated, so each entry is briefed once) (never auto-refreshed → daily hook is free unless new seeds). Runs in the daily scheduler slot via `jobs.AfterFetch` (kind `fbrief` in job_runs → shared budget) or via the "brief" / "re-brief all" buttons (`/admin/funding/brief`, `all=1`). The card shows amount + date in the header, the brief as collapsed `<details>` (teaser = Next line), and the curated Window/Curator/Score lines beneath it.
- Trash = status `skip` via the shared `react` partial (`Hidden` = skip|rejected); skipped rows are rendered with `hidden` and shown with `?hidden=1` (trash chip in the filter bar). The status `<select>` is gone.

## Style
- Mobile-first, text-first admin UI; no frameworks, inline SVG symbols, system fonts. Descriptive commits (see `git log`).

## Access model
- **Owner** (`ADMIN_EMAIL`, or basic-auth fallback) → everything. **Viewers** (allowlist in `settings.viewer_emails`, edited at `/admin/config`) → read-only `/admin/jobs`, `/admin/funding`, `report.txt`, `status.json`; every POST and `/admin/config` stays owner-only (`requireAuth`). Handlers use `requireViewer` → `(ok, owner)`; templates hide controls with `{{if .Owner}}`.
