package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/besten/internal/modules"
	"github.com/besten/internal/runtime"
)

func args() []runtime.Object {
	res := make([]runtime.Object, len(os.Args))
	for i := range os.Args {
		res[i] = os.Args[i]
	}
	return res
}

func main() {
	var step string = "compilation"
	defer func() {
		if e := recover(); e != nil {
			fmt.Printf("besten error during %s: %s\n\t", step, e)
		}
	}()
	var file string
	flag.StringVar(&file, "file", "", "File to be compiled")
	flag.Parse()
	if flag.NArg() > 0 {
		file = flag.Arg(0)
	}
	if len(file) == 0 {
		panic("No file provided")
	}
	symbols, cname, err := modules.New().MainFile(file)
	if err != nil {
		panic(err)
	}
	step = "execution"
	vm := runtime.NewVM()
	vm.LoadSymbols(symbols)
	/*{
		f, err := os.Create("cpu.prof")
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}*/
	process, err := vm.InitSpawn(cname, []runtime.Object{runtime.MakeVec(args()...)})
	if err != nil {
		panic(err)
	}
	err = vm.Wait(process)
	if err != nil {
		panic(err)
	}
}
