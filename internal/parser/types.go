package parser

import "github.com/Besten/internal/runtime"

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
	}
	return base
}

type OBJType interface {
	TypeName() string               //For identifying type
	Primitive() PrimitiveType       //For basic object kind identification
	Items() OBJType                 //For containers with unique and unnamed types
	FixedItems() []OBJType          //For tuple with fixed fields
	NamedItems() map[string]OBJType //For structures with fixed fields
	Create() []runtime.Instruction
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

func CompareMapyOfTypes(a, b map[string]OBJType) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v1 := range a {
		if v2, e := b[k]; !e || !CompareTypes(v1, v2) {
			return false
		}
	}
	return true
}

func CompareTypes(a, b OBJType) bool {
	if a.Primitive() == ANY || b.Primitive() == ANY {
		return true
	}
	if a.Primitive() != b.Primitive() {
		return false
	}
	switch a.Primitive() {
	case VECTOR, MAP, VARIADIC:
		return CompareTypes(a.Items(), b.Items())
	case STRUCT:
		return CompareMapyOfTypes(a.NamedItems(), b.NamedItems())
	case TUPLE:
		return CompareArrayOfTypes(a.FixedItems(), b.FixedItems())
	}
	return true
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

func (nc *Literal) NamedItems() map[string]OBJType {
	return nil
}

func (nc *Literal) Create() []runtime.Instruction {
	return runtime.MKInstruction(runtime.PSH, nc.DefaultObj).Fragment()
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

func (nc *Container) NamedItems() map[string]OBJType {
	return nil
}

func (nc *Container) Create() []runtime.Instruction {
	return runtime.MKInstruction(nc.CreateInstruction).Fragment()
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

func (nc *Tuple) NamedItems() map[string]OBJType {
	return nil
}

func (nc *Tuple) Create() []runtime.Instruction {
	data := make([]runtime.Instruction, 0)
	for _, v := range nc.ItemTypes {
		data = append(data, v.Create()...)
	}
	return append(data, runtime.MKInstruction(runtime.CSE, len(nc.ItemTypes)))
}

/*type Structure struct {
	ItemsType map[string]OBJType
	Name      string
}

func StructOf(fields map[string]OBJType, name string) OBJType {
	return &Structure{fields, name}
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
	return nil
}

func (nc *Structure) NamedItems() map[string]OBJType {
	return nc.ItemsType
}*/
