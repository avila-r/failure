package modifier

type ClassModifier int

const (
	// ClassModifierTransparent is a type modifier; an error type with such modifier creates transparent wrappers by default
	ClassModifierTransparent ClassModifier = 1
	// ClassModifierOmitStackTrace is a type modifier; an error type with such modifier omits the stack trace collection upon creation of an error instance
	ClassModifierOmitStackTrace ClassModifier = 2
)

type Modifiers interface {
	CollectStackTrace() bool
	Transparent() bool
	ReplaceWith(new Modifiers) Modifiers
}

type (
	NoModifiers struct{}
)

var (
	None Modifiers = NoModifiers{}
)

// CollectStackTrace implements Modifiers.
func (n NoModifiers) CollectStackTrace() bool {
	return true
}

// ReplaceWith implements Modifiers.
func (n NoModifiers) ReplaceWith(new Modifiers) Modifiers {
	return new
}

// Transparent implements Modifiers.
func (n NoModifiers) Transparent() bool {
	return false
}

type (
	ClassModifiers struct {
		OmitStackTrace bool
		IsTransparent  bool
	}
)

var (
	_ Modifiers = ClassModifiers{}
)

func Class(modifiers ...ClassModifier) Modifiers {
	m := ClassModifiers{}

	for _, modifier := range modifiers {
		switch modifier {
		case ClassModifierOmitStackTrace:
			m.OmitStackTrace = true
		case ClassModifierTransparent:
			m.IsTransparent = true
		}
	}

	return m
}

// CollectStackTrace implements Modifiers.
func (c ClassModifiers) CollectStackTrace() bool {
	return !c.OmitStackTrace
}

// ReplaceWith implements Modifiers.
func (c ClassModifiers) ReplaceWith(new Modifiers) Modifiers {
	panic("attempt to modify class modifiers more than once")
}

// Transparent implements Modifiers.
func (c ClassModifiers) Transparent() bool {
	return c.IsTransparent
}

type (
	InheritedModifiers struct {
		Parent   Modifiers
		Override Modifiers
	}
)

var (
	_ Modifiers = InheritedModifiers{}
)

func Inherited(modifiers Modifiers) Modifiers {
	if _, ok := modifiers.(NoModifiers); ok {
		return None
	}

	return InheritedModifiers{
		Parent:   modifiers,
		Override: None,
	}
}

// CollectStackTrace implements Modifiers.
func (i InheritedModifiers) CollectStackTrace() bool {
	return i.Parent.CollectStackTrace() && i.Override.CollectStackTrace()
}

// ReplaceWith implements Modifiers.
func (i InheritedModifiers) ReplaceWith(new Modifiers) Modifiers {
	i.Override = new
	return i
}

// Transparent implements Modifiers.
func (i InheritedModifiers) Transparent() bool {
	return i.Parent.Transparent() || i.Override.Transparent()
}
