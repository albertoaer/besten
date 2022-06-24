package runtime

type ICode uint16

const (
	NOP ICode = 0 //No operation

	//ARITHMETIC

	ADD  = 10
	SUB  = 11
	MUL  = 12
	DIV  = 13
	MOD  = 14
	ADDF = 15
	SUBF = 16
	MULF = 17
	DIVF = 18

	//CONVERSION

	ITD = 20 //int to double
	DTI = 21 //double to int
	IRE = 22 //int representation (to string)
	DRE = 23 //double representation (to string)
	IPA = 24 //int parse (to int)
	DPA = 25 //double parse (to double)

	//COMPARISON

	EQI = 30 //equal ints
	EQD = 31 //equal doubles
	EQS = 32 //equal strings
	ILE = 33 //int less
	DLE = 34 //double less
	IGR = 35 //int greater
	DGR = 36 //double greater
	ILQ = 37 //int less equals
	DLQ = 38 //double less equals
	IGQ = 39 //int greater equals
	DGQ = 40 //double greater equals

	//LOGIC

	NOT = 50 //logic not
	AND = 51 //logic and
	OR  = 52 //logic or
	XOR = 53 //logic xor

	//STRINGS

	CCS = 60 //Concatenate strings
	CAI = 61 //Char at index

	//MEMORY

	LEI = 70 //Load environment item
	SEI = 71 //Save environment item
	LLI = 72 //Load local item
	SLI = 73 //Save local item

	//STACK

	PSH = 90 //Push top
	POP = 91 //Pop top
	CLR = 92 //Clear stack
	DUP = 93 //Duplicate last
	SWT = 94 //Switch two last

	//CONTROl

	CLL = 120 //Jumps into symbol address and saves actual address
	JMP = 121 //Jumps into symbol address
	RET = 122 //Return to last address in the stack
	MVR = 125 //Moves pc relative to position
	MVT = 126 //Moves pc relative to position if true
	MVF = 127 //Moves pc relative to position if false

	//MAPS AND VECTORS

	KVC = 160 //Creates a map
	PRP = 161 //Gets value from map
	ATT = 162 //Attaches value to map
	EXK = 163 //Finds out if key is already in map
	VEC = 164 //Creates a vector
	ACC = 165 //Accesses position of vector
	APP = 166 //Appends element at the end of a vector
	SVI = 167 //Set vector item at position
	DMI = 168 //Delete map item
	PFV = 169 //Pop from vector
	CSE = 171 //Collapse stack elements

	//SIZE

	SOS = 180 //Size of string
	SOV = 181 //Size of vector
	SOM = 182 //Size of map

	//STATE

	SWR = 200 //State write
	SRE = 201 //State read

	//Threads

	FRK = 240 //Forks, next line will run in a different thread (stack is copied) and adds pid to father thread
	ELL = 241 //Ends life line, stops current thread
	FPI = 242 //Pushes to stack the father pid
	MSG = 243 //Sends message to pid
	AWA = 244 //Blocks threads until message (pushed into the stack)

	//Interaction

	INV = 250 //Invokes native function
	SYS = 251 //Invokes a system call
	IFD = 252 //Invocation for debugging, run directly methods, NOT SAFE!

	LDOP = 256 //Last defined operation, just a mark
)

var opNumTable [LDOP]uint8

func init() {

	opNumTable[NOP] = 0

	//ARITHMETIC
	opNumTable[ADD] = 2
	opNumTable[SUB] = 2
	opNumTable[MUL] = 2
	opNumTable[DIV] = 2
	opNumTable[MOD] = 2
	opNumTable[ADDF] = 2
	opNumTable[SUBF] = 2
	opNumTable[MULF] = 2
	opNumTable[DIVF] = 2

	//CONVERSION
	opNumTable[ITD] = 1
	opNumTable[DTI] = 1
	opNumTable[IRE] = 1
	opNumTable[DRE] = 1
	opNumTable[IPA] = 1
	opNumTable[DPA] = 1

	//COMPARISON
	opNumTable[EQI] = 2
	opNumTable[EQD] = 2
	opNumTable[EQS] = 2
	opNumTable[ILE] = 2
	opNumTable[DLE] = 2
	opNumTable[IGR] = 2
	opNumTable[DGR] = 2
	opNumTable[ILQ] = 2
	opNumTable[DLQ] = 2
	opNumTable[IGQ] = 2
	opNumTable[DGQ] = 2

	//LOGIC
	opNumTable[NOT] = 1
	opNumTable[AND] = 2
	opNumTable[OR] = 2
	opNumTable[XOR] = 2

	//STRINGS
	opNumTable[CCS] = 2
	opNumTable[CAI] = 2

	//MEMORY
	opNumTable[LEI] = 1
	opNumTable[SEI] = 2
	opNumTable[LLI] = 1
	opNumTable[SLI] = 2

	//STACK
	opNumTable[PSH] = 1
	opNumTable[POP] = 1
	opNumTable[CLR] = 0
	opNumTable[DUP] = 1
	opNumTable[SWT] = 2

	//CONTROL
	opNumTable[CLL] = 1
	opNumTable[JMP] = 1
	opNumTable[RET] = 0
	opNumTable[MVR] = 1
	opNumTable[MVT] = 2
	opNumTable[MVF] = 2

	//MAPS AND VECTORS
	opNumTable[KVC] = 0
	opNumTable[PRP] = 2
	opNumTable[ATT] = 3
	opNumTable[EXK] = 2
	opNumTable[VEC] = 0
	opNumTable[ACC] = 2
	opNumTable[APP] = 2
	opNumTable[SVI] = 3
	opNumTable[DMI] = 2
	opNumTable[PFV] = 1
	opNumTable[CSE] = 1

	//SIZE
	opNumTable[SOS] = 1
	opNumTable[SOV] = 1
	opNumTable[SOM] = 1

	//STATE
	opNumTable[SWR] = 1
	opNumTable[SRE] = 0

	//INTERACTION
	opNumTable[INV] = 1
	opNumTable[IFD] = 1
}
