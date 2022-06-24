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
