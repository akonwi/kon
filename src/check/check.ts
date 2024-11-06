import type { Point, Tree, TreeCursor } from "tree-sitter";
import {
	type NumberNode,
	type StringNode,
	SyntaxType,
	type BooleanNode,
	type ExpressionNode,
	type FunctionCallNode,
	type MemberAccessNode,
	type NamedNode,
	type StructDefinitionNode,
	type StructInstanceNode,
	type SyntaxNode,
	type TypeDeclarationNode,
	type TypedTreeCursor,
	type VariableDefinitionNode,
} from "../ast.ts";
import console from "node:console";

export type Diagnostic = {
	level: "error" | "warning";
	location: Point;
	message: string;
};

const RESERVED_KEYWORDS = new Set([
	"let",
	"of",
	"in",
	"if",
	"else",
	"true",
	"false",
	"or",
	"and",
	"struct",
	"enum",
]);

const tokenTypeToText: Record<
	SyntaxType.Boolean | SyntaxType.Number | SyntaxType.String,
	string
> = {
	string: "Str",
	number: "Num",
	boolean: "Bool",
};

class Variable {
	constructor(
		readonly name: string,
		readonly node: VariableDefinitionNode,
	) {}

	get is_mutable(): boolean {
		return this.node.bindingNode.text === "mut";
	}

	get static_type(): string | null {
		const type_declaration = this.node.typeNode?.typeNode;
		if (!type_declaration) {
			return null;
		}
		switch (type_declaration.type) {
			case SyntaxType.PrimitiveType:
				return type_declaration.text;
			case SyntaxType.ListType:
				return SyntaxType.ListType;
			// return `[${type_declaration.innerNode.text}]`;
			case SyntaxType.MapType:
				return `{${type_declaration.keyNode.text}: ${type_declaration.valueNode.text}}`;
			default:
				return null;
		}
	}
}

class StructDef {
	// Map<field_name, field_kon_type>
	readonly fields: Map<string, string> = new Map();

	constructor(readonly node: StructDefinitionNode) {
		for (const field of this.node.fieldNodes) {
			this.fields.set(field.nameNode.text, field.typeNode.text);
		}
	}

	get name(): string {
		return this.node.nameNode.text;
	}

	get static_type(): string {
		return this.name;
	}
}

class LexScope {
	readonly variables: Map<string, Variable> = new Map();
	readonly structs: Map<string, StructDef> = new Map();
}

export class Checker {
	cursor: TreeCursor;
	errors: Diagnostic[] = [];
	scopes: LexScope[] = [new LexScope()];

	constructor(readonly tree: Tree) {
		this.cursor = tree.walk();
	}

	check(): Diagnostic[] {
		const cursor = this.cursor as unknown as TypedTreeCursor;
		this.visit(cursor.currentNode);
		return this.errors;
	}

	visit(node: SyntaxNode) {
		if (node === null) return;
		const methodName = `visit${node.type
			.split("_")
			.map((part) => part.charAt(0).toUpperCase() + part.slice(1))
			.join("")}`;
		console.log("visiting ", node.type);

		// @ts-expect-error - dynamic method call
		const method = this[methodName]?.bind(this);
		if (method) {
			return method(node);
		}
		console.debug(`No visit method for ${node.type}, going through children`);

		for (const child of node.namedChildren) {
			this.visit(child);
		}
		return;
	}

	private error(error: Diagnostic) {
		this.errors.push(error);
	}

	private scope(): LexScope {
		if (this.scopes.length === 0) throw new Error("No scope found");
		return this.scopes.at(0)!;
	}

	visitStructDefinition(node: StructDefinitionNode) {
		const def = new StructDef(node);
		this.scope().structs.set(def.name, def);
	}

	visitStructInstance(node: StructInstanceNode) {
		const struct_name = node.nameNode.text;
		const struct_def = this.scope().structs.get(struct_name);
		if (!struct_def) {
			this.error({
				level: "error",
				message: `Missing definition for type '${struct_name}'.`,
				location: node.startPosition,
			});
			return;
		}

		const expected_fields = struct_def.fields;
		const received_fields = new Set<string>();
		for (const inputFieldNode of node.fieldNodes) {
			const member_name = inputFieldNode.nameNode.text;
			if (expected_fields.has(member_name)) {
				const expected_type = expected_fields.get(member_name)!;
				const provided_type = this.getTypeFromExpressionNode(
					inputFieldNode.valueNode,
				);
				if (
					expected_type !== provided_type &&
					provided_type !== "unknown" &&
					expected_type !== "unknown"
				) {
					this.error({
						location: inputFieldNode.valueNode.startPosition,
						level: "error",
						message: `Expected a '${expected_type}' but got '${provided_type}'`,
					});
				}
				received_fields.add(member_name);
			}
			if (!expected_fields.has(member_name)) {
				this.error({
					level: "warning",
					message: `Struct '${struct_name}' does not have a field named ${member_name}.`,
					location: inputFieldNode.startPosition,
				});
				continue;
			}
		}

		const missing_field_names = new Set<string>();
		for (const field of expected_fields.keys()) {
			if (!received_fields.has(field)) {
				missing_field_names.add(field);
			}
		}
		if (missing_field_names.size > 0) {
			this.error({
				level: "error",
				message: `Missing fields for struct '${struct_name}': ${Array.from(
					missing_field_names,
				).join(", ")}.`,
				location: node.startPosition,
			});
		}
	}

	visitVariableDefinition(node: VariableDefinitionNode) {
		const name = node.nameNode;
		this.validateIdentifier(node.nameNode);

		const typeNode = node.typeNode?.typeNode;
		// todo: type inference
		// if (!type)
		// 	this.error({
		// 		level: "error",
		// 		message: `Missing type declaration for variable ${name?.text ?? ""}.`,
		// 		location: node.startPosition,
		// 	});
		const value = node.valueNodes.filter((n) => n.isNamed).at(0);
		if (!(typeNode && value)) {
			return;
		}

		const declared_type = this.getTypeFromTypeDefNode(typeNode);
		const provided_type = this.getTypeFromExpressionNode(value);
		const assigment_error = this.validateCompatibility(
			declared_type,
			provided_type,
		);
		if (assigment_error != null) {
			this.error({
				location: value.startPosition,
				level: "error",
				message: assigment_error,
			});
		}

		const variable = new Variable(name.text, node);
		this.scope().variables.set(name.text, variable);

		this.cursor.gotoParent();
		return;
	}

	private validateCompatibility(
		expected: string,
		received: string,
	): string | null {
		if (expected === received) {
			return null;
		}
		if (expected === "unknown" || received === "unknown") {
			return "Unable to determine type compatibility";
		}
		if (expected.startsWith("[") && expected.endsWith("]")) {
			if (received === "[]") return null;
			if (expected.slice(1, -1) !== received.slice(1, -1)) {
				return `Expected a '${expected}' and received a list containing '${received.slice(
					1,
					-1,
				)}'.`;
			}
		}
		return null;
	}

	private getTypeFromTypeDefNode(
		node: TypeDeclarationNode["typeNode"],
	): string {
		switch (node.type) {
			case SyntaxType.PrimitiveType:
				return node.text;
			case SyntaxType.ListType: {
				switch (node.innerNode.type) {
					case "identifier": {
						// check that the type exists
						const declaration =
							this.scope().variables.get(node.innerNode.text) ??
							this.scope().structs.get(node.innerNode.text);
						if (declaration == null) {
							this.error({
								level: "error",
								location: node.innerNode.startPosition,
								message: `Missing definition for type '${node.innerNode.text}'.`,
							});
						}
						return node.text;
					}
					case "primitive_type": {
						return node.text;
					}
					default:
						return "unknown";
				}
			}
			default:
				return "unknown";
		}
	}

	private getTypeFromExpressionNode(
		node: ExpressionNode | BooleanNode | NumberNode | StringNode,
	): string {
		switch (node.type) {
			case SyntaxType.Boolean:
			case SyntaxType.Number:
			case SyntaxType.String: {
				return tokenTypeToText[node.type];
			}
			case SyntaxType.PrimitiveValue: {
				if (node.primitiveNode.type === SyntaxType.ListValue) {
					if (node.primitiveNode.namedChildren.length === 0) {
						return "[]";
					}
					const isConsistent =
						new Set(node.primitiveNode.namedChildren.map((c) => c?.type))
							.size === 1;
					if (!isConsistent) {
						console.debug(
							"checking assignment of a list value",
							node.primitiveNode.namedChild(0)?.text,
						);
						return "[unknown]";
					}
					const first = node.primitiveNode.namedChild(0)!;
					if (first.type === SyntaxType.Identifier) {
						const variable = this.scope().variables.get(
							node.primitiveNode.text,
						);
						if (!variable) {
							this.error({
								level: "error",
								location: node.primitiveNode.startPosition,
								message: `Missing definition for variable '${node.primitiveNode.text}'.`,
							});
							return "[unknown]";
						}
						return `[${variable.static_type}]`;
					}
					switch (first.type) {
						case SyntaxType.Boolean:
							return "[Bool]";
						case SyntaxType.Number:
							return "[Num]";
						case SyntaxType.String:
							return "[Str]";
					}
					if (first.type === SyntaxType.StructInstance) {
						console.debug(
							"checking assignment of a struct instance",
							first.nameNode.text,
						);
						const struct = this.scope().structs.get(first.nameNode.text);
						if (!struct) {
							this.error({
								level: "error",
								location: first.nameNode.startPosition,
								message: `Missing definition for struct '${first.nameNode.text}'.`,
							});
						}
						return `[${first.nameNode.text}]`;
					}
					return "[unknown]";
				}
				if (node.primitiveNode.type === SyntaxType.MapValue) {
					return "unknown";
				}
				return tokenTypeToText[node.primitiveNode.type] ?? "unknown";
			}
			case SyntaxType.StructInstance: {
				this.visitStructInstance(node);
				return node.nameNode.text;
			}
			default: {
				return "unknown";
			}
		}
	}

	visitFunctionCall(node: FunctionCallNode) {
		const { targetNode, argumentsNode } = node;
		const targetVariable = (() => {
			switch (targetNode.type) {
				case "identifier": {
					const variable = this.scope().variables.get(targetNode.text);
					if (!variable) {
						this.error({
							level: "error",
							location: targetNode.startPosition,
							message: `Missing declaration for '${targetNode.text}'.`,
						});
						return;
					}
					return variable;
				}
				case SyntaxType.Identifier: {
					// TODO
					return null;
				}
				default:
					return null;
			}
		})();

		switch (targetVariable?.static_type) {
			case "list_type": {
				if (LIST_MEMBERS.has(argumentsNode.text)) {
					const signature = LIST_MEMBERS.get(argumentsNode.text)!;
					if (signature.mutates && !targetVariable.is_mutable) {
						this.error({
							level: "error",
							location: argumentsNode.startPosition,
							message: `Cannot mutate an immutable list. Use 'mut' to make it mutable.`,
						});
					}
				} else {
					this.error({
						level: "error",
						location: argumentsNode.startPosition,
						message: `Unknown member '${argumentsNode.text}' for list type.`,
					});
				}
				break;
			}
			case undefined:
			case null: {
				this.error({
					level: "error",
					location: targetNode.startPosition,
					message: `The type of '${targetNode.text}' is unknown.`,
				});
				break;
			}
			default: {
				console.log(`Unknown type: ${targetVariable?.static_type}`);
				break;
			}
		}
	}

	visitMemberAccess(node: MemberAccessNode) {
		const { memberNode } = node;
		const targetNode = node.targetNodes.filter((n) => n.isNamed).at(0);
		if (targetNode == null) throw new Error("Invalid member access");

		switch (targetNode.type) {
			case SyntaxType.Identifier: {
				const variable = this.scope().variables.get(targetNode.text);
				if (!variable) {
					this.error({
						level: "error",
						location: targetNode.startPosition,
						message: `Missing declaration for '${targetNode.text}'.`,
					});
					return;
				}

				switch (variable.static_type) {
					case "list_type": {
						switch (memberNode.type) {
							case SyntaxType.FunctionCall: {
								const member = memberNode.targetNode.text;
								if (!member) {
									return;
								}
								if (LIST_MEMBERS.has(member)) {
									const signature = LIST_MEMBERS.get(member)!;
									if (signature.mutates && !variable.is_mutable) {
										this.error({
											level: "error",
											location: memberNode.startPosition,
											message: `Cannot mutate an immutable list. Use 'mut' to make it mutable.`,
										});
									}
									if (!signature.callable) {
										this.error({
											level: "error",
											location: memberNode.startPosition,
											message: `${variable.name}.${member} is not a callable function.`,
										});
									}
								} else {
									this.error({
										level: "error",
										location: memberNode.startPosition,
										message: `Unknown member '${member}' for list type.`,
									});
								}
							}
						}
						break;
					}
					case null: {
						this.error({
							level: "error",
							location: targetNode.startPosition,
							message: `The type of '${targetNode.text}' is unknown.`,
						});
						break;
					}
					default: {
						console.log(`Unknown type: ${variable.static_type}`);
						break;
					}
				}
			}
		}
	}

	validateIdentifier(node: NamedNode) {
		if (RESERVED_KEYWORDS.has(node.text)) {
			this.error({
				location: node.startPosition,
				level: "error",
				message: `'${node.text}' is a reserved keyword and cannot be used as a variable name`,
			});
		}
	}
}

const LIST_MEMBERS = new Map<string, { callable: boolean; mutates: boolean }>([
	["at", { mutates: false, callable: true }],
	["concat", { mutates: false, callable: true }],
	["copyWithin", { mutates: true, callable: true }],
	["length", { mutates: false, callable: false }],
	["pop", { mutates: true, callable: true }],
	["push", { mutates: true, callable: true }],
	["reverse", { mutates: true, callable: true }],
	["shift", { mutates: true, callable: true }],
	["slice", { mutates: false, callable: true }],
	["sort", { mutates: true, callable: true }],
	["splice", { mutates: true, callable: true }],
]);
