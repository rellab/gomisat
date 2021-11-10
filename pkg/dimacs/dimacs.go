package dimacs

import (
	"bufio"
	"bytes"
	"log"
	"strconv"
	"strings"
)

func ParseDimacs(b []byte) ([][]int, error) {
	in := bytes.NewBuffer(b)
	s := bufio.NewScanner(in)
	var nlits, nclauses int
	for s.Scan() {
		a := strings.Fields(s.Text())
		if len(a) >= 1 && a[0] == "c" { // comment line
			continue
		}
		if len(a) >= 2 && a[0] == "p" && a[1] == "cnf" {
			nlits, _ = strconv.Atoi(a[2])
			nclauses, _ = strconv.Atoi(a[3])
			log.Println("Number of literals:", nlits)
			log.Println("Number of clauses:", nclauses)
			break
		}
	}
	clauses := make([][]int, 0)
	for s.Scan() {
		a := strings.Fields(s.Text())
		if len(a) >= 2 {
			lits := make([]int, 0, len(a)-1)
			if len(a) >= 1 && a[0] == "c" { // comment line
				continue
			}
			for _, x := range a {
				if v, err := strconv.Atoi(x); err == nil {
					if v != 0 {
						lits = append(lits, int(v))
					}
				} else {
					log.Println("Atoi fails", err)
				}
			}
			clauses = append(clauses, lits)
		}
	}
	if len(clauses) != nclauses {
		log.Println("Did not match the number of clauses generated to the number of clauses defined on the header of DIMCS:", len(clauses))
	}
	return clauses, nil
}

