import { describe, test, expect } from "bun:test";
import { Token, TokenType } from "../lexer/lexer";
import type { Expr } from "../ast";
import { print } from "./printer";

describe("Printer", () => {
	test("printing AST in lisp", () => {
		// "-123 * 45.67"
		const expression: Expr = {
			type: "Binary",
			operator: Token.init({
				type: TokenType.STAR,
				lexeme: "*",
				line: 1,
				column: 1,
			}),
			right: {
				type: "Unary",
				operator: Token.init({
					type: TokenType.MINUS,
					lexeme: "-",
					line: 1,
					column: 2,
				}),
				right: {
					type: "Literal",
					token: Token.init({
						type: TokenType.INTEGER,
						lexeme: "123",
						line: 1,
						column: 3,
					}),
				},
			},
			left: {
				type: "Grouping",
				expression: {
					type: "Literal",
					token: Token.init({
						type: TokenType.DOUBLE,
						lexeme: "45.67",
						line: 1,
						column: 8,
					}),
				},
			},
		};
		expect(print(expression)).toEqual("(* (- 123) (group 45.67))");
	});
});
