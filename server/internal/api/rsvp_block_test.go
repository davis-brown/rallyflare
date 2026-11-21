package api_test

import (
	"net/http"
	"testing"
	"time"
)

func TestRSVPGoingRejectedAcrossBlock(t *testing.T) {
	ts, _ := newTestServer(t)
	organizer := registerUser(t, ts, "rsvp-organizer@test.local", "RSVP Organizer")
	blockedByOrganizer := registerUser(t, ts, "rsvp-blocked@test.local", "RSVP Blocked")
	blocker := registerUser(t, ts, "rsvp-blocker@test.local", "RSVP Blocker")
	friend := registerUser(t, ts, "rsvp-friend@test.local", "RSVP Friend")
	court := createTestCourt(t, ts, organizer.AccessToken, "RSVP Block Court", ruckerLat, ruckerLng)

	resp := doJSON(t, ts, http.MethodPost, "/courts/"+court.ID+"/sessions", organizer.AccessToken, map[string]any{
		"starts_at": time.Now().Add(time.Hour).UTC().Format(time.RFC3339),
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create session: status %d: %s", resp.StatusCode, readBody(t, resp))
	}
	session := decodeJSON[struct {
		ID string `json:"id"`
	}](t, resp)

	rsvp := func(token, status string) int {
		resp := doJSON(t, ts, http.MethodPut, "/sessions/"+session.ID+"/rsvp", token, map[string]any{"status": status})
		resp.Body.Close()
		return resp.StatusCode
	}
	goingCount := func() int {
		resp := doJSON(t, ts, http.MethodGet, "/sessions/"+session.ID+"/attendees", "", nil)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("list attendees: status %d", resp.StatusCode)
		}
		return decodeJSON[struct {
			GoingCount int `json:"going_count"`
		}](t, resp).GoingCount
	}

	// blockedByOrganizer and blocker RSVP before any block, so "out" can be
	// checked as a withdrawal across it.
	for _, u := range []testUser{blockedByOrganizer, blocker} {
		if code := rsvp(u.AccessToken, "going"); code != http.StatusOK {
			t.Fatalf("pre-block going: status %d", code)
		}
	}
	if got := goingCount(); got != 3 {
		t.Fatalf("going_count before blocks = %d, want 3", got)
	}

	doJSON(t, ts, http.MethodPut, "/users/"+blockedByOrganizer.User.ID+"/block", organizer.AccessToken, nil).Body.Close()
	doJSON(t, ts, http.MethodPut, "/users/"+organizer.User.ID+"/block", blocker.AccessToken, nil).Body.Close()

	for name, u := range map[string]testUser{"blocked by organizer": blockedByOrganizer, "blocker of organizer": blocker} {
		if code := rsvp(u.AccessToken, "out"); code != http.StatusOK {
			t.Fatalf("%s: out across block: status %d", name, code)
		}
		if code := rsvp(u.AccessToken, "going"); code != http.StatusNotFound {
			t.Fatalf("%s: going across block: status %d, want 404", name, code)
		}
	}
	if got := goingCount(); got != 1 {
		t.Fatalf("going_count after blocked RSVPs = %d, want 1", got)
	}

	if code := rsvp(friend.AccessToken, "going"); code != http.StatusOK {
		t.Fatalf("unblocked going: status %d", code)
	}
	if got := goingCount(); got != 2 {
		t.Fatalf("going_count after unblocked RSVP = %d, want 2", got)
	}
}
