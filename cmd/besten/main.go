package main

import (
	"flag"

	"github.com/Besten/internal/modules"
)

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
	_, _, err := modules.New().File(file)
	if err != nil {
		panic(err)
	}
}
