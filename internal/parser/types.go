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
	STRUCT  PrimitiveType = 9
	TUPLE   PrimitiveType = 10
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
	base := a.TypeName()
	switch a.Primitive() {
	case VECTOR, MAP:
		base += "|" + Repr(a.Items())
	case STRUCT: //TODO: Named types repr
	case TUPLE: //TODO: Fixed length types repr
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

type StructureContainer struct {
	ContainerType PrimitiveType
	ItemsType     map[string]OBJType
	Name          string
}

func StructOf(fields map[string]OBJType, name string) OBJType {
	return &StructureContainer{STRUCT, fields, name}
}

func (nc *StructureContainer) TypeName() string {
	return nc.Name
}

func (nc *StructureContainer) Primitive() PrimitiveType {
	return nc.ContainerType
}

func (nc *StructureContainer) Items() OBJType {
	return nil
}

func (nc *StructureContainer) FixedItems() []OBJType {
	return nil
}

func (nc *StructureContainer) NamedItems() map[string]OBJType {
	return nc.ItemsType
}
