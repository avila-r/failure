package failure

import (
	"encoding"
	"strings"

	"github.com/avila-r/failure/id"
	"github.com/avila-r/failure/modifier"
	"github.com/avila-r/failure/trait"
)

type ErrorClass struct {
	Namespace ErrorNamespace
	Parent    *ErrorClass
	ID        uint64
	Name      string
	Traits    map[trait.Trait]bool
	Modifiers modifier.Modifiers
}

func (c *ErrorClass) Of(message string, v ...any) *Error {
	return c.New(message, v...)
}

func (c *ErrorClass) New(message string, v ...any) *Error {
	return Builder(c).
		Message(message, v...).
		Build()
}

func (c *ErrorClass) Blank() *Error {
	return Builder(c).
		Build()
}

func (c *ErrorClass) Wrap(err error, message string, v ...any) *Error {
	return Builder(c).
		Message(message, v...).
		Cause(err).
		Build()
}

func (c *ErrorClass) From(err error) *Error {
	return Builder(c).
		Cause(err).
		Build()
}

func Class(name string, traits ...trait.Trait) *ErrorClass {
	class := &ErrorClass{
		Namespace: DefaultNamespace,
		Parent:    nil,
		ID:        id.Next(),
		Name:      name,
		Traits: func() map[trait.Trait]bool {
			result := make(map[trait.Trait]bool)
			for trait := range DefaultNamespace.CollectTraits() {
				result[trait] = true
			}
			for _, trait := range traits {
				result[trait] = true
			}
			return result
		}(),
		Modifiers: modifier.Inherited(DefaultNamespace.Modifiers),
	}

	class.register()

	return class
}

func (c ErrorClass) Class(name string, traits ...trait.Trait) *ErrorClass {
	class := &ErrorClass{
		Namespace: c.Namespace,
		Parent:    &c,
		ID:        id.Next(),
		Name: func() string {
			if strings.Contains(c.Name, DefaultClass.Name) {
				return name
			} else {
				return c.Name + "." + name
			}
		}(),
		Traits: func() map[trait.Trait]bool {
			result := make(map[trait.Trait]bool)
			for trait := range c.Traits {
				result[trait] = true
			}
			for trait := range c.Namespace.CollectTraits() {
				result[trait] = true
			}
			for _, trait := range traits {
				result[trait] = true
			}
			return result
		}(),
		Modifiers: modifier.Inherited(c.Modifiers),
	}

	class.register()

	return class
}

func (c *ErrorClass) Is(other *ErrorClass) bool {
	current := c
	for current != nil {
		if current.ID == other.ID {
			return true
		}
		current = current.Parent
	}
	return false
}

func (c *ErrorClass) Has(trait trait.Trait) bool {
	_, ok := c.Traits[trait]
	return ok
}

func (c *ErrorClass) Apply(modifiers ...modifier.ClassModifier) *ErrorClass {
	new := modifier.Class(modifiers...)
	c.Modifiers = c.Modifiers.ReplaceWith(new)
	return c
}

func (c *ErrorClass) String() string {
	return c.Name
}

func (c *ErrorClass) RootNamespace() ErrorNamespace {
	n := c.Namespace
	for n.Parent != nil {
		n = *n.Parent
	}
	return n
}

func (c *ErrorClass) register() {
	Registry.mu.Lock()
	defer Registry.mu.Unlock()

	Registry.Classes = append(Registry.Classes, c)
	for _, s := range Registry.Listeners {
		s.OnClassCreated(c)
	}
}

var _ encoding.TextMarshaler = (*ErrorClass)(nil)

// MarshalText implements encoding.TextMarshaler
func (c *ErrorClass) MarshalText() (text []byte, err error) {
	return []byte(c.String()), nil
}
