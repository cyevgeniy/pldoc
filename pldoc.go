// Copyright 2022 Yevgeniy Chaban.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"github.com/cyevgeniy/pldoc/ast"
	"github.com/cyevgeniy/pldoc/parser"
	"github.com/cyevgeniy/pldoc/template"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func genFileSet(description string, files []string) (*ast.Files, error) {
	var fileSet ast.Files = ast.Files{
		Description: description,
	}

	for i := range files {

		data, err := os.ReadFile(files[i])
		if err != nil {
			log.Fatal(files[i])
			return nil, err
		}

		file := parser.ParseFile(files[i], data)
		fileSet.Add(file)
	}

	return &fileSet, nil

}

func main() {

	var ext = flag.String("ext", "pks", "The extension of specification files")
	var outDir = flag.String("output", ".", "The output directory for documentation")

	flag.Parse()

	args := flag.Args()

	packages := make([]string, 0)

	for i := range args {
		err := filepath.Walk(args[i], func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() && strings.HasSuffix(info.Name(), "."+*ext) {
				packages = append(packages, path)
			}

			return nil
		})

		if err != nil {
			panic(err)
		}
	}

	fset, err := genFileSet("Documentation", packages)

	if err != nil {
		panic(err)
	}

	err = template.Execute(*outDir, fset)

	if err != nil {
		panic(err)
	}
}
