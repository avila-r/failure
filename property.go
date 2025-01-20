package failure

import "reflect"

var (
	PropertyStatusCode = "code"
	PropertyContext    = "context"
	PropertyPayload    = "payload"
	PropertyUnderlying = "underlying"
)

// Properties represents map of properties.
// Compared to builtin type, it uses less allocations and reallocations on copy.
// It is implemented as a simple linked list.
type Properties struct {
	key   string
	value any
	next  *Properties
}

func (p *Properties) Set(key string, value any) *Properties {
	return &Properties{key: key, value: value, next: p}
}

func (p *Properties) Get(key string) (value any, ok bool) {
	for p != nil {
		if p.key == key {
			return p.value, true
		}
		p = p.next
	}
	return nil, false
}

type PropertyResult struct {
	Value any
	Ok    bool
}

func (r PropertyResult) Get() (value any, ok bool) {
	return r.Value, r.Ok
}

func (r PropertyResult) Bind(out any) bool {
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
