package gomisat

import (
	"fmt"
	"testing"
)

func TestClause01(t *testing.T) {
	c := MkClause([]Lit{MkLit(1, true), MkLit(2, false)}, true, false)
	fmt.Println(c)
}
