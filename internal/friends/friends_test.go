package friends

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hexdek/hexdek/internal/db"
)

func newTestTracker(t *testing.T) *Tracker {
	t.Helper()
	dbPath := t.TempDir() + "/friends_test.db"
	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	tr, err := New(database)
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	return tr
}

func TestAddListAreFriends(t *testing.T) {
	tr := newTestTracker(t)
	ctx := context.Background()

	if err := tr.AddFriend(ctx, "josh", "meglin"); err != nil {
		t.Fatalf("add: %v", err)
	}
	// Idempotent.
	if err := tr.AddFriend(ctx, "josh", "meglin"); err != nil {
		t.Fatalf("add idempotent: %v", err)
	}

	for _, pair := range [][2]string{{"josh", "meglin"}, {"meglin", "josh"}} {
		ok, err := tr.AreFriends(ctx, pair[0], pair[1])
		if err != nil {
			t.Fatalf("areFriends: %v", err)
		}
		if !ok {
			t.Errorf("expected %s ↔ %s to be friends", pair[0], pair[1])
		}
	}

	js, err := tr.ListFriends(ctx, "josh")
	if err != nil {
		t.Fatalf("list josh: %v", err)
	}
	if len(js) != 1 || js[0] != "meglin" {
		t.Errorf("josh's friends: got %v, want [meglin]", js)
	}
	mg, _ := tr.ListFriends(ctx, "meglin")
	if len(mg) != 1 || mg[0] != "josh" {
		t.Errorf("meglin's friends: got %v, want [josh]", mg)
	}
}

func TestRemoveFriend(t *testing.T) {
	tr := newTestTracker(t)
	ctx := context.Background()

	_ = tr.AddFriend(ctx, "a", "b")
	if err := tr.RemoveFriend(ctx, "a", "b"); err != nil {
		t.Fatalf("remove: %v", err)
	}
	// Idempotent.
	if err := tr.RemoveFriend(ctx, "a", "b"); err != nil {
		t.Fatalf("remove idempotent: %v", err)
	}
	ok, _ := tr.AreFriends(ctx, "a", "b")
	if ok {
		t.Errorf("expected a/b to no longer be friends after remove")
	}
}

func TestRejectSelfAndEmpty(t *testing.T) {
	tr := newTestTracker(t)
	ctx := context.Background()
	if err := tr.AddFriend(ctx, "josh", "josh"); err == nil {
		t.Errorf("expected error on self-add")
	}
	if err := tr.AddFriend(ctx, "", "meglin"); err == nil {
		t.Errorf("expected error on empty owner")
	}
}

func TestNormalizesCaseAndWhitespace(t *testing.T) {
	tr := newTestTracker(t)
	ctx := context.Background()
	if err := tr.AddFriend(ctx, "  Josh ", "MEGLIN"); err != nil {
		t.Fatalf("add: %v", err)
	}
	ok, _ := tr.AreFriends(ctx, "josh", "meglin")
	if !ok {
		t.Errorf("normalization didn't lowercase/trim")
	}
}

func TestEndpoints(t *testing.T) {
	tr := newTestTracker(t)
	mux := http.NewServeMux()
	tr.Register(mux)

	// Add: POST /api/friends/meglin?as=josh
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest("POST", "/api/friends/meglin?as=josh", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("add status: got %d, body %s", rr.Code, rr.Body.String())
	}

	// List: GET /api/friends?as=josh
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest("GET", "/api/friends?as=josh", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("list status: got %d", rr.Code)
	}
	var resp struct {
		Friends []string `json:"friends"`
	}
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	if len(resp.Friends) != 1 || resp.Friends[0] != "meglin" {
		t.Errorf("list: got %v, want [meglin]", resp.Friends)
	}

	// Missing ?as= returns 400.
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest("GET", "/api/friends", nil))
	if rr.Code != http.StatusBadRequest {
		t.Errorf("missing as: got %d, want 400", rr.Code)
	}

	// Remove: DELETE /api/friends/meglin?as=josh
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest("DELETE", "/api/friends/meglin?as=josh", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("remove status: got %d", rr.Code)
	}
	ok, _ := tr.AreFriends(context.Background(), "josh", "meglin")
	if ok {
		t.Errorf("still friends after DELETE")
	}
}
