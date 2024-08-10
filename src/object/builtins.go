package object

import "fmt"

var Builtins = []struct {
	Name    string
	Builtin *Builtin
}{
	{
		"len",
		&Builtin{
			Name: "len",
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return newError("wrong number of arguments to `len`. got=%d, want=1", len(args))
				}

				switch arg := args[0].(type) {
				case *String:
					return &Integer{Value: int64(len(arg.Value))}
				case *Array:
					return &Integer{Value: int64(len(arg.Elements))}
				case *Hash:
					return &Integer{Value: int64(len(arg.Pairs))}
				default:
					return newError("argument to `len` not supported, got=%s", args[0].Type())
				}
			},
		},
	},
	{"first",
		&Builtin{
			Name: "first",
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return newError("wrong number of arguments to `len`. got=%d, want=1", len(args))
				}

				if args[0].Type() != ARRAY_OBJ {
					return newError("argument to `first` must be ARRAY, got %s", args[0].Type())
				}

				arr := args[0].(*Array)
				if len(arr.Elements) > 0 {
					return arr.Elements[0]
				}

				return nil
			},
		}},
	{
		"last",
		&Builtin{
			Name: "last",
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return newError("wrong number of arguments to `len`. got=%d, want=1", len(args))
				}

				if args[0].Type() != ARRAY_OBJ {
					return newError("argument to `last` must be ARRAY, got %s", args[0].Type())
				}

				arr := args[0].(*Array)
				if len(arr.Elements) > 0 {
					return arr.Elements[len(arr.Elements)-1]
				}

				return nil
			},
		}},
	{
		"rest",
		&Builtin{
			Name: "rest",
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return newError("wrong number of arguments to `len`. got=%d, want=1", len(args))
				}

				if args[0].Type() != ARRAY_OBJ {
					return newError("argument to `rest` must be ARRAY, got %s", args[0].Type())
				}

				arr := args[0].(*Array)
				if len(arr.Elements) > 0 {
					return &Array{
						Elements: arr.Elements[1:],
					}
				}

				return nil
			},
		}},
	{
		"push",
		&Builtin{
			Name: "push",
			Fn: func(args ...Object) Object {
				if len(args) != 2 {
					return newError("wrong number of arguments to `len`. got=%d, want=2", len(args))
				}

				if args[0].Type() != ARRAY_OBJ {
					return newError("first argument to `push` must be ARRAY, got %s", args[0].Type())
				}

				arr := args[0].(*Array)
				return &Array{
					Elements: append(arr.Elements, args[1]),
				}

			},
		}}, {
		"puts",
		&Builtin{
			Name: "puts",
			Fn: func(args ...Object) Object {
				for _, arg := range args {
					fmt.Println(arg.Inspect())
				}

				return nil
			},
		}},
	{
		"range",
		&Builtin{
			Name: "range",
			Fn: func(args ...Object) Object {
				if len(args) != 2 {
					return newError("wrong number of arguments to `range`. got=%d, want=2", len(args))
				}

				if args[0].Type() != INTEGER_OBJ || args[1].Type() != INTEGER_OBJ {
					return newError("arg must be INTEGERS")
				}

				start := args[0].(*Integer).Value
				end := args[1].(*Integer).Value

				arr := make([]Object, end-start)

				for i := range arr {
					arr[i] = &Integer{Value: start + int64(i)}
				}

				return &Array{
					Elements: arr,
				}
			},
		},
	},
}

func newError(format string, a ...interface{}) *Error {
	return &Error{Message: fmt.Sprintf(format, a...)}
}

func GetBuiltinByName(name string) *Builtin {
	for _, b := range Builtins {
		if b.Name == name {
			return b.Builtin
		}
	}

	return nil
}
