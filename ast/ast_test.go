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
var compareOptions = cmp.Options{
	cmp.FilterPath(func(p cmp.Path) bool {
		return p.Last().String() == ".BaseNode" || p.Last().String() == ".Range"
	}, cmp.Ignore()),
}

func init() {
	language := tree_sitter.NewLanguage(tree_sitter_kon.Language())
	treeSitterParser = tree_sitter.NewParser()
	treeSitterParser.SetLanguage(language)
}

type test struct {
	name        string
	input       string
	ast         *Program
	diagnostics []checker.Diagnostic
}

func runTests(t *testing.T, tests []test) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := treeSitterParser.Parse([]byte(tt.input), nil)
			parser := NewParser([]byte(tt.input), tree)
			ast, err := parser.Parse()
			if err != nil && len(tt.diagnostics) == 0 {
				t.Fatal(fmt.Errorf("Error parsing tree: %v", err))
			}

			// Compare the ASTs
			if tt.ast != nil {
				diff := cmp.Diff(tt.ast, ast, compareOptions)
				if diff != "" {
					t.Errorf("Built AST does not match (-want +got):\n%s", diff)
				}
			}

			// Compare the errors
			if len(parser.typeErrors) != len(tt.diagnostics) {
				t.Fatalf(
					"There were a different number of errors than expected: wanted %v\n actual errors:\n%v",
					len(tt.diagnostics),
					parser.typeErrors,
				)
			}
			for i, want := range tt.diagnostics {
				if diff := cmp.Diff(want, parser.typeErrors[i], compareOptions); diff != "" {
					t.Errorf("Error does not match (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestEmptyProgram(t *testing.T) {
	runTests(t, []test{
		{
			name:        "Empty program",
			input:       "",
			ast:         &Program{Statements: []Statement{}},
			diagnostics: []checker.Diagnostic{},
		},
	})
}

func TestVariableDeclarations(t *testing.T) {
	tests := []test{
		{
			name: "Valid variables",
			input: `
				let name: Str = "Alice"
    		mut age: Num = 30
      	let is_student: Bool = true`,
			ast: &Program{
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
			diagnostics: []checker.Diagnostic{},
		},
	}

	runTests(t, tests)
}

func TestVariableAssignment(t *testing.T) {
	tests := []test{
		{
			name: "Valid Str variable reassignment",
			input: `
				mut name = "Alice"
				name = "Bob"`,
			ast: &Program{
				Statements: []Statement{
					&VariableDeclaration{
						Mutable:      true,
						Name:         "name",
						InferredType: checker.StrType,
						Value:        &StrLiteral{Value: `"Alice"`},
					},
					&VariableAssignment{
						Name:     "name",
						Operator: Assign,
						Value: &StrLiteral{
							Value: `"Bob"`,
						},
					},
				},
			},
			diagnostics: []checker.Diagnostic{},
		},
		{
			name: "Immutable Str variable reassignment",
			input: `
				let name = "Alice"
				name = "Bob"`,
			ast: &Program{
				Statements: []Statement{
					&VariableDeclaration{
						Mutable:      false,
						Name:         "name",
						InferredType: checker.StrType,
						Value:        &StrLiteral{Value: `"Alice"`},
					},
					&VariableAssignment{
						Name:     "name",
						Operator: Assign,
						Value: &StrLiteral{
							Value: `"Bob"`,
						},
					},
				},
			},
			diagnostics: []checker.Diagnostic{
				{
					Msg: "'name' is not mutable",
				},
			},
		},
		{
			name: "Invalid Str variable reassignment",
			input: `
				mut name = "Alice"
				name = 500`,
			ast: &Program{
				Statements: []Statement{
					&VariableDeclaration{
						Mutable:      true,
						Name:         "name",
						InferredType: checker.StrType,
						Value:        &StrLiteral{Value: `"Alice"`},
					},
					&VariableAssignment{
						Name:     "name",
						Operator: Assign,
						Value: &NumLiteral{
							Value: `500`,
						},
					},
				},
			},
			diagnostics: []checker.Diagnostic{
				{
					Msg: "Expected a 'Str' and received 'Num'",
				},
			},
		},
		{
			name:  "Unknown variable reassignment",
			input: `name = "Bob"`,
			ast: &Program{
				Statements: []Statement{
					&VariableAssignment{
						Name:     "name",
						Operator: Assign,
						Value: &StrLiteral{
							Value: `"Bob"`,
						},
					},
				},
			},
			diagnostics: []checker.Diagnostic{
				{
					Msg: "Undefined: 'name'",
				},
			},
		},
		{
			name: "Valid Num increment assignment",
			input: `
				mut count = 0
				count =+ 2`,
			ast: &Program{
				Statements: []Statement{
					&VariableDeclaration{
						Mutable:      true,
						Name:         "count",
						InferredType: checker.NumType,
						Value:        &NumLiteral{Value: `0`},
					},
					&VariableAssignment{
						Name:     "count",
						Operator: Increment,
						Value: &NumLiteral{
							Value: `2`,
						},
					},
				},
			},
			diagnostics: []checker.Diagnostic{},
		},
		{
			name: "Cannot increment an immutable variable",
			input: `
				let count = 0
				count =+ 2`,
			ast: &Program{
				Statements: []Statement{
					&VariableDeclaration{
						Mutable:      false,
						Name:         "count",
						InferredType: checker.NumType,
						Value:        &NumLiteral{Value: `0`},
					},
					&VariableAssignment{
						Name:     "count",
						Operator: Increment,
						Value: &NumLiteral{
							Value: `2`,
						},
					},
				},
			},
			diagnostics: []checker.Diagnostic{
				{
					Msg: "'count' is not mutable",
				},
			},
		},
		{
			name: "Valid decrement",
			input: `
				mut count = 0
				count =- 2`,
			ast: &Program{
				Statements: []Statement{
					&VariableDeclaration{
						Mutable:      true,
						Name:         "count",
						InferredType: checker.NumType,
						Value:        &NumLiteral{Value: `0`},
					},
					&VariableAssignment{
						Name:     "count",
						Operator: Decrement,
						Value: &NumLiteral{
							Value: `2`,
						},
					},
				},
			},
			diagnostics: []checker.Diagnostic{},
		},
		{
			name: "Invalid decrement",
			input: `
						mut name = "joe"
						name =- 2`,
			ast: &Program{
				Statements: []Statement{
					&VariableDeclaration{
						Mutable:      true,
						Name:         "name",
						InferredType: checker.StrType,
						Value:        &StrLiteral{Value: `"joe"`},
					},
					&VariableAssignment{
						Name:     "name",
						Operator: Decrement,
						Value: &NumLiteral{
							Value: `2`,
						},
					},
				},
			},
			diagnostics: []checker.Diagnostic{
				{
					Msg: "'=-' can only be used with 'Num'",
				},
			},
		},
		{
			name: "Cannot decrement an immutable variable",
			input: `
				let count = 0
				count =- 2`,
			ast: &Program{
				Statements: []Statement{
					&VariableDeclaration{
						Mutable:      false,
						Name:         "count",
						InferredType: checker.NumType,
						Value:        &NumLiteral{Value: `0`},
					},
					&VariableAssignment{
						Name:     "count",
						Operator: Decrement,
						Value: &NumLiteral{
							Value: `2`,
						},
					},
				},
			},
			diagnostics: []checker.Diagnostic{
				{
					Msg: "'count' is not mutable",
				},
			},
		},
	}

	runTests(t, tests)
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

func TestVariableTypeInference(t *testing.T) {
	tests := []test{
		{
			name:  "Inferred type",
			input: `let name = "Alice"`,
			ast: &Program{
				Statements: []Statement{
					&VariableDeclaration{
						Mutable:      false,
						Name:         "name",
						InferredType: checker.StrType,
						Value: &StrLiteral{
							Value: `"Alice"`,
						},
					},
				},
			},
			diagnostics: []checker.Diagnostic{},
		},
		{
			name:  "Str mismatch",
			input: `let name: Str = false`,
			diagnostics: []checker.Diagnostic{
				{
					Msg: "Type mismatch: expected Str, got Bool",
				},
			},
		},
		{
			name:  "Num mismatch",
			input: `let name: Num = "Alice"`,
			diagnostics: []checker.Diagnostic{
				{
					Msg: "Type mismatch: expected Num, got Str",
				},
			},
		},
		{
			name:  "Bool mismatch",
			input: `let is_bool: Bool = "Alice"`,
			diagnostics: []checker.Diagnostic{
				{
					Msg: "Type mismatch: expected Bool, got Str",
				},
			},
		},
	}

	runTests(t, tests)
}

func TestFunctionDeclaration(t *testing.T) {
	tests := []test{
		{
			name:  "Empty function",
			input: `fn empty() {}`,
			ast: &Program{
				Statements: []Statement{
					&FunctionDeclaration{
						Name:       "empty",
						Parameters: []Parameter{},
						ReturnType: checker.VoidType,
						Body:       []Statement{},
					},
				},
			},
			diagnostics: []checker.Diagnostic{},
		},
		{
			name:  "Inferred function return type",
			input: `fn get_msg() { "Hello, world!" }`,
			ast: &Program{
				Statements: []Statement{
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
				},
			},
			diagnostics: []checker.Diagnostic{},
		},
		{
			name:  "Function with a parameter and declared return type",
			input: `fn greet(person: Str) Str { "hello" }`,
			ast: &Program{
				Statements: []Statement{
					&FunctionDeclaration{
						Name: "greet",
						Parameters: []Parameter{
							{
								Name: "person",
							},
						},
						ReturnType: checker.StrType,
						Body: []Statement{
							&StrLiteral{Value: `"hello"`},
						},
					},
				},
			},
		},
		{
			name:  "Function return must match declared return type",
			input: `fn greet(person: Str) Str { }`,
			diagnostics: []checker.Diagnostic{
				{
					Msg: "Type mismatch: expected Str, got Void",
				},
			},
		},
		{
			name:  "Function with two parameters",
			input: `fn add(x: Num, y: Num) Num { 10 }`,
			ast: &Program{
				Statements: []Statement{
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
						Body: []Statement{
							&NumLiteral{Value: "10"},
						},
					},
				},
			},
		},
	}

	runTests(t, tests)
}

func TestUnaryExpressions(t *testing.T) {
	tests := []test{
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
			diagnostics: []checker.Diagnostic{},
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
			diagnostics: []checker.Diagnostic{
				{
					Msg: "The '-' operator can only be used on 'Num'",
				},
			},
		},
	}

	runTests(t, tests)
}

func TestBinaryExpressions(t *testing.T) {
	tests := []test{
		{
			name:  "Valid addition",
			input: `-30 + 20`,
			ast: &Program{
				Statements: []Statement{
					&BinaryExpression{
						Operator: Plus,
						Left: &UnaryExpression{
							Operator: Minus,
							Operand: &NumLiteral{
								Value: `30`,
							},
						},
						Right: &NumLiteral{
							Value: "20",
						},
					},
				},
			},
			diagnostics: []checker.Diagnostic{},
		},
		{
			name:  "Invalid addition",
			input: `30 + "f12"`,
			ast: &Program{
				Statements: []Statement{
					&BinaryExpression{
						Operator: Plus,
						Left: &NumLiteral{
							Value: `30`,
						},
						Right: &StrLiteral{
							Value: `"f12"`,
						},
					},
				},
			},
			diagnostics: []checker.Diagnostic{
				{
					Msg: "The '+' operator can only be used between instances of 'Num'",
				},
			},
		},
		{
			name:  "+ operator is only allowed on Num",
			input: `"foo" + "bar"`,
			ast: &Program{
				Statements: []Statement{
					&BinaryExpression{
						Operator: Plus,
						Left: &StrLiteral{
							Value: `"foo"`,
						},
						Right: &StrLiteral{
							Value: `"bar"`,
						},
					},
				},
			},
			diagnostics: []checker.Diagnostic{
				{
					Msg: "The '+' operator can only be used between instances of 'Num'",
				},
			},
		},
		{
			name:  "Valid subtraction",
			input: `30 - 12`,
			ast: &Program{
				Statements: []Statement{
					&BinaryExpression{
						Operator: Minus,
						Left: &NumLiteral{
							Value: `30`,
						},
						Right: &NumLiteral{
							Value: `12`,
						},
					},
				},
			},
			diagnostics: []checker.Diagnostic{},
		},
		{
			name:  "Invalid subtraction",
			input: `30 - "f12"`,
			ast: &Program{
				Statements: []Statement{
					&BinaryExpression{
						Operator: Minus,
						Left: &NumLiteral{
							Value: `30`,
						},
						Right: &StrLiteral{
							Value: `"f12"`,
						},
					},
				},
			},
			diagnostics: []checker.Diagnostic{
				{
					Msg: "The '-' operator can only be used between instances of 'Num'",
				},
			},
		},
		{
			name:  "Valid division",
			input: `30 / 6`,
			ast: &Program{
				Statements: []Statement{
					&BinaryExpression{
						Operator: Divide,
						Left: &NumLiteral{
							Value: `30`,
						},
						Right: &NumLiteral{
							Value: `6`,
						},
					},
				},
			},
			diagnostics: []checker.Diagnostic{},
		},
		{
			name:  "Invalid division",
			input: `30 / "f12"`,
			ast: &Program{
				Statements: []Statement{
					&BinaryExpression{
						Operator: Divide,
						Left: &NumLiteral{
							Value: `30`,
						},
						Right: &StrLiteral{
							Value: `"f12"`,
						},
					},
				},
			},
			diagnostics: []checker.Diagnostic{
				{
					Msg: "The '/' operator can only be used between instances of 'Num'",
				},
			},
		},
		{
			name:  "Valid multiplication",
			input: `30 * 10`,
			ast: &Program{
				Statements: []Statement{
					&BinaryExpression{
						Operator: Multiply,
						Left: &NumLiteral{
							Value: `30`,
						},
						Right: &NumLiteral{
							Value: `10`,
						},
					},
				},
			},
			diagnostics: []checker.Diagnostic{},
		},
		{
			name:  "Invalid multiplication",
			input: `30 * "f12"`,
			ast: &Program{
				Statements: []Statement{
					&BinaryExpression{
						Operator: Multiply,
						Left: &NumLiteral{
							Value: `30`,
						},
						Right: &StrLiteral{
							Value: `"f12"`,
						},
					},
				},
			},
			diagnostics: []checker.Diagnostic{
				{
					Msg: "The '*' operator can only be used between instances of 'Num'",
				},
			},
		},
		{
			name:  "Valid modulo",
			input: `3 % 9`,
			ast: &Program{
				Statements: []Statement{
					&BinaryExpression{
						Operator: Modulo,
						Left: &NumLiteral{
							Value: `3`,
						},
						Right: &NumLiteral{
							Value: `9`,
						},
					},
				},
			},
			diagnostics: []checker.Diagnostic{},
		},
		{
			name:  "Invalid modulo",
			input: `30 % "f12"`,
			ast: &Program{
				Statements: []Statement{
					&BinaryExpression{
						Operator: Modulo,
						Left: &NumLiteral{
							Value: `30`,
						},
						Right: &StrLiteral{
							Value: `"f12"`,
						},
					},
				},
			},
			diagnostics: []checker.Diagnostic{
				{
					Msg: "The '%' operator can only be used between instances of 'Num'",
				},
			},
		},
		{
			name:  "Valid greater than",
			input: `30 > 12`,
			ast: &Program{
				Statements: []Statement{
					&BinaryExpression{
						Operator: GreaterThan,
						Left: &NumLiteral{
							Value: `30`,
						},
						Right: &NumLiteral{
							Value: `12`,
						},
					},
				},
			},
			diagnostics: []checker.Diagnostic{},
		},
		{
			name:  "Invalid greater than",
			input: `30 > "f12"`,
			ast: &Program{
				Statements: []Statement{
					&BinaryExpression{
						Operator: GreaterThan,
						Left: &NumLiteral{
							Value: `30`,
						},
						Right: &StrLiteral{
							Value: `"f12"`,
						},
					},
				},
			},
			diagnostics: []checker.Diagnostic{
				{
					Msg: "The '>' operator can only be used between instances of 'Num'",
				},
			},
		},
		{
			name:  "Valid greater than or equal",
			input: `30 >= 12`,
			ast: &Program{
				Statements: []Statement{
					&BinaryExpression{
						Operator: GreaterThanOrEqual,
						Left: &NumLiteral{
							Value: `30`,
						},
						Right: &NumLiteral{
							Value: `12`,
						},
					},
				},
			},
			diagnostics: []checker.Diagnostic{},
		},
		{
			name:  "Invalid greater than or equal",
			input: `30 >= "f12"`,
			ast: &Program{
				Statements: []Statement{
					&BinaryExpression{
						Operator: GreaterThanOrEqual,
						Left: &NumLiteral{
							Value: `30`,
						},
						Right: &StrLiteral{
							Value: `"f12"`,
						},
					},
				},
			},
			diagnostics: []checker.Diagnostic{
				{
					Msg: "The '>=' operator can only be used between instances of 'Num'",
				},
			},
		},
		{
			name:  "Valid less than",
			input: `30 < 12`,
			ast: &Program{
				Statements: []Statement{
					&BinaryExpression{
						Operator: LessThan,
						Left: &NumLiteral{
							Value: `30`,
						},
						Right: &NumLiteral{
							Value: `12`,
						},
					},
				},
			},
			diagnostics: []checker.Diagnostic{},
		},
		{
			name:  "Invalid les than",
			input: `30 < "f12"`,
			ast: &Program{
				Statements: []Statement{
					&BinaryExpression{
						Operator: LessThan,
						Left: &NumLiteral{
							Value: `30`,
						},
						Right: &StrLiteral{
							Value: `"f12"`,
						},
					},
				},
			},
			diagnostics: []checker.Diagnostic{
				{
					Msg: "The '<' operator can only be used between instances of 'Num'",
				},
			},
		},
		{
			name:  "Valid less than or equal",
			input: `30 <= 12`,
			ast: &Program{
				Statements: []Statement{
					&BinaryExpression{
						Operator: LessThanOrEqual,
						Left: &NumLiteral{
							Value: `30`,
						},
						Right: &NumLiteral{
							Value: `12`,
						},
					},
				},
			},
			diagnostics: []checker.Diagnostic{},
		},
		{
			name:  "Invalid less than or equal",
			input: `30 <= true`,
			ast: &Program{
				Statements: []Statement{
					&BinaryExpression{
						Operator: LessThanOrEqual,
						Left: &NumLiteral{
							Value: `30`,
						},
						Right: &BoolLiteral{
							Value: true,
						},
					},
				},
			},
			diagnostics: []checker.Diagnostic{
				{
					Msg: "The '<=' operator can only be used between instances of 'Num'",
				},
			},
		},
		{
			name:  "Valid string equality checks",
			input: `"Joe" == "Joe"`,
			ast: &Program{
				Statements: []Statement{
					&BinaryExpression{
						Operator: Equal,
						Left: &StrLiteral{
							Value: `"Joe"`,
						},
						Right: &StrLiteral{
							Value: `"Joe"`,
						},
					},
				},
			},
			diagnostics: []checker.Diagnostic{},
		},
		{
			name:  "Invalid string equality check",
			input: `"Joe" == true`,
			ast: &Program{
				Statements: []Statement{
					&BinaryExpression{
						Operator: Equal,
						Left: &StrLiteral{
							Value: `"Joe"`,
						},
						Right: &BoolLiteral{
							Value: true,
						},
					},
				},
			},
			diagnostics: []checker.Diagnostic{
				{
					Msg: "The '==' operator can only be used between instances of 'Num', 'Str', or 'Bool'",
				},
			},
		},
		{
			name:  "Valid number equality checks",
			input: `1 == 1`,
			ast: &Program{
				Statements: []Statement{
					&BinaryExpression{
						Operator: Equal,
						Left: &NumLiteral{
							Value: `1`,
						},
						Right: &NumLiteral{
							Value: `1`,
						},
					},
				},
			},
			diagnostics: []checker.Diagnostic{},
		},
		{
			name:  "Invalid number equality checks",
			input: `1 == "eleventy"`,
			ast: &Program{
				Statements: []Statement{
					&BinaryExpression{
						Operator: Equal,
						Left: &NumLiteral{
							Value: `1`,
						},
						Right: &StrLiteral{
							Value: `"eleventy"`,
						},
					},
				},
			},
			diagnostics: []checker.Diagnostic{
				{
					Msg: "The '==' operator can only be used between instances of 'Num', 'Str', or 'Bool'",
				},
			},
		},
		{
			name:  "Valid boolean equality checks",
			input: `true == false`,
			ast: &Program{
				Statements: []Statement{
					&BinaryExpression{
						Operator: Equal,
						Left: &BoolLiteral{
							Value: true,
						},
						Right: &BoolLiteral{
							Value: false,
						},
					},
				},
			},
			diagnostics: []checker.Diagnostic{},
		},
		{
			name:  "Invalid boolean equality checks",
			input: `true == "eleventy"`,
			ast: &Program{
				Statements: []Statement{
					&BinaryExpression{
						Operator: Equal,
						Left: &BoolLiteral{
							Value: true,
						},
						Right: &StrLiteral{
							Value: `"eleventy"`,
						},
					},
				},
			},
			diagnostics: []checker.Diagnostic{
				{
					Msg: "The '==' operator can only be used between instances of 'Num', 'Str', or 'Bool'",
				},
			},
		},

		// Test cases for the '!=' operator
		{
			name:  "Valid string inequality checks",
			input: `"Joe" != "Joe"`,
			ast: &Program{
				Statements: []Statement{
					&BinaryExpression{
						Operator: NotEqual,
						Left: &StrLiteral{
							Value: `"Joe"`,
						},
						Right: &StrLiteral{
							Value: `"Joe"`,
						},
					},
				},
			},
			diagnostics: []checker.Diagnostic{},
		},
		{
			name:  "Invalid string inequality check",
			input: `"Joe" != true`,
			ast: &Program{
				Statements: []Statement{
					&BinaryExpression{
						Operator: NotEqual,
						Left: &StrLiteral{
							Value: `"Joe"`,
						},
						Right: &BoolLiteral{
							Value: true,
						},
					},
				},
			},
			diagnostics: []checker.Diagnostic{
				{
					Msg: "The '!=' operator can only be used between instances of 'Num', 'Str', or 'Bool'",
				},
			},
		},
		{
			name:  "Valid number inequality checks",
			input: `1 != 1`,
			ast: &Program{
				Statements: []Statement{
					&BinaryExpression{
						Operator: NotEqual,
						Left: &NumLiteral{
							Value: `1`,
						},
						Right: &NumLiteral{
							Value: `1`,
						},
					},
				},
			},
			diagnostics: []checker.Diagnostic{},
		},
		{
			name:  "Invalid number inequality checks",
			input: `1 != "eleventy"`,
			ast: &Program{
				Statements: []Statement{
					&BinaryExpression{
						Operator: NotEqual,
						Left: &NumLiteral{
							Value: `1`,
						},
						Right: &StrLiteral{
							Value: `"eleventy"`,
						},
					},
				},
			},
			diagnostics: []checker.Diagnostic{
				{
					Msg: "The '!=' operator can only be used between instances of 'Num', 'Str', or 'Bool'",
				},
			},
		},
		{
			name:  "Valid boolean inequality checks",
			input: `true != false`,
			ast: &Program{
				Statements: []Statement{
					&BinaryExpression{
						Operator: NotEqual,
						Left: &BoolLiteral{
							Value: true,
						},
						Right: &BoolLiteral{
							Value: false,
						},
					},
				},
			},
			diagnostics: []checker.Diagnostic{},
		},
		{
			name:  "Invalid boolean inequality checks",
			input: `true != "eleventy"`,
			ast: &Program{
				Statements: []Statement{
					&BinaryExpression{
						Operator: NotEqual,
						Left: &BoolLiteral{
							Value: true,
						},
						Right: &StrLiteral{
							Value: `"eleventy"`,
						},
					},
				},
			},
			diagnostics: []checker.Diagnostic{
				{
					Msg: "The '!=' operator can only be used between instances of 'Num', 'Str', or 'Bool'",
				},
			},
		},

		// logic operator checks
		{
			name:  "Valid use of 'and' operator",
			input: `true and false`,
			ast: &Program{
				Statements: []Statement{
					&BinaryExpression{
						Operator: And,
						Left: &BoolLiteral{
							Value: true,
						},
						Right: &BoolLiteral{
							Value: false,
						},
					},
				},
			},
			diagnostics: []checker.Diagnostic{},
		},
		{
			name:  "Ivalid use of 'and' operator",
			input: `true and "eleventy"`,
			ast: &Program{
				Statements: []Statement{
					&BinaryExpression{
						Operator: And,
						Left: &BoolLiteral{
							Value: true,
						},
						Right: &StrLiteral{
							Value: `"eleventy"`,
						},
					},
				},
			},
			diagnostics: []checker.Diagnostic{
				{
					Msg: "The 'and' operator can only be used between instances of 'Bool'",
				},
			},
		},
		{
			name:  "Valid use of 'or' operator",
			input: `true or false`,
			ast: &Program{
				Statements: []Statement{
					&BinaryExpression{
						Operator: Or,
						Left: &BoolLiteral{
							Value: true,
						},
						Right: &BoolLiteral{
							Value: false,
						},
					},
				},
			},
			diagnostics: []checker.Diagnostic{},
		},
		{
			name:  "Ivalid use of 'or' operator",
			input: `true or "eleventy"`,
			ast: &Program{
				Statements: []Statement{
					&BinaryExpression{
						Operator: Or,
						Left: &BoolLiteral{
							Value: true,
						},
						Right: &StrLiteral{
							Value: `"eleventy"`,
						},
					},
				},
			},
			diagnostics: []checker.Diagnostic{
				{
					Msg: "The 'or' operator can only be used between instances of 'Bool'",
				},
			},
		},

		// range operator
		{
			name:  "Valid use of range operator",
			input: "1...10",
			ast: &Program{
				Statements: []Statement{
					&BinaryExpression{
						Operator: Range,
						Left: &NumLiteral{
							Value: `1`,
						},
						Right: &NumLiteral{
							Value: `10`,
						},
					},
				},
			},
			diagnostics: []checker.Diagnostic{},
		},
		{
			name:  "Invalid use of range operator",
			input: `"fizz"...10`,
			ast: &Program{
				Statements: []Statement{
					&BinaryExpression{
						Operator: Range,
						Left: &StrLiteral{
							Value: `"fizz"`,
						},
						Right: &NumLiteral{
							Value: `10`,
						},
					},
				},
			},
			diagnostics: []checker.Diagnostic{{
				Msg: "A range must be between two Num",
			}},
		},
	}

	runTests(t, tests)
}
