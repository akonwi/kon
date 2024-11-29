package ast

import (
	"testing"

	checker "github.com/akonwi/kon/checker"
)

func TestFunctionDeclaration(t *testing.T) {
	tests := []test{
		{
			name:  "Empty function",
			input: `fn empty() {}`,
			output: &Program{
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
			output: &Program{
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
			output: &Program{
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
			output: &Program{
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

func testFunctionCalls(t *testing.T) {
	get_name := checker.FunctionType{Mutates: false, Parameters: []checker.Type{}, ReturnType: checker.StrType}
	greet := checker.FunctionType{
		Mutates:    false,
		Parameters: []checker.Type{checker.StrType},
		ReturnType: checker.StrType,
	}
	tests := []test{
		{
			name: "Valid function call with no arguments",
			input: `
				fn get_name() Str { "name" }
				get_name()`,
			output: &Program{
				Statements: []Statement{
					&FunctionDeclaration{
						Name:       "get_name",
						Parameters: []Parameter{},
						ReturnType: get_name.ReturnType,
						Body:       []Statement{&StrLiteral{Value: `"name"`}},
					},
					FunctionCall{
						Name: "get_name",
						Args: []Expression{},
						Type: get_name,
					},
				},
			},
			diagnostics: []checker.Diagnostic{},
		},
		{
			name: "Providing arguments when none are expected",
			input: `
				fn get_name() Str { "name" }
				get_name("bo")
			`,
			diagnostics: []checker.Diagnostic{
				{Msg: "Expected 0 arguments, got 1"},
			},
		},
		{
			name: "Valid function call with one argument",
			input: `
				fn greet(name: Str) Str { "hello" }
				greet("Alice")`,
			output: &Program{
				Statements: []Statement{
					&FunctionDeclaration{
						Name: "greet",
						Parameters: []Parameter{
							{Name: "name"},
						},
						ReturnType: greet.ReturnType,
						Body:       []Statement{&StrLiteral{Value: `"hello"`}},
					},
					FunctionCall{
						Name: "greet",
						Args: []Expression{
							&StrLiteral{Value: `"Alice"`},
						},
						Type: greet,
					},
				},
			},
			diagnostics: []checker.Diagnostic{},
		},
		// {
		// 	name: "Valid function call with two arguments",
		// 	input: `
		// 		fn add(x: Num, y: Num) Num { 0 }
		// 		add(1, 2)`,
		// 	output: &Program{
		// 		Statements: []Statement{
		// 			&FunctionCall{
		// 				Name: "add",
		// 				Arguments: []Expression{
		// 					&NumLiteral{Value: "1"},
		// 					&NumLiteral{Value: "2"},
		// 				},
		// 			},
		// 		},
		// 	},
		// 	diagnostics: []checker.Diagnostic{},
		// },
		// {
		// 	name: "Wrong number of arguments",
		// 	input: `
		// 		fn greet(name: Str) Str { "hello" }
		// 		greet("Alice", "Bob")`,
		// 	output: &Program{
		// 		Statements: []Statement{
		// 			&FunctionCall{
		// 				Name: "greet",
		// 				Arguments: []Expression{
		// 					&StrLiteral{Value: `"Alice"`},
		// 					&StrLiteral{Value: `"Bob"`},
		// 				},
		// 			},
		// 		},
		// 	},
		// 	diagnostics: []checker.Diagnostic{
		// 		{
		// 			Msg: "Function 'greet' expects 1 argument(s), got 2",
		// 		},
		// 	},
		// },
		// {
		// 	name: "Wrong argument type",
		// 	input: `
		// 		fn add(x: Num, y: Num) Num { 0 }
		// 		add("one", 2)`,
		// 	output: &Program{
		// 		Statements: []Statement{
		// 			&FunctionCall{
		// 				Name: "add",
		// 				Arguments: []Expression{
		// 					&StrLiteral{Value: `"one"`},
		// 					&NumLiteral{Value: "2"},
		// 				},
		// 			},
		// 		},
		// 	},
		// 	diagnostics: []checker.Diagnostic{
		// 		{
		// 			Msg: "Function 'add' expects argument of type 'Num', got 'Str'",
		// 		},
		// 	},
		// },
	}

	runTests(t, tests)
}