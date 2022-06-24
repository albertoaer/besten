package parser

import (
	. "github.com/Besten/internal/lexer"
)

var (
	IMPORT   Token = Token{Data: "import", Kind: KeywordToken}
	STRUCTK        = Token{Data: "struct", Kind: KeywordToken}
	FN             = Token{Data: "fn", Kind: KeywordToken}
	OP             = Token{Data: "op", Kind: KeywordToken}
	DO             = Token{Data: "do", Kind: KeywordToken}
	IN             = Token{Data: "in", Kind: KeywordToken}
	IF             = Token{Data: "if", Kind: KeywordToken}
	FOR            = Token{Data: "for", Kind: KeywordToken}
	WHILE          = Token{Data: "while", Kind: KeywordToken}
	VAL            = Token{Data: "val", Kind: KeywordToken}
	VAR            = Token{Data: "var", Kind: KeywordToken}
	REF            = Token{Data: "ref", Kind: KeywordToken}
	SPAWN          = Token{Data: "spawn", Kind: KeywordToken}
	DIRECT         = Token{Data: "direct", Kind: KeywordToken}
	DOT            = Token{Data: ".", Kind: SpecialToken}
	INDEXOP        = Token{Data: "[]", Kind: OperatorToken}
	ASSIGN         = Token{Data: "=", Kind: OperatorToken}
	NOTOP          = Token{Data: "!", Kind: OperatorToken}
	DOUBLES        = Token{Data: ":", Kind: SpecialToken}
	SPLITTER       = Token{Data: "|", Kind: OperatorToken}
	COMA           = Token{Data: ",", Kind: SpecialToken}
	QUOTE          = Token{Data: "'", Kind: OperatorToken}
	POPEN          = Token{Data: "(", Kind: SpecialToken}
	PCLOSE         = Token{Data: ")", Kind: SpecialToken}
	BOPEN          = Token{Data: "[", Kind: SpecialToken}
	BCLOSE         = Token{Data: "]", Kind: SpecialToken}
	CBOPEN         = Token{Data: "{", Kind: SpecialToken}
	CBCLOSE        = Token{Data: "}", Kind: SpecialToken}
	TRUE           = Token{Data: "true", Kind: KeywordToken}
	FALSE          = Token{Data: "false", Kind: KeywordToken}
)
