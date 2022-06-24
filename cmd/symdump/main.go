package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/Besten/internal/modules"
	"github.com/Besten/internal/runtime"
)

func itable(path string) map[int]string {
	b, e := ioutil.ReadFile(path)
	if e != nil {
		panic(e)
	}
	result := make(map[int]string)
	lines := strings.Split(string(b), "\n")
	flag := false
	for i := 0; i < len(lines); i++ {
		if flag {
			n := strings.Split(lines[i], "/")
			if len(n) > 0 {
				target := strings.TrimSpace(n[0])
				if target == ")" {
					break
				}
				if len(target) > 0 && target[0] != '/' {
					parts := strings.Split(target, "=")
					ids := strings.Split(parts[0], " ")
					icode, e := strconv.Atoi(strings.TrimSpace(parts[1]))
					if e != nil {
						panic(e)
					}
					result[icode] = ids[0]
				}
			}
		} else if strings.TrimSpace(lines[i]) == "const (" {
			flag = true
		}
	}
	return result
}

func dumpInto(s runtime.Symbol, table map[int]string, dest io.Writer) {
	fmt.Fprintln(dest, "Dumped from: ", s.Name)
	for i, v := range s.Source {
		fmt.Fprintf(dest, "\t%4d %s", i, table[int(v.Code)])
		for _, v := range v.Inspect() {
			fmt.Fprint(dest, " ", v)
		}
		fmt.Fprintln(dest)
	}
}

func main() {
	defer func() {
		if e := recover(); e != nil {
			fmt.Println("Error while dumping: \n\t", e)
		}
	}()
	table := itable("./internal/runtime/instructionset.go")
	var file string
	var name string
	flag.StringVar(&file, "file", "", "File to be compiled")
	flag.StringVar(&name, "name", "", "Prefix to compare the symbol compilation name")
	flag.Parse()
	if len(file) == 0 {
		panic("No file provided")
	}
	symbols, _, err := modules.New().MainFile(file)
	if err != nil {
		panic(err)
	}
	for k, s := range symbols {
		if strings.HasPrefix(k, name) {
			dumpInto(s, table, os.Stdout)
		}
	}
	fmt.Println("Dumping done!")
}
