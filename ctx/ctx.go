package ctx

import "reflect"

type Context map[string]any

func Evaluated(ctx Context) Context {
	for key, value := range ctx {
		switch v := value.(type) {
		case map[string]any:
			ctx[key] = Evaluated(v)
		default:
			val := reflect.ValueOf(value)
			if !val.IsValid() || val.Kind() != reflect.Func {
				ctx[key] = value
			} else if val.Type().NumIn() != 0 || val.Type().NumOut() != 1 {
				ctx[key] = value
			}

			ctx[key] = val.Call([]reflect.Value{})[0].Interface()
		}
	}

	for key, value := range ctx {
		val := reflect.ValueOf(value)
		for val.Kind() == reflect.Ptr && !val.IsNil() {
			val = val.Elem()
		}
		ctx[key] = val.Interface()
	}

	return ctx
}
