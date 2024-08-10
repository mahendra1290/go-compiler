package evaluator

import (
	"bytes"
	"fmt"
	"monkey/src/ast"
	"monkey/src/object"
	"strings"
)

var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

func Eval(node ast.Node, env *object.Environment, buffer *bytes.Buffer) object.Object {
	switch node := node.(type) {
	case *ast.IntegerLiteral:

		return &object.Integer{Value: node.Value}

	case *ast.StringLiteral:
		return &object.String{Value: node.Value}

	case *ast.Boolean:
		return nativeBoolToBooleanObject(node.Value)

	case *ast.PrefixExpression:
		right := Eval(node.Right, env, buffer)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)

	case *ast.InfixExpression:
		left := Eval(node.Left, env, buffer)
		if isError(left) {
			return left
		}
		right := Eval(node.Right, env, buffer)
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right)

	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue, env, buffer)
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}

	case *ast.Program:
		return evalProgram(node.Statements, env, buffer)

	case *ast.BlockStatement:
		return evalBlockStatement(node, env, buffer)

	case *ast.IfExpression:
		return evalIfExpression(node, env, buffer)

	case *ast.ExpressionStatement:
		return Eval(node.Expression, env, buffer)

	case *ast.LetStatement:
		val := Eval(node.Value, env, buffer)
		if isError(val) {
			return val
		}
		env.Set(node.Name.Value, val)

	case *ast.Identifier:
		return evalIdentifier(node, env)

	case *ast.FunctionLiteral:
		params := node.Parameters
		body := node.Body
		return &object.Function{Parameters: params, Env: env, Body: body}

	case *ast.CallExpression:
		function := Eval(node.Function, env, buffer)
		if isError(function) {
			return function
		}
		args := evalExpressions(node.Arguments, env, buffer)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}

		return applyFunction(function, args, buffer)

	case *ast.ArrayLiteral:
		elements := evalExpressions(node.Elements, env, buffer)
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}

		return &object.Array{Elements: elements}

	case *ast.IndexExpression:
		left := Eval(node.Left, env, buffer)
		if isError(left) {
			return left
		}
		index := Eval(node.Index, env, buffer)
		if isError(index) {
			return index
		}

		return evalIndexExpression(left, index)

	case *ast.AssignStatement:

		val := Eval(node.Value, env, buffer)
		if isError(val) {
			return val
		}

		ok := env.UpdateValue(node.Variable.Value, val)
		if !ok {
			return newError("invalid assignment to non declared identifier %s", node.Variable.Value)
		}

	case *ast.IndexAssignmentExpression:
		index := Eval(node.Index.Index, env, buffer)
		if isError(index) {
			return index
		}

		val := Eval(node.Value, env, buffer)
		if isError(val) {
			return val
		}

		left := Eval(node.Index.Left, env, buffer)

		return evalIndexAssignmentExpression(left, index, val)

	case *ast.ForStatement:
		iterator := Eval(node.Iterator, env, buffer)
		if isError(iterator) {
			return iterator
		}

		forEnv := object.NewEnclosedEnvironement(env)

		switch {
		case iterator.Type() == object.ARRAY_OBJ:
			arr := iterator.(*object.Array)
			for i, v := range arr.Elements {

				forEnv.Set(node.Index.Value, &object.Integer{Value: int64(i)})
				forEnv.Set(node.Value.Value, v)

				evalBlockStatement(node.Block, forEnv, buffer)

			}
		case iterator.Type() == object.STRING_OBJ:
			str := iterator.(*object.String)
			for i, v := range str.Value {
				forEnv.Set(node.Index.Value, &object.Integer{Value: int64(i)})
				forEnv.Set(node.Value.Value, &object.String{Value: string(v)})

				evalBlockStatement(node.Block, forEnv, buffer)
			}

		case iterator.Type() == object.HASH_OBJ:
			pairs := iterator.(*object.Hash)
			for _, v := range pairs.Pairs {
				forEnv.Set(node.Index.Value, v.Key)
				forEnv.Set(node.Value.Value, v.Value)

				evalBlockStatement(node.Block, forEnv, buffer)
			}

		default:
			return newError("for iterator must resolve to array, string or hash got %T", iterator)
		}

		return NULL

	case *ast.HashLiteral:
		return evalHashLiteral(node, env, buffer)
	}

	return nil
}

func applyFunction(fn object.Object, args []object.Object, buffer *bytes.Buffer) object.Object {
	switch fn := fn.(type) {
	case *object.Function:
		extendedEnv := extendFunctionEnv(fn, args)
		evaluated := Eval(fn.Body, extendedEnv, buffer)
		return unwrapReturnValue(evaluated)
	case *object.Builtin:
		if fn.Name == "puts" {
			var put object.BuiltinFunction = func(args ...object.Object) object.Object {
				values := []string{}
				for _, arg := range args {
					values = append(values, arg.Inspect())
				}

				buffer.WriteString(strings.Join(values, ", ") + "\n")

				return NULL
			}
			if result := put(args...); result != nil {
				return result
			}
			return NULL
		}
		if result := fn.Fn(args...); result != nil {
			return result
		}
		return NULL
	default:
		return newError("not a function: %s", fn.Type())
	}

}

func extendFunctionEnv(fn *object.Function, args []object.Object) *object.Environment {
	env := object.NewEnclosedEnvironement(fn.Env)

	for paramIdx, param := range fn.Parameters {
		env.Set(param.Value, args[paramIdx])
	}

	return env
}

func unwrapReturnValue(obj object.Object) object.Object {
	if returnValue, ok := obj.(*object.ReturnValue); ok {
		return returnValue.Value
	}

	return obj
}

func evalExpressions(exps []ast.Expression, env *object.Environment, buffer *bytes.Buffer) []object.Object {
	var result []object.Object

	for _, e := range exps {
		evaluated := Eval(e, env, buffer)
		if isError(evaluated) {
			return []object.Object{evaluated}
		}

		result = append(result, evaluated)
	}

	return result
}

func evalIfExpression(ie *ast.IfExpression, env *object.Environment, buffer *bytes.Buffer) object.Object {
	condition := Eval(ie.Condition, env, buffer)
	if isError(condition) {
		return condition
	}
	if isTruthy(condition) {
		return Eval(ie.Consequence, env, buffer)
	} else if ie.Alternative != nil {
		return Eval(ie.Alternative, env, buffer)
	} else {
		return NULL
	}
}

func isTruthy(obj object.Object) bool {
	switch obj {
	case NULL:
		return false
	case TRUE:
		return true
	case FALSE:
		return false
	default:
		return true
	}
}

func evalProgram(stmts []ast.Statement, env *object.Environment, buffer *bytes.Buffer) object.Object {
	var result object.Object

	for _, stmt := range stmts {
		result = Eval(stmt, env, buffer)

		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}

	}

	return result
}

func evalBlockStatement(block *ast.BlockStatement, env *object.Environment, buffer *bytes.Buffer) object.Object {
	var result object.Object

	for _, stmt := range block.Statements {
		result = Eval(stmt, env, buffer)

		if result != nil && (result.Type() == object.RETURN_VALUE_OBJ || result.Type() == object.ERROR_OBJ) {
			return result
		}
	}

	return result

}

func nativeBoolToBooleanObject(value bool) *object.Boolean {
	if value {
		return TRUE
	}
	return FALSE
}

func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusPrefixOperatorExpression(right)
	default:
		return newError("unknown operator: %s%s", operator, right.Type())
	}
}

func evalBangOperatorExpression(right object.Object) object.Object {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		return FALSE
	}
}

func evalMinusPrefixOperatorExpression(right object.Object) object.Object {
	if right.Type() != object.INTEGER_OBJ {
		return newError("unknown operator: -%s", right.Type())
	}

	value := right.(*object.Integer).Value
	return &object.Integer{Value: -value}
}

func evalInfixExpression(operator string, left object.Object, right object.Object) object.Object {

	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(operator, left, right)
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return evalStringInfixExpression(operator, left, right)
	case operator == "==":
		return nativeBoolToBooleanObject(left == right)
	case operator == "!=":
		return nativeBoolToBooleanObject(left != right)
	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s", left.Type(), operator, right.Type())
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}

}

func evalIntegerInfixExpression(operator string, left object.Object, right object.Object) object.Object {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value
	switch operator {
	case "+":
		return &object.Integer{Value: leftVal + rightVal}
	case "-":
		return &object.Integer{Value: leftVal - rightVal}
	case "*":
		return &object.Integer{Value: leftVal * rightVal}
	case "/":
		return &object.Integer{Value: leftVal / rightVal}
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalStringInfixExpression(operator string, left object.Object, right object.Object) object.Object {
	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value
	switch operator {
	case "+":
		return &object.String{Value: leftVal + rightVal}
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	if val, ok := env.Get(node.Value); ok {
		return val
	}
	if builtin, ok := builtins[node.Value]; ok {
		return builtin
	}

	return newError("identifier not found: " + node.Value)
}

func evalIndexExpression(left, index object.Object) object.Object {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return evalArrayIndexExpression(left, index)
	case left.Type() == object.HASH_OBJ:
		return evalHashIndexExpression(left, index)
	default:
		return newError("index operator not supported: %s", left.Type())
	}
}

func evalIndexAssignmentExpression(left, index, value object.Object) object.Object {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return evalArrayIndexAssignmentExpression(left, index, value)

	case left.Type() == object.HASH_OBJ:
		return evalHashIndexAssignmnetExpression(left, index, value)

	default:
		return newError("index assignemnt not supported: %s", left.Type())
	}

}

func evalArrayIndexAssignmentExpression(left, index, value object.Object) object.Object {
	arr := left.(*object.Array)

	idx := index.(*object.Integer).Value

	if idx >= 0 && idx < int64(len(arr.Elements)) {
		arr.Elements[idx] = value

		return value
	}

	return newError("index out of range: got = %d for array of size = %d", idx, len(arr.Elements))

}

func evalHashLiteral(node *ast.HashLiteral, env *object.Environment, buffer *bytes.Buffer) object.Object {
	pairs := make(map[object.HashKey]object.HashPair)

	for keyNode, valueNode := range node.Pairs {
		key := Eval(keyNode, env, buffer)
		if isError(key) {
			return key
		}

		hashKey, ok := key.(object.Hashable)
		if !ok {
			return newError("unusable as hask key: %s", key.Type())
		}

		value := Eval(valueNode, env, buffer)
		if isError(value) {
			return value
		}

		hashed := hashKey.HashKey()
		pairs[hashed] = object.HashPair{Key: key, Value: value}

	}

	return &object.Hash{Pairs: pairs}
}

func evalArrayIndexExpression(array, index object.Object) object.Object {
	arrayObject := array.(*object.Array)
	idx := index.(*object.Integer).Value

	max := int64(len(arrayObject.Elements) - 1)

	if idx < 0 || idx > max {
		return NULL
	}

	return arrayObject.Elements[idx]
}

func evalHashIndexExpression(hash, index object.Object) object.Object {
	hashObject := hash.(*object.Hash)
	key, ok := index.(object.Hashable)
	if !ok {
		return newError("unusable as hash key: %s", index.Type())
	}
	pair, ok := hashObject.Pairs[key.HashKey()]
	if !ok {
		return NULL
	}

	return pair.Value
}

func evalHashIndexAssignmnetExpression(hash, index, val object.Object) object.Object {
	hashObject := hash.(*object.Hash)
	key, ok := index.(object.Hashable)
	if !ok {
		return newError("unusable as hash key: %s", index.Type())
	}
	hashObject.Pairs[key.HashKey()] = object.HashPair{
		Key:   index,
		Value: val,
	}

	return NULL
}

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}

	return false
}
