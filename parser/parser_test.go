// Copyright 2022 Yevgeniy Chaban.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package parser

import (
	"github.com/cyevgeniy/pldoc/ast"
	"testing"
)

var src = []byte(`
-- Test package
create or replace package test is
	;
end test;
`)

func TestPackageCount(t *testing.T) {
	file := ParseFile("testfile", src)
	if len(file.Packages) != 1 {
		// Seems like parsePackage return empty package. Need to
		// fix its return statement
		t.Fatalf("Expected 1 package, got %d", len(file.Packages))
	}
}

var pckDocs = []struct {
	src []byte
	doc string
}{
	{[]byte(`--pck docs
		create or replace package test is
		null;
		end test;`),
		"pck docs\n",
	},
	{
		[]byte("/*pck\ndocs\nis\ngreat\n*/\ncreate package pck is null; end pck"),
		"pck\ndocs\nis\ngreat\n",
	},
	// Currently not supported. We expect package's documentation at one line up
	// than the "package" keyword
	//
	// {
	// 	[]byte("/*pck\ndocs\nis\ngreat\n*/\ncreate\n package\n pck\n is\n null; end pck"),
	// 	"pck\ndocs\nis\ngreat\n",
	// },
}

func TestPackageDoc(t *testing.T) {
	for i := 0; i < len(pckDocs); i++ {
		file := ParseFile("testfile", pckDocs[i].src)
		docText := file.Packages[0].Doc.Text()
		if docText != string(pckDocs[i].doc) {
			t.Fatalf("Package docs error. Expected %s; Got: %s\n", string(pckDocs[i].doc), docText)
		}
	}
}

var funcDocs = []struct {
	src []byte
	doc string
}{
	{
		[]byte("create package test is\n--func docs\nfunction t return number; end test;"),
		"func docs\n",
	},
	{
		[]byte("create package test is\n--proc docs\nprocedure t; end test;"),
		"proc docs\n",
	},
	{
		[]byte("create package test is\nl number;--var docs\n--proc docs\nprocedure t; end test;"),
		"proc docs\n",
	},
	{
		[]byte("create package test is\n--proc docs\n--group\n--\nprocedure t; end test;"),
		"proc docs\ngroup\n",
	},
}

func TestFuncDocs(t *testing.T) {
	for i := range funcDocs {
		file := ParseFile("testfile", funcDocs[i].src)
		doc := file.Packages[0].FuncSpecs[0].Doc.Text()
		if doc != funcDocs[i].doc {
			t.Fatalf("Func docs error; Expected %s; Got %s\n", funcDocs[i].doc, doc)
		}
	}
}

var funcs = []struct {
	src  []byte
	name string
	cnt  byte
}{
	{
		[]byte("create package test is\nfunction\n --t\nf return number;PROCEDURE a; procedure b; end test;"),
		"f",
		3,
	},
	{
		[]byte("create package test is\n--proc docs\nfunction\n\n t_funcname(pvar number) return number; end test;"),
		"t_funcname",
		1,
	},
	{
		[]byte("create package test is\nl number;--var docs\n--proc docs\nprocedure my_proc_name; end test;"),
		"my_proc_name",
		1,
	},
	{
		[]byte("create package test is\n--proc docs\n--group\n--\nprocedure t; end test;"),
		"t",
		1,
	},
}

func TestFuncs(t *testing.T) {
	for i := range funcs {
		file := ParseFile("testfile", funcs[i].src)
		name := file.Packages[0].FuncSpecs[0].Name.Name
		if name != funcs[i].name {
			t.Fatalf("Func docs error; Expected %s; Got %s\n", funcs[i].name, name)
		}

		cnt := byte(len(file.Packages[0].FuncSpecs))
		if cnt != funcs[i].cnt {
			t.Fatalf("Func count error; Expected %d functions; Got %d\n", funcs[i].cnt, cnt)
		}
	}
}

var balancedParensSrc = []struct {
	Src []byte
	Exp string
}{
	{[]byte(`(200)`), "(200)"},
	{[]byte("(t.someval(300))"), "(t.someval(300))"},
	{[]byte("()"), "()"},
}

func TestBalancedParens(t *testing.T) {
	var p Parser
	for i := range balancedParensSrc {
		p.Init("testfile", balancedParensSrc[i].Src, false)

		res := p.scanBalancedParens()

		if res != balancedParensSrc[i].Exp {
			t.Fatalf("Balanced parens scanning error. Expected content: %s; Given content: %s", balancedParensSrc[i].Exp, res)
		}
	}

}

var paramsSrc = `
create or replace package tst is

function tst(pvar number) return number;
function tst(pvar2 varchar2 default null, pid_value pck_package.t_mytype) return varchar2;
procedure tst(
	pname_of_the_param in out pls_integer default 3.14,
	pvar3 in date default sysdate
);
procedure tst(
	pvar4 mytable.id%type default pck_const.id_default,
	pvar_row table_name%rowtype
);

procedure tst(
	pvar5 mytable.id%type, -- comment
	-- comment
	pvar_row2 table_name%rowtype,
	pvar6 schema.tablename.column%type,
    -- In function declarations, we may meet a 
    -- parameter which is typed as some keyword
    type schema.tablename."column"%type,
    -- Same issue as with previous parameter
    exception in clob default empty_clob(),
    pvar9 in number
);
end tst;
`

var parCnt []int = []int{1, 2, 2, 2, 6}

func TestFuncParamsCnt(t *testing.T) {
	file := ParseFile("testfile", []byte(paramsSrc))

	fc := file.Packages[0].FuncSpecs
	for i := range fc {
		if len(fc[i].Params.List) != parCnt[i] {
			t.Fatalf("Func params count error; Expected %d params; Got %d\n", parCnt[i], len(fc[i].Params.List))
		}
	}
}

var parNames []string = []string{"pvar", "pvar2", "pid_value", "pname_of_the_param", "pvar3", "pvar4", "pvar_row", "pvar5", "pvar_row2", "pvar6", "type", "exception", "pvar9"}

func TestFuncParamNames(t *testing.T) {
	file := ParseFile("testfile", []byte(paramsSrc))

	fc := file.Packages[0].FuncSpecs
	params := make([]string, 0, 9)
	for i := range fc {
		for j := range fc[i].Params.List {
			params = append(params, fc[i].Params.List[j].Name.Name)
		}
	}

	for i := range params {
		if params[i] != parNames[i] {
			t.Fatalf("Func params names error; Expected param %s; Got %s\n", parNames[i], params[i])
		}

	}
}

var parTypes []string = []string{"number", "varchar2", "pck_package.t_mytype", "pls_integer", "date", "mytable.id%type", "table_name%rowtype",
	"mytable.id%type", "table_name%rowtype", "schema.tablename.column%type", "schema.tablename.\"column\"%type", "clob", "number"}

func TestParamTypes(t *testing.T) {
	file := ParseFile("testfile", []byte(paramsSrc))

	fc := file.Packages[0].FuncSpecs
	types := make([]string, 0)
	for i := range fc {
		for j := range fc[i].Params.List {
			types = append(types, fc[i].Params.List[j].T.Name)
		}
	}

	for i := range types {
		if types[i] != parTypes[i] {
			t.Fatalf("Func types names error; Expected param %s; Got %s\n", parTypes[i], types[i])
		}

	}
}

var parDefs []string = []string{"null", "3.14", "sysdate", "pck_const.id_default", "empty_clob()"}

func TestParamDefaults(t *testing.T) {
	file := ParseFile("testfile", []byte(paramsSrc))

	fc := file.Packages[0].FuncSpecs
	defs := make([]string, 0)
	for i := range fc {
		for j := range fc[i].Params.List {
			if fc[i].Params.List[j].Def != nil {
				defs = append(defs, fc[i].Params.List[j].Def.Name)
			}
		}
	}

	for i := range defs {
		if defs[i] != parDefs[i] {
			t.Fatalf("Func default values error; Expected value %s; Got %s\n", parDefs[i], defs[i])
		}

	}
}

var varSrc = `
create or replace package tst is

procedure test;

/*
Exception's docs
*/
E_ERROR exception;
-- const docs
c_const constant varchar2(20 char) := 'Hello';
-- var docs
--
myvar number;

function test(pvar number) return number;

-- var1 docs
myvar number(20, 4);

end tst;
/
`

var varDocs []string = []string{"Exception's docs\n", "const docs\n", "var docs\n", "var1 docs\n"}

func TestVarDocs(t *testing.T) {

	file := ParseFile("testfile", []byte(varSrc))

	vd := file.Packages[0].VarDecls

	for i := range vd {
		if vd[i].Doc.Text() != varDocs[i] {
			t.Fatalf("Var docs exception. Expected: %s; Got: %s\n", varDocs[i], vd[i].Doc.Text())
		}
	}
}

var varNames []string = []string{"e_error", "c_const", "myvar", "myvar"}

func TestVarNames(t *testing.T) {
	file := ParseFile("testfile", []byte(varSrc))

	vd := file.Packages[0].VarDecls

	for i := range vd {
		if vd[i].Name.Name != varNames[i] {
			t.Fatalf("Var names exception. Expected: %s; Got: %s\n", varNames[i], vd[i].Name.Name)
		}
	}
}

var curSrc = `
create or replace package test is

-- Cursor 1 documentation
cursor cur1 is
select * from dual;

-- First documemtation string
type t_table is table of number;

-- cur2 docs
cursor cur_2_cursor 
    return my_table%rowtype;

-- t_varray docs
type
	t_varray is varray(20) of number;
cursor cur_3(par_1 number, par2 varchar2) return pck_package.some_type;
end test;
`

func TestCurCount(t *testing.T) {
	file := ParseFile("testfile", []byte(curSrc))

	cnt := len(file.Packages[0].CursorDecls)

	if cnt != 3 {
		t.Fatalf("Cursor's count exception. Expected: %d; Got: %d\n", 3, cnt)
	}
}

var curNames []string = []string{"cur1", "cur_2_cursor", "cur_3"}

func TestCurNames(t *testing.T) {
	file := ParseFile("testfile", []byte(curSrc))

	curs := file.Packages[0].CursorDecls

	for i := range curs {
		if curs[i].Name.Name != curNames[i] {
			t.Fatalf("Cursors' names exception; Expected: %s, Got: %s\n", curNames[i], curs[i].Name.Name)
		}
	}
}

var curParNames []string = []string{"par_1", "par2"}

func TestCurParams(t *testing.T) {
	file := ParseFile("testfile", []byte(curSrc))

	params := file.Packages[0].CursorDecls[2].Params.List

	for i := range params {
		if params[i].Name.Name != curParNames[i] {
			t.Fatalf("Cursor's params exception; Expected: %s; Got: %s\n", curParNames[i], params[i].Name.Name)
		}
	}
}

func TestListTypesCnt(t *testing.T) {
	file := ParseFile("testfile", []byte(curSrc))

	cnt := len(file.Packages[0].TypeDecls)
	if cnt != 2 {
		t.Fatalf("List type's count exception; Expected %d types; Got: %d\n", 2, cnt)
	}
}

var listNames []string = []string{"t_table", "t_varray"}

func TestListTypesNames(t *testing.T) {
	file := ParseFile("testfile", []byte(curSrc))

	ltypes := file.Packages[0].TypeDecls

	for i := range ltypes {
		if ltypes[i].Name.Name != listNames[i] {
			t.Fatalf("List types's names exception; Expected: %s; Got: %s", listNames[i], ltypes[i].Name.Name)
		}
	}
}

var listDocs []string = []string{"First documemtation string\n", "t_varray docs\n"}

func TestListTypesDocs(t *testing.T) {
	file := ParseFile("testfile", []byte(curSrc))

	ltypes := file.Packages[0].TypeDecls

	for i := range ltypes {
		if ltypes[i].Doc.Text() != listDocs[i] {
			t.Fatalf("List types's docs exception; Expected: %s; Got: %s", listDocs[i], ltypes[i].Doc.Text())
		}
	}
}

var listTypes []string = []string{"number", "number"}

func TestListTypesTypes(t *testing.T) {
	file := ParseFile("testfile", []byte(curSrc))

	ltypes := file.Packages[0].TypeDecls

	for i := range ltypes {
		if ltypes[i].T.Name != listTypes[i] {
			t.Fatalf("List type's types exception; Expected: %s; Got: %s", listTypes[i], ltypes[i].T.Name)
		}
	}
}

var recordsSrc = `
create or replace package test is

-- EmpInfo documentation
type EmpInfo is record(
	-- Name field
	name varchar2,
	-- Age field
	age number,
	status emp_table.status%type default pck_package.default()
);

-- Period docs
type Period is record(
	-- Start date field
	start_date date,
	-- End date field
	end_date date
);

end test;
`

func TestRecordTypesCount(t *testing.T) {
	file := ParseFile("testfile", []byte(recordsSrc))

	ltypes := file.Packages[0].TypeDecls

	if len(ltypes) != 2 {
		t.Fatalf("Record types' count exception. Expected: %d; Got: %d", 2, len(ltypes))
	}
}

var recordNames []string = []string{"empinfo", "period"}

func TestRecordtypesNames(t *testing.T) {
	file := ParseFile("testfile", []byte(recordsSrc))

	ltypes := file.Packages[0].TypeDecls

	for i := range ltypes {
		if ltypes[i].Name.Name != recordNames[i] {
			t.Fatalf("Record names' exception. Expected: %s; Got: %s", recordNames[i], ltypes[i].Name.Name)
		}
	}
}

var recordDocs []string = []string{"EmpInfo documentation\n", "Period docs\n"}

func TestRecordtypesDocs(t *testing.T) {
	file := ParseFile("testfile", []byte(recordsSrc))

	ltypes := file.Packages[0].TypeDecls

	for i := range ltypes {
		if ltypes[i].Doc.Text() != recordDocs[i] {
			t.Fatalf("Record docs' exception. Expected: %s; Got: %s", recordDocs[i], ltypes[i].Doc.Text())
		}
	}
}

var recordFieldsCnt []int = []int{3, 2}

func TestRecordtypesFieldCount(t *testing.T) {
	file := ParseFile("testfile", []byte(recordsSrc))

	ltypes := file.Packages[0].TypeDecls

	for i := range ltypes {
		if len(ltypes[i].Params.List) != recordFieldsCnt[i] {
			t.Fatalf("Record fields' count exception. Expected: %d; Got: %d", recordFieldsCnt[i], len(ltypes[i].Params.List))
		}
	}
}

var recFieldNames [][]string = [][]string{
	{"name", "age", "status"},
	{"start_date", "end_date"},
}

func TestRecordFieldNames(t *testing.T) {
	file := ParseFile("testfile", []byte(recordsSrc))

	ltypes := file.Packages[0].TypeDecls

	for i := range ltypes {
		for j := range ltypes[i].Params.List {
			if ltypes[i].Params.List[j].Name.Name != recFieldNames[i][j] {
				t.Fatalf("Record fields' name exception. Expected: %s; Got: %s", recFieldNames[i][j], ltypes[i].Params.List[j].Name.Name)
			}
		}
	}
}

var recFieldTypes [][]string = [][]string{
	{"varchar2", "number", "emp_table.status%type"},
	{"date", "date"},
}

func TestRecordFieldTypes(t *testing.T) {
	file := ParseFile("testfile", []byte(recordsSrc))

	ltypes := file.Packages[0].TypeDecls

	for i := range ltypes {
		for j := range ltypes[i].Params.List {
			if ltypes[i].Params.List[j].T.Name != recFieldTypes[i][j] {
				t.Fatalf("Record fields' type exception. Expected: %s; Got: %s", recFieldTypes[i][j], ltypes[i].Params.List[j].T.Name)
			}
		}
	}
}

var refCurSrc = `
create or replace package test is

procedure test;

function test return number;

c_const constant number := 33;

-- Weakly typed ref cursor
type t_my_cursor is ref cursor;

-- Strongly typed ref cursor
type t_my_cursor_2 is ref cursor
    return
    my_table%rowtype;

end test;
`

var refTest = []struct {
	name string
	typ  *ast.Ident
}{
	{name: "t_my_cursor", typ: nil},
	{name: "t_my_cursor_2",
		typ: &ast.Ident{
			Name: "my_table%rowtype",
		}},
}

func TestRefCursors(t *testing.T) {
	file := ParseFile("testfile", []byte(refCurSrc))

	ltypes := file.Packages[0].TypeDecls

	for i := range ltypes {
		if ltypes[i].Name.Name != refTest[i].name {
			t.Fatalf("Ref cursor's name exception. Expected: %s; Got: %s", refTest[i].name, ltypes[i].Name)
		}
	}
}
