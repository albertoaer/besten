package parser

import (
	"errors"

	"github.com/Besten/internal/runtime"
)

type PrimitiveType uint8

const (
	VOID     PrimitiveType = 0
	ANY      PrimitiveType = 1
	NULL     PrimitiveType = 2
	INTEGER  PrimitiveType = 3
	DECIMAL  PrimitiveType = 4
	BOOL     PrimitiveType = 5
	STRING   PrimitiveType = 6
	VECTOR   PrimitiveType = 7
	MAP      PrimitiveType = 8
	TUPLE    PrimitiveType = 9
	STRUCT   PrimitiveType = 10
	VARIADIC PrimitiveType = 11
	ALIAS    PrimitiveType = 12
	FUNCTION PrimitiveType = 13
)

func FnCArrRepr(arr []OBJType) string {
	return ArrRepr(arr, '(', ')')
}

func ArrRepr(arr []OBJType, o, c rune) string {
	rep := string(o)
	for i, a := range arr {
		rep += Repr(a)
		if i < len(arr)-1 {
			rep += ","
		}
	}
	rep += string(c)
	return rep
}

func Repr(a OBJType) string {
	base := a.TypeName()
	switch a.Primitive() {
	case VECTOR, MAP, VARIADIC:
		base += "|" + Repr(a.Items())
	case TUPLE:
		return ArrRepr(a.FixedItems(), '{', '}')
	case FUNCTION:
		fn := a.(*FunctionType)
		return ArrRepr(fn.args, '{', '}') + "|" + Repr(fn.ret)
	}
	return base
}

type OBJType interface {
	TypeName() string                       //For identifying type
	Primitive() PrimitiveType               //For basic object kind identification
	Items() OBJType                         //For containers with unique and unnamed types
	FixedItems() []OBJType                  //For tuple and structures with fixed fields
	NamedItems() map[string]int             //For structures with fixed and indexed fields
	Create() ([]runtime.Instruction, error) //Instructions to create the object
}

func CloneType(o OBJType) *OBJType {
	newtp := o
	return &newtp
}

func CompareArrayOfTypes(a, b []OBJType) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !CompareTypes(a[i], b[i]) {
			return false
		}
	}
	return true
}

func CompareTypes(a, b OBJType) bool {
	if a.Primitive() == ANY || b.Primitive() == ANY {
		return true
	}
	if a.Primitive() == ALIAS || b.Primitive() == ALIAS {
		return a.TypeName() == b.TypeName()
	}
	if a.Primitive() == FUNCTION || b.Primitive() == FUNCTION {
		at := (a.(*FunctionType))
		bt := (b.(*FunctionType))
		return CompareArrayOfTypes(at.args, bt.args) && CompareTypes(at.ret, bt.ret)
	}
	if a.Primitive() != b.Primitive() {
		return false
	}
	switch a.Primitive() {
	case VECTOR, MAP, VARIADIC:
		return CompareTypes(a.Items(), b.Items())
	case STRUCT, TUPLE:
		return CompareArrayOfTypes(a.FixedItems(), b.FixedItems())
	}
	return true
}

func checkCompatibility(from, to OBJType) bool {
	switch to.Primitive() {
	case ALIAS:
		a := to.(*Alias).Holds
		b := from
		if b.Primitive() == ALIAS {
			b = b.(*Alias).Holds
		}
		return CompareTypes(a, b)
	case STRUCT:
		a := to.(*Structure).ItemTypes
		if from.Primitive() == TUPLE {
			return CompareTypes(TupleOf(a), from)
		}
		if from.Primitive() == STRUCT {
			return CompareArrayOfTypes(a, from.(*Structure).ItemTypes)
		}
	}
	return false
}

type Literal struct {
	RawType    PrimitiveType
	Name       string
	DefaultObj runtime.Object
}

var (
	Void OBJType = &Literal{VOID, "Void", nil}
	Int  OBJType = &Literal{INTEGER, "Int", 0}
	Dec  OBJType = &Literal{DECIMAL, "Dec", float64(0.0)}
	Bool OBJType = &Literal{BOOL, "Bool", true}
	Str  OBJType = &Literal{STRING, "Str", ""}
	Any  OBJType = &Literal{ANY, "Any", nil}
)

func (nc *Literal) TypeName() string {
	return nc.Name
}

func (nc *Literal) Primitive() PrimitiveType {
	return nc.RawType
}

func (nc *Literal) Items() OBJType {
	return nil
}

func (nc *Literal) FixedItems() []OBJType {
	return nil
}

func (nc *Literal) NamedItems() map[string]int {
	return nil
}

func (nc *Literal) Create() ([]runtime.Instruction, error) {
	return runtime.MKInstruction(runtime.PSH, nc.DefaultObj).Fragment(), nil
}

type Container struct {
	ContainerType     PrimitiveType
	ItemsType         OBJType
	Name              string
	CreateInstruction runtime.ICode
}

func VecOf(t OBJType) OBJType {
	return &Container{VECTOR, t, "Vec", runtime.VEC}
}

func MapOf(t OBJType) OBJType {
	return &Container{MAP, t, "Map", runtime.KVC}
}

func VariadicOf(t OBJType) OBJType {
	return &Container{VARIADIC, t, "Variadic", runtime.VEC}
}

func (nc *Container) TypeName() string {
	return nc.Name
}

func (nc *Container) Primitive() PrimitiveType {
	return nc.ContainerType
}

func (nc *Container) Items() OBJType {
	return nc.ItemsType
}

func (nc *Container) FixedItems() []OBJType {
	return nil
}

func (nc *Container) NamedItems() map[string]int {
	return nil
}

func (nc *Container) Create() ([]runtime.Instruction, error) {
	return runtime.MKInstruction(nc.CreateInstruction).Fragment(), nil
}

type Tuple struct {
	ItemTypes []OBJType
}

func TupleOf(items []OBJType) OBJType {
	return &Tuple{items}
}

func (nc *Tuple) TypeName() string {
	return "Tuple"
}

func (nc *Tuple) Primitive() PrimitiveType {
	return TUPLE
}

func (nc *Tuple) Items() OBJType {
	return nil
}

func (nc *Tuple) FixedItems() []OBJType {
	return nc.ItemTypes
}

func (nc *Tuple) NamedItems() map[string]int {
	return nil
}

func (nc *Tuple) Create() ([]runtime.Instruction, error) {
	data := make([]runtime.Instruction, 0)
	for _, v := range nc.ItemTypes {
		i, e := v.Create()
		if e != nil {
			return nil, e
		}
		data = append(i, data...)
	}
	return append(data, runtime.MKInstruction(runtime.CSE, len(nc.ItemTypes))), nil
}

type Alias struct {
	Holds OBJType
	Name  string
}

func AliasFor(name string, obj OBJType) OBJType {
	if obj.Primitive() == ALIAS {
		obj = obj.(*Alias).Holds
	}
	return &Alias{obj, name}
}

func (nc *Alias) TypeName() string {
	return nc.Name
}

func (nc *Alias) Primitive() PrimitiveType {
	return ALIAS
}

func (nc *Alias) Items() OBJType {
	return nc.Holds.Items()
}

func (nc *Alias) FixedItems() []OBJType {
	return nc.Holds.FixedItems()
}

func (nc *Alias) NamedItems() map[string]int {
	return nc.Holds.NamedItems()
}

func (nc *Alias) Create() ([]runtime.Instruction, error) {
	return nc.Holds.Create()
}

type Structure struct {
	ItemTypes []OBJType
	Fields    map[string]int
	Name      string
}

func StructOf(items []OBJType, fields map[string]int, name string) OBJType {
	return &Structure{items, fields, name}
}

func (nc *Structure) TypeName() string {
	return nc.Name
}

func (nc *Structure) Primitive() PrimitiveType {
	return STRUCT
}

func (nc *Structure) Items() OBJType {
	return nil
}

func (nc *Structure) FixedItems() []OBJType {
	return nc.ItemTypes
}

func (nc *Structure) NamedItems() map[string]int {
	return nc.Fields
}

func (nc *Structure) Create() ([]runtime.Instruction, error) {
	data := make([]runtime.Instruction, 0)
	for _, v := range nc.ItemTypes {
		i, e := v.Create()
		if e != nil {
			return nil, e
		}
		data = append(i, data...)
	}
	return append(data, runtime.MKInstruction(runtime.CSE, len(nc.ItemTypes))), nil
}

type FunctionType struct {
	args []OBJType
	ret  OBJType
}

func FunctionTypeOf(args []OBJType, ret OBJType) OBJType {
	return &FunctionType{args, ret}
}

func (nc *FunctionType) TypeName() string {
	return "Function"
}

func (nc *FunctionType) Primitive() PrimitiveType {
	return FUNCTION
}

func (nc *FunctionType) Items() OBJType {
	return nil
}

func (nc *FunctionType) FixedItems() []OBJType {
	return nil
}

func (nc *FunctionType) NamedItems() map[string]int {
	return nil
}

func (nc *FunctionType) Create() ([]runtime.Instruction, error) {
	return nil, errors.New("Trying to instance a type with no default value")
}
