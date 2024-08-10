package parser

import (
	"fmt"
	"monkey/src/ast"
	"monkey/src/lexer"
	"testing"
)

func setup(t *testing.T, input string) *ast.Program {
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	return program
}

func TestLetStatement(t *testing.T) {
	input := `
let x = 5;
let y = 10;
let foobar = 838383;
`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)
	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	if len(program.Statements) != 3 {
		t.Fatalf("program.Statements does not contain 3 statements. got=%d", len(program.Statements))
	}

	tests := []struct {
		expectedIdentifier string
	}{
		{"x"},
		{"y"},
		{"foobar"},
	}

	for i, tt := range tests {
		stmt := program.Statements[i]
		if !testLetStatment(t, stmt, tt.expectedIdentifier) {
			return
		}
	}

}

func TestIndexAssignmentStatment(t *testing.T) {
	tests := []struct {
		input            string
		expectedVariable string
		expectedIndex    string
		expected         string
	}{
		{"let a = [1, 2]; a[0] = 5;", "a", "0", "5"},
		{"let a = [1, 2]; a[1] = 5;", "a", "1", "5"},
		{"let a = [1, 2]; a[3] = 5 + 5;", "a", "3", "(5 + 5)"},
		{"let a = [1, 2]; a[2+2] = 5;", "a", "(2 + 2)", "5"},
		{`let a = {"name": 5}; a["name"] = 6;`, "a", "name", "6"},
	}

	for _, tt := range tests {
		program := setup(t, tt.input)

		fmt.Println(program.String())

		if len(program.Statements) != 2 {
			t.Fatalf("program.Statements does not contain 3 statements. got=%d", len(program.Statements))
		}

		stmt := program.Statements[1].(*ast.ExpressionStatement)
		exp, ok := stmt.Expression.(*ast.IndexAssignmentExpression)
		if !ok {
			t.Error("p", program.String())
			t.Fatalf("program.Statement[1] not IndexAssignmentStatement. got=%T %+v", stmt, stmt)
		}

		if exp.Index.Left.String() != tt.expectedVariable {
			t.Fatalf("exp.Variable not %s. got=%s", tt.expectedVariable, exp.Index.Left.String())
		}

		if exp.Index.Index.String() != tt.expectedIndex {
			t.Fatalf("exp.Index.Index not %s. got=%s", tt.expectedIndex, exp.Index.Index.String())
		}

		if exp.Value.String() != tt.expected {
			t.Fatalf("exp.Value not %s. got=%s", tt.expected, exp.Value.String())

		}

	}

}

func TestLetStatements(t *testing.T) {

	tests := []struct {
		input              string
		expectedIdentifier string
		expectedValue      interface{}
	}{
		{"let x = 5;", "x", 5},
		{"let y = true;", "y", true},
		{"let foobar = y;", "foobar", "y"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain 1 statements. got %d", len(program.Statements))
		}

		stmt := program.Statements[0]
		if !testLetStatment(t, stmt, tt.expectedIdentifier) {
			return
		}

		val := stmt.(*ast.LetStatement).Value

		if !testLiteralExpression(t, val, tt.expectedValue) {
			return
		}
	}
}

func TestAssignStatements(t *testing.T) {

	tests := []struct {
		input              string
		expectedIdentifier string
		expectedValue      string
	}{
		{"x = 5;", "x", "5"},
		{"y = true;", "y", "true"},
		{"foobar = 2 * 2;", "foobar", "(2 * 2)"},
	}

	for _, tt := range tests {
		program := setup(t, tt.input)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain 1 statements. got %d", len(program.Statements))
		}

		stmt := program.Statements[0]
		if !testAssignStatment(t, stmt, tt.expectedIdentifier, tt.expectedValue) {
			return
		}
	}
}

func TestReturnStatement(t *testing.T) {

	input := `
return 5;
return 10;
return 993322;
`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 3 {
		t.Fatalf("program.Statements does not contain 3 statements. got=%d", len(program.Statements))
	}

	for _, stmt := range program.Statements {
		returnStmt, ok := stmt.(*ast.ReturnStatement)
		if !ok {
			t.Errorf("stmt not *ast.ReturnStatement. got=%T", stmt)
			continue
		}
		if returnStmt.TokenLiteral() != "return" {
			t.Errorf("returnStmt.TokenLiteral not 'return'. got %q", returnStmt.TokenLiteral())
		}

	}

}

func TestIdentifierExpression(t *testing.T) {
	input := "foobar;"

	program := setup(t, input)

	if len(program.Statements) != 1 {
		t.Fatalf("program has not enough statement. go=%d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	ident, ok := stmt.Expression.(*ast.Identifier)
	if !ok {
		t.Fatalf("stmt.Expression is not *ast.Identifier. got=%T", stmt.Expression)
	}

	if ident.Value != "foobar" {
		t.Errorf("ident.Value not %s. got=%s", "foobar", ident.Value)
	}

	if ident.TokenLiteral() != "foobar" {
		t.Errorf("ident.TokenLiteral not %s. got=%s", "foobar", ident.TokenLiteral())
	}

}

func TestIntegerLiteralExpression(t *testing.T) {
	input := "5;"

	program := setup(t, input)

	if len(program.Statements) != 1 {
		t.Fatalf("program has not enough statement. go=%d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	literal, ok := stmt.Expression.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("exp not *ast.IntegerLiteral. got=%T", stmt.Expression)

	}
	if literal.Value != 5 {
		t.Fatalf("literal.Value not %d. got=%d", 4, literal.Value)
	}
	if literal.TokenLiteral() != "5" {
		t.Fatalf("literal.TokenLiteral() not %s. got=%s", "5", literal.TokenLiteral())
	}

}

func TestStringLiteralExpression(t *testing.T) {
	input := `"hello world";`

	program := setup(t, input)
	stmt := program.Statements[0].(*ast.ExpressionStatement)
	literal, ok := stmt.Expression.(*ast.StringLiteral)
	if !ok {
		t.Fatalf("stmt.Expression not ast.StringLiteral. got=%T", stmt.Expression)
	}

	if literal.Value != "hello world" {
		t.Errorf("literal.Value not %q, got=%q", "hello world", literal.Value)
	}

}

func testLetStatment(t *testing.T, s ast.Statement, name string) bool {
	if s.TokenLiteral() != "let" {
		t.Errorf("s.TokenLiteral not 'let'. got=%q", s.TokenLiteral())
		return false
	}
	letStmt, ok := s.(*ast.LetStatement)
	if !ok {
		t.Errorf("s not *ast.LetStatement. got=%T", s)
		return false
	}

	if letStmt.Name.Value != name {
		t.Errorf("letStmt.Name.Value not '%s'. got %q", name, letStmt.Name.Value)
		return false
	}

	if letStmt.Name.TokenLiteral() != name {
		t.Errorf("letStmt.Name not '%s'. got=%s", name, letStmt.Name)
		return false
	}
	return true
}

func testAssignStatment(t *testing.T, s ast.Statement, name string, value string) bool {
	if s.TokenLiteral() != name {
		t.Errorf("s.TokenLiteral not %q. got=%q", name, s.TokenLiteral())
		return false
	}
	assignStmt, ok := s.(*ast.AssignStatement)
	if !ok {
		t.Errorf("s not *ast.LetStatement. got=%T", s)
		return false
	}

	if assignStmt.Variable.Value != name {
		t.Errorf("letStassignStmt.Variable.Value not '%s'. got %q", name, assignStmt.Variable.Value)
		return false
	}

	if assignStmt.Variable.TokenLiteral() != name {
		t.Errorf("letStmt.Value not '%s'. got=%s", name, assignStmt.Variable.String())
		return false
	}

	if assignStmt.Value.String() != value {
		t.Errorf("assigStmt.Value not %s. got=%s", value, assignStmt.Value.String())
		return false
	}

	return true
}

func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.Errors()

	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}
	t.FailNow()
}

func TestParsingPrefixExpressions(t *testing.T) {
	prefixTests := []struct {
		input    string
		operator string
		value    interface{}
	}{
		{"!5;", "!", 5},
		{"-15;", "-", 15},
		{"!true", "!", true},
		{"!false", "!", false},
	}

	for _, tt := range prefixTests {
		program := setup(t, tt.input)

		fmt.Print("===", program.String(), program.Statements, "====")

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain %d statements. got=%d\n", 1, len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statement[0] is not ast.ExpressionStatement. got=%T\n", stmt)

		}

		exp, ok := stmt.Expression.(*ast.PrefixExpression)
		if !ok {
			t.Fatalf("program.Statemets[0] is not ast.PrefixExpresion. got=%T\n", stmt.Expression)
		}

		if exp.Operator != tt.operator {
			t.Fatalf("exp.Operator not equal to %s. got=%s\n", tt.operator, exp.Operator)

		}

		if !testLiteralExpression(t, exp.Right, tt.value) {
			return
		}

	}
}

func TestParsingInfixExpresssions(t *testing.T) {
	tests := []struct {
		input    string
		left     interface{}
		operator string
		right    interface{}
	}{
		{"5 + 5;", 5, "+", 5},
		{"5 - 5;", 5, "-", 5},
		{"5 / 5;", 5, "/", 5},
		{"5 * 5;", 5, "*", 5},
		{"5 > 5;", 5, ">", 5},
		{"5 > 5;", 5, ">", 5},
		{"5 == 5;", 5, "==", 5},
		{"5 != 5;", 5, "!=", 5},
		{"alice == bob;", "alice", "==", "bob"},
		{"alice != bob;", "alice", "!=", "bob"},
		{"alice != 5;", "alice", "!=", 5},
		{"alice == true", "alice", "==", true},
		{"false == true", false, "==", true},
	}

	for _, tt := range tests {
		program := setup(t, tt.input)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements len not equal to 1. got = %d", len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T", program.Statements[0])
		}

		if !testInfixExpression(t, stmt.Expression, tt.left, tt.operator, tt.right) {
			t.Fatalf("testInfixExpression failed")
		}
	}
}

func TestOperatorPrecedenceParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			"-a * b",
			"((-a) * b)",
		},
		{
			"!-a",
			"(!(-a))",
		},
		{
			"a+b+c",
			"((a + b) + c)",
		},
		{
			"a+b-c",
			"((a + b) - c)",
		},
		{
			"a*b/c",
			"((a * b) / c)",
		},
		{
			"a/b*c",
			"((a / b) * c)",
		},
		{
			"a+b/c",
			"(a + (b / c))",
		},
		{
			"a + b * c + d / e - f",
			"(((a + (b * c)) + (d / e)) - f)",
		}, {
			"3 + 4; -5 * 5",
			"(3 + 4)((-5) * 5)",
		},
		{
			"5 > 4 == 3 < 4",
			"((5 > 4) == (3 < 4))",
		}, {
			"5 < 4 != 3 > 4",
			"((5 < 4) != (3 > 4))",
		},
		{
			"3 + 4 * 5 == 3 * 1 + 4 * 5",
			"((3 + (4 * 5)) == ((3 * 1) + (4 * 5)))",
		},
		{
			"3 + 4 * 5 == 3 * 1 + 4 * 5",
			"((3 + (4 * 5)) == ((3 * 1) + (4 * 5)))",
		},
		{
			"true",
			"true",
		},
		{
			"false",
			"false",
		},
		{
			"3 > 5 == false",
			"((3 > 5) == false)",
		},
		{
			"3 < 5 == false",
			"((3 < 5) == false)",
		},
		{
			"!false",
			"(!false)",
		},
		{
			"true == false",
			"(true == false)",
		},
		{
			"(1 + 2) * 5",
			"((1 + 2) * 5)",
		},
		{
			"1 + (2 + 5) + 6",
			"((1 + (2 + 5)) + 6)",
		},
		{
			"a + add(b * c) + d",
			"((a + add((b * c))) + d)",
		},
		{
			"add(a, b, 1, 2 * 3, 4 + 5, add(6, 7 * 8))",
			"add(a, b, 1, (2 * 3), (4 + 5), add(6, (7 * 8)))",
		},
		{
			"add(a + b + c * d / f + g)",
			"add((((a + b) + ((c * d) / f)) + g))",
		},
		{
			"a * [1, 2, 3, 4][b * c] * d",
			"((a * ([1, 2, 3, 4][(b * c)])) * d)",
		},
	}

	for _, tt := range tests {
		program := setup(t, tt.input)

		actual := program.String()

		if actual != tt.expected {
			t.Errorf("expected=%q, got=%q", tt.expected, actual)
		}
	}
}

func TestBooleanExpression(t *testing.T) {
	input := "true;"

	program := setup(t, input)

	stmtLen := len(program.Statements)

	if stmtLen != 1 {
		t.Fatalf("program.Statements len not 1. got=%T", stmtLen)
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("stmt not of type (*ast.ExpressionStatement). got=%T", stmt)
	}

	boolExp, ok := stmt.Expression.(*ast.Boolean)
	if !ok {
		t.Fatalf("stmt.Expression not of type *ast.BooleanExpression. got=%T", boolExp)
	}

	if boolExp.Value != true {
		t.Fatalf("boolExp.Value not true. got=%v", boolExp.Value)
	}

	if boolExp.TokenLiteral() != "true" {
		t.Fatalf("boolExp.TokenLiteral not 'true'. got=%q", boolExp.TokenLiteral())
	}

}

func TestIfExpression(t *testing.T) {
	input := "if (x < y) { x }"

	program := setup(t, input)

	pLen := len(program.Statements)

	if pLen != 1 {
		t.Fatalf("len(program.Statements) is not 1. got=%d", pLen)
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] not of type (*ast.ExpressionStatement). got=%T", stmt)
	}

	exp, ok := stmt.Expression.(*ast.IfExpression)
	if !ok {
		t.Fatalf("stmt.Expression not of type (*ast.IfExpression). got=%T", exp)
	}

	if !testInfixExpression(t, exp.Condition, "x", "<", "y") {
		return
	}

	if len(exp.Consequence.Statements) != 1 {
		t.Fatalf("len(exp.Consequence.Statements) is not 1. got=%d", len(exp.Consequence.Statements))
	}

	consequence, ok := exp.Consequence.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("consequence is not *ast.ExpressionStatement. got=%T", consequence)
	}

	if !testIdentifier(t, consequence.Expression, "x") {
		return
	}

	if exp.Alternative != nil {
		t.Fatalf("exp.Alternative was not nil. got=%+v", exp.Alternative)
	}

}

func TestForStatement(t *testing.T) {
	input := `
for i, v in arr {
	let x = i * 2;
}
`

	program := setup(t, input)

	pLen := len(program.Statements)

	if pLen != 1 {
		t.Fatalf("len(program.Statements) is not 1. got=%d", pLen)
	}
	stmt, ok := program.Statements[0].(*ast.ForStatement)

	if !ok {
		t.Fatalf("program.Statements[0] not of type (*ast.ForStatement). got=%T", stmt)
	}

	if stmt.Iterator.String() != "arr" {
		t.Fatalf("stmt.Iterator.Value not %q got=%q", "arr", stmt.Iterator.String())
	}

	if stmt.Index.Value != "i" {
		t.Fatalf("stmt.Index.Value not i got=%q", stmt.Index.Value)
	}

	if stmt.Value.Value != "v" {
		t.Fatalf("stmt.Value.Value not i got=%q", stmt.Value.Value)
	}

	if len(stmt.Block.Statements) != 1 {
		t.Fatalf("stmt.Block.Statements len not 1 got=%d", len(stmt.Block.Statements))
	}

	block, ok := stmt.Block.Statements[0].(*ast.LetStatement)
	if !ok {
		t.Fatalf("body statment 1 not LetStatement got=%T", stmt.Block.Statements[0])
	}

	testLetStatment(t, block, "x")
}

func TestIfElseExpression(t *testing.T) {
	input := "if (x < y) { x } else { y }"

	program := setup(t, input)

	pLen := len(program.Statements)

	if pLen != 1 {
		t.Fatalf("len(program.Statements) is not 1. got=%d", pLen)
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] not of type (*ast.ExpressionStatement). got=%T", stmt)
	}

	exp, ok := stmt.Expression.(*ast.IfExpression)
	if !ok {
		t.Fatalf("stmt.Expression not of type (*ast.IfExpression). got=%T", exp)
	}

	if !testInfixExpression(t, exp.Condition, "x", "<", "y") {
		return
	}

	if len(exp.Consequence.Statements) != 1 {
		t.Fatalf("len(exp.Consequence.Statements) is not 1. got=%d", len(exp.Consequence.Statements))
	}

	consequence, ok := exp.Consequence.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("consequence is not *ast.ExpressionStatement. got=%T", consequence)
	}

	if !testIdentifier(t, consequence.Expression, "x") {
		return
	}

	if exp.Alternative == nil {
		t.Fatalf("exp.Alternative was not nil. got=%+v", exp.Alternative)
	}

	if len(exp.Alternative.Statements) != 1 {
		t.Fatalf("len(exp.Alternative.Statements) is not 1. got=%d", len(exp.Alternative.Statements))
	}

	alternative, ok := exp.Alternative.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("alternative is not *ast.ExpressionStatement. got=%T", alternative)
	}

	if !testIdentifier(t, alternative.Expression, "y") {
		return
	}
}

func TestFunctionLiteral(t *testing.T) {
	input := `fn(x, y) { x + y }`

	program := setup(t, input)

	if !testStatementsLen(t, program, 1) {
		return
	}

	stmt, ok := testExpressionStatement(t, program)
	if !ok {
		return
	}

	function, ok := stmt.Expression.(*ast.FunctionLiteral)
	if !ok {
		t.Fatalf("stmt.Expression is not *ast.FunctionLiteral. got=%T", function)
	}

	if len(function.Parameters) != 2 {
		t.Fatalf("function literal parameters wrong. want 2, got=%d\n", len(function.Parameters))
	}

	testLiteralExpression(t, function.Parameters[0], "x")
	testLiteralExpression(t, function.Parameters[1], "y")

	if len(function.Body.Statements) != 1 {
		t.Fatalf("function.Body.Statements has not 1 statement. got=%d\n", len(function.Body.Statements))
	}

	bodyStmt, ok := function.Body.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("function body stmt is not ast.ExpressionStatement. got=%T",
			function.Body.Statements[0])
	}

	testInfixExpression(t, bodyStmt.Expression, "x", "+", "y")
}

func TestFunctionParameterParsing(t *testing.T) {
	tests := []struct {
		input          string
		expectedParams []string
	}{
		{
			input: "fn() {};", expectedParams: []string{},
		},
		{
			input: "fn(x) {};", expectedParams: []string{"x"},
		},

		{
			input: "fn(x, y) {};", expectedParams: []string{"x", "y"},
		},
	}

	for _, tt := range tests {
		program := setup(t, tt.input)

		stmt := program.Statements[0].(*ast.ExpressionStatement)
		function := stmt.Expression.(*ast.FunctionLiteral)

		if len(function.Parameters) != len(tt.expectedParams) {
			t.Errorf("len of parameters wrong. want %d, got %d", len(tt.expectedParams), len(function.Parameters))
		}

		for i, ident := range tt.expectedParams {
			testLiteralExpression(t, function.Parameters[i], ident)
		}
	}
}

func TestCallExpressionParsing(t *testing.T) {
	input := "add(1, 2 * 3, 4 + 5);"

	program := setup(t, input)

	if !testStatementsLen(t, program, 1) {
		return
	}

	stmt, ok := testExpressionStatement(t, program)
	if !ok {
		return
	}

	exp, ok := stmt.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("stmt.Expression is not *ast.CallExpression. got=%T", exp)
	}

	if !testIdentifier(t, exp.Function, "add") {
		return
	}

	if len(exp.Arguments) != 3 {
		t.Fatalf("wrong length of arguments. gpt=%d", len(exp.Arguments))
	}

	testLiteralExpression(t, exp.Arguments[0], 1)
	testInfixExpression(t, exp.Arguments[1], 2, "*", 3)
	testInfixExpression(t, exp.Arguments[2], 4, "+", 5)
}

func TestCallExpressionArgumentParsing(t *testing.T) {
	tests := []struct {
		input        string
		expectedArgs []string
	}{
		{
			input:        "add()",
			expectedArgs: []string{},
		},
		{
			input:        "add(x)",
			expectedArgs: []string{"x"},
		},
		{
			input:        "add(1, 2 + 5)",
			expectedArgs: []string{"1", "(2 + 5)"},
		},
	}

	for _, tt := range tests {
		program := setup(t, tt.input)
		stmt := program.Statements[0].(*ast.ExpressionStatement)
		callExp := stmt.Expression.(*ast.CallExpression)

		if len(callExp.Arguments) != len(tt.expectedArgs) {
			t.Fatalf("arguments length mismatch. want %d, got %d", len(tt.expectedArgs), len(callExp.Arguments))
		}

		for i, arg := range tt.expectedArgs {
			if callExp.Arguments[i].String() != arg {
				t.Fatalf("callExp.Arguments[%d] want %s got %s", i, arg, callExp.Arguments[i].String())
			}
		}
	}
}

func TestParsingArrayLiterals(t *testing.T) {
	input := "[1, 2 * 2, 3 + 3]"

	program := setup(t, input)
	stmt, _ := program.Statements[0].(*ast.ExpressionStatement)
	array, ok := stmt.Expression.(*ast.ArrayLiteral)
	if !ok {
		t.Fatalf("exp not ast.ArrayLiteral. got=%T", stmt.Expression)
	}

	if len(array.Elements) != 3 {
		t.Fatalf("len(array.Elements) not 3, got=%d", len(array.Elements))
	}

	testIntegerLiteral(t, array.Elements[0], 1)
	testInfixExpression(t, array.Elements[1], 2, "*", 2)
	testInfixExpression(t, array.Elements[2], 3, "+", 3)
}

func TestParsingIndexExpressions(t *testing.T) {
	input := "myArray[1+1]"

	program := setup(t, input)

	stmt, _ := program.Statements[0].(*ast.ExpressionStatement)
	indexExp, ok := stmt.Expression.(*ast.IndexExpression)
	if !ok {
		t.Fatalf("stmt.Expression not *ast.IndexExpression. got=%T", stmt.Expression)
	}

	if !testIdentifier(t, indexExp.Left, "myArray") {
		return
	}

	if !testInfixExpression(t, indexExp.Index, 1, "+", 1) {
		return
	}
}

func TestParsingHashLiteralsStringKeys(t *testing.T) {
	input := `{"one": 1, "two": 2, "three": 3}`

	program := setup(t, input)
	stmt, _ := program.Statements[0].(*ast.ExpressionStatement)
	hash, ok := stmt.Expression.(*ast.HashLiteral)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.HashLiteral. got=%T", stmt.Expression)
	}

	if len(hash.Pairs) != 3 {
		t.Fatalf("hash.Pairs has wrong length, want=3, got=%d", len(hash.Pairs))
	}

	expected := map[string]int64{
		"one":   1,
		"two":   2,
		"three": 3,
	}

	for key, value := range hash.Pairs {
		literal, ok := key.(*ast.StringLiteral)
		if !ok {
			t.Errorf("key is not *ast.StringLiteral. got=%T", key)
		}

		expectedValue := expected[literal.String()]

		testIntegerLiteral(t, value, expectedValue)
	}
}

func TestParsingEmptyHashLiteral(t *testing.T) {
	input := `{}`

	program := setup(t, input)

	stmt, _ := program.Statements[0].(*ast.ExpressionStatement)
	hash, ok := stmt.Expression.(*ast.HashLiteral)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.HashLiteral. got=%T", stmt.Expression)
	}

	if len(hash.Pairs) != 0 {
		t.Fatalf("hash.Pairs has wrong length, want=0, got=%d", len(hash.Pairs))
	}
}

func TestParsingHashLiteralsWithExpression(t *testing.T) {
	input := `{"one": 0+1, "two": 10-8, "three": 15/5}`

	program := setup(t, input)
	stmt, _ := program.Statements[0].(*ast.ExpressionStatement)
	hash, ok := stmt.Expression.(*ast.HashLiteral)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.HashLiteral. got=%T", stmt.Expression)
	}

	if len(hash.Pairs) != 3 {
		t.Fatalf("hash.Pairs has wrong length, want=3, got=%d", len(hash.Pairs))
	}

	tests := map[string]func(ast.Expression){
		"one": func(e ast.Expression) {
			testInfixExpression(t, e, 0, "+", 1)
		},
		"two": func(e ast.Expression) {
			testInfixExpression(t, e, 10, "-", 8)
		},
		"three": func(e ast.Expression) {
			testInfixExpression(t, e, 15, "/", 5)
		},
	}

	for key, value := range hash.Pairs {
		literal, ok := key.(*ast.StringLiteral)
		if !ok {
			t.Errorf("key is not *ast.StringLiteral. got=%T", key)
		}

		testFunc, ok := tests[literal.String()]
		if !ok {
			t.Errorf("No test function for key %q found", literal.String())
		}

		testFunc(value)
	}

}

func testStatementsLen(t *testing.T, p *ast.Program, expectedLen int) bool {
	if len(p.Statements) != expectedLen {
		t.Errorf("program.Statements does not contain %d statements. got=%d\n", expectedLen, len(p.Statements))
		return false
	}

	return true
}

func testExpressionStatement(t *testing.T, program *ast.Program) (*ast.ExpressionStatement, bool) {
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Errorf("program.Statements[0] not of type (*ast.ExpressionStatement). got=%T", stmt)
		return nil, false
	}

	return stmt, true
}

func testIntegerLiteral(t *testing.T, il ast.Expression, value int64) bool {
	integ, ok := il.(*ast.IntegerLiteral)
	if !ok {
		t.Errorf("il not *ast.IntegerLiteral. got=%T", il)
		return false
	}
	if integ.Value != value {
		t.Errorf("integ.Value not %d. got=%d", value, integ.Value)
		return false
	}
	if integ.TokenLiteral() != fmt.Sprintf("%d", value) {
		t.Errorf("integ.TokenLiteral not %d. got=%s", value,
			integ.TokenLiteral(),
		)
		return false
	}
	return true
}

func testIdentifier(t *testing.T, exp ast.Expression, value string) bool {
	ident, ok := exp.(*ast.Identifier)
	if !ok {
		t.Errorf("exp not *ast.Identifier. got=%T", exp)
		return false
	}

	if ident.Value != value {
		t.Errorf("ident.Value not %q. got=%q", value, ident.Value)
		return false
	}

	if ident.TokenLiteral() != value {
		t.Errorf("ident.TokenLiteral not %s. got=%s", value, ident.TokenLiteral())
		return false
	}

	return true
}

func testLiteralExpression(t *testing.T, exp ast.Expression, expected interface{}) bool {
	switch v := expected.(type) {
	case int:
		return testIntegerLiteral(t, exp, int64(v))
	case int64:
		return testIntegerLiteral(t, exp, v)
	case bool:
		return testBooleanLiteral(t, exp, v)
	case string:
		return testIdentifier(t, exp, v)
	}
	t.Errorf("type of exp not handled. got=%T", exp)
	return false
}

func testInfixExpression(t *testing.T, exp ast.Expression, left interface{}, operator string, right interface{}) bool {

	opExp, ok := exp.(*ast.InfixExpression)
	if !ok {
		t.Errorf("exp not *ast.InfixExpression. got=%T", opExp)
		return false
	}

	if !testLiteralExpression(t, opExp.Left, left) {
		return false
	}

	if opExp.Operator != operator {
		t.Errorf("opExp.Operator not %q. got=%q", operator, opExp.Operator)
		return false
	}

	if !testLiteralExpression(t, opExp.Right, right) {
		return false
	}

	return true
}

func testBooleanLiteral(t *testing.T, exp ast.Expression, value bool) bool {
	b, ok := exp.(*ast.Boolean)
	if !ok {
		t.Errorf("exp not *ast.Boolean. got=%T", b)
		return false
	}

	if b.Value != value {
		t.Errorf("b.Value not %t, got=%t", value, b.Value)
		return false
	}

	if b.TokenLiteral() != fmt.Sprintf("%t", value) {
		t.Errorf("b.TokenLiteral not %t. got=%s", value, b.TokenLiteral())
		return false
	}

	return true
}
