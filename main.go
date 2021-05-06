package main

import (
	"bufio"
	"flag"
	"fmt"
	"go/ast"
	"go/types"
	"io"
	"os"
	"reflect"
	"strings"

	"golang.org/x/tools/go/packages"
)

const (
	securityTag           = "security"
	noDirectEqualityValue = "nodirectequality"
)

func main() {
	flag.Parse()

	mode := packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo
	cfg := &packages.Config{Mode: mode}
	pkgs, err := packages.Load(cfg, flag.Args()...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load: %v\n", err)
		os.Exit(1)
	}
	if packages.PrintErrors(pkgs) > 0 {
		os.Exit(1)
	}

	rawComparisonsFound := false

	for _, pkg := range pkgs {
		for _, f := range pkg.Syntax {
			hasRawComparisons := walkFile(pkg, f)
			if hasRawComparisons {
				rawComparisonsFound = true
			}
		}
	}

	if rawComparisonsFound {
		os.Exit(1)
	}
}

func walkFile(pkg *packages.Package, file *ast.File) bool {
	rawComparisonsFound := false

	ast.Inspect(file, func(n ast.Node) bool {

		ifStmt, ok := n.(*ast.IfStmt)
		if ok {

			binaryCondition, ok := ifStmt.Cond.(*ast.BinaryExpr)
			if ok {

				left := binaryCondition.X
				right := binaryCondition.Y

				if prohibited, fieldName := isProhibited(pkg, left); prohibited {
					printWarningMessageForExpression(pkg, left, fieldName)
					rawComparisonsFound = true
				}
				if prohibited, fieldName := isProhibited(pkg, right); prohibited {
					printWarningMessageForExpression(pkg, right, fieldName)
					rawComparisonsFound = true
				}
			}

			return true
		}
		return true
	})

	return rawComparisonsFound
}

func printWarningMessageForExpression(pkg *packages.Package, expr ast.Expr, fieldName string) {
	fmt.Printf("\033[1;31m[SECURITY]\033[0m Found raw comparison of field '%s'. Use constant time comparison function.\n", fieldName)
	pos := pkg.Fset.Position(expr.Pos())
	fmt.Printf("%s:%d\n", pos.Filename, pos.Line)

	code, err := lineOfCode(pos.Filename, pos.Offset)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	fmt.Println(code)
}

func lineOfCode(path string, offset int) (string, error) {
	readFile, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer readFile.Close()

	scanner := bufio.NewScanner(readFile)
	scanner.Split(bufio.ScanLines)

	if _, err := readFile.Seek(int64(offset), io.SeekStart); err != nil {
		return "", fmt.Errorf("Seek to line numer (%d) failed for file (%s)", offset, path)
	}

	if scanner.Scan() {
		return scanner.Text(), nil
	}

	return "", fmt.Errorf("line number (%d) exceeded max of file (%s)", offset, path)
}

// returns a types.Struct by diving into underlying types or nil if not applicable
func structFromSelectorXType(xType types.Type) *types.Struct {
	xTypeStruct, ok := xType.(*types.Struct)
	if ok {
		return xTypeStruct
	}

	xTypePointer, ok := xType.(*types.Pointer)
	if ok {
		return structFromSelectorXType(xTypePointer.Elem())
	}

	xTypeNamed, ok := xType.(*types.Named)
	if ok {
		return structFromSelectorXType(xTypeNamed.Underlying())
	}

	return nil
}

func isProhibited(pkg *packages.Package, x ast.Expr) (bool, string) {
	selectorExpr, ok := x.(*ast.SelectorExpr)
	if ok {
		return isProbibitedSelector(pkg, selectorExpr)
	}

	return false, ""
}

// Returns true if the selector's field name is one which has the tag to indicate it shouldn't be directly compared
// When it returns true, it also returns the field name
func isProbibitedSelector(pkg *packages.Package, selectorExpr *ast.SelectorExpr) (bool, string) {
	if selectorExpr.Sel == nil || len(selectorExpr.Sel.Name) == 0 {
		return false, ""
	}

	selectorFieldName := selectorExpr.Sel.Name
	xTypeAndValue := pkg.TypesInfo.Types[selectorExpr.X]

	xTypeStruct := structFromSelectorXType(xTypeAndValue.Type)
	if xTypeStruct == nil {
		return false, ""
	}

	result := isProhibitedField(xTypeStruct, selectorFieldName)
	return result, selectorFieldName
}

func isProhibitedField(definition *types.Struct, fieldSelector string) bool {
	for i := 0; i < definition.NumFields(); i++ {
		field := definition.Field(i)
		if field.Name() == fieldSelector {
			tag := definition.Tag(i)
			return hasNoDirectEqualityTag(tag)
		}
	}
	return false
}

// `json:"token,omitempty" security:"nodirectequality"` -> true
func hasNoDirectEqualityTag(tagString string) bool {

	structTag := reflect.StructTag(tagString)

	tag := structTag.Get(securityTag)
	values := strings.Split(tag, ",")
	for _, value := range values {
		if value == noDirectEqualityValue {
			return true
		}
	}

	return false
}
