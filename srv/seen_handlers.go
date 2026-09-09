package srv

import (
	"encoding/json"
	"io"
	"net/http"

	"srv.exe.dev/srv/seen"
)

// HandleAdminJobsSeen / HandleAdminFundingSeen record cards the current user
// has had on screen. Body: JSON {"ids":[1,2,3]} (radar.js batches + sendBeacon
// on pagehide). Viewers are allowed — the state is theirs, not the owner's.
func (s *Server) HandleAdminJobsSeen(w http.ResponseWriter, r *http.Request) { s.markSeen(w, r, "job") }
func (s *Server) HandleAdminFundingSeen(w http.ResponseWriter, r *http.Request) {
	s.markSeen(w, r, "grant")
}

func (s *Server) markSeen(w http.ResponseWriter, r *http.Request, kind string) {
	if ok, _ := s.requireViewer(w, r); !ok {
		return
	}
	var body struct {
		IDs []int64 `json:"ids"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 64<<10)).Decode(&body); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	n, err := seen.Mark(r.Context(), s.DB, kind, seen.User(r.Header.Get("X-ExeDev-Email")), body.IDs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	json.NewEncoder(w).Encode(map[string]any{"ok": true, "marked": n})
}
