package failure

import (
	"fmt"

	"github.com/avila-r/failure/id"
	"github.com/avila-r/failure/modifier"
	"github.com/avila-r/failure/trait"
)

type ErrorNamespace struct {
	Parent    *ErrorNamespace
	ID        uint64
	Name      string
	Traits    []trait.Trait
	Modifiers modifier.Modifiers
}

func Namespace(name string, traits ...trait.Trait) ErrorNamespace {
	namespace := ErrorNamespace{
		Parent:    nil,
		ID:        id.Next(),
		Name:      name,
		Traits:    append([]trait.Trait{}, traits...),
		Modifiers: modifier.None,
	}

	namespace.register()

	return namespace
}

func (n ErrorNamespace) Namespace(name string, traits ...trait.Trait) ErrorNamespace {
	namespace := ErrorNamespace{
		Parent:    &n,
		ID:        id.Next(),
		Name:      fmt.Sprintf("%s.%s", n.Name, name),
		Traits:    append([]trait.Trait{}, traits...),
		Modifiers: modifier.Inherited(n.Modifiers),
	}

	namespace.register()

	return namespace
}

func (n ErrorNamespace) Apply(modifiers ...modifier.ClassModifier) ErrorNamespace {
	new := modifier.Class(modifiers...)
	n.Modifiers = n.Modifiers.ReplaceWith(new)
	return n
}

func (n ErrorNamespace) Class(name string, traits ...trait.Trait) *ErrorClass {
	return Class(n, name, traits...)
}

func (n ErrorNamespace) Contains(class *ErrorClass) bool {
	other := &class.Namespace

	for other != nil {
		if n.ID == other.ID {
			return true
		}
		other = other.Parent
	}

	return false
}

func (n ErrorNamespace) CollectTraits() map[trait.Trait]bool {
	result := make(map[trait.Trait]bool)

	namespace := &n
	for namespace != nil {
		for _, trait := range namespace.Traits {
			result[trait] = true
		}
		namespace = namespace.Parent
	}

	return result
}

func (n ErrorNamespace) register() {
	Registry.mu.Lock()
	defer Registry.mu.Unlock()

	Registry.Namespaces = append(Registry.Namespaces, n)
	for _, s := range Registry.Listeners {
		s.OnNamespaceCreated(n)
	}
}
