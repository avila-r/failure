package failure

import (
	"github.com/avila-r/failure/modifier"
	"github.com/avila-r/failure/trait"
)

var (
	// CommonErrors is a namespace for general purpose errors designed for universal use.
	// These errors should typically be used in opaque manner, implying no handing in user code.
	// When handling is required, it is best to use custom error types with both standard and custom traits.
	CommonErrors = Namespace("common")

	// IllegalArgument is a type for invalid argument error
	IllegalArgument = CommonErrors.Class("illegal_argument")

	// IllegalState is a type for invalid state error
	IllegalState = CommonErrors.Class("illegal_state")

	// IllegalFormat is a type for invalid format error
	IllegalFormat = CommonErrors.Class("illegal_format")

	// InitializationFailed is a type for initialization error
	InitializationFailed = CommonErrors.Class("initialization_failed")

	// DataUnavailable is a type for unavailable data error
	DataUnavailable = CommonErrors.Class("data_unavailable")

	// UnsupportedOperation is a type for unsupported operation error
	UnsupportedOperation = CommonErrors.Class("unsupported_operation")

	// RejectedOperation is a type for rejected operation error
	RejectedOperation = CommonErrors.Class("rejected_operation")

	// Interrupted is a type for interruption error
	Interrupted = CommonErrors.Class("interrupted")

	// AssertionFailed is a type for assertion error
	AssertionFailed = CommonErrors.Class("assertion_failed")

	// InternalError is a type for internal error
	InternalError = CommonErrors.Class("internal_error")

	// ExternalError is a type for external error
	ExternalError = CommonErrors.Class("external_error")

	// ConcurrentUpdate is a type for concurrent update error
	ConcurrentUpdate = CommonErrors.Class("concurrent_update")

	// TimeoutElapsed is a type for timeout error
	TimeoutElapsed = CommonErrors.Class("timeout", trait.Timeout())

	// NotImplemented is an error type for lacking implementation
	NotImplemented = UnsupportedOperation.Class("not_implemented")

	// UnsupportedVersion is a type for unsupported version error
	UnsupportedVersion = UnsupportedOperation.Class("version")
)

var (
	// Most errors from this namespace are made private in order to disallow and direct type checks in the user code
	synthetic = Namespace("synthetic")

	// Private error type for non-errors errors, used as a not-nil substitute that cannot be type-checked directly
	foreignType = synthetic.Class("foreign")

	// Private error type used as a universal wrapper, meant to add nothing at all to the error apart from some message
	transparentWrapper = synthetic.Class("decorate").Apply(modifier.ClassModifierTransparent)

	// Private error type used as a densely opaque wrapper which hides both the original error and its own type
	_ = synthetic.Class("wrap")

	// Private error type used for stack trace capture
	_ = synthetic.Class("stacktrace").Apply(modifier.ClassModifierTransparent)
)
