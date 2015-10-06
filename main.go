// Copyright (c) 2015 Arista Networks, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"fmt"
	"go/build"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"syscall"
	"time"
)

var h = flag.String("header", "", "Header of the generated .go files")
var hf = flag.String("headerfile", "", "Header of the generated .go files (from the content of the file)")
var allowdestexists = flag.Bool("allowdestexists", false, "Do not fail if destination folder already exists")

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s <source_pkg> <dest_folder>\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\tsource_folder: the source folder of package to process\n")
	fmt.Fprintf(os.Stderr, "\tdest_folder  : The folder that will be created with the fake .go files\n")
	flag.PrintDefaults()
}

func main() {

	flag.Usage = usage
	flag.Parse()

	args := flag.Args()
	if len(args) != 2 {
		log.Printf("Invalid number of arguments")
		flag.Usage()
		os.Exit(1)
	}
	source := args[0]
	dest := args[1]

	// Get possible header
	header := getHeader()

	// Get "build" package info (to have the list of .go files)
	buildPkg, err := build.ImportDir(source, build.ImportComment)
	if err != nil {
		log.Fatalf("Unable to read package %s: %s", source, err)
	}

	// Create dest directory
	err = os.Mkdir(dest, 0777)
	if err != nil {
		if perr, ok := err.(*os.PathError); !ok || perr.Err != syscall.EEXIST || !*allowdestexists {
			log.Fatalf("Unable to create destination folder %s: %s", dest, err)
		}
	}

	// Parse the list of .go source files
	for _, f := range buildPkg.GoFiles {
		writeFakeFile(path.Join(buildPkg.Dir, f), path.Join(dest, f), header, buildPkg.Name)
	}
}

var set = token.NewFileSet()

func writeFakeFile(srcFile, destFile, header, pkgName string) {
	file, err := parser.ParseFile(set, srcFile, nil, parser.ImportsOnly)
	if err != nil {
		log.Fatalf("Error reading source file %s: %s", srcFile, err)
	}
	// Write dest file
	fd, err := os.Create(destFile)
	if err != nil {
		log.Fatalf("Error creating destination file %s: %s", destFile, err)
	}

	fileInfo, err := os.Stat(srcFile)
	if err != nil {
		log.Fatalf("Unable to get stats for file %s: %s", srcFile, err)
	}

	defer onClose(fd, destFile, fileInfo.ModTime())

	if len(header) > 0 {
		fmt.Fprintf(fd, "%s\n\n", header)
	}
	fmt.Fprintf(fd, "package %s\n\n", pkgName)
	if len(file.Imports) == 0 {
		return
	}
	fmt.Fprintf(fd, "import (\n")
	for _, imp := range file.Imports {
		fmt.Fprintf(fd, "\t_ %s\n", imp.Path.Value)
	}
	fmt.Fprintf(fd, ")\n")
}

func onClose(fd *os.File, destFile string, t time.Time) {
	fd.Close()
	// Preserve the original timestamp.
	os.Chtimes(destFile, t, t)
}

func getHeader() string {
	header := ""
	if len(*h) != 0 {
		header = *h
	} else if len(*hf) != 0 {
		content, err := ioutil.ReadFile(*hf)
		if err != nil {
			log.Fatalf("Unable to read the header file %s: %s", hf, err)
		}
		header = string(content)
	}
	header = strings.Trim(header, " \t\n")

	return header
}
