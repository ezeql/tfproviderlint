// Package acctestcheckdestroy defines an Analyzer that checks for
// TestCase missing CheckDestroy
package acctestcheckdestroy

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const Doc = `check for TestCase missing CheckDestroy

The acctestcheckdestroy analyzer reports likely incorrect uses of TestCase
which do not define a CheckDestroy function. CheckDestroy is used to verify
that test infrastructure has been removed at the end of an acceptance test.

More information can be found at:
https://www.terraform.io/docs/extend/testing/acceptance-tests/testcase.html#checkdestroy`

var Analyzer = &analysis.Analyzer{
	Name:     "acctestcheckdestroy",
	Doc:      Doc,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	testCases := ResourceTestCases(pass)
	for _, testCase := range testCases {
		var found bool

		for _, elt := range testCase.Elts {
			switch v := elt.(type) {
			default:
				continue
			case *ast.KeyValueExpr:
				if v.Key.(*ast.Ident).Name == "CheckDestroy" {
					found = true
					break
				}
			}
		}

		if !found {
			pass.Reportf(testCase.Type.(*ast.SelectorExpr).Sel.Pos(), "missing CheckDestroy")
		}
	}

	return nil, nil
}

// ResourceTestCases returns all github.com/hashicorp/terraform/helper/resource.TestCase AST
func ResourceTestCases(pass *analysis.Pass) []*ast.CompositeLit {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{
		(*ast.CompositeLit)(nil),
	}
	var result []*ast.CompositeLit

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		x := n.(*ast.CompositeLit)

		if !IsResourceTestCase(pass, x) {
			return
		}

		result = append(result, x)
	})

	return result
}

func IsResourceTestCase(pass *analysis.Pass, cl *ast.CompositeLit) bool {
	switch v := cl.Type.(type) {
	default:
		return false
	case *ast.SelectorExpr:
		switch t := pass.TypesInfo.TypeOf(v).(type) {
		default:
			return false
		case *types.Named:
			if t.Obj().Name() != "TestCase" {
				return false
			}
			// HasSuffix here due to vendoring
			if !strings.HasSuffix(t.Obj().Pkg().Path(), "github.com/hashicorp/terraform/helper/resource") {
				return false
			}
		}
	}
	return true
}