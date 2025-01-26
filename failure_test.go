package failure_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/avila-r/failure"
	"github.com/avila-r/failure/modifier"
	"github.com/avila-r/failure/trait"
)

var (
	TestNamespace              = failure.Namespace("foo")
	TestClass                  = TestNamespace.Class("bar")
	TestClassSilent            = TestClass.Class("silent").Apply(modifier.ClassModifierOmitStackTrace)
	TestClassTransparent       = TestClass.Class("transparent").Apply(modifier.ClassModifierTransparent)
	TestClassSilentTransparent = TestClass.Class("silent_transparent").Apply(modifier.ClassModifierTransparent, modifier.ClassModifierOmitStackTrace)
	TestSubtype0               = TestClass.Class("internal")
	TestSubtype1               = TestSubtype0.Class("wat")
	TestClassBar1              = TestNamespace.Class("bar1")
	TestClassBar2              = TestNamespace.Class("bar2")
)

func Test_Err(t *testing.T) {
	cases := []struct {
		name     string
		message  string
		args     []any
		expected string
	}{
		{"SimpleMessage", "error occurred", nil, "error occurred"},
		{"FormattedMessage", "error %d occurred", []any{404}, "error 404 occurred"},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if err := failure.Err(test.message, test.args...); err.Error() != test.expected {
				t.Errorf("expected %v, got %v", test.expected, err.Error())
			}
		})
	}
}

func Test_New(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		args     []any
		expected string
	}{
		{"SimpleError", "new error", nil, "new error"},
		{"FormattedError", "error %s", []any{"test"}, "error test"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := failure.New(test.message, test.args...); err.Error() != test.expected {
				t.Errorf("expected %v, got %v", test.expected, err.Error())
			}
		})
	}
}

func Test_From(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"NilError", nil},
		{"NonFailureError", errors.New("base error")},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if result := failure.From(test.err); result != nil {
				t.Errorf("expected nil, got %v", result)
			}
		})
	}
}

func Test_Cast(t *testing.T) {
	base, custom := errors.New("base error"), failure.Of("custom error")

	tests := []struct {
		name     string
		err      error
		expected *failure.Error
	}{
		{"NilError", nil, nil},
		{"NonFailureError", base, nil},
		{"FailureError", custom, custom},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := failure.Cast(test.err)
			if result != test.expected {
				t.Errorf("expected result to be %v, got %v", test.expected, result)
			}
		})
	}
}

func Test_Decorate(t *testing.T) {
	base := errors.New("base error")

	tests := []struct {
		name     string
		err      error
		message  string
		args     []any
		expected string
	}{
		{"WithDecoration", base, "decorated %s", []any{"message"}, "decorated message, cause: base error"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := failure.Decorate(test.err, test.message, test.args...)
			if err.Summary() != test.expected {
				t.Errorf("expected %v, got %v", test.expected, err.Summary())
			}
		})
	}
}

func Test_EnhanceStackTrace(t *testing.T) {
}

func Test_EnsureStackTrace(t *testing.T) {
}

func Test_Extends(t *testing.T) {
}

func Test_Is(t *testing.T) {
	base := errors.New("base error")
	wrapped := fmt.Errorf("wrapped: %w", base)

	tests := []struct {
		name   string
		err    error
		target error
		expect bool
	}{
		{"SameError", base, base, true},
		{"WrappedError", wrapped, base, true},
		{"DifferentError", base, errors.New("another error"), false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := failure.Is(test.err, test.target)
			if result != test.expect {
				t.Errorf("expected %v, got %v", test.expect, result)
			}
		})
	}
}

func Test_Has(t *testing.T) {
	var (
		custom = trait.New("custom")
		class  = failure.Class("class", custom)
		err    = class.New("error").With("key", "value")
	)

	tests := []struct {
		name   string
		err    error
		trait  trait.Trait
		expect bool
	}{
		{"HasTrait", err, custom, true},
		{"NoTrait", err, trait.New("other"), false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := failure.Has(test.err, test.trait)
			if result != test.expect {
				t.Errorf("expected %v, got %v", test.expect, result)
			}
		})
	}
}

func Test_As(t *testing.T) {
}

func Test_Cause(t *testing.T) {
}

func Test_Unwrap(t *testing.T) {
}
