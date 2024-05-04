// Copyright 2009 The Go Authors. All rights reserved.
// Copyright 2022, 2024 Evgeny Chaban.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package parser

import (
	"fmt"
	"github.com/cyevgeniy/pldoc/ast"
	"github.com/cyevgeniy/pldoc/scanner"
	"github.com/cyevgeniy/pldoc/token"
)

type Parser struct {
	scanner scanner.Scanner
	file    *token.File

	trace bool

	pos token.Pos
	tok token.Token
	lit string

	comments    []*ast.CommentGroup
	leadComment *ast.CommentGroup
	lineComment *ast.CommentGroup

	// For Cursor SQL query's text.
	src []byte
}

func (p *Parser) Init(fname string, src []byte, trace bool) {
	p.file = token.NewFile(fname)
	p.trace = trace
	p.scanner.Init(p.file, src)
	p.pos = token.NoPos
	p.src = src
	p.next()
}

// Read next token
func (p *Parser) next0() {
	if p.trace {
		fmt.Printf("Token=%s, Literal=%s, Line: %d\n", p.tok, p.lit, p.file.Line(p.pos))
	}
	p.pos, p.tok, p.lit = p.scanner.Scan()
}

// Consume a comment and return it and the line on which it ends.
func (p *Parser) consumeComment() (comment *ast.Comment, endline int) {
	// /*-style comments may end on a different line than where they start.
	// Scan the comment for '\n' chars and adjust endline accordingly.
	endline = p.file.Line(p.pos)
	if p.lit[1] == '*' {
		// don't use range here - no need to decode Unicode code points
		for i := 0; i < len(p.lit); i++ {
			if p.lit[i] == '\n' {
				endline++
			}
		}
	}

	comment = &ast.Comment{First: p.pos, Text: p.lit}
	p.next0()

	return
}

// Consume a group of adjacent comments, add it to the parser's
// comments list, and return it together with the line at which
// the last comment in the group ends. A non-comment token or n
// empty lines terminate a comment group.
func (p *Parser) consumeCommentGroup(n int) (comments *ast.CommentGroup, endline int) {
	var list []*ast.Comment
	endline = p.file.Line(p.pos)
	for p.tok == token.COMMENT && p.file.Line(p.pos) <= endline+n {
		var comment *ast.Comment
		comment, endline = p.consumeComment()
		list = append(list, comment)
	}

	// add comment group to the comments list
	comments = &ast.CommentGroup{List: list}
	p.comments = append(p.comments, comments)

	return
}

// Advance to the next non-comment token. In the process, collect
// any comment groups encountered, and remember the last lead and
// line comments.
//
// A lead comment is a comment group that starts and ends in a
// line without any other tokens and that is followed by a non-comment
// token on the line immediately after the comment group.
//
// A line comment is a comment group that follows a non-comment
// token on the same line, and that has no tokens after it on the line
// where it ends.
//
// Lead and line comments may be considered documentation that is
// stored in the AST.
func (p *Parser) next() {
	p.leadComment = nil
	p.lineComment = nil
	prev := p.pos
	p.next0()

	if p.tok == token.COMMENT {
		var comment *ast.CommentGroup
		var endline int

		if p.file.Line(p.pos) == p.file.Line(prev) && prev != token.NoPos {
			// The comment is on same line as the previous token; it
			// cannot be a lead comment but may be a line comment.
			comment, endline = p.consumeCommentGroup(0)
			if p.file.Line(p.pos) != endline || p.tok == token.EOF {
				// The next token is on a different line, thus
				// the last comment group is a line comment.
				p.lineComment = comment
			}
		}

		// consume successor comments, if any
		endline = -1
		for p.tok == token.COMMENT {
			comment, endline = p.consumeCommentGroup(1)
		}

		if endline+1 == p.file.Line(p.pos) {
			// The next token is following on the line immediately after the
			// comment group, thus the last comment group is a lead comment.
			p.leadComment = comment
		}
	}
}

// Scans untill EOF or specified token.
// When this function is completed, p.tok will contain
// tok or token.EOF
func (p *Parser) scanTo(tok token.Token) {
	for p.tok != tok && p.tok != token.EOF {
		p.next()
	}
	return
}

func (p *Parser) parseFile() *ast.File {
	pck := p.parsePackages()

	return &ast.File{
		Name:     p.file.Filename,
		Packages: pck,
	}
}

func (p *Parser) parsePackages() []*ast.Package {
	var pcks []*ast.Package

	for p.tok != token.EOF {
		if pck := p.parsePackage(); pck != nil {
			pcks = append(pcks, pck)
		}
	}
	return pcks
}

func (p *Parser) skipPackageBody() {
	// We are at the beginning of a package. It may
	// be any position before first package's function, procedure or
	// anything else that may contain the `begin` keyword.
	var balance byte = 1

	// Skip until the package's `end` or EOF
	for balance != 0 && p.tok != token.EOF {
		p.next()

		if p.tok == token.BEGIN {
			balance += 1
		} else if p.tok == token.END {
			balance -= 1
		}
	}

	return
}

func (p *Parser) parsePackage() *ast.Package {
	var doc *ast.CommentGroup

	p.scanTo(token.CREATE)
	if p.tok == token.CREATE {
		doc = p.leadComment
	}

	p.scanTo(token.PACKAGE)

	// We may be at EOF or at PACKAGE
	if p.tok == token.PACKAGE {

		pckName := p.parsePackageName()
		if pckName == nil {
			return nil
		}

		pckNodes := p.parsePackageNodes(pckName.Name)

		var fSpecs []*ast.FuncSpec
		var vDecls []*ast.Field
		var sTypeDecls []*ast.SubtypeDecl
		var cDecls []*ast.CursorDecl
		var tDecls []*ast.TypeDecl

		for i := range pckNodes {
			switch pckNodes[i].(type) {
			case *ast.FuncSpec:
				fSpecs = append(fSpecs, pckNodes[i].(*ast.FuncSpec))
			case *ast.SubtypeDecl:
				sTypeDecls = append(sTypeDecls, pckNodes[i].(*ast.SubtypeDecl))
			case *ast.CursorDecl:
				cDecls = append(cDecls, pckNodes[i].(*ast.CursorDecl))
			case *ast.Field:
				vDecls = append(vDecls, pckNodes[i].(*ast.Field))
			case *ast.TypeDecl:
				tDecls = append(tDecls, pckNodes[i].(*ast.TypeDecl))
			}
		}

		return &ast.Package{
			Doc:          doc,
			First:        token.Pos(0),
			Last:         token.Pos(0),
			Name:         pckName,
			VarDecls:     vDecls,
			SubtypeDecls: sTypeDecls,
			FuncSpecs:    fSpecs,
			CursorDecls:  cDecls,
			TypeDecls:    tDecls,
		}
	} else {
		return nil
	}

}

// Function parsePackageName returns package name
// identifier. If package name is specified with
// schema (like "create or replace package sys.utl_pck as ..."),
// the schema is ignored. It means that returned package name identifier
// will be "utl_pck", not "sys.utl_pck". Skips package body and returns nil if
// current position in the source is package's body, but not
// its specification
func (p *Parser) parsePackageName() *ast.Ident  {
	// we are at token "PACKAGE" now
	start := p.pos
	p.next()

	lit := p.lit

	if p.tok == token.BODY {
		p.skipPackageBody()
		return nil
	}

	// Here, our current token may be a schema name (sys.utl_pck)
	// or a package name (utl_pck). Scan next token to be sure
	// what to do next. The next token may be a token.DOT (".") in the
	// first scenario, or one of the (token.AS, token.IS, token.AUTHID) in the
	// second.
	p.test(token.IDENT)
	p.next()

	// Ignore a schema name and parse package name next to it
	if p.tok == token.DOT {
		p.next()

		return &ast.Ident{
			Name: p.lit,
			First: token.Pos(int(p.pos) - len(p.lit)),
		}
	}

	return &ast.Ident{
		Name: lit,
		First: start,
	}
}

func (p *Parser) expect(tok token.Token) token.Pos {
	p.test(tok)

	pos := p.pos
	p.next()

	return pos
}

func (p *Parser) parsePackageNodes(pckName string) []ast.Node {

	var res []ast.Node

	for p.tok != token.EOF {
		p.next()

		if p.tok == token.DOLLAR {
			// skip any conditional compilation statements and
			p.skipCond()
			continue
		}

		if p.tok == token.IDENT {
			v := p.parseField()
			res = append(res, v)
		}

		if p.tok == token.CURSOR {
			doc := p.leadComment
			c := p.parseCursor()
			c.Doc = doc
			res = append(res, c)
		}

		if p.tok == token.FUNCTION || p.tok == token.PROCEDURE {
			fSpec := p.parseFuncSpec()
			res = append(res, fSpec)
		}

		if p.tok == token.PRAGMA {
			// skip type declarations and pragma section for now
			p.scanTo(token.SEMICOLON)
		}

		if p.tok == token.TYPE {
			typ := p.parseType()
			res = append(res, typ)
		}

		if p.tok == token.END {
			p.next()

			if p.tok == token.IDENT && p.lit != pckName {
				p.panic("Incorrect package name! Expecting " + pckName)
			}
			break
		}
	}

	return res
}

func (p *Parser) skipCond() {
	p.next()

	// Scan twice for statements like $$ERROR
	if p.tok == token.DOLLAR {
		p.next()
	}
}

func (p *Parser) parseType() ast.Node {

	var node ast.Node

	doc := p.leadComment
	// Now we are at token.TYPE. Scan next token for
	// type's name
	p.next()
	p.test(token.IDENT)
	name := p.genIdent()

	p.next()
	p.test(token.IS)

	p.next()

	if p.tok == token.TABLE || p.tok == token.VARRAY {
		ltype := p.parseListType()
		ltype.Doc = doc
		ltype.Name = name
		node = ltype
	} else if p.tok == token.RECORD {
		ltype := p.parseRecordType()
		ltype.Doc = doc
		ltype.Name = name
		node = ltype
	} else if p.tok == token.REF {
		ltype := p.parseRefCursorType()
		ltype.Doc = doc
		ltype.Name = name
		node = ltype
	} else {
		// TODO: Check if there are any types that aren't parsed yet
		//       For now, just jump to semicolon
		p.scanTo(token.SEMICOLON)
	}

	return node
}

func (p *Parser) parseRefCursorType() *ast.TypeDecl {
	// We are at REF token now. Check if next token is CURSOR
	p.next()
	p.expect(token.CURSOR)

	var typ *ast.Ident

	if p.tok != token.SEMICOLON {
		// Strongly typed ref cursor. It means that
		// it declared like:
		//   type cur_name is ref cursor returning type%name;
		//
		// So we need to extract its type.
		p.expect(token.RETURN)

		var typeLit string
		start := p.pos

		for {
			if p.tok != token.EOF && p.tok != token.SEMICOLON {
				typeLit += p.lit
				p.next()
			} else {
				break
			}
		}

		typ = &ast.Ident{
			Name:  typeLit,
			First: start,
		}
	}
	return &ast.TypeDecl{
		Kind: ast.TkRefCursor,
		T:    typ,
	}
}

func (p *Parser) parseRecordType() *ast.TypeDecl {
	// Params should be *ast.FieldList
	p.next()
	p.test(token.LPAREN)

	params := p.parseFieldList()

	return &ast.TypeDecl{
		Kind:   ast.TkRecord,
		Params: params,
	}
}

func (p *Parser) parseListType() *ast.TypeDecl {

	var tKind ast.TypeKind
	var typeName string
	start := token.Pos(int(p.pos) - len(p.lit))
	if p.tok == token.TABLE {
		tKind = ast.TkVarray
	} else {
		tKind = ast.TkTable
	}

	// ignore varray size(like varray(40)) and
	// anything before the OF keyword.
	// TODO: Don't ignore varray's size. It is important
	//       information, especially when we print
	//       type's listing in documentation
	p.scanTo(token.OF)

	prevKeyword := false
	for {
		p.next()

		if p.tok != token.EOF && p.tok != token.SEMICOLON {
			if token.IsKeyword(p.tok) {
				if !prevKeyword {
					typeName += " "
				}

				typeName += p.lit + " "
				prevKeyword = true
			} else {
				prevKeyword = false
				typeName += p.lit
			}
		} else {
			break
		}
	}

	return &ast.TypeDecl{
		// Doc and Name should be filled outside this function,
		//
		Kind: tKind,
		T:    &ast.Ident{Name: typeName, First: start},
	}
}

func (p *Parser) parseCursor() *ast.CursorDecl {
	name := p.parseIdent()

	p.next()

	var params *ast.FieldList

	if p.tok == token.LPAREN {
		params = p.parseFieldList()
	}

	var t *ast.Ident
	start := token.Pos(-1)

	if p.tok == token.RETURN {
		t = p.parseCursorResult()
	}

	var sql ast.Sql

	if p.tok == token.SEMICOLON {
		goto ret
	}

	p.scanTo(token.IS)

	for {
		p.next()
		if p.tok != token.SEMICOLON && p.tok != token.EOF {
			if start == token.Pos(-1) {
				start = p.pos
			}
		} else {
			break
		}
	}

	sql.First = start
	sql.Text = string(p.src[start:p.pos])

ret:

	return &ast.CursorDecl{
		Name:   name,
		Params: params,
		T:      t,
		SQL:    &sql,
	}
}

func (p *Parser) parseField() *ast.Field {

	doc := p.leadComment

	name := p.genIdent()
	p.next()

	if p.tok == token.EXCEPTION {
		return &ast.Field{
			Doc:  doc,
			Name: name,
			T:    &ast.Ident{First: token.Pos(int(p.pos) - len(p.lit)), Name: p.lit},
			Kind: ast.VExc,
		}
	}

	vkind := ast.VVar
	var typeName string
	start := p.pos
	if p.tok == token.CONSTANT {
		vkind = ast.VConst
		start = p.pos
	} else {
		typeName = p.lit
	}

	// For now, compose all before semicolon into the one big Ident
	// In the real world, there can be many options that we need to
	// parse
	for {
		p.next()
		if p.tok != token.EOF && p.tok != token.SEMICOLON {
			lit := p.lit

			// String literals may appear in a field declaration
			// as default values, like:
			//     l_var constant varchar2(20) := 'hello';
			// So, we wrap string literal with "'", because scanner
			// removes quotes.
			if p.tok == token.STRING {
				lit = "'" + p.lit + "'"
			}
			typeName = typeName + lit
		} else {
			break
		}
	}

	return &ast.Field{
		Doc:  doc,
		Name: name,
		T:    &ast.Ident{Name: typeName, First: start},
		Kind: ast.VarType(vkind),
	}
}

func (p *Parser) panic(msg string) {
	panic(fmt.Sprintf("File: %s; line: %d; %s", p.file.Filename,p.file.Line(p.pos), msg))
}

// Generates Ident from the current parser's state.
func (p *Parser) genIdent() *ast.Ident {
	p.test(token.IDENT)

	return &ast.Ident{Name: p.lit, First: token.Pos(int(p.pos) - len(p.lit))}
}

// Get next ident
func (p *Parser) parseIdent() *ast.Ident {
	p.next()

	p.test(token.IDENT)

	return &ast.Ident{Name: p.lit, First: token.Pos(int(p.pos) - len(p.lit))}
}

// Test current token
func (p *Parser) test(tok token.Token) {
	if p.tok != tok {
		p.panic(fmt.Sprintf("Expected token %s, got %s", tok, p.tok))
	}
}

// Parse function/procedure/cursor parameters or
// record fields
func (p *Parser) parseFieldList() *ast.FieldList {
	open := p.pos
	var close token.Pos

	var fields []*ast.Field
	for {
		fields = append(fields, p.parseParam())
		if p.tok == token.RPAREN {
			close = p.pos
			break
		}
		// If we have not reached the final right paren,
		// we expect comma, because there should be
		// another one parameter
		p.test(token.COMMA)
	}

	// Make progress
	p.next()

	return &ast.FieldList{
		Opening: open,
		List:    fields,
		Closing: close,
	}
}

// Parse parameter or record field
//
//	TODO: Has to be refactored into a more
//	      smaller chunks
func (p *Parser) parseParam() *ast.Field {
	// Don't use p.scanIdent here, because
	// function/procedure/record parameter or record field
	// can be a keyword as well as identifier, so we just
	// scan what we can and treat scanned literal as Ident
	p.next()
	ident := &ast.Ident{
		Name:  p.lit,
		First: token.Pos(int(p.pos) - len(p.lit)),
	}

	var doc *ast.CommentGroup
	if p.leadComment != nil {
		doc = p.leadComment
	} else if p.lineComment != nil {
		doc = p.lineComment
	}

	var typ ast.FieldMod = ast.ModNone

	p.next()
	if p.tok == token.OUT {
		typ = ast.ModOut
	} else if p.tok == token.IN {
		p.next()
		if p.tok == token.OUT {
			typ = ast.ModInOut
		} else {
			typ = ast.ModIn
		}
	}

	var parType *ast.Ident

	// If parameter is IN or not specified, then at this point we
	// have the whole type(or its part) in the parser's
	// state. If parameter is OUT or IN OUT, the IN or OUT
	// token is stored in the state. Here, we arrange the state, so
	// we will be at the position of the type's end
	if typ == ast.ModIn || typ == ast.ModNone {
		p.test(token.IDENT)
		// Not need to move forward, the type is already
		// parsed
		parType = p.genIdent()
	} else {
		// Need to move forward to parse type
		parType = p.parseIdent()
	}

	var balance int
	for {
		p.next()

		// We shouldn't stop if current right paren
		// is closing in a field's declaration, like
		// pfield varchar2(2000)
		if p.tok == token.LPAREN {
			balance += 1
		} else if p.tok == token.RPAREN && balance > 0 {
			balance -= 1
			parType.Name = parType.Name + p.lit
			continue
		}

		if p.tok != token.EOF && p.tok != token.DEFAULT && p.tok != token.COMMA && p.tok != token.RPAREN {
			parType.Name = parType.Name + p.lit
		} else {
			break
		}
	}

	// Scan default params
	var start token.Pos
	var name string
	if p.tok == token.DEFAULT {
		for {
			p.next()

			// Remember start position only if current token is
			// the first in a default value
			if int(start) == 0 {
				start = token.Pos(int(p.pos) - len(p.lit))
			}

			var inParens string

			if p.tok == token.LPAREN {
				inParens = p.scanBalancedParens()

				name = name + inParens

				// At the end of scanBalancedParens, current
				// token is the RPAREN, we have to make progress
				// to the next token
				p.next()
			}

			if p.tok != token.EOF && p.tok != token.COMMA && p.tok != token.RPAREN {
				name = name + p.lit
			} else {
				break
			}
		}
	}

	var def *ast.Ident
	if name != "" {
		def = &ast.Ident{First: start, Name: name}
	}

	return &ast.Field{
		Doc:  doc,
		Name: ident,
		T:    parType,
		Mod:  typ,
		Def:  def,
		Null: false,
		Kind: ast.VPar,
	}
}

// Scan string that starts with "(" and ends with
// corresponding closing ")". Any nested parentseses
// are included into result string.
func (p *Parser) scanBalancedParens() string {

	p.test(token.LPAREN)

	balance := 1

	// Because we start from left paren, it should be
	// included into the result
	lit := "("

	for balance != 0 && p.tok != token.EOF {
		p.next()

		lit = lit + p.lit

		if p.tok == token.LPAREN {
			balance += 1
		} else if p.tok == token.RPAREN {
			balance -= 1
		}
	}

	return lit
}

// Returns function modificators: pipelined,
// deterministic or result_cache is enabled
// TODO: Parse function modificators(pipelined, deterministic,
//
//	result_cache
func (p *Parser) parseFuncOpts() (bool, bool, bool) {
	return false, false, false
}

func (p *Parser) parseFuncResult() *ast.Ident {
	// skip RETURN keyword
	p.test(token.RETURN)

	var name string

	for {
		p.next()
		if p.tok != token.EOF &&
			p.tok != token.RESULT_CACHE &&
			p.tok != token.DETERMINISTIC &&
			p.tok != token.PIPELINED &&
			p.tok != token.SEMICOLON {
			name = name + p.lit
		} else {
			break
		}
	}

	return &ast.Ident{Name: name, First: token.Pos(int(p.pos) - len(name))}
}

func (p *Parser) parseCursorResult() *ast.Ident {
	// check if we are really stay at return keyword
	p.test(token.RETURN)

	var name string

	for {
		p.next()
		if p.tok != token.EOF && p.tok != token.IS && p.tok != token.SEMICOLON {
			name = name + p.lit
		} else {
			break
		}
	}

	return &ast.Ident{Name: name, First: token.Pos(int(p.pos) - len(name))}
}

func (p *Parser) parseFuncSpec() *ast.FuncSpec {
	var ftype ast.FuncType

	if p.tok == token.PROCEDURE {
		ftype = ast.FtProc
	} else {
		ftype = ast.FtFunc
	}

	doc := p.leadComment
	name := p.parseIdent()
	var params *ast.FieldList
	var typ *ast.Ident
	var pipelined, deterministic, resultCache bool

	p.next()
	if p.tok == token.LPAREN {
		params = p.parseFieldList()
	}

	if ftype == ast.FtFunc {
		typ = p.parseFuncResult()
		pipelined, deterministic, resultCache = p.parseFuncOpts()
	}

	p.scanTo(token.SEMICOLON)

	return &ast.FuncSpec{
		Doc:           doc,
		Name:          name,
		Params:        params,
		Ftype:         ftype,
		T:             typ,
		Pipelined:     pipelined,
		Deterministic: deterministic,
		ResultCache:   resultCache,
	}
}
