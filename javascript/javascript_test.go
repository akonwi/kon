package javascript

import (
	"fmt"
	"strings"
	"testing"

	"github.com/akonwi/kon/ast"
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

func assertEquality(t *testing.T, got, want string) {
	t.Helper()
	diff := cmp.Diff(want, got)
	if diff != "" {
		t.Errorf("Generated code does not match (-want +got):\n%s", diff)
	}
}

type test struct {
	name, input, output string
}

func runTests(t *testing.T, tests []test) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := treeSitterParser.Parse([]byte(tt.input), nil)
			parser := ast.NewParser([]byte(tt.input), tree)
			ast, err := parser.Parse()
			if err != nil {
				t.Fatal(fmt.Errorf("Error parsing tree: %v", err))
			}

			js := GenerateJS(ast)

			if diff := cmp.Diff(tt.output, js, cmp.Transformer("SpaceRemover", strings.TrimSpace)); diff != "" {
				t.Errorf("Generated javascript does not match (-want +got):\n%s", diff)
			}
		})
	}
}

func TestLiteralExpressions(t *testing.T) {
	tests := []test{
		{
			name: "identifier",
			input: `
let x = 42
let y = x`,
			output: `
const x = 42
const y = x`,
		},
		{
			name:   "raw string",
			input:  `"foobar"`,
			output: `"foobar"`,
		},
		{
			name:   "interpolated string",
			input:  `"foobar {{ 42 }}"`,
			output: "`foobar ${42}`",
		},
		{
			name:   "number",
			input:  `42`,
			output: `42`,
		},
		{
			name:   "booleans",
			input:  `false`,
			output: `false`,
		},
		{
			name:   "list literal",
			input:  `[1, 2, 3]`,
			output: `[1, 2, 3]`,
		},
		{
			name:   "map literal",
			input:  `["jane": 1, "joe": 2]`,
			output: `new Map([["jane", 1], ["joe", 2]])`,
		},
	}

	runTests(t, tests)
}

func TestBinaryExpressions(t *testing.T) {
	tests := []test{
		{
			name:   "addition",
			input:  `42 + 20`,
			output: `42 + 20`,
		},
		{
			name:   "subtraction",
			input:  `42 - 20`,
			output: `42 - 20`,
		},
		{
			name:   "multiplication",
			input:  `42 * 20`,
			output: `42 * 20`,
		},
		{
			name:   "division",
			input:  `42 / 2`,
			output: `42 / 2`,
		},
		{
			name:   "modulo",
			input:  `42 % 2`,
			output: `42 % 2`,
		},
		{
			name:   "math with precedence",
			input:  `(70 - 32) * 5 / 9`,
			output: `(70 - 32) * 5 / 9`,
		},
		{
			name:   "equality",
			input:  `42 == 42`,
			output: `42 === 42`,
		},
		{
			name:   "inequality",
			input:  `42 != 42`,
			output: `42 !== 42`,
		},
		{
			name:   "logical or",
			input:  `true or false`,
			output: `true || false`,
		},
		{
			name:   "logical and",
			input:  `true and false`,
			output: `true && false`,
		},
		{
			name:   "less than",
			input:  `20 < 100`,
			output: `20 < 100`,
		},
		{
			name:   "less than or equal",
			input:  `20 <= 100`,
			output: `20 <= 100`,
		},
		{
			name:   "greater than",
			input:  `20 > 100`,
			output: `20 > 100`,
		},
		{
			name:   "greater than or equal",
			input:  `20 >= 100`,
			output: `20 >= 100`,
		},
	}

	runTests(t, tests)
}

func TestUnaryExpressions(t *testing.T) {
	tests := []test{
		{
			name:   "numeric negation",
			input:  `-42`,
			output: `-42`,
		},
		{
			name:   "boolean negation",
			input:  `!true`,
			output: `!true`,
		},
	}

	runTests(t, tests)
}

func TestFunctionDeclaration(t *testing.T) {
	tests := []test{
		{
			name:   "noop",
			input:  `fn noop() {}`,
			output: "function noop() {}",
		},
		{
			name:   "with parameters",
			input:  `fn add(x: Num, y: Num) {}`,
			output: "function add(x, y) {}",
		},
		{
			name:   "with return type",
			input:  `fn add(x: Num, y: Num) Num {}`,
			output: "function add(x, y) {}",
		},
		{
			name:  "single statement body: return is implicit",
			input: `fn add(x: Num, y: Num) Num { x + y }`,
			output: `
function add(x, y) {
  return x + y
}`,
		},
		{
			name: "the last statement is the return statement",
			input: `
fn add(x: Num, y: Num) Num {
  let result = x + y
	result
}`,
			output: `
function add(x, y) {
  const result = x + y
  return result
}`,
		},
	}

	runTests(t, tests)
}

func TestAnonymousFunctions(t *testing.T) {
	tests := []test{
		{
			name:   "noop",
			input:  `() {}`,
			output: `() => {}`,
		},
		{
			name:  "with parameters and body",
			input: `(one, two) { one / two }`,
			output: `
(one, two) => {
  return one / two
}`,
		},
	}

	runTests(t, tests)
}

func TestVariableDeclaration(t *testing.T) {
	tests := []test{
		{
			name:   "mutable string",
			input:  `mut explicit: Str = "Alice"`,
			output: `let explicit = "Alice"`,
		},
		{
			name:   "immutable string",
			input:  `let explicit = "Alice"`,
			output: `const explicit = "Alice"`,
		},
		{
			name:   "mutable number",
			input:  `mut power = 200`,
			output: `let power = 200`,
		},
		{
			name:   "immutable number",
			input:  `let power = 200`,
			output: `const power = 200`,
		},
		{
			name:   "mutable boolean",
			input:  `mut is_valid = true`,
			output: `let is_valid = true`,
		},
		{
			name:   "immutable boolean",
			input:  `let is_valid = false`,
			output: `const is_valid = false`,
		},
	}

	runTests(t, tests)
}

func TestVariableAssignment(t *testing.T) {
	runTests(t, []test{
		{
			name: "string assignment",
			input: `
mut name = "Alice"
name = "Bob"`,
			output: `
let name = "Alice"
name = "Bob"`,
		},
		{
			name: "compound assignment",
			input: `
mut x = 10
x =+ 5
x =- 5`,
			output: `
let x = 10
x += 5
x -= 5`,
		},
	})
}

func TestStructs(t *testing.T) {
	runTests(t, []test{
		{
			name: "empty struct",
			input: `
struct Foo {}
let a_foo = Foo{}`,
			output: `
const a_foo = {}`,
		},
		{
			name: "full struct",
			input: `
struct Person { name: Str, age: Num, employed: Bool }
Person{ name: "Joe", age: 42, employed: true }`,
			output: `
{ name: "Joe", age: 42, employed: true }`,
		},
	})
}

func TestEnums(t *testing.T) {
	runTests(t, []test{
		{
			name:  "basic enum",
			input: `enum Color { Red, Green, Yellow }`,
			output: `
const Color = Object.freeze({
  Red: Object.freeze({ index: 0 }),
  Green: Object.freeze({ index: 1 }),
  Yellow: Object.freeze({ index: 2 })
})`,
		},
	})
}

func TestWhileLoops(t *testing.T) {
	runTests(t, []test{
		{
			name:  "basic while loop",
			input: `while true { 20 }`,
			output: `
while (true) {
  20
}`,
		},
	})
}

func TestForLoops(t *testing.T) {
	runTests(t, []test{
		{
			name:  "looping over a number",
			input: `for i in 10 { i }`,
			output: `
for (let i = 0; i < 10; i++) {
  i
}`,
		},
		{
			name:  "looping over a range",
			input: `for num in 0..10 { num }`,
			output: `
for (let num = 0; num < 10; num++) {
  num
}`,
		},
		{
			name:  "looping over an array",
			input: `for name in ["jane", "joe"] { name }`,
			output: `
for (const name of ["jane", "joe"]) {
  name
}`,
		},
		{
			name: "looping over a string",
			input: `
let msg = "hello world"
for char in msg { char }`,
			output: `
const msg = "hello world"
for (const char of msg) {
  char
}`,
		},
	})
}
