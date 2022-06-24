package main

import (
	"flag"
	"os"

	"github.com/Besten/internal/modules"
	"github.com/Besten/internal/runtime"
)

func args() []runtime.Object {
	res := make([]runtime.Object, len(os.Args))
	for i := range os.Args {
		res[i] = os.Args[i]
	}
	return res
}

func main() {
	var file string
	flag.StringVar(&file, "file", "", "File to be compiled")
	flag.Parse()
	if flag.NArg() > 0 {
		file = flag.Arg(0)
		if flag.NArg() != 1 {
			panic("Expecting just one source")
		}
	}
	if len(file) == 0 {
		panic("No file provided")
	}
	symbols, cname, err := modules.New().MainFile(file)
	if err != nil {
		panic(err)
	}
	vm := runtime.NewVM()
	vm.LoadSymbols(symbols)
	process, err := vm.InitSpawn(cname, []runtime.Object{runtime.MakeVec(args()...)})
	if err != nil {
		panic(err)
	}
	err = vm.Wait(process)
	if err != nil {
		panic(err)
	}
}
