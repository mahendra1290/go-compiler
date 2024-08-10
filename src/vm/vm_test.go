package vm

import (
	"fmt"
	"monkey/src/ast"
	"monkey/src/compiler"
	"monkey/src/lexer"
	"monkey/src/object"
	"monkey/src/parser"
	"testing"
)

func parse(input string) *ast.Program {
	l := lexer.New(input)
	p := parser.New(l)

	return p.ParseProgram()
}

func testIntegerObject(expected int64, actual object.Object) error {
	result, ok := actual.(*object.Integer)
	if !ok {
		return fmt.Errorf("object is not Integer. got=%T (%+v)", actual, actual)
	}

	if result.Value != expected {
		return fmt.Errorf("object has wrong value. got=%d, want=%d", result.Value, expected)
	}

	return nil
}

func testBooleanObject(expected bool, actual object.Object) error {
	result, ok := actual.(*object.Boolean)
	if !ok {
		return fmt.Errorf("object is not Boolean. got=%T (%+v)", actual, actual)
	}

	if result.Value != expected {
		return fmt.Errorf("object has wrong value. got=%t, want=%t", result.Value, expected)
	}

	return nil
}

func testStringObject(expected string, actual object.Object) error {
	result, ok := actual.(*object.String)
	if !ok {
		return fmt.Errorf("object is not String. got=%T (%+v)", actual, actual)
	}

	if result.Value != expected {
		return fmt.Errorf("object has wrong value. got=%q, want=%q", result.Value, expected)
	}

	return nil
}

type vmTestCase struct {
	input    string
	expected interface{}
}

func runVmTests(t *testing.T, tests []vmTestCase) {
	t.Helper()

	for _, tt := range tests {
		program := parse(tt.input)

		comp := compiler.New()
		err := comp.Compile(program)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		vm := New(comp.Bytecode())
		err = vm.Run()
		if err != nil {
			t.Fatalf("vm error: %s", err)
		}

		stackElem := vm.LastPoppedStackElem()

		textExpectedObject(t, tt.expected, stackElem)
	}
}

func textExpectedObject(t *testing.T, expected interface{}, actual object.Object) {
	t.Helper()

	switch expected := expected.(type) {

	case int:
		err := testIntegerObject(int64(expected), actual)
		if err != nil {
			t.Fatalf("testIntegerObject failed: %s", err)
		}

	case bool:
		err := testBooleanObject(bool(expected), actual)
		if err != nil {
			t.Fatalf("testBooleanObject failed: %s", err)
		}

	case string:
		err := testStringObject(string(expected), actual)
		if err != nil {
			t.Fatalf("testStringObject failed: %s", err)
		}

	case []int:
		array, ok := actual.(*object.Array)
		if !ok {
			t.Errorf("object not array: %T (%+v)", actual, actual)
		}

		if len(array.Elements) != len(expected) {
			t.Errorf("wrong num of elements. want=%d, got=%d", len(expected), len(array.Elements))
		}

		for i, expectedElem := range expected {
			err := testIntegerObject(int64(expectedElem), array.Elements[i])
			if err != nil {
				t.Errorf("testIntegerObject failed: %s", err)
			}
		}

	case map[object.HashKey]int64:
		hash, ok := actual.(*object.Hash)
		if !ok {
			t.Errorf("object not Hash: %T (%+v)", actual, actual)
		}

		if len(hash.Pairs) != len(expected) {
			t.Errorf("wrong num of pairs. want=%d, got=%d", len(expected), len(hash.Pairs))
		}

		for expectedKey, expectedValue := range expected {
			pair, ok := hash.Pairs[expectedKey]
			if !ok {
				t.Errorf("no pair found with key %q", expectedKey)
			}

			err := testIntegerObject(int64(expectedValue), pair.Value)
			if err != nil {
				t.Errorf("testIntegerObject failed: %s", err)
			}
		}

	case *object.Null:
		if actual != Null {
			t.Fatalf("object is not Null: %T (%+v)", actual, actual)
		}

	}
}

func TestIntegerArithmetic(t *testing.T) {
	tests := []vmTestCase{
		{"1", 1},
		{"2", 2},
		{"1 + 2", 3}, // FIXME
		{"2 - 1", 1},
		{"2 * 2", 4},
		{"4 / 2", 2},
		{"2 + 2 * 2", 6},
		{"2 * 2 + 2", 6},
		{"(2 + 2) * 2", 8},
		{"2 * 2 * 2 * 2", 16},
		{"8 / 2 * 5 + 6", 26},
		{"2 + 2 * 2 / 2", 4},
		{"2 * (2 + 2)", 8},
		{"-5", -5},
		{"-6", -6},
	}

	runVmTests(t, tests)
}

func TestConditionals(t *testing.T) {
	tests := []vmTestCase{
		{"if (false) { 10 }", Null},
		{"if (1 > 2) { 10 }", Null},
		{"if (true) { 10 }", 10},
		{"if (false) { 10 } else { 20 }", 20},
		{"if (true) { 10 } else { 20 }", 10},
		{"if (1) { 10 } else { 20 }", 10},
		{"if (1 < 2) { 10 } else { 20 }", 10},
		{"if (1 > 2) { 10 } else { 20 }", 20},
		{"if (1 < 2) { 10 }", 10},
		{"if ((if (false) { 10 })) { 10 } else { 20 }", 20},
	}

	runVmTests(t, tests)

}

func TestBooleanExpressions(t *testing.T) {
	tests := []vmTestCase{
		{"!(if (false) { 5; })", true},
		{"true", true},
		{"false", false},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"1 == 1", true},
		{"1 != 1", false},
		{"1 == 2", false},
		{"1 != 2", true},
		{"true == true", true},
		{"false == false", true},
		{"true == false", false},
		{"true != false", true},
		{"false != true", true},
		{"(1 < 2) == true", true},
		{"(1 < 2) == false", false},
		{"(1 > 2) == true", false},
		{"(1 > 2) == false", true},
		{"!true", false},
		{"!false", true},
		{"!!true", true},
		{"!!false", false},
		{"!5", false},
		{"!!5", true},
	}

	runVmTests(t, tests)
}

func TestGlobalLetStatements(t *testing.T) {
	tests := []vmTestCase{
		{"let one = 1; one", 1},
		{"let one = 1; let two = 2; one + two", 3},
		{"let one = 1; let two = one + one; one + two", 3},
	}

	runVmTests(t, tests)
}

func TestStringExpressions(t *testing.T) {
	tests := []vmTestCase{
		{`"monkey"`, "monkey"},
		{`"mon" + "key"`, "monkey"},
		{`"mon" + "key" + "banana"`, "monkeybanana"},
	}
	runVmTests(t, tests)
}

func TestArrayLiterals(t *testing.T) {
	tests := []vmTestCase{
		{"[]", []int{}},
		{"[1, 2, 3]", []int{1, 2, 3}},
		{"[1 + 2, 4 - 2, 3 * 4]", []int{3, 2, 12}},
	}

	runVmTests(t, tests)
}

func TestHashLiterals(t *testing.T) {
	tests := []vmTestCase{
		{"{}", map[object.HashKey]int64{}},
		{"{1: 2, 3: 4}", map[object.HashKey]int64{
			(&object.Integer{Value: 1}).HashKey(): 2,
			(&object.Integer{Value: 3}).HashKey(): 4,
		}},
		{"{1 + 1: 2 + 3, 4 * 2: 5 * 6}", map[object.HashKey]int64{
			(&object.Integer{Value: 2}).HashKey(): 5,
			(&object.Integer{Value: 8}).HashKey(): 30,
		}},
	}

	runVmTests(t, tests)
}

func TestIndexExpressions(t *testing.T) {
	tests := []vmTestCase{
		{"[][0]", Null},
		{"[1][0]", 1},
		{"[1, 2][1]", 2},
		{"[1, 2][10]", Null},
		{"[[1, 2, 3], [4, 5]][0][1]", 2},
		{"{}[0]", Null},
		{"{1: 2, 3: 4}[1]", 2},
		{"{1: 2, 3: 4}[3]", 4},
		{"{1: 2, 3: 4}[4]", Null},
	}

	runVmTests(t, tests)
}

func TestIndexAssignmentExpressions(t *testing.T) {
	tests := []vmTestCase{
		{"let arr = [1]; arr[0] = 2; arr[0];", 2},
		{"let arr = []; arr[0] = 2; arr[0];", Null},
		{"let obj = {}; obj[1] = 5; obj[1]", 5},
		{"let obj = {1: 2}; obj[1] = 3; obj[1]", 3},
		{"let obj = {}; obj[1+1] = 2; obj[2]", 2},
		{"let arr = [[1, 2], 3]; arr[0] = [3, 4 + 4]; arr[0]", []int{3, 8}},
	}

	runVmTests(t, tests)
}

func TestCallingFunctionsWithoutArguments(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
			let fivePlusTen = fn() { 5 + 10; };
			fivePlusTen();`,
			expected: 15,
		},
		{
			input: `
			let one = fn() { 1; };
			let two = fn() { 2; };
			one() + two();`,
			expected: 3,
		},
		{
			input: `
			let a = fn() { 1 }
			let b = fn() { a() + 1 }
			let c = fn() { b() + 1 }
			c()`,
			expected: 3,
		},
	}

	runVmTests(t, tests)
}

func TestCallingFunctionsWithArguments(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
			let fivePlusTen = fn(x) { x + 5 + 10; };
			fivePlusTen(5);`,
			expected: 20,
		},
		{
			input: `
			let one = fn(x) { x; };
			let two = fn(x) { x; };
			one(3) + two(2);`,
			expected: 5,
		},
		{
			input: `
			let sum = fn(a, b) { a+b; }
			sum(1, 3)`,
			expected: 4,
		},
		{
			input: `
			let sum = fn(a, b) {
				let c = a + b;
				return c;
			}
			sum(3, 4)
			`,
			expected: 7,
		},
		{
			input: `
			let sum = fn(a, b) {
				let c = a + b;
				return c;
			}
			sum(2, 2) + sum(1, 2)
			`,

			expected: 7,
		},
		{
			input: `
			let globalNum = 10;
           	let sum = fn(a, b) {
               let c = a + b;
               c + globalNum;
			};
            let outer = fn() {
               sum(1, 2) + sum(3, 4) + globalNum;
			};
           	outer() + globalNum;
			`,
			expected: 50,
		},
	}

	runVmTests(t, tests)
}

func TestFunctionsWithReturnStatements(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
			let earlyExit = fn() { return 99; 100 };
			earlyExit()`,
			expected: 99,
		},
		{
			input: `
			let earlyExit = fn() { return 99; return 100; };
			earlyExit()`,
			expected: 99,
		},
	}

	runVmTests(t, tests)
}

func TestFunctionsWithoutReturnValue(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
			let noReturn = fn() { };
			noReturn();`,
			expected: Null,
		},
		{
			input: `
			let first = fn() {};
			let second = fn() { first(); };
			second();`,
			expected: Null,
		},
	}

	runVmTests(t, tests)
}

func TestFirstClassFunctions(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
			   let returnsOne = fn() { 1; };
			   let returnsOneReturner = fn() { returnsOne; };
			   returnsOneReturner()();
			   `,
			expected: 1,
		},
	}
	runVmTests(t, tests)
}

func TestCallingFunctionsWithBindings(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
			let one = fn() { let one = 1; one; }
			one();
			`,
			expected: 1,
		},
		{
			input: `
			let oneAndTwo = fn() { let one = 1; let two = 2; one + two; };
			oneAndTwo();`,
			expected: 3,
		},
		{
			input: `
			let oneAndTwo = fn() { let one = 1; let two = 2; one + two; }
			let threeAndFour = fn() { let three = 3; let four = 4; three + four; }
			oneAndTwo() + threeAndFour();`,
			expected: 10,
		},
		{
			input: `
			let firstFoobar = fn() { let foobar = 100; foobar; }
			let secondFoobar = fn() { let foobar = 50; foobar; }
			firstFoobar() + secondFoobar();`,
			expected: 150,
		},
		{
			input: `
			let globalSeed = 50;
			let minusOne = fn() {
			  let num = 1;
			  return globalSeed - num;
			}
			let minusTwo = fn() {
			  let num = 2;
			  return globalSeed - num;
			}
			minusOne() + minusTwo();`,
			expected: 97,
		},
		{
			input: `
			let returnsOneReturner = fn() {
			let returnsOne = fn() { 1; };
				returnsOne;
			};
			returnsOneReturner()();
			`,
			expected: 1,
		},
	}

	runVmTests(t, tests)
}

func TestCallingWithWrongNumOfArguments(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
			fn() { 5; }(1);`,
			expected: "wrong number of arguments: want=0 got=1",
		},
		{
			input: `
			fn(a, b){ a + b}(3)`,
			expected: "wrong number of arguments: want=2 got=1",
		},
	}

	for _, tt := range tests {
		program := parse(tt.input)

		comp := compiler.New()
		err := comp.Compile(program)
		if err != nil {
			t.Errorf("compiler error: %s", err)
		}

		vm := New(comp.Bytecode())
		err = vm.Run()

		if err == nil {
			t.Errorf("expected vm error but got none")
		}

		if err.Error() != tt.expected {
			t.Errorf("expected vm error %s, got %s", tt.expected, err)
		}
	}

}
