package jobs

import "testing"

func TestDedupe(t *testing.T) {
	s := func(v int64) *int64 { return &v }
	rows := []Row{
		{ID: 1, Source: "Noé · nous rejoindre", Title: "Directeur.trice du Parc National de Conkouati-Douli (PNCD)", URL: "https://noe.org/a", Score: s(95), FirstSeen: "2026-09-01 00:00:00"},
		{ID: 2, Source: "Noé · nous rejoindre", Title: "Conkouati-Douli Park Manager", URL: "https://noe.org/b", Score: s(95), FirstSeen: "2026-08-20 00:00:00", Deadline: "2026-10-01"},
		{ID: 3, Source: "LinkedIn · CI", Org: "Conservation International", Title: "Senior Director, KPA Resource Mobilization", Location: "Kenya", URL: "https://ke.linkedin.com/jobs/view/x-4454403176", Score: s(15)},
		{ID: 4, Source: "LinkedIn · CI", Org: "Conservation International", Title: "Senior Director, KPA Resource Mobilization", Location: "Nairobi", URL: "https://ke.linkedin.com/jobs/view/x-4455895309", Score: s(15)},
		{ID: 5, Source: "LinkedIn · CI", Org: "Conservation International", Title: "Senior Manager, Ocean Policy", URL: "https://cr.linkedin.com/jobs/view/y-4439153466", Score: s(15)},
		{ID: 6, Source: "LinkedIn · X", Org: "Conservation International", Title: "Senior Manager, Ocean Policy", URL: "https://gy.linkedin.com/jobs/view/y-4439153466?trk=1", Score: s(10)},
		{ID: 7, Source: "LinkedIn · Ramboll", Org: "Ramboll", Title: "Senior Consultant (m/w/d) Corporate Biodiversity Strategy", URL: "https://de.linkedin.com/jobs/view/1", Score: s(10)},
		{ID: 8, Source: "LinkedIn · Ramboll", Org: "Ramboll", Title: "Senior Consultant Corporate Biodiversity Strategy", URL: "https://de.linkedin.com/jobs/view/2", Score: s(10)},
		{ID: 9, Source: "Noé · nous rejoindre", Title: "Expertise scientifique sur la conservation des espèces", URL: "https://noe.org/c", Score: s(10)},
		{ID: 10, Source: "TNC", Org: "The Nature Conservancy", Title: "Utah State Director", URL: "https://l/1", Score: s(40)},
		{ID: 11, Source: "TNC", Org: "The Nature Conservancy", Title: "New Mexico State Director", URL: "https://l/2", Score: s(40)},
		// same job posted from a sub-brand account: org "WWF Cities" vs "WWF", same location
		{ID: 12, Source: "LinkedIn · npd", Org: "WWF Cities", Title: "Director WWF-North Africa", Location: "Tunis, Tunis, Tunisia", URL: "https://tn.linkedin.com/jobs/view/director-wwf-north-africa-at-wwf-cities-4473741490", Score: s(45)},
		{ID: 13, Source: "LinkedIn · npd", Org: "WWF", Title: "Director WWF-North Africa", Location: "Tunis, Tunis, Tunisia", URL: "https://tn.linkedin.com/jobs/view/director-wwf-north-africa-at-wwf-4473730526", Score: s(45)},
		// related org but different location → stays separate
		{ID: 14, Source: "LinkedIn · npd", Org: "WWF Germany", Title: "Director WWF-North Africa", Location: "Berlin", URL: "https://de.linkedin.com/jobs/view/z-1", Score: s(45)},
	}
	out := Dedupe(rows)
	ids := map[int64]Row{}
	for _, r := range out {
		ids[r.ID] = r
	}
	if len(out) != 9 {
		for _, r := range out {
			t.Logf("%d %q dupes=%d", r.ID, r.Title, r.Dupes)
		}
		t.Fatalf("want 9 groups, got %d", len(out))
	}
	if ids[12].Dupes != 1 {
		t.Errorf("wwf sub-brand merge wrong: %+v", ids[12])
	}
	if _, ok := ids[14]; !ok {
		t.Error("wwf germany wrongly merged")
	}
	if r := ids[1]; r.Dupes != 1 || r.FirstSeen != "2026-08-20 00:00:00" || r.Deadline != "2026-10-01" {
		t.Errorf("noe merge wrong: %+v", r)
	}
	if ids[3].Dupes != 1 || ids[5].Dupes != 1 || ids[7].Dupes != 1 {
		t.Errorf("linkedin merges wrong: %+v %+v %+v", ids[3], ids[5], ids[7])
	}
	if _, ok := ids[10]; !ok {
		t.Error("utah lost")
	}
	if _, ok := ids[11]; !ok {
		t.Error("new mexico wrongly merged")
	}
}

func TestDedupeWithPairs(t *testing.T) {
	rows := []Row{
		{ID: 1, Title: "Marine Protected Area Technical Consultant", Source: "UNjobnet · protected area", URL: "https://www.unjobnet.org/jobs/detail/x-1"},
		{ID: 2, Title: "Marine Protected Area Technical Consultant", Org: "UNEP - United Nations Environment Programme", Source: "Impactpool · protected area", URL: "https://www.impactpool.org/jobs/1239251"},
		{ID: 3, Title: "Marine Protected Area Technical Consultant", Org: "WWF", Source: "x", URL: "https://x/3"},
	}
	if got := len(Dedupe(rows)); got != 3 {
		t.Fatalf("deterministic rules should keep 3 (different org buckets), got %d", got)
	}
	out := DedupeWith(rows, [][2]int64{{1, 2}})
	if len(out) != 2 || out[0].ID != 1 || out[0].Dupes != 1 {
		t.Fatalf("pair should merge 1+2: %+v", out)
	}
}

func TestParseDuplicate(t *testing.T) {
	for in, want := range map[string]int64{
		"What: x\nDuplicate: 1003": 1003,
		"Fit: y\nDuplicate: none":  0,
		"Duplicate: [#42] (same)":  42,
		"no line":                  0,
		"- **Duplicate**: 7":       0, // bold label not matched → treated as none (safe)
	} {
		if got := parseDuplicate(in); got != want {
			t.Errorf("%q → %d, want %d", in, got, want)
		}
	}
}

func TestDedupeSameTitleRelatedOrg(t *testing.T) {
	rows := []Row{
		{ID: 1, Title: "Marine Protected Area Technical Consultant", Org: "UNEP", Source: "UNjobnet · protected area", URL: "https://www.unjobnet.org/jobs/detail/x-1"},
		{ID: 2, Title: "Marine Protected Area Technical Consultant", Org: "UNEP - United Nations Environment Programme", Location: "Remote | Bangkok", Source: "Impactpool · protected area", URL: "https://www.impactpool.org/jobs/1239251"},
		{ID: 3, Title: "Marine Protected Area Technical Consultant", Org: "WWF", Source: "x", URL: "https://x/3"},
	}
	out := Dedupe(rows)
	if len(out) != 2 || out[0].Dupes != 1 || out[0].Location != "Remote | Bangkok" {
		t.Fatalf("UNEP copies should merge, WWF stay: %+v", out)
	}
}

func TestParseMeta(t *testing.T) {
	org, loc := parseMeta("Org: UNEP - United Nations Environment Programme\nLocation: unknown\nWhat: x")
	if org != "UNEP - United Nations Environment Programme" || loc != "" {
		t.Fatalf("got %q %q", org, loc)
	}
}
