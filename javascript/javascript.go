package javascript

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/akonwi/kon/ast"
)

type jsGenerator struct {
	builder     strings.Builder
	indentLevel int
}

func (g *jsGenerator) indent() {
	g.indentLevel++
}

func (g *jsGenerator) dedent() {
	if g.indentLevel > 0 {
		g.indentLevel--
	}
}

func (g *jsGenerator) writeIndent() {
	g.builder.WriteString(strings.Repeat("  ", g.indentLevel))
}

func (g *jsGenerator) write(format string, args ...interface{}) {
	g.builder.WriteString(fmt.Sprintf(format, args...))
}

func (g *jsGenerator) writeLine(line string, args ...interface{}) {
	g.writeIndent()
	g.builder.WriteString(fmt.Sprintf(line, args...))
	g.builder.WriteString("\n")
}

func (g *jsGenerator) generateVariableDeclaration(decl ast.VariableDeclaration) {
	g.writeIndent()
	if decl.Mutable {
		g.write("let ")
	} else {
		g.write("const ")
	}

	g.write("%s = ", decl.Name)
	g.generateExpression(decl.Value)
	g.write("\n")
}

func (g *jsGenerator) generateFunctionDeclaration(decl ast.FunctionDeclaration) {
	g.writeIndent()
	g.write("function %s", decl.Name)
	g.write("(")
	for i, param := range decl.Parameters {
		if i > 0 {
			g.write(", ")
		}
		g.write(param.Name)
	}
	g.write(") ")

	if len(decl.Body) == 0 {
		g.write("{}\n")
	} else {
		g.writeLine("{")
		g.indent()
		for i, statement := range decl.Body {
			if i == len(decl.Body)-1 {
				if expr, ok := statement.(ast.Expression); ok {
					g.writeIndent()
					g.write("return ")
					g.generateExpression(expr)
					g.write("\n")
					continue
				}
			} else {
				g.generateStatement(statement)
			}
		}
		g.dedent()
		g.writeLine("}")
	}
}

func (g *jsGenerator) generateAnonymousFunction(decl ast.AnonymousFunction) {
	g.write("(")
	for i, param := range decl.Parameters {
		if i > 0 {
			g.write(", ")
		}
		g.write(param.Name)
	}
	g.write(") => {")

	if len(decl.Body) == 0 {
		g.write("}")
		return
	}

	g.write("\n")
	g.indent()
	for i, statement := range decl.Body {
		if i == len(decl.Body)-1 {
			if expr, ok := statement.(ast.Expression); ok {
				g.writeIndent()
				g.write("return ")
				g.generateExpression(expr)
				g.write("\n")
				continue
			}
		} else {
			g.generateStatement(statement)
		}
	}
	g.dedent()
	g.write("}")
}

func resolveOperator(operator ast.Operator) string {
	switch operator {
	case ast.Assign:
		return "="
	case ast.Equal:
		return "==="
	case ast.NotEqual:
		return "!=="
	case ast.Increment:
		return "+="
	case ast.Decrement:
		return "-="
	case ast.Multiply:
		return "*"
	case ast.Divide:
		return "/"
	case ast.Plus:
		return "+"
	case ast.Minus:
		return "-"
	case ast.Modulo:
		return "%"
	case ast.Or:
		return "||"
	case ast.And:
		return "&&"
	case ast.LessThan:
		return "<"
	case ast.LessThanOrEqual:
		return "<="
	case ast.GreaterThan:
		return ">"
	case ast.GreaterThanOrEqual:
		return ">="
	case ast.Bang:
		return "!"
	default:
		panic(fmt.Errorf("Unresolved operator: %v", operator))
	}
}

func (g *jsGenerator) generateVariableAssignment(assignment ast.VariableAssignment) {
	g.write("%s %s ", assignment.Name, resolveOperator(assignment.Operator))
	g.generateExpression(assignment.Value)
	g.write("\n")
}

func (g *jsGenerator) generateStatement(statement ast.Statement) {
	switch statement.(type) {
	case ast.StructDefinition: // skipped
	case ast.VariableDeclaration:
		g.generateVariableDeclaration(statement.(ast.VariableDeclaration))
	case ast.VariableAssignment:
		g.generateVariableAssignment(statement.(ast.VariableAssignment))
	case ast.FunctionDeclaration:
		g.generateFunctionDeclaration(statement.(ast.FunctionDeclaration))
	default:
		{
			if expr, ok := statement.(ast.Expression); ok {
				g.writeIndent()
				g.generateExpression(expr)
			} else {
				panic(fmt.Errorf("Unhandled statement node: [%s] - %s\n", reflect.TypeOf(statement), statement))
			}
		}
	}
}

func (g *jsGenerator) generateStructInstance(instance ast.StructInstance) {
	g.write("{")
	if len(instance.Properties) > 0 {
		i := 0
		g.write(" ")
		for key, value := range instance.Properties {
			if i > 0 {
				g.write(", ")
			} else {
				i++
			}
			g.write("%s: ", key)
			g.generateExpression(value)
		}
		g.write(" ")
	}
	g.write("}")
}

func (g *jsGenerator) generateExpression(expr ast.Expression) {
	switch expr.(type) {
	case ast.InterpolatedStr:
		g.write("`")
		for _, chunk := range expr.(ast.InterpolatedStr).Chunks {
			if _, ok := chunk.(ast.StrLiteral); ok {
				g.write(chunk.(ast.StrLiteral).Value)
			} else {
				g.write("${")
				g.generateExpression(chunk)
				g.write("}")
			}
		}
		g.write("`")
	case ast.StrLiteral:
		g.write(expr.(ast.StrLiteral).Value)
	case ast.NumLiteral:
		g.write(expr.(ast.NumLiteral).Value)
	case ast.BoolLiteral:
		g.write("%v", expr.(ast.BoolLiteral).Value)
	case ast.ListLiteral:
		g.write("[")
		for i, item := range expr.(ast.ListLiteral).Items {
			if i > 0 {
				g.write(", ")
			}
			g.generateExpression(item)
		}
		g.write("]")
	case ast.MapLiteral:
		g.write("new Map([")
		i := 0
		for key, value := range expr.(ast.MapLiteral).Entries {
			if i > 0 {
				g.write(", ")
			} else {
				i++
			}
			g.write("[")
			g.write(`%s, `, key)
			g.generateExpression(value)
			g.write("]")
		}
		g.write("])")
	case ast.Identifier:
		g.write("%s", expr.(ast.Identifier).Name)
	case ast.BinaryExpression:
		binary := expr.(ast.BinaryExpression)
		if binary.HasPrecedence {
			g.write("(")
		}
		g.generateExpression(binary.Left)
		g.write(" %s ", resolveOperator(binary.Operator))
		g.generateExpression(binary.Right)
		if binary.HasPrecedence {
			g.write(")")
		}
	case ast.UnaryExpression:
		unary := expr.(ast.UnaryExpression)
		g.write("%s", resolveOperator(unary.Operator))
		g.generateExpression(unary.Operand)
	case ast.AnonymousFunction:
		g.generateAnonymousFunction(expr.(ast.AnonymousFunction))
	case ast.StructInstance:
		g.generateStructInstance(expr.(ast.StructInstance))
	default:
		panic(fmt.Errorf("Unhandled expression node: [%s] - %s\n", reflect.TypeOf(expr), expr))
	}
}

func GenerateJS(program ast.Program) string {
	generator := jsGenerator{
		builder:     strings.Builder{},
		indentLevel: 0,
	}

	for _, statement := range program.Statements {
		generator.generateStatement(statement)
	}

	return generator.builder.String()
}
