package gomisat

import "testing"

func TestClauseArenaAlloc(t *testing.T) {
	a := newClauseArena()
	c := a.alloc([]Lit{MkLit(1, true), MkLit(2, false)}, false, true)
	if got := a.Size(c); got != 2 {
		t.Errorf("Size = %d, want 2", got)
	}
	if a.Learnt(c) {
		t.Error("clause is reported as learnt")
	}
	if got := a.String(c); got == "" {
		t.Error("String is empty")
	}
}

// TestClauseArenaCompactKeepsReferences is the property the whole design rests
// on: compacting the literal store must not change any clause reference.
func TestClauseArenaCompactKeepsReferences(t *testing.T) {
	a := newClauseArena()
	refs := make([]CRef, 0, 8)
	want := make([][]Lit, 0, 8)
	for i := 0; i < 8; i++ {
		lits := []Lit{MkLit(Var(i), false), MkLit(Var(i+1), true), MkLit(Var(i+2), false)}
		refs = append(refs, a.alloc(lits, true, false))
		want = append(want, lits)
	}
	// Delete every other clause, then compact.
	for i := 0; i < len(refs); i += 2 {
		a.markDead(refs[i])
	}
	if a.wastedFraction() <= 0 {
		t.Fatal("deleting clauses did not produce any waste")
	}
	a.compact()
	if a.wastedFraction() != 0 {
		t.Errorf("wastedFraction = %v after compaction, want 0", a.wastedFraction())
	}
	for i := 1; i < len(refs); i += 2 {
		got := a.Lits(refs[i])
		if len(got) != len(want[i]) {
			t.Fatalf("clause %d has %d literals after compaction, want %d", i, len(got), len(want[i]))
		}
		for j := range got {
			if got[j] != want[i][j] {
				t.Fatalf("clause %d literal %d = %v, want %v", i, j, got[j], want[i][j])
			}
		}
	}
}
