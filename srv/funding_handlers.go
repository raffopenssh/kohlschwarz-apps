package srv

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"srv.exe.dev/srv/feedback"
	"srv.exe.dev/srv/funding"
	"srv.exe.dev/srv/jobs"
	"srv.exe.dev/srv/seen"
)

type fundingPage struct {
	Rows       []funding.Entry
	Upcoming   []funding.Entry
	Msg        string
	Today      string
	Statuses   []string
	Owner      bool // false for allowlisted viewers (read-only)
	Viewers    int
	Reasons    []feedback.Reason
	Feedback   feedbackPanel
	Activity   jobs.ActivityState
	ShowHidden bool   // ?hidden=1 lists skipped/rejected entries too
	Briefed    int    // entries with an LLM brief
	Unbriefed  int    // entries still waiting for one (open, not yet briefed)
	BriefDate  string // newest brief date
	MonthUSD   string
	Budget     string
	Unseen     int // non-skipped entries the current user has not had on screen yet
}

func (s *Server) HandleAdminFunding(w http.ResponseWriter, r *http.Request) {
	ok, owner := s.requireViewer(w, r)
	if !ok {
		return
	}
	ctx := r.Context()
	rows, err := funding.List(ctx, s.DB)
	if err != nil {
		slog.Warn("funding list", "error", err)
	}
	up, _ := funding.Upcoming(ctx, s.DB, 60, 40)
	data := fundingPage{Rows: rows, Upcoming: up, Msg: r.URL.Query().Get("msg"), Today: time.Now().UTC().Format("2006-01-02"), Statuses: []string{"open", "applied", "rejected", "won", "skip"}, Owner: owner, Viewers: len(s.viewers(ctx)), Reasons: feedback.Reasons, Feedback: s.feedbackPanel(ctx, "grant"), Activity: jobs.Current.State(), ShowHidden: r.URL.Query().Get("hidden") == "1"}
	sn, _ := seen.Set(ctx, s.DB, "grant", seen.User(r.Header.Get("X-ExeDev-Email")))
	for i := range rows {
		rows[i].Seen = sn[rows[i].ID]
		if !rows[i].Seen && rows[i].Status != "skip" && rows[i].Status != "rejected" {
			data.Unseen++
		}
	}
	for _, e := range rows {
		if e.Brief != "" {
			data.Briefed++
			if d := e.BriefedDate(); d > data.BriefDate {
				data.BriefDate = d
			}
		}
	}
	if ub, err := funding.Unbriefed(ctx, s.DB, funding.BriefMinScore, 1000); err == nil {
		data.Unbriefed = len(ub)
	}
	cost := jobs.GetCost(ctx, s.DB)
	data.MonthUSD, data.Budget = jobs.USD(cost.MonthUSD), jobs.USD(jobs.MaxMonthUSD())
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.renderTemplate(w, "funding.html", data); err != nil {
		slog.Warn("render funding", "error", err)
	}
}

func (s *Server) HandleAdminFundingReseed(w http.ResponseWriter, r *http.Request) {
	if !s.requireAuth(w, r) {
		return
	}
	n, err := funding.Seed(r.Context(), s.DB)
	msg := "reseeded " + strconv.Itoa(n) + " entries"
	if err != nil {
		msg = "reseed error: " + err.Error()
	}
	http.Redirect(w, r, "/admin/funding?msg="+msg, http.StatusSeeOther)
}

// HandleAdminFundingBrief generates LLM briefs for entries that lack one
// (?all=1 first clears every brief so they are regenerated).
func (s *Server) HandleAdminFundingBrief(w http.ResponseWriter, r *http.Request) {
	if !s.requireAuth(w, r) {
		return
	}
	all := r.FormValue("all") == "1"
	s.startBackground(w, r, "brief", 15*time.Minute, func(ctx context.Context) string {
		if all {
			if err := funding.ClearBriefs(ctx, s.DB); err != nil {
				return "clear briefs: " + err.Error()
			}
		}
		run := funding.BriefPending(ctx, s.DB, 200)
		return fmt.Sprintf("briefed %d funding entries, cost %s", run.Ranked, jobs.USD(run.CostUSD))
	})
}
