import type { Token } from "./lexer/lexer";

export type Literal = { type: "Literal"; token: Token; value: any };
export type Variable = { type: "Variable"; token: Token };
export type Grouping = { type: "Grouping"; expression: Expr };
export type Unary = {
	type: "Unary";
	operator: Token;
	right: Expr;
};
export type Binary = {
	type: "Binary";
	left: Expr;
	operator: Token;
	right: Expr;
};

/*
 although assigning should technically be a statement,
 i'm treating it as an expression for a few reasons:
  * for precedence
  * it's right associative like most expressions
*/
export type Assign = { type: "Assign"; name: Token; value: Expr };
export type Increment = { type: "Increment"; name: Token; value: Expr };
export type Decrement = { type: "Decrement"; name: Token; value: Expr };
export type Logical = {
	type: "Logical";
	left: Expr;
	operator: Token;
	right: Expr;
};
export type RangeExpr = {
	type: "RangeExpr";
	start: Token;
	end: Token;
};

export type Expr =
	| Literal
	| Grouping
	| Unary
	| Binary
	| Variable
	| Assign
	| Increment
	| Decrement
	| Logical
	| RangeExpr;

export type Print = { type: "Print"; expression: Expr };
export type ExprStmt = { type: "ExprStatement"; expression: Expr };
export type MutDecl = {
	type: "MutDecl";
	name: Token;
	_staticType: Token | null;
	initializer: Expr;
};
export type LetDecl = {
	type: "LetDecl";
	name: Token;
	_staticType: Token | null;
	initializer: Expr;
};
export type Block = { type: "Block"; statements: Stmt[] };
export type If = {
	type: "If";
	condition: Expr;
	thenBranch: Stmt;
	elseBranch: Stmt | null;
};
export type While = {
	type: "While";
	condition: Expr;
	body: Stmt;
};
export type ForIn = {
	type: "ForIn";
	cursor: Token;
	range: RangeExpr;
	body: Stmt;
};
export type Stmt =
	| Print
	| ExprStmt
	| MutDecl
	| LetDecl
	| Block
	| If
	| While
	| ForIn;
// | { type: "Call"; callee: Expr; paren: Token; arguments: Expr[] };

// | { type: "Let"; name: Token; initializer: Expr | null }
// | { type: "Mut"; name: Token; initializer: Expr | null }
// | { type: "If"; condition: Expr; thenBranch: Stmt; elseBranch: Stmt | null }
// | { type: "While"; condition: Expr; body: Stmt }
// | { type: "Function"; name: Token; params: Token[]; body: Stmt[] }
// | { type: "Return"; keyword: Token; value: Expr | null };
