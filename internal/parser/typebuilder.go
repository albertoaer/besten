package parser

import (
	"errors"
	"fmt"

	. "github.com/Besten/internal/lexer"
)

func solveTypeFromTokens(tokens []Token) (OBJType, error) {
	return solveContextedTypeFromTokens(tokens, nil)
}

func solveContextedTypeFromTokens(tokens []Token, parser *Parser) (OBJType, error) {
	parts, err := splitByToken(tokens, func(t Token) bool { return t == SPLITTER }, []struct {
		open  Token
		close Token
	}{{CBOPEN, CBCLOSE}}, false, false, false)
	if err != nil {
		return nil, err
	}
	return genericSolveType(parts, parser)
}

func genericSolveType(parts [][]Token, parser *Parser) (OBJType, error) {
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
		return solveTypeTuple(parts, parser)
	}
	if len(base) > 1 || base[0].Kind != IdToken {
		//TODO: Modify in order to allow type route
		return nil, errors.New("Type name must be one word identifier")
	}
	if o := isTypeLiteral(base[0]); o != nil {
		if len(parts) > 1 {
			return nil, errors.New("Literal must have no child type")
		}
		return o, nil
	}
	switch base[0].Data {
	case "Vec":
		return solveTypeVec(parts[1:], parser)
	case "Map":
		return solveTypeMap(parts[1:], parser)
	default:
		return nil, errors.New(fmt.Sprintf("Type not available: %s", base[0].Data))
	}
}

func isTypeLiteral(tk Token) OBJType {
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
		/*case Any.TypeName(): //Any type may generate undefined behaviour
		return Any*/
	}
	return nil
}

func solveTypeMap(parts [][]Token, parser *Parser) (OBJType, error) {
	inner, e := genericSolveType(parts, parser)
	if e != nil {
		return nil, e
	}
	return MapOf(inner), nil
}

func solveTypeVec(parts [][]Token, parser *Parser) (OBJType, error) {
	inner, e := genericSolveType(parts, parser)
	if e != nil {
		return nil, e
	}
	return VecOf(inner), nil
}

func solveTypeTuple(parts [][]Token, parser *Parser) (OBJType, error) {
	if len(parts) == 0 {
		return nil, errors.New("Wrong tuple generation type call")
	}
	if len(parts) > 1 {
		return nil, errors.New("Tuple type definition has no child type")
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
		result, e := solveContextedTypeFromTokens(typedef, parser)
		if e != nil {
			return nil, e
		}
		types = append(types, result)
	}
	return TupleOf(types), nil
}
