// Copyright 2009 The Go Authors. All rights reserved.
// Copyright 2022 Yevgeniy Chaban.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ast

import (
	"github.com/cyevgeniy/pldoc/token"
	"strings"
)

// Nodes interfaces

type Node interface {
	Start() token.Pos // position of the first character of the node
	End() token.Pos   // position of the first character immediately after the node
}

type Comment struct {
	Text  string    // Comment text, including dashs or slashes
	First token.Pos // position of the first dash or slash
}

func (c *Comment) Start() token.Pos { return c.First }
func (c *Comment) End() token.Pos   { return token.Pos(int(c.First) + len(c.Text)) }

type CommentGroup struct {
	List []*Comment
}

func (cg *CommentGroup) Start() token.Pos {
	return cg.List[0].First
}

func (cg *CommentGroup) End() token.Pos {
	return cg.List[len(cg.List)-1].End()
}

func isWhitespace(ch byte) bool { return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' }

func stripTrailingWhitespace(s string) string {
	i := len(s)
	for i > 0 && isWhitespace(s[i-1]) {
		i--
	}
	return s[0:i]
}

// Text returns the text of the comment.
// Comment markers (--, /*, and */), the first space of a line comment, and
// leading and trailing empty lines are removed.
// Comment directives like "//line" and "//go:noinline" are also removed.
// Multiple empty lines are reduced to one, and trailing space on lines is trimmed.
// Unless the result is empty, it is newline-terminated.
func (g *CommentGroup) Text() string {
	if g == nil {
		return ""
	}
	comments := make([]string, len(g.List))
	for i, c := range g.List {
		comments[i] = c.Text
	}

	lines := make([]string, 0, 10) // most comments are less than 10 lines
	for _, c := range comments {
		// Remove comment markers.
		// The parser has given us exactly the comment text.
		switch c[1] {
		case '-':
			//-- -style comment (no newline at the end)
			c = c[2:]
			if len(c) == 0 {
				// empty line
				break
			}
			if c[0] == ' ' {
				// strip first space - required for Example tests
				c = c[1:]
				break
			}
		case '*':
			/*-style comment */
			c = c[2 : len(c)-2]
		}

		// Split on newlines.
		cl := strings.Split(c, "\n")

		// Walk lines, stripping trailing white space and adding to list.
		for _, l := range cl {
			lines = append(lines, stripTrailingWhitespace(l))
		}
	}

	// Remove leading blank lines; convert runs of
	// interior blank lines to a single blank line.
	n := 0
	for _, line := range lines {
		if line != "" || n > 0 && lines[n-1] != "" {
			lines[n] = line
			n++
		}
	}
	lines = lines[0:n]

	// Add final "" entry to get trailing newline from Join.
	if n > 0 && lines[n-1] != "" {
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

type Ident struct {
	Name  string
	First token.Pos
}

func (i *Ident) Start() token.Pos { return i.First }
func (i *Ident) End() token.Pos   { return token.Pos(int(i.First) + len(i.Name)) }
func (i *Ident) String() string {
	if i != nil {
		return i.Name
	}

	return ""
}

type FieldMod byte

const (
	ModNone FieldMod = iota
	ModIn
	ModOut
	ModInOut
)

func (fm FieldMod) String() string {
	switch fm {
	case ModNone:
		return ""
	case ModIn:
		return "in"
	case ModOut:
		return "out"
	case ModInOut:
		return "in out"
	}

	return ""
}

// Function's parameter, field in a record type or cursor
type VarType byte

const (
	VConst VarType = iota
	VVar
	VExc
	VPar
)

type Field struct {
	Doc  *CommentGroup
	Name *Ident
	T    *Ident
	Kind VarType  // constant, variable, exception or parameter
	Mod  FieldMod // IN, OUT, or IN OUT param. modNone for variable declaration
	Def  *Ident   // Default value. Nil for exceptions
	Null bool     // not null modificator. Used only in Variable declarations
}

func (f *Field) Start() token.Pos { return f.Name.Start() }
func (f *Field) End() token.Pos   { return f.T.End() }
func (f *Field) String() string {
	if f.Name == nil {
		return ""
	}

	if (f.Kind == VVar || f.Kind == VConst || f.Kind == VPar) &&
		f.T == nil {
		return ""
	}

	var s string

	if f.Kind == VVar {
		s = f.Name.String() + " " + f.T.String()
	}

	if f.Kind == VConst {
		s = f.Name.String() + " constant " + f.T.String()
		if f.Def != nil {
			s += " := " + f.Def.String()
		}
	}

	if f.Kind == VExc {
		s = f.Name.String() + " exception"
	}

	if f.Kind == VPar {
		var modStr string
		if f.Mod == ModNone {
			modStr = " "
		} else {
			modStr = " " + f.Mod.String() + " "
		}

		s = f.Name.String() + modStr + f.T.String()
		if f.Def != nil {
			s += " default " + f.Def.String()
		}
	}

	return s
}

// List of function's parameters, fields in a record type
// declaration or cursor's parameters
type FieldList struct {
	Opening token.Pos
	List    []*Field
	Closing token.Pos
}

func (l *FieldList) Start() token.Pos { return l.Opening }
func (l *FieldList) End() token.Pos   { return l.Closing + 1 }

// Subtype declaration
type SubtypeDecl struct {
	Doc  *CommentGroup
	Name *Ident
	Base *Ident // base type
}

func (s *SubtypeDecl) Start() token.Pos {
	return s.Name.Start()
}

func (s *SubtypeDecl) End() token.Pos {
	return s.Base.End()
}

// Represents raw SQL text.
type Sql struct {
	First token.Pos
	Text  string
}

func (s *Sql) Start() token.Pos {
	return s.First
}

func (s *Sql) End() token.Pos {
	return token.Pos(int(s.First) + len(s.Text))
}

// Cursor declaration
type CursorDecl struct {
	Doc    *CommentGroup
	Name   *Ident
	Params *FieldList
	T      *Ident
	SQL    *Sql
}

func (c *CursorDecl) Start() token.Pos {
	return c.Name.Start()
}

func (c *CursorDecl) End() token.Pos {
	return c.SQL.End()
}

type FuncType byte

const (
	FtFunc = iota
	FtProc
)

// Function specification
type FuncSpec struct {
	Doc           *CommentGroup
	Name          *Ident
	Params        *FieldList
	Ftype         FuncType
	Pipelined     bool   // ignored for procedures
	Deterministic bool   // ignored for procedures
	ResultCache   bool   // ignored for procedures
	T             *Ident // ignored for procedures
}

func (f *FuncSpec) Start() token.Pos {
	return f.Name.Start()
}

func (f *FuncSpec) End() token.Pos {
	return f.T.End()
}

// Table type or Varray type
type TypeKind byte

const (
	TkTable TypeKind = iota
	TkVarray
	TkRecord
	TkRefCursor
)

type TypeDecl struct {
	Doc    *CommentGroup
	Name   *Ident
	Kind   TypeKind
	T      *Ident // for all remaining part of table or varray or ref_cursor declaration
	Params *FieldList
}

func (l *TypeDecl) Start() token.Pos {
	return l.Name.First
}

func (l *TypeDecl) End() token.Pos {
	return l.T.End()
}

// Package specification
type Package struct {
	Doc   *CommentGroup
	First token.Pos // Position of the 'package' token
	Last  token.Pos // Next token after ...end pck_name(most probably semicolon)
	Name  *Ident

	VarDecls     []*Field
	SubtypeDecls []*SubtypeDecl
	FuncSpecs    []*FuncSpec
	CursorDecls  []*CursorDecl
	TypeDecls    []*TypeDecl
}

func (p *Package) Start() token.Pos {
	return p.First
}

func (p *Package) End() token.Pos {
	return p.Last
}

// File
type File struct {
	Name     string
	Packages []*Package
}

type Files struct {
	Description string
	Files       []*File
}

func (fset *Files) Add(f *File) {
	fset.Files = append(fset.Files, f)
}

func (fset *Files) GetPackages() []*Package {
	res := make([]*Package, 0)

	for i := range fset.Files {
		res = append(res, fset.Files[i].Packages...)
	}

	return res
}
