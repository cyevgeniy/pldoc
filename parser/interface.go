// Copyright 2009 The Go Authors. All rights reserved.
// Copyright 2022 Yevgeniy Chaban.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package parser

import (
	"github.com/cyevgeniy/pldoc/ast"
)

func ParseFile(fname string, src []byte) (f *ast.File) {
	if fname == "" {
		panic("Empty file provided")
	}

	var p Parser
	p.Init(fname, src, false)

	f = p.parseFile()

	return
}
