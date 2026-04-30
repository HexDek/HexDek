package versioning

import (
	"testing"

	"github.com/hexdek/hexdek/internal/trueskill"
)

func TestHashCardList_DeterministicAndOrderIndependent(t *testing.T) {
	a := HashCardList([]string{"Sol Ring", "Mountain", "Lightning Bolt"})
	b := HashCardList([]string{"Lightning Bolt", "Mountain", "Sol Ring"})
	if a != b {
		t.Errorf("hash should be order-independent: %s vs %s", a, b)
	}

	c := HashCardList([]string{"Sol Ring", "Mountain", "Counterspell"})
	if a == c {
		t.Errorf("different cards should produce different hash")
	}
}

func TestCardDelta(t *testing.T) {
	old := []string{"A", "B", "C"}
	new1 := []string{"A", "B", "C"}
	if d := CardDelta(old, new1); d != 0 {
		t.Errorf("identical lists: got delta %d, want 0", d)
	}

	new2 := []string{"A", "B", "D"} // swapped C for D
	if d := CardDelta(old, new2); d != 2 {
		t.Errorf("one swap: got delta %d, want 2 (one removed, one added)", d)
	}

	new3 := []string{"A", "B", "C", "D"} // added one
	if d := CardDelta(old, new3); d != 1 {
		t.Errorf("one addition: got delta %d, want 1", d)
	}

	// Multiple of same card.
	if d := CardDelta([]string{"Forest", "Forest"}, []string{"Forest", "Forest", "Forest"}); d != 1 {
		t.Errorf("forest count change: got %d, want 1", d)
	}
}

func TestRegisterVersion_RootNode(t *testing.T) {
	dag := NewDeckDAG()
	cards := []string{"Sol Ring", "Mountain"}
	node := dag.RegisterVersion("alice", "burn", "Krenko", cards)

	if node == nil {
		t.Fatal("nil node")
	}
	if node.Version != 1 {
		t.Errorf("root version: got %d, want 1", node.Version)
	}
	if node.ParentHash != "" {
		t.Errorf("root should have empty ParentHash, got %q", node.ParentHash)
	}
	if !node.IsHead {
		t.Errorf("root should be HEAD")
	}
	if node.CardCount != 2 {
		t.Errorf("CardCount: got %d", node.CardCount)
	}
	defaultR := trueskill.DefaultRating()
	if node.Rating.Mu != defaultR.Mu {
		t.Errorf("root rating should be default mu, got %v", node.Rating.Mu)
	}
}

func TestRegisterVersion_Idempotent(t *testing.T) {
	dag := NewDeckDAG()
	cards := []string{"A", "B"}
	n1 := dag.RegisterVersion("alice", "deck", "Cmdr", cards)
	n2 := dag.RegisterVersion("alice", "deck", "Cmdr", cards)
	if n1 != n2 {
		t.Errorf("re-registering same cards should return same node")
	}
	if len(dag.Nodes) != 1 {
		t.Errorf("expected 1 node, got %d", len(dag.Nodes))
	}
}

func TestRegisterVersion_ChildInheritsAndDethronesParent(t *testing.T) {
	dag := NewDeckDAG()
	// Use 10+ cards so the fallback CardDelta (len/5) is non-zero and
	// sigma actually inflates from inheritance.
	parentCards := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J"}
	parent := dag.RegisterVersion("alice", "deck", "Cmdr", parentCards)

	// Mutate parent rating so we can verify inheritance carries μ.
	parent.Rating = trueskill.Rating{Mu: 30.0, Sigma: 5.0}

	childCards := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "K"}
	child := dag.RegisterVersion("alice", "deck", "Cmdr", childCards)
	if child == parent {
		t.Fatal("child should be a new node")
	}
	if child.ParentHash != parent.Hash {
		t.Errorf("child parent: got %q, want %q", child.ParentHash, parent.Hash)
	}
	if child.Version != 2 {
		t.Errorf("child version: got %d, want 2", child.Version)
	}
	if !child.IsHead {
		t.Errorf("child should be HEAD")
	}
	if parent.IsHead {
		t.Errorf("parent should no longer be HEAD")
	}
	if child.Rating.Mu != 30.0 {
		t.Errorf("child should inherit μ=30.0, got %v", child.Rating.Mu)
	}
	if child.Rating.Sigma <= 5.0 {
		t.Errorf("child σ should inflate above parent's 5.0, got %v", child.Rating.Sigma)
	}
	if child.CardDelta == 0 {
		t.Errorf("child CardDelta should be non-zero")
	}
}

func TestGetHeadAndLineage(t *testing.T) {
	dag := NewDeckDAG()
	v1 := dag.RegisterVersion("alice", "deck", "Cmdr", []string{"A", "B"})
	v2 := dag.RegisterVersion("alice", "deck", "Cmdr", []string{"A", "C"})
	v3 := dag.RegisterVersion("alice", "deck", "Cmdr", []string{"A", "D"})

	head := dag.GetHead("alice", "deck")
	if head == nil || head.Hash != v3.Hash {
		t.Errorf("HEAD should be v3")
	}

	lineage := dag.GetLineage("alice", "deck")
	if len(lineage) != 3 {
		t.Fatalf("lineage length: got %d, want 3", len(lineage))
	}
	if lineage[0].Hash != v3.Hash || lineage[1].Hash != v2.Hash || lineage[2].Hash != v1.Hash {
		t.Errorf("lineage order should be newest-first")
	}

	if dag.GetHead("alice", "missing") != nil {
		t.Errorf("missing deck should return nil head")
	}
	if dag.GetLineage("alice", "missing") != nil {
		t.Errorf("missing deck should return nil lineage")
	}
}

func TestLeaderboard(t *testing.T) {
	dag := NewDeckDAG()
	a := dag.RegisterVersion("alice", "deckA", "CA", []string{"X1"})
	b := dag.RegisterVersion("bob", "deckB", "CB", []string{"X2"})
	c := dag.RegisterVersion("cora", "deckC", "CC", []string{"X3"})

	a.Rating = trueskill.Rating{Mu: 30, Sigma: 2}
	b.Rating = trueskill.Rating{Mu: 25, Sigma: 2}
	c.Rating = trueskill.Rating{Mu: 35, Sigma: 2}

	board := dag.Leaderboard()
	if len(board) != 3 {
		t.Fatalf("got %d entries, want 3", len(board))
	}
	if board[0].Owner != "cora" || board[1].Owner != "alice" || board[2].Owner != "bob" {
		t.Errorf("leaderboard order wrong: %v / %v / %v", board[0].Owner, board[1].Owner, board[2].Owner)
	}
}

func TestLeaderboard_OnlyHeadsIncluded(t *testing.T) {
	dag := NewDeckDAG()
	dag.RegisterVersion("alice", "deck", "Cmdr", []string{"A"})
	dag.RegisterVersion("alice", "deck", "Cmdr", []string{"B"}) // bumps HEAD

	board := dag.Leaderboard()
	if len(board) != 1 {
		t.Errorf("only HEAD should be on leaderboard, got %d", len(board))
	}
}

func TestLeaderboard_Empty(t *testing.T) {
	dag := NewDeckDAG()
	if got := dag.Leaderboard(); len(got) != 0 {
		t.Errorf("empty DAG leaderboard: got %d", len(got))
	}
}

func TestUpdateRating(t *testing.T) {
	dag := NewDeckDAG()
	n := dag.RegisterVersion("a", "d", "C", []string{"X"})
	dag.UpdateRating(n.Hash, trueskill.Rating{Mu: 28, Sigma: 4}, 7)
	if n.Rating.Mu != 28 || n.GamesPlayed != 7 {
		t.Errorf("update did not apply: %+v games=%d", n.Rating, n.GamesPlayed)
	}

	// No-op for unknown hash.
	dag.UpdateRating("nonexistent", trueskill.Rating{Mu: 99}, 99)
}

func TestLookupByCommander(t *testing.T) {
	dag := NewDeckDAG()
	dag.RegisterVersion("a", "d1", "Krenko, Mob Boss", []string{"X"})
	dag.RegisterVersion("b", "d2", "Edgar Markov", []string{"Y"})

	got := dag.LookupByCommander("krenko, mob boss")
	if got == nil || got.DeckID != "d1" {
		t.Errorf("case-insensitive lookup failed: %+v", got)
	}
	if got := dag.LookupByCommander("Unknown"); got != nil {
		t.Errorf("unknown commander should return nil")
	}
}

func TestSaveAndLoadDAG(t *testing.T) {
	dir := t.TempDir()
	dag := NewDeckDAG()
	dag.RegisterVersion("a", "d", "Cmdr", []string{"X", "Y"})
	dag.RegisterVersion("a", "d", "Cmdr", []string{"X", "Z"})

	if err := SaveDAG(dir, dag); err != nil {
		t.Fatalf("SaveDAG: %v", err)
	}

	loaded, err := LoadDAG(dir)
	if err != nil {
		t.Fatalf("LoadDAG: %v", err)
	}
	if len(loaded.Nodes) != 2 {
		t.Errorf("loaded nodes: got %d, want 2", len(loaded.Nodes))
	}
	if len(loaded.Heads) != 1 {
		t.Errorf("loaded heads: got %d, want 1", len(loaded.Heads))
	}
}

func TestLoadDAG_MissingFile(t *testing.T) {
	dir := t.TempDir()
	dag, err := LoadDAG(dir)
	if err != nil {
		t.Errorf("missing file should return empty DAG, got: %v", err)
	}
	if dag == nil || len(dag.Nodes) != 0 {
		t.Errorf("expected empty DAG, got %+v", dag)
	}
}

func TestSigmaInflation(t *testing.T) {
	if v := SigmaInflation(0); v != 0 {
		t.Errorf("delta 0: got %v, want 0", v)
	}
	// Capped at 25/6 ≈ 4.166...
	if v := SigmaInflation(100); v < 4.16 || v > 4.17 {
		t.Errorf("large delta should cap at 25/6, got %v", v)
	}
	if v := SigmaInflation(4); v != 2.0 {
		t.Errorf("delta 4: got %v, want 2.0", v)
	}
}
