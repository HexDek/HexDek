package matchmaking

import (
	"math/rand"
	"testing"
)

func newPool(n int) []DeckEntry {
	pool := make([]DeckEntry, n)
	for i := 0; i < n; i++ {
		pool[i] = DeckEntry{
			Index:     i,
			Commander: "C" + string(rune('A'+i)),
			Mu:        25.0,
			Sigma:     8.33,
			Games:     0,
		}
	}
	return pool
}

func TestAssemblePod_PoolEqualToSeats(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	pool := newPool(4)
	got := AssemblePod(rng, pool, 4)
	if len(got) != 4 {
		t.Fatalf("expected 4 indices, got %d", len(got))
	}
	seen := map[int]bool{}
	for _, idx := range got {
		if seen[idx] {
			t.Errorf("duplicate index %d", idx)
		}
		seen[idx] = true
	}
}

func TestAssemblePod_PoolSmallerThanSeats(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	pool := newPool(2)
	got := AssemblePod(rng, pool, 4)
	if len(got) != 2 {
		t.Errorf("undersized pool should return all decks, got %d", len(got))
	}
}

func TestAssemblePod_CorrectPodSize(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	pool := newPool(20)
	for podSize := 2; podSize <= 6; podSize++ {
		got := AssemblePod(rng, pool, podSize)
		if len(got) != podSize {
			t.Errorf("pod size %d: got %d", podSize, len(got))
		}
		seen := map[int]bool{}
		for _, idx := range got {
			if seen[idx] {
				t.Errorf("duplicate index %d in pod size %d", idx, podSize)
			}
			seen[idx] = true
			if idx < 0 || idx >= 20 {
				t.Errorf("index %d out of range", idx)
			}
		}
	}
}

func TestAssemblePod_RatingAware_PrefersClose(t *testing.T) {
	// One cluster of high-Mu decks and one cluster of low-Mu decks. With
	// low sigma, the algorithm should keep the seed and its pod close.
	pool := []DeckEntry{
		{Index: 0, Commander: "Hi1", Mu: 35, Sigma: 1, Games: 50},
		{Index: 1, Commander: "Hi2", Mu: 36, Sigma: 1, Games: 50},
		{Index: 2, Commander: "Hi3", Mu: 34, Sigma: 1, Games: 50},
		{Index: 3, Commander: "Hi4", Mu: 35, Sigma: 1, Games: 50},
		{Index: 4, Commander: "Lo1", Mu: 15, Sigma: 1, Games: 50},
		{Index: 5, Commander: "Lo2", Mu: 14, Sigma: 1, Games: 50},
		{Index: 6, Commander: "Lo3", Mu: 16, Sigma: 1, Games: 50},
		{Index: 7, Commander: "Lo4", Mu: 15, Sigma: 1, Games: 50},
	}

	// Run several iterations — each pod should be a single cluster.
	for trial := 0; trial < 20; trial++ {
		rng := rand.New(rand.NewSource(int64(trial)))
		got := AssemblePod(rng, pool, 4)
		if len(got) != 4 {
			t.Fatalf("trial %d: pod size %d, want 4", trial, len(got))
		}
		highCount, lowCount := 0, 0
		for _, idx := range got {
			if idx < 4 {
				highCount++
			} else {
				lowCount++
			}
		}
		// Allow occasional crossover from the random tail-fallback,
		// but most of the pod should be one cluster.
		if highCount < 3 && lowCount < 3 {
			t.Errorf("trial %d: pod is mixed, expected dominant cluster: high=%d low=%d",
				trial, highCount, lowCount)
		}
	}
}

func TestAssemblePod_SingleDeck(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	pool := newPool(1)
	got := AssemblePod(rng, pool, 4)
	if len(got) != 1 || got[0] != 0 {
		t.Errorf("single deck: got %v, want [0]", got)
	}
}

func TestAssemblePod_EmptyPool(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	got := AssemblePod(rng, []DeckEntry{}, 4)
	if len(got) != 0 {
		t.Errorf("empty pool: got %d, want 0", len(got))
	}
}

func TestAssemblePod_RespectsIndexField(t *testing.T) {
	// Verify the returned values are pool[i].Index, not raw slice positions.
	pool := []DeckEntry{
		{Index: 100, Mu: 25, Sigma: 8.33},
		{Index: 200, Mu: 25, Sigma: 8.33},
		{Index: 300, Mu: 25, Sigma: 8.33},
		{Index: 400, Mu: 25, Sigma: 8.33},
		{Index: 500, Mu: 25, Sigma: 8.33},
	}
	rng := rand.New(rand.NewSource(7))
	got := AssemblePod(rng, pool, 4)
	if len(got) != 4 {
		t.Fatalf("got %d", len(got))
	}
	for _, idx := range got {
		if idx < 100 || idx%100 != 0 {
			t.Errorf("expected pool.Index value, got %d", idx)
		}
	}
}
