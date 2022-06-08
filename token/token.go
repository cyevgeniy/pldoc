// Copyright 2009 The Go Authors. All rights reserved.
// Copyright 2022 Yevgeniy Chaban.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package token

import "strconv"

type Token int

const (
	EOF Token = iota
	COMMENT

	literal_start
	IDENT  // identifiers
	NUMBER // floating-point and int
	STRING // 'user'
	literal_end

	operators_start
	ADD // +
	SUB // -
	MUL // *
	DIV // /
	REM // %
	CON // ||
	EXP // **

	EQL // =
	NEQ // <>, !=
	GRT // >
	LSS // <
	GEQ // >=
	LEQ // <=

	ASSIGN // :=
	RANGE  // ..

	LPAREN    // (
	LBRACK    // [
	RPAREN    // )
	RBRACK    // ]
	COMMA     // ,
	DOT       // .
	SEMICOLON // ;
	COLON     // :
	DQUOTE    // "

	operators_end

	keywords_start
	LIKE    // LIKE
	BETWEEN // BETWEEN

	AND           // AND
	OR            // OR
	NOT           // NOT
	AS            //AS
	IS            //IS
	DEFAULT       // DEFAULT
	BEGIN         // BEGIN
	END           // end
	IF            //IF
	ELSE          //ELSE
	ELSIF         // elsif
	NULL          // NULL
	RECORD        // record
	PROCEDURE     // procedure
	FUNCTION      // function
	TYPE          // type
	RETURN        // return
	ROWTYPE       // rowtype
	PACKAGE       // package
	IN            // in
	OUT           // out
	PIPELINED     // pipelined
	DETERMINISTIC // deterministic
	RESULT_CACHE  //result_cache
	PRAGMA        // pragma
	EXCEPTION     // exception
	CONSTANT      // constant
	CURSOR        // cursor
	TABLE         // table
	VARRAY        //varray
	OF            // of
	CREATE        // create
	INDEX         // index
	BY            // by
	REF           // ref

	keywords_end
)

var tokens = [...]string{
	EOF:     "EOF",
	COMMENT: "COMMENT",

	IDENT:  "IDENT",
	NUMBER: "NUMBER",
	STRING: "STRING",

	ADD:    "+",
	SUB:    "-",
	MUL:    "*",
	DIV:    "/",
	REM:    "%",
	CON:    "||",
	EXP:    "**",
	DQUOTE: "\"",

	EQL:     "=",
	NEQ:     "<>",
	GRT:     ">",
	LSS:     "<",
	GEQ:     ">=",
	LEQ:     "<=",
	LIKE:    "like",
	BETWEEN: "between",

	AND: "and", // AND
	OR:  "or",  // OR
	NOT: "not", // NOT

	ASSIGN: ":=",
	RANGE:  "..",

	LPAREN:    "(",
	LBRACK:    "[",
	RPAREN:    ")",
	RBRACK:    "]",
	COMMA:     ",", // ,
	DOT:       ".",
	SEMICOLON: ";", // ;
	COLON:     ":",

	AS:            "as",
	IS:            "is",
	DEFAULT:       "default",
	BEGIN:         "begin",
	END:           "end",
	IF:            "if",
	ELSE:          "else",
	ELSIF:         "elsif",
	NULL:          "null",
	RECORD:        "record",
	PROCEDURE:     "procedure",
	FUNCTION:      "function",
	TYPE:          "type",
	RETURN:        "return",
	ROWTYPE:       "rowtype",
	PACKAGE:       "package",
	IN:            "in",
	OUT:           "out",
	PIPELINED:     "pipelined",
	DETERMINISTIC: "deterministic",
	RESULT_CACHE:  "result_cache",
	PRAGMA:        "pragma",
	EXCEPTION:     "exception",
	CONSTANT:      "constant",
	CURSOR:        "cursor",
	TABLE:         "table",
	VARRAY:        "varray",
	OF:            "of",
	CREATE:        "create",
	INDEX:         "index",
	BY:            "by",
	REF:           "ref",
}

var keywords map[string]Token

func init() {
	keywords = make(map[string]Token)
	for i := keywords_start + 1; i < keywords_end; i++ {
		keywords[tokens[i]] = i
	}
}

// String returns the string corresponding to the token tok.
// For operators, delimiters, and keywords the string is the actual
// token character sequence (e.g., for the token ADD, the string is
// "+"). For all other tokens the string corresponds to the token
// constant name (e.g. for the token IDENT, the string is "IDENT").
//
func (tok Token) String() string {
	s := ""
	if 0 <= tok && tok < Token(len(tokens)) {
		s = tokens[tok]
	}
	if s == "" {
		s = "token(" + strconv.Itoa(int(tok)) + ")"
	}
	return s
}

func Lookup(ident string) Token {
	v, ok := keywords[ident]
	if ok {
		return v
	}

	return IDENT
}

func IsKeyword(tok Token) bool {
	return tok > keywords_start && tok < keywords_end
}
