let count = 0;
while (count <= 9) {
	count += 1;
}
const is_true = false;
while (is_true || count > 0) {
	count -= 1;
}
count = 3;
do {
	count -= 1;
} while (count > 0);
for (let i = 1; i <= 10; i++) {
	console.log(i);
}
console.log("counting from 1 to count + 3");
for (let i = 1; i <= count + 3; i++) {
	console.log(i);
}
