package parser

import (
	"errors"
	"fmt"

	. "github.com/Besten/internal/lexer"
)

func solveTypeFromTokens(tokens []Token, allowany bool) (OBJType, error) {
	return solveContextedTypeFromTokens(tokens, nil, allowany)
}

func solveContextedTypeFromTokens(tokens []Token, parser *Parser, allowany bool) (OBJType, error) {
	parts, err := splitByToken(tokens, func(t Token) bool { return t == SPLITTER }, genericPairs, false, false, false)
	if err != nil {
		return nil, err
	}
	return genericSolveType(parts, parser, allowany)
}

func genericSolveType(parts [][]Token, parser *Parser, allowany bool) (OBJType, error) {
	if len(parts) == 0 {
		return nil, errors.New("Expecting type")
	}
	base := parts[0]
	if len(base) == 0 {
		return nil, errors.New("Void type")
	}
	if base[0] == REF {
		if parser == nil {
			return nil, errors.New("Referenced type is not valid in the current context")
		}
		if len(parts) > 1 {
			return nil, errors.New("Referenced type has no child type")
		}
		exp := discardOne(base)
		o, _, err := parser.parseExpressionInto(exp, nil, false)
		return o, err
	}
	if base[0] == CBOPEN {
		if base[len(base)-1] != CBCLOSE {
			return nil, errors.New("Expecting tuple closer")
		}
		return solveTypeTuple(parts, parser, allowany)
	}
	if len(base) > 1 || base[0].Kind != IdToken {
		//TODO: Modify in order to allow type route
		return nil, errors.New("Type name must be one word identifier")
	}
	if o := isTypeLiteral(base[0], allowany); o != nil {
		if len(parts) > 1 {
			return nil, errors.New("Literal must have no child type")
		}
		return o, nil
	}
	switch base[0].Data {
	case "Vec":
		return solveTypeVec(parts[1:], parser, allowany)
	case "Map":
		return solveTypeMap(parts[1:], parser, allowany)
	default:
		if parser != nil {
			obj, e := parser.currentScope().FetchType(base[0].Data)
			if len(parts) > 1 {
				return nil, errors.New("Unexpected child type")
			}
			if e != nil {
				return nil, e //Return error first in order to avoid pointer dereference
			}
			return *obj, nil
		}
		return nil, errors.New(fmt.Sprintf("Type not available: %s", base[0].Data))
	}
}

func isTypeLiteral(tk Token, allowany bool) OBJType {
	if tk.Kind != IdToken {
		return nil
	}
	switch tk.Data {
	case Int.TypeName():
		return Int
	case Dec.TypeName():
		return Dec
	case Bool.TypeName():
		return Bool
	case Str.TypeName():
		return Str
	case Any.TypeName():
		if allowany {
			return Any
		}
	}
	return nil
}

func solveTypeMap(parts [][]Token, parser *Parser, allowany bool) (OBJType, error) {
	inner, e := genericSolveType(parts, parser, allowany)
	if e != nil {
		return nil, e
	}
	return MapOf(inner), nil
}

func solveTypeVec(parts [][]Token, parser *Parser, allowany bool) (OBJType, error) {
	inner, e := genericSolveType(parts, parser, allowany)
	if e != nil {
		return nil, e
	}
	return VecOf(inner), nil
}

func solveTypeTuple(parts [][]Token, parser *Parser, allowany bool) (OBJType, error) {
	if len(parts) == 0 {
		return nil, errors.New("Wrong tuple generation type call")
	}
	p := parts[0][1 : len(parts[0])-1] //Asume call was {*data*} form like
	tokens, e := splitByToken(p, func(t Token) bool { return t == COMA }, []struct {
		open  Token
		close Token
	}{{CBOPEN, CBCLOSE}}, false, false, false)
	if e != nil {
		return nil, e
	}
	types := make([]OBJType, 0)
	for _, typedef := range tokens {
		result, e := solveContextedTypeFromTokens(typedef, parser, allowany)
		if e != nil {
			return nil, e
		}
		types = append(types, result)
	}
	if len(parts) == 2 {
		rettype, e := genericSolveType(parts[1:], parser, allowany)
		if e != nil {
			return nil, e
		}
		return FunctionTypeOf(types, rettype), nil
	}
	if len(parts) > 2 {
		return nil, errors.New("Function type definition has no child type")
	}
	return TupleOf(types), nil
}
