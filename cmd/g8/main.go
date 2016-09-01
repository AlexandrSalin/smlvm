package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"e8vm.io/e8vm/arch8"
	"e8vm.io/e8vm/dasm"
	"e8vm.io/e8vm/glang"
	"e8vm.io/e8vm/lexing"
)

func exit(e error) {
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
	}
	os.Exit(-1)
}

func printErrs(es []*lexing.Error) {
	if len(es) == 0 {
		return
	}
	for _, e := range es {
		fmt.Println(e)
	}
	exit(nil)
}

var (
	bare       = flag.Bool("bare", false, "parse as bare function")
	ir         = flag.Bool("ir", false, "prints out the IR")
	doDasm     = flag.Bool("d", false, "deassemble the image")
	ncycle     = flag.Int("n", 100000, "maximum number of cycles")
	ncycleTest = flag.Int("ntest", 0, "maximum number of cycles for tests")
	verbose    = flag.Bool("v", false, "verbose")
	golike     = flag.Bool("golike", false, "using strict go-like syntax")
)

func main() {
	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		exit(errors.New("need exactly one input input file"))
	}
	fname := args[0]
	input, e := ioutil.ReadFile(fname)
	if e != nil {
		exit(e)
	}

	if *bare {
		bs, es, irLog := glang.CompileBareFunc(fname, string(input))
		printErrs(es)
		printIRLog(irLog, *ir)
		runImage(bs, *doDasm, *ncycle)
	} else {
		bs, es, irLog := glang.CompileAndTestSingle(
			fname, string(input), *golike, *ncycleTest,
		)
		printErrs(es)
		printIRLog(irLog, *ir)
		runImage(bs, *doDasm, *ncycle)
	}
}

func runImage(bs []byte, doDasm bool, n int) {
	if doDasm {
		err := dasm.DumpImage(bytes.NewReader(bs), os.Stdout)
		if err != nil {
			fmt.Println(err)
		}
	}
	if len(bs) == 0 {
		fmt.Println("(the image is empty)")
		return
	}

	m := arch8.NewMachine(new(arch8.Config))
	if err := m.LoadImageBytes(bs); err != nil {
		fmt.Println(err)
		return
	}

	ncycle, exp := m.Run(n)
	fmt.Printf("(%d cycles)\n", ncycle)
	if exp != nil {
		if !arch8.IsHalt(exp) {
			fmt.Println(exp)
			err := arch8.FprintStack(os.Stdout, m, exp)
			if err != nil {
				fmt.Println(err)
			}
		}
	} else {
		fmt.Println("(end of time)")
	}
}

func printIRLog(irLog []byte, ir bool) {
	if !ir {
		return
	}
	if irLog == nil {
		fmt.Println("(no IR log produced)")
	} else {
		fmt.Println(string(irLog))
	}
}
