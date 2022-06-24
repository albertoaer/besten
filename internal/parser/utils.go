package parser

import (
	"errors"
	"fmt"

	. "github.com/Besten/internal/lexer"
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
		return errors.New(fmt.Sprintf("Unexpected token: %v", tks[0]))
	} else {
		return nil
	}
}

func expect(tks []Token, t Token) ([]Token, error) {
	if len(tks) == 0 || tks[0] != t {
		return nil, errors.New(fmt.Sprintf("Expecting token: %s", t.Data))
	}
	return tks[1:], nil
}

func expectT(tks []Token, t TokenType) (Token, []Token, error) {
	if len(tks) == 0 || tks[0].Kind != t {
		return Token{}, nil, errors.New(fmt.Sprintf("Expecting token type %d", t))
	}
	return tks[0], tks[1:], nil
}

func expectV(tks []Token, v string) (Token, []Token, error) {
	if len(tks) == 0 || tks[0].Data != v {
		return Token{}, nil, errors.New(fmt.Sprintf("Expecting token type %s", v))
	}
	return tks[0], tks[1:], nil
}

/*
the key will split, the pairs of openers and closers will avoid the inner tokens to be splitted
other options:
	allowvoid: allow empty token list between separators
	lastmustkey: if it is true last token must be also a delimiter
*/
func splitByToken(data []Token, key Token, pairs []struct {
	open  Token
	close Token
}, allowvoid, lastmustkey bool) ([][]Token, error) {
	res := make([][]Token, 0)
	current := make([]Token, 0)
	opened_idx := make([]int, 0)
	for i, tk := range data {
		if len(opened_idx) > 0 {
			if pairs[opened_idx[len(opened_idx)-1]].close == tk {
				opened_idx = opened_idx[:len(opened_idx)-1]
			}
		} else {
			for i, p := range pairs {
				if p.open == tk {
					opened_idx = append(opened_idx, i)
					break
				}
			}
			if tk == key {
				if (!allowvoid && len(current) == 0) || (i == len(data)-1 && !lastmustkey) {
					return nil, errors.New(fmt.Sprintf("Unexpected token %s", key.Data))
				}
				res = append(res, current)
				current = make([]Token, 0)
			} else {
				current = append(current, tk)
			}
		}
	}
	if len(opened_idx) > 0 {
		return nil, errors.New(fmt.Sprintf("%s not closed", pairs[opened_idx[len(opened_idx)-1]].open.Data))
	}
	if len(current) > 0 {
		if lastmustkey {
			return nil, errors.New(fmt.Sprintf("Expecting token %s", key.Data))
		} else {
			res = append(res, current)
		}
	}
	return res, nil
}

func readUntilToken(data []Token, key Token) ([]Token, []Token) {
	for i := range data {
		if data[i] == key {
			return data[:i], data[i:]
		}
	}
	return data, nil
}
