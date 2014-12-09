package main

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/token"
	"regexp"
	"spew"

	"github.com/tmc/refactor_utils/pos"

	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/go/pointer"
	"golang.org/x/tools/go/ssa"
)

type refactor struct {
	iprog         *loader.Program
	prog          *ssa.Program
	ptraCfg       *pointer.Config
	packageNameRe *regexp.Regexp
}

func newRefactor(args []string, packageNameRe *regexp.Regexp) (*refactor, error) {
	conf := loader.Config{Build: &build.Default, SourceImports: true}

	args, err := conf.FromArgs(args, true)
	if err != nil {
		return nil, err
	}
	if len(args) > 0 {
		return nil, fmt.Errorf("surplus arguments: %q", args)
	}

	iprog, err := conf.Load()
	if err != nil {
		return nil, err
	}

	var mode ssa.BuilderMode
	prog := ssa.Create(iprog, mode)
	prog.BuildAll()

	// For each initial package (specified on the command line),
	// if it has a main function, analyze that,
	// otherwise analyze its tests, if any.
	var testPkgs, mains []*ssa.Package
	for _, info := range iprog.InitialPackages() {
		initialPkg := prog.Package(info.Pkg)

		// Add package to the pointer analysis scope.
		if initialPkg.Func("main") != nil {
			mains = append(mains, initialPkg)
		} else {
			testPkgs = append(testPkgs, initialPkg)
		}
	}
	if testPkgs != nil {
		if p := prog.CreateTestMainPackage(testPkgs...); p != nil {
			mains = append(mains, p)
		}
	}
	if mains == nil {
		return nil, fmt.Errorf("analysis scope has no main and no tests")
	}

	return &refactor{
		iprog,
		prog,
		&pointer.Config{Mains: mains, BuildCallGraph: true},
		packageNameRe,
	}, nil
}

func (r *refactor) callers(qpos *pos.QueryPos) ([]*pos.QueryPos, error) {
	pkg := r.prog.Package(qpos.Info.Pkg)
	if pkg == nil {
		return nil, fmt.Errorf("no SSA package")
	}
	if !ssa.HasEnclosingFunction(pkg, qpos.Path) {
		return nil, fmt.Errorf("this position is not inside a function")
	}

	target := ssa.EnclosingFunction(pkg, qpos.Path)
	if target == nil {
		return nil, fmt.Errorf("no SSA function built for this location (dead code?)")
	}

	ptrAnalysis, err := pointer.Analyze(r.ptraCfg)
	if err != nil {
		return nil, err
	}

	cg := ptrAnalysis.CallGraph
	cg.DeleteSyntheticNodes()
	edges := cg.CreateNode(target).In

	callers := []*pos.QueryPos{}
	for _, edge := range edges {
		if edge.Caller.ID <= 1 {
			continue
		}
		caller, err := r.posToQueryPos(edge.Pos())
		if err != nil {
			return callers, err
		}
		callers = append(callers, caller)
	}

	return callers, nil
}

func (r *refactor) callersAndCallsites(qpos *pos.QueryPos) ([]*pos.QueryPos, []*pos.QueryPos, error) {
	allCallers, allCallsites := map[token.Pos]*pos.QueryPos{}, map[token.Pos]*pos.QueryPos{}

	err := r.addCallersAndCallsites(qpos, allCallers, allCallsites)
	if err != nil {
		return nil, nil, err
	}

	resultCallers, resultCallsites := []*pos.QueryPos{}, []*pos.QueryPos{}
	for _, caller := range allCallers {
		resultCallers = append(resultCallers, caller)
	}
	for _, caller := range allCallsites {
		resultCallsites = append(resultCallsites, caller)
	}
	return resultCallers, resultCallsites, nil
}

func (r *refactor) addCallersAndCallsites(qpos *pos.QueryPos, allCallers, allCallsites map[token.Pos]*pos.QueryPos) error {
	if _, present := allCallers[qpos.Start]; present {
		return nil
	}
	allCallers[qpos.Start] = qpos
	callers, err := r.callers(qpos)
	if err != nil {
		return err
	}
	for _, caller := range callers {
		allCallsites[caller.Start] = caller

		parent, err := r.parentFunc(caller.Path)
		if err != nil {
			return err
		}
		if parent != nil {
			if err := r.addCallersAndCallsites(parent, allCallers, allCallsites); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *refactor) queryPos(position string, reflection bool) (*pos.QueryPos, error) {
	return pos.ParseQueryPos(r.iprog, position, reflection)
}

func (r *refactor) posToQueryPos(pos token.Pos) (*pos.QueryPos, error) {
	p := r.prog.Fset.Position(pos)
	return r.queryPos(fmt.Sprintf("%s:#%d", p.Filename, p.Offset), false)
}

func (r *refactor) parentFunc(path []ast.Node) (*pos.QueryPos, error) {
	for _, node := range path {
		// TODO consider function literals
		if fn, ok := node.(*ast.FuncDecl); ok {
			return r.posToQueryPos(fn.Type.Params.Pos())
		}
	}
	spew.Dump(path)
	return nil, fmt.Errorf("no parent found")
}
