import { expect } from "jsr:@std/expect";
import { makeParser } from "../parser/parser.ts";
import { Checker, Diagnostic } from "./check.ts";

const parser = makeParser();

Deno.test("variables can be referenced from nested scope", () => {
	const for_loop = parser.parse(`
struct Todo {
  title: Str,
  completed: Bool
}
let list: [Todo] = []
for todo in list {
  if !todo.completed {
    print(todo.title)
  }
}
`);
	expect(new Checker(for_loop).check()).toEqual([]);

	const while_loop = parser.parse(`
mut is_true = true
while (is_true) {
  print(name)
  is_true = false
}
`);
	expect(new Checker(while_loop).check()).toEqual([
		{
			level: "error",
			location: { row: 3, column: 8 },
			message: "Cannot find name 'name'.",
		} satisfies Diagnostic,
	]);
});
