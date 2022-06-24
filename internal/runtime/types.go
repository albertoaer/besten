package runtime

type OBJType uint8

const (
	VOID    OBJType = 0
	INTEGER         = 1
	DECIMAL         = 2
	STRING          = 3
	VECTOR          = 4
	MAP             = 5
	ALIAS           = 6
)

type Object interface{}

/* type (
	Integer int
	Decimal float64
	String  string
	Vector  []Object
	Map     map[Object]Object
	Alias   string
)
*/
