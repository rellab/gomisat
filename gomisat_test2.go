package test
/*
import (
	"bufio"
	"bytes"
	"fmt"
	//"github.com/mitchellh/go-sat"
	//"github.com/mitchellh/go-sat/cnf"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
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




func BenchmarkDimacs01(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/unsat-dimacs-dubois/dubois100.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs02(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/unsat-dimacs-dubois/dubois20.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs03(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/unsat-dimacs-dubois/dubois21.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs04(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/unsat-dimacs-dubois/dubois22.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs05(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/unsat-dimacs-dubois/dubois23.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs06(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/unsat-dimacs-dubois/dubois24.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs07(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/unsat-dimacs-dubois/dubois25.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs08(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/unsat-dimacs-dubois/dubois26.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs09(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/unsat-dimacs-dubois/dubois27.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs10(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/unsat-dimacs-dubois/dubois28.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs11(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/unsat-dimacs-dubois/dubois29.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs12(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/unsat-dimacs-dubois/dubois30.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs13(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/unsat-dimacs-dubois/dubois50.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs14(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-100-1_6-no-1.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs15(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-100-1_6-no-2.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs16(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-100-1_6-no-3.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs17(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-100-1_6-no-4.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs18(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-100-1_6-yes1-1.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs19(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-100-1_6-yes1-2.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs20(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-100-1_6-yes1-3.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs21(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-100-1_6-yes1-4.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs22(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-100-2_0-no-1.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs23(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-100-2_0-no-2.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs24(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-100-2_0-no-3.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs25(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-100-2_0-no-4.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs26(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-100-2_0-yes1-1.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs27(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-100-2_0-yes1-2.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs28(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-100-2_0-yes1-3.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs29(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-100-2_0-yes1-4.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs30(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-100-3_4-yes1-1.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs31(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-100-3_4-yes1-2.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs32(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-100-3_4-yes1-3.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs33(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-100-3_4-yes1-4.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs34(b *testing.B) {satlib/u
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-100-6_0-yes1-1.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs35(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-100-6_0-yes1-2.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs36(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-100-6_0-yes1-3.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs37(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-100-6_0-yes1-4.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs38(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-200-1_6-no-1.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs39(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-200-1_6-no-2.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs40(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-200-1_6-no-3.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs41(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-200-1_6-no-4.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs42(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-200-1_6-yes1-1.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs43(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-200-1_6-yes1-2.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs44(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-200-1_6-yes1-3.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs45(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-200-1_6-yes1-4.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs46(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-200-2_0-no-1.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs47(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-200-2_0-no-2.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs48(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-200-2_0-no-3.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs49(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-200-2_0-no-4.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs50(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-200-2_0-yes1-1.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs51(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-200-2_0-yes1-2.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs52(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-200-2_0-yes1-3.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs53(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-200-2_0-yes1-4.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs54(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-200-3_4-yes1-1.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs55(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-200-3_4-yes1-2.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs56(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-200-3_4-yes1-3.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs57(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-200-3_4-yes1-4.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs58(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-200-6_0-yes1-1.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs59(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-200-6_0-yes1-2.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs60(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-200-6_0-yes1-3.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs61(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-200-6_0-yes1-4.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs62(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-50-1_6-no-1.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs63(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-50-1_6-no-2.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs64(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-50-1_6-no-3.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs65(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-50-1_6-no-4.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs66(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-50-1_6-yes1-1.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs67(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-50-1_6-yes1-2.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs68(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-50-1_6-yes1-3.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs69(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-50-1_6-yes1-4.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs70(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-50-2_0-no-1.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs71(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-50-2_0-no-2.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs72(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-50-2_0-no-3.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs73(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-50-2_0-no-4.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs74(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-50-2_0-yes1-1.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs75(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-50-2_0-yes1-2.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs76(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-50-2_0-yes1-3.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs77(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-50-2_0-yes1-4.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs78(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-50-3_4-yes1-1.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs79(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-50-3_4-yes1-2.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs80(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-50-3_4-yes1-3.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs81(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-50-3_4-yes1-4.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs82(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-50-6_0-yes1-1.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs83(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-50-6_0-yes1-2.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs84(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-50-6_0-yes1-3.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

func BenchmarkDimacs85(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-50-6_0-yes1-4.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)

	
		formula := cnf.NewFormulaFromInts(cs)
		s := sat.New()
		s.AddFormula(formula)
		sat := s.Solve()
		end := time.Now()
	}
}

*/