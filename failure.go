package failure

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/avila-r/failure/ctx"
	"github.com/avila-r/failure/property"
	"github.com/avila-r/failure/stacktrace"
	"github.com/avila-r/failure/trait"
)

func Err(message string, v ...any) error {
	if len(v) > 0 {
		return fmt.Errorf(message, v...)
	}

	return errors.New(message)
}

func Of(message string, v ...any) *Error {
	return New(message, v...)
}

func New(message string, v ...any) *Error {
	return Builder(DefaultClass).
		Message(message, v...).
		Build()
}

func Blank() *Error {
	return Builder(DefaultClass).
		Build()
}

func From(err error) *Error {
	return Cast(err)
}

// Cast attempts to cast an error to errorx Type, returns nil if cast has failed.
func Cast(err error) *Error {
	if e, ok := err.(*Error); ok && e != nil {
		return e
	}

	return nil
}

func Decorate(err error, message string, v ...any) *Error {
	return Builder(transparentWrapper).
		Message(message, v...).
		Cause(err).
		StackTrace().
		Build()
}

func Enhance(err error, message string, v ...any) *Error {
	return Builder(transparentWrapper).
		Message(message, v...).
		Cause(err).
		EnhanceStackTrace().
		Build()
}

func Enhanced(err error) *Error {
	if casted := Cast(err); casted != nil && casted.stacktrace != nil {
		return casted
	}

	return Builder(stackTraceWrapper).
		Message("").
		Cause(err).
		EnhanceStackTrace().
		Build()
}

func Extends(err error, c *ErrorClass) bool {
	casted := func() *Error {
		raw := err
		for raw != nil {
			typed := Cast(raw)
			if typed != nil {
				return typed
			}
			raw = errors.Unwrap(raw)
		}

		return nil
	}()

	return casted != nil && casted.Extends(c)
}

func Has(err error, trait trait.Trait) bool {
	if casted := Cast(err); casted != nil {
		return casted.Has(trait)
	}

	return false
}

func Is(err, target error) bool {
	if target == nil {
		return err == target
	}

	for {
		if reflect.TypeOf(target).Comparable() && err == target {
			return true
		}
		if x, ok := err.(interface{ Is(error) bool }); ok && x.Is(target) {
			return true
		}
		if err = Unwrap(err); err == nil {
			return false
		}
	}
}

func As(err error, target any) bool {
	if errors.As(err, &target) {
		return true
	}

	if target == nil || err == nil {
		return false
	}

	val := reflect.ValueOf(target)
	typ := val.Type()

	// target must be a non-nil pointer
	if typ.Kind() != reflect.Ptr || val.IsNil() {
		return false
	}

	// *target must be interface or implement error
	if typ != reflect.TypeOf(&Error{}) && reflect.TypeOf(err).AssignableTo(typ.Elem()) {
		return false
	}

	for {
		typeof := reflect.TypeOf(err)
		if typeof != reflect.TypeOf(&Error{}) && reflect.TypeOf(err).AssignableTo(typ.Elem()) {
			val.Elem().Set(reflect.ValueOf(err))
			return true
		}
		if x, ok := err.(interface{ As(any) bool }); ok && x.As(target) {
			return true
		}
		if err = Unwrap(err); err == nil {
			return false
		}
	}
}

func Cause(err error) error {
	for {
		previous := Unwrap(err)
		if previous == nil {
			return err
		}
		err = previous
	}
}

func Property(err error, key string) property.Result {
	if err := Cast(err); err != nil {
		return err.Property(key)
	}

	return property.Empty()
}

func Extract[T any](err error, key string) (out T) {
	if err := Cast(err); err != nil {
		err.Property(key).Bind(&out)
	}

	return
}

func Contains(err error, key string) bool {
	if err := Cast(err); err != nil {
		return err.Property(key).Ok
	}

	return false
}

func Inspect(err error) string {
	if err := Cast(err); err != nil {
		return err.Summary()
	}

	return err.Error()
}

func Unwrap(err error) error {
	u, ok := err.(interface {
		Unwrap() error
	})

	if !ok {
		return nil
	}

	return u.Unwrap()
}

func Must[T any](v T, err error) T {
	if err != nil {
		panic(err.Error())
	}
	return v
}

func Try(f func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			switch v := r.(type) {
			case error:
				err = Err("recovered value: %w", v)
			default:
				err = Err("recovered value: %v", r)
			}
		}
	}()
	f()
	return
}

func Pie(err error) {
	if err != nil {
		panic(err)
	}
}

func Deref[T any](p *T, def ...T) (t T) {
	if p != nil {
		return *p
	}
	if len(def) > 0 {
		return def[0]
	}
	return
}

func Trait(label string) trait.Trait {
	return trait.New(label)
}

func Deep[T comparable](err *Error, getter func(*Error) T) T {
	if err.cause == nil {
		return getter(err)
	}

	if casted := Cast(err.cause); casted != nil {
		return func(values ...T) T {
			var zero T
			for _, v := range values {
				if v != zero {
					return v
				}
			}
			return zero
		}(Deep(casted, getter), getter(err))
	}

	return getter(err)
}

func Recurse(err *Error, do func(*Error)) {
	do(err)

	if err.cause == nil {
		return
	}

	if casted := Cast(err.cause); casted != nil {
		Recurse(casted, do)
	}
}

func Gather(err *Error, getter func(*Error) ctx.Context) ctx.Context {
	if err.cause == nil {
		return getter(err)
	}

	if casted := Cast(err.cause); casted != nil {
		return func(maps ...ctx.Context) ctx.Context {
			count := 0
			for i := range maps {
				count += len(maps[i])
			}

			out := make(ctx.Context, count)
			for i := range maps {
				for k := range maps[i] {
					out[k] = maps[i][k]
				}
			}

			return out
		}(ctx.Context{}, getter(err), Gather(casted, getter))
	}

	return getter(err)
}

func InitializeStackTraceTransformer(subtransformer stacktrace.FilePathTransformer) (stacktrace.FilePathTransformer, error) {
	stacktrace.Transformer.Mu.Lock()
	defer stacktrace.Transformer.Mu.Unlock()

	old := stacktrace.Transformer.Transform.Load().(stacktrace.FilePathTransformer)
	stacktrace.Transformer.Transform.Store(subtransformer)

	if stacktrace.Transformer.Initialized {
		return old, InitializationFailed.New("stack trace transformer was already set up: %#v", old)
	}

	stacktrace.Transformer.Initialized = true
	return nil, nil
}
