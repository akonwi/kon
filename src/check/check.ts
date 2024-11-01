import type { Point, SyntaxNode, Tree, TreeCursor } from "tree-sitter";

export type CheckError = {
	point: Point;
	type: "variable_type";
	message: string;
};

export class Checker {
	cursor: TreeCursor;
	errors: CheckError[] = [];

	constructor(readonly tree: Tree) {
		this.cursor = tree.walk();
	}

	check(): CheckError[] {
		// go through children
		if (this.cursor.gotoFirstChild()) {
			do {
				this.visitNode(this.cursor.currentNode);
			} while (this.cursor.gotoNextSibling());
		}

		return this.errors;
	}

	private error(error: CheckError) {
		this.errors.push(error);
	}

	// if a check can be done, do it and go back to the parent node
	// otherwise, continue to the next child
	private visitNode(node: SyntaxNode) {
		if (node.type === "variable_definition") {
			const name = node.childForFieldName("name");
			const type = node.namedChildren
				.find((n) => n.grammarType === "type_declaration")
				?.childForFieldName("type");
			const value = node.childForFieldName("value");
			if (!(name && type && value)) {
				return;
			}

			const getValueLabel = (value: SyntaxNode) => {
				switch (value.type) {
					case "primitive_value": {
						const child = value.firstChild;
						switch (child?.grammarType) {
							case "number":
								return "Num";
							case "string":
								return "Str";
							case "boolean":
								return "Bool";
							default:
								return "unknown";
						}
					}
				}
			};

			switch (type.text) {
				case "Str": {
					if (value.firstChild?.type !== "string") {
						this.error({
							point: value.startPosition,
							type: "variable_type",
							message: `Expected a 'Str' but got '${getValueLabel(value)}'`,
						});
					}
					break;
				}
				case "Num": {
					if (value.firstChild?.type !== "number") {
						this.error({
							point: value.startPosition,
							type: "variable_type",
							message: `Expected a 'Num' but got '${getValueLabel(value)}'`,
						});
					}
					break;
				}
				case "Bool": {
					if (value.firstChild?.type !== "boolean") {
						this.error({
							point: value.startPosition,
							type: "variable_type",
							message: `Expected a 'Bool' but got '${getValueLabel(value)}'`,
						});
					}
				}
			}
			this.cursor.gotoParent();
			return;
		}

		this.check();
	}
}