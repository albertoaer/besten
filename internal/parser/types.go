package parser

type PrimitiveType uint8

const (
	VOID    PrimitiveType = 0
	ANY     PrimitiveType = 1
	NULL    PrimitiveType = 2
	INTEGER PrimitiveType = 3
	DECIMAL PrimitiveType = 4
	STRING  PrimitiveType = 5
	VECTOR  PrimitiveType = 6
	MAP     PrimitiveType = 7
	STRUCT  PrimitiveType = 8
)

type OBJType interface {
	Primitive() PrimitiveType //For basic object kind identification
	Items() OBJType           //For containers with unique and unnamed types
	FixedItems() []OBJType    //For structures with fixed fields
}

type Literal struct {
	RawType PrimitiveType
}

var (
	Void *Literal = &Literal{VOID}
	Int  *Literal = &Literal{INTEGER}
	Dec  *Literal = &Literal{DECIMAL}
	Str  *Literal = &Literal{STRING}
)

func (nc *Literal) Primitive() PrimitiveType {
	return nc.RawType
}

func (nc *Literal) Items() OBJType {
	return nil
}

func (nc *Literal) FixedItems() []OBJType {
	return nil
}

type Container struct {
	ContainerType PrimitiveType
	ItemsType     OBJType
}

func VecOf(t OBJType) *Container {
	return &Container{VECTOR, t}
}

func MapOf(t OBJType) *Container {
	return &Container{MAP, t}
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

type FixedContainer struct {
	ContainerType PrimitiveType
	ItemsType     []OBJType
}

func StructOf(fields []OBJType) *FixedContainer {
	return &FixedContainer{STRUCT, fields}
}

func (nc *FixedContainer) Primitive() PrimitiveType {
	return nc.ContainerType
}

func (nc *FixedContainer) Items() OBJType {
	return nil
}

func (nc *FixedContainer) FixedItems() []OBJType {
	return nc.ItemsType
}
