package property

import (
	"reflect"
)

const (
	StatusCode = "code"
	Context    = "context"
	Payload    = "payload"
	Underlying = "underlying"
)

// List represents map of properties.
// Compared to builtin type, it uses less allocations and reallocations on copy.
// It is implemented as a simple linked list.
type List struct {
	Key   string
	Value any
	Next  *List
}

func (p *List) Set(key string, value any) *List {
	return &List{Key: key, Value: value, Next: p}
}

func (p *List) Get(key string) (value any, ok bool) {
	for p != nil {
		if p.Key == key {
			return p.Value, true
		}
		p = p.Next
	}
	return nil, false
}

type Result struct {
	Value any
	Ok    bool
}

func Empty() Result {
	return Result{
		Value: nil,
		Ok:    false,
	}
}

func (r Result) Get() (value any, ok bool) {
	return r.Value, r.Ok
}

func (r Result) Bind(out any) bool {
	if !r.Ok {
		return false
	}

	v := reflect.ValueOf(out)
	if v.Kind() != reflect.Ptr {
		return false
	}

	target := v.Elem()
	if target.CanSet() {
		bind := reflect.ValueOf(r.Value)
		target.Set(bind)
		return true
	}

	return false
}
