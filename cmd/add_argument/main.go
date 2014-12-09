// Program add_argument inserts a new argument into a function and all of it's callers
//
// Example:
//   $ add_argument -arg="foo int" -pos=$GOPATH/src/github.com/tmc/refactor_utils/test/original/z.go:#20 github.com/tmc/refactor_utils/test/original
package main

import (
	"flag"
	"fmt"
	"os"
)

type Options struct {
	position      string   // position
	argument      string   // argument to add
	args          []string // ssa FromArgs
	write         bool
	skipExists    bool   // skip if specified name and type are already present
	packageNameRe string // package name regexp
}

var options Options

func init() {
	flag.StringVar(&options.argument, "arg", "",
		"argument to add to the specified function")
	flag.StringVar(&options.position, "pos", "",
		"Filename and byte offset or extent of a syntax element about which to "+
			"query, e.g. foo.go:#123,#456, bar.go:#123.")
	flag.BoolVar(&options.write, "w", false,
		"write result to (source) file instead of stdout")
	flag.BoolVar(&options.skipExists, "skip-exists", true,
		"if an argument appears to exist already don't add it")
	flag.StringVar(&options.packageNameRe, "package-regexp", ".*",
		"package name regex")
}

func main() {
	flag.Parse()
	options.args = flag.Args()
	if err := commandAddArgument(options); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s.\n", err)
		os.Exit(1)
	}
}
