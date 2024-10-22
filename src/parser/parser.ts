import type { Expr, Stmt } from "../ast";
import { TokenType, type Token } from "../lexer/lexer";

class ParseError extends Error {}

export class Parser {
	private current = 0;

	constructor(private tokens: Token[]) {}

	parse(): Stmt[] {
		const statements: Stmt[] = [];
		while (!this.isAtEnd()) {
			statements.push(this.declaration());
		}
		return statements;
	}

	private expression(): Expr {
		return this.equality();
	}

	private equality(): Expr {
		let expr = this.comparison();
		while (this.match(TokenType.NOT_EQUAL, TokenType.EQUAL)) {
			const operator = this.previous();
			const right = this.comparison();
			expr = { type: "Binary", left: expr, operator, right };
		}
		return expr;
	}

	private comparison(): Expr {
		let expr = this.term();
		while (
			this.match(
				TokenType.GREATER_THAN,
				TokenType.GREATER_EQUAL,
				TokenType.LESS_THAN,
				TokenType.LESS_EQUAL,
			)
		) {
			const operator = this.previous();
			const right = this.term();
			expr = { type: "Binary", left: expr, operator, right };
		}
		return expr;
	}

	private term(): Expr {
		let expr = this.factor();
		while (this.match(TokenType.MINUS, TokenType.PLUS)) {
			const operator = this.previous();
			const right = this.factor();
			expr = { type: "Binary", left: expr, operator, right };
		}
		return expr;
	}

	private factor(): Expr {
		let expr = this.unary();
		while (this.match(TokenType.SLASH, TokenType.STAR)) {
			const operator = this.previous();
			const right = this.unary();
			expr = { type: "Binary", left: expr, operator, right };
		}
		return expr;
	}

	private unary(): Expr {
		if (this.match(TokenType.BANG, TokenType.MINUS)) {
			const operator = this.previous();
			const right = this.unary();
			return { type: "Unary", operator, right };
		}
		return this.primary();
	}

	private declaration(): Stmt {
		try {
			if (this.match(TokenType.LET)) return this.letDeclaration();
			if (this.match(TokenType.MUT)) return this.mutDeclaration();
			if (this.match(TokenType.FUNC)) return this.function("function");
			return this.statement();
		} catch (error) {
			this.synchronize();
			console.error("Encountered an error while parsing.", error);
			throw error;
		}
	}

	private letDeclaration(): Stmt {
		const name = this.consume(TokenType.IDENTIFIER, "Expect variable name.");
		let initializer = null;
		if (this.match(TokenType.ASSIGN)) {
			initializer = this.expression();
		}
		return { type: "Let", name, initializer };
	}

	private mutDeclaration(): Stmt {
		const name = this.consume(TokenType.IDENTIFIER, "Expect variable name.");
		let initializer = null;
		if (this.match(TokenType.ASSIGN)) {
			initializer = this.expression();
		}
		return { type: "Mut", name, initializer };
	}

	private function(kind: string): Stmt {
		const name = this.consume(TokenType.IDENTIFIER, `Expect ${kind} name.`);
		this.consume(TokenType.LEFT_PAREN, `Expect '(' after ${kind} name.`);
		const parameters: Token[] = [];
		if (!this.check(TokenType.RIGHT_PAREN)) {
			do {
				if (parameters.length >= 255) {
					this.error(this.peek(), "Can't have more than 255 parameters.");
				}
				parameters.push(
					this.consume(TokenType.IDENTIFIER, "Expect parameter name."),
				);
			} while (this.match(TokenType.COMMA));
		}
		this.consume(TokenType.RIGHT_PAREN, "Expect ')' after parameters.");
		this.consume(TokenType.LEFT_BRACE, `Expect '{' before ${kind} body.`);
		const body = this.block();
		return { type: "Function", name, params: parameters, body };
	}

	private statement(): Stmt {
		if (this.match(TokenType.IF)) return this.ifStatement();
		if (this.match(TokenType.WHILE)) return this.whileStatement();
		if (this.match(TokenType.FOR)) return this.forStatement();
		if (this.match(TokenType.RETURN)) return this.returnStatement();
		if (this.match(TokenType.LEFT_BRACE))
			return { type: "Block", statements: this.block() };
		return this.expressionStatement();
	}

	private ifStatement(): Stmt {
		this.consume(TokenType.LEFT_PAREN, "Expect '(' after 'if'.");
		const condition = this.expression();
		this.consume(TokenType.RIGHT_PAREN, "Expect ')' after if condition.");
		const thenBranch = this.statement();
		let elseBranch = null;
		if (this.match(TokenType.ELSE)) {
			elseBranch = this.statement();
		}
		return { type: "If", condition, thenBranch, elseBranch };
	}

	private whileStatement(): Stmt {
		this.consume(TokenType.LEFT_PAREN, "Expect '(' after 'while'.");
		const condition = this.expression();
		this.consume(TokenType.RIGHT_PAREN, "Expect ')' after condition.");
		const body = this.statement();
		return { type: "While", condition, body };
	}

	private forStatement(): Stmt {
		throw new Error("For statement not implemented yet.");
		// this.consume(TokenType.LEFT_PAREN, "Expect '(' after 'for'.");
		// let initializer: Stmt | null;
		// if (this.match(TokenType.COLON)) {
		// 	initializer = null;
		// } else if (this.match(TokenType.LET)) {
		// 	initializer = this.letDeclaration();
		// } else if (this.match(TokenType.MUT)) {
		// 	initializer = this.mutDeclaration();
		// } else {
		// 	initializer = this.expressionStatement();
		// }
		// let condition: Expr | null = null;
		// if (!this.check(TokenType.SEMICOLON)) {
		// 	condition = this.expression();
		// }
		// this.consume(TokenType.SEMICOLON, "Expect ';' after loop condition.");
		// let increment = null;
		// if (!this.check(TokenType.RIGHT_PAREN)) {
		// 	increment = this.expression();
		// }
		// this.consume(TokenType.RIGHT_PAREN, "Expect ')' after for clauses.");
		// let body = this.statement();
		// if (increment !== null) {
		// 	body = {
		// 		type: "Block",
		// 		statements: [body, { type: "Expression", expression: increment }],
		// 	};
		// }
		// if (condition === null) condition = { type: "Literal", value: true };
		// body = { type: "While", condition, body };
		// if (initializer !== null) {
		// 	body = { type: "Block", statements: [initializer, body] };
		// }
		// return body;
	}

	private returnStatement(): Stmt {
		const keyword = this.previous();
		let value = null;
		// if (!this.check(TokenType.SEMICOLON)) {
		value = this.expression();
		// }
		return { type: "Return", keyword, value };
	}

	private expressionStatement(): Stmt {
		const expr = this.expression();
		return { type: "Expression", expression: expr };
	}

	private block(): Stmt[] {
		const statements: Stmt[] = [];
		while (!this.check(TokenType.RIGHT_BRACE) && !this.isAtEnd()) {
			statements.push(this.declaration());
		}
		this.consume(TokenType.RIGHT_BRACE, "Expect '}' after block.");
		return statements;
	}

	private literal(): Expr {
		const token = this.previous();
		return { type: "Literal", value: token.lexeme };
	}

	private assignment(): Expr {
		const expr = this.or();
		if (this.match(TokenType.ASSIGN)) {
			const equals = this.previous();
			const value = this.assignment();
			if (expr.type === "Variable") {
				return { type: "Assign", name: expr.name, value };
			}
			this.error(equals, "Invalid assignment target.");
		}
		return expr;
	}

	private or(): Expr {
		let expr = this.and();
		while (this.match(TokenType.OR)) {
			const operator = this.previous();
			const right = this.and();
			expr = { type: "Binary", left: expr, operator, right };
		}
		return expr;
	}

	private and(): Expr {
		let expr = this.equality();
		while (this.match(TokenType.AND)) {
			const operator = this.previous();
			const right = this.equality();
			expr = { type: "Binary", left: expr, operator, right };
		}
		return expr;
	}

	private primary(): Expr {
		if (this.match(TokenType.FALSE)) {
			return { type: "Literal", value: false, token: this.previous() };
		}
		if (this.match(TokenType.TRUE)) {
			return { type: "Literal", value: true, token: this.previous() };
		}
		if (this.match(TokenType.INTEGER)) {
			const token = this.previous();
			return { type: "Literal", value: Number(token.lexeme), token };
		}
		// memo: this distinction probably doesn't matter in a JS runtime
		if (this.match(TokenType.DOUBLE)) {
			const token = this.previous();
			return { type: "Literal", value: parseFloat(token.lexeme), token };
		}
		if (this.match(TokenType.STRING)) {
			const token = this.previous();
			return {
				type: "Literal",
				value: token.lexeme,
				token,
			};
		}
		// if (this.match(TokenType.IDENTIFIER)) {
		// 	return { type: "Variable", name: this.previous() };
		// }
		if (this.match(TokenType.LEFT_PAREN)) {
			const expr = this.expression();
			this.consume(TokenType.RIGHT_PAREN, "Expect ')' after expression.");
			return { type: "Grouping", expression: expr };
		}
	}

	private call(): Expr {
		let expr = this.primary();
		while (true) {
			if (this.match(TokenType.LEFT_PAREN)) {
				expr = this.finishCall(expr);
			} else {
				break;
			}
		}
		return expr;
	}

	private finishCall(callee: Expr): Expr {
		const args: Expr[] = [];
		if (!this.check(TokenType.RIGHT_PAREN)) {
			do {
				if (args.length >= 255) {
					this.error(this.peek(), "Can't have more than 255 arguments.");
				}
				args.push(this.expression());
			} while (this.match(TokenType.COMMA));
		}
		const paren = this.consume(
			TokenType.RIGHT_PAREN,
			"Expect ')' after arguments.",
		);
		return { type: "Call", callee, paren, arguments: args };
	}

	private match(...types: TokenType[]): boolean {
		for (const type of types) {
			if (this.check(type)) {
				this.advance();
				return true;
			}
		}
		return false;
	}

	private consume(type: TokenType, message: string): Token {
		if (this.check(type)) return this.advance();
		throw this.error(this.peek(), message);
	}

	private check(type: TokenType): boolean {
		if (this.isAtEnd()) return false;
		return this.peek().type === type;
	}

	private advance(): Token {
		if (!this.isAtEnd()) this.current++;
		return this.previous();
	}

	private isAtEnd(): boolean {
		return this.peek().type === TokenType.EOF;
	}

	private peek(): Token {
		const token = this.tokens[this.current];
		if (!token) throw new ParseError();
		return token;
	}

	private previous(): Token {
		const token = this.tokens[this.current - 1];
		if (!token) throw new ParseError();
		return token;
	}

	private error(token: Token, message: string): ParseError {
		// Report error here
		return new ParseError(
			message + " at " + token.line + ":" + token.column + ".",
		);
	}

	private synchronize() {
		this.advance();
		while (!this.isAtEnd()) {
			// @ts-expect-error TODO: remove semicolons
			if (this.previous().type === TokenType.SEMICOLON) return;
			switch (this.peek().type) {
				case TokenType.FUNC:
				case TokenType.LET:
				case TokenType.MUT:
				case TokenType.FOR:
				case TokenType.IF:
				case TokenType.WHILE:
				case TokenType.RETURN:
					return;
			}
			this.advance();
		}
	}
}