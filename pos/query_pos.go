package pos

import (
	"fmt"
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/loader"
)

// A QueryPos represents the position provided as input to a query:
// a textual extent in the program's source code, the AST node it
// corresponds to, and the package to which it belongs.
// Instances are created by ParseQueryPos.
//
type QueryPos struct {
	Fset       *token.FileSet
	Start, End token.Pos           // source extent of query
	Path       []ast.Node          // AST path from query node to root of ast.File
	Exact      bool                // 2nd result of PathEnclosingInterval
	Info       *loader.PackageInfo // type info for the queried package
}

// ParseQueryPos parses the source query position pos.
// If needExact, it must identify a single AST subtree;
// this is appropriate for queries that allow fairly arbitrary syntax,
// e.g. "describe".
//
func ParseQueryPos(iprog *loader.Program, posFlag string, needExact bool) (*QueryPos, error) {
	filename, startOffset, endOffset, err := ParsePosFlag(posFlag)
	if err != nil {
		return nil, err
	}
	start, end, err := findQueryPos(iprog.Fset, filename, startOffset, endOffset)

	if err != nil {
		return nil, err
	}
	info, path, exact := iprog.PathEnclosingInterval(start, end)
	if path == nil {
		return nil, fmt.Errorf("no syntax here")
	}
	if needExact && !exact {
		return nil, fmt.Errorf("ambiguous selection within %s", astutil.NodeDescription(path[0]))
	}
	return &QueryPos{iprog.Fset, start, end, path, exact, info}, nil
}
