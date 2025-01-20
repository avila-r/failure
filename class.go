package failure

import (
	"encoding"

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

func Class(namespace ErrorNamespace, name string, traits ...trait.Trait) *ErrorClass {
	class := &ErrorClass{
		Namespace: namespace,
		Parent:    nil,
		ID:        id.Next(),
		Name:      namespace.Name + "." + name,
		Traits: func() map[trait.Trait]bool {
			result := make(map[trait.Trait]bool)
			for trait := range namespace.CollectTraits() {
				result[trait] = true
			}
			for _, trait := range traits {
				result[trait] = true
			}
			return result
		}(),
		Modifiers: modifier.Inherited(namespace.Modifiers),
	}

	class.register()

	return class
}

func (c ErrorClass) Class(name string, traits ...trait.Trait) *ErrorClass {
	class := &ErrorClass{
		Namespace: c.Namespace,
		Parent:    &c,
		ID:        id.Next(),
		Name:      c.Name + "." + name,
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
