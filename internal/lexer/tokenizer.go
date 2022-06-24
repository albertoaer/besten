package lexer

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

type TokenType uint8

const (
	NoneToken     TokenType = 0
	KeywordToken  TokenType = 1
	IdToken       TokenType = 2
	SpecialToken  TokenType = 4 //For . , ( ) :
	OperatorToken TokenType = 8
	IntegerToken  TokenType = 16
	DecimalToken  TokenType = 32
	StringToken   TokenType = 64
)

func (ttype TokenType) Representation() string {
	switch ttype {
	case NoneToken:
		return "None"
	case KeywordToken:
		return "Keyword"
	case IdToken:
		return "Identifier"
	case SpecialToken:
		return "Special Markup Identifier"
	case OperatorToken:
		return "Operator"
	case IntegerToken:
		return "Numeric Integer"
	case DecimalToken:
		return "Numeric decimal"
	case StringToken:
		return "String"
	default:
		return "UNKNOWN TYPE"
	}
}

type Token struct {
	Data string
	Kind TokenType
}

var string_mark rune = '"'
var decimal_mark rune = '.'
var underscore_mark rune = '_'
var specials []string = []string{",", ".", "(", ")", ":", "[", "]", "{", "}"}
var keywords []string = []string{"require", "import", "struct", "return", "fn", "op", "do",
	"val", "var", "if", "else", "for", "in", "while", "collect", "done", "throw", "catch",
	"true", "false", "direct", "ref", "break", "continue", "omit", "drop", "alias", "spawn"}

func strArrContains(arr []string, elem string) bool {
	for _, a := range arr {
		if a == elem {
			return true
		}
	}
	return false
}

func maskIs(mask, expect int) bool {
	return (mask & expect) == expect
}

func digit(char rune) bool {
	return char >= 48 && char <= 57
}

func leter(char rune) bool {
	return char >= 48 && char <= 57
}

func solveToken(mask TokenType, value string) (Token, error) {
	if mask == OperatorToken && strArrContains(specials, value) {
		return Token{value, SpecialToken}, nil
	} else if mask == IntegerToken || mask == DecimalToken || mask == OperatorToken || mask == StringToken {
		return Token{value, mask}, nil
	} else if mask == IdToken {
		if strArrContains(keywords, value) {
			mask = KeywordToken
		}
		return Token{value, mask}, nil
	} else if mask == NoneToken {
		return Token{}, errors.New("Trying to solve none type")
	}
	return Token{}, errors.New(fmt.Sprintf("No matching type for: %s", value))
}

const (
	mergeTokens int8 = 0
	pushToken   int8 = 1
	pushNoAdd   int8 = 2
)

func splitOp(prev []rune, char rune) bool {
	if len(prev) == 1 && prev[0] == '[' && char == ']' { //special case for operator []
		return false
	}
	return len(prev) > 0 && (strArrContains(specials, string(prev[len(prev)-1])) || strArrContains(specials, string(char)))
}

//action: It the way tokens will be treat after mask fushions
func updateMask(previous []rune, mask TokenType, char rune) (newmask TokenType, action int8, err error) {
	newmask = mask
	if unicode.IsSymbol(char) || unicode.IsMark(char) || unicode.IsPunct(char) {
		if char == decimal_mark && mask == IntegerToken {
			newmask = DecimalToken
			action = mergeTokens
		} else if char == string_mark {
			newmask = StringToken
			action = pushNoAdd
		} else if char == underscore_mark {
			if mask == IdToken {
				action = mergeTokens
			} else {
				action = pushToken
				newmask = IdToken
			}
		} else {
			newmask = OperatorToken
			if mask != OperatorToken || (mask == OperatorToken && splitOp(previous, char)) {
				action = pushToken
			} else {
				action = mergeTokens
			}
		}
	} else if digit(char) {
		if mask == DecimalToken || mask == IntegerToken || mask == IdToken {
			action = mergeTokens
		} else {
			action = pushToken
			newmask = IntegerToken
		}
	} else if unicode.IsLetter(char) {
		if mask == IdToken {
			action = mergeTokens
		} else {
			action = pushToken
			newmask = IdToken
		}
	} else {
		err = errors.New(fmt.Sprintf("Unexpected character: %s", string(char)))
	}
	return
}

/*
return exit string, characters to append, escaping error
*/
func string_analysis(last rune, char rune) (bool, []rune, error) {
	if last == '\\' {
		switch char {
		case 'n':
			return false, []rune{'\n'}, nil
		case '"':
			return false, []rune{'"'}, nil
		case '\\':
			return false, []rune{'\\'}, nil
		case 'r':
			return false, []rune{'\r'}, nil
		case 't':
			return false, []rune{'\t'}, nil
		case 'v':
			return false, []rune{'\v'}, nil
		case 'b':
			return false, []rune{'\b'}, nil
		case 'a':
			return false, []rune{'\a'}, nil
		case 'f':
			return false, []rune{'\f'}, nil
		default:
			return false, nil, errors.New(fmt.Sprintf("Unexpected scape character %c", char))
		}
	}
	if char == '"' {
		return true, []rune{}, nil
	}
	if char == '\\' {
		return false, []rune{}, nil
	}
	return false, []rune{char}, nil
}

func tokens(line string) (tokens []Token, err error) {
	mask := NoneToken
	value := make([]rune, 0)
	characters := []rune(line)

	for i, r := range characters {
		//String lock
		if mask == StringToken {
			end, push, e := string_analysis(characters[i-1], r)
			if e != nil {
				err = e
				return
			}
			if len(push) > 0 {
				value = append(value, push...)
			}
			if end {
				t, e := solveToken(mask, string(value))
				if e != nil {
					err = e
					return
				}
				tokens = append(tokens, t)
				value = make([]rune, 0)
				mask = NoneToken
			}
			continue
		}

		sep := strings.Contains(separators, string(r))

		//If it is not a separator, we operate with it
		if !sep && r != '#' {
			m, o, e := updateMask(value, mask, r)
			if e != nil {
				err = e
				return
			}
			if o == mergeTokens {
				value = append(value, r)
			} else {
				if len(value) > 0 {
					t, e := solveToken(mask, string(value))
					if e != nil {
						err = e
						return
					}
					tokens = append(tokens, t)
				}
				if o == pushToken {
					value = []rune{r}
				} else {
					value = make([]rune, 0)
				}
			}
			mask = m
		}

		//Solves and inserts the token
		if (i == len(characters)-1 || sep || r == '#') && len(value) > 0 {
			t, e := solveToken(mask, string(value))
			if e != nil {
				err = e
				return
			}
			tokens = append(tokens, t)
			value = make([]rune, 0)
			mask = NoneToken
		}

		//It is a comment so we skip it
		if r == '#' {
			break
		}
	}
	if mask != NoneToken {
		if mask == StringToken {
			err = errors.New("Unclosed string literal")
		} else {
			err = errors.New("Unclosed token")
		}
	}
	return
}

//make_sublevel: Indicates that is creating a new scope
func GetTokens(line string) (result []Token, make_sublevel bool, err error) {
	result, err = tokens(line)
	do := Token{"do", KeywordToken}
	make_sublevel = len(result) > 0 && result[len(result)-1] == do
	return
}
