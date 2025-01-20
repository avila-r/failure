package trait

import "github.com/avila-r/failure/id"

type Trait struct {
	ID    uint64
	Label string
}

func New(label string) Trait {
	return Trait{
		ID:    id.Next(),
		Label: label,
	}
}
