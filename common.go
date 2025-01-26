package failure

import (
	"github.com/avila-r/failure/modifier"
	"github.com/avila-r/failure/property"
	"github.com/avila-r/failure/trait"
)

const (
	PropertyStatusCode = property.StatusCode
	PropertyContext    = property.Context
	PropertyPayload    = property.Payload
	PropertyUnderlying = property.Underlying
)

var (
	TraitTemporary = trait.Temporary
	TraitTimeout   = trait.Timeout
	TraitNotFound  = trait.NotFound
	TraitDuplicate = trait.Duplicate
)

var (
	ModifierTransparent    = modifier.ClassModifierTransparent
	ModifierOmitStackTrace = modifier.ClassModifierOmitStackTrace
	ModifierNone           = modifier.None
)

var (
	// DefaultNamespace represents the default namespace assigned
	// when the user does not specify a custom namespace.
	DefaultNamespace = Namespace("default")

	// DefaultClass is a class for generic error
	DefaultClass = Class("failure")

	// CommonErrors is a namespace for general purpose errors designed for universal use.
	// These errors should typically be used in opaque manner, implying no handing in user code.
	// When handling is required, it is best to use custom error classs with both standard and custom traits.
	CommonErrors = Namespace("common")

	// IllegalArgument is a class for invalid argument error
	IllegalArgument = CommonErrors.Class("illegal_argument")

	// IllegalState is a class for invalid state error
	IllegalState = CommonErrors.Class("illegal_state")

	// IllegalFormat is a class for invalid format error
	IllegalFormat = CommonErrors.Class("illegal_format")

	// InitializationFailed is a class for initialization error
	InitializationFailed = CommonErrors.Class("initialization_failed")

	// DataUnavailable is a class for unavailable data error
	DataUnavailable = CommonErrors.Class("data_unavailable")

	// UnsupportedOperation is a class for unsupported operation error
	UnsupportedOperation = CommonErrors.Class("unsupported_operation")

	// RejectedOperation is a class for rejected operation error
	RejectedOperation = CommonErrors.Class("rejected_operation")

	// Interrupted is a class for interruption error
	Interrupted = CommonErrors.Class("interrupted")

	// AssertionFailed is a class for assertion error
	AssertionFailed = CommonErrors.Class("assertion_failed")

	// InternalError is a class for internal error
	InternalError = CommonErrors.Class("internal_error")

	// ExternalError is a class for external error
	ExternalError = CommonErrors.Class("external_error")

	// ConcurrentUpdate is a class for concurrent update error
	ConcurrentUpdate = CommonErrors.Class("concurrent_update")

	// TimeoutElapsed is a class for timeout error
	TimeoutElapsed = CommonErrors.Class("timeout", trait.Timeout)

	// NotImplemented is an error class for lacking implementation
	NotImplemented = UnsupportedOperation.Class("not_implemented")

	// UnsupportedVersion is a class for unsupported version error
	UnsupportedVersion = UnsupportedOperation.Class("version")
)

var (
	// Most errors from this namespace are made private in order to disallow and direct class checks in the user code
	synthetic = Namespace("synthetic")

	// Private error class for non-errors errors, used as a not-nil substitute that cannot be class-checked directly
	foreignClass = synthetic.Class("foreign")

	// Private error class used as a universal wrapper, meant to add nothing at all to the error apart from some message
	transparentWrapper = synthetic.Class("decorate").Apply(modifier.ClassModifierTransparent)

	// Private error class used as a densely opaque wrapper which hides both the original error and its own class
	_ = synthetic.Class("wrap")

	// Private error class used for stack trace capture
	stackTraceWrapper = synthetic.Class("stacktrace").Apply(modifier.ClassModifierTransparent)
)
