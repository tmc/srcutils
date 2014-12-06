package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"io/ioutil"
	"log"
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

	return r.addArgument(argumentName, argumentType, options.position, options.skipExists)
}

func (r *refactor) addArgument(argumentName, argumentType, position string, skipExists bool) error {
	qpos, err := r.queryPos(position, false)
	if err != nil {
		return err
	}

	funcPositions, callSites, err := r.callersAndCallsites(qpos)
	if err != nil {
		return err
	}

	for _, callPos := range funcPositions {
		if err := addArgument(argumentName, argumentType, callPos, skipExists); err != nil {
			log.Println(err)
		}
	}

	for _, callSite := range callSites {
		if err := addParameter(argumentName, callSite, skipExists); err != nil {
			log.Println(err)
		}
	}

	modifiedFiles := map[*ast.File]bool{}
	for _, pos := range append(funcPositions, callSites...) {
		fileNode := pos.Path[len(pos.Path)-1].(*ast.File)
		modifiedFiles[fileNode] = true
	}

	for file, _ := range modifiedFiles {
		var buf bytes.Buffer
		cfg := &printer.Config{Mode: printer.SourcePos}
		cfg.Fprint(&buf, qpos.Fset, file)
		if options.write {
			err := ioutil.WriteFile(r.iprog.Fset.Position(file.Pos()).Filename, buf.Bytes(), 644)
			if err != nil {
				return err
			}
			log.Println("wrote", r.iprog.Fset.Position(file.Pos()).Filename)
		} else {
			fmt.Println(string(buf.Bytes()))
		}
	}

	return nil
}

func addArgument(name, argType string, position *pos.QueryPos, skipExists bool) error {
	if len(position.Path) == 0 {
		return fmt.Errorf("got empty node path")
	}
	node := position.Path[0]

	fieldList, ok := node.(*ast.FieldList)
	if !ok {
		ast.Print(position.Fset, node)
		return fmt.Errorf("pos must be in a FieldList, got: %T instead", node)
	}

	newField := &ast.Field{
		Names: []*ast.Ident{{Name: name}},
		Type:  &ast.Ident{Name: argType},
	}
	if len(fieldList.List) > 0 {
		if fieldList.List[0].Names[0].Name == name &&
			fieldList.List[0].Type.(*ast.Ident).Name == argType {
			return nil
		}
	}
	fieldList.List = append([]*ast.Field{newField}, fieldList.List...)
	return nil
}

func addParameter(name string, position *pos.QueryPos, skipExists bool) error {
	if len(position.Path) == 0 {
		return fmt.Errorf("got empty node path")
	}
	node := position.Path[0]

	fieldList, ok := node.(*ast.CallExpr)
	if !ok {
		return fmt.Errorf("pos must be in a CallExpr, got: %T instead", node)
	}
	newParam := &ast.Ident{Name: name}
	if len(fieldList.Args) > 0 {
		if field, ok := fieldList.Args[0].(*ast.Ident); ok {
			if field.Name == name {
				return nil
			}
		}
	}
	fieldList.Args = append([]ast.Expr{newParam}, fieldList.Args...)
	return nil
}
