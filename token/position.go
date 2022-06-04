// Copyright 2010 The Go Authors. All rights reserved.
// Copyright 2022 Yevgeniy Chaban.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package token

import (
	"fmt"
)

type Position struct {
	Filename string // Filename, under the name it was opened
	Offset   int    // Offset, starting at 0
	Line     int    // Line number, starting at 1
	Column   int    // Column number, starting at 1
}

// For a file, this is the offset.
// It may be converted to Position, if needed.
type Pos int

const NoPos Pos = 0

func (pos Position) String() string {
	return fmt.Sprintf("File: %s; Line: %d; Position: %d", pos.Filename, pos.Line, pos.Column)
}

type File struct {
	Filename string
	lines    []int // offsets of the first characters for each line(first is always 0)
}

func NewFile(filename string) *File {
	return &File{Filename: filename, lines: []int{0}}
}

func (f *File) AddLine(offs int) {
	if f.lines[len(f.lines)-1] < offs {
		f.lines = append(f.lines, offs)
	}
}

func (f *File) Pos(offs int) Pos {
	return Pos(offs)
}

func (f *File) Line(pos Pos) int {
	// Later this search should be optimized
	var i int
	for i = len(f.lines) - 1; i > 0; i-- {
		if int(pos) >= f.lines[i] {
			break
		}
	}

	return i + 1
}

func (f *File) Position(p Pos) *Position {
	pos := Position{}

	pos.Line = f.Line(p)
	pos.Column = int(p) - f.lines[pos.Line-1] + 1
	pos.Filename = f.Filename

	return &pos
}
