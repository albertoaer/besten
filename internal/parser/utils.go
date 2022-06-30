package parser

import (
	"fmt"

	. "github.com/besten/internal/lexer"
)

func discardOne(tks []Token) []Token {
	if len(tks) == 0 {
		return tks
	}
	return tks[1:]
}

func isNext(tks []Token) bool {
	return len(tks) > 0
}

func next(tks []Token, t Token) bool {
	return len(tks) > 0 && tks[0] == t
}

func nextT(tks []Token, t TokenType) bool {
	return len(tks) > 0 && tks[0].Kind == t
}

func nextV(tks []Token, v string) bool {
	return len(tks) > 0 && tks[0].Data == v
}

func unexpect(tks []Token) error {
	if len(tks) > 0 {
		return fmt.Errorf("Unexpected token: %s of type %s", tks[0].Data, tks[0].Kind.Representation())
	} else {
		return nil
	}
}

func expect(tks []Token, t Token) ([]Token, error) {
	if len(tks) == 0 || tks[0] != t {
		return nil, fmt.Errorf("Expecting token: %s", t.Data)
	}
	return tks[1:], nil
}

func expectT(tks []Token, t TokenType) (Token, []Token, error) {
	if len(tks) == 0 {
		return Token{}, nil, fmt.Errorf("Expecting token type %s", t.Representation())
	}
	if tks[0].Kind != t {
		return Token{}, nil, fmt.Errorf("Expecting token type %s instead of %s", t.Representation(), tks[0].Kind.Representation())
	}
	return tks[0], tks[1:], nil
}

func expectV(tks []Token, v string) (Token, []Token, error) {
	if len(tks) == 0 || tks[0].Data != v {
		return Token{}, nil, fmt.Errorf("Expecting token: %s", v)
	}
	return tks[0], tks[1:], nil
}

/*
the key will split, the pairs of openers and closers will avoid the inner tokens to be splitted
other options:
	allowvoid: allow empty token list between separators
	lastmustkey: if it is true last token must be also a delimiter
	lastsplit: if last is a key will add a void token array
*/
func splitByToken(data []Token, key func(Token) bool, pairs []struct {
	open  Token
	close Token
}, allowvoid, includekey bool, lastsplit bool) ([][]Token, error) {
	res := make([][]Token, 0)
	current := make([]Token, 0)
	opened_idx := -1
	opened_level := 0
mainloop:
	for i, tk := range data {
		if opened_idx >= 0 {
			if pairs[opened_idx].close == tk {
				opened_level--
			} else if pairs[opened_idx].open == tk {
				opened_level++
			}
			if opened_level == 0 {
				opened_idx = -1
			}
			current = append(current, tk)
		} else {
			for i, p := range pairs {
				if p.open == tk {
					opened_idx = i
					opened_level = 1
					current = append(current, tk)
					continue mainloop
				}
			}
			if key(tk) {
				if (!allowvoid && len(current) == 0) || (!lastsplit && i == len(data)-1) {
					return nil, fmt.Errorf("Unexpected token %s", tk.Data)
				}
				res = append(res, current)
				if includekey {
					res = append(res, []Token{tk})
				}
				if lastsplit && i == len(data)-1 {
					res = append(res, make([]Token, 0))
				}
				current = make([]Token, 0)
			} else {
				current = append(current, tk)
			}
		}
	}
	if opened_level > 0 {
		return nil, fmt.Errorf("%s not closed", pairs[opened_idx].open.Data)
	}
	if len(current) > 0 {
		res = append(res, current)
	}
	return res, nil
}

/*
split the tokens outside from the tokens inside of a delimited block
tokens inside pairs does not count
if after block, there is another one, it doesn't get operated
pairs are not valid inside left part, so will asume there is no block
*/
func blockSubtract(data []Token, open, close Token, pairs []struct {
	open  Token
	close Token
}) (left []Token, inner []Token, right []Token, err error) {
	opened_idx := -1
	opened_level := 0

	stage := 0 //0 left, 1 inside, 2 right
	for _, tk := range data {
		switch stage {
		case 0:
			if tk == open {
				stage = 1
			} else {
				for _, p := range pairs {
					if p.open == tk {
						left = data
						return //Block is inside pairs
					}
				}
				left = append(left, tk)
			}
		case 1:
			if opened_idx >= 0 {
				if pairs[opened_idx].close == tk {
					opened_level--
				} else if pairs[opened_idx].open == tk {
					opened_level++
				}
				if opened_level == 0 {
					opened_idx = -1
				}
				inner = append(inner, tk)
			} else {
				if tk == close {
					stage = 2
				} else {
					for i, p := range pairs {
						if p.open == tk {
							opened_idx = i
							opened_level = 1
						}
					}
					inner = append(inner, tk)
				}
			}
		case 2:
			right = append(right, tk)
		}
	}
	if opened_idx >= 0 {
		err = fmt.Errorf("%s not closed", pairs[opened_idx].open.Data)
	}
	return
}

func readUntilToken(data []Token, key Token) ([]Token, []Token) {
	for i := range data {
		if data[i] == key {
			return data[:i], data[i:]
		}
	}
	return data, nil
}

/*
returns -1 if the token is not found, the index otherwise
*/
func indexOfFirstToken(data []Token, tk Token) int {
	for i := range data {
		if data[i] == tk {
			return i
		}
	}
	return -1
}
