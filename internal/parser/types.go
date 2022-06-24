package parser

type PrimitiveType uint8

const (
	VOID    PrimitiveType = 0
	ANY     PrimitiveType = 1
	NULL    PrimitiveType = 2
	INTEGER PrimitiveType = 3
	DECIMAL PrimitiveType = 4
	BOOL    PrimitiveType = 5
	STRING  PrimitiveType = 6
	VECTOR  PrimitiveType = 7
	MAP     PrimitiveType = 8
	TUPLE   PrimitiveType = 9
	STRUCT  PrimitiveType = 10
)

func ArrRepr(arr []OBJType) string {
	rep := "("
	for i, a := range arr {
		rep += Repr(a)
		if i < len(arr)-1 {
			rep += ","
		}
	}
	rep += ")"
	return rep
}

func Repr(a OBJType) string {
	//TODO: Modify representation, for example, tuple: {type1, type2}
	base := a.TypeName()
	switch a.Primitive() {
	case VECTOR, MAP:
		base += "|" + Repr(a.Items())
	case STRUCT:
		var items []OBJType
		for _, i := range items {
			items = append(items, i)
		}
		base += "|" + ArrRepr(items)
	case TUPLE:
		base += "|" + ArrRepr(a.FixedItems())
	}
	return base
}

type OBJType interface {
	TypeName() string               //For identifying type
	Primitive() PrimitiveType       //For basic object kind identification
	Items() OBJType                 //For containers with unique and unnamed types
	FixedItems() []OBJType          //For tuple with fixed fields
	NamedItems() map[string]OBJType //For structures with fixed fields
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
	case VECTOR, MAP:
		return CompareTypes(a.Items(), b.Items())
	case STRUCT:
		return CompareMapyOfTypes(a.NamedItems(), b.NamedItems())
	case TUPLE:
		return CompareArrayOfTypes(a.FixedItems(), b.FixedItems())
	}
	return true
}

type Literal struct {
	RawType PrimitiveType
	Name    string
}

var (
	Void OBJType = &Literal{VOID, "Void"}
	Int  OBJType = &Literal{INTEGER, "Int"}
	Dec  OBJType = &Literal{DECIMAL, "Decimal"}
	Bool OBJType = &Literal{BOOL, "Boolean"}
	Str  OBJType = &Literal{STRING, "String"}
	Any  OBJType = &Literal{ANY, "Any"}
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

type Container struct {
	ContainerType PrimitiveType
	ItemsType     OBJType
	Name          string
}

func VecOf(t OBJType) OBJType {
	return &Container{VECTOR, t, "Vector"}
}

func MapOf(t OBJType) OBJType {
	return &Container{MAP, t, "Map"}
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

type Structure struct {
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
}
