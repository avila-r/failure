## failure.Error

`*failure.Error` implements the error interface and can be handled like a native error in your code.

```go
func (e *Error) Error() string {
	return e.Message
}
```

If you want to create native errors, use the failure.Err() function. You can create both static and dynamic errors as follows:
```go
// Static message, allowing error matching
err := failure.Err("error message!")
if err != nil {
	failure.Is(err, failure.Err("error message!")) // true
}

// Dynamic message, making error matching impossible
err := failure.Err("error message with arg %v!", 1)
if err != nil {
	failure.Is(err, failure.Err("error message with arg %v!", 2)) // false
}
```

Otherwise, `failure.Error` is a more powerful and versatile error type:
```go
err := failure.Of("message, arg %v", 1)
other := failure.Of("message, arg %v", 2)

if failure.As(err, other) { // True
	// ...
}
```

You can enhance your errors by decorating them:
```go
var (
	ErrNotFound = failure.New("user wasn't found")
)

func Find(id int, out *User) error {
	return ErrNotFound
}

func Delete(id int) error {
	user := User{}
	if err := Find(id, &user); err != nil {
		return failure.Decorate(err, "unable to delete user")
	}
}

//

if err := users.Delete(1); err != nil {
	log.Fatalf("Failed: %+v", err) // Includes stacktrace
}

```

You can organize your errors into namespaces, classes, and traits:
```go
var (
	UserErrors = failure.Class("user")
	
	ErrNotFound = UserErrors.New("user wasn't found")
)

if err := Find(); failure.Is(err, ErrNotFound) /* or ErrNotFound.Is(err) */ {
	// ...
}

// or

if err := Find(); failure.Extends(err, UserErrors) {
	// ...
}
```
```go
var (
	UserErrors = failure.Class("user")
	
	NotFound = UserErrors.Class("not_found", trait.NotFound)
)

Find := func() error {
	return NotFound.New("unable to find user with the provided email")
}

if err := Find(); failure.Extends(err, NotFound) {
	// ...
}

// or

if err := Find(); err != nil {
	return failure.Decorate(err, "failed to delete user")
}

// or

if err := Find(); failure.Has(err, trait.NotFound) {
	// ...
}
```

To enrich errors with additional metadata, use the `ErrorChain` methods:

```go
tags := failure.Tags{
	"critical": "true",
}

err := failure.Of("operation failed").
	Chain().
	Owner("admin").
	Public("user-facing message").
	Hint("check network connection").
	Trace("trace-id-123").
	Tags(tags).
	Done()

fmt.Printf("%+v\n", err)
```

Since `failure.Error` implements `slog.LogValuer`, it can be logged directly:

```go
import (
    "log/slog"
)

logger := slog.Default()

err := failure.Of("database connection failed").
	Chain().
	In("db").
	Owner("backend-team").
	Trace("trace-xyz").
	Done()

logger.Error("An error occurred", "error", err)
```


### Properties:

Properties can be used to encapsulate additional payload/context or details for your errors. For example:
```go
import (
	"github.com/avila-r/failure"
	"github.com/avila-r/failure/property"
)

var (
	ErrNotFound = failure.
		New("user not found").
		With(property.StatusCode, http.StatusNotFound) // Additional payload
)

func FindByID(id int) (*User, error) {
	return nil, ErrNotFound
}

if _, err := FindByID(id); err != nil {
	code := failure.Extract[int](err, property.StatusCode)
	// ...
}
```

`failure.Property(err error, key string)` returns a `property.Result`:
```go
type Result struct {
	Value any
	Ok    bool
}

func (r Result) Get() (value any, ok bool)

func (r Result) Bind(out any) bool
```

This allows for better result handling:

```go
result := failure.Property(err, "key")

value := 0
if ok := result.Bind(&value); !ok {
	// Handle
}
```

The function `failure.Extract[T]()` is also available to unwrap errors' properties:
```go
ErrNotFound := failure.
	Of("user not found").
	With(property.StatusCode, http.StatusNotFound)

code := failure.Extract[int](ErrNotFound, "code")

// ...
```
