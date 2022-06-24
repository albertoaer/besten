package parser

import (
	"fmt"
	"time"

	. "github.com/Besten/internal/runtime"
)

func mathOpInstruction(int_, float_ ICode) []*FunctionSymbol {
	return []*FunctionSymbol{{"none", false, MKInstruction(int_).Fragment(), &Int, []OBJType{Int, Int}},
		{"none", false, MKInstruction(float_).Fragment(), &Dec, []OBJType{Dec, Dec}}}
}

func multiTypeInstruction(length int, returntype OBJType, matches map[OBJType]Instruction) []*FunctionSymbol {
	ret := make([]*FunctionSymbol, 0)
	for k, v := range matches {
		tp := make([]OBJType, length)
		for i := range tp {
			tp[i] = k
		}
		ret = append(ret, &FunctionSymbol{
			"none",
			false,
			v.Fragment(),
			&returntype,
			tp,
		})
	}
	return ret
}

func comparisonInstruction(flags int, matches map[OBJType]ICode) []*FunctionSymbol {
	nmatch := make(map[OBJType]Instruction)
	for k, v := range matches {
		nmatch[k] = MKInstruction(v, flags)
	}
	return multiTypeInstruction(2, Bool, nmatch)
}

func wrapOpInstruction(code ICode, tp OBJType, unary bool) *FunctionSymbol {
	tps := []OBJType{tp, tp}
	if unary {
		tps = []OBJType{tp}
	}
	return &FunctionSymbol{"none", false, MKInstruction(code).Fragment(), &tp, tps}
}

func injectBuiltinFunctions(to *FunctionCollection) {
	to.AddDynamicSymbol("constructor", func(o []OBJType) *FunctionSymbol {
		if len(o) == 1 {
			/*DEFAULT CONSTRUCTOR FOR ANY TYPE*/
			return &FunctionSymbol{"none", false, make([]Instruction, 0), CloneType(o[0]), o}
		}
		return nil
	})
	to.AddDynamicSymbol("static_cast", func(o []OBJType) *FunctionSymbol {
		if len(o) == 2 && checkCompatibility(o[0], o[1]) {
			/*DEFAULT CAST FOR ANY TWO TYPES*/
			return &FunctionSymbol{"none", false, make([]Instruction, 0), CloneType(o[1]), o}
		}
		return nil
	})
	to.AddSymbol("print", &FunctionSymbol{"none", true, MKInstruction(IFD, EmbeddedFunction{
		Name:     "print",
		ArgCount: 1,
		Function: func(args []Object) Object {
			v := *args[0].(VecT)
			for i, e := range v {
				if i == len(v)-1 {
					fmt.Println(e)
				} else {
					fmt.Print(e)
				}
			}
			return nil
		},
		Returns: false,
	}).Fragment(), CloneType(Void), []OBJType{VecOf(Any)}})
	to.AddSymbol("puts", &FunctionSymbol{"none", true, MKInstruction(IFD, EmbeddedFunction{
		Name:     "puts",
		ArgCount: 1,
		Function: func(args []Object) Object {
			for _, e := range *args[0].(VecT) {
				fmt.Print(e)
			}
			return nil
		},
		Returns: false,
	}).Fragment(), CloneType(Void), []OBJType{VecOf(Any)}})
	to.AddSymbol("clock", &FunctionSymbol{"none", false, MKInstruction(IFD, EmbeddedFunction{
		Name:     "clock",
		ArgCount: 0,
		Function: func(args []Object) Object {
			return int(time.Now().UnixMicro())
		},
		Returns: true,
	}).Fragment(), CloneType(Int), []OBJType{}})
	to.AddSymbol("raw", &FunctionSymbol{"none", false, MKInstruction(IFD, EmbeddedFunction{
		Name:     "raw",
		ArgCount: 1,
		Function: func(args []Object) Object {
			return fmt.Sprintf("%v", args[0])
		},
		Returns: true,
	}).Fragment(), CloneType(Str), []OBJType{Any}})
	to.AddDynamicSymbol("stref", func(o []OBJType) *FunctionSymbol {
		if len(o) == 1 {
			return &FunctionSymbol{"none", false, []Instruction{MKInstruction(POP), MKInstruction(PSH, Repr(o[0]))}, CloneType(Str), o}
		}
		return nil
	})
	to.AddSymbol("dec", &FunctionSymbol{"none", false, MKInstruction(ITD).Fragment(), CloneType(Dec), []OBJType{Int}})
	to.AddSymbol("int", &FunctionSymbol{"none", false, MKInstruction(DTI).Fragment(), CloneType(Int), []OBJType{Dec}})
	to.AddSymbol("int", &FunctionSymbol{"none", false, []Instruction{}, CloneType(Int), []OBJType{Bool}})
	to.AddSymbol("str", &FunctionSymbol{"none", false, MKInstruction(IFD, EmbeddedFunction{
		Name:     "vec_to_str",
		ArgCount: 1,
		Function: func(args []Object) Object {
			r := make([]rune, 0)
			for _, v := range *args[0].(VecT) {
				r = append(r, rune(v.(int)))
			}
			return string(r)
		},
		Returns: true,
	}).Fragment(), CloneType(Str), []OBJType{VecOf(Int)}})
	to.AddSymbols("len", multiTypeInstruction(1, Int, map[OBJType]Instruction{Str: MKInstruction(SOS), MapOf(Any): MKInstruction(SOM), VecOf(Any): MKInstruction(SOV)}))
	to.AddDynamicSymbol("vec", func(o []OBJType) *FunctionSymbol {
		if len(o) > 0 {
			ret := VecOf(o[0])
			return &FunctionSymbol{"none", true, []Instruction{}, &ret, []OBJType{VecOf(o[0])}}
		}
		return nil
	})
	to.AddDynamicSymbol("pop", func(o []OBJType) *FunctionSymbol {
		if len(o) == 1 && o[0].Primitive() == VECTOR {
			ret := o[0].Items()
			return &FunctionSymbol{"none", false, []Instruction{MKInstruction(PFV)}, &ret, o}
		}
		return nil
	})
	to.AddDynamicSymbol("setbykey", func(o []OBJType) *FunctionSymbol {
		if len(o) == 3 {
			var ins []Instruction
			if o[2].Primitive() == VECTOR && o[1].Primitive() == INTEGER && CompareTypes(o[0], o[2].Items()) {
				ins = []Instruction{MKInstruction(SVI)}
			} else if o[2].Primitive() == MAP && o[1].Primitive() == STRING && CompareTypes(o[0], o[2].Items()) {
				ins = []Instruction{MKInstruction(ATT)}
			} else {
				return nil
			}
			return &FunctionSymbol{"none", false, ins, CloneType(Void), o}
		}
		return nil
	})
	to.AddDynamicSymbol("callfn", func(o []OBJType) *FunctionSymbol {
		if o[0].Primitive() == FUNCTION {
			fn := o[0].(*FunctionType)
			if !CompareArrayOfTypes(o[1:], fn.args) {
				return nil
			}
			return &FunctionSymbol{"none", false, MKInstruction(CLL).Fragment(), CloneType(fn.ret), o}
		}
		return nil
	})
	to.AddDynamicSymbol("callmapfn", func(o []OBJType) *FunctionSymbol {
		if len(o) == 2 && o[0].Primitive() == FUNCTION && o[1].Primitive() == TUPLE {
			fn := o[0].(*FunctionType)
			if !CompareArrayOfTypes(o[1].FixedItems(), fn.args) {
				return nil
			}
			return &FunctionSymbol{"none", false, MKInstruction(CLX).Fragment(), CloneType(fn.ret), o}
		}
		return nil
	})
	to.AddDynamicSymbol("argsfn", func(o []OBJType) *FunctionSymbol {
		if len(o) == 1 && o[0].Primitive() == FUNCTION {
			fn := o[0].(*FunctionType)
			tp := TupleOf(fn.args)
			tpins, e := tp.Create()
			if e != nil {
				return nil
			}
			return &FunctionSymbol{"none", false, tpins, &tp, o}
		}
		return nil
	})
	to.AddDynamicSymbol("end", func(o []OBJType) *FunctionSymbol {
		if len(o) == 1 {
			return &FunctionSymbol{"none", false,
				[]Instruction{MKInstruction(POP), MKInstruction(PSH, 0)}, CloneType(Bool), o}
		}
		return nil
	})
	to.AddSymbol("not", wrapOpInstruction(NOTB, Bool, true))
}

func injectBuiltinOperators(to *FunctionCollection) {
	to.AddSymbols("+", mathOpInstruction(ADD, ADDF))
	to.AddSymbols("-", mathOpInstruction(SUB, SUBF))
	to.AddSymbol("-", &FunctionSymbol{"none", false, MKInstruction(SUB, 0).Fragment(), CloneType(Int), []OBJType{Int}})
	to.AddSymbol("-", &FunctionSymbol{"none", false, MKInstruction(SUBF, 0).Fragment(), CloneType(Dec), []OBJType{Dec}})
	to.AddSymbols("*", mathOpInstruction(MUL, MULF))
	to.AddSymbols("/", mathOpInstruction(DIV, DIVF))
	to.AddSymbol("<<", wrapOpInstruction(SHL, Int, false))
	to.AddSymbol(">>", wrapOpInstruction(SHR, Int, false))
	to.AddSymbol("%", wrapOpInstruction(MOD, Int, false))
	to.AddSymbol("!", wrapOpInstruction(NOT, Int, true))
	to.AddSymbol("&", wrapOpInstruction(AND, Int, false))
	to.AddSymbol("|", wrapOpInstruction(OR, Int, false))
	to.AddSymbol("^", wrapOpInstruction(XOR, Int, false))
	to.AddSymbol("!", wrapOpInstruction(NOTB, Bool, true))
	to.AddSymbol("&&", wrapOpInstruction(AND, Bool, false))
	to.AddSymbol("||", wrapOpInstruction(OR, Bool, false))
	to.AddSymbol("^^", wrapOpInstruction(XOR, Bool, false))
	to.AddSymbols("==", comparisonInstruction(1, map[OBJType]ICode{Int: CMPI, Dec: CMPF, Bool: CMPI}))
	to.AddSymbols("!=", comparisonInstruction(5, map[OBJType]ICode{Int: CMPI, Dec: CMPF, Bool: CMPI}))
	to.AddSymbols("<", comparisonInstruction(2, map[OBJType]ICode{Int: CMPI, Dec: CMPF}))
	to.AddSymbols(">", comparisonInstruction(7, map[OBJType]ICode{Int: CMPI, Dec: CMPF}))
	to.AddSymbols("<=", comparisonInstruction(3, map[OBJType]ICode{Int: CMPI, Dec: CMPF}))
	to.AddSymbols(">=", comparisonInstruction(6, map[OBJType]ICode{Int: CMPI, Dec: CMPF}))
	to.AddDynamicSymbol("[]", func(o []OBJType) *FunctionSymbol {
		if len(o) == 2 {
			var ins []Instruction
			var ret *OBJType
			if o[0].Primitive() == VECTOR && o[1].Primitive() == INTEGER {
				ins = []Instruction{MKInstruction(ACC)}
				ret = CloneType(o[0].Items())
			} else if o[0].Primitive() == MAP && o[1].Primitive() == STRING {
				ins = []Instruction{MKInstruction(PRP)}
				ret = CloneType(TupleOf([]OBJType{o[0].Items(), Bool}))
			} else {
				return nil
			}
			return &FunctionSymbol{"none", false, ins, ret, o}
		}
		return nil
	})
	to.AddDynamicSymbol("->", func(o []OBJType) *FunctionSymbol {
		if len(o) == 2 {
			var ins []Instruction
			if o[1].Primitive() == VECTOR && CompareTypes(o[0], o[1].Items()) {
				ins = []Instruction{MKInstruction(SWT), MKInstruction(APP)}
			} else if o[1].Primitive() == MAP && CompareTypes(o[0], TupleOf([]OBJType{o[1].Items(), Str})) {
				ins = []Instruction{MKInstruction(EIS), MKInstruction(ATT)}
			} else {
				return nil
			}
			return &FunctionSymbol{"none", false, ins, CloneType(Void), o}
		}
		return nil
	})
	to.AddDynamicSymbol("'", func(o []OBJType) *FunctionSymbol {
		if len(o) == 1 {
			if o[0].Primitive() == VECTOR {
				v := VariadicOf(o[0].Items())
				return &FunctionSymbol{"none", false, []Instruction{}, &v, o}
			} else if o[0].Primitive() == VARIADIC {
				v := VecOf(o[0].Items())
				return &FunctionSymbol{"none", false, []Instruction{}, &v, o}
			}
		}
		return nil
	})
	to.AddDynamicSymbol("*", func(o []OBJType) *FunctionSymbol {
		if len(o) == 1 {
			if o[0].Primitive() == ALIAS {
				r := o[0].(*Alias).Holds
				return &FunctionSymbol{"none", false, []Instruction{}, &r, o}
			} else if o[0].Primitive() == STRUCT {
				t := TupleOf(o[0].FixedItems())
				return &FunctionSymbol{"none", false, []Instruction{}, &t, o}
			}
		}
		return nil
	})
	to.AddSymbol("*", &FunctionSymbol{"none", false, MKInstruction(IFD, EmbeddedFunction{
		Name:     "str_to_vec",
		ArgCount: 1,
		Function: func(args []Object) Object {
			r := make([]Object, 0)
			for _, v := range []rune(args[0].(string)) {
				r = append(r, int(v))
			}
			return VecT(&r)
		},
		Returns: true,
	}).Fragment(), CloneType(VecOf(Int)), []OBJType{Str}})
	to.AddDynamicSymbol("%%", func(o []OBJType) *FunctionSymbol {
		if len(o) == 2 {
			i := 0
			if CompareTypes(o[0], o[1]) {
				i = 1
			}
			return &FunctionSymbol{"none", false, MKInstruction(PSH, i).Fragment(), CloneType(Bool), o}
		}
		return nil
	})
}
