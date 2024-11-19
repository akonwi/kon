package ast

import (
	"fmt"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

type BaseNode struct {
	TSNode *tree_sitter.Node
}

func (b *BaseNode) GetTSNode() *tree_sitter.Node {
	return b.TSNode
}

type Node interface {
	String() string
	GetTSNode() *tree_sitter.Node
}

type Program struct {
	BaseNode
	Statements []Statement
}

type Statement interface {
	Node
	StatementNode()
}

type Expression interface {
	Node
	ExpressionNode()
}

type VariableDeclaration struct {
	BaseNode
	Name    string
	Mutable bool
	Value   Expression
}

func (v *VariableDeclaration) StatementNode() {}
func (v *VariableDeclaration) String() string {
	return fmt.Sprintf("TODO")
}

type StrLiteral struct {
	BaseNode
	Value string
}

func (s *StrLiteral) ExpressionNode() {}
func (s *StrLiteral) String() string {
	return s.Value
}

type NumLiteral struct {
	BaseNode
	Value string
}

func (n *NumLiteral) ExpressionNode() {}
func (n *NumLiteral) String() string {
	return n.Value
}

type BoolLiteral struct {
	BaseNode
	Value bool
}

func (b *BoolLiteral) ExpressionNode() {}
func (b *BoolLiteral) String() string {
	return fmt.Sprintf("%t", b.Value)
}

type Parser struct {
	sourceCode []byte
}

func NewParser(sourceCode []byte) *Parser {
	return &Parser{sourceCode: sourceCode}
}

func (p *Parser) text(node *tree_sitter.Node) string {
	return string(p.sourceCode[node.StartByte():node.EndByte()])
}

func (p *Parser) Parse(tree *tree_sitter.Tree) (*Program, error) {
	rootNode := tree.RootNode()
	program := &Program{
		BaseNode:   BaseNode{TSNode: rootNode},
		Statements: []Statement{}}

	for i := range rootNode.NamedChildCount() {
		stmt, err := p.parseStatement(rootNode.NamedChild(i))
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
	}

	return program, nil
}

func (p *Parser) parseStatement(node *tree_sitter.Node) (Statement, error) {
	child := node.NamedChild(0)
	if child == nil {
		println("statement child is nil")
		return nil, fmt.Errorf("statement child is nil")
	}
	switch child.GrammarName() {
	case "variable_definition":
		return p.parseVariableDecl(child)
	default:
		return nil, fmt.Errorf("Unhandled statement: %s", child.GrammarName())
	}
}

func (p *Parser) parseVariableDecl(node *tree_sitter.Node) (*VariableDeclaration, error) {
	isMutable := p.text(node.NamedChild(0)) == "mut"
	name := p.text(node.NamedChild(1))
	value, err := p.parseExpression(node.ChildByFieldName("value"))

	if err != nil {
		return nil, err
	}
	return &VariableDeclaration{
		BaseNode: BaseNode{TSNode: node},
		Mutable:  isMutable,
		Name:     name,
		Value:    value,
	}, nil
}

func (p *Parser) parseExpression(node *tree_sitter.Node) (Expression, error) {
	child := node.Child(0)
	switch child.GrammarName() {
	case "primitive_value":
		return p.parsePrimitiveValue(child)
	default:
		return nil, fmt.Errorf("Unhandled expression: %s", child.GrammarName())
	}
}

func (p *Parser) parsePrimitiveValue(node *tree_sitter.Node) (Expression, error) {
	child := node.Child(0)
	switch child.GrammarName() {
	case "string":
		return &StrLiteral{
			BaseNode: BaseNode{TSNode: node},
			Value:    p.text(child)}, nil
	case "number":
		return &NumLiteral{
			BaseNode: BaseNode{TSNode: node},
			Value:    p.text(child)}, nil
	case "boolean":
		return &BoolLiteral{
			BaseNode: BaseNode{TSNode: node},
			Value:    p.text(child) == "true"}, nil
	default:
		return nil, fmt.Errorf("Unhandled expression: %s", child.GrammarName())
	}
}
