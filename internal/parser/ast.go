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
			ins = MKInstruction(PSH, 1)
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
	tps := make([]OBJType, len(s.elements))
	for i := range s.elements {
		t, e := s.elements[len(s.elements)-i-1].runIntoStack(p, stack)
		if e != nil {
			return nil, e
		}
		tps[len(s.elements)-i-1] = t
	}
	*stack = append(*stack, MKInstruction(CSE, len(s.elements)))
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
		return p.solveFunctionCall(INDEXOP.Data, true, []OBJType{r, Int},
			[][]Instruction{tempstack, MKInstruction(PSH, s.idx).Fragment()}, stack)
	}
	*stack = append(*stack, tempstack...)
	elems := r.FixedItems()
	if s.idx < 0 || len(elems) <= s.idx {
		return nil, errors.New(fmt.Sprintf("Invalid index %d for %s", s.idx, Repr(r)))
	}
	*stack = append(*stack, MKInstruction(ACC, nil, s.idx))
	return elems[s.idx], nil
}

func (s *syntaxConstantAccess) runIntoStackSet(p *Parser, stack *[]Instruction, branch syntaxBranch) error {
	tempstack := make([]Instruction, 0) //Important to insert the push index
	r, e := s.target.runIntoStack(p, &tempstack)
	if e != nil {
		return e
	}
	tempstackval := make([]Instruction, 0) //Important to insert the push index
	r2, e := branch.runIntoStack(p, &tempstackval)
	if e != nil {
		return e
	}
	if r.Primitive() != TUPLE { //When is not tuple related search for index operator
		_, err := p.solveFunctionCall("setbykey", false, []OBJType{r2, Int, r},
			[][]Instruction{tempstackval, MKInstruction(PSH, s.idx).Fragment(), tempstack}, stack)
		return err
	}
	*stack = append(*stack, tempstack...)
	*stack = append(*stack, tempstackval...)
	elems := r.FixedItems()
	if s.idx < 0 || len(elems) <= s.idx {
		return errors.New(fmt.Sprintf("Invalid index %d for %s", s.idx, Repr(r)))
	}
	if !CompareTypes(elems[s.idx], r2) {
		return errors.New(fmt.Sprintf("Invalid type %s at position %d for %s", Repr(r2), s.idx, Repr(r)))
	}
	*stack = append(*stack, MKInstruction(SVI, nil, s.idx))
	return nil
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
		if idx, e := tp.NamedItems()[r]; e {
			tp = tp.FixedItems()[idx]
			*stack = append(*stack, MKInstruction(ACC, nil, idx))
		} else {
			return Void, errors.New(fmt.Sprintf("Type %s does not have property %s", tp.TypeName(), r))
		}
	}
	return tp, err
}

func (s *syntaxRoute) runIntoStackSet(p *Parser, stack *[]Instruction, branch syntaxBranch) error {
	objtype, err := branch.runIntoStack(p, stack)
	if err != nil {
		return err
	}
	if len(s.route) == 0 {
		return errors.New("No route provided")
	}
	var tp OBJType
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
			if idx, e := tp.NamedItems()[r]; e {
				if i == len(route)-1 {
					*stack = append(*stack, MKInstruction(SWT), MKInstruction(SVI, nil, idx))
				} else {
					tp = tp.FixedItems()[idx]
					*stack = append(*stack, MKInstruction(ACC, nil, idx))
				}
			} else {
				return errors.New(fmt.Sprintf("Type %s does not have property %s", tp.TypeName(), r))
			}
		}
	}
	return err
}

type syntaxCast struct {
	origin syntaxBranch
	into   string
}

func (s *syntaxCast) runIntoStack(p *Parser, stack *[]Instruction) (OBJType, error) {
	o, e := s.origin.runIntoStack(p, stack)
	if e != nil {
		return nil, e
	}
	n, e := p.currentScope().FetchType(s.into)
	if e != nil {
		return nil, e
	}
	var err error = nil
	if !checkCompatibility(o, *n) {
		err = errors.New(fmt.Sprintf("%s is not compatible with %s", Repr(o), Repr(*n)))
	}
	return *n, err
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
	ops, stacks, err := runBranchesIntoStacks(p, s.operands)
	if err != nil {
		return nil, err
	}
	return p.solveFunctionCall(s.relation.route[0], false, ops, stacks, stack)
}

type syntaxOpCall struct {
	operator string
	operands []syntaxBranch
	owner    *SyntaxTree
}

func (s *syntaxOpCall) runIntoStack(p *Parser, stack *[]Instruction) (OBJType, error) {
	if v, e := s.operands[0].(*syntaxLiteral); (s.operator == "+" || s.operator == "-") && len(s.operands) == 1 && e &&
		(v.kind == IntegerToken || v.kind == DecimalToken) {
		v.value = s.operator + v.value
		return v.runIntoStack(p, stack)
	}
	ops := make([]OBJType, len(s.operands))
	ops, stacks, err := runBranchesIntoStacks(p, s.operands)
	if err != nil {
		return nil, err
	}
	return p.solveFunctionCall(s.operator, true, ops, stacks, stack)
}

func (s *syntaxOpCall) runIntoStackSet(p *Parser, stack *[]Instruction, branch syntaxBranch) error {
	if s.operator != INDEXOP.Data {
		return errors.New("Cannot set")
	}
	ops, stacks, err := runBranchesIntoStacks(p, []syntaxBranch{branch, s.operands[1], s.operands[0]})
	if err != nil {
		return err
	}
	_, err = p.solveFunctionCall("setbykey", false, ops, stacks, stack)
	return err
}

type syntaxFnReference struct {
	relation *syntaxRoute
	args     []Token
	owner    *SyntaxTree
}

func (s *syntaxFnReference) runIntoStack(p *Parser, stack *[]Instruction) (OBJType, error) {
	operandTuple, err := solveContextedTypeFromTokens(s.args, p, true)
	if err != nil {
		return nil, err
	}
	if operandTuple.Primitive() != TUPLE {
		return nil, errors.New("Expecting tuple in order to identify function arguments")
	}
	args := operandTuple.FixedItems()
	if s.relation == nil || len(s.relation.route) == 0 {
		return Void, errors.New("No way to fetch function")
	}
	if s.relation.origin != nil {
		return Void, errors.New("A function must always be global defined")
	}
	if len(s.relation.route) > 1 {
		return Void, errors.New("Not implemented function route navigation")
	}
	sym, err := p.getSymbolForCall(s.relation.route[0], false, args)
	if err != nil {
		return nil, err
	}
	fntp, cname := p.getFunctionTypeFrom(s.relation.route[0], sym)
	*stack = append(*stack, MKInstruction(PSH, cname))
	return fntp, nil
}

type syntaxOpReference struct {
	identifier string
	args       []Token
	owner      *SyntaxTree
}

func (s *syntaxOpReference) runIntoStack(p *Parser, stack *[]Instruction) (OBJType, error) {
	operandTuple, err := solveContextedTypeFromTokens(s.args, p, true)
	if err != nil {
		return nil, err
	}
	if operandTuple.Primitive() != TUPLE {
		return nil, errors.New("Expecting tuple in order to identify function arguments")
	}
	args := operandTuple.FixedItems()
	sym, err := p.getSymbolForCall(s.identifier, true, args)
	if err != nil {
		return nil, err
	}
	fntp, cname := p.getFunctionTypeFrom(s.identifier, sym)
	*stack = append(*stack, MKInstruction(PSH, cname))
	return fntp, nil
}

type syntaxTypeCreation struct {
	typeref []Token
	owner   *SyntaxTree
}

func (s *syntaxTypeCreation) runIntoStack(p *Parser, stack *[]Instruction) (OBJType, error) {
	obj, e := solveContextedTypeFromTokens(s.typeref, p, false)
	if e != nil {
		return nil, e
	}
	i, e := obj.Create()
	if e != nil {
		return nil, e
	}
	*stack = append(*stack, i...)
	return obj, nil
}

type syntaxHighLevelCall struct {
	relation *syntaxRoute
	operands []syntaxBranch
	spawned  bool
	owner    *SyntaxTree
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
	ops, stacks, err := runBranchesIntoStacks(p, s.operands)
	if err != nil {
		return nil, err
	}
	ret, err := p.solveFunctionCall(s.relation.route[0], false, ops, stacks, stack)
	if err == nil {
		if s.owner.inReturn {
			if (*stack)[len(*stack)-1].Code == CLL {
				(*stack)[len(*stack)-1].Code = JMP
			} else if (*stack)[len(*stack)-1].Code == CLX {
				(*stack)[len(*stack)-1].Code = JMX
			}
		}
	}
	return ret, err
}

type syntaxAssignment struct {
	value  syntaxBranch
	assign syntaxBranchSet
	owner  *SyntaxTree
}

func (s *syntaxAssignment) runIntoStack(p *Parser, stack *[]Instruction) (OBJType, error) {
	return Void, s.assign.runIntoStackSet(p, stack, s.value)
}

type syntaxBranch interface {
	runIntoStack(p *Parser, stack *[]Instruction) (OBJType, error)
}

type syntaxBranchSet interface {
	runIntoStackSet(p *Parser, stack *[]Instruction, branch syntaxBranch) error
}

func runBranchesIntoStacks(p *Parser, branches []syntaxBranch) ([]OBJType, [][]Instruction, error) {
	ops := make([]OBJType, len(branches))
	stacks := make([][]Instruction, len(branches))
	for i := len(branches) - 1; i >= 0; i-- {
		var e error
		if ops[i], e = branches[i].runIntoStack(p, &stacks[i]); e != nil {
			return nil, nil, e
		}
		if ops[i].Primitive() == VOID {
			return nil, nil, errors.New("Using Void as function argument")
		}
	}
	return ops, stacks, nil
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

	leftterm, err := s.identifyExpressionBranch(ttks[0])
	if err != nil {
		return nil, err
	}
	if setter, v := leftterm.(syntaxBranchSet); v {
		return &syntaxAssignment{value: rightterm, assign: setter, owner: s}, nil
	}
	return nil, errors.New("Cannot set")
}

func (s *SyntaxTree) generateFirstLevelExpression(tks []Token, children []Block) (syntaxBranch, error) {
	if len(tks) > 0 {
		if tks[0] == FOR || tks[0] == WHILE {
			return nil, errors.New("Generators are not implemented yet")
		}
		name, args, spawned, err := splitFirstLevelFunctionCall(tks)
		if err != nil {
			return nil, err
		}
		if len(name) != 0 {
			var op syntaxHighLevelCall
			op.owner = s
			op.spawned = spawned
			route, e := getRoute(name)
			if e != nil {
				return nil, e
			}
			op.relation = &syntaxRoute{origin: nil, route: route}
			op.operands, err = s.generateOperands(args)
			if err != nil {
				return nil, err
			}
			return &op, nil
		}
		//No high level function
		return s.generateSecondLevelExpression(tks)
	}
	return nil, nil
}

func (s *SyntaxTree) generateSecondLevelExpression(tks []Token) (syntaxBranch, error) {
	if len(tks) > 0 && (tks[0] == FN || tks[0] == OP) {
		return s.generateFunctionReference(discardOne(tks), tks[0] == OP)
	}
	ttks, err := splitByToken(tks, func(tk Token) bool { return tk.Kind == OperatorToken }, genericPairs, true, true, false)
	if err != nil {
		return nil, err
	}
	return s.generateOperatorBranch(ttks)
}

func (s *SyntaxTree) generateFunctionReference(tks []Token, op bool) (syntaxBranch, error) {
	id, args := readUntilToken(tks, DOUBLES)
	if len(args) == 0 || args[0] != DOUBLES {
		return nil, errors.New("Expecting token ':'")
	} else {
		args = discardOne(args)
	}
	var branch syntaxBranch
	var e error
	if op {
		var o Token
		if o, id, e = expectT(id, OperatorToken); e == nil {
			branch = &syntaxOpReference{o.Data, args, s}
			id = discardOne(id)
		}
		if e == nil {
			e = unexpect(id)
		}
	} else {
		var route []string
		if route, e = getRoute(id); e == nil {
			branch = &syntaxFnReference{&syntaxRoute{nil, route, s}, args, s}
		}
	}
	return branch, e
}

//we asume the odd positions are operators
func (s *SyntaxTree) generateOperatorBranch(tks [][]Token) (syntaxBranch, error) {
	if len(tks) == 0 {
		return nil, errors.New("Expecting expression")
	}
	if len(tks) == 1 {
		return s.identifyExpressionBranch(tks[0])
	}
	if len(tks) == 2 {
		return nil, errors.New("No operator sequence")
	}
	//Find first binary operator
	p := -1
	for i := 1; i < len(tks); i += 2 {
		if len(tks[i-1]) != 0 {
			p = i
			break
		}
	}
	if p < 0 { //Unary operator
		op, err := s.generateOperatorBranch(tks[2:])
		return &syntaxOpCall{
			operator: tks[1][0].Data,
			operands: []syntaxBranch{op},
		}, err
	}
	left, err := s.generateOperatorBranch(tks[0:p])
	if err != nil {
		return nil, err
	}
	right, err := s.generateOperatorBranch(tks[p+1:])
	if err != nil {
		return nil, err
	}
	opb := syntaxOpCall{
		operator: tks[p][0].Data,
		operands: []syntaxBranch{left, right},
	}
	return &opb, nil
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
		literal := &syntaxLiteral{tks[0].Data, tks[0].Kind, s}
		if len(tks) > 1 {
			return s.identifySubrouting(literal, tks[1:])
		}
		return literal, nil
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
			preceded = &syntaxTypeCreation{inner, s}
		} else { //Indexation
			route, err := getRoute(left)
			if err != nil {
				return nil, err
			}
			if len(inner) == 1 && inner[0].Kind == IntegerToken {
				i, err := strconv.Atoi(inner[0].Data)
				e = err
				preceded = &syntaxConstantAccess{target: &syntaxRoute{nil, route, s}, idx: i, owner: s}
			} else {
				indexer, err := s.generateSecondLevelExpression(inner)
				e = err
				preceded = &syntaxOpCall{operator: INDEXOP.Data, operands: []syntaxBranch{&syntaxRoute{nil, route, s}, indexer}}
			}
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
		if !next(left, DOT) {
			var id Token
			if id, left, e = expectT(left, IdToken); e != nil {
				return nil, e
			}
			if e = unexpect(left); e != nil {
				return nil, e
			}
			return &syntaxCast{preceded, id.Data}, nil
		}
		left = discardOne(left)
		route, e := getRoute(left)
		if e != nil {
			return nil, e
		}
		return &syntaxRoute{preceded, route, s}, nil
	} else {
		branch := preceded
		if len(left) != 0 {
			if left, e = expect(left, DOT); e != nil {
				return nil, e
			}
			route, e := getRoute(left)
			if e != nil {
				return nil, e
			}
			branch = &syntaxRoute{preceded, route, s}
		}
		if len(inner) == 1 && inner[0].Kind == IntegerToken {
			i, err := strconv.Atoi(inner[0].Data)
			e = err
			idx = &syntaxConstantAccess{target: branch, idx: i, owner: s}
		} else {
			indexer, err := s.generateSecondLevelExpression(inner)
			e = err
			idx = &syntaxOpCall{operator: INDEXOP.Data, operands: []syntaxBranch{branch, indexer}, owner: s}
		}
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

func splitFirstLevelFunctionCall(tks []Token) (name []Token, args []Token, spawned bool, err error) {
	split := indexOfFirstToken(tks, DOUBLES)
	if split >= 0 {
		for _, p := range genericPairs {
			i := indexOfFirstToken(tks, p.open)
			if i >= 0 && i < split {
				return
			}
		}
		name = tks[:split]
		args = tks[split+1:]
		if len(args) > 0 {
			spawned = args[len(args)-1] == SPAWN
			if spawned {
				args = args[:len(args)-1]
			}
		}
	}
	return
}
