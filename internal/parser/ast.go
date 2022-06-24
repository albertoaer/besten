package parser

import (
	"errors"
	"fmt"
	"strconv"

	. "github.com/Besten/internal/lexer"
	. "github.com/Besten/internal/runtime"
)

var genericPairs = []struct {
	open  Token
	close Token
}{{CBOPEN, CBCLOSE}, {POPEN, PCLOSE}, {BOPEN, BCLOSE}}

type syntaxLiteral struct {
	value string
	kind  TokenType
	owner *SyntaxTree
}

func (s *syntaxLiteral) runIntoStack(p *Parser, stack *[]Instruction) (OBJType, error) {
	var ins Instruction
	var toret OBJType = Void
	switch s.kind {
	case StringToken:
		toret = Str
		ins = MKInstruction(PSH, s.value)
	case IntegerToken:
		toret = Int
		i, e := strconv.Atoi(s.value)
		if e != nil {
			return Void, e
		}
		ins = MKInstruction(PSH, i)
	case DecimalToken:
		toret = Dec
		f, e := strconv.ParseFloat(s.value, 64)
		if e != nil {
			return Void, e
		}
		ins = MKInstruction(PSH, f)
	case KeywordToken:
		toret = Bool
		if s.value == TRUE.Data {
			ins = MKInstruction(PSH, -1)
			break
		} else if s.value == FALSE.Data {
			ins = MKInstruction(PSH, 0)
			break
		}
		fallthrough
	default:
		return nil, errors.New("Wrong literal")
	}
	*stack = append(*stack, ins)
	return toret, nil
}

func isLiteral(tk Token) bool {
	kind := tk.Kind
	return kind == StringToken || kind == IntegerToken || kind == DecimalToken || tk == TRUE || tk == FALSE
}

func getRoute(tk []Token) ([]string, error) {
	if len(tk) == 0 {
		return nil, errors.New("Expecting identifier")
	}
	res := make([]string, 0)
	for i := 0; i < len(tk); i += 2 {
		if tk[i].Kind != IdToken {
			return nil, errors.New(fmt.Sprintf("Unexpected token: %s", tk[i].Data))
		}
		if i+1 < len(tk) && tk[i+1] != DOT {
			return nil, errors.New(fmt.Sprintf("Unexpected token: %s", tk[i+1].Data))
		}
		res = append(res, tk[i].Data)
	}
	return res, nil
}

type syntaxTupleDefinition struct {
	elements []syntaxBranch
	owner    *SyntaxTree
}

func (s *syntaxTupleDefinition) runIntoStack(p *Parser, stack *[]Instruction) (OBJType, error) {
	tps := make([]OBJType, 0)
	for i := range s.elements {
		t, e := s.elements[len(s.elements)-i-1].runIntoStack(p, stack)
		if e != nil {
			return nil, e
		}
		tps = append(tps, t)
	}
	*stack = append(*stack, MKInstruction(CSE, len(s.elements)), MKInstruction(WTP))
	return TupleOf(tps), nil
}

type syntaxConstantAccess struct {
	target syntaxBranch
	idx    int
	owner  *SyntaxTree
}

func (s *syntaxConstantAccess) runIntoStack(p *Parser, stack *[]Instruction) (OBJType, error) {
	tempstack := make([]Instruction, 0) //Important to insert the push index
	r, e := s.target.runIntoStack(p, &tempstack)
	if e != nil {
		return nil, e
	}
	if r.Primitive() != TUPLE { //When is not tuple related search for index operator
		ins, ret, err := p.solveFunctionCall(INDEXOP.Data, true, []OBJType{r, Int})
		*stack = append(*stack, MKInstruction(PSH, s.idx))
		*stack = append(*stack, tempstack...)
		*stack = append(*stack, ins...)
		return ret, err
	}
	*stack = append(*stack, tempstack...)
	elems := r.FixedItems()
	if s.idx < 0 || len(elems) <= s.idx {
		return nil, errors.New(fmt.Sprintf("Invalid index %d for %s", s.idx, Repr(r)))
	}
	*stack = append(*stack, MKInstruction(PTW), MKInstruction(ACC, s.idx))
	return elems[s.idx], nil
}

type syntaxRoute struct {
	origin syntaxBranch //if origin is not null is where the route access to
	route  []string     //Maybe include de single variable case, already implemented in syntaxLiteral
	owner  *SyntaxTree
}

func (s *syntaxRoute) runIntoStack(p *Parser, stack *[]Instruction) (OBJType, error) {
	if len(s.route) == 0 {
		return nil, errors.New("No route provided")
	}
	var tp OBJType
	var err error
	route := s.route
	if s.origin != nil {
		tp, err = s.origin.runIntoStack(p, stack)
		if err != nil {
			return Void, err
		}
	} else {
		var ins Instruction
		ins, tp, err = p.currentScope().GetVariableIns(route[0])
		if err != nil {
			return nil, err
		}
		*stack = append(*stack, ins)
		route = s.route[1:]
	}
	for _, r := range route {
		if t, e := tp.NamedItems()[r]; e {
			tp = t
			*stack = append(*stack, MKInstruction(PRP, r))
		} else {
			return Void, errors.New(fmt.Sprintf("Type %s does not have property %s", tp.TypeName(), r))
		}
	}
	return tp, err
}

func (s *syntaxRoute) runIntoStackSet(p *Parser, stack *[]Instruction, objtype OBJType) error {
	if len(s.route) == 0 {
		return errors.New("No route provided")
	}
	var tp OBJType
	var err error
	route := s.route
	if s.origin != nil {
		tp, err = s.origin.runIntoStack(p, stack)
		if err != nil {
			return err
		}
	} else if len(route) == 1 {
		var ins Instruction
		ins, err = p.currentScope().SetVariableIns(route[0], objtype)
		if err != nil {
			return err
		}
		*stack = append(*stack, ins)
	} else {
		var ins Instruction
		ins, tp, err = p.currentScope().GetVariableIns(route[0])
		if err != nil {
			return err
		}
		*stack = append(*stack, ins)
		route = s.route[1:]
		for i, r := range route {
			if t, e := tp.NamedItems()[r]; e {
				if i == len(route)-1 {
					*stack = append(*stack, MKInstruction(ATT, r))
				} else {
					tp = t
					*stack = append(*stack, MKInstruction(PRP, r))
				}
			} else {
				return errors.New(fmt.Sprintf("Type %s does not have property %s", tp.TypeName(), r))
			}
		}
	}
	return err
}

type syntaxCall struct {
	relation *syntaxRoute
	operands []syntaxBranch
	owner    *SyntaxTree
}

func (s *syntaxCall) runIntoStack(p *Parser, stack *[]Instruction) (OBJType, error) {
	if s.relation == nil || len(s.relation.route) == 0 {
		return Void, errors.New("No way to fetch function")
	}
	if s.relation.origin != nil {
		return Void, errors.New("A function must always be global defined")
	}
	if len(s.relation.route) > 1 {
		return Void, errors.New("Not implemented function route navigation")
	}
	ops := make([]OBJType, len(s.operands))
	for i := len(s.operands) - 1; i >= 0; i-- {
		var e error
		if ops[i], e = s.operands[i].runIntoStack(p, stack); e != nil {
			return nil, e
		}
		if ops[i].Primitive() == VOID {
			return nil, errors.New("Using Void as function argument")
		}
	}
	ins, ret, err := p.solveFunctionCall(s.relation.route[0], false, ops)
	*stack = append(*stack, ins...)
	return ret, err
}

type syntaxOpCall struct {
	operator string
	operands []syntaxBranch
	owner    *SyntaxTree
}

func (s *syntaxOpCall) runIntoStack(p *Parser, stack *[]Instruction) (OBJType, error) {
	ops := make([]OBJType, len(s.operands))
	for i := len(s.operands) - 1; i >= 0; i-- {
		var e error
		if ops[i], e = s.operands[i].runIntoStack(p, stack); e != nil {
			return nil, e
		}
		if ops[i].Primitive() == VOID {
			return nil, errors.New("Using Void as operator argument")
		}
	}
	ins, ret, err := p.solveFunctionCall(s.operator, true, ops)
	*stack = append(*stack, ins...)
	return ret, err
}

type syntaxTypeCreation struct {
	typeref []Token
	owner   *SyntaxTree
}

func (s *syntaxTypeCreation) runIntoStack(p *Parser, stack *[]Instruction) (OBJType, error) {
	obj, e := solveTypeFromTokens(s.typeref)
	if e != nil {
		return nil, e
	}
	//TODO: Create the type
	return obj, nil
}

type syntaxHighLevelCall struct {
	relation  *syntaxRoute
	operands  []syntaxBranch
	modifiers []syntaxBranch
	lambda    string
	owner     *SyntaxTree
}

func (s *syntaxHighLevelCall) runIntoStack(p *Parser, stack *[]Instruction) (OBJType, error) {
	if s.relation == nil || len(s.relation.route) == 0 {
		return Void, errors.New("No way to fetch function")
	}
	if s.relation.origin != nil {
		return Void, errors.New("A function must always be global defined")
	}
	if len(s.relation.route) > 1 {
		return Void, errors.New("Not implemented function route navigation")
	}
	ops := make([]OBJType, len(s.operands))
	for i := len(s.operands) - 1; i >= 0; i-- {
		var e error
		if ops[i], e = s.operands[i].runIntoStack(p, stack); e != nil {
			return nil, e
		}
		if ops[i].Primitive() == VOID {
			return nil, errors.New("Using Void as function argument")
		}
	}
	ins, ret, err := p.solveFunctionCall(s.relation.route[0], false, ops)
	if s.owner.inReturn && ins[len(ins)-1].Code == CLL {
		ins[len(ins)-1].Code = JMP
	}
	*stack = append(*stack, ins...)
	return ret, err
}

type syntaxAssignment struct {
	value  syntaxBranch
	assign syntaxBranchGetSet
	owner  *SyntaxTree
}

func (s *syntaxAssignment) runIntoStack(p *Parser, stack *[]Instruction) (OBJType, error) {
	objt, err := s.value.runIntoStack(p, stack)
	if err != nil {
		return nil, err
	}
	return Void, s.assign.runIntoStackSet(p, stack, objt)
}

type syntaxBranch interface {
	runIntoStack(p *Parser, stack *[]Instruction) (OBJType, error)
}

type syntaxBranchGetSet interface {
	runIntoStack(p *Parser, stack *[]Instruction) (OBJType, error)
	runIntoStackSet(p *Parser, stack *[]Instruction, objtype OBJType) error
}

type SyntaxTree struct {
	root     syntaxBranch
	parser   *Parser
	inReturn bool
}

func (s *SyntaxTree) runIntoStack(stack *[]Instruction) (OBJType, error) {
	if s.root == nil {
		return Void, errors.New("No expression node to parse")
	}
	return s.root.runIntoStack(s.parser, stack)
}

func GenerateTree(parser *Parser, tks []Token, children []Block, returning bool) (tree *SyntaxTree, err error) {
	tree = &SyntaxTree{nil, parser, returning}
	if returning {
		tree.root, err = tree.generateFirstLevelExpression(tks, children)
	} else {
		tree.root, err = tree.splitAssignmentExpression(tks, children)
	}
	return
}

func (s *SyntaxTree) splitAssignmentExpression(tks []Token, children []Block) (syntaxBranch, error) {
	if len(tks) == 0 {
		return nil, errors.New("No expression to parse")
	}
	ttks, err := splitByToken(tks, func(tk Token) bool { return tk == ASSIGN }, genericPairs, false, false, false)
	if err != nil {
		return nil, err
	}
	if len(ttks) > 2 {
		return nil, errors.New("Multiassignment is not implemented")
	}
	rightterm, err := s.generateFirstLevelExpression(ttks[len(ttks)-1], children)
	if err != nil {
		return nil, err
	}
	if len(ttks) == 1 {
		return rightterm, nil
	}
	return s.identifyAssignmentExpressionBranch(ttks[0], rightterm)
}

func (s *SyntaxTree) generateFirstLevelExpression(tks []Token, children []Block) (syntaxBranch, error) {
	if len(tks) > 0 {
		if tks[0] == FOR || tks[0] == WHILE {
			return nil, errors.New("Generators are not implemented yet")
		}
		haslambda := tks[len(tks)-1] == DO
		name, args, mods, lambdatemplate, err := splitFirstLevelFunctionCall(tks, haslambda, children)
		if err != nil {
			return nil, err
		}
		if len(name) != 0 {
			var op syntaxHighLevelCall
			op.owner = s
			route, e := getRoute(name)
			if e != nil {
				return nil, e
			}
			op.relation = &syntaxRoute{origin: nil, route: route}
			op.operands, err = s.generateOperands(args)
			if err != nil {
				return nil, err
			}
			op.modifiers, err = generateModifiers(mods)
			if err != nil {
				return nil, err
			}
			if haslambda {
				op.lambda = s.parser.solveLambdaTemplate(lambdatemplate)
			}
			return &op, nil
		} else if tks[len(tks)-1] == DO {
			return nil, errors.New("Unexpected lambda")
		} else {
			//No high level function
			return s.generateSecondLevelExpression(tks)
		}
	}
	return nil, nil
}

func (s *SyntaxTree) generateSecondLevelExpression(tks []Token) (syntaxBranch, error) {
	ttks, err := splitByToken(tks, func(tk Token) bool { return tk.Kind == OperatorToken }, genericPairs, true, true, false)
	if err != nil {
		return nil, err
	}
	return s.generateOperatorBranch(ttks)
}

//we asume the odd positions are operators
func (s *SyntaxTree) generateOperatorBranch(tks [][]Token) (syntaxBranch, error) {
	if len(tks) == 0 {
		return nil, errors.New("Expecting expression")
	}
	if len(tks) == 1 {
		return s.identifyExpressionBranch(tks[0])
	}
	if len(tks[0]) == 0 { //Unary operator
		op, err := s.generateOperatorBranch(tks[2:])
		return &syntaxOpCall{
			operator: tks[1][0].Data,
			operands: []syntaxBranch{op},
		}, err
	}
	firstop, err := s.identifyExpressionBranch(tks[0])
	if err != nil {
		return nil, err
	}
	opb := syntaxOpCall{
		operator: tks[1][0].Data,
		operands: []syntaxBranch{firstop},
	}
	if len(tks) > 2 {
		op, err := s.generateOperatorBranch(tks[2:])
		if err != nil {
			return nil, err
		}
		opb.operands = append(opb.operands, op)
	}
	return &opb, nil
}

func (s *SyntaxTree) identifyAssignmentExpressionBranch(tks []Token, obj syntaxBranch) (syntaxBranch, error) {
	route, e := getRoute(tks)
	if e != nil {
		return nil, e
	}
	return &syntaxAssignment{obj, &syntaxRoute{nil, route, s}, s}, nil
}

//Detects if it's an object literal, function call, etc and generates it
func (s *SyntaxTree) identifyExpressionBranch(tks []Token) (syntaxBranch, error) {
	if len(tks) == 0 {
		return nil, errors.New("Expecting expression")
	}
	if route, e := getRoute(tks); e == nil {
		return &syntaxRoute{nil, route, s}, nil
	}
	if isLiteral(tks[0]) {
		if len(tks) > 1 {
			return nil, errors.New(fmt.Sprintf("Unexpected token: %s", tks[1].Data))
		}
		return &syntaxLiteral{tks[0].Data, tks[0].Kind, s}, nil
	}
	if tks[0] == CBOPEN {
		_, inner, right, err := blockSubtract(tks, CBOPEN, CBCLOSE, genericPairs)
		if err != nil {
			return nil, err
		}
		args, err := splitByToken(inner, func(t Token) bool { return t == COMA }, genericPairs, false, false, false)
		if err != nil {
			return nil, err
		}
		tupledef := syntaxTupleDefinition{elements: make([]syntaxBranch, 0), owner: s}
		for _, arg := range args {
			exp, err := s.generateSecondLevelExpression(arg)
			if err != nil {
				return nil, err
			}
			tupledef.elements = append(tupledef.elements, exp)
		}
		if len(right) > 0 {
			return s.identifySubrouting(&tupledef, right)
		}
		return &tupledef, nil
	}
	left, inner, right, err := blockSubtract(tks, BOPEN, BCLOSE, genericPairs)
	if err != nil {
		return nil, err
	}
	if len(left) == len(tks) { //No bracket block, forced to be parenthesis related
		return s.identifyParenthesisBranch(tks)
	} else {
		if len(inner) == 0 { //Index by nothing or creating nothing
			return nil, errors.New("Expecting something between []")
		}
		var preceded syntaxBranch
		var e error
		if len(left) == 0 { //Create type
			preceded = &syntaxTypeCreation{left, s}
		} else { //Indexation
			route, e := getRoute(left)
			if e != nil {
				return nil, e
			}
			if len(inner) == 1 && inner[0].Kind == IntegerToken {
				i, r := strconv.Atoi(inner[0].Data)
				if r != nil {
					return nil, r
				}
				return &syntaxConstantAccess{target: &syntaxRoute{nil, route, s}, idx: i, owner: s}, nil
			}
			indexer, err := s.generateSecondLevelExpression(inner)
			e = err
			preceded = &syntaxOpCall{operator: INDEXOP.Data, operands: []syntaxBranch{&syntaxRoute{nil, route, s}, indexer}}
		}
		if len(right) == 0 || e != nil {
			return preceded, e
		}
		return s.identifySubrouting(preceded, right)
	}
}

func (s *SyntaxTree) identifySubrouting(preceded syntaxBranch, nexttks []Token) (syntaxBranch, error) {
	left, inner, right, err := blockSubtract(nexttks, BOPEN, BCLOSE, genericPairs)
	if err != nil {
		return nil, err
	}
	var idx syntaxBranch = nil
	var e error
	if len(left) == len(nexttks) {
		route, e := getRoute(left)
		if e != nil {
			return nil, e
		}
		return &syntaxRoute{preceded, route, s}, nil
	} else {
		branch := preceded
		if len(left) != 0 {
			route, e := getRoute(left)
			if e != nil {
				return nil, e
			}
			branch = &syntaxRoute{preceded, route, s}
		}
		if len(inner) == 1 && inner[0].Kind == IntegerToken {
			i, r := strconv.Atoi(inner[0].Data)
			if r != nil {
				return nil, r
			}
			return &syntaxConstantAccess{target: branch, idx: i, owner: s}, nil
		}
		indexer, err := s.generateSecondLevelExpression(inner)
		e = err
		idx = &syntaxOpCall{operator: INDEXOP.Data, operands: []syntaxBranch{branch, indexer}, owner: s}
	}
	if len(right) == 0 || e != nil {
		return idx, e
	}
	return s.identifySubrouting(idx, right)
}

func (s *SyntaxTree) identifyParenthesisBranch(tks []Token) (syntaxBranch, error) {
	left, inner, right, err := blockSubtract(tks, POPEN, PCLOSE, genericPairs)
	if err != nil {
		return nil, err
	}
	var branch syntaxBranch
	if len(left) == 0 {
		branch, err = s.generateSecondLevelExpression(inner)
	} else {
		branch, err = s.generateFunctionCall(left, inner)
	}
	if len(right) == 0 || err != nil {
		return branch, err
	}
	return s.identifySubrouting(branch, right)
}

func (s *SyntaxTree) generateFunctionCall(head []Token, callbody []Token) (syntaxBranch, error) {
	args, err := splitByToken(callbody, func(t Token) bool { return t == COMA }, genericPairs, false, false, false)
	if err != nil {
		return nil, err
	}
	route, e := getRoute(head)
	if e != nil {
		return nil, e
	}
	call := syntaxCall{relation: &syntaxRoute{origin: nil, route: route}, operands: make([]syntaxBranch, 0)}
	for _, arg := range args {
		exp, err := s.generateSecondLevelExpression(arg)
		if err != nil {
			return nil, err
		}
		call.operands = append(call.operands, exp)
	}
	return &call, nil
}

func (s *SyntaxTree) generateOperands(tks []Token) ([]syntaxBranch, error) {
	ttks, err := splitByToken(tks, func(t Token) bool { return t == COMA }, genericPairs, false, false, false)
	if err != nil {
		return nil, err
	}
	branchs := make([]syntaxBranch, 0)
	for _, opl := range ttks {
		b, e := s.generateSecondLevelExpression(opl)
		if e != nil {
			return nil, e
		}
		branchs = append(branchs, b)
	}
	return branchs, nil
}

func generateModifiers(ttks [][]Token) ([]syntaxBranch, error) {
	for range ttks {
		return nil, errors.New("Modifier not implemented")
	}
	return make([]syntaxBranch, 0), nil
}

func splitFirstLevelFunctionCall(tks []Token, haslambda bool, children []Block) (name []Token, args []Token, modifiers [][]Token, template FunctionTemplate, err error) {
	ttks, err := splitByToken(tks, func(t Token) bool { return t == DOUBLES }, genericPairs, true, false, true)
	switch len(ttks) {
	case 0:
		err = errors.New("No function to be parsed") //Weird
	case 1:
		return //Is not a first level function
	case 2: //without lambda
		name = ttks[0]
		args, modifiers, err = substractModifiers(ttks[1])
	case 3: //with lambda
		if !haslambda {
			err = errors.New("Too much argument lists")
		} else {
			name = ttks[0]
			args, modifiers, err = substractModifiers(ttks[1])
			//FIXME: solution for types
			template.Args, _, _, template.Varargs, err = parseArguments(ttks[2])
			template.Children = children
			if err == nil {
				return
			}
		}
	default:
		err = errors.New("Too much argument lists")
	}
	return
}

func substractModifiers(tks []Token) ([]Token, [][]Token, error) {
	ttks, err := splitByToken(tks, func(t Token) bool { return t.Kind == KeywordToken && t != TRUE && t != FALSE }, genericPairs, true, true, false)
	if err != nil {
		return nil, nil, err
	}
	if len(ttks) == 0 {
		return make([]Token, 0), make([][]Token, 0), nil
	}
	return ttks[0], ttks[1:], nil
}
