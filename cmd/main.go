package main

import (
	"com.github/rellab/gomisat/pkg/gomisat"
	"fmt"
	"io"
	"os"
	"time"
)

/*
func main() {
	fname := os.Args[1]
	file, err := os.Open(fname)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	buf, _ := io.ReadAll(file)
	cs, _ := gomisat.ParseDimacs(buf)
	s := gomisat.NewSolver()
	options := gomisat.DefaultSolverOptions()
	for _, x := range cs {
		s.AddClauseFromCode(x, options)
	}

	var dummy []gomisat.Lit
	//dummy = append(dummy, gomisat.MkLit(30, true))
	//dummy = append(dummy, gomisat.MkLit(109, true))
	//dummy = append(dummy, gomisat.MkLit(142, true))
	//dummy = append(dummy, gomisat.MkLit(199, false))
	s.SolveLimited(dummy)

	start := time.Now()
	s.Simplify()
	fmt.Println(s.Solve(options))
	end := time.Now()
	fmt.Println(s.Conflicts)
	fmt.Println(s.Propagations)
	fmt.Printf("computation time : %.8f (sec)\n", (end.Sub(start)).Seconds())
	fmt.Printf("Memory used : %.2f MB\n", s.Mem_used())
}
*/



/*
func main() {
	fname := os.Args[1]
	file, err := os.Open(fname)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	buf, _ := io.ReadAll(file)
	cs, _ := gomisat.ParseDimacs(buf)
	options := gomisat.DefaultSolverOptions()

	var (
		sat gomisat.Solver
		slv gomisat.LBool
		t float64
	)
	sat1 := make(chan *gomisat.Solver)
	slv1 := make(chan gomisat.LBool)
	t1 := make(chan float64)
	sat2 := make(chan *gomisat.Solver)
	slv2 := make(chan gomisat.LBool)
	t2 := make(chan float64)

	go func(){
		s := gomisat.NewSolver()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}

		var dummy []gomisat.Lit
		dummy = append(dummy, gomisat.MkLit(0, false))
		s.SolveLimited(dummy)

		start := time.Now()
		s.Simplify()
		slv1 <- s.Solve(options)
		end := time.Now()
		sat1 <- s
		t1 <- (end.Sub(start)).Seconds()
	}()

	go func(){
		s := gomisat.NewSolver()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}

		var dummy []gomisat.Lit
		dummy = append(dummy, gomisat.MkLit(1, true))
		s.SolveLimited(dummy)

		start := time.Now()
		s.Simplify()
		slv2 <- s.Solve(options)
		end := time.Now()
		sat2 <- s
		t2 <- (end.Sub(start)).Seconds()
	}()

	slv = <-slv1
	sat = *<-sat1
	fmt.Println(slv)
	fmt.Println(sat.Conflicts)
	fmt.Println(sat.Propagations)
	fmt.Printf("Memory used : %.2f MB\n", sat.Mem_used())
	t = <-t1
	fmt.Printf("computation time : %.8f (sec)\n", t)

	slv = <-slv2
	sat = *<-sat2
	fmt.Println(slv)
	fmt.Println(sat.Conflicts)
	fmt.Println(sat.Propagations)
	fmt.Printf("Memory used : %.2f MB\n", sat.Mem_used())
	t = <-t2
	fmt.Printf("computation time : %.8f (sec)\n", t)

}
*/


/*
func main() {
	fname := os.Args[1]
	file, err := os.Open(fname)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	buf, _ := io.ReadAll(file)
	cs, _ := gomisat.ParseDimacs(buf)
	options := gomisat.DefaultSolverOptions()

	var (
		sat gomisat.Solver
		slv gomisat.LBool
		t float64
	)
	sat1 := make(chan *gomisat.Solver)
	slv1 := make(chan gomisat.LBool)
	t1 := make(chan float64)
	sat2 := make(chan *gomisat.Solver)
	slv2 := make(chan gomisat.LBool)
	t2 := make(chan float64)

	go func(){
		s := gomisat.NewSolver()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}

		var dummy []gomisat.Lit
		dummy = append(dummy, gomisat.MkLit(0, false))
		s.SolveLimited(dummy)

		start := time.Now()
		s.Simplify()
		slv1 <- s.Solve(options)
		end := time.Now()
		sat1 <- s
		t1 <- (end.Sub(start)).Seconds()
	}()

	go func(){
		s := gomisat.NewSolver()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}

		var dummy []gomisat.Lit
		dummy = append(dummy, gomisat.MkLit(1, true))
		s.SolveLimited(dummy)

		start := time.Now()
		s.Simplify()
		slv2 <- s.Solve(options)
		end := time.Now()
		sat2 <- s
		t2 <- (end.Sub(start)).Seconds()
	}()

	for i:=0; i<2; i++{
		select {
		case slv = <-slv1:
			sat = *<-sat1
			t = <-t1
			if slv == gomisat.LTrue {
				fmt.Println(slv)
				fmt.Println(sat.Conflicts)
				fmt.Println(sat.Propagations)
				fmt.Printf("Memory used : %.2f MB\n", sat.Mem_used())
				fmt.Printf("computation time : %.8f (sec)\n", t)
				os.Exit(0)
			}
		case slv = <-slv2:
			sat = *<-sat2
			t = <-t2
			if slv == gomisat.LTrue {
				fmt.Println(slv)
				fmt.Println(sat.Conflicts)
				fmt.Println(sat.Propagations)
				fmt.Printf("Memory used : %.2f MB\n", sat.Mem_used())
				fmt.Printf("computation time : %.8f (sec)\n", t)
				os.Exit(0)
			}
		}
	}
	
	fmt.Println(slv)
	fmt.Println(sat.Conflicts)
	fmt.Println(sat.Propagations)
	fmt.Printf("Memory used : %.2f MB\n", sat.Mem_used())
	fmt.Printf("computation time : %.8f (sec)\n", t)
	
}
*/




//2^1
func main() {
	fname := os.Args[1]
	file, err := os.Open(fname)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	buf, _ := io.ReadAll(file)
	cs, _ := gomisat.ParseDimacs(buf)
	options := gomisat.DefaultSolverOptions()

	var (
		sat gomisat.Solver
		slv gomisat.LBool
		t float64
	)
	sat1 := make(chan *gomisat.Solver)
	slv1 := make(chan gomisat.LBool)
	sat2 := make(chan *gomisat.Solver)
	slv2 := make(chan gomisat.LBool)
	t1 := make(chan float64)
	t2 := make(chan float64)

	var v1 gomisat.Var = 15


	go func(){
		s := gomisat.NewSolver()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}

		var dummy []gomisat.Lit
		dummy = append(dummy, gomisat.MkLit(v1, false))
		s.SolveLimited(dummy) //1

		start := time.Now()
		s.Simplify()
		slv1 <- s.Solve(options)
		end := time.Now()
		sat1 <- s
		t1 <- (end.Sub(start)).Seconds()
	}()

	go func(){
		s := gomisat.NewSolver()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}

		var dummy []gomisat.Lit
		dummy = append(dummy, gomisat.MkLit(v1, true))
		s.SolveLimited(dummy) //-1

		start := time.Now()
		s.Simplify()
		slv2 <- s.Solve(options)
		end := time.Now()
		sat2 <- s
		t2 <- (end.Sub(start)).Seconds()
	}()

	st := time.Now()

	for i:=0; i<2; i++{
		select {
		case slv = <-slv1:	
			sat = *<-sat1
			t = <-t1
		case slv = <-slv2:
			sat = *<-sat2
			t = <-t2
		}
		gl := time.Now()
		if slv == gomisat.LTrue {
			break 
		}
		t = (gl.Sub(st)).Seconds()
	}
	
	fmt.Println(slv)
	fmt.Println(sat.Conflicts)
	fmt.Println(sat.Propagations)
	//fmt.Printf("Memory used : %.2f MB\n", sat.Mem_used())
	fmt.Printf("computation time : %.6f (sec)\n", t)
}


/*
//2^2
func main() {
	fname := os.Args[1]
	file, err := os.Open(fname)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	buf, _ := io.ReadAll(file)
	cs, _ := gomisat.ParseDimacs(buf)
	options := gomisat.DefaultSolverOptions()

	var (
		sat gomisat.Solver
		slv gomisat.LBool
		t float64
	)
	sat1 := make(chan *gomisat.Solver)
	slv1 := make(chan gomisat.LBool)
	sat2 := make(chan *gomisat.Solver)
	slv2 := make(chan gomisat.LBool)
	sat3 := make(chan *gomisat.Solver)
	slv3 := make(chan gomisat.LBool)
	sat4 := make(chan *gomisat.Solver)
	slv4 := make(chan gomisat.LBool)
	t1 := make(chan float64)
	t2 := make(chan float64)
	t3 := make(chan float64)
	t4 := make(chan float64)

	var v1,v2 gomisat.Var = 15, 62

	go func(){
		s := gomisat.NewSolver()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}

		var dummy []gomisat.Lit
		dummy = append(dummy, gomisat.MkLit(v1, false))
		dummy = append(dummy, gomisat.MkLit(v2, false))
		s.SolveLimited(dummy) //1, 2

		start := time.Now()
		s.Simplify()
		slv1 <- s.Solve(options)
		end := time.Now()
		sat1 <- s
		t1 <- (end.Sub(start)).Seconds()
	}()

	go func(){
		s := gomisat.NewSolver()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}

		var dummy []gomisat.Lit
		dummy = append(dummy, gomisat.MkLit(v1, true))
		dummy = append(dummy, gomisat.MkLit(v2, false))
		s.SolveLimited(dummy) //-1, 2

		start := time.Now()
		s.Simplify()
		slv2 <- s.Solve(options)
		end := time.Now()
		sat2 <- s
		t2 <- (end.Sub(start)).Seconds()
	}()

	go func(){
		s := gomisat.NewSolver()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}

		var dummy []gomisat.Lit
		dummy = append(dummy, gomisat.MkLit(v1, false))
		dummy = append(dummy, gomisat.MkLit(v2, true))
		s.SolveLimited(dummy) //1, -2

		start := time.Now()
		s.Simplify()
		slv3 <- s.Solve(options)
		end := time.Now()
		sat3 <- s
		t3 <- (end.Sub(start)).Seconds()
	}()

	go func(){
		s := gomisat.NewSolver()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}

		var dummy []gomisat.Lit
		dummy = append(dummy, gomisat.MkLit(v1, true))
		dummy = append(dummy, gomisat.MkLit(v2, true))
		s.SolveLimited(dummy) //-1, -2

		start := time.Now()
		s.Simplify()
		slv4 <- s.Solve(options)
		end := time.Now()
		sat4 <- s
		t4 <- (end.Sub(start)).Seconds()
	}()

	st := time.Now()

	for i:=0; i<4; i++{
		select {
		case slv = <-slv1:	
			sat = *<-sat1
			t = <-t1
		case slv = <-slv2:
			sat = *<-sat2
			t = <-t2
		case slv = <-slv3:
			sat = *<-sat3
			t = <-t3
		case slv = <-slv4:
			sat = *<-sat4
			t = <-t4
		}
		gl := time.Now()
		if slv == gomisat.LTrue {
			break 
		}
		t = (gl.Sub(st)).Seconds()
	}
	
	fmt.Println(slv)
	fmt.Println(sat.Conflicts)
	fmt.Println(sat.Propagations)
	//fmt.Printf("Memory used : %.2f MB\n", sat.Mem_used())
	fmt.Printf("computation time : %.6f (sec)\n", t)
}
*/

/*
//2^3
func main() {
	fname := os.Args[1]
	file, err := os.Open(fname)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	buf, _ := io.ReadAll(file)
	cs, _ := gomisat.ParseDimacs(buf)
	options := gomisat.DefaultSolverOptions()

	var (
		sat gomisat.Solver
		slv gomisat.LBool
		t float64
	)
	sat1 := make(chan *gomisat.Solver)
	slv1 := make(chan gomisat.LBool)
	sat2 := make(chan *gomisat.Solver)
	slv2 := make(chan gomisat.LBool)
	sat3 := make(chan *gomisat.Solver)
	slv3 := make(chan gomisat.LBool)
	sat4 := make(chan *gomisat.Solver)
	slv4 := make(chan gomisat.LBool)
	sat5 := make(chan *gomisat.Solver)
	slv5 := make(chan gomisat.LBool)
	sat6 := make(chan *gomisat.Solver)
	slv6 := make(chan gomisat.LBool)
	sat7 := make(chan *gomisat.Solver)
	slv7 := make(chan gomisat.LBool)
	sat8 := make(chan *gomisat.Solver)
	slv8 := make(chan gomisat.LBool)
	t1 := make(chan float64)
	t2 := make(chan float64)
	t3 := make(chan float64)
	t4 := make(chan float64)
	t5 := make(chan float64)
	t6 := make(chan float64)
	t7 := make(chan float64)
	t8 := make(chan float64)

	var v1,v2,v3 gomisat.Var = 15, 62, 127

	go func(){
		s := gomisat.NewSolver()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}

		var dummy []gomisat.Lit
		dummy = append(dummy, gomisat.MkLit(v1, false))
		dummy = append(dummy, gomisat.MkLit(v2, false))
		dummy = append(dummy, gomisat.MkLit(v3, false))
		s.SolveLimited(dummy) //1, 2, 3

		start := time.Now()
		s.Simplify()
		slv1 <- s.Solve(options)
		end := time.Now()
		sat1 <- s
		t1 <- (end.Sub(start)).Seconds()
	}()

	go func(){
		s := gomisat.NewSolver()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}

		var dummy []gomisat.Lit
		dummy = append(dummy, gomisat.MkLit(v1, true))
		dummy = append(dummy, gomisat.MkLit(v2, false))
		dummy = append(dummy, gomisat.MkLit(v3, false))
		s.SolveLimited(dummy) //-1, 2, 3

		start := time.Now()
		s.Simplify()
		slv2 <- s.Solve(options)
		end := time.Now()
		sat2 <- s
		t2 <- (end.Sub(start)).Seconds()
	}()

	go func(){
		s := gomisat.NewSolver()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}

		var dummy []gomisat.Lit
		dummy = append(dummy, gomisat.MkLit(v1, false))
		dummy = append(dummy, gomisat.MkLit(v2, true))
		dummy = append(dummy, gomisat.MkLit(v3, false))
		s.SolveLimited(dummy) //1, -2, 3

		start := time.Now()
		s.Simplify()
		slv3 <- s.Solve(options)
		end := time.Now()
		sat3 <- s
		t3 <- (end.Sub(start)).Seconds()
	}()

	go func(){
		s := gomisat.NewSolver()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}

		var dummy []gomisat.Lit
		dummy = append(dummy, gomisat.MkLit(v1, true))
		dummy = append(dummy, gomisat.MkLit(v2, true))
		dummy = append(dummy, gomisat.MkLit(v3, false))
		s.SolveLimited(dummy) //-1, -2, 3

		start := time.Now()
		s.Simplify()
		slv4 <- s.Solve(options)
		end := time.Now()
		sat4 <- s
		t4 <- (end.Sub(start)).Seconds()
	}()

	go func(){
		s := gomisat.NewSolver()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}

		var dummy []gomisat.Lit
		dummy = append(dummy, gomisat.MkLit(v1, false))
		dummy = append(dummy, gomisat.MkLit(v2, false))
		dummy = append(dummy, gomisat.MkLit(v3, true))
		s.SolveLimited(dummy) //1, 2, -3

		start := time.Now()
		s.Simplify()
		slv5 <- s.Solve(options)
		end := time.Now()
		sat5 <- s
		t5 <- (end.Sub(start)).Seconds()
	}()

	go func(){
		s := gomisat.NewSolver()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}

		var dummy []gomisat.Lit
		dummy = append(dummy, gomisat.MkLit(v1, true))
		dummy = append(dummy, gomisat.MkLit(v2, false))
		dummy = append(dummy, gomisat.MkLit(v3, true))
		s.SolveLimited(dummy) //-1, 2, -3

		start := time.Now()
		s.Simplify()
		slv6 <- s.Solve(options)
		end := time.Now()
		sat6 <- s
		t6 <- (end.Sub(start)).Seconds()
	}()

	go func(){
		s := gomisat.NewSolver()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}

		var dummy []gomisat.Lit
		dummy = append(dummy, gomisat.MkLit(v1, false))
		dummy = append(dummy, gomisat.MkLit(v2, true))
		dummy = append(dummy, gomisat.MkLit(v3, true))
		s.SolveLimited(dummy) //1, -2, -3

		start := time.Now()
		s.Simplify()
		slv7 <- s.Solve(options)
		end := time.Now()
		sat7 <- s
		t7 <- (end.Sub(start)).Seconds()
	}()

	go func(){
		s := gomisat.NewSolver()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}

		var dummy []gomisat.Lit
		dummy = append(dummy, gomisat.MkLit(v1, true))
		dummy = append(dummy, gomisat.MkLit(v2, true))
		dummy = append(dummy, gomisat.MkLit(v3, true))
		s.SolveLimited(dummy) //-1, -2, -3

		start := time.Now()
		s.Simplify()
		slv8 <- s.Solve(options)
		end := time.Now()
		sat8 <- s
		t8 <- (end.Sub(start)).Seconds()
	}()

	st := time.Now()

	for i:=0; i<8; i++{
		select {
		case slv = <-slv1:	
			sat = *<-sat1
			t = <-t1
		case slv = <-slv2:
			sat = *<-sat2
			t = <-t2
		case slv = <-slv3:
			sat = *<-sat3
			t = <-t3
		case slv = <-slv4:
			sat = *<-sat4
			t = <-t4
		case slv = <-slv5:
			sat = *<-sat5
			t = <-t5
		case slv = <-slv6:
			sat = *<-sat6
			t = <-t6
		case slv = <-slv7:
			sat = *<-sat7
			t = <-t7
		case slv = <-slv8:
			sat = *<-sat8
			t = <-t8
		}
		gl := time.Now()
		if slv == gomisat.LTrue {
			break 
		}
		t = (gl.Sub(st)).Seconds()
	}
	
	fmt.Println(slv)
	fmt.Println(sat.Conflicts)
	fmt.Println(sat.Propagations)
	//fmt.Printf("Memory used : %.2f MB\n", sat.Mem_used())
	fmt.Printf("computation time : %.6f (sec)\n", t)
}
*/

/*
//2^4
func main() {
	fname := os.Args[1]
	file, err := os.Open(fname)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	buf, _ := io.ReadAll(file)
	cs, _ := gomisat.ParseDimacs(buf)
	options := gomisat.DefaultSolverOptions()

	var (
		sat gomisat.Solver
		slv gomisat.LBool
		t float64
	)
	sat1 := make(chan *gomisat.Solver)
	slv1 := make(chan gomisat.LBool)
	sat2 := make(chan *gomisat.Solver)
	slv2 := make(chan gomisat.LBool)
	sat3 := make(chan *gomisat.Solver)
	slv3 := make(chan gomisat.LBool)
	sat4 := make(chan *gomisat.Solver)
	slv4 := make(chan gomisat.LBool)
	sat5 := make(chan *gomisat.Solver)
	slv5 := make(chan gomisat.LBool)
	sat6 := make(chan *gomisat.Solver)
	slv6 := make(chan gomisat.LBool)
	sat7 := make(chan *gomisat.Solver)
	slv7 := make(chan gomisat.LBool)
	sat8 := make(chan *gomisat.Solver)
	slv8 := make(chan gomisat.LBool)
	sat9 := make(chan *gomisat.Solver)
	slv9 := make(chan gomisat.LBool)
	sat10 := make(chan *gomisat.Solver)
	slv10 := make(chan gomisat.LBool)
	sat11 := make(chan *gomisat.Solver)
	slv11 := make(chan gomisat.LBool)
	sat12 := make(chan *gomisat.Solver)
	slv12 := make(chan gomisat.LBool)
	sat13 := make(chan *gomisat.Solver)
	slv13 := make(chan gomisat.LBool)
	sat14 := make(chan *gomisat.Solver)
	slv14 := make(chan gomisat.LBool)
	sat15 := make(chan *gomisat.Solver)
	slv15 := make(chan gomisat.LBool)
	sat16 := make(chan *gomisat.Solver)
	slv16 := make(chan gomisat.LBool)
	t1 := make(chan float64)
	t2 := make(chan float64)
	t3 := make(chan float64)
	t4 := make(chan float64)
	t5 := make(chan float64)
	t6 := make(chan float64)
	t7 := make(chan float64)
	t8 := make(chan float64)
	t9 := make(chan float64)
	t10 := make(chan float64)
	t11 := make(chan float64)
	t12 := make(chan float64)
	t13 := make(chan float64)
	t14 := make(chan float64)
	t15 := make(chan float64)
	t16 := make(chan float64)


	var v1,v2,v3,v4 gomisat.Var = 26, 53, 137, 196

	go func(){
		s := gomisat.NewSolver()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}

		var dummy []gomisat.Lit
		dummy = append(dummy, gomisat.MkLit(v1, false))
		dummy = append(dummy, gomisat.MkLit(v2, false))
		dummy = append(dummy, gomisat.MkLit(v3, false))
		dummy = append(dummy, gomisat.MkLit(v4, false))
		s.SolveLimited(dummy) //1, 2, 3, 4

		start := time.Now()
		s.Simplify()
		slv1 <- s.Solve(options)
		end := time.Now()
		sat1 <- s
		t1 <- (end.Sub(start)).Seconds()
	}()

	go func(){
		s := gomisat.NewSolver()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}

		var dummy []gomisat.Lit
		dummy = append(dummy, gomisat.MkLit(v1, true))
		dummy = append(dummy, gomisat.MkLit(v2, false))
		dummy = append(dummy, gomisat.MkLit(v3, false))
		dummy = append(dummy, gomisat.MkLit(v4, false))
		s.SolveLimited(dummy) //-1, 2, 3, 4

		start := time.Now()
		s.Simplify()
		slv2 <- s.Solve(options)
		end := time.Now()
		sat2 <- s
		t2 <- (end.Sub(start)).Seconds()
	}()

	go func(){
		s := gomisat.NewSolver()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}

		var dummy []gomisat.Lit
		dummy = append(dummy, gomisat.MkLit(v1, false))
		dummy = append(dummy, gomisat.MkLit(v2, true))
		dummy = append(dummy, gomisat.MkLit(v3, false))
		dummy = append(dummy, gomisat.MkLit(v4, false))
		s.SolveLimited(dummy) //1, -2, 3, 4

		start := time.Now()
		s.Simplify()
		slv3 <- s.Solve(options)
		end := time.Now()
		sat3 <- s
		t3 <- (end.Sub(start)).Seconds()
	}()

	go func(){
		s := gomisat.NewSolver()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}

		var dummy []gomisat.Lit
		dummy = append(dummy, gomisat.MkLit(v1, true))
		dummy = append(dummy, gomisat.MkLit(v2, true))
		dummy = append(dummy, gomisat.MkLit(v3, false))
		dummy = append(dummy, gomisat.MkLit(v4, false))
		s.SolveLimited(dummy) //-1, -2, 3, 4

		start := time.Now()
		s.Simplify()
		slv4 <- s.Solve(options)
		end := time.Now()
		sat4 <- s
		t4 <- (end.Sub(start)).Seconds()
	}()

	go func(){
		s := gomisat.NewSolver()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}

		var dummy []gomisat.Lit
		dummy = append(dummy, gomisat.MkLit(v1, false))
		dummy = append(dummy, gomisat.MkLit(v2, false))
		dummy = append(dummy, gomisat.MkLit(v3, true))
		dummy = append(dummy, gomisat.MkLit(v4, false))
		s.SolveLimited(dummy) //1, 2, -3, 4

		start := time.Now()
		s.Simplify()
		slv5 <- s.Solve(options)
		end := time.Now()
		sat5 <- s
		t5 <- (end.Sub(start)).Seconds()
	}()

	go func(){
		s := gomisat.NewSolver()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}

		var dummy []gomisat.Lit
		dummy = append(dummy, gomisat.MkLit(v1, true))
		dummy = append(dummy, gomisat.MkLit(v2, false))
		dummy = append(dummy, gomisat.MkLit(v3, true))
		dummy = append(dummy, gomisat.MkLit(v4, false))
		s.SolveLimited(dummy) //-1, 2, -3, 4

		start := time.Now()
		s.Simplify()
		slv6 <- s.Solve(options)
		end := time.Now()
		sat6 <- s
		t6 <- (end.Sub(start)).Seconds()
	}()

	go func(){
		s := gomisat.NewSolver()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}

		var dummy []gomisat.Lit
		dummy = append(dummy, gomisat.MkLit(v1, false))
		dummy = append(dummy, gomisat.MkLit(v2, true))
		dummy = append(dummy, gomisat.MkLit(v3, true))
		dummy = append(dummy, gomisat.MkLit(v4, false))
		s.SolveLimited(dummy) //1, -2, -3, 4

		start := time.Now()
		s.Simplify()
		slv7 <- s.Solve(options)
		end := time.Now()
		sat7 <- s
		t7 <- (end.Sub(start)).Seconds()
	}()

	go func(){
		s := gomisat.NewSolver()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}

		var dummy []gomisat.Lit
		dummy = append(dummy, gomisat.MkLit(v1, true))
		dummy = append(dummy, gomisat.MkLit(v2, true))
		dummy = append(dummy, gomisat.MkLit(v3, true))
		dummy = append(dummy, gomisat.MkLit(v4, false))
		s.SolveLimited(dummy) //-1, -2, -3, 4

		start := time.Now()
		s.Simplify()
		slv8 <- s.Solve(options)
		end := time.Now()
		sat8 <- s
		t8 <- (end.Sub(start)).Seconds()
	}()

	go func(){
		s := gomisat.NewSolver()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}

		var dummy []gomisat.Lit
		dummy = append(dummy, gomisat.MkLit(v1, false))
		dummy = append(dummy, gomisat.MkLit(v2, false))
		dummy = append(dummy, gomisat.MkLit(v3, false))
		dummy = append(dummy, gomisat.MkLit(v4, true))
		s.SolveLimited(dummy) //1, 2, 3, -4

		start := time.Now()
		s.Simplify()
		slv9 <- s.Solve(options)
		end := time.Now()
		sat9 <- s
		t9 <- (end.Sub(start)).Seconds()
	}()

	go func(){
		s := gomisat.NewSolver()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}

		var dummy []gomisat.Lit
		dummy = append(dummy, gomisat.MkLit(v1, true))
		dummy = append(dummy, gomisat.MkLit(v2, false))
		dummy = append(dummy, gomisat.MkLit(v3, false))
		dummy = append(dummy, gomisat.MkLit(v4, true))
		s.SolveLimited(dummy) //-1, 2, 3, -4

		start := time.Now()
		s.Simplify()
		slv10 <- s.Solve(options)
		end := time.Now()
		sat10 <- s
		t10 <- (end.Sub(start)).Seconds()
	}()

	go func(){
		s := gomisat.NewSolver()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}

		var dummy []gomisat.Lit
		dummy = append(dummy, gomisat.MkLit(v1, false))
		dummy = append(dummy, gomisat.MkLit(v2, true))
		dummy = append(dummy, gomisat.MkLit(v3, false))
		dummy = append(dummy, gomisat.MkLit(v4, true))
		s.SolveLimited(dummy) //1, -2, 3, -4

		start := time.Now()
		s.Simplify()
		slv11 <- s.Solve(options)
		end := time.Now()
		sat11 <- s
		t11 <- (end.Sub(start)).Seconds()
	}()

	go func(){
		s := gomisat.NewSolver()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}

		var dummy []gomisat.Lit
		dummy = append(dummy, gomisat.MkLit(v1, true))
		dummy = append(dummy, gomisat.MkLit(v2, true))
		dummy = append(dummy, gomisat.MkLit(v3, false))
		dummy = append(dummy, gomisat.MkLit(v4, true))
		s.SolveLimited(dummy) //-1, -2, 3, -4

		start := time.Now()
		s.Simplify()
		slv12 <- s.Solve(options)
		end := time.Now()
		sat12 <- s
		t12 <- (end.Sub(start)).Seconds()
	}()

	go func(){
		s := gomisat.NewSolver()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}

		var dummy []gomisat.Lit
		dummy = append(dummy, gomisat.MkLit(v1, false))
		dummy = append(dummy, gomisat.MkLit(v2, false))
		dummy = append(dummy, gomisat.MkLit(v3, true))
		dummy = append(dummy, gomisat.MkLit(v4, true))
		s.SolveLimited(dummy) //1, 2, -3, -4

		start := time.Now()
		s.Simplify()
		slv13 <- s.Solve(options)
		end := time.Now()
		sat13 <- s
		t13 <- (end.Sub(start)).Seconds()
	}()

	go func(){
		s := gomisat.NewSolver()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}

		var dummy []gomisat.Lit
		dummy = append(dummy, gomisat.MkLit(v1, true))
		dummy = append(dummy, gomisat.MkLit(v2, false))
		dummy = append(dummy, gomisat.MkLit(v3, true))
		dummy = append(dummy, gomisat.MkLit(v4, true))
		s.SolveLimited(dummy) //-1, 2, -3, -4

		start := time.Now()
		s.Simplify()
		slv14 <- s.Solve(options)
		end := time.Now()
		sat14 <- s
		t14 <- (end.Sub(start)).Seconds()
	}()

	go func(){
		s := gomisat.NewSolver()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}

		var dummy []gomisat.Lit
		dummy = append(dummy, gomisat.MkLit(v1, false))
		dummy = append(dummy, gomisat.MkLit(v2, true))
		dummy = append(dummy, gomisat.MkLit(v3, true))
		dummy = append(dummy, gomisat.MkLit(v4, true))
		s.SolveLimited(dummy) //1, -2, -3, -4

		start := time.Now()
		s.Simplify()
		slv15 <- s.Solve(options)
		end := time.Now()
		sat15 <- s
		t15 <- (end.Sub(start)).Seconds()
	}()

	go func(){
		s := gomisat.NewSolver()
		for _, x := range cs {
			s.AddClauseFromCode(x, options)
		}

		var dummy []gomisat.Lit
		dummy = append(dummy, gomisat.MkLit(v1, true))
		dummy = append(dummy, gomisat.MkLit(v2, true))
		dummy = append(dummy, gomisat.MkLit(v3, true))
		dummy = append(dummy, gomisat.MkLit(v4, true))
		s.SolveLimited(dummy) //-1, -2, -3, -4

		start := time.Now()
		s.Simplify()
		slv16 <- s.Solve(options)
		end := time.Now()
		sat16 <- s
		t16 <- (end.Sub(start)).Seconds()
	}()

	st := time.Now()

	for i:=0; i<16; i++{
		select {
		case slv = <-slv1:	
			sat = *<-sat1
			t = <-t1
		case slv = <-slv2:
			sat = *<-sat2
			t = <-t2
		case slv = <-slv3:
			sat = *<-sat3
			t = <-t3
		case slv = <-slv4:
			sat = *<-sat4
			t = <-t4
		case slv = <-slv5:
			sat = *<-sat5
			t = <-t5
		case slv = <-slv6:
			sat = *<-sat6
			t = <-t6
		case slv = <-slv7:
			sat = *<-sat7
			t = <-t7
		case slv = <-slv8:
			sat = *<-sat8
			t = <-t8
		case slv = <-slv9:
			sat = *<-sat9
			t = <-t9
		case slv = <-slv10:
			sat = *<-sat10
			t = <-t10
		case slv = <-slv11:
			sat = *<-sat11
			t = <-t11
		case slv = <-slv12:
			sat = *<-sat12
			t = <-t12
		case slv = <-slv13:
			sat = *<-sat13
			t = <-t13
		case slv = <-slv14:
			sat = *<-sat14
			t = <-t14
		case slv = <-slv15:
			sat = *<-sat15
			t = <-t15
		case slv = <-slv16:
			sat = *<-sat16
			t = <-t16
		}
		gl := time.Now()
		if slv == gomisat.LTrue {
			break 
		}
		t = (gl.Sub(st)).Seconds()
	}
	
	fmt.Println(slv)
	fmt.Println(sat.Conflicts)
	fmt.Println(sat.Propagations)
	//fmt.Printf("Memory used : %.2f MB\n", sat.Mem_used())
	fmt.Printf("computation time : %.6f (sec)\n", t)
}
*/