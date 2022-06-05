// Copyright 2022 Yevgeniy Chaban.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package template

import (
	_ "embed"
	"github.com/cyevgeniy/pldoc/ast"
	"html/template"
	"os"
	"path/filepath"
	"strings"
)

var (
	//go:embed static/single.html
	tmpl string

	//go:embed static/main.css
	css []byte

	//go:embed static/pldoc.js
	js []byte
)

func varHeader(vd *ast.Field) string {
	switch vd.Kind {
	case ast.VConst:
		return "const"
	case ast.VVar:
		return "var"
	case ast.VExc:
		return "exception"
	}

	return ""
}

func funcHeader(fd *ast.FuncSpec) string {
	if fd.Ftype == ast.FtFunc {
		return "function"
	}

	return "procedure"
}

func funcListing(fd *ast.FuncSpec) string {
	res := funcHeader(fd)

	if fd.Name != nil && fd.Name.Name != "" {
		res += " " + fd.Name.Name
	}

	res += fieldListListing(fd.Params)

	if fd.Ftype == ast.FtFunc {
		res += " return " + fd.T.Name
	}

	return res
}

func fieldListListing(fl *ast.FieldList) string {
	if fl == nil {
		return ""
	}

	var res string

	if fl.List != nil {
		res += "(\n"
		for i := range fl.List {
			last := i == len(fl.List)-1
			res += "    " + fl.List[i].String()
			if !last {
				res += ","
			}

			res += "\n"
		}
		res += ")"
	}

	return res
}

func typeHeader(td *ast.TypeDecl) string {
	switch td.Kind {
	case ast.TkTable:
		return "table"
	case ast.TkVarray:
		return "varray"
	case ast.TkRecord:
		return "record"
	case ast.TkRefCursor:
		return "ref cursor"
	}
	return ""
}

func typeListing(td *ast.TypeDecl) string {
	res := "type " + td.Name.Name + " is " + typeHeader(td) +
		fieldListListing(td.Params)

	if (td.Kind == ast.TkVarray || td.Kind == ast.TkTable) && td.T != nil {
		res += " of " + td.T.Name
	}

	return res
}

func cursorListing(cd *ast.CursorDecl) string {
	return "cursor " + cd.Name.Name + fieldListListing(cd.Params) +
		" is\n" + cd.SQL.Text
}

// Returns index of the first non-space
// character in a string. If string doesn't
// have any spaces at the beginning, returns 0
// If string doesn't have non-space characters,
// returns -1
func getFirstCharOffset(s string) int {
	for i := range s {
		if s[i] != ' ' {
			return i
		}
	}

	return -1
}

// Trims n spaces from the left side of a string
func trimNSpaces(s string, n int) string {
	var i int
	for i = 0; i <= n && i < len(s); i++ {
		if s[i] != ' ' {
			break
		}
	}

	return s[i:]

}

func formatComment(cg *ast.CommentGroup) template.HTML {
	if cg == nil {
		return ""
	}

	lines := strings.Split(cg.Text(), "\n")

	if len(lines) == 0 {
		return ""
	}

	// Get the first non-space character's offset for the
	// first line in the comment group.
	offset := getFirstCharOffset(lines[0])

	var res = make([]string, 0)
	var buf = make([]string, 0)
	fromPre := false

	// Preformatted text offset related to ordinary comment text
	var preOffs int

	for i := range lines {
		currOff := getFirstCharOffset(lines[i])

		// Preformatted block starts.
		// All previous text should be formatted as paragraph.
		if currOff > offset {
			// If we have not processing preformatted block
			if !fromPre {
				res = append(res, "<p>"+strings.Join(buf, "\n")+"</p>")
				buf = make([]string, 0)
				preOffs = currOff - offset
				fromPre = true
			}
		} else if fromPre {
			res = append(res, "<pre>"+strings.Join(buf, "\n")+"</pre>")
			buf = make([]string, 0)
			fromPre = false
		}

		var tmp string

		if fromPre {
			tmp = trimNSpaces(lines[i], preOffs)
		} else {
			tmp = lines[i]
		}

		buf = append(buf, tmp)

	}

	// Process unprocessed data
	if len(buf) > 0 {
		var str string

		if fromPre {
			str = "<pre>" + strings.Join(buf, "\n") + "</pre>"
		} else {
			str = "<p>" + strings.Join(buf, "\n") + "</p>"
		}

		res = append(res, str)
	}

	return template.HTML(strings.Join(res, "\n"))
}

type reportData struct {
	Package     *ast.Package
	PackageList []*ast.Package
}

func Execute(dir string, f *ast.Files) error {
	fm := template.FuncMap{
		"varHeader":     varHeader,
		"funcHeader":    funcHeader,
		"funcListing":   funcListing,
		"typeHeader":    typeHeader,
		"typeListing":   typeListing,
		"cursorListing": cursorListing,
		"formatComment": formatComment,
	}

	// Prepare directory
	err := os.Mkdir(dir, 0750)
	if err != nil && !os.IsExist(err) {
		return err
	}

	// Create css and js files
	err = os.WriteFile(filepath.Join(dir, "main.css"), css, 0666)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join(dir, "pldoc.js"), js, 0666)
	if err != nil {
		return err
	}

	
	t, err := template.New("Documentation").Funcs(fm).Parse(tmpl)
	if err != nil {
		return err
	}


	for i := range f.Files {
		for fn := range f.Files[i].Packages {
			// Create file for each pl/sql package
			fdoc, err := os.Create(filepath.Join(dir, f.Files[i].Packages[fn].Name.Name+".html"))
			if err != nil {
				return err
			}

			pckList := f.GetPackages()
			err = t.Execute(fdoc,
				reportData{
					Package:     f.Files[i].Packages[fn],
					PackageList: pckList,
				})

			if err != nil {
				return err
			}

			if err = fdoc.Close(); err != nil {
				panic(err)
			}
		}
	}

	return nil
}
