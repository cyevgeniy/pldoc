// Copyright 2022 Yevgeniy Chaban.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ast

import (
	"testing"
)

type FldStringCase struct {
	fld *Field
	str string
}

var fldStringCases []FldStringCase = []FldStringCase{
	{
		fld: &Field{
			Doc:  nil,
			Kind: VVar,
			Name: &Ident{
				Name: "l_age",
			},
			T: &Ident{
				Name: "number",
			},
		},
		str: "l_age number",
	},

	{
		fld: &Field{
			Doc:  nil,
			Kind: VVar,
			Name: nil,
			T:    nil,
		},
		str: "",
	},

	{
		fld: &Field{
			Doc:  nil,
			Kind: VConst,
			Name: &Ident{
				Name: "l_age",
			},
			T: &Ident{
				Name: "number",
			},
		},
		str: "l_age constant number",
	},

	{
		fld: &Field{
			Doc:  nil,
			Kind: VConst,
			Name: &Ident{
				Name: "l_age",
			},
			T: &Ident{
				Name: "number",
			},
			Def: &Ident{
				Name: "10",
			},
		},
		str: "l_age constant number := 10",
	},

	{
		fld: &Field{
			Doc:  nil,
			Kind: VExc,
			Name: &Ident{
				Name: "e_wrong_value",
			},
		},
		str: "e_wrong_value exception",
	},

	{
		fld: &Field{
			Doc:  nil,
			Kind: VExc,
			Name: &Ident{
				Name: "e_wrong_value",
			},
		},
		str: "e_wrong_value exception",
	},

	{
		fld: &Field{
			Doc:  nil,
			Kind: VPar,
			Name: &Ident{
				Name: "p_var",
			},
			T: &Ident{
				Name: "number",
			},
			Mod: ModNone,
		},
		str: "p_var number",
	},

	{
		fld: &Field{
			Doc:  nil,
			Kind: VPar,
			Name: &Ident{
				Name: "p_var",
			},
			T: &Ident{
				Name: "number",
			},
			Mod: ModIn,
		},
		str: "p_var in number",
	},

	{
		fld: &Field{
			Doc:  nil,
			Kind: VPar,
			Name: &Ident{
				Name: "p_var",
			},
			T: &Ident{
				Name: "number",
			},
			Mod: ModOut,
		},
		str: "p_var out number",
	},

	{
		fld: &Field{
			Doc:  nil,
			Kind: VPar,
			Name: &Ident{
				Name: "p_var",
			},
			T: &Ident{
				Name: "number",
			},
			Mod: ModInOut,
		},
		str: "p_var in out number",
	},

	{
		fld: &Field{
			Doc:  nil,
			Kind: VPar,
			Name: &Ident{
				Name: "p_var",
			},
			T: &Ident{
				Name: "number",
			},
			Mod: ModInOut,
			Def: &Ident{
				Name: "10",
			},
		},
		str: "p_var in out number default 10",
	},
}

func TestFieldString(t *testing.T) {
	for i := range fldStringCases {
		if fldStringCases[i].fld.String() != fldStringCases[i].str {
			t.Fatalf("Field to String exception. Expected: %s; got: %s; Testcase #%d\n",
				fldStringCases[i].str,
				fldStringCases[i].fld.String(),
				i)
		}
	}
}
