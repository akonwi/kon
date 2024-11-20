package ast

import (
	"fmt"
	"testing"

	"github.com/akonwi/kon/checker"
	tree_sitter_kon "github.com/akonwi/tree-sitter-kon/bindings/go"
	"github.com/google/go-cmp/cmp"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

var treeSitterParser *tree_sitter.Parser

func init() {
	language := tree_sitter.NewLanguage(tree_sitter_kon.Language())
	treeSitterParser = tree_sitter.NewParser()
	treeSitterParser.SetLanguage(language)
}

func TestEmptyProgram(t *testing.T) {
	assertAST(t, []byte(""), &Program{Statements: []Statement{}})
}

func TestVariableDeclarations(t *testing.T) {
	assertAST(t, []byte(`
    let name: Str = "Alice"
    mut age: Num = 30
    let is_student: Bool = true`),
		&Program{
			Statements: []Statement{
				&VariableDeclaration{
					Name:         "name",
					Mutable:      false,
					Type:         checker.StrType,
					InferredType: checker.StrType,
					Value: &StrLiteral{
						Value: `"Alice"`,
					},
				},
				&VariableDeclaration{
					Name:         "age",
					Mutable:      true,
					Type:         checker.NumType,
					InferredType: checker.NumType,
					Value: &NumLiteral{
						Value: "30",
					},
				},
				&VariableDeclaration{
					Name:         "is_student",
					Mutable:      false,
					Type:         checker.BoolType,
					InferredType: checker.BoolType,
					Value: &BoolLiteral{
						Value: true,
					},
				},
			},
		},
	)
}

var compareOptions = cmp.Options{
	cmp.FilterPath(func(p cmp.Path) bool {
		return p.Last().String() == ".BaseNode"
	}, cmp.Ignore()),
}

func assertAST(t *testing.T, input []byte, want *Program) {
	t.Helper()

	tree := treeSitterParser.Parse(input, nil)
	ast, err := NewParser(input, tree).Parse()
	if err != nil {
		t.Fatalf("Error parsing tree: %v", err)
	}

	diff := cmp.Diff(want, ast, compareOptions)
	if diff != "" {
		t.Errorf("Generated code does not match (-want +got):\n%s", diff)
	}
}

// this could be combined with the above tests
func TestVariableTypeInference(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantErrors []checker.Error
	}{
		{
			name:       "Inferred type",
			input:      `let name = "Alice"`,
			wantErrors: []checker.Error{},
		},
		{
			name:  "Str mismatch",
			input: `let name: Str = false`,
			wantErrors: []checker.Error{
				{
					Msg:   "Type mismatch: expected Str, got Bool",
					Start: checker.Position{Line: 1, Column: 17},
					End:   checker.Position{Line: 1, Column: 21},
				},
			},
		},
		{
			name:  "Num mismatch",
			input: `let name: Num = "Alice"`,
			wantErrors: []checker.Error{
				{
					Msg:   "Type mismatch: expected Num, got Str",
					Start: checker.Position{Line: 1, Column: 17},
					End:   checker.Position{Line: 1, Column: 23},
				},
			},
		},
		{
			name:  "Bool mismatch",
			input: `let is_bool: Bool = "Alice"`,
			wantErrors: []checker.Error{
				{
					Msg:   "Type mismatch: expected Bool, got Str",
					Start: checker.Position{Line: 1, Column: 21},
					End:   checker.Position{Line: 1, Column: 27},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := treeSitterParser.Parse([]byte(tt.input), nil)
			parser := NewParser([]byte(tt.input), tree)
			_, err := parser.Parse()
			if err != nil {
				t.Fatal(fmt.Errorf("Error parsing tree: %v", err))
			}

			if len(parser.typeErrors) != len(tt.wantErrors) {
				t.Errorf(
					"There were a different number of errors than expected: wanted %v\n actual errors:\n%v",
					len(tt.wantErrors),
					parser.typeErrors,
				)
			}
			for i, want := range tt.wantErrors {
				if diff := cmp.Diff(want, parser.typeErrors[i]); diff != "" {
					t.Errorf("Error does not match (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestFunctionDeclaration(t *testing.T) {
	assertAST(t, []byte(`
		fn empty() {}
		fn get_msg() {
			"Hello, world!"
		}
		fn greet(person: Str) Str {
		}
		fn add(x: Num, y: Num) Num {
		}
	`), &Program{
		Statements: []Statement{
			&FunctionDeclaration{
				Name:       "empty",
				Parameters: []Parameter{},
				ReturnType: checker.VoidType,
				Body:       []Statement{},
			},
			&FunctionDeclaration{
				Name:       "get_msg",
				Parameters: []Parameter{},
				ReturnType: checker.StrType,
				Body: []Statement{
					&StrLiteral{
						Value: `"Hello, world!"`,
					},
				},
			},
			&FunctionDeclaration{
				Name: "greet",
				Parameters: []Parameter{
					{
						Name: "person",
					},
				},
				ReturnType: checker.StrType,
				Body:       []Statement{},
			},
			&FunctionDeclaration{
				Name: "add",
				Parameters: []Parameter{
					{
						Name: "x",
					},
					{
						Name: "y",
					},
				},
				ReturnType: checker.NumType,
				Body:       []Statement{},
			},
		},
	})
}

func TestFunctionDeclarationTypes(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantErrors []checker.Error
	}{
		{
			name:  "Return type mismatch",
			input: `fn get_greeting(person: Str) Str { 42 }`,
			wantErrors: []checker.Error{
				{
					Msg:   "Type mismatch: expected Str, got Num",
					Start: checker.Position{Line: 1, Column: 36},
					End:   checker.Position{Line: 1, Column: 37},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := treeSitterParser.Parse([]byte(tt.input), nil)
			parser := NewParser([]byte(tt.input), tree)
			_, err := parser.Parse()
			if err != nil {
				t.Fatal(fmt.Errorf("Error parsing tree: %v", err))
			}

			if len(parser.typeErrors) != len(tt.wantErrors) {
				t.Fatalf(
					"There were a different number of errors than expected: wanted %v\n actual errors:\n%v",
					len(tt.wantErrors),
					parser.typeErrors,
				)
			}
			for i, want := range tt.wantErrors {
				if diff := cmp.Diff(want, parser.typeErrors[i]); diff != "" {
					t.Errorf("Error does not match (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestUnaryExpressions(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		ast    *Program
		errors []checker.Error
	}{
		{
			name:  "Valid negation",
			input: `let negative_number = -30`,
			ast: &Program{
				Statements: []Statement{
					&VariableDeclaration{
						Name:         "negative_number",
						Mutable:      false,
						InferredType: checker.NumType,
						Value: &UnaryExpression{
							Operator: Minus,
							Operand: &NumLiteral{
								Value: `30`,
							}},
					},
				},
			},
			errors: []checker.Error{},
		},
		{
			name:  "Invalid negation",
			input: `-false`,
			ast: &Program{
				Statements: []Statement{
					&UnaryExpression{
						Operator: Minus,
						Operand: &BoolLiteral{
							Value: false,
						}},
				},
			},
			errors: []checker.Error{
				{
					Msg: "The '-' operator can only be used on 'Num'",
					Start: checker.Position{
						Line:   1,
						Column: 1,
					},
					End: checker.Position{
						Line:   1,
						Column: 1,
					},
				},
			},
		},
		{
			name:  "Valid bang",
			input: `!false`,
			ast: &Program{
				Statements: []Statement{
					&UnaryExpression{
						Operator: Bang,
						Operand: &BoolLiteral{
							Value: false,
						}},
				},
			},
			errors: []checker.Error{},
		},
		{
			name:  "Invali bang",
			input: `!"foo"`,
			ast: &Program{
				Statements: []Statement{
					&UnaryExpression{
						Operator: Bang,
						Operand: &StrLiteral{
							Value: `"foo"`,
						}},
				},
			},
			errors: []checker.Error{
				{
					Msg: "The '!' operator can only be used on 'Bool'",
					Start: checker.Position{
						Line:   1,
						Column: 1,
					},
					End: checker.Position{
						Line:   1,
						Column: 1,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := treeSitterParser.Parse([]byte(tt.input), nil)
			parser := NewParser([]byte(tt.input), tree)
			ast, err := parser.Parse()
			if err != nil {
				t.Fatal(fmt.Errorf("Error parsing tree: %v", err))
			}

			// Compare the ASTs
			diff := cmp.Diff(tt.ast, ast, compareOptions)
			if diff != "" {
				t.Errorf("Built AST does not match (-want +got):\n%s", diff)
			}

			// Compare the errors
			if len(parser.typeErrors) != len(tt.errors) {
				t.Fatalf(
					"There were a different number of errors than expected: wanted %v\n actual errors:\n%v",
					len(tt.errors),
					parser.typeErrors,
				)
			}
			for i, want := range tt.errors {
				if diff := cmp.Diff(want, parser.typeErrors[i]); diff != "" {
					t.Errorf("Error does not match (-want +got):\n%s", diff)
				}
			}
		})
	}
}
