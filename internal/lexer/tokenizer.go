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
	OperatorToken TokenType = 4
	IntegerToken  TokenType = 8
	DecimalToken  TokenType = 16
	StringToken   TokenType = 32
)

type Token struct {
	Data string
	Kind TokenType
}

var string_mark rune = '"'
var decimal_mark rune = '.'
var keywords []string = []string{"require", "import", "struct", "fn", "do", "val", "var"}

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
	if mask == IntegerToken || mask == DecimalToken || mask == OperatorToken || mask == StringToken {
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

//action: It the way tokens will be treat after mask fushions
func updateMask(mask TokenType, char rune) (newmask TokenType, action int8, err error) {
	newmask = mask
	if unicode.IsSymbol(char) || unicode.IsMark(char) || unicode.IsPunct(char) {
		if char == decimal_mark && mask == IntegerToken {
			newmask = DecimalToken
			action = mergeTokens
		} else if char == string_mark {
			newmask = StringToken
			action = pushNoAdd
		} else {
			newmask = OperatorToken
			if mask != OperatorToken {
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

func string_analysis(chars []rune, char rune) (exit_str bool, push []rune) {
	push = []rune{char}
	exit_str = false
	if char == '"' && (len(chars) == 0 || chars[len(chars)-1] != '\\') {
		push = make([]rune, 0)
		exit_str = true
	}
	return
}

func tokens(line string) (tokens []Token, err error) {
	mask := NoneToken
	value := make([]rune, 0)
	characters := []rune(line)

	for i, r := range characters {
		//String lock
		if mask == StringToken {
			end, push := string_analysis(value, r)
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
			m, o, e := updateMask(mask, r)
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
