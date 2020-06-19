package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
)

func main() {
	var jsonTag = flag.String("jsonTag", "timestamp", "json tag for the timestamp field that is unix timestamp")
	var structField = flag.String("structField", "Timestamp", "struct field name that contains timestamp as time.Time")
	var file = flag.String("file", "", "path to generated _easyjson.go file")

	flag.Parse()

	// will create
	// time.Unix(in.Int64(), 0).UTC()
	var fromUnixTimestamp = []ast.Expr{
		&ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X: &ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X: &ast.Ident{
							Name: "time",
						},
						Sel: &ast.Ident{
							Name: "Unix",
						},
					},
					Args: []ast.Expr{
						&ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X: &ast.Ident{
									Name: "in",
								},
								Sel: &ast.Ident{
									Name: "Int64",
								},
							},
						},
						&ast.BasicLit{
							Kind:  token.INT,
							Value: "0",
						},
					},
				},
				Sel: &ast.Ident{
					Name: "UTC",
				},
			},
		},
	}

	// will create:
	// out.Timestamp = time.Unix(in.Int64(), 0).UTC()
	var fromUnixTimestampAssignStmt = &ast.AssignStmt{
		Tok: token.ASSIGN,
		Lhs: []ast.Expr{
			&ast.SelectorExpr{
				X:   &ast.Ident{Name: "out"},
				Sel: &ast.Ident{Name: *structField},
			},
		},
		Rhs: fromUnixTimestamp,
	}

	// will create:
	// out.Int64(in.Timestamp.Unix())
	var toUnixTimestamp = &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "out"},
				Sel: &ast.Ident{Name: "Int64"},
			},
			Args: []ast.Expr{
				&ast.SelectorExpr{
					X: &ast.SelectorExpr{
						X: &ast.Ident{
							Name: "in",
						},
						Sel: &ast.Ident{
							Name: "Timestamp",
						},
					},
					Sel: &ast.Ident{
						Name: "Unix()",
					},
				},
			},
		},
	}

	// parse file
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, *file, nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	astutil.AddNamedImport(fset, node, "", "time")
	ast.SortImports(fset, node)

	ast.Inspect(node, func(n ast.Node) bool {
		f, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}
		if strings.Contains(f.Name.Name, "Decode") {
			// Decode function body
			for _, stmt := range f.Body.List {
				// for !in.IsDelim('}')
				forStmt, ok := stmt.(*ast.ForStmt)
				if !ok {
					continue
				}
				for _, innerForStmt := range forStmt.Body.List {
					// switch key
					switchStmt, ok := innerForStmt.(*ast.SwitchStmt)
					if !ok {
						continue
					}
					for _, caseStmt := range switchStmt.Body.List {
						innerCaseStmt, ok := caseStmt.(*ast.CaseClause)
						if !ok {
							continue
						}
						// we are looking for: case "timestamp"
						if len(innerCaseStmt.List) == 1 {
							// case "timestamp"
							if innerCaseStmt.List[0].(*ast.BasicLit).Value == fmt.Sprintf("\"%s\"", *jsonTag) {
								innerCaseStmt.Body = []ast.Stmt{fromUnixTimestampAssignStmt}
							}
						}
					}
				}
			}
		}
		if strings.Contains(f.Name.Name, "Encode") {
			// Encode function body
			for _, stmt := range f.Body.List {
				innerStmt, ok := stmt.(*ast.BlockStmt)
				if !ok {
					continue
				}
				// looking for a encoder block for timestamp
				if len(innerStmt.List) == 3 {
					// const prefix string = ",\"timestamp\":"
					declStmt, _ := innerStmt.List[0].(*ast.DeclStmt).Decl.(*ast.GenDecl).
						Specs[0].(*ast.ValueSpec).Values[0].(*ast.BasicLit)
					if declStmt.Value == fmt.Sprintf("\",\\\"%s\\\":\"", *jsonTag) {
						innerStmt.List[2] = toUnixTimestamp
					}
				}
			}
		}

		return true
	})

	f, err := os.Create(*file)
	defer f.Close()

	if err := format.Node(f, fset, node); err != nil {
		log.Fatal(err)
	}
}
