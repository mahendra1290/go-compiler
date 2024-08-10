package object

type Environment struct {
	store map[string]Object
	outer *Environment
}

func NewEnclosedEnvironement(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer

	return env
}

func NewEnvironment() *Environment {
	s := make(map[string]Object)
	return &Environment{store: s}
}

func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		obj, ok = e.outer.Get(name)
	}
	return obj, ok
}

func (e *Environment) Set(name string, val Object) Object {
	e.store[name] = val
	return val
}

func (e *Environment) UpdateValue(name string, val Object) bool {

	var setValue func(env *Environment) bool
	setValue = func(env *Environment) bool {
		_, ok := env.store[name]
		if !ok && env.outer != nil {
			return setValue(e.outer)
		}

		if !ok {
			return false
		}

		env.Set(name, val)

		return true
	}

	return setValue(e)
}
