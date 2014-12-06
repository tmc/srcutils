package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"strings"

	"github.com/tmc/refactor_utils/pos"
)

func commandAddArgument(options Options) error {
	r, err := newRefactor(options.args)
	if err != nil {
		return err
	}
	parts := strings.SplitN(options.argument, " ", 2)
	argumentName, argumentType := parts[0], parts[1]

	return r.addArgument(argumentName, argumentType, options.position)
}

func (r *refactor) addArgument(argumentName, argumentType, position string) error {
	qpos, err := r.queryPos(position, false)
	if err != nil {
		return err
	}

	callPositions, callSites, err := r.callersAndCallsites(qpos)

	if err != nil {
		return err
	}

	for _, callPos := range callPositions {
		if err := addArgument(argumentName, argumentType, callPos); err != nil {
			return err
		}
	}

	for _, callSite := range callSites {
		if err := addParameter(argumentName, callSite); err != nil {
			return err
		}
	}

	fmt.Println("callers:", callPositions)
	fmt.Println("call sites:", callSites)

	modifiedFiles := map[*ast.File]bool{}
	for _, pos := range append(callPositions, callSites...) {
		fileNode := pos.Path[len(pos.Path)-1].(*ast.File)
		modifiedFiles[fileNode] = true
	}

	for file, _ := range modifiedFiles {
		var buf bytes.Buffer
		printer.Fprint(&buf, qpos.Fset, file)
		fmt.Println(string(buf.Bytes()))
	}

	if options.write {
		return fmt.Errorf("not implemented")
	}
	return nil
}

func addArgument(name, argType string, position *pos.QueryPos) error {
	if len(position.Path) == 0 {
		return fmt.Errorf("got empty node path")
	}
	node := position.Path[0]

	fieldList, ok := node.(*ast.FieldList)
	if !ok {
		return fmt.Errorf("pos must be in a FieldList, got: %T instead", node)
	}

	newField := &ast.Field{
		Names: []*ast.Ident{{Name: name}},
		Type:  &ast.Ident{Name: argType},
	}
	fieldList.List = append([]*ast.Field{newField}, fieldList.List...)
	return nil
}

func addParameter(name string, position *pos.QueryPos) error {
	if len(position.Path) == 0 {
		return fmt.Errorf("got empty node path")
	}
	node := position.Path[0]

	fieldList, ok := node.(*ast.CallExpr)
	if !ok {
		return fmt.Errorf("pos must be in a CallExpr, got: %T instead", node)
	}

	newParam := &ast.BasicLit{Kind: token.STRING, Value: name}
	fieldList.Args = append([]ast.Expr{newParam}, fieldList.Args...)
	return nil
}
