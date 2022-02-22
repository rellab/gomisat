package test

import (
	"com.github/rellab/gomisat/pkg/gomisat"
	"fmt"
	"io"
	"log"
	"os"
	"testing"
)

func memReadStat() int {
	pid := os.Getpid()
	name := fmt.Sprintf("/proc/%d/statm", pid)
	var value int
	file, err := os.Open(name)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	n, err := fmt.Fscanf(file, "%d", &value)
	n = n +1
	return value
}

func BenchmarkDimacs01(b *testing.B) {
	log.SetOutput(io.Discard)

	file, _ := os.Open("testdata/satlib/unsat-dimacs-dubois/dubois100.cnf")
	defer file.Close()
	buf, _ := io.ReadAll(file)
	cs, _ := gomisat.ParseDimacs(buf)
	s := gomisat.NewSolver()
	options := gomisat.DefaultSolverOptions()
	for _, x := range cs {
		s.AddClauseFromCode(x, options)
	}
	s.Simplify()
	s.Solve(options)
	
	mem_used := float64(memReadStat()) * float64(os.Getpagesize()) / 1024.0*1024.0
	fmt.Printf("Memory used           : %.2f MB\n", mem_used)
}

func BenchmarkDimacs02(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/unsat-dimacs-dubois/dubois20.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}


func BenchmarkDimacs03(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/unsat-dimacs-dubois/dubois21.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs04(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/unsat-dimacs-dubois/dubois22.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs05(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/unsat-dimacs-dubois/dubois23.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs06(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/unsat-dimacs-dubois/dubois24.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs07(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/unsat-dimacs-dubois/dubois25.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs08(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/unsat-dimacs-dubois/dubois26.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs09(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/unsat-dimacs-dubois/dubois27.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs10(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/unsat-dimacs-dubois/dubois28.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs11(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/unsat-dimacs-dubois/dubois29.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs12(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/unsat-dimacs-dubois/dubois30.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs13(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/unsat-dimacs-dubois/dubois50.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs14(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-100-1_6-no-1.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs15(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-100-1_6-no-2.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs16(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-100-1_6-no-3.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs17(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-100-1_6-no-4.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs18(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-100-1_6-yes1-1.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs19(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-100-1_6-yes1-2.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs20(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-100-1_6-yes1-3.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs21(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-100-1_6-yes1-4.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs22(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-100-2_0-no-1.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs23(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-100-2_0-no-2.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs24(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-100-2_0-no-3.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs25(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-100-2_0-no-4.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs26(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-100-2_0-yes1-1.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs27(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-100-2_0-yes1-2.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs28(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-100-2_0-yes1-3.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs29(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-100-2_0-yes1-4.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs30(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-100-3_4-yes1-1.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs31(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-100-3_4-yes1-2.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs32(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-100-3_4-yes1-3.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs33(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-100-3_4-yes1-4.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs34(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-100-6_0-yes1-1.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs35(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-100-6_0-yes1-2.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs36(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-100-6_0-yes1-3.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs37(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-100-6_0-yes1-4.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs38(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-200-1_6-no-1.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs39(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-200-1_6-no-2.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs40(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-200-1_6-no-3.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs41(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-200-1_6-no-4.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs42(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-200-1_6-yes1-1.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs43(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-200-1_6-yes1-2.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs44(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-200-1_6-yes1-3.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs45(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-200-1_6-yes1-4.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs46(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-200-2_0-no-1.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs47(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-200-2_0-no-2.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs48(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-200-2_0-no-3.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs49(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-200-2_0-no-4.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs50(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-200-2_0-yes1-1.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs51(b *testing.B) {
	log.SetOutput(io.Discard)
	file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-200-2_0-yes1-2.cnf")
	defer file.Close()
	buf, _ := io.ReadAll(file)
	cs, _ := gomisat.ParseDimacs(buf)
	for i := 0; i < b.N; i++ {
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs52(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-200-2_0-yes1-3.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs53(b *testing.B) {
	log.SetOutput(io.Discard)
	file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-200-2_0-yes1-4.cnf")
	defer file.Close()
	buf, _ := io.ReadAll(file)
	cs, _ := gomisat.ParseDimacs(buf)
	for i := 0; i < b.N; i++ {
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs54(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-200-3_4-yes1-1.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs55(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-200-3_4-yes1-2.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs56(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-200-3_4-yes1-3.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs57(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-200-3_4-yes1-4.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs58(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-200-6_0-yes1-1.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs59(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-200-6_0-yes1-2.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs60(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-200-6_0-yes1-3.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs61(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-200-6_0-yes1-4.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs62(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-50-1_6-no-1.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs63(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-50-1_6-no-2.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs64(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-50-1_6-no-3.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs65(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-50-1_6-no-4.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs66(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-50-1_6-yes1-1.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs67(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-50-1_6-yes1-2.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs68(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-50-1_6-yes1-3.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs69(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-50-1_6-yes1-4.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs70(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-50-2_0-no-1.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs71(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-50-2_0-no-2.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs72(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-50-2_0-no-3.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs73(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-50-2_0-no-4.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs74(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-50-2_0-yes1-1.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs75(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-50-2_0-yes1-2.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs76(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-50-2_0-yes1-3.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs77(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-50-2_0-yes1-4.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs78(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-50-3_4-yes1-1.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs79(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-50-3_4-yes1-2.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs80(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-50-3_4-yes1-3.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs81(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-50-3_4-yes1-4.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs82(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-50-6_0-yes1-1.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs83(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-50-6_0-yes1-2.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs84(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-50-6_0-yes1-3.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}

func BenchmarkDimacs85(b *testing.B) {
	log.SetOutput(io.Discard)
	for i := 0; i < b.N; i++ {
		file, _ := os.Open("testdata/satlib/file-dimacs-aim/aim-50-6_0-yes1-4.cnf")
		defer file.Close()
		buf, _ := io.ReadAll(file)
		cs, _ := gomisat.ParseDimacs(buf)
	
		s := gomisat.NewSolver()
		options := gomisat.DefaultSolverOptions()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}
		s.Simplify()
		s.Solve(options)
	}
}