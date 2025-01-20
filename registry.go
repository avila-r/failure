package failure

import "sync"

var Registry = struct {
	Listeners  []RegistryListener
	Namespaces []ErrorNamespace
	Classes    []*ErrorClass

	mu sync.Mutex
}{}

type RegistryListener interface {
	// OnNamespaceCreated is called exactly once for each namespace
	OnNamespaceCreated(namespace ErrorNamespace)

	// OnClassCreated is called exactly once for each class
	OnClassCreated(t *ErrorClass)
}

func Subscribe(listener RegistryListener) {
	for _, namespace := range Registry.Namespaces {
		listener.OnNamespaceCreated(namespace)
	}

	for _, class := range Registry.Classes {
		listener.OnClassCreated(class)
	}

	Registry.Listeners = append(Registry.Listeners, listener)
}
