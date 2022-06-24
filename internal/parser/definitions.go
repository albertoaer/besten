package parser

import (
	. "github.com/Besten/internal/lexer"
)

var (
	REQUIRE  Token = Token{Data: "require", Kind: KeywordToken}
	IMPORT         = Token{Data: "import", Kind: KeywordToken}
	STRUCTK        = Token{Data: "struct", Kind: KeywordToken}
	FN             = Token{Data: "fn", Kind: KeywordToken}
	DO             = Token{Data: "do", Kind: KeywordToken}
	FOR            = Token{Data: "for", Kind: KeywordToken}
	WHILE          = Token{Data: "while", Kind: KeywordToken}
	VAL            = Token{Data: "val", Kind: KeywordToken}
	VAR            = Token{Data: "var", Kind: KeywordToken}
	DOT            = Token{Data: ".", Kind: OperatorToken}
	INDEXOP        = Token{Data: "[]", Kind: OperatorToken}
	ASSIGN         = Token{Data: "=", Kind: OperatorToken}
	DOUBLES        = Token{Data: ":", Kind: OperatorToken}
	SPLITTER       = Token{Data: "|", Kind: OperatorToken}
	COMA           = Token{Data: ",", Kind: OperatorToken}
	QUOTE          = Token{Data: "'", Kind: OperatorToken}
	POPEN          = Token{Data: "(", Kind: OperatorToken}
	PCLOSE         = Token{Data: ")", Kind: OperatorToken}
	BOPEN          = Token{Data: "[", Kind: OperatorToken}
	BCLOSE         = Token{Data: "]", Kind: OperatorToken}
)
